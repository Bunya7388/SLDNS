# UDPGW - UDP Gateway & Port Forwarding

A lightweight, high-performance UDP gateway for forwarding UDP packets with session management and automatic cleanup.

## Features

- ✅ **UDP Port Forwarding**: Forward UDP traffic to remote servers
- ✅ **Session Management**: Automatic session tracking and idle timeout
- ✅ **Concurrent Processing**: Configurable worker threads (1-256)
- ✅ **Bi-Directional**: Full-duplex UDP communication
- ✅ **High Performance**: Multi-threaded packet processing
- ✅ **Real-Time Statistics**: Live throughput and packet counting
- ✅ **Zero Dependencies**: Standalone executable

## Quick Start

### Basic Usage

```bash
# Terminal 1 - Start server (expose local service)
./bin/udpgw-server -l 0.0.0.0:5555 -r 127.0.0.1:53

# Terminal 2 - Start client (on another machine)
./bin/udpgw-client -l 127.0.0.1:5555 -r 192.168.1.100:5555
```

### Environment Variables

```bash
# Server
export LISTEN="0.0.0.0:5555"
export REMOTE="127.0.0.1:53"
export WORKERS=16
export VERBOSE=true
./run-udpgw-server.sh

# Client
export LISTEN="127.0.0.1:5555"
export REMOTE="192.168.1.100:5555"
export VERBOSE=true
./run-udpgw-client.sh
```

## Command-Line Options

### Server (`udpgw-server`)

```
-l address:port    Listen address (default: 0.0.0.0:5555)
-r address:port    Remote target (default: 127.0.0.1:53)
-w workers         Worker threads (default: 8, range: 1-256)
-t timeout         Session timeout in seconds (default: 300)
-v                 Verbose logging
```

### Client (`udpgw-client`)

```
-l address:port    Local listen (default: 127.0.0.1:5555)
-r address:port    Remote gateway (default: 127.0.0.1:5555)
-v                 Verbose logging
```

## Use Cases

### 1. DNS Tunneling
```bash
# Make local DNS service available remotely
./bin/udpgw-server -l 0.0.0.0:53 -r 127.0.0.1:5353

# Connect from remote client
./bin/udpgw-client -l 127.0.0.1:53 -r dns-server.com:53
```

### 2. Game Server Port Forwarding
```bash
# Forward game server traffic
./bin/udpgw-server -l 0.0.0.0:27015 -r game-backend:27015

# Connect from game client
./bin/udpgw-client -l 127.0.0.1:27015 -r proxy.example.com:27015
```

### 3. VoIP Gateway
```bash
# Forward VoIP packets
./bin/udpgw-server -l 0.0.0.0:5060 -r voip-server:5060 -w 32

# Client access
./bin/udpgw-client -l 127.0.0.1:5060 -r vpn.company.com:5060 -v
```

### 4. IoT Device Communication
```bash
# Gateway for IoT devices
./bin/udpgw-server -l 0.0.0.0:8888 -r iot-hub:8888 -w 64 -t 600

# Remote access
./bin/udpgw-client -l 127.0.0.1:8888 -r cloud.iot-platform.com:8888
```

## Performance Configuration

### Worker Thread Recommendations

| Scenario | Workers | Notes |
|----------|---------|-------|
| Light | 2-4 | Low traffic, minimal CPU |
| Medium | 8-16 | Standard production |
| High | 32-64 | Heavy traffic |
| Very High | 128-256 | Maximum throughput |

### Example Configurations

**Development/Testing**
```bash
./bin/udpgw-server -l 0.0.0.0:5555 -r 127.0.0.1:53 -w 4 -v
```

**Production (Medium Load)**
```bash
./bin/udpgw-server -l 0.0.0.0:5555 -r 10.0.0.5:5555 -w 32 -t 600
```

**High Performance**
```bash
./bin/udpgw-server -l 0.0.0.0:5555 -r 10.0.0.5:5555 -w 128 -t 300 -v
```

## Statistics

The gateway provides real-time statistics every 30 seconds:

```
[STATS] Recv: 10000 | Sent: 10000 | Bytes In: 5.2 MB | Out: 5.2 MB | Sessions: 42
```

**Metrics:**
- **Recv**: Packets received from clients
- **Sent**: Packets sent to clients
- **Sessions**: Active concurrent sessions

## File Binaries

| Binary | Size | Type |
|--------|------|------|
| udpgw-server | 2.2 MB | Linux x86_64 |
| udpgw-client | 2.2 MB | Linux x86_64 |

## Session Management

- **Automatic Creation**: Sessions created on first packet from client
- **Auto Cleanup**: Idle sessions closed after timeout (default: 300s)
- **Thread-Safe**: All session operations are protected with mutexes
- **Scalable**: Supports thousands of concurrent sessions

## Network Architecture

```
┌─────────────────────────────────────────────────┐
│ Client Application (local)                      │
└────────────────┬────────────────────────────────┘
                 │
        ┌────────▼────────┐
        │  udpgw-client   │
        │ 127.0.0.1:5555  │
        └────────┬────────┘
                 │ UDP Tunnel
        ┌────────▼────────┐
        │  udpgw-server   │
        │  0.0.0.0:5555   │
        └────────┬────────┘
                 │
        ┌────────▼────────────┐
        │  Remote Service     │
        │  10.0.0.5:5555      │
        └─────────────────────┘
```

## Deployment

### Direct Execution

```bash
# Server
./bin/udpgw-server -l 0.0.0.0:5555 -r backend:5555 -w 32

# Client
./bin/udpgw-client -l 127.0.0.1:5555 -r proxy.com:5555
```

### Using Scripts

```bash
chmod +x run-udpgw-server.sh run-udpgw-client.sh
./run-udpgw-server.sh
./run-udpgw-client.sh
```

### Systemd Service

Create `/etc/systemd/system/udpgw.service`:

```ini
[Unit]
Description=UDPGW UDP Gateway
After=network.target

[Service]
Type=simple
User=nobody
ExecStart=/usr/local/bin/udpgw-server -l 0.0.0.0:5555 -r backend:5555 -w 32
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable udpgw
sudo systemctl start udpgw
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using the port
sudo lsof -i :5555

# Use different port
./bin/udpgw-server -l 0.0.0.0:6666 -r backend:5555
```

### High CPU Usage

Reduce worker count:
```bash
./bin/udpgw-server -l 0.0.0.0:5555 -r backend:5555 -w 4
```

### No Data Flow

Enable verbose logging:
```bash
./bin/udpgw-server -l 0.0.0.0:5555 -r backend:5555 -v
./bin/udpgw-client -l 127.0.0.1:5555 -r gateway:5555 -v
```

### Memory Leaks

Check session cleanup:
```bash
./bin/udpgw-server -l 0.0.0.0:5555 -r backend:5555 -t 60 -v
# Reduce timeout to force cleanup
```

## Performance Tips

1. **Buffer Sizes**: Automatically set to 256KB per socket
2. **Worker Scaling**: Start with 8, increase if CPU < 70%
3. **Timeout**: Lower for transient connections, higher for persistent
4. **Monitoring**: Check stats every 30 seconds for trends

## Security Considerations

- No encryption - use VPN/TLS wrapper for sensitive data
- No authentication - implement at application layer
- Firewall incoming port appropriately
- Monitor for unusual traffic patterns
- Limit max concurrent sessions if needed

## Version

UDPGW v1.0 - UDP Port Forwarding Gateway
