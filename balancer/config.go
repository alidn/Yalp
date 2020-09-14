package balancer

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Algorithm string

const (
	RoundRobin      Algorithm = "round-robin"
	LeastConnection           = "least-connection"
)

type SessionPersistenceConfig struct {
	Enabled bool `yaml:"enabled"`
	// the cookie expiration time in seconds.
	ExpirationPeriod int32 `yaml:"expiration_period"`
}

type Config struct {
	Algorithm                Algorithm                `yaml:"algorithm"`
	SessionPersistenceConfig SessionPersistenceConfig `yaml:"session_persistence"`
	URLs                     []string                 `yaml:"backend_urls"`
}

func ReadConfigFile(filename string) (Config, error) {
	config := Config{}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(file, &config)

	return config, err
}
