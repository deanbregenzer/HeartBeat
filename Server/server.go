// Package server provides WebSocket server implementation with enhanced security
// and heartbeat functionality for maintaining persistent connections.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// Server configuration constants
const (
	ServerAddr          = ":8080"          // Server listen address
	maxMessageSize      = 1024 * 1024      // 1 MB - Maximum allowed message size
	maxConnectionsPerIP = 50               // Max concurrent connections per IP address
	readTimeout         = 10 * time.Second // Timeout for reading messages
	writeTimeout        = 10 * time.Second // Timeout for writing messages
)

// Global connection tracking and management
var (
	activeConnections atomic.Int64                                // Thread-safe active connection counter
	connManager       = NewConnectionManager(maxConnectionsPerIP) // IP-based connection limiter
)

// Start initializes and starts the WebSocket server
func Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebSocket)
	mux.HandleFunc("/health", healthCheck)

	server := &http.Server{
		Addr:         ServerAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("Starting WebSocket server on %s", ServerAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-errChan:
		return fmt.Errorf("server failed to start: %w", err)
	case <-ctx.Done():
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		log.Println("Server stopped")
	}

	return nil
}

// handleWebSocket handles incoming WebSocket connections with comprehensive
// security checks including IP-based rate limiting and connection counting.
// Each connection runs in its own goroutine with automatic heartbeat monitoring.
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Step 1: Check connection limit for this IP address
	// Prevents a single IP from exhausting server resources
	clientIP := r.RemoteAddr
	if !connManager.CheckLimit(clientIP) {
		http.Error(w, "Too many connections from your IP", http.StatusTooManyRequests)
		log.Printf("Connection limit exceeded for %s", clientIP)
		return
	}
	defer connManager.Release(clientIP) // Always release the connection slot

	// Step 2: Upgrade HTTP connection to WebSocket with security options
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  []string{"localhost:*"},       // Only allow local connections
		CompressionMode: websocket.CompressionDisabled, // Disabled for security
	})
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}

	// Step 3: Configure connection limits and tracking
	conn.SetReadLimit(maxMessageSize) // Prevent oversized message attacks
	activeConnections.Add(1)
	defer activeConnections.Add(-1) // Decrement counter on disconnect

	log.Printf("New WebSocket connection from %s (active: %d, ip_conns: %d)",
		r.RemoteAddr, activeConnections.Load(), connManager.GetConnectionCount(clientIP))

	// Step 4: Set up context for graceful shutdown and cleanup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer conn.Close(websocket.StatusInternalError, "") // Ensure connection closure

	// Step 5: Start enhanced heartbeat monitoring in background goroutine
	// This continuously checks connection health via ping/pong frames
	cfg := DefaultHeartbeatConfig()
	go func() {
		metrics, err := EnhancedHeartbeat(ctx, conn, cfg)
		if err != nil {
			// Log detailed metrics on heartbeat failure
			log.Printf("Heartbeat failed for %s: %v | Pings=%d Pongs=%d Failed=%d Latency=%dms",
				r.RemoteAddr, err,
				metrics.PingsSent.Load(),
				metrics.PongsReceived.Load(),
				metrics.FailedPings.Load(),
				metrics.AvgLatency.Load())
		}
		// Cancel main context to trigger cleanup on heartbeat failure
		cancel()
	}()

	// Step 6: Main message handling loop - reads and echoes messages
	for {
		// Read message with timeout to prevent blocking indefinitely
		readCtx, readCancel := context.WithTimeout(ctx, readTimeout)
		msgType, msg, err := conn.Read(readCtx)
		readCancel()

		if err != nil {
			log.Printf("Read error from %s: %v", r.RemoteAddr, err)
			break // Exit loop on any read error
		}

		log.Printf("Server received from %s: %s", r.RemoteAddr, string(msg))

		// Echo the received message back to the client
		writeCtx, writeCancel := context.WithTimeout(ctx, writeTimeout)
		err = conn.Write(writeCtx, msgType, []byte(fmt.Sprintf("Server echoes: %s", msg)))
		writeCancel()

		if err != nil {
			log.Printf("Write error to %s: %v", r.RemoteAddr, err)
			break // Exit loop on write failure
		}
	}

	// Clean shutdown with normal closure status
	conn.Close(websocket.StatusNormalClosure, "")
	log.Printf("Connection closed for %s (active: %d)",
		r.RemoteAddr, activeConnections.Load())
}

// healthCheck provides a simple HTTP health check endpoint for monitoring
// Returns JSON with server status and current active connection count
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","active_connections":` +
		fmt.Sprintf("%d", activeConnections.Load()) + `}`))
}
