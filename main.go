package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/vinit-chauhan/load-balancer/config"
	"github.com/vinit-chauhan/load-balancer/internal"
	"github.com/vinit-chauhan/load-balancer/logger"
)

func init() {
	logger.Init()
	logger.SetLogLevel(logger.LevelDebug)
	logger.Debug("init", "logger initialized")

	logger.Debug("init", "start loading config")

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		logger.Debug("init", "CONFIG_PATH not set, using default config path")
		path = "./config.yml"
	}
	config.Load(path)
	logger.Info("init", "config loaded successfully")
}

func main() {
	conf := config.GetConfig()

	logger.Debug("main", "setting up load balancer")
	loadBalancer := internal.NewLoadBalancer(&conf)
	logger.Debug("main", "load balancer initiated")

	logger.Debug("main", "setting up multiple routes")
	handler := http.NewServeMux()

	for _, service := range conf.Services {
		path := service.UrlPath
		if path == "" {
			logger.Panic("main", "Service URL path cannot be empty")
			os.Exit(1)
		}

		handler.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug("main", "Load balancing incoming requests")
			proxy := loadBalancer.GetServices(path).GetNextBackend()
			proxy.ServeHTTP(w, r)
		}))
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: handler,
	}

	logger.Info("main", "Starting reverse proxy with multiple backends on https://0.0.0.0:8080...")
	if err := server.ListenAndServe(); err != nil {
		logger.Panic("main", fmt.Sprintf("Server failed: %v", err))
	}
}
