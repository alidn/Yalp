package balancer

import (
	"Yalp/backend"
	"net/http/httputil"
)

type Balancer interface {
	NextBackend() (*backend.Backend, error)
	NewReverseProxy() *httputil.ReverseProxy
}
