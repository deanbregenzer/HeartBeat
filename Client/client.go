package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coder/websocket"
)

const (
	defaultServerURL = "ws://localhost:8080/ws"
	dialTimeout      = 30 * time.Second
	messageTimeout   = 10 * time.Second
)

// Run connects to the WebSocket server and sends test messages
func Run(ctx context.Context) error {
	// Get server URL from environment or use default
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = os.Getenv("WEBSOCKET_SERVER")
	}
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	// Create a context with timeout for dial
	dialCtx, dialCancel := context.WithTimeout(ctx, dialTimeout)
	defer dialCancel()

	// Establish WebSocket connection
	log.Printf("Connecting to server: %s", serverURL)
	conn, resp, err := websocket.Dial(dialCtx, serverURL, &websocket.DialOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close(websocket.StatusInternalError, "")

	log.Printf("Connection established. Server response status: %s", resp.Status)

	// Start client-side heartbeat monitoring
	heartbeatCtx, heartbeatCancel := context.WithCancel(ctx)
	defer heartbeatCancel()

	cfg := DefaultClientHeartbeatConfig()
	go func() {
		metrics, err := ClientHeartbeat(heartbeatCtx, conn, cfg)
		if err != nil {
			log.Printf("Client heartbeat failed: %v | Pings=%d Pongs=%d Failed=%d",
				err,
				metrics.PingsSent.Load(),
				metrics.PongsReceived.Load(),
				metrics.FailedPings.Load())
		}
	}()

	// Send test messages to the server
	for i := 1; i <= 5; i++ {
		select {
		case <-ctx.Done():
			log.Println("Client shutting down...")
			conn.Close(websocket.StatusNormalClosure, "Client shutting down")
			return ctx.Err()
		default:
		}

		// Send ping message
		message := fmt.Sprintf("Client Ping #%d", i)
		log.Printf("Sending message: %s", message)

		writeCtx, writeCancel := context.WithTimeout(ctx, messageTimeout)
		err := conn.Write(writeCtx, websocket.MessageText, []byte(message))
		writeCancel()

		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}

		// Wait for response
		readCtx, readCancel := context.WithTimeout(ctx, messageTimeout)
		_, response, err := conn.Read(readCtx)
		readCancel()

		if err != nil {
			return fmt.Errorf("error reading response: %w", err)
		}

		log.Printf("Received response: %s", string(response))

		// Wait between messages
		time.Sleep(2 * time.Second)
	}

	// Gracefully close the connection
	conn.Close(websocket.StatusNormalClosure, "Client finished")
	log.Println("WebSocket connection closed")

	return nil
}
