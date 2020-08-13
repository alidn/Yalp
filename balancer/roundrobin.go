package balancer

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/Yalp/backend"

	"github.com/google/uuid"
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

func checkSessionPersistenceCookie(req *http.Request) (uuid.UUID, bool, error) {
	for _, cookie := range req.Cookies() {
		if cookie.Name == SessionPersistenceCookieName {
			id, err := uuid.Parse(cookie.Value)
			if err != nil {
				return uuid.UUID{}, true, err
			}
			return id, true, nil
		}
	}
	return uuid.UUID{}, false, nil
}

func (r *RoundRobinBalancer) checkBackendSession(req *http.Request) (*backend.Backend, bool, error) {
	id, foundCookie, err := checkSessionPersistenceCookie(req)
	if err != nil {
		return nil, false, err
	}
	if foundCookie {
		nextBackend, err := r.backendPool.Get(id)
		if err != nil {
			return nil, false, err
		}
		return nextBackend, true, nil
	}
	return nil, false, nil
}

// NewReverseProxy returns a new ReverseProxy that routes URLs to one of the servers among
// the load balancer servers.
func (r *RoundRobinBalancer) NewReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		var nextBackend *backend.Backend = nil
		if r.Config.SessionPersistenceConfig.Enabled {
			b, foundCookie, err := r.checkBackendSession(req)
			if err != nil {
				fmt.Errorf("checking the session of thr request")
				return
			}
			if foundCookie {
				nextBackend = b
			}
		}
		// no session found, start a new one
		if nextBackend == nil {
			var err error
			nextBackend, err = r.NextBackend()
			if err != nil {
				log.Fatal("could not get the next nextBackend", err)
				return
			}
			if r.Config.SessionPersistenceConfig.Enabled {
				sessionPersistenceCookie := r.createCookie(nextBackend.Id)
				req.AddCookie(&sessionPersistenceCookie)
				c := http.Cookie{
					Name:  "SessionExists",
					Value: "false",
				}
				req.AddCookie(&c)
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
				fmt.Printf("%s = %s\n", cookie.Name, cookie.Value)
				if cookie.Name == "SessionExists" && cookie.Value == "true" {
					return nil
				}
			}

			for _, cookie := range response.Request.Cookies() {
				if cookie.Name == SessionPersistenceCookieName {
					expirationPeriod := time.Duration(r.Config.SessionPersistenceConfig.ExpirationPeriod) * time.Second
					cookie.Expires = time.Now().Add(expirationPeriod)
					response.Header.Add("Set-Cookie", cookie.String())
					c := http.Cookie{
						Name:  "SessionExists",
						Value: "true",
					}
					c.Expires = time.Now().Add(expirationPeriod)
					response.Header.Add("Set-Cookie", c.String())
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
