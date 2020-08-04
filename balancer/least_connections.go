package balancer

import (
	"Yalp/backend"
	"errors"
	"github.com/google/uuid"
	"math"
	"net/http/httputil"
	"sync/atomic"
)

type BackendWithConnState struct {
	backend.Backend
	OpenConnections uint32
}

func (b *BackendWithConnState) addOpenConnections(delta uint32) {
	atomic.AddUint32(&b.OpenConnections, delta)
}

func (b *BackendWithConnState) reduceOpenConnections(delta uint32) {
	atomic.AddUint32(&b.OpenConnections, ^uint32(delta-1))
}

type BackendPoolWithConnState struct {
	Backends []*BackendWithConnState
}

func NewBackendPoolFromURLs(urls ...string) (*BackendPoolWithConnState, error) {
	pool := &BackendPoolWithConnState{
		Backends: make([]*BackendWithConnState, 0),
	}
	for _, url := range urls {
		b, err := backend.NewBackend(url)
		if err != nil {
			return nil, err
		}
		pool.Backends = append(pool.Backends, &BackendWithConnState{
			Backend:         *b,
			OpenConnections: 0,
		})
	}
	return pool, nil
}

type LeastConnectionsBalancer struct {
	backendPool BackendPoolWithConnState
	Config      Config
}

func (l *LeastConnectionsBalancer) NextBackend() (*BackendWithConnState, error) {
	if len(l.backendPool.Backends) == 0 {
		return nil, errors.New("there is no backend")
	}
	index, minConnections := 0, uint32(math.MaxUint32)
	for i, b := range l.backendPool.Backends {
		if b.OpenConnections < minConnections {
			minConnections = b.OpenConnections
			index = i
		}
	}
	return l.backendPool.Backends[index], nil
}

func (l *LeastConnectionsBalancer) NewReverseProxy() *httputil.ReverseProxy {

}
