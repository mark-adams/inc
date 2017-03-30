package backends

import "fmt"

var inMemoryTokens map[string]int64
var inMemoryNamespacedCounters map[string]int64

type InMemoryBackend struct{}

func NewInMemoryBackend() (*InMemoryBackend, error) {
	return &InMemoryBackend{}, nil
}

func (b *InMemoryBackend) Close() error {
	return nil
}

func (b *InMemoryBackend) CreateSchema() error {
	inMemoryTokens = make(map[string]int64)
	inMemoryNamespacedCounters = make(map[string]int64)
	return nil
}

func (b *InMemoryBackend) DropSchema() error {
	inMemoryTokens = nil
	return nil
}

func (b *InMemoryBackend) CreateToken(token string) error {
	inMemoryTokens[token] = 0
	return nil
}

func (b *InMemoryBackend) IncrementAndGetNamespacedToken(token string, namespace string) (int64, error) {
	_, tokenExists := inMemoryTokens[token]

	if !tokenExists {
		return 0, errInvalidToken
	}

	namespacedToken := fmt.Sprintf("%s:%s", token, namespace)
	val, ok := inMemoryNamespacedCounters[namespacedToken]

	if !ok {
		// Namespaced token doesnt exist, initialize it at 0
		inMemoryNamespacedCounters[namespacedToken] = val
		return val, nil
	} else {
		// Namespaced token exists, increment it
		val++
		inMemoryNamespacedCounters[namespacedToken] = val
		return val, nil
	}

}

func (b *InMemoryBackend) IncrementAndGetToken(token string) (int64, error) {
	val, ok := inMemoryTokens[token]
	if !ok {
		return 0, errInvalidToken
	}

	inMemoryTokens[token] = val + 1

	return val + 1, nil
}
