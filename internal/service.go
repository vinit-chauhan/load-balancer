package internal

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Service struct {
	backends []*Backend
	counter  uint64
}

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
}

func (s *Service) GetNextBackend() *httputil.ReverseProxy {
	index := atomic.AddUint64(&s.counter, 1) % uint64(len(s.backends))
	return s.backends[index].ReverseProxy
}
