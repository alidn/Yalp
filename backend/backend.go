package backend

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Backend represents a backend server.
type Backend struct {
	Addr string
	URL  url.URL
	// shows whether or not the server is alive.
	IsAlive           bool
	healthCheckTicker *time.Ticker
	sync.RWMutex
}

func NewBackend(addr string) (*Backend, error) {
	parsedURL, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	backend := &Backend{
		Addr:    addr,
		IsAlive: true,
		URL:     *parsedURL,
	}
	go func() {
		isAlive, err := backend.CheckAlive()
		if err != nil {
			backend.IsAlive = false
		} else {
			backend.IsAlive = isAlive
		}
	}()

	go func() {
		err2 := backend.StartHealthCheck()
		err = err2
	}()
	if err != nil {
		return nil, err
	}
	return backend, nil
}

// StartHealthCheck starts the health-check which checks if the backend
// is alive every second.
func (b *Backend) StartHealthCheck() error {
	// TODO: decide a better duration or use a config file
	b.healthCheckTicker = time.NewTicker(time.Second * 10)
	for {
		select {
		case _ = <-b.healthCheckTicker.C:
			isAlive, err := b.CheckAlive()
			if err != nil {
				return err
			}
			b.IsAlive = isAlive
		}
	}
}

// StopHealthCheck stops the health-check ticker.
func (b *Backend) StopHealthCheck() {
	b.healthCheckTicker.Stop()
}

// CheckAlive checks if the backend is still alive using a TCP connection
// with a timeout of 2 seconds. It returns an error if the TCP connection
// fails.
func (b *Backend) CheckAlive() (bool, error) {
	// println("Checking health", b.Addr)
	b.Lock()
	defer b.Unlock()
	// _tcpURL := b.URL.Host

	// The number of consecutive failed health checks that must occur before
	// declaring the server unhealthy. This number is based on AWS Elastic Load Balancing
	// default value. See here: https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-healthchecks.html
	unhealthyThreshold := 2
	// The number of consecutive successful health checks that must occur before
	// declaring the server healthy. This number is based on AWS Elastic Load Balancing
	// default value. See here: https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-healthchecks.html
	healthyThreshold := 10

	consecutiveSuccessfulHealthChecks := 0
	consecutiveFailedHealthChecks := 0
	var closeError error
	for consecutiveFailedHealthChecks < unhealthyThreshold && consecutiveSuccessfulHealthChecks < healthyThreshold {
		_, err := http.Get(b.Addr)
		if err != nil {
			consecutiveFailedHealthChecks++
			consecutiveSuccessfulHealthChecks = 0
			continue
		}

		consecutiveSuccessfulHealthChecks++
		consecutiveFailedHealthChecks = 0
	}

	if consecutiveFailedHealthChecks >= unhealthyThreshold {
		return false, closeError
	}
	return true, closeError
}
