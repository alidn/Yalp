package backend

type BackendPool struct {
	Backends []*Backend
}

func NewBackendPool() *BackendPool {
	return &BackendPool{
		Backends: make([]*Backend, 0),
	}
}

// NewBackendPoolFromURLs constructs and returns a new BackendPool using the
// given urls.
func NewBackendPoolFromURLs(urls ...string) (*BackendPool, error) {
	backendPool := &BackendPool{
		Backends: make([]*Backend, 0),
	}
	for _, url := range urls {
		backend, err := NewBackend(url)
		if err != nil {
			return nil, err
		}
		backendPool.Backends = append(backendPool.Backends, backend)
	}

	return backendPool, nil
}
