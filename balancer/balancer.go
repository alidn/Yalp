package balancer

import (
	"net/http/httputil"

	"github.com/alidn/Yalp/backend"
)

type Balancer interface {
	NextBackend() (*backend.Backend, error)
	NewReverseProxy() *httputil.ReverseProxy
}
