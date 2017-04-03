package backends

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-redis/redis"
)

type RedisBackend struct {
	client *redis.Client
}

func NewRedisBackend(connString string) (*RedisBackend, error) {
	u, _ := url.Parse(connString)
	dbParam := u.Query().Get("database")

	// Default to DB 0
	var db int
	if dbParam != "" {
		fmt.Print(dbParam)
		db, _ = strconv.Atoi(dbParam)
	} else {
		db = 0
	}

	host := u.Host
	password := u.Query().Get("password")

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	return &RedisBackend{
		client: client,
	}, nil
}

func (r *RedisBackend) Close() error {
	return r.client.Close()
}

func (r *RedisBackend) CreateSchema() error {
	return nil
}

func (r *RedisBackend) DropSchema() error {
	r.client.FlushDb()
	return nil
}

func (r *RedisBackend) CreateToken(token string) error {
	err := r.client.Set(token, 0, 0).Err()

	if err != nil {
		return err
	}

	return nil
}

func (r *RedisBackend) IncrementAndGetNamespacedToken(token string, namespace string) (int64, error) {
	nsToken := token + namespace
	_, err := r.client.Get(nsToken).Result()

	// If namespace doesn't exist, create it and set count to 0
	if err != nil {
		// Assert that the base token exists
		_, err := r.client.Get(token).Result()
		if err != nil {
			return 0, errInvalidToken
		}

		err = r.client.Set(nsToken, 0, 0).Err()
		if err != nil {
			return 0, err
		}

		return 0, nil
	}

	// Otherwise, get and increment
	if err := r.client.Incr(nsToken).Err(); err != nil {
		return 0, err
	}

	n, err := r.client.Get(nsToken).Int64()

	if err != nil {
		return 0, err
	}

	return n, nil
}

func (r *RedisBackend) IncrementAndGetToken(token string) (int64, error) {
	// Assert the token has been created already
	_, err := r.client.Get(token).Result()

	if err != nil {
		return 0, errInvalidToken
	}

	if err := r.client.Incr(token).Err(); err != nil {
		return 0, errInvalidToken
	}

	n, err := r.client.Get(token).Int64()

	if err != nil {
		return 0, err
	}

	return n, nil
}
