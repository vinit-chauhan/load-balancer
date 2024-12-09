package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

// Backend struct holds backend URL information
type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
}

// LoadBalancer struct manages multiple backends
type LoadBalancer struct {
	backends []*Backend
	counter  uint64
}

// NewLoadBalancer initializes the load balancer with backends
func NewLoadBalancer(targets []string) *LoadBalancer {
	backends := make([]*Backend, len(targets))
	for i, target := range targets {
		url, err := url.Parse(target)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}
		backends[i] = &Backend{
			URL:          url,
			ReverseProxy: httputil.NewSingleHostReverseProxy(url),
		}
		fmt.Printf("Backend URL: %s\n", backends[i].URL)
	}

	return &LoadBalancer{backends: backends}
}

// GetNextBackend returns the next backend using round-robin
func (lb *LoadBalancer) GetNextBackend() *httputil.ReverseProxy {
	index := atomic.AddUint64(&lb.counter, 1) % uint64(len(lb.backends))
	return lb.backends[index].ReverseProxy
}

func main() {
	// Define multiple target servers
	targets := []string{
		"http://backend1.local",
		"http://backend2.local",
		"http://backend3.local",
	}

	// Create a load balancer
	loadBalancer := NewLoadBalancer(targets)

	// Define an HTTPS server with load balancer
	server := &http.Server{
		Addr: "0.0.0.0:80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy := loadBalancer.GetNextBackend()
			proxy.ServeHTTP(w, r)
		}),
	}

	log.Println("Starting encrypted reverse proxy with multiple backends on https://0.0.0.0:80...")
	// Start the HTTPS server
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
