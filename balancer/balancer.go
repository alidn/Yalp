package balancer

import "loadBalancer/backend"

type Balancer interface {
	NextBackend() (*backend.Backend, error)
	AddBackend(backend *backend.Backend)
}
