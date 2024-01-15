package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/tty-monkey/load-balancer/internal/models"
	"github.com/tty-monkey/load-balancer/internal/service"
)

func main() {
	var serverList string
	var port int

	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 80, "Port to serve")
	flag.Parse()

	serverUrls := strings.Split(serverList, ",")

	balancer, err := setupLoadBalancer(serverUrls)
	if err != nil {
		log.Fatal(err)
	}

	if err := startServer(balancer, port); err != nil {
		log.Fatal(err)
	}

	log.Printf("Load Balancer started at :%d\n", port)
}

func setupLoadBalancer(serverUrls []string) (service.Balancer, error) {
	balancer := service.NewBalancer(service.NewServerPool())

	for _, urlString := range serverUrls {
		serverUrl, err := url.Parse(urlString)
		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = balancer.HandleError(serverUrl, proxy)

		balancer.AddServer(
			&models.Server{
				URL:   serverUrl,
				Alive: true,
				Proxy: proxy,
			},
		)
	}

	return balancer, nil
}

func startServer(balancer service.Balancer, port int) error {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(balancer.LoadBalance),
	}

	go startHealthChecks(balancer)

	return server.ListenAndServe()
}

func startHealthChecks(balancer service.Balancer) {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			balancer.HealthCheck()
		}
	}
}
