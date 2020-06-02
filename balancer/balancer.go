package balancer

import "Yalp/backend"

type Balancer interface {
	NextBackend() (*backend.Backend, error)
	AddBackend(backend *backend.Backend)
}
