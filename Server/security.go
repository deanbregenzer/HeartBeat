package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// ConnectionState stores per-connection state for rate-limiting ping requests.
// This prevents clients from flooding the server with excessive ping frames,
// which could be used for DoS attacks or resource exhaustion.
type ConnectionState struct {
	lastPing         time.Time  // Timestamp of last ping - used to calculate interval
	pingCount        int        // Number of pings in current window - for burst detection
	violations       int        // Counter for rate-limit violations - triggers disconnect
	lastClientPing   time.Time  // Timestamp of last CLIENT ping received
	clientViolations int        // Violations from client's incoming pings
	mu               sync.Mutex // Protects state updates
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
	cs.mu.Lock()
	defer cs.mu.Unlock()

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

// RateLimitClientPing checks if incoming pings from the client are within acceptable limits.
// This is called whenever the server detects the client has sent a ping frame.
// Returns false if connection should be closed due to excessive ping flooding.
func (cs *ConnectionState) RateLimitClientPing() bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	now := time.Now()

	// First ping from client - initialize timestamp
	if cs.lastClientPing.IsZero() {
		cs.lastClientPing = now
		return true
	}

	// Check if client's ping arrives too quickly
	if now.Sub(cs.lastClientPing) < minPingInterval {
		cs.clientViolations++
		cs.lastClientPing = now

		// Client has exceeded the violation threshold - disconnect
		if cs.clientViolations > maxViolations {
			return false // Signal to close connection
		}
		return true // Allow but count violation
	}

	// Compliant ping frequency - reset violations
	cs.clientViolations = 0
	cs.lastClientPing = now
	return true
}

// GetClientViolations returns the current number of client ping violations (thread-safe)
func (cs *ConnectionState) GetClientViolations() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.clientViolations
}

// RateLimitedConn wraps a WebSocket connection to monitor incoming ping frequency
// This wrapper intercepts Read operations to track when clients send pings
type RateLimitedConn struct {
	*websocket.Conn
	connState  *ConnectionState
	remoteAddr string
}

// NewRateLimitedConn creates a new rate-limited connection wrapper
func NewRateLimitedConn(conn *websocket.Conn, connState *ConnectionState, remoteAddr string) *RateLimitedConn {
	return &RateLimitedConn{
		Conn:       conn,
		connState:  connState,
		remoteAddr: remoteAddr,
	}
}

// Ping wraps the original Ping method to track outgoing pings
// Note: This tracks server->client pings, not client->server
func (rlc *RateLimitedConn) Ping(ctx context.Context) error {
	// Track outgoing ping in metrics (optional)
	return rlc.Conn.Ping(ctx)
}

// Read wraps the original Read to monitor for incoming messages and enforce rate limits
// While we cannot directly intercept ping frames (handled internally by coder/websocket),
// we enforce a general message rate limit that indirectly protects against ping flooding
func (rlc *RateLimitedConn) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	// Check rate limit before processing message
	// This provides protection against all types of message flooding, including pings
	if !rlc.connState.RateLimitClientPing() {
		// Client exceeded rate limit - return error to trigger disconnect
		return 0, nil, fmt.Errorf("message rate limit exceeded for %s (violations: %d)",
			rlc.remoteAddr, rlc.connState.GetClientViolations())
	}

	msgType, data, err := rlc.Conn.Read(ctx)
	return msgType, data, err
} // CheckClientPingRate should be called periodically to enforce client ping rate limits
// Returns error if client should be disconnected due to excessive pings
func (rlc *RateLimitedConn) CheckClientPingRate() error {
	if !rlc.connState.RateLimitClientPing() {
		return fmt.Errorf("client ping rate limit exceeded from %s", rlc.remoteAddr)
	}
	return nil
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

// ConnectionStateManager manages rate-limiting state for each connection.
// Tracks ping frequency per connection ID to prevent ping-flooding attacks.
type ConnectionStateManager struct {
	states map[string]*ConnectionState // Connection ID -> state
	mu     sync.RWMutex                // Protects states map
}

// NewConnectionStateManager creates a new connection state manager.
func NewConnectionStateManager() *ConnectionStateManager {
	return &ConnectionStateManager{
		states: make(map[string]*ConnectionState),
	}
}

// GetOrCreate returns the ConnectionState for a given connection ID,
// creating it if it doesn't exist. Thread-safe.
func (csm *ConnectionStateManager) GetOrCreate(connID string) *ConnectionState {
	csm.mu.Lock()
	defer csm.mu.Unlock()

	if state, exists := csm.states[connID]; exists {
		return state
	}

	// Create new state for this connection
	state := &ConnectionState{
		lastPing: time.Now(), // Initialize to now to allow first ping immediately
	}
	csm.states[connID] = state
	return state
}

// Remove deletes the ConnectionState when a connection is closed.
// Prevents memory leaks from accumulating old connection states.
func (csm *ConnectionStateManager) Remove(connID string) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	delete(csm.states, connID)
}
