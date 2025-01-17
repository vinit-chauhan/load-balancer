package internal

import (
	"net/http/httputil"
	"net/url"

	"github.com/vinit-chauhan/load-balancer/config"
	"github.com/vinit-chauhan/load-balancer/logger"
)

type Path string

type LoadBalancer struct {
	Services map[Path]Service
}

func NewLoadBalancer(conf *config.ConfigType) *LoadBalancer {
	logger.Debug("NewLoadBalancer", "creating new load balancer instance from config")

	services := make(map[Path]Service)

	for _, service := range conf.Services {
		backends := make([]*Backend, len(service.Backends))
		for i, backend := range service.Backends {
			url, err := url.Parse(backend)
			if err != nil {
				logger.Error("NewLoadBalancer", "error parsing url:"+backend+":"+err.Error())
			}
			backends[i] = &Backend{
				URL:          url,
				ReverseProxy: httputil.NewSingleHostReverseProxy(url),
			}
		}
		services[Path(service.UrlPath)] = Service{backends: backends, counter: new(uint64)}
	}

	return &LoadBalancer{Services: services}
}

func (lb *LoadBalancer) GetServices(path string) *Service {
	service, exists := lb.Services[Path(path)]
	if !exists {
		return nil
	}

	return &service
}
