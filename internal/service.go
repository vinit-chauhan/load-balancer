package internal

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/vinit-chauhan/reverse-proxy/logger"
)

type Service struct {
	backends []*Backend
	counter  *uint64
}

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
}

func (s *Service) GetNextBackend() *httputil.ReverseProxy {
	logger.Debug("GetNextBackend", "fetching next backend")
	index := atomic.AddUint64(s.counter, 1) % uint64(len(s.backends))
	return s.backends[index].ReverseProxy
}
