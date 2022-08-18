package backends

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type RedisBackend struct {
	client *redis.Client
}

func NewRedisBackend(url string) (*RedisBackend, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("error parsing redis URL: %w", err)
	}

	return &RedisBackend{client: redis.NewClient(opts)}, nil
}

func (b *RedisBackend) Close() error {
	return b.client.Close()
}

func (b *RedisBackend) CreateSchema() error {
	return nil
}

func (b *RedisBackend) DropSchema() error {
	return nil
}

func (b *RedisBackend) CreateToken(token string) error {
	return b.client.Set(context.TODO(), token, 0, 0).Err()
}

func (b *RedisBackend) IncrementAndGetNamespacedToken(token string, namespace string) (int64, error) {
	namespacedToken := fmt.Sprintf("%s:%s", token, namespace)
	return b.IncrementAndGetToken(namespacedToken)
}

func (b *RedisBackend) IncrementAndGetToken(token string) (int64, error) {
	res := b.client.Incr(context.TODO(), token)
	if err := res.Err(); err != nil {
		return 0, err
	}

	return res.Val(), nil
}
