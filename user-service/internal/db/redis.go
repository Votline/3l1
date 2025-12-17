package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"

	gc "users/internal/graceful"
)

type RedisRepo struct {
	ctx context.Context
	log *zap.Logger
	rdb *redis.Client
}

func NewRR(log *zap.Logger) *RedisRepo {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_SK_HOST") + ":6379",
		Password: os.Getenv("REDIS_SK_PSWD"),
		DB:       0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed connect to sk redis", zap.Error(err))
		return nil
	}

	return &RedisRepo{rdb: rdb, log: log, ctx: ctx}
}
func (r *RedisRepo) Stop(ctx context.Context) error {
	return gc.Shutdown(r.rdb.Close, ctx)
}

func (r *RedisRepo) NewSession(id, role string) (string, error) {
	const op = "UserRedisRepository.NewSession"
	
	sk := uuid.NewString()
	tx := r.rdb.TxPipeline()

	if err := tx.HSet(r.ctx, sk, map[string]string{
		"id":   id,
		"role": role,
	}).Err(); err != nil {
		return "", fmt.Errorf("%s: tx add entry: %w", op, err)
	}

	if err := tx.Expire(r.ctx, sk, time.Hour*720).Err(); err != nil {
		return "", fmt.Errorf("%s: tx expire entry: %w", op, err)
	}

	if _, err := tx.Exec(r.ctx); err != nil {
		return "", fmt.Errorf("%s: new session: %w", op, err)
	}

	return sk, nil
}

func (r *RedisRepo) Validate(id, role, sk string) error {
	const op = "UserRedisRepository.Validate"

	fields, err := r.rdb.HGetAll(r.ctx, sk).Result()
	if err != nil {
		return fmt.Errorf("%s: get all: %w", op, err)
	}
	if fields["id"] != id || fields["role"] != role {
		err := fmt.Errorf("Data are different: %s: %s | %s: %s",
			id, fields["id"], role, fields["role"])
		return fmt.Errorf("%s: match data: %w", op, err)
	}
	return nil
}

func (r *RedisRepo) DelSession(sk string) error {
	const op = "UserRedisRepository.DelSession"

	if err := r.rdb.Del(r.ctx, sk).Err(); err != nil {
		return fmt.Errorf("%s: delete session: %w", op, err)
	}
	return nil
}
