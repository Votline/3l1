package routers

import (
	"os"
	"time"
	"net/http"

	"go.uber.org/zap"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"gateway/internal/users"
	"gateway/internal/service"
)

func NewServer(log *zap.Logger) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

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
	}

	for _, svc := range services {
		path := "/api/"+svc.GetName()
		log.Debug("service: ", zap.String("path", path))

		r.Route(path, func(g chi.Router){
			svc.RegisterRoutes(g)
		})
	}
}
