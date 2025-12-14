package mdwr

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"go.uber.org/zap"

	"gateway/internal/service"
	"gateway/internal/users"
)

type mdwr struct {
	log *zap.Logger
	uc  *users.UsersClient
	svc service.Service
}

func NewMdwr(svc service.Service, uc *users.UsersClient, log *zap.Logger) *mdwr {

	return &mdwr{svc: svc, uc: uc, log: log}
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
			data, err := m.uc.ExtJWTData(tokenString, sk.String())
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

func (m *mdwr) Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			// /api/serviceName/etc

			var svcName, oper string
			if len(parts) >= 2 && parts[0] == "api" {
				svcName = parts[1]
			} else if len(parts) >= 1 && parts[0] == "api" {
				svcName = "api"
			} else if len(parts) >= 1 && parts[0] != "" {
				svcName = parts[0]
			} else {
				svcName = "root"
			}

			if len(parts) == 1 {
				// "/" or "/api"
				if parts[0] == "api" {
					oper = "api"
				} else {
					oper = "root"
				}
			} else if len(parts) == 2 && parts[0] == "api" {
				// /api/serviceName
				if r.Method == http.MethodPost {
					oper = "post"
				} else {
					oper = "root"
				}

			} else if len(parts) >= 3 {
				// /api/serviceName/etc
				candidate := parts[2]

				if len(candidate) >= 32 {
					oper = "by_uuid"
				} else {
					oper = candidate
				}
			} else {
				oper = "root"
			}

			timer := m.svc.NewTimer(svcName, oper)
			defer timer.ObserveDuration()

			m.svc.GetCounter(svcName).Inc()
			m.svc.GetCounter(svcName + "_" + oper).Inc()

			m.svc.GetActive().Inc()
			defer m.svc.GetActive().Dec()

			next.ServeHTTP(w, r)
		})
	}
}
