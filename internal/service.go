package internal

import (
	"fmt"
	"hash/crc32"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vinit-chauhan/load-balancer/config"
	"github.com/vinit-chauhan/load-balancer/logger"
)

// Backend represents a single backend server that a service can route requests to.
type Backend struct {
	URL          *url.URL           // The URL of the backend server.
	ReverseProxy *httputil.ReverseProxy // The reverse proxy configured to forward requests to this backend.
	Alive        bool               // Current liveness status of the backend (true if alive, false otherwise).
	mux          sync.RWMutex       // Mutex to protect access to the Alive status.
	ActiveConns  int64              // Atomic counter for active connections, used by least-connections algorithm.
}

// Service represents a load-balanced service with multiple backends and a specific load balancing algorithm.
type Service struct {
	Name        string
	Backends    []*Backend
	counter     uint64 // For Round Robin: atomic counter to keep track of the next backend to use.
	Algorithm   string // The load balancing algorithm to use (e.g., "round-robin", "least-connections", "ip-hash").
	HealthCheck config.HealthCheckConfig // Configuration for active health checks.
	// For Consistent Hashing (ip-hash algorithm):
	hashRing []uint32           // Sorted slice of hash values representing virtual nodes on the consistent hash ring.
	hashMap  map[uint32]*Backend // Maps hash values on the ring to actual backend instances.
	ringMux  sync.RWMutex       // Mutex to protect access to hashRing and hashMap.
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

func (b *Backend) IncConn() {
	atomic.AddInt64(&b.ActiveConns, 1)
}

func (b *Backend) DecConn() {
	atomic.AddInt64(&b.ActiveConns, -1)
}

func (b *Backend) GetActiveConns() int64 {
	return atomic.LoadInt64(&b.ActiveConns)
}

// responseWriter is a wrapper around http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.IncConn()
	ActiveConnections.WithLabelValues("", b.URL.String()).Inc()
	defer func() {
		b.DecConn()
		ActiveConnections.WithLabelValues("", b.URL.String()).Dec()
	}()
	b.ReverseProxy.ServeHTTP(w, r)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use a custom ResponseWriter to capture the status code
	rw := newResponseWriter(w)
	
	// Start timer for request duration metric
	start := time.Now()

	backend := s.GetNextBackend(r)
	if backend != nil {
		backend.ServeHTTP(rw, r)
		// Record metrics after the request has been served by the backend
		statusCode := strconv.Itoa(rw.statusCode)
		HttpRequestsTotal.WithLabelValues(s.Name, r.URL.Path, r.Method, statusCode).Inc()
		HttpRequestDurationSeconds.WithLabelValues(s.Name, r.URL.Path, r.Method, statusCode).Observe(time.Since(start).Seconds())
		return
	}

	http.Error(rw, "Service Unavailable", http.StatusServiceUnavailable)
	statusCode := strconv.Itoa(rw.statusCode)
	HttpRequestsTotal.WithLabelValues(s.Name, r.URL.Path, r.Method, statusCode).Inc()
	HttpRequestDurationSeconds.WithLabelValues(s.Name, r.URL.Path, r.Method, statusCode).Observe(time.Since(start).Seconds())
}

// GetNextBackend selects the next available backend based on the configured load balancing algorithm.
// It takes an http.Request as input, which might be used by certain algorithms (e.g., IP Hash).
func (s *Service) GetNextBackend(r *http.Request) *Backend {
	switch s.Algorithm {
	case "least-connections":
		return s.leastConnections()
	case "ip-hash":
		return s.ipHash(r)
	case "round-robin":
		fallthrough
	default:
		return s.roundRobin()
	}
}

// roundRobin implements the Round Robin load balancing algorithm.
// It atomically increments a counter and cycles through the backends to select the next alive one.
func (s *Service) roundRobin() *Backend {
	count := len(s.Backends)
	if count == 0 {
		return nil
	}

	start := atomic.AddUint64(&s.counter, 1)
	// Iterate through backends starting from 'start' to find an alive one.
	// This ensures that even if some backends are down, the load balancer attempts to find an available one.
	for i := 0; i < count; i++ {
		idx := (int(start) + i) % count
		if s.Backends[idx].IsAlive() {
			return s.Backends[idx]
		}
	}
	return nil // No alive backend found
}

// leastConnections implements the Least Connections load balancing algorithm.
// It selects the backend with the fewest active connections among the alive backends.
func (s *Service) leastConnections() *Backend {
	var best *Backend
	min := int64(-1) // Initialize with -1 to ensure the first alive backend is chosen.

	for _, b := range s.Backends {
		if !b.IsAlive() {
			continue // Skip dead backends
		}
		conns := b.GetActiveConns()
		if min == -1 || conns < min {
			min = conns
			best = b
		}
	}
	return best
}

// ipHash implements the IP Hash (Consistent Hashing) load balancing algorithm.
// It uses the client's IP address to consistently route requests to the same backend.
// If the hash ring is empty, it falls back to round-robin.
func (s *Service) ipHash(r *http.Request) *Backend {
	s.ringMux.RLock() // Protect hash ring access
	defer s.ringMux.RUnlock()

	if len(s.hashRing) == 0 {
		return s.roundRobin() // Fallback if hash ring is not initialized or empty
	}

	ip := r.RemoteAddr
	hash := crc32.ChecksumIEEE([]byte(ip)) // Compute hash of the client's IP

	// Find the backend on the hash ring (closest clockwise virtual node)
	idx := sort.Search(len(s.hashRing), func(i int) bool {
		return s.hashRing[i] >= hash
	})

	if idx == len(s.hashRing) {
		idx = 0 // Wrap around to the beginning of the ring
	}

	return s.hashMap[s.hashRing[idx]]
}

// UpdateHashRing builds or updates the consistent hash ring based on the currently alive backends.
// It creates multiple virtual nodes for each alive backend to improve distribution.
func (s *Service) UpdateHashRing() {
	s.ringMux.Lock() // Protect hash ring modification
	defer s.ringMux.Unlock()

	s.hashRing = []uint32{}
	s.hashMap = make(map[uint32]*Backend)

	// Add alive backends to the hash ring with multiple virtual nodes.
	for _, b := range s.Backends {
		if b.IsAlive() {
			for i := 0; i < 3; i++ {
				key := fmt.Sprintf("%s-%d", b.URL.String(), i) // Create unique key for virtual node
				hash := crc32.ChecksumIEEE([]byte(key))    // Compute hash for the virtual node
				s.hashRing = append(s.hashRing, hash)
				s.hashMap[hash] = b
			}
		}
	}
	// Sort the hash ring to enable binary search for backend selection.
	sort.Slice(s.hashRing, func(i, j int) bool {
		return s.hashRing[i] < s.hashRing[j]
	})
}

// StartHealthCheck initializes and runs periodic health checks for the backends of a service.
// If health checks are disabled, all backends are initially marked as alive.
func (s *Service) StartHealthCheck() {
	if !s.HealthCheck.Enabled {
		// If health checks are disabled, mark all backends as alive and update the hash ring.
		for _, b := range s.Backends {
			b.SetAlive(true)
		}
		s.UpdateHashRing()
		return
	}

	// Perform an initial health check when the service starts.
	s.checkBackends()

	// Parse the health check interval, defaulting to 10 seconds if parsing fails.
	interval, err := time.ParseDuration(s.HealthCheck.Interval)
	if err != nil {
		interval = 10 * time.Second
	}

	// Start a goroutine to periodically check backend health.
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.checkBackends()
			}
		}
	}()
}

// checkBackends iterates through all backends and updates their liveness status based on health check results.
// If a backend's status changes, it logs the event and triggers an update to the consistent hash ring.
func (s *Service) checkBackends() {
	changed := false
	for _, b := range s.Backends {
		alive := isBackendAlive(b.URL, s.HealthCheck.Path)
		if b.IsAlive() != alive {
			b.SetAlive(alive)
			changed = true
			status := "down"
			if alive {
				status = "up"
			}
			logger.Info("HealthCheck", "Backend status changed", "backend", b.URL.String(), "status", status)
		}
	}
	if changed {
		s.UpdateHashRing() // Update hash ring if any backend status changed
	}
}

// isBackendAlive performs a simple HTTP HEAD or GET request to a backend to determine its liveness.
// It returns true if the backend responds with a 2xx, 3xx, or 4xx status code within a 2-second timeout, false otherwise.
func isBackendAlive(u *url.URL, path string) bool {
	// Configure an HTTP client with a short timeout to prevent blocking indefinitely.
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	// Construct the full target URL for the health check.
	target := u.Scheme + "://" + u.Host + path

	// Attempt a HEAD request first, as it's generally lighter.
	resp, err := client.Head(target)
	if err != nil {
		// If HEAD fails, fallback to a GET request.
		resp, err = client.Get(target)
		if err != nil {
			return false // Both HEAD and GET failed, backend is considered down.
		}
	}
	defer resp.Body.Close()

	// Consider the backend alive if the status code is between 200 and 499 (inclusive).
	// This broadly covers successful responses and client-side errors, indicating the server is reachable.
	return resp.StatusCode >= 200 && resp.StatusCode < 500
}
