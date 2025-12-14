package mdwr

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	ck "gateway/internal/contextKeys"
)

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

			next.ServeHTTP(w, r)
		})
	}
}

func (m *mdwr) RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := uuid.NewString()
			ctx := context.WithValue(r.Context(), ck.ReqKey, reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
