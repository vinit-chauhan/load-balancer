package internal

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/vinit-chauhan/load-balancer/config"
	"github.com/vinit-chauhan/load-balancer/logger"
)

type Path string

type LoadBalancer struct {
	Services map[Path]*Service
	mux      sync.RWMutex
}

// UpdateServices updates the services in a thread-safe manner.
func (lb *LoadBalancer) UpdateServices(conf *config.ConfigType) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	logger.Debug("UpdateServices", "updating load balancer services from new config")
	newServices := make(map[Path]*Service)

	for _, serviceConf := range conf.Services {
		serviceConf.Validate() // Validate service configuration
		backends := make([]*Backend, len(serviceConf.Backends))
		for i, backendURL := range serviceConf.Backends {
			u, err := url.Parse(backendURL)
			if err != nil {
				logger.Error("UpdateServices", "error parsing url", "url", backendURL, "error", err)
				continue
			}
			proxy := httputil.NewSingleHostReverseProxy(u)
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
				logger.ErrorContext(r.Context(), "Proxy error", "backend", u.String(), "error", e.Error())
				w.WriteHeader(http.StatusBadGateway)
			}
			backends[i] = &Backend{
				URL:          u,
				ReverseProxy: proxy,
				Alive:        true,
			}
		}

		svc := &Service{
			Name:        serviceConf.Name,
			Backends:    backends,
			Algorithm:   serviceConf.Algorithm,
			HealthCheck: serviceConf.HealthCheck,
		}
		svc.StartHealthCheck()
		svc.UpdateHashRing()

		newServices[Path(serviceConf.UrlPath)] = svc
	}
	lb.Services = newServices
}

func NewLoadBalancer(conf *config.ConfigType) *LoadBalancer {
	logger.Debug("NewLoadBalancer", "creating new load balancer instance from config")

	services := make(map[Path]*Service)

	for _, serviceConf := range conf.Services {
		serviceConf.Validate() // Validate service configuration
		backends := make([]*Backend, len(serviceConf.Backends))
		for i, backendURL := range serviceConf.Backends {
			u, err := url.Parse(backendURL)
			if err != nil {
				logger.Error("NewLoadBalancer", "error parsing url", "url", backendURL, "error", err)
				continue
			}
			proxy := httputil.NewSingleHostReverseProxy(u)

			// Custom Error Handler for Passive Health Check
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
				logger.ErrorContext(r.Context(), "Proxy error", "backend", u.String(), "error", e.Error())
				w.WriteHeader(http.StatusBadGateway)
			}

			backends[i] = &Backend{
				URL:          u,
				ReverseProxy: proxy,
				Alive:        true,
			}
		}

		svc := &Service{
			Name:        serviceConf.Name,
			Backends:    backends,
			Algorithm:   serviceConf.Algorithm,
			HealthCheck: serviceConf.HealthCheck,
		}
		
		// Initialize Health Check & Hash Ring
		svc.StartHealthCheck()
		svc.UpdateHashRing()

		services[Path(serviceConf.UrlPath)] = svc
	}

	return &LoadBalancer{Services: services}
}

func (lb *LoadBalancer) GetServices(path string) *Service {
	lb.mux.RLock()
	defer lb.mux.RUnlock()
	service, exists := lb.Services[Path(path)]
	if !exists {
		return nil
	}
	return service
}
