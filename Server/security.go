package server

import (
	"sync"
	"time"
)

// ConnectionState stores state for rate-limiting per connection
type ConnectionState struct {
	lastPing   time.Time // Timestamp of last ping
	pingCount  int       // Number of pings in current window
	violations int       // Counter for rate-limit violations
}

const (
	minPingInterval = 10 * time.Second // Minimum interval between pings
	maxViolations   = 3                // Max allowed violations before disconnect
)

// RateLimitPing checks if ping frequency is acceptable
func (cs *ConnectionState) RateLimitPing() bool {
	// Check if ping comes too early
	if time.Since(cs.lastPing) < minPingInterval {
		cs.violations++
		// Too many violations: signal to close connection
		if cs.violations > maxViolations {
			return false // Signal to close connection
		}
	} else {
		// Compliant ping frequency: reset violations
		cs.violations = 0
	}
	cs.lastPing = time.Now()
	return true // Ping allowed
}

// ConnectionManager manages connection limits per IP
type ConnectionManager struct {
	connections map[string]int
	mu          sync.Mutex
	maxPerIP    int
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(maxPerIP int) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]int),
		maxPerIP:    maxPerIP,
	}
}

// CheckLimit checks and increments connection count for an IP
func (cm *ConnectionManager) CheckLimit(ip string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if limit already reached
	if cm.connections[ip] >= cm.maxPerIP {
		return false // Reject connection
	}

	// Increment connection counter for this IP
	cm.connections[ip]++
	return true // Allow connection
}

// Release decrements the connection count for an IP
func (cm *ConnectionManager) Release(ip string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.connections[ip] > 0 {
		cm.connections[ip]--
	}

	// Clean up if no connections from this IP
	if cm.connections[ip] == 0 {
		delete(cm.connections, ip)
	}
}

// GetConnectionCount returns current connection count for an IP
func (cm *ConnectionManager) GetConnectionCount(ip string) int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.connections[ip]
}
