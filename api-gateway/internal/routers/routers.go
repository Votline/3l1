package routers

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"gateway/internal/mdwr"
	"gateway/internal/orders"
	"gateway/internal/service"
	"gateway/internal/users"
)

var resTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "seconds_for_operation",
	Help:    "Time spent processing requests",
	Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0},
}, []string{"service", "operation"})

const (
	gzipLevel              = 5
	maxConcurrencyRequests = 10
)

type Server struct {
	log  *zap.Logger
	Srv  *http.Server
	svcs []service.Service
}

func NewServer(log *zap.Logger) *Server {
	r := chi.NewRouter()
	s := Server{log: log}

	corsOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	corsMethods := strings.Split(os.Getenv("CORS_METHODS"), ",")
	if corsOrigins[0] == "" {
		corsOrigins = []string{"*"}
	}
	if corsMethods[0] == "" {
		corsMethods = []string{"*"}
	}
	c := cors.Options{
		MaxAge:           3600,
		AllowCredentials: true,
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   corsMethods,
	}

	rl := mdwr.NewRl(log)
	r.Use(rl.Middleware())

	groups := s.activateMdwr()
	r.Use(cors.Handler(c))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(gzipLevel))
	r.Use(middleware.Throttle(maxConcurrencyRequests))

	s.routing(r, groups)
	r.Handle("/metrics", promhttp.Handler())

	addr := ":" + os.Getenv("API_PORT")
	s.Srv = &http.Server{
		Handler:      r,
		Addr:         addr,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	s.Srv.SetKeepAlivesEnabled(true)

	return &s
}

func (s *Server) activateMdwr() []chi.Router {
	uc := users.New(resTime, s.log).(*users.UsersClient)
	services := []service.Service{
		uc,
		orders.New(resTime, s.log),
	}

	groups := make([]chi.Router, len(services))

	for i, svc := range services {
		g := chi.NewRouter()
		m := mdwr.NewMdwr(svc, uc.ExtJWTData, s.log)

		g.Use(m.RequestID())
		if svc.GetName() == "users" {
			g.Use(m.JWTAuth())
		}
		g.Use(m.Metrics())
		groups[i] = g
	}
	s.svcs = services
	return groups
}

func (s *Server) routing(r *chi.Mux, groups []chi.Router) {
	for i, svc := range s.svcs {
		path := "/api/" + svc.GetName()
		s.log.Debug("service: ", zap.String("path", path))

		r.Mount(path, groups[i])

		groups[i].Route("/", func(g chi.Router) {
			svc.RegisterRoutes(g)
		})
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root"))
	})
}

func (s *Server) ShutdownServices(ctx context.Context) error {
	var wg sync.WaitGroup
	done := make(chan struct{})
	for _, svc := range s.svcs {
		wg.Add(1)
		s.log.Info("Shutting down " + svc.GetName() + " service")
		go func() {
			defer wg.Done()
			svc.Close(ctx)
		}()
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
