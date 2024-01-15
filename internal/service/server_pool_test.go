package service

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/tty-monkey/load-balancer/internal/models"
)

// Mock function for isServerAlive to avoid real network calls
func mockIsServerAlive(u *url.URL) bool {
	// Mock implementation, can be customized based on test needs
	return true
}

func setupServerPool() (*DefaultServerPool, *models.Server, *url.URL) {
	serverPool := NewServerPool().(*DefaultServerPool)
	serverURL, _ := url.Parse("http://example.com")
	server := &models.Server{
		URL:   serverURL,
		Alive: true,
		Proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}
	serverPool.AddServer(server)

	return serverPool, server, serverURL
}

func TestPoolAddServer(t *testing.T) {
	serverPool, server, _ := setupServerPool()

	got := serverPool.GetNextServer()
	if got != server {
		t.Errorf("Added server was not found in the server pool")
	}
}

func TestPoolMarkServerStatus(t *testing.T) {
	serverPool, server, serverURL := setupServerPool()

	serverPool.MarkServerStatus(serverURL, false)
	if server.IsAlive() {
		t.Errorf("Server should have been marked as not alive")
	}

	serverPool.MarkServerStatus(serverURL, true)
	if !server.IsAlive() {
		t.Errorf("Server should have been marked as alive")
	}
}

func TestPoolGetNextServer(t *testing.T) {
	serverPool, server, serverURL := setupServerPool()

	got := serverPool.GetNextServer()
	if got != server {
		t.Errorf("Expected GetNextServer to return the server, but it did not")
	}

	serverPool.MarkServerStatus(serverURL, false)
	got = serverPool.GetNextServer()
	if got != nil {
		t.Errorf("Expected GetNextServer to return nil for a non-alive server")
	}
}

func TestPoolHealthCheck(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer testServer.Close()

	serverURL, _ := url.Parse(testServer.URL)
	server := &models.Server{
		URL:   serverURL,
		Alive: false,
		Proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}

	serverPool := NewServerPool()
	serverPool.AddServer(server)

	serverPool.HealthCheck()

	if !server.IsAlive() {
		t.Errorf("Expected HealthCheck to set the server as alive")
	}
}
