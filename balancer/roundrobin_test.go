package balancer

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"
)

func CreateTestServer(id int, logs *[]int) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		*logs = append(*logs, id)
	}
	return httptest.NewServer(http.HandlerFunc(handler))
}

func AssertInRange(t *testing.T, actual int, lowerbound int, upperbound int, message string) {
	t.Helper()

	if actual < lowerbound || actual > upperbound {
		t.Errorf("%s, received = %d, lowerbound = %d, upperbound = %d", message, actual, lowerbound, upperbound)
	}
}

func GetClient(config Config, urls ...string) *httptest.Server {
	loadBalancer, err := NewRoundRobinBalancerWithURLs(urls...)

	if err != nil {
		log.Fatal("Could not get the load balancer", err)
	}

	loadBalancer.Config = config

	reverseProxy := loadBalancer.NewReverseProxy()

	return httptest.NewServer(reverseProxy)
}

func MakeRequests(count int, url string) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal("Couldn't create the cookie jar")
	}
	client := http.Client{Jar: jar}
	for i := 0; i < count; i++ {
		client.Get(url)
	}
}

func CountOccurrences(list []int, target int) int {
	count := 0
	for _, x := range list {
		if target == x {
			count++
		}
	}
	return count
}

func TestOneServer(t *testing.T) {
	logs := make([]int, 0)
	testServer := CreateTestServer(1, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          false,
			ExpirationPeriod: 0,
		},
	}
	client := GetClient(config, testServer.URL)
	defer client.Close()

	MakeRequests(1000, client.URL)

	AssertInRange(t, len(logs), 950, 1050, "expected the server to receive ~1000 requests")

	for _, log := range logs {
		if log != 1 {
			t.Errorf("expected the log to be 1, found %d", log)
		}
	}
}

func TestOneServerHighVolume(t *testing.T) {
	logs := make([]int, 0)
	testServer := CreateTestServer(1, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          false,
			ExpirationPeriod: 0,
		},
	}
	client := GetClient(config, testServer.URL)
	defer client.Close()

	MakeRequests(10000, client.URL)

	AssertInRange(t, len(logs), 9900, 10100, "expected the server to receive ~10000 requests")

	for _, log := range logs {
		if log != 1 {
			t.Errorf("expected the log to be 1, found %d", log)
		}
	}
}

func TestTwoServersNoSession(t *testing.T) {
	logs := make([]int, 0)
	testServer1 := CreateTestServer(1, &logs)
	testServer2 := CreateTestServer(2, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          false,
			ExpirationPeriod: 0,
		},
	}

	client := GetClient(config, testServer1.URL, testServer2.URL)
	defer client.Close()

	MakeRequests(1000, client.URL)

	AssertInRange(t, len(logs), 950, 1050, "expected the servers to receive ~1000 requests")

	firstServerN := CountOccurrences(logs, 1)
	secondServerN := CountOccurrences(logs, 2)

	AssertInRange(t, firstServerN, 450, 550, "expected the server 1 to receive ~500 requests")
	AssertInRange(t, secondServerN, 450, 550, "expected the server 2 to receive ~500 requests")
}

func TestTwoServersNoSessionHighVolume(t *testing.T) {
	logs := make([]int, 0)
	testServer1 := CreateTestServer(1, &logs)
	testServer2 := CreateTestServer(2, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          false,
			ExpirationPeriod: 0,
		},
	}

	client := GetClient(config, testServer1.URL, testServer2.URL)
	defer client.Close()

	MakeRequests(10000, client.URL)

	AssertInRange(t, len(logs), 9900, 10100, "expected the servers to receive ~1000 requests")

	firstServerN := CountOccurrences(logs, 1)
	secondServerN := CountOccurrences(logs, 2)

	AssertInRange(t, firstServerN, 4950, 5050, "expected the server 1 to receive ~500 requests")
	AssertInRange(t, secondServerN, 4950, 5050, "expected the server 2 to receive ~500 requests")
}

func TestTwoServersWithSessionPersistence(t *testing.T) {
	logs := make([]int, 0)
	testServer1 := CreateTestServer(1, &logs)
	testServer2 := CreateTestServer(2, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          true,
			ExpirationPeriod: 3,
		},
	}

	client := GetClient(config, testServer1.URL, testServer2.URL)
	defer client.Close()

	MakeRequests(100, client.URL)

	AssertInRange(t, len(logs), 100, 120, "expected the servers to receive ~100 requests")

	firstServerN := CountOccurrences(logs, 1)
	secondServerN := CountOccurrences(logs, 2)

	AssertInRange(t, firstServerN, 100, 120, "expected the server 1 to receive ~100 requests")
	AssertInRange(t, secondServerN, 0, 10, "expected the server 2 to receive ~0 requests")
}

func TestTwoServersWithSessionPersistence2(t *testing.T) {
	logs := make([]int, 0)
	testServer1 := CreateTestServer(1, &logs)
	testServer2 := CreateTestServer(2, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          true,
			ExpirationPeriod: 3,
		},
	}

	client := GetClient(config, testServer1.URL, testServer2.URL)
	defer client.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal("Couldn't create the cookie jar")
	}
	c := http.Client{Jar: jar}
	for i := 0; i < 3000; i++ {
		c.Get(client.URL)
	}
	time.Sleep(time.Second * 3)
	for i := 0; i < 1000; i++ {
		c.Get(client.URL)
	}

	AssertInRange(t, len(logs), 3950, 4050, "expected the servers to receive ~4000 requests")

	firstServerN := CountOccurrences(logs, 1)
	secondServerN := CountOccurrences(logs, 2)

	AssertInRange(t, firstServerN, 2950, 3050, "expected the server 1 to receive ~3000 requests")
	AssertInRange(t, secondServerN, 950, 1050, "expected the server 2 to receive ~1000 requests")
}

func TestThreeServersNoSession(t *testing.T) {
	logs := make([]int, 0)
	testServer1 := CreateTestServer(1, &logs)
	testServer2 := CreateTestServer(2, &logs)
	testServer3 := CreateTestServer(3, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          false,
			ExpirationPeriod: 0,
		},
	}

	client := GetClient(config, testServer1.URL, testServer2.URL, testServer3.URL)
	defer client.Close()

	MakeRequests(1000, client.URL)

	AssertInRange(t, len(logs), 950, 1050, "expected the servers to receive ~1000 requests")

	firstServerN := CountOccurrences(logs, 1)
	secondServerN := CountOccurrences(logs, 2)
	thirdServerN := CountOccurrences(logs, 3)

	AssertInRange(t, firstServerN, 250, 350, "expected the server 1 to receive ~500 requests")
	AssertInRange(t, secondServerN, 250, 350, "expected the server 2 to receive ~500 requests")
	AssertInRange(t, thirdServerN, 250, 350, "expected the server 3 to receive ~500 requests")
}

func TestThreeServersNoSessionHighVolume(t *testing.T) {
	logs := make([]int, 0)
	testServer1 := CreateTestServer(1, &logs)
	testServer2 := CreateTestServer(2, &logs)
	testServer3 := CreateTestServer(3, &logs)

	config := Config{
		Algorithm: "round-robin",
		SessionPersistenceConfig: SessionPersistenceConfig{
			Enabled:          false,
			ExpirationPeriod: 0,
		},
	}

	client := GetClient(config, testServer1.URL, testServer2.URL, testServer3.URL)
	defer client.Close()

	MakeRequests(10000, client.URL)

	AssertInRange(t, len(logs), 9950, 10100, "expected the servers to receive ~1000 requests")

	firstServerN := CountOccurrences(logs, 1)
	secondServerN := CountOccurrences(logs, 2)
	thirdServerN := CountOccurrences(logs, 3)

	AssertInRange(t, firstServerN, 3300, 3360, "expected the server 1 to receive ~500 requests")
	AssertInRange(t, secondServerN, 3300, 3360, "expected the server 2 to receive ~500 requests")
	AssertInRange(t, thirdServerN, 3000, 3360, "expected the server 3 to receive ~500 requests")
}
