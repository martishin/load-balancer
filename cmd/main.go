package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/tty-monkey/load-balancer/internal/models"
	"github.com/tty-monkey/load-balancer/internal/service"
)

func main() {
	app := application{
		serverPool: service.NewServerPool(),
	}

	var serverList string
	var port int

	flag.StringVar(&serverList, "backends", "", "Load balanced backends, use commas to separate")
	flag.IntVar(&port, "port", 80, "Port to serve")
	flag.Parse()

	serverUrls := strings.Split(serverList, ",")

	for _, urlString := range serverUrls {
		serverUrl, err := url.Parse(urlString)
		if err != nil {
			log.Fatal(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = app.handleError(serverUrl, proxy)

		app.serverPool.AddServer(
			&models.Server{
				URL:   serverUrl,
				Alive: true,
				Proxy: proxy,
			},
		)
		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(app.loadBalance),
	}

	go app.healthCheck()

	log.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
