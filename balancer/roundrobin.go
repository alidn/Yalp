package balancer

import (
	"Yalp/backend"
	"errors"
	"sync"
)

// RoundRobinBalancer is a load balancer that uses the Round-Robin approach
// to distribute requests across a group of servers.
type RoundRobinBalancer struct {
	backendPool   backend.BackendPool
	curBackendIdx int
	sync.Mutex
}

// NewRoundRobinBalancer constructs and returns a RoundRobinBalancer with
// no backends.
func NewRoundRobinBalancer() *RoundRobinBalancer {
	backendPool := backend.NewBackendPool()
	return &RoundRobinBalancer{
		backendPool:   *backendPool,
		curBackendIdx: 0,
	}
}

// NewRoundRobinBalancerWithURLs constructs and returns a RoundRobinBalancer
// for the given list of urls.
func NewRoundRobinBalancerWithURLs(urls ...string) (*RoundRobinBalancer, error) {
	backendPool, err := backend.NewBackendPoolFromURLs(urls...)
	if err != nil {
		return nil, err
	}
	return &RoundRobinBalancer{
		backendPool:   *backendPool,
		curBackendIdx: -1,
	}, nil
}

// NextBackend returns the next avaiable server. If it reaches the end,
// it starts from the first server, and if no server is alive, it returns
// and error.
func (r *RoundRobinBalancer) NextBackend() (*backend.Backend, error) {
	if len(r.backendPool.Backends) == 0 {
		return nil, errors.New("There is no backend")
	}

	// TODO: should lock the backendpool?
	r.Lock()
	defer r.Unlock()

	i := (r.curBackendIdx + 1) % len(r.backendPool.Backends)

	for counter := 0; counter < len(r.backendPool.Backends); counter++ {
		candidateBackend := r.backendPool.Backends[i]

		if candidateBackend.IsAlive {
			r.curBackendIdx = i
			return candidateBackend, nil
		}
		i = (i + 1) % len(r.backendPool.Backends)
	}
	return nil, errors.New("None of the servers is alive")
}

func (r RoundRobinBalancer) GetCurIndex() int {
	return r.curBackendIdx
}

func (r *RoundRobinBalancer) AddBackend(backend *backend.Backend) {
	r.backendPool.Backends = append(r.backendPool.Backends, backend)
}
