package mdwr

import (
	"os"
	"time"
	"strings"
	"net/http"

	"go.uber.org/zap"
	"github.com/go-redis/redis"
)

const exp = time.Minute

type rateLimiter struct {
	maxRate  int64
	log *zap.Logger
	rdb *redis.Client
}
func NewRl(log *zap.Logger) *rateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PSWD"),
		DB: 0,
	})
	return &rateLimiter{
		rdb: rdb,
		log: log,
		maxRate: 50,
	}
}

func (rl *rateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			key := "rl:" + ip

			count, err := rl.rdb.Incr(key).Result()
			if err != nil {
				rl.log.Error("Failed to increment by key", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rl.rdb.Expire(key, exp)
			}

			if count > rl.maxRate {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
