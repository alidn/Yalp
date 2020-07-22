package main

import (
	"Yalp/balancer"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

func main() {
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
