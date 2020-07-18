package main

import (
	"Yalp/balancer"
	"log"
)

func main() {
	config, err := balancer.ReadConfigFile("config.yaml")
	if err != nil {
		log.Fatal("Couldn't read the config file", err)
	}
	println(config.Algorithm, config.SessionPersistence)
}
