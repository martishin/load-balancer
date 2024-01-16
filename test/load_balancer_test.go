package test

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"github.com/tty-monkey/load-balancer/internal/models"
	"github.com/tty-monkey/load-balancer/internal/service"
)

func TestLoadBalancerIntegration(t *testing.T) {
	pool := service.NewServerPool()
	balancer := service.NewBalancer(pool)

	numServers := 3
	servers := make([]*httptest.Server, numServers)
	for i := 0; i < numServers; i++ {
		servers[i] = createMockServer()
		defer servers[i].Close()

		serverURL, _ := url.Parse(servers[i].URL)
		pool.AddServer(
			&models.Server{
				URL:   serverURL,
				Alive: true,
				Proxy: httputil.NewSingleHostReverseProxy(serverURL),
			},
		)
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	rr := httptest.NewRecorder()
	balancer.LoadBalance(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rr.Result().StatusCode)
	}

	for _, server := range servers {
		u, _ := url.Parse(server.URL)
		pool.MarkServerStatus(u, false)
	}
	rr = httptest.NewRecorder()
	balancer.LoadBalance(rr, req)

	if rr.Result().StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status ServiceUnavailable, got %v", rr.Result().StatusCode)
	}

	balancer.HealthCheck()
	time.Sleep(1 * time.Second)

	for _, server := range servers {
		u, _ := url.Parse(server.URL)
		if !pool.IsServerAlive(u) {
			t.Errorf("Expected server %s to be alive after health check", server.URL)
		}
	}
}

func createMockServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
}
