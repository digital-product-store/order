package cart

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.elastic.co/apm/v2"
)

type RedisCartStorage struct {
	redisClient *redis.Client
}

func NewRedisCartStorage(url string) (*RedisCartStorage, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	storage := &RedisCartStorage{
		redisClient: client,
	}
	return storage, nil
}

func (r RedisCartStorage) Get(ctx context.Context, key string) (*Cart, error) {
	span, ctx := apm.StartSpan(ctx, "Get", "RedisCartStorage")
	defer span.End()

	value, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	cart := new(Cart)
	err = json.Unmarshal([]byte(value), cart)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (r RedisCartStorage) Set(ctx context.Context, key string, cart *Cart) error {
	span, ctx := apm.StartSpan(ctx, "Set", "RedisCartStorage")
	defer span.End()

	j, err := json.Marshal(cart)
	if err != nil {
		return err
	}

	value := string(j)
	err = r.redisClient.Set(ctx, key, value, 0).Err()
	return err
}

func (r RedisCartStorage) Delete(ctx context.Context, key string) error {
	span, ctx := apm.StartSpan(ctx, "Delete", "RedisCartStorage")
	defer span.End()

	err := r.redisClient.Del(ctx, key).Err()
	return err
}
