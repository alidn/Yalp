package backend

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type Pool struct {
	Backends []*RoundRobinBackend
}

func NewBackendPool() *Pool {
	return &Pool{
		Backends: make([]*RoundRobinBackend, 0),
	}
}

func (p *Pool) Get(id uuid.UUID) (*RoundRobinBackend, error) {
	for _, backend := range p.Backends {
		if backend.Id == id {
			return backend, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("did not find a backend with the given id: %s", id))
}

// NewBackendPoolFromURLs constructs and returns a new BackendPool using the
// given urls.
func NewBackendPoolFromURLs(urls ...string) (*Pool, error) {
	backendPool := &Pool{
		Backends: make([]*RoundRobinBackend, 0),
	}
	for _, url := range urls {
		backend, err := NewBackend(url)
		if err != nil {
			return nil, err
		}
		backendPool.Backends = append(backendPool.Backends, backend)
	}

	return backendPool, nil
}
