package balancer

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"sync/atomic"

	"github.com/alidn/Yalp/backend"
	"github.com/google/uuid"
)

type BackendWithConnState struct {
	backend.RoundRobinBackend
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

func (b *BackendPoolWithConnState) getBackend(backendID string) (*BackendWithConnState, error) {
	backendUUID, err := uuid.Parse(backendID)
	if err != nil {
		return nil, err
	}
	for _, bckend := range b.Backends {
		if bckend.Id == backendUUID {
			return bckend, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("did not find a backend with the given id: %s", backendID))
}

func NewConnBackendPoolFromURLs(urls ...string) (*BackendPoolWithConnState, error) {
	pool := &BackendPoolWithConnState{
		Backends: make([]*BackendWithConnState, 0),
	}
	for _, url := range urls {
		b, err := backend.NewBackend(url)
		if err != nil {
			return nil, err
		}
		pool.Backends = append(pool.Backends, &BackendWithConnState{
			RoundRobinBackend: *b,
			OpenConnections:   0,
		})
	}
	return pool, nil
}

func NewLeastConnectionBalancerFromURLs(config Config, urls ...string) (*LeastConnectionsBalancer, error) {
	backendPool, err := NewConnBackendPoolFromURLs(urls...)
	if err != nil {
		return nil, err
	}
	return &LeastConnectionsBalancer{
		backendPool: *backendPool,
		Config:      config,
	}, nil
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

func attachBackendIDCookie(request *http.Request, backendID uuid.UUID) {
	cookie := &http.Cookie{
		Name:  "BackendID",
		Value: backendID.String(),
	}
	request.AddCookie(cookie)
}

func (l *LeastConnectionsBalancer) NewReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		nextBackend, err := l.NextBackend()
		// log.Print(nextBackend.OpenConnections, nextBackend.URL)
		if err != nil {
			log.Fatal("Could not get the next backend")
			return
		}

		attachBackendIDCookie(req, nextBackend.Id)
		nextBackend.addOpenConnections(1)

		targetURL := nextBackend.URL
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		req.URL.RawQuery = targetURL.RawQuery
		req.URL.Path = targetURL.Path
	}

	return &httputil.ReverseProxy{
		Director: director,
		ModifyResponse: func(response *http.Response) error {
			go func() error {
				for _, cookie := range response.Request.Cookies() {
					if cookie.Name == "BackendID" {
						backendID := cookie.Value
						bckend, err := l.backendPool.getBackend(backendID)
						if err != nil {
							return err
						}
						bckend.reduceOpenConnections(1)
						return nil
					}
				}
				return nil
			}()
			return nil
		},
	}
}
