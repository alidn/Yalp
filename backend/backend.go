package backend

import (
	"net"
	"time"
)

type Backend struct {
	addr string
	IsAlive bool
	healthCheckTicker *time.Ticker
}

func NewBackend(addr string) *Backend {
	return &Backend{
		addr: addr,
		IsAlive: false,
	}
}

func (b Backend) StartHealthCheck() error {
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

func (b Backend) StopHealthCheck() {
	b.healthCheckTicker.Stop()
}

func (b Backend) CheckAlive() (bool, error) {
	// TODO: find a better timeout, or maybe use a config file
	conn, err := net.DialTimeout("tcp", b.addr, time.Second*2)
	if err != nil {
		return false, err
	}

	err = conn.Close()
	if err != nil {
		return true, err
	}

	return true, nil
}
