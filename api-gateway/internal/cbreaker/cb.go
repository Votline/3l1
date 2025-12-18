package cbreaker

import (
	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
	"time"
)

func NewCb(name string, log *zap.Logger) *gobreaker.CircuitBreaker[any] {
	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5,
		Interval:    time.Minute,
		Timeout:     5 * time.Minute,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Error("CB changed",
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}
	return gobreaker.NewCircuitBreaker[any](st)
}
