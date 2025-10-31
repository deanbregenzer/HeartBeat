package server

import (
	"sync"
	"time"
)

// ConnectionState stores per-connection state for rate-limiting ping requests.
// This prevents clients from flooding the server with excessive ping frames,
// which could be used for DoS attacks or resource exhaustion.
type ConnectionState struct {
	lastPing   time.Time // Timestamp of last ping - used to calculate interval
	pingCount  int       // Number of pings in current window - for burst detection
	violations int       // Counter for rate-limit violations - triggers disconnect
}

// Rate limiting constants
const (
	minPingInterval = 10 * time.Second // Minimum interval between pings - prevents flooding
	maxViolations   = 3                // Max allowed violations before disconnect - prevents abuse
)

// RateLimitPing checks if the ping frequency is acceptable and enforces rate limits.
// Returns false if connection should be closed due to excessive violations.
// This implements a simple but effective rate limiting algorithm:
// - Track time since last ping
// - Count violations (pings that arrive too quickly)
// - Disconnect after too many violations
func (cs *ConnectionState) RateLimitPing() bool {
	// Check if ping arrives before minimum interval has elapsed
	if time.Since(cs.lastPing) < minPingInterval {
		cs.violations++
		// Exceeded violation threshold - this client is misbehaving
		if cs.violations > maxViolations {
			return false // Signal to close connection
		}
	} else {
		// Compliant ping frequency: reset violation counter
		// This gives clients a clean slate after proper behavior
		cs.violations = 0
	}
	cs.lastPing = time.Now() // Update timestamp for next check
	return true              // Ping allowed - connection continues
}

// ConnectionManager manages connection limits per IP address to prevent
// a single client from exhausting server resources by opening too many
// concurrent connections. This is a critical DoS protection mechanism.
type ConnectionManager struct {
	connections map[string]int // IP address -> connection count
	mu          sync.Mutex     // Protects connections map from concurrent access
	maxPerIP    int            // Maximum connections allowed per IP
}

// NewConnectionManager creates a new connection manager with specified
// per-IP connection limit. The manager uses a mutex for thread-safety
// as it's accessed concurrently by multiple goroutines (one per connection).
func NewConnectionManager(maxPerIP int) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]int),
		maxPerIP:    maxPerIP,
	}
}

// CheckLimit checks if the IP has reached its connection limit and atomically
// increments the counter if allowed. This operation must be atomic to prevent
// race conditions where multiple goroutines check the limit simultaneously.
// Returns true if connection is allowed, false if limit is exceeded.
func (cm *ConnectionManager) CheckLimit(ip string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock() // Ensure lock is released even if panic occurs

	// Check if limit already reached for this IP
	if cm.connections[ip] >= cm.maxPerIP {
		return false // Reject connection - limit exceeded
	}

	// Atomically increment connection counter for this IP
	// This prevents race conditions in concurrent connection attempts
	cm.connections[ip]++
	return true // Allow connection
}

// Release atomically decrements the connection count for an IP when a
// connection is closed. This must be called in a defer statement to ensure
// the count is always decremented even if connection handler panics.
func (cm *ConnectionManager) Release(ip string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Sanity check: only decrement if count is positive
	if cm.connections[ip] > 0 {
		cm.connections[ip]--
	}

	// Clean up map entry if no more connections from this IP
	// This prevents unbounded memory growth over time
	if cm.connections[ip] == 0 {
		delete(cm.connections, ip)
	}
}

// GetConnectionCount returns the current connection count for an IP.
// Used for logging and monitoring purposes. Thread-safe via mutex.
func (cm *ConnectionManager) GetConnectionCount(ip string) int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.connections[ip]
}
