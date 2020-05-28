package types

type Balancer interface {
	NextBackend() (*Backend, error)
}
