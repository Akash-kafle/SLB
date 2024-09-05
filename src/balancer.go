package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Major balancing act is performed here
type Backend struct {
	Url          *url.URL
	IsAlive      bool
	mux          sync.Mutex
	ReverseProxy *httputil.ReverseProxy
}

type Pool struct {
	backends []*Backend
	current  uint64
}

var pool Pool

// SetAlive for this backend
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.IsAlive = alive
	b.mux.Unlock()
}

// IsAlive returns true when backend is alive
func (b *Backend) Alive() (alive bool) {
	b.mux.Lock()
	alive = b.IsAlive
	b.mux.Unlock()
	return
}

// Returns the next index to the currently being used ones
func (s *Pool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

// GetNextPeer returns next active peer to take a connection
func (s *Pool) GetNextPeer() *Backend {
	// loop entire backends to find out an Alive backend
	next := s.NextIndex()
	l := len(s.backends) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(s.backends) // take an index by modding with length
		// if we have an alive backend, use it and store if its not the original one
		if s.backends[idx].Alive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx)) // mark the current one
			}
			return s.backends[idx]
		}
	}
	return nil
}

// lb load balances the incoming request
func Lb(w http.ResponseWriter, r *http.Request) {
	peer := pool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// GetAttemptsFromContext returns the attempts for request
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value( /*retry value (placeholder for now)*/ 3).(int); ok {
		return retry
	}
	return 0
}

// isAlive checks whether a backend is Alive by establishing a TCP connection
func IsBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	_ = conn.Close() // close it, we dont need to maintain this connection
	return true
}

// HealthCheck pings the backends and update the status
func (s *Pool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := IsBackendAlive(b.Url)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.Url, status)
	}
}
