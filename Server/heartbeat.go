package server

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// HeartbeatConfig contains all configurable heartbeat parameters
type HeartbeatConfig struct {
	Interval       time.Duration // Time between pings (e.g. 30s)
	Timeout        time.Duration // Max wait time for pong (e.g. 20s)
	MaxMissedPings int           // Max failed pings before giving up (e.g. 2)
	EnableMetrics  bool          // Enable metrics collection
}

// HeartbeatMetrics collects performance and health metrics
// Uses atomic.Int64 for thread-safety without locks
type HeartbeatMetrics struct {
	PingsSent     atomic.Int64 // Total pings sent
	PongsReceived atomic.Int64 // Total pongs received
	FailedPings   atomic.Int64 // Failed pings
	AvgLatency    atomic.Int64 // Average latency (ms)
}

// DefaultHeartbeatConfig returns a production-ready configuration
func DefaultHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Interval:       30 * time.Second,
		Timeout:        20 * time.Second,
		MaxMissedPings: 2,
		EnableMetrics:  true,
	}
}

// EnhancedHeartbeat implements a production-ready heartbeat solution
// with monitoring, timeout handling and configurable failure threshold
func EnhancedHeartbeat(ctx context.Context, conn *websocket.Conn,
	cfg HeartbeatConfig) (*HeartbeatMetrics, error) {
	// Initialize metrics
	metrics := &HeartbeatMetrics{}
	timer := time.NewTimer(cfg.Interval)
	defer timer.Stop()
	missedPings := 0 // Counter for consecutive failures

	for {
		select {
		case <-ctx.Done():
			// Context cancelled - cleanly exit with metrics
			return metrics, ctx.Err()
		case <-timer.C:
			// Timer expired - time for next ping
		}

		// Context with timeout for this specific ping
		pingCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		start := time.Now() // Start latency measurement

		// Send protocol-level ping
		err := conn.Ping(pingCtx)
		cancel() // Always clean up context (avoid resource leak)

		metrics.PingsSent.Add(1) // Atomic increment

		if err != nil {
			// Ping failed - could be network issue
			metrics.FailedPings.Add(1)
			missedPings++

			// Threshold exceeded? Consider connection dead
			if missedPings >= cfg.MaxMissedPings {
				return metrics, fmt.Errorf("max missed pings (%d) exceeded", cfg.MaxMissedPings)
			}
		} else {
			// Ping successful - calculate latency and reset counter
			latency := time.Since(start).Milliseconds()
			metrics.AvgLatency.Store(latency)
			metrics.PongsReceived.Add(1)
			missedPings = 0 // Reset on success
		}

		// Reset timer for next ping
		timer.Reset(cfg.Interval)
	}
}

// HeartBeat sends periodic pings to keep the connection alive.
// Deprecated: Use EnhancedHeartbeat for production
func HeartBeat(ctx context.Context, c *websocket.Conn, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		err := c.Ping(ctx)
		if err != nil {
			return
		}

		t.Reset(time.Minute)
	}
}
