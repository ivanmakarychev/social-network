package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivanmakarychev/social-network/counter-service/internal/config"

	"github.com/go-redis/redis/v8"
	"github.com/ivanmakarychev/social-network/counter-service/internal/models"
)

type CounterGetter interface {
	Get(ctx context.Context, key models.UnreadMessagesCounterKey) (int, error)
}

type CounterRepository interface {
	CounterGetter
	Increment(ctx context.Context, key models.UnreadMessagesCounterKey, by int) error
}

type RedisCounterRepository struct {
	redisClient *redis.Client
}

func NewRedisCounterRepository(redisClient *redis.Client) *RedisCounterRepository {
	return &RedisCounterRepository{redisClient: redisClient}
}

func MakeRedisCounterRepository(cfg config.Redis) (*RedisCounterRepository, error) {
	cl := redis.NewClient(&redis.Options{
		Addr: cfg.Address,
	})
	pingErr := cl.Ping(context.Background()).Err()
	return NewRedisCounterRepository(cl), pingErr
}

func (r *RedisCounterRepository) Get(ctx context.Context, key models.UnreadMessagesCounterKey) (int, error) {
	count, err := r.redisClient.Get(ctx, toRedisKey(key)).Int()
	if err == nil {
		return count, nil
	}
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return 0, err
}

func (r *RedisCounterRepository) Increment(ctx context.Context, key models.UnreadMessagesCounterKey, by int) error {
	return r.redisClient.IncrBy(ctx, toRedisKey(key), int64(by)).Err()
}

func toRedisKey(key models.UnreadMessagesCounterKey) string {
	return fmt.Sprintf("%d:%d", key.From, key.To)
}
