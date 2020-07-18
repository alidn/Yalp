package balancer

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Algorithm string

const (
	RoundRobin         = "round-robin"
	WeightedRoundRobin = "weighted-round-robin"
)

type Config struct {
	Algorithm          Algorithm
	SessionPersistence bool
}

func ReadConfigFile(filename string) (Config, error) {
	config := Config{}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		// TODO: return default config
		return config, err
	}

	err = yaml.Unmarshal(file, &config)

	return config, err
}
