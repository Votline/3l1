package routers

import (
	"time"
	"net/http"

	"go.uber.org/zap"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewServer(log *zap.Logger) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	routing(r)

	addr := ":8443"
	return &http.Server{
		Handler: r,
		Addr: addr,
		ReadTimeout: 20*time.Second,
		WriteTimeout: 20*time.Second,
		IdleTimeout: 60*time.Second,
	}
}

func routing(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte("root"))
	})
}
