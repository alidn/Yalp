package balancer

import "loadBalancer/backend"

type RoundRobinBalancer struct {
	backendPool backend.BackendPool
	curBackendIdx uint
}

func (r RoundRobinBalancer) NextBackend() (*backend.Backend, error) {
	i := r.curBackendIdx
	// TODO: return error if none of the backends is alive
	for {
		candidateBackend := r.backendPool.Backends[i]
		if candidateBackend.IsAlive {
			r.curBackendIdx = i
			return candidateBackend, nil
		} else {
			i = i + 1
		}
	}
}

func (r RoundRobinBalancer) AddBackend(backend *backend.Backend) {
	r.backendPool.Backends = append(r.backendPool.Backends, backend)
	err := backend.StartHealthCheck()
	if err != nil {
		println("Could not start health check", err.Error())
	}
}


