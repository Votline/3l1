package routers

import (
	"os"
	"time"
	"strings"
	"net/http"

	"go.uber.org/zap"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"gateway/internal/mdwr"
	"gateway/internal/users"
	"gateway/internal/orders"
	"gateway/internal/service"
)

var resTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "seconds_for_operation",
	Help: "Time spent processing requests",
	Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0},
},[]string{"service", "operation"})

const (
	gzipLevel = 5
	maxConcurrencyRequests = 10
)

func NewServer(log *zap.Logger) *http.Server {
	r := chi.NewRouter()

	corsOrigins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
	corsMethods := strings.Split(os.Getenv("CORS_METHODS"), ",")
	if corsOrigins[0] == "" {
		corsOrigins = []string{"*"}
	}
	if corsMethods[0] == "" {
		corsMethods = []string{"*"}
	}
	c := cors.Options{
		MaxAge: 3600,
		AllowCredentials: true,
		AllowedOrigins: corsOrigins,
		AllowedMethods: corsMethods,
	}

	rl := mdwr.NewRl(log)
	r.Use(rl.Middleware())

	svcs, groups := activateMdwr(log)
	r.Use(cors.Handler(c))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(gzipLevel))
	r.Use(middleware.Throttle(maxConcurrencyRequests))

	routing(r, svcs, groups, log)
	r.Handle("/metrics", promhttp.Handler())

	addr := ":"+os.Getenv("API_PORT")
	return &http.Server{
		Handler: r,
		Addr: addr,
		ReadTimeout: 20*time.Second,
		WriteTimeout: 20*time.Second,
		IdleTimeout: 60*time.Second,
	}
}

func activateMdwr(log *zap.Logger) ([]service.Service, []chi.Router) {
	uc := users.New(resTime, log).(*users.UsersClient)
	services := []service.Service{
		uc,
		orders.New(resTime, log),
	}

	groups := make([]chi.Router, len(services))

	for i, svc := range services {
		g := chi.NewRouter()
		m := mdwr.NewMdwr(svc, uc, log)

		if svc.GetName() == "users" {
			g.Use(m.JWTAuth())
		}
		g.Use(m.Metrics())
		groups[i] = g
	}
	return services, groups
}

func routing(r *chi.Mux, services []service.Service, groups []chi.Router, log *zap.Logger) {
	for i, svc := range services {
		path := "/api/"+svc.GetName()
		log.Debug("service: ", zap.String("path", path))

		r.Mount(path, groups[i])

		groups[i].Route("/", func(g chi.Router){
			svc.RegisterRoutes(g)
		})
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("root"))
	})
}
