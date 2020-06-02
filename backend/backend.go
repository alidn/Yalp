package backend

import (
	"errors"
	"net"
	"net/url"
	"time"
)

// Backend represents a backend server.
type Backend struct {
	Addr string
	// shows whether or not the server is alive.
	IsAlive           bool
	healthCheckTicker *time.Ticker
}

func NewBackend(addr string) (*Backend, error) {
	backend := &Backend{
		Addr: addr,
		// TODO: having a default value might cause errors
		IsAlive: true,
	}
	isAlive, err := backend.CheckAlive()
	if err != nil {
		backend.IsAlive = false
	} else {
		backend.IsAlive = isAlive
	}

	// TODO: find out why a new goroutine is absolutely necessary here.
	go func() {
		err = backend.StartHealthCheck()
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
	parsedURL, err := url.Parse(b.Addr)
	if err != nil {
		return false, errors.New("could not parser url: %s: %s" + b.Addr + " : " + err.Error())
	}
	tcpURL := parsedURL.Host

	// TODO: find a better timeout, or maybe use a config file
	conn, err := net.DialTimeout("tcp", tcpURL, time.Second)
	if err != nil {
		return false, nil
		// return false, errors.New("could not do a tcp dial: " + err.Error())
	}

	err = conn.Close()
	if err != nil {
		return true, err
	}

	return true, nil
}
