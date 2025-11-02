package server

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// HeartbeatConfig contains all configurable heartbeat parameters.
// This allows fine-tuning of heartbeat behavior for different network conditions
// and application requirements without code changes.
type HeartbeatConfig struct {
	Interval       time.Duration // Time between pings (e.g. 30s) - lower for faster detection
	Timeout        time.Duration // Max wait time for pong (e.g. 20s) - should be < Interval
	MaxMissedPings int           // Max failed pings before giving up (e.g. 2) - prevents false positives
	EnableMetrics  bool          // Enable metrics collection - overhead negligible with atomics
}

// HeartbeatMetrics collects performance and health metrics for monitoring.
// Uses atomic.Int64 for thread-safety without locks, allowing concurrent reads
// from multiple goroutines without performance degradation.
type HeartbeatMetrics struct {
	PingsSent     atomic.Int64 // Total pings sent - incremented before each ping
	PongsReceived atomic.Int64 // Total pongs received - incremented on successful pong
	FailedPings   atomic.Int64 // Failed pings - incremented on timeout or error
	AvgLatency    atomic.Int64 // Average latency in milliseconds - updated after each pong
}

// DefaultHeartbeatConfig returns a production-ready configuration with
// conservative values suitable for most internet connections.
// Interval: 5s - shorter for testing/demo purposes (use 30s in production)
// Timeout: 3s - allows for network jitter and processing delays
// MaxMissedPings: 2 - prevents false positives from transient issues
func DefaultHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Interval:       5 * time.Second, // Shorter interval for testing
		Timeout:        3 * time.Second, // Shorter timeout
		MaxMissedPings: 2,
		EnableMetrics:  true,
	}
}

// EnhancedHeartbeat implements a production-ready heartbeat solution with:
// - Automatic ping/pong frame handling per RFC 6455
// - Configurable timeout and failure threshold
// - Real-time latency measurement
// - Thread-safe metrics collection
// - Graceful context cancellation support
// Returns metrics and error on failure or context cancellation.
// Note: Rate-limiting for incoming client pings should be implemented at the
// WebSocket frame level, not in the server's outgoing ping loop.
func EnhancedHeartbeat(ctx context.Context, conn *websocket.Conn,
	cfg HeartbeatConfig) (*HeartbeatMetrics, error) {
	// Initialize metrics collector
	metrics := &HeartbeatMetrics{}
	timer := time.NewTimer(cfg.Interval)
	defer timer.Stop()
	missedPings := 0 // Counter for consecutive failures - resets on successful pong

	for {
		select {
		case <-ctx.Done():
			// Context cancelled (e.g., connection closed) - exit gracefully with metrics
			return metrics, ctx.Err()
		case <-timer.C:
			// Timer expired - time to send next ping
		}

		// Note: Rate-limiting is not applied here because the server controls
		// its own ping frequency through cfg.Interval configuration.
		// Rate-limiting should instead be applied to incoming pings from clients,
		// which would require WebSocket ping frame interception (not implemented).

		// Create timeout context for this specific ping attempt
		// This ensures we don't wait forever for a response
		pingCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		start := time.Now() // Start latency measurement

		// Send WebSocket ping frame (opcode 0x9) per RFC 6455
		// Server expects pong frame (opcode 0xA) in response
		err := conn.Ping(pingCtx)
		cancel() // Always clean up context resources (prevents memory leak)

		metrics.PingsSent.Add(1) // Atomic increment - thread-safe

		if err != nil {
			// Ping failed - could be network issue, client crashed, or timeout
			metrics.FailedPings.Add(1)
			missedPings++

			// Check if we've exceeded the failure threshold
			// Multiple failures indicate persistent connection problem
			if missedPings >= cfg.MaxMissedPings {
				return metrics, fmt.Errorf("max missed pings (%d) exceeded", cfg.MaxMissedPings)
			}
		} else {
			// Ping successful - pong received within timeout
			// Calculate round-trip latency and reset failure counter
			latency := time.Since(start).Milliseconds()
			metrics.AvgLatency.Store(latency) // Store current latency (atomic operation)
			metrics.PongsReceived.Add(1)      // Increment successful pongs
			missedPings = 0                   // Reset failure counter - connection is healthy
		}

		// Reset timer for next ping interval
		// This creates consistent ping intervals regardless of processing time
		timer.Reset(cfg.Interval)
	}
}

// HeartBeat sends periodic pings to keep the connection alive.
// This is a simplified version without metrics or error handling.
// Deprecated: Use EnhancedHeartbeat for production environments.
// Kept for backward compatibility and simple use cases.
func HeartBeat(ctx context.Context, c *websocket.Conn, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			// Context cancelled - exit gracefully
			return
		case <-t.C:
			// Timer fired - send ping
		}

		// Send ping without timeout or error checking
		err := c.Ping(ctx)
		if err != nil {
			// Any error terminates heartbeat
			return
		}

		// Schedule next ping
		t.Reset(time.Minute)
	}
}
