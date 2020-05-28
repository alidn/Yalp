package types

type Backend struct {
	addr string
}

func NewBackend(addr string) *Backend {
	return &Backend{
		addr: addr,
	}
}