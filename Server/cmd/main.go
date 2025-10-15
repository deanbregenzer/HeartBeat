package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	serverAddr = ":8080"
	readLimit  = 1024 // 1 KB read limit for messages
)

func main() {
	// Create a context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,   // Interrupt signal (Ctrl+C)
		syscall.SIGTERM // Termination signal
	)
	defer stop()

	// Create a new HTTP server with WebSocket handler
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebSocket)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the server in a separate goroutine
	go func() {
		log.Printf("Starting WebSocket server on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Create a shutdown context with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	log.Println("Shutting down server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// handleWebSocket manages incoming WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Accept the WebSocket connection with some options
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // For development only - remove in production
		CompressionMode:    websocket.CompressionDisabled, // Optional: can be configured
	})
	if err != nil {
		log.Printf("Failed to accept WebSocket connection: %v", err)
		http.Error(w, "Could not open WebSocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "closing connection")

	// Log new connection
	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Set read limit to prevent excessive message sizes
	conn.SetReadLimit(readLimit)

	// Context for the connection
	ctx := r.Context()

	// Start a background goroutine for connection monitoring
	go monitorConnection(ctx, conn)

	// Echo server: read and send back messages
	for {
		// Read JSON message
		var message string
		err := wsjson.Read(ctx, conn, &message)
		if err != nil {
			// Check if connection is closed normally
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Log received message
		log.Printf("Received message: %s", message)

		// Echo the message back to the client
		err = wsjson.Write(ctx, conn, fmt.Sprintf("Server received: %s", message))
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}

	// Close the connection with a normal closure status
	conn.Close(websocket.StatusNormalClosure, "")
}

// monitorConnection provides basic connection monitoring and heartbeat functionality
func monitorConnection(ctx context.Context, conn *websocket.Conn) {
	// Periodic ping to check connection health
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send a ping to check connection
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Ping(pingCtx)
			cancel()

			if err != nil {
				log.Printf("Connection ping failed: %v", err)
				conn.Close(websocket.StatusAbnormalClosure, "ping failed")
				return
			}

		case <-ctx.Done():
			return
		}
	}
}