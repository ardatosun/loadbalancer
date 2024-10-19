package backends

import (
	"net/url"
	"sync"
	"time"
)

// Backend holds the data for each backend server
type Backend struct {
	URL     *url.URL
	Alive   bool
	Mutex   sync.RWMutex
	Latency time.Duration
}

// SetAlive sets the status of the backend (uses a write lock)
func (b *Backend) SetAlive(alive bool) {
	b.Mutex.Lock()
	b.Alive = alive
	b.Mutex.Unlock()
}

// IsAlive checks if the backend is alive (uses a read lock)
func (b *Backend) IsAlive() bool {
	b.Mutex.RLock()
	alive := b.Alive
	b.Mutex.RUnlock()
	return alive
}

// SetLatency sets the latency for the backend (uses a write lock)
func (b *Backend) SetLatency(latency time.Duration) {
	b.Mutex.Lock()
	b.Latency = latency
	b.Mutex.Unlock()
}

// GetLatency gets the current latency for the backend (uses a read lock)
func (b *Backend) GetLatency() time.Duration {
	b.Mutex.RLock()
	latency := b.Latency
	b.Mutex.RUnlock()
	return latency
}
