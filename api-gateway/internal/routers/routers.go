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

	"gateway/internal/users"
	"gateway/internal/orders"
	"gateway/internal/service"
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

	r.Use(cors.Handler(c))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Throttle(10))

	routing(r, log)

	addr := ":"+os.Getenv("API_PORT")
	return &http.Server{
		Handler: r,
		Addr: addr,
		ReadTimeout: 20*time.Second,
		WriteTimeout: 20*time.Second,
		IdleTimeout: 60*time.Second,
	}
}

func routing(r *chi.Mux, log *zap.Logger) {
	services := []service.Service{
		users.New(log),
		orders.New(log),
	}

	for _, svc := range services {
		path := "/api/"+svc.GetName()
		log.Debug("service: ", zap.String("path", path))

		r.Route(path, func(g chi.Router){
			svc.RegisterRoutes(g)
		})
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("root"))
	})
}
