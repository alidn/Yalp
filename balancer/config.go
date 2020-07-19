package balancer

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Algorithm string

const (
	RoundRobin         Algorithm = "round-robin"
	WeightedRoundRobin Algorithm = "weighted-round-robin"
)

type SessionPersistenceConfig struct {
	Enabled bool
	// the cookie expiration time in seconds.
	ExpirationPeriod int32 `yaml:"expiration_period"`
}

type Config struct {
	Algorithm                Algorithm
	SessionPersistenceConfig SessionPersistenceConfig `yaml:"session_persistence"`
}

func ReadConfigFile(filename string) (Config, error) {
	config := Config{}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		// TODO: return default Config
		return config, err
	}

	err = yaml.Unmarshal(file, &config)

	return config, err
}
