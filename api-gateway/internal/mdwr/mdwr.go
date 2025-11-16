package mdwr

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"go.uber.org/zap"

	"gateway/internal/users"
)

type Mdwr struct {
	log *zap.Logger
	uc  *users.UsersClient
}

func NewMdwr(uc *users.UsersClient, log *zap.Logger) *Mdwr {
	return &Mdwr{uc: uc, log: log}
}

func (m *Mdwr) JWTAuth() func(http.Handler) http.Handler {
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
			data, err := m.uc.ExtJWTData(tokenString)
			if err != nil {
				m.log.Error("Failed to extract data from JWT token", zap.Error(err))
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
		"/metrics",
		"/",
	}

	return slices.Contains(publicRotues, path)
}

func (m *Mdwr) Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			oper := strings.SplitN(strings.TrimSuffix(r.URL.Path, "/"), "/", 5)

			var mainLabel, operLabel string
			switch len(oper) { //'/api/serviceName/oper/etc
			case 0, 1, 2: //"", api
				mainLabel, operLabel = "api", "api"
			case 3: //"", api, serviceName
				mainLabel, operLabel = oper[2], oper[2] //serviceName
			case 4, 5: //"", api, serviceName, operation, etc will be ignored
				operLabel = oper[2] + "_" + oper[3] //serviceName+oper
				mainLabel = oper[2] //serviceName
			default:
				m.log.Error("Failed to split url", zap.Int("oper len:", len(oper)))
			}

			timer := m.uc.NewTimer(operLabel)
			defer timer.ObserveDuration()

			m.uc.Counter.WithLabelValues(mainLabel).Inc()
			m.uc.Counter.WithLabelValues(operLabel).Inc()
			m.uc.Active.Inc()
			defer m.uc.Active.Dec()

			next.ServeHTTP(w, r)
		})
	}
}
