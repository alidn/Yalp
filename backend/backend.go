package backend

import (
	"errors"
	"net"
	"net/url"
	"sync"
	"time"
)

// Backend represents a backend server.
type Backend struct {
	// TODO: consider using net.URL instead of raw string.
	Addr string
	// shows whether or not the server is alive.
	IsAlive           bool
	healthCheckTicker *time.Ticker
	sync.RWMutex
}

func NewBackend(addr string) (*Backend, error) {
	backend := &Backend{
		Addr:    addr,
		IsAlive: true,
	}
	go func() {
		isAlive, err := backend.CheckAlive()
		if err != nil {
			backend.IsAlive = false
		} else {
			backend.IsAlive = isAlive
		}
	}()

	var err error
	// TODO: find out why a new goroutine is absolutely necessary here.
	go func() {
		err2 := backend.StartHealthCheck()
		err = err2
	}()
	if err != nil {
		return nil, err
	}
	return backend, nil
}

// StartHealthCheck starts the healt-check which checks if the backend
// is alive every second.
func (b *Backend) StartHealthCheck() error {
	// TODO: decide a better duration or use a config file
	b.healthCheckTicker = time.NewTicker(time.Second)
	for {
		select {
		case _ = <-b.healthCheckTicker.C:
			isAlive, err := b.CheckAlive()
			if err != nil {
				return err
			}
			// TODO: should lock the backend?
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
	parsedURL, err := url.Parse(b.Addr)
	if err != nil {
		return false, errors.New("could not parser url: %s: %s" + b.Addr + " : " + err.Error())
	}
	tcpURL := parsedURL.Host

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
		// the response timeout of 5 seconds is based on AWS Elastic Load Balancing
		// default value. See here: https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-healthchecks.html
		conn, err := net.DialTimeout("tcp", tcpURL, time.Second*5)
		if err != nil {
			consecutiveFailedHealthChecks++
			consecutiveSuccessfulHealthChecks = 0
			// println("In health check", b.Addr, "Unsuccessful")
			continue
		}

		// println("In health check", b.Addr, "successful")

		err = conn.Close()
		if err != nil {
			closeError = err
			consecutiveSuccessfulHealthChecks++
			consecutiveFailedHealthChecks = 0
		}

		consecutiveSuccessfulHealthChecks++
		consecutiveFailedHealthChecks = 0
	}

	if consecutiveFailedHealthChecks >= unhealthyThreshold {
		return false, closeError
	}
	return true, closeError
}
