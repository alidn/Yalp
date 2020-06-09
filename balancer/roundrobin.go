package balancer

import (
	"Yalp/backend"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// RoundRobinBalancer is a load balancer that uses the Round-Robin approach
// to distribute requests across a group of servers.
type RoundRobinBalancer struct {
	backendPool   backend.Pool
	curBackendIdx int
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

// NewReverseProxy returns a New ReverseProxy that routes URLs to one of the servers among
// the load balancer servers.
func (r *RoundRobinBalancer) NewReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		backend, err := r.NextBackend()
		if err != nil {
			log.Fatal("ERR, could not get the next backend", err)
			// TODO: redirect to a url that shows the error?
			return
		}
		targetURL := backend.URL
		// directURL(req.URL, &targetURL)
		// req.URL = &targetURL
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		req.URL.RawQuery = targetURL.RawQuery
		req.URL.Path = targetURL.Path
	}

	return &httputil.ReverseProxy{Director: director}
}

func directURL(source, target *url.URL) {
	source = target
}

// NextBackend returns the next available server. If it reaches the end,
// it starts from the first server, and if no server is alive, it returns
// and error.
func (r *RoundRobinBalancer) NextBackend() (*backend.Backend, error) {
	if len(r.backendPool.Backends) == 0 {
		return nil, errors.New("There is no backend")
	}

	i := (r.curBackendIdx + 1) % len(r.backendPool.Backends)

	for counter := 0; counter < len(r.backendPool.Backends); counter++ {
		candidateBackend := r.backendPool.Backends[i]

		if candidateBackend.IsAlive {
			r.curBackendIdx = i
			return candidateBackend, nil
		}
		i = (i + 1) % len(r.backendPool.Backends)
	}
	return nil, errors.New("none of the servers is alive")
}

func (r *RoundRobinBalancer) GetCurIndex() int {
	return r.curBackendIdx
}

func (r *RoundRobinBalancer) AddBackend(backend *backend.Backend) {
	r.backendPool.Backends = append(r.backendPool.Backends, backend)
}
