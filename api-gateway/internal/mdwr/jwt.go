package mdwr

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"go.uber.org/zap"

	ck "gateway/internal/contextKeys"
	"gateway/internal/service"
)

type mdwr struct {
	log *zap.Logger
	ext func(string, string, string) (ck.UserInfo, error)
	svc service.Service
}

func NewMdwr(svc service.Service, ext func(string, string, string) (ck.UserInfo, error), log *zap.Logger) *mdwr {

	return &mdwr{svc: svc, ext: ext, log: log}
}

func (m *mdwr) JWTAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicRoute(r.URL.String()) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				m.log.Error("Failed to extract Authorization Header")
				http.Error(w, "Authorization header required", http.StatusBadRequest)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				m.log.Error("Invalid Authorization format", zap.Int("len parts", len(parts)), zap.String("Part 0", parts[0]))
				http.Error(w, "Invalid Authorization format", http.StatusBadRequest)
				return
			}

			tokenString := parts[1]
			sk, err := r.Cookie("session_key")

			if err != nil {
				m.log.Error("Failed to extract jwt data", zap.Error(err))
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			rq := r.Context().Value(ck.ReqKey).(string)
			data, err := m.ext(tokenString, sk.Value, rq)
			if err != nil {
				m.log.Error("Failed to extract data from JWT token", zap.Error(err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), ck.UserKey, data)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isPublicRoute(path string) bool {
	publicRotues := []string{
		"/api/users/reg",
		"/api/users/log",
		"/metrics",
		"/",
	}
	return slices.Contains(publicRotues, path)
}
