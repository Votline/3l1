package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxRetries = 3

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

func (c *ctx) Context() context.Context {
	return c.r.Context()
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

func rpc[T any](fn func() (T, error)) (T, error) {
	var zero T
	for i := 0; i < maxRetries; i++ {
		res, err := fn()
		if err == nil {
			return res, nil
		}

		if !shouldRetry(err) {
			return zero, err
		}

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return zero, fmt.Errorf("max retries exceeded")
}

func shouldRetry(err error) bool {
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case
			codes.Canceled,
			codes.DeadlineExceeded,
			codes.ResourceExhausted,
			codes.Aborted,
			codes.Unavailable,
			codes.DataLoss:

			return true
		case
			codes.InvalidArgument,
			codes.NotFound,
			codes.AlreadyExists,
			codes.PermissionDenied,
			codes.FailedPrecondition,
			codes.OutOfRange,
			codes.Unimplemented,
			codes.Internal,
			codes.Unauthenticated:

			return false
		default:
			return false
		}
	}

	return false
}

func Execute[T any](cb *gobreaker.CircuitBreaker[any], fn func() (T, error)) (T, error) {
	var zero T

	resCb, err := cb.Execute(func() (any, error) {
		return rpc(func() (T, error) {
			return fn()
		})
	})

	if err != nil {
		return zero, err
	}

	res, ok := resCb.(T)
	if !ok {
		return zero, fmt.Errorf("Failed parse response to needed type")
	}
	return res, nil
}
