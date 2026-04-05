# SLDNS - High-Speed DNS Tunnel over UDP

A high-performance DNS tunnel implementation for fast, reliable data transmission over DNS queries and responses. Optimized for low-latency tunneling with configurable worker pools for maximum throughput.

## Features

- ✅ **High-Speed UDP**: Optimized for fast DNS tunneling
- ✅ **Worker Pool**: Configurable thread count via `-h` flag (1-256 workers)
- ✅ **Session Management**: Automatic session tracking and cleanup
- ✅ **Performance Monitoring**: Real-time statistics (packets/sec, throughput)
- ✅ **Buffer Optimization**: Large socket buffers (4MB read/write)
- ✅ **Ready-to-Run**: Pre-built binaries included
- ✅ **Concurrent Processing**: Non-blocking request handling

## Architecture

```
┌──────────────────────────────────────────────────────┐
│          Client Application                          │
└──────────────────┬───────────────────────────────────┘
                   │
┌──────────────────▼───────────────────────────────────┐
│     dnstt-client (127.0.0.1:8888)                    │
│  - Converts TCP/UDP to DNS queries                   │
│  - Manages local tunnel endpoint                     │
└──────────────────┬───────────────────────────────────┘
                   │ DNS over UDP
┌──────────────────▼───────────────────────────────────┐
│     dnstt-server (0.0.0.0:53)                        │
│  - Worker Pool (configurable threads)                │
│  - Session Manager                                   │
│  - DNS Query Processor                               │
│  - High-speed forwarding                             │
└──────────────────┬───────────────────────────────────┘
                   │
┌──────────────────▼───────────────────────────────────┐
│          Internet / Network                          │
└──────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Linux system (x86_64 or ARM64)
- Go 1.21+ (for building from source)
- Root privileges (for binding to port 53)

### Option 1: Use Pre-Built Binaries

```bash
# Make scripts executable
chmod +x build.sh run-server.sh run-client.sh

# Build binaries
./build.sh

# Server (requires root)
sudo ./run-server.sh

# Client (in another terminal)
./run-client.sh
```

### Option 2: Manual Build

```bash
# Download dependencies
go mod download

# Build server
go build -o bin/dnstt-server ./cmd/server

# Build client
go build -o bin/dnstt-client ./cmd/client

# Run server
sudo bin/dnstt-server -l 0.0.0.0:53 -h 16

# Run client
bin/dnstt-client -s SERVER_IP:53 -l 127.0.0.1:8888
```

## Usage

### Server (dnstt-server)

```bash
sudo ./bin/dnstt-server [OPTIONS]
```

**Options:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-l` | string | `0.0.0.0:53` | Listen address and port |
| `-h` | int | `4` | Number of worker threads (1-256) |
| `-v` | bool | `false` | Enable verbose logging |
| `-log` | string | `` | Log file path |
| `-stats` | int | `30` | Stats reporting interval (seconds) |

**Examples:**

```bash
# High-performance server with 64 workers
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 64 -v

# Specific interface with moderate workers
sudo ./bin/dnstt-server -l 192.168.1.1:53 -h 32

# With file logging
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 48 -log /var/log/dnstt.log
```

### Client (dnstt-client)

```bash
./bin/dnstt-client [OPTIONS]
```

**Options:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-s` | string | `127.0.0.1:53` | DNS server address |
| `-l` | string | `127.0.0.1:8888` | Local listen address |
| `-w` | int | `4` | Number of worker threads |
| `-v` | bool | `false` | Enable verbose logging |

**Examples:**

```bash
# Connect to remote server
./bin/dnstt-client -s 192.168.1.100:53 -l 127.0.0.1:8888 -v

# High-performance client
./bin/dnstt-client -s 10.0.0.5:53 -l 0.0.0.0:9999 -w 8

# Local testing
./bin/dnstt-client -s 127.0.0.1:53 -l 127.0.0.1:8888
```

## Environment Variables

Configure via environment variables (used by run scripts):

```bash
# Server configuration
export LISTEN_ADDR="0.0.0.0:53"
export WORKERS=64
export VERBOSE=true
export LOG_FILE="/var/log/dnstt.log"
./run-server.sh

# Client configuration
export SERVER_ADDR="192.168.1.100:53"
export LOCAL_ADDR="0.0.0.0:9999"
export WORKERS=8
export VERBOSE=true
./run-client.sh
```

## Performance Tuning

### Recommended Worker Counts

| Scenario | Workers | Notes |
|----------|---------|-------|
| Low-traffic | 2-4 | Minimal resource usage |
| Medium-traffic | 8-16 | Balanced performance |
| High-traffic | 32-64 | Good throughput |
| Very high-traffic | 128-256 | Maximum performance |

### System Optimization

For maximum performance, adjust system limits:

```bash
# Increase file descriptor limits
ulimit -n 65536

# Increase socket buffer sizes
sysctl -w net.core.rmem_max=134217728
sysctl -w net.core.wmem_max=134217728
sysctl -w net.ipv4.udp_mem="102400 873800 1677600"
```

### Monitoring

Server statistics are printed at regular intervals:

```
[STATS] Packets: 10000 | PPS: 333.0 | Bytes: 5242880 | BPS: 174.8 Mbps | Sessions: 42
```

**Metrics:**
- **Packets**: Total packets processed
- **PPS**: Packets per second
- **Bytes**: Total bytes transferred
- **BPS**: Bytes per second (in Mbps)
- **Sessions**: Active tunnel sessions

## Systemd Service (Optional)

Create `/etc/systemd/system/dnstt-server.service`:

```ini
[Unit]
Description=DNSTT UDP Tunnel Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/dnstt
ExecStart=/opt/dnstt/bin/dnstt-server -l 0.0.0.0:53 -h 64 -log /var/log/dnstt.log
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable dnstt-server
sudo systemctl start dnstt-server
sudo systemctl status dnstt-server
```

## Binary Sizes

Optimized builds with stripping:

```
-rwxr-xr-x dnstt-server (3.2 MB)
-rwxr-xr-x dnstt-client (2.8 MB)
```

## Architecture Details

### Worker Pool

The server uses a configurable worker pool for concurrent DNS request processing:

- Each worker runs in its own goroutine
- Requests queued with 2x worker buffer capacity
- Lock-free submission via channels
- Graceful shutdown on termination

### Session Management

Each tunnel session:

- Identified by FNV-32a hash of query data
- Tracks creation time and last activity
- Automatic cleanup of idle sessions (5-minute timeout)
- Thread-safe access via RWMutex

### UDP Optimization

- 4MB read/write socket buffers
- Non-blocking DNS query processing
- Efficient memory allocation patterns
- Zero-copy data forwarding where possible

## Troubleshooting

### "Permission denied" error

Port 53 requires root:

```bash
sudo ./run-server.sh
```

### High CPU usage

Adjust worker count downward or system limits:

```bash
./run-server.sh  # Uses WORKERS env var
```

### Statistics show 0 packets

Ensure client is sending traffic:

```bash
# Test with DNS query
./bin/dnstt-client -s 127.0.0.1:53 -l 127.0.0.1:8888 -v
```

### Connection timeouts

Increase stats interval and verify network:

```bash
./bin/dnstt-server -l 0.0.0.0:53 -h 16 -stats 10 -v
```

## Security Notes

- This is a tunneling tool, not an encryption layer
- Implement authentication separately if needed
- Use with VPN/encryption for sensitive data
- Monitor firewall rules for DNS traffic

## Version

DNSTT v1.0 - High-Speed DNS Tunnel