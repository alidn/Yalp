package balancer

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// The tests should not be run concurrently; use `go test -p 1`.

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

	deadServerURL1 := "http://127.0.0.1:00010"
	deadServerURL2 := "http://127.0.0.1:00011"

	roundRobinBalancer, err := NewRoundRobinBalancerWithURLs(testServer1.URL, deadServerURL1, testServer2.URL, deadServerURL2, testServer3.URL)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 1)

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
	time.Sleep(time.Second * 1)

	nextBackend, err = roundRobinBalancer.NextBackend()
	if err != nil {
		t.Error("ERR!", err)
	}
	if nextBackend.Addr != testServer1.URL {
		t.Errorf("third server url is wrong, got %s, expected %s", nextBackend.Addr, testServer1.URL)
	}
}

func TestNextBackend1(t *testing.T) {
	testServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("Got a request for server 1")
	}))

	testServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("Got a request for server 1")
	}))

	roundRobinBalancer, err := NewRoundRobinBalancerWithURLs(testServer1.URL, testServer2.URL)
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

	testServer1.Close()
	time.Sleep(time.Second * 1)
	nextBackend, err = roundRobinBalancer.NextBackend()
	if err != nil {
		t.Error("ERR!", err)
	}
	if nextBackend.Addr != testServer2.URL {
		t.Errorf("third server url is wrong, got %s, expected %s", nextBackend.Addr, testServer2.URL)
	}
}

func TestServerProxy(t *testing.T) {
	resSeq := make([]int, 0)
	server1 := newTestServer(&resSeq, 1)
	server2 := newTestServer(&resSeq, 2)
	server3 := newTestServer(&resSeq, 3)

	balancer, err := NewRoundRobinBalancerWithURLs(server1.URL, server2.URL, server3.URL)
	if err != nil {
		t.Error("Could not get the balancer", err)
	}

	frontendProxy := httptest.NewServer(balancer.NewReverseProxy())
	defer frontendProxy.Close()

	for i := 0; len(resSeq) < 10; i++ {
		_, err = http.Get(frontendProxy.URL)
		if err != nil {
			t.Error("Could not get the reponse", err)
		}
	}

	expected := []int{1, 2, 3, 1, 2, 3, 1, 2, 3, 1}
	if !reflect.DeepEqual(expected, resSeq) {
		t.Error("The sequence of servers isn't the same")
	}

	server1.Close()
	server2.Close()
	server3.Close()
}

func TestServerProxyWithFailure(t *testing.T) {
	resSeq := make([]int, 0)
	server1 := newTestServer(&resSeq, 1)
	server2 := newTestServer(&resSeq, 2)
	server3 := newTestServer(&resSeq, 3)

	balancer, err := NewRoundRobinBalancerWithURLs(server1.URL, server2.URL, server3.URL)
	if err != nil {
		t.Error("Could not get the balancer", err)
	}

	frontendProxy := httptest.NewServer(balancer.NewReverseProxy())
	defer frontendProxy.Close()

	for i := 0; len(resSeq) < 10; i++ {
		if i == 6 {
			server2.Close()
		}
		_, err = http.Get(frontendProxy.URL)
		if err != nil {
			t.Error("Could not get the reponse", err)
		}
	}

	expected := []int{1, 2, 3, 1, 2, 3, 1, 3, 1, 3}
	if !reflect.DeepEqual(expected, resSeq) {
		t.Errorf("The sequence of servers isn't the same\n, got %#v \n, expected %#v", resSeq, expected)
	}

	server1.Close()
	server3.Close()
}

func TestServerProxyWithFailure2(t *testing.T) {
	resSeq := make([]int, 0)
	server1 := newTestServer(&resSeq, 1)
	server2 := newTestServer(&resSeq, 2)
	server3 := newTestServer(&resSeq, 3)

	balancer, err := NewRoundRobinBalancerWithURLs(server1.URL, server2.URL, server3.URL)
	if err != nil {
		t.Error("Could not get the balancer", err)
	}

	frontendProxy := httptest.NewServer(balancer.NewReverseProxy())
	defer frontendProxy.Close()

	for i := 0; len(resSeq) < 10; i++ {
		if i == 2 {
			server3.Close()
		}
		_, err = http.Get(frontendProxy.URL)
		if err != nil {
			t.Error("Could not get the reponse", err)
		}
	}

	expected := []int{1, 2, 1, 2, 1, 2, 1, 2, 1, 2}
	if !reflect.DeepEqual(expected, resSeq) {
		t.Errorf("The sequence of servers isn't the same\n, got %#v \n, expected %#v", resSeq, expected)
	}

	server1.Close()
	server3.Close()
}

func TestServerProxyWithNoAliveBackend(t *testing.T) {
	balancer, err := NewRoundRobinBalancerWithURLs("http://127.0.0.1:00010")
	if err != nil {
		t.Error("Could not get the balancer", err)
	}

	time.Sleep(2 * time.Second)

	backend, err := balancer.NextBackend()
	if backend != nil {
		t.Errorf("Backend should be nil, got %+v", backend)
	}
	if err.Error() != "none of the servers is alive" {
		t.Errorf("Got the wrong error")
	}
}

func newTestServer(resSeq *[]int, id int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*resSeq = append(*resSeq, id)
	}))
	return server
}
