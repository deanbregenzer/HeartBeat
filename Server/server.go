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

const (
	ServerAddr          = ":8080"
	maxMessageSize      = 1024 * 1024 // 1 MB
	maxConnectionsPerIP = 50          // Max connections per IP address
	readTimeout         = 10 * time.Second
	writeTimeout        = 10 * time.Second
)

var activeConnections atomic.Int64
var connManager = NewConnectionManager(maxConnectionsPerIP)

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

// handleWebSocket handles incoming WebSocket connections with security checks
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 1. Check connection limit for this IP
	clientIP := r.RemoteAddr
	if !connManager.CheckLimit(clientIP) {
		http.Error(w, "Too many connections from your IP", http.StatusTooManyRequests)
		log.Printf("Connection limit exceeded for %s", clientIP)
		return
	}
	defer connManager.Release(clientIP)

	// 2. Accept WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns:  []string{"localhost:*"},
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		return
	}

	conn.SetReadLimit(maxMessageSize)
	activeConnections.Add(1)
	defer activeConnections.Add(-1)

	log.Printf("New WebSocket connection from %s (active: %d, ip_conns: %d)",
		r.RemoteAddr, activeConnections.Load(), connManager.GetConnectionCount(clientIP))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer conn.Close(websocket.StatusInternalError, "")

	// Start enhanced heartbeat in separate goroutine
	cfg := DefaultHeartbeatConfig()
	go func() {
		metrics, err := EnhancedHeartbeat(ctx, conn, cfg)
		if err != nil {
			log.Printf("Heartbeat failed for %s: %v | Pings=%d Pongs=%d Failed=%d Latency=%dms",
				r.RemoteAddr, err,
				metrics.PingsSent.Load(),
				metrics.PongsReceived.Load(),
				metrics.FailedPings.Load(),
				metrics.AvgLatency.Load())
		}
		// Cancel context on heartbeat failure
		cancel()
	}()

	// Handle messages from client
	for {
		readCtx, readCancel := context.WithTimeout(ctx, readTimeout)
		msgType, msg, err := conn.Read(readCtx)
		readCancel()

		if err != nil {
			log.Printf("Read error from %s: %v", r.RemoteAddr, err)
			break
		}

		log.Printf("Server received from %s: %s", r.RemoteAddr, string(msg))

		// Echo the message back
		writeCtx, writeCancel := context.WithTimeout(ctx, writeTimeout)
		err = conn.Write(writeCtx, msgType, []byte(fmt.Sprintf("Server echoes: %s", msg)))
		writeCancel()

		if err != nil {
			log.Printf("Write error to %s: %v", r.RemoteAddr, err)
			break
		}
	}

	conn.Close(websocket.StatusNormalClosure, "")
	log.Printf("Connection closed for %s (active: %d)",
		r.RemoteAddr, activeConnections.Load())
}

// healthCheck provides a health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","active_connections":` +
		fmt.Sprintf("%d", activeConnections.Load()) + `}`))
}
