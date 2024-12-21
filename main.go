package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/vinit-chauhan/reverse-proxy/config"
	"github.com/vinit-chauhan/reverse-proxy/internal"
	"github.com/vinit-chauhan/reverse-proxy/logger"
)

func init() {
	logger.Init()
	logger.Debug("init", "logger initialized")

	logger.Debug("init", "start loading config")
	config.Load()
	logger.Info("init", "config loaded successfully")
}

func main() {
	conf := config.GetConfig()

	fmt.Println("[debug] setting up load balancer")
	loadBalancer := internal.NewLoadBalancer(&conf)
	fmt.Println("[debug] load balancer initiated")

	fmt.Println("[debug] setting up multiple routes")
	handler := http.NewServeMux()

	for _, service := range conf.Services {
		path := service.UrlPath
		if path == "" {
			log.Fatalf("Service URL path cannot be empty")
		}

		handler.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Load balancing incoming requests")
			proxy := loadBalancer.GetServices(path).GetNextBackend()
			proxy.ServeHTTP(w, r)
		}))
	}

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: handler,
	}

	log.Println("Starting reverse proxy with multiple backends on https://0.0.0.0:8080...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
