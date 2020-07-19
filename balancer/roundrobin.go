package balancer

import (
	"Yalp/backend"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

const SessionPersistenceCookieName string = "LoadBalancerSessionCookie"

// RoundRobinBalancer is a load balancer that uses the Round-Robin approach
// to distribute requests across a group of servers.
type RoundRobinBalancer struct {
	backendPool   backend.Pool
	curBackendIdx int
	Config        Config
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
		var nextBackend *backend.Backend = nil
		if r.Config.SessionPersistenceConfig.Enabled {
			for _, cookie := range req.Cookies() {
				if cookie.Name == SessionPersistenceCookieName {
					id, err := uuid.Parse(cookie.Value)
					if err != nil {
						log.Fatal("Could not parse session-persistence cookie value", err)
						return
					}
					nextBackend, err = r.backendPool.Get(id)
					if err != nil {
						log.Fatal(err)
						return
					}
				}
			}
		}
		if nextBackend == nil {
			var err error
			nextBackend, err = r.NextBackend()
			if err != nil {
				log.Fatal("ERR, could not get the next nextBackend", err)
				// TODO: redirect to a url that shows the error?
				return
			}
			if r.Config.SessionPersistenceConfig.Enabled {
				println("setting new cookie")
				sessionPersistenceCookie := r.createCookie(nextBackend.Id)
				req.AddCookie(&sessionPersistenceCookie)
			}
		}

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
			for _, cookie := range response.Request.Cookies() {
				if cookie.Name == SessionPersistenceCookieName {
					println("Setting cookie")
					cookie.Expires = time.Now().Add(time.Second * 1)
					response.Header.Add("Set-Cookie", cookie.String())
				}
			}
			return nil
		},
	}
}

func (r *RoundRobinBalancer) createCookie(id uuid.UUID) http.Cookie {
	expirationPeriod := time.Duration(r.Config.SessionPersistenceConfig.ExpirationPeriod) * time.Second
	cookie := http.Cookie{
		Name:     SessionPersistenceCookieName,
		Value:    id.String(),
		Expires:  time.Now().Add(expirationPeriod),
		SameSite: http.SameSiteNoneMode,
	}
	return cookie
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
