package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"github.com/tty-monkey/load-balancer/internal/models"
)

func TestBalancerAddServer(t *testing.T) {
	pool := NewServerPool()
	balancer := NewBalancer(pool)

	serverURL, _ := url.Parse("http://example.com")
	server := &models.Server{
		URL:   serverURL,
		Alive: true,
		Proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}

	balancer.AddServer(server)

	if got := pool.GetNextServer(); got == nil || got.URL.String() != server.URL.String() {
		t.Errorf("Server not added correctly to the pool")
	}
}

func TestBalancerLoadBalance(t *testing.T) {
	pool := NewServerPool()
	balancer := NewBalancer(pool)

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
		Alive: true,
		Proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}

	balancer.AddServer(server)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	balancer.LoadBalance(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Result().StatusCode)
	}
}

func TestBalancerHandleError(t *testing.T) {
	pool := NewServerPool()
	balancer := NewBalancer(pool)

	serverURL, _ := url.Parse("http://example.com")
	proxy := httputil.NewSingleHostReverseProxy(serverURL)
	server := &models.Server{
		URL:   serverURL,
		Alive: true,
		Proxy: proxy,
	}
	pool.AddServer(server)

	errorHandler := balancer.HandleError(serverURL, proxy)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), Retry, 3)
	errorHandler(w, req.WithContext(ctx), fmt.Errorf("test error"))

	if w.Result().StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status ServiceUnavailable, got %v", w.Result().StatusCode)
	}
}

func TestBalancerHealthCheck(t *testing.T) {
	pool := NewServerPool()
	balancer := NewBalancer(pool)

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

	pool.AddServer(server)
	balancer.HealthCheck()

	time.Sleep(1 * time.Second)

	if !server.IsAlive() {
		t.Errorf("Expected server to be set alive by HealthCheck")
	}
}
