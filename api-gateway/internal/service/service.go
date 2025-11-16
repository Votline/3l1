package service

import (
	"context"
	"net/http"
	"encoding/json"

	"github.com/go-chi/chi"
)

type Service interface{
	Close() error
	GetName() string
	RegisterRoutes(chi.Router)
}
type ctx struct {
	w http.ResponseWriter
	r *http.Request
}
func NewContext(w http.ResponseWriter, r *http.Request) *ctx {
	return &ctx{w:w, r:r}
}

func(c *ctx) Bind(v any) error {
	return json.NewDecoder(c.r.Body).Decode(v)
}
func (c *ctx) JSON(status int, v any) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(status)
	return json.NewEncoder(c.w).Encode(v)
}

func (c *ctx) Context() context.Context {
	return c.r.Context()
}
