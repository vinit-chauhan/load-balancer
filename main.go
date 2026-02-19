package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/vinit-chauhan/load-balancer/config"
	"github.com/vinit-chauhan/load-balancer/internal"
	"github.com/vinit-chauhan/load-balancer/logger"
	"github.com/vinit-chauhan/load-balancer/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var configPath string

func init() {
	logger.Init()
	logger.SetLogLevel(logger.LevelDebug)
	logger.Debug("init", "logger initialized")

	logger.Debug("init", "start loading config")

	configPath = os.Getenv("CONFIG_PATH")
	if configPath == "" {
		logger.Debug("init", "CONFIG_PATH not set, using default config path")
		configPath = "./config.yml"
	}
	config.Load(configPath)
	logger.Info("init", "config loaded successfully")
}

func main() {
	// Initialize Tracer
	shutdown := tracer.InitTracer()
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logger.Error("main", "failed to shutdown tracer: "+err.Error())
		}
	}()

	conf := config.GetConfig()

	logger.Debug("main", "setting up load balancer")
	loadBalancer := internal.NewLoadBalancer(&conf)
	logger.Debug("main", "load balancer initiated")

	// Start watching config file for changes
	go watchConfig(loadBalancer)

	logger.Debug("main", "setting up multiple routes")
	handler := http.NewServeMux()

	// Initialize and expose Prometheus metrics
	internal.InitMetrics()
	handler.Handle("/metrics", promhttp.Handler())

	for _, serviceConf := range conf.Services {
		path := serviceConf.UrlPath
		if path == "" {
			logger.Panic("main", "Service URL path cannot be empty")
		}

		// Closure to capture service
		svc := loadBalancer.GetServices(path)
		
		handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			// Start Trace Span
			ctx := r.Context()
			tr := otel.Tracer("load-balancer")
			ctx, span := tr.Start(ctx, "proxy_request")
			defer span.End()

			// Add attributes
			span.SetAttributes(attribute.String("http.path", r.URL.Path))
			span.SetAttributes(attribute.String("service.name", svc.Name))

			// Inject trace context into headers for backend
			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))
			
			// Pass context with span
			r = r.WithContext(ctx)

			logger.DebugContext(ctx, "Forwarding request", "tag", "Proxy", "path", r.URL.Path, "service", svc.Name)
			
			svc.ServeHTTP(w, r)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: handler,
	}

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("main", "Starting reverse proxy with multiple backends on 0.0.0.0:"+port+"...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Panic("main", "Server failed", "error", err)
		}
	}()

	<-stop
	logger.Info("main", "Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("main", "Server shutdown failed", "error", err)
	} else {
		logger.Info("main", "Server stopped gracefully")
	}
}

// watchConfig watches the config file for changes and reloads the configuration.
func watchConfig(loadBalancer *internal.LoadBalancer) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Panic("watchConfig", "Failed to create watcher", "error", err)
	}
	defer watcher.Close()

	err = watcher.Add(configPath)
	if err != nil {
		logger.Panic("watchConfig", "Failed to add config file to watcher", "path", configPath, "error", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// We only care about Write or Create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				logger.Info("watchConfig", "Config file modified, reloading...", "path", event.Name)
				// Reload the configuration
				config.Load(configPath)
				newConf := config.GetConfig()
				loadBalancer.UpdateServices(&newConf)
				logger.Info("watchConfig", "Config reloaded successfully")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Error("watchConfig", "Watcher error", "error", err)
		}
	}
}
