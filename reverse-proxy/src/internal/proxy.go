package internal

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/vinit-chauhan/reverse-proxy/config"
	"github.com/vinit-chauhan/reverse-proxy/logger"
)

type Path string

type LoadBalancer struct {
	Services map[Path]Service
}

func NewLoadBalancer(conf *config.ConfigType) *LoadBalancer {
	fmt.Println("[debug] in the function NewLoadBalancer")

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
		services[Path(service.UrlPath)] = Service{backends: backends}
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
