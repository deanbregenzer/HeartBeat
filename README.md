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

### Quick Start with Docker Compose (Recommended)

The easiest way to run the entire setup is with Docker Compose:

```bash
docker compose up
```

This will:
- Build the server and client containers
- Start the WebSocket server on `localhost:8080`
- Start two client instances that automatically connect to the server
- Enable team members to connect by accessing your machine's IP on port 8080

To stop all services:
```bash
docker compose down
```

To rebuild after code changes:
```bash
docker compose up --build
```

### Building the Application (Local Development)

First, build the application:

```bash
go build -o cysl .
```

### Running the Server

Start the WebSocket server:

```bash
./cysl -mode=server
```

Or using go run:

```bash
go run main.go -mode=server
```

The server will start on `http://localhost:8080`

### Running the Client

In a separate terminal, start the client:

```bash
./cysl -mode=client
```

Or using go run:
```bash
go run main.go -mode=client
```

The client will:
- Connect to the server at `ws://localhost:8080/ws`
- Start heartbeat monitoring (pings every 30s)
- Send 5 test messages
- Display server responses
- Show heartbeat metrics

### Custom Server URL

You can specify a custom server URL for the client using the `SERVER_URL` or `WEBSOCKET_SERVER` environment variable:
```bash
SERVER_URL=ws://example.com:8080/ws go run main.go -mode=client
```

### Team Connectivity with Docker

To allow team members to connect to your Docker server:

1. Start the server with Docker Compose:
   ```bash
   docker compose up server
   ```

2. Find your machine's IP address:
   ```bash
   hostname -I  # Linux
   ipconfig     # Windows
   ```

3. Team members can connect using your IP:
   ```bash
   SERVER_URL=ws://YOUR_IP:8080/ws ./cysl -mode=client
   ```

   Or with Docker:
   ```bash
   docker run --rm \
     -e SERVER_URL=ws://YOUR_IP:8080/ws \
     $(docker build -q .) \
     ./cysl -mode=client
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
