package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/alidn/Yalp/balancer"
)

func main() {
	config, err := balancer.ReadConfigFile("config.yaml")
	if err != nil {
		panic("Could not find the config file")
	}

	urls := config.URLs
	var loadBalancer balancer.Balancer
	if config.Algorithm == balancer.RoundRobin {
		loadBalancer, err = balancer.NewRoundRobinBalancerWithURLs(urls...)
		if err != nil {
			log.Fatal("could not start the round-robin balancer: ", err)
		}
	} else if config.Algorithm == balancer.LeastConnection {
		loadBalancer, err = balancer.NewRoundRobinBalancerWithURLs(urls...)
		if err != nil {
			log.Fatal("could not start the least-connection balancer: ", err)
		}
	}

	reverseProxy := loadBalancer.NewReverseProxy()
	println("Server listening on port 9000")

	err = http.ListenAndServe(":9000", reverseProxy)
	if err != nil {
		log.Fatal("ERROR, could not start the server", err)
	}
}

func example() {
	config, err := balancer.ReadConfigFile("config.yaml")

	if err != nil {
		log.Fatal("Couldn't read the config file", err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hi there")
	}

	handler2 := func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hi there2")
	}

	handler3 := func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hi there3")
	}

	testServer := httptest.NewServer(http.HandlerFunc(handler))
	testServer2 := httptest.NewServer(http.HandlerFunc(handler2))
	testServer3 := httptest.NewServer(http.HandlerFunc(handler3))

	loadBalancer, err := balancer.NewRoundRobinBalancerWithURLs(testServer.URL,
		testServer2.URL, testServer3.URL)
	if err != nil {
		log.Fatal(err)
	}
	loadBalancer.Config = config

	reverseProxy := loadBalancer.NewReverseProxy()

	err = http.ListenAndServe(":9000", reverseProxy)
	log.Fatal("Could not start the load loadBalancer reverse proxy")
}
