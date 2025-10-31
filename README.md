# Cysl - WebSocket Server and Client

A simple WebSocket communication demo using the `coder/websocket` library with both server and client implementations.

## Project Structure

```
/Cysl
├── main.go           # Entry point for the application
├── Server/
│   └── server.go     # WebSocket server implementation
├── Client/
│   └── client.go     # WebSocket client implementation
├── go.mod            # Go module dependencies
└── README.md         # This file
```

## Features

- **Server**: WebSocket server that listens on port 8080
  - Enhanced heartbeat with configurable parameters (interval, timeout, max missed pings)
  - Performance metrics collection (pings sent/received, latency, failures)
  - Connection limiting per IP address (max 50 connections)
  - Rate limiting to prevent ping flooding attacks
  - Health check endpoint at `/health`
  - Echoes received messages back to clients
  - Logs connection events with detailed metrics
  - Graceful shutdown support

- **Client**: WebSocket client that connects to the server
  - Client-side heartbeat monitoring with metrics
  - Automatic latency measurement for each ping
  - Configurable failure threshold (default: 2 missed pings)
  - Sends test messages to the server
  - Receives and displays server responses
  - Graceful connection handling

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd Cysl
```

2. Install dependencies:
```bash
go mod download
```

## Usage

### Running the Server

Start the WebSocket server:
```bash
go run main.go -mode=server
```

The server will start listening on `localhost:8080`.

### Running the Client

In a separate terminal, start the client:
```bash
go run main.go -mode=client
```

The client will connect to the server and send 5 test messages.

### Custom Server URL

You can specify a custom server URL for the client using the `WEBSOCKET_SERVER` environment variable:
```bash
WEBSOCKET_SERVER=ws://example.com:8080/ws go run main.go -mode=client
```

### Health Check

To check server health:
```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"healthy","active_connections":0}
```

## Building

Build the application:
```bash
go build -o cysl main.go
```

Run the built binary:
```bash
./cysl -mode=server   # Start server
./cysl -mode=client   # Start client
```

## Testing

Test the WebSocket communication by:

1. Starting the server in one terminal:
   ```bash
   go run main.go -mode=server
   ```

2. Starting the client in another terminal:
   ```bash
   go run main.go -mode=client
   ```

You should see:
- Server logs showing connection establishment and message reception
- Client logs showing messages sent and responses received
- Heartbeat Ping/Pong working between client and server

## Graceful Shutdown

Both server and client support graceful shutdown:
- Press `Ctrl+C` to trigger shutdown
- Server will complete ongoing requests before stopping
- Client will close connections properly

## Dependencies

- [github.com/coder/websocket](https://github.com/coder/websocket) - WebSocket implementation

## Module Information

Module: `github.com/deanbregenzer/cysl`
Go Version: 1.24
