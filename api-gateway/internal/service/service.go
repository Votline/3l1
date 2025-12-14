package service

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
)

type Service interface {
	Close(context.Context) error
	GetName() string
	RegisterRoutes(chi.Router)
	NewTimer(string, string) *prometheus.Timer
	GetCounter(string) prometheus.Counter
	GetActive() prometheus.Gauge
}
type ctx struct {
	w http.ResponseWriter
	r *http.Request
}

func NewContext(w http.ResponseWriter, r *http.Request) *ctx {
	return &ctx{w: w, r: r}
}

func (c *ctx) Bind(v any) error {
	return json.NewDecoder(c.r.Body).Decode(v)
}
func (c *ctx) JSON(status int, v any) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(status)
	return json.NewEncoder(c.w).Encode(v)
}
func (c *ctx) SetSession(key string) {
	http.SetCookie(c.w, &http.Cookie{
		Name:     "session_key",
		Value:    key,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   os.Getenv("mode") == "production",
		SameSite: http.SameSiteLaxMode,
	})
}
func (c *ctx) Validate(req any) error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		if c.w != nil {
			http.Error(c.w, err.Error(), http.StatusBadRequest)
		}
		return err
	}
	return nil
}

func (c *ctx) Context() context.Context {
	return c.r.Context()
}
