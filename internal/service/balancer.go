package service

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/tty-monkey/load-balancer/internal/models"
)

const (
	Attempts int = iota
	Retry
)

type Balancer interface {
	AddServer(server *models.Server)
	LoadBalance(w http.ResponseWriter, r *http.Request)
	HandleError(serverUrl *url.URL, proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request, error)
	HealthCheck()
}

type DefaultBalancer struct {
	serverPool ServerPool
}

func NewBalancer(pool ServerPool) Balancer {
	return &DefaultBalancer{
		serverPool: pool,
	}
}

func getAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func getRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func (b *DefaultBalancer) AddServer(server *models.Server) {
	b.serverPool.AddServer(server)
}

func (b *DefaultBalancer) LoadBalance(w http.ResponseWriter, r *http.Request) {
	attempts := getAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	server := b.serverPool.GetNextServer()
	if server != nil {
		server.Proxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func (b *DefaultBalancer) HandleError(serverUrl *url.URL, proxy *httputil.ReverseProxy) func(
	http.ResponseWriter, *http.Request, error,
) {
	return func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
		retries := getRetryFromContext(request)
		if retries < 3 {
			select {
			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				proxy.ServeHTTP(writer, request.WithContext(ctx))
			}
			return
		}

		b.serverPool.MarkServerStatus(serverUrl, false)

		attempts := getAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		b.LoadBalance(writer, request.WithContext(ctx))
	}
}

func (b *DefaultBalancer) HealthCheck() {
	log.Println("Starting health check...")
	b.serverPool.HealthCheck()
	log.Println("Health check completed")
}
