package balancer

import (
	"net/http/httputil"
)

type Balancer interface {
	NewReverseProxy() *httputil.ReverseProxy
}
