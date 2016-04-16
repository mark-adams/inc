package backends

import (
	"errors"
	"os"
	"strings"
)

// Backend represents the data store used by inc
type Backend interface {
	CreateSchema() error
	DropSchema() error
	CreateToken(token string) error
	IncrementAndGetToken(token string) (int64, error)
}

var errDatabase = errors.New("😧 The database is having some trouble... try again?")
var errInvalidToken = errors.New("invalid token")

func NewBackendFromString(url string) (Backend, error) {
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return nil, errors.New("Invalid connection string")
	}

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
	default:
		return nil, errors.New("Invalid backend type")
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
