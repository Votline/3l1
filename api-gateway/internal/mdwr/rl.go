package mdwr

import (
	"os"
	"time"
	"strings"
	"context"
	"net/http"

	"go.uber.org/zap"
	"github.com/go-redis/redis/v8"
)

const exp = time.Minute

type rateLimiter struct {
	maxRate  int64
	log *zap.Logger
	rdb *redis.Client
	ctx context.Context
}
func NewRl(log *zap.Logger) *rateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_RL_HOST") + ":6379",
		Password: os.Getenv("REDIS_RL_PSWD"),
		DB: 0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed connect to rl redis", zap.Error(err))
	}

	return &rateLimiter{
		rdb: rdb,
		ctx: ctx,
		log: log,
		maxRate: 50,
	}
}

func (rl *rateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			key := "rl:" + ip

			count, err := rl.rdb.Incr(rl.ctx, key).Result()
			if err != nil {
				rl.log.Error("Failed to increment by key", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rl.rdb.Expire(rl.ctx, key, exp)
			}

			if count > rl.maxRate {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
