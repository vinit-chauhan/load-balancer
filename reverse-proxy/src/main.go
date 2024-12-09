package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
}

type LoadBalancer struct {
	backends []*Backend
	counter  uint64
}

func NewLoadBalancer(targets []string) *LoadBalancer {
	fmt.Println("[debug] in the function NewLoadBalancer")
	backends := make([]*Backend, len(targets))
	for i, target := range targets {
		fmt.Printf("[debug] setting up connection with target: %s\n", target)
		url, err := url.Parse(target)
		if err != nil {
			log.Fatalf("Failed to parse target URL: %v", err)
		}
		backends[i] = &Backend{
			URL:          url,
			ReverseProxy: httputil.NewSingleHostReverseProxy(url),
		}
		fmt.Printf("[debug] setup complete with target: %s\n", target)
	}

	return &LoadBalancer{backends: backends}
}

func (lb *LoadBalancer) GetNextBackend() *httputil.ReverseProxy {
	index := atomic.AddUint64(&lb.counter, 1) % uint64(len(lb.backends))
	return lb.backends[index].ReverseProxy
}

func main() {
	targets := []string{
		"http://backend1.local",
		"http://backend2.local",
		"http://backend3.local",
	}

	fmt.Println("[debug] setting up load balancer")
	loadBalancer := NewLoadBalancer(targets)
	fmt.Println("[debug] load balancer initiated")

	fmt.Println("[debug] setting up multiple routes")
	handler := http.NewServeMux()

	handler.Handle("/backend1/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, _ := url.Parse(targets[0])
		fmt.Println("redirecting user to", url.String())
		httputil.NewSingleHostReverseProxy(url).ServeHTTP(w, r)
	}))

	handler.Handle("/welcome", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Load balancing incoming requests")
		proxy := loadBalancer.GetNextBackend()
		proxy.ServeHTTP(w, r)
	}))

	server := &http.Server{
		Addr:    "0.0.0.0:80",
		Handler: handler,
	}

	log.Println("Starting reverse proxy with multiple backends on https://0.0.0.0:80...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
