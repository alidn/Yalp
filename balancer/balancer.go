package balancer

import (
	"net/http/httputil"

	"github.com/Yalp/backend"
)

type Balancer interface {
	NextBackend() (*backend.Backend, error)
	NewReverseProxy() *httputil.ReverseProxy
}
