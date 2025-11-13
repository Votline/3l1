package mdwr

import (
	"slices"
	"strings"
	"context"
	"net/http"
	
	"go.uber.org/zap"

	"gateway/internal/users"
)

type Auth struct {
	log *zap.Logger
	uc *users.UsersClient
}
func NewAuth(uc *users.UsersClient, log *zap.Logger) *Auth {
	return &Auth{uc: uc, log: log}
}

func (a *Auth) JWTAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicRoute(r.URL.String()) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				a.log.Error("Failed to extract Authorization Header")
				http.Error(w, "Authorization header required", http.StatusBadRequest)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				a.log.Error("Invalid Authorization format", zap.Int("len parts", len(parts)), zap.String("Part 0", parts[0]))
				http.Error(w, "Invalid Authorization format", http.StatusBadRequest)
				return
			}

			tokenString := parts[1]
			data, err := a.uc.ExtJWTData(tokenString)
			if err != nil {
				a.log.Error("Failed to extract data from JWT token", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), "userInfo", data)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicRoute(path string) bool {
	publicRotues := []string{
		"/api/users/reg",
		"/api/users/log",
		"/",
	}

	return slices.Contains(publicRotues, path)
}
