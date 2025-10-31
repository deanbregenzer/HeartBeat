// Package main provides the entry point for the WebSocket application.
// It supports both server and client modes via command-line flags,
// making it suitable for Docker containers with different configurations.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	client "github.com/deanbregenzer/cysl/Client"
	server "github.com/deanbregenzer/cysl/Server"
)

var (
	// mode determines whether to run as server or client
	// Set via -mode flag: ./cysl -mode=server or ./cysl -mode=client
	mode string
)

// init runs before main() and sets up command-line flags
func init() {
	flag.StringVar(&mode, "mode", "server", "Run mode: server or client")
	flag.Parse()
}

// main is the application entry point that:
// 1. Sets up graceful shutdown via signal handling
// 2. Starts either server or client based on -mode flag
// 3. Handles errors and ensures clean shutdown
func main() {
	// Create context that listens for OS interrupt signals (Ctrl+C, SIGTERM)
	// This enables graceful shutdown in both Docker and terminal environments
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop() // Ensure signal notification is stopped

	var err error
	// Route to appropriate mode based on flag
	switch mode {
	case "server":
		log.Println("Starting in server mode...")
		err = server.Start(ctx) // Start WebSocket server
	case "client":
		log.Println("Starting in client mode...")
		err = client.Run(ctx) // Start WebSocket client
	default:
		// Invalid mode - exit with error
		log.Fatalf("Invalid mode: %s. Use 'server' or 'client'", mode)
	}

	// Check for errors during execution
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Application shutdown complete")
}
