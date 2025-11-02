# Quick Start Guide - Docker Setup

## Overview
This Docker setup allows you to run the WebSocket server and multiple clients with a single command. Perfect for team collaboration and testing.

## Prerequisites
- Docker installed ([Get Docker](https://docs.docker.com/get-docker/))
- Docker Compose installed (included with Docker Desktop)

## Quick Start

### 1. Start Everything
```bash
docker compose up
```

This will:
- Build the server and client Docker images
- Start the WebSocket server on `localhost:8080`
- Start two client instances that automatically connect
- Show live logs from all containers

### 2. Stop Everything
Press `Ctrl+C` in the terminal, or run:
```bash
docker compose down
```

### 3. Rebuild After Code Changes
```bash
docker compose up --build
```

## Architecture

```
┌─────────────────────┐
│   Docker Network    │
│  websocket-network  │
│                     │
│  ┌──────────────┐   │
│  │   Server     │   │  ← Port 8080 exposed to host
│  │  :8080       │   │
│  └──────────────┘   │
│         ▲  ▲        │
│         │  │        │
│  ┌──────┘  └─────┐  │
│  │               │  │
│ Client1      Client2│
│                     │
└─────────────────────┘
```

## Team Connectivity

### Allow Team Members to Connect

1. **Find Your IP Address:**
   ```bash
   # Linux/Mac
   hostname -I | awk '{print $1}'
   
   # Windows
   ipconfig | findstr IPv4
   ```

2. **Start the Server:**
   ```bash
   docker compose up server
   ```

3. **Team Members Connect:**
   
   Using the compiled binary:
   ```bash
   SERVER_URL=ws://YOUR_IP:8080/ws ./cysl -mode=client
   ```
   
   Using Docker:
   ```bash
   docker run --rm \
     -e SERVER_URL=ws://YOUR_IP:8080/ws \
     aufgabe1-client1 \
     ./cysl -mode=client
   ```
   
   Using Go directly:
   ```bash
   SERVER_URL=ws://YOUR_IP:8080/ws go run main.go -mode=client
   ```

## Configuration

### Running Individual Services

**Server only:**
```bash
docker compose up server
```

**Single client:**
```bash
docker compose up client1
```

**Server + one client:**
```bash
docker compose up server client1
```

### Environment Variables

The client uses the `SERVER_URL` environment variable. You can customize it in `docker-compose.yml`:

```yaml
environment:
  - SERVER_URL=ws://server:8080/ws
```

### Port Configuration

To change the server port, update the `ports` mapping in `docker-compose.yml`:

```yaml
ports:
  - "9000:8080"  # Maps host port 9000 to container port 8080
```

## Monitoring

### Check Server Health
```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"healthy","active_connections":2}
```

### View Logs

All services:
```bash
docker compose logs
```

Specific service:
```bash
docker compose logs server
docker compose logs client1
```

Follow logs in real-time:
```bash
docker compose logs -f
```

## Troubleshooting

### Port Already in Use
If port 8080 is already in use:

```bash
# Find and kill the process
lsof -i :8080
kill -9 <PID>

# Or use a different port
# Edit docker-compose.yml: ports: "8081:8080"
```

### Container Won't Start
```bash
# Clean up and rebuild
docker compose down
docker compose up --build
```

### View Container Status
```bash
docker compose ps
```

### Access Container Shell
```bash
docker compose exec server /bin/sh
```

## Advanced Usage

### Scale Clients
Run multiple client instances:

```bash
docker compose up --scale client1=5
```

### Background Mode
Run services in the background:

```bash
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

### Restart Policy
Services are configured with `restart: unless-stopped` (server) and `restart: on-failure` (clients).

To change:
```yaml
restart: always        # Always restart
restart: no           # Never restart
restart: on-failure   # Restart on error
restart: unless-stopped  # Restart unless manually stopped
```

## Network Details

- **Network Name:** `aufgabe1_websocket-network`
- **Network Type:** Bridge
- **Server Hostname:** `server` (resolvable within the network)
- **Client Access:** Clients use `ws://server:8080/ws` internally

## What Happens During Startup

1. **Build Phase** (if needed):
   - Go dependencies downloaded
   - Binary compiled
   - Minimal Alpine image created

2. **Network Creation**:
   - Docker creates isolated network
   - DNS resolution configured

3. **Server Startup**:
   - Container starts
   - Server binds to `:8080`
   - Health check begins

4. **Client Startup**:
   - Waits for server health check
   - Connects to `ws://server:8080/ws`
   - Sends 5 test messages
   - Heartbeat monitoring active

5. **Shutdown**:
   - Clients finish and exit
   - Server receives graceful shutdown signal
   - All connections closed cleanly

## Tips

- Keep docker compose terminal open to see live logs
- Use `Ctrl+C` once for graceful shutdown, twice to force
- The health check ensures clients don't start before server is ready
- Check `docker-compose.yml` for detailed configuration
- Images are cached - rebuild only when code changes
