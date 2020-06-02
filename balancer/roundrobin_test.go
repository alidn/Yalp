package balancer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNextBackend(t *testing.T) {
	testServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("Got a request for server 1")
	}))

	testServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("Got a request for server 1")
	}))

	testServer3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("Got a request for server 1")
	}))

	t.Logf(testServer1.URL, testServer2.URL, testServer3.URL)

	deadServerURL1 := "http://127.0.0.1:00000"
	deadServerURL2 := "http://127.0.0.1:00001"

	roundRobinBalancer, err := NewRoundRobinBalancerWithURLs(testServer1.URL, deadServerURL1, testServer2.URL, deadServerURL2, testServer3.URL)
	if err != nil {
		t.Error(err)
	}

	nextBackend, err := roundRobinBalancer.NextBackend()
	if err != nil {
		t.Error("ERR!", err)
	}
	if roundRobinBalancer.GetCurIndex() != 0 {
		t.Errorf("index wrong, got %d, expected %d", roundRobinBalancer.GetCurIndex(), 0)
	}
	if nextBackend.Addr != testServer1.URL {
		t.Errorf("first server url is wrong, got %s, expected %s", nextBackend.Addr, testServer1.URL)
	}

	nextBackend, err = roundRobinBalancer.NextBackend()
	if err != nil {
		t.Error("ERR!", err)
	}
	if nextBackend.Addr != testServer2.URL {
		t.Errorf("second server url is wrong, got %s, expected %s", nextBackend.Addr, testServer2.URL)
	}

	testServer3.Close()
	nextBackend, err = roundRobinBalancer.NextBackend()
	if err != nil {
		t.Error("ERR!", err)
	}
	if nextBackend.Addr != testServer1.URL {
		t.Errorf("third server url is wrong, got %s, expected %s", nextBackend.Addr, testServer1.URL)
	}
}
