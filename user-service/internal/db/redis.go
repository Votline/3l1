package db

import (
	"context"
	"errors"
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
	sk := uuid.NewString()
	tx := r.rdb.TxPipeline()

	if err := tx.HSet(r.ctx, sk, map[string]string{
		"id":   id,
		"role": role,
	}).Err(); err != nil {
		r.log.Error("Failed to add new redis hash entry", zap.Error(err))
		return "", err
	}

	if err := tx.Expire(r.ctx, sk, time.Hour*720).Err(); err != nil {
		r.log.Error("Failed to expire new session", zap.Error(err))
		return "", err
	}

	if _, err := tx.Exec(r.ctx); err != nil {
		r.log.Error("Failed to create session", zap.Error(err))
		return "", err
	}

	return sk, nil
}

func (r *RedisRepo) Validate(id, role, sk string) error {
	fields, err := r.rdb.HGetAll(r.ctx, sk).Result()
	if err != nil {
		r.log.Error("Failed to get all fields", zap.Error(err))
		return err
	}
	if fields["id"] != id || fields["role"] != role {
		r.log.Error("Different data",
			zap.String("Original id", id),
			zap.String("Original role", role),
			zap.String("Session id", fields["id"]),
			zap.String("Session role", fields[role]))
		return errors.New("Data are different")
	}
	return nil
}

func (r *RedisRepo) DelSession(sk string) error {
	if err := r.rdb.Del(r.ctx, sk).Err(); err != nil {
		r.log.Error("Failed to delete session key", zap.Error(err))
		return err
	}
	return nil
}
