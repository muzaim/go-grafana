package queue

import (
	"context"
	"fmt"
	"time"

	"go-up-monitoring/internal/config"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Queue struct {
	RedisClient *redis.Client
	AsynqClient *asynq.Client
	RedisOpt    asynq.RedisClientOpt
}

func NewQueue(cfg config.RedisConfig) *Queue {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	asynqClient := asynq.NewClient(redisOpt)

	return &Queue{
		RedisClient: rdb,
		AsynqClient: asynqClient,
		RedisOpt:    redisOpt,
	}
}

func (q *Queue) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := q.RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}
	return nil
}

func (q *Queue) Close() error {
	var errs []error
	if q.AsynqClient != nil {
		if err := q.AsynqClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if q.RedisClient != nil {
		if err := q.RedisClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing queue: %v", errs)
	}
	return nil
}
