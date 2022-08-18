package backends

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Backend represents the data store used by inc
type Backend interface {
	io.Closer
	CreateSchema() error
	DropSchema() error
	CreateToken(token string) error
	IncrementAndGetNamespacedToken(token string, namespace string) (int64, error)
	IncrementAndGetToken(token string) (int64, error)
}

var errDatabase = errors.New("😧 The database is having some trouble... try again?")
var errInvalidToken = errors.New("invalid token")

func NewBackendFromString(url string) (Backend, error) {
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return nil, errors.New("invalid connection string")
	}

	fmt.Println(url)

	switch parts[0] {
	case "postgres":
		backend, err := NewPostgresBackend(url)
		if err != nil {
			return nil, err
		}
		return backend, nil
	case "memory":
		backend, err := NewInMemoryBackend()
		if err != nil {
			return nil, err
		}
		return backend, nil
	case "redis":
		backend, err := NewRedisBackend(url)
		if err != nil {
			return nil, err
		}
		return backend, nil
	default:
		return nil, errors.New("invalid backend type")
	}
}

func GetBackendURL() string {
	if os.Getenv("DB_URL") != "" {
		return os.Getenv("DB_URL")
	} else if os.Getenv("PG_DB_URL") != "" {
		return os.Getenv("PG_DB_URL")
	} else if os.Getenv("REDIS_DB_URL") != "" {
		return os.Getenv("REDIS_DB_URL")
	}

	return ""
}

func GetBackend() (Backend, error) {
	return NewBackendFromString(GetBackendURL())
}
