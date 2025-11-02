package client

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// HeartbeatConfig contains all configurable heartbeat parameters for client
type HeartbeatConfig struct {
	Interval       time.Duration // Time between pings
	Timeout        time.Duration // Max wait time for pong
	MaxMissedPings int           // Max failed pings before giving up
	EnableMetrics  bool          // Enable metrics collection
}

// HeartbeatMetrics collects performance and health metrics
type HeartbeatMetrics struct {
	PingsSent     atomic.Int64 // Total pings sent
	PongsReceived atomic.Int64 // Total pongs received
	FailedPings   atomic.Int64 // Failed pings
	AvgLatency    atomic.Int64 // Average latency (ms)
}

// DefaultClientHeartbeatConfig returns client-side heartbeat configuration
func DefaultClientHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Interval:       5 * time.Second, // Shorter interval for testing
		Timeout:        3 * time.Second, // Shorter timeout
		MaxMissedPings: 2,
		EnableMetrics:  true,
	}
}

// ClientHeartbeat implements client-side heartbeat monitoring
// The client reads pong responses automatically through the Read() loop
func ClientHeartbeat(ctx context.Context, conn *websocket.Conn,
	cfg HeartbeatConfig) (*HeartbeatMetrics, error) {
	metrics := &HeartbeatMetrics{}
	timer := time.NewTimer(cfg.Interval)
	defer timer.Stop()
	missedPings := 0

	for {
		select {
		case <-ctx.Done():
			return metrics, ctx.Err()
		case <-timer.C:
		}

		// Send ping with timeout
		pingCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		start := time.Now()

		err := conn.Ping(pingCtx)
		cancel()

		metrics.PingsSent.Add(1)

		if err != nil {
			metrics.FailedPings.Add(1)
			missedPings++
			log.Printf("Client ping failed: %v (missed: %d/%d)",
				err, missedPings, cfg.MaxMissedPings)

			if missedPings >= cfg.MaxMissedPings {
				return metrics, fmt.Errorf("max missed pings (%d) exceeded", cfg.MaxMissedPings)
			}
		} else {
			latency := time.Since(start).Milliseconds()
			metrics.AvgLatency.Store(latency)
			metrics.PongsReceived.Add(1)
			missedPings = 0
			log.Printf("Client ping successful (latency: %dms)", latency)
		}

		timer.Reset(cfg.Interval)
	}
}
