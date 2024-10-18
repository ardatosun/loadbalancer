package serverpool

import (
	"loadbalancer/backends"
	"net/url"
	"testing"
)

// Helper function to create a new backend
func newBackend(rawurl string, alive bool) *backends.Backend {
	parsedURL, _ := url.Parse(rawurl)
	return &backends.Backend{
		URL:   parsedURL,
		Alive: alive,
	}
}

// Test that the round-robin algorithm correctly selects backends
func TestServerPool_NextIndex(t *testing.T) {
	pool := &ServerPool{
		backends: []*backends.Backend{
			newBackend("http://server1", true),
			newBackend("http://server2", true),
			newBackend("http://server3", true),
		},
	}

	// Test round-robin indexing
	first := pool.NextIndex()
	second := pool.NextIndex()
	third := pool.NextIndex()

	if first != 1 || second != 2 || third != 0 {
		t.Errorf("Expected round-robin to cycle through backends, got first: %d, second: %d, third: %d", first, second, third)
	}
}

// Test that GetNextPeer skips unhealthy backends
func TestServerPool_GetNextPeer_SkipUnhealthy(t *testing.T) {
	pool := &ServerPool{
		backends: []*backends.Backend{
			newBackend("http://server1", true),
			newBackend("http://server2", false), // Mark this backend as unhealthy
			newBackend("http://server3", true),
		},
	}

	peer1 := pool.GetNextPeer()
	peer2 := pool.GetNextPeer()

	if peer1.URL.String() == "http://server2" || peer2.URL.String() == "http://server2" {
		t.Errorf("Expected round-robin to skip unhealthy backend 'http://server2', but got peer1: %s, peer2: %s", peer1.URL, peer2.URL)
	}
}

// Test that GetNextPeer returns nil if all backends are unhealthy
func TestServerPool_GetNextPeer_AllUnhealthy(t *testing.T) {
	pool := &ServerPool{
		backends: []*backends.Backend{
			newBackend("http://server1", false),
			newBackend("http://server2", false),
			newBackend("http://server3", false),
		},
	}

	peer := pool.GetNextPeer()

	if peer != nil {
		t.Errorf("Expected GetNextPeer to return nil when all backends are unhealthy, got: %s", peer.URL)
	}
}
