# DNSTT Deployment Guide

## Pre-Built Binaries Available

Your repository now includes pre-compiled, ready-to-run binaries:

- **dnstt-server**: High-speed DNS tunnel server (2.2MB)
- **dnstt-client**: DNS tunnel client (2.3MB)

These binaries support Linux x86_64 systems and can be deployed immediately without compilation.

## Quick Deployment

### 1. Direct Binary Execution

```bash
# Download/clone the repository
git clone https://github.com/Bunya7388/SLDNS.git
cd SLDNS

# Make scripts executable
chmod +x bin/dnstt-server bin/dnstt-client run-server.sh run-client.sh

# Run server (requires root for port 53)
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 32

# Run client (in another terminal)
./bin/dnstt-client -s <SERVER_IP>:53 -l 127.0.0.1:8888
```

### 2. Using Convenience Scripts

```bash
# Server with defaults (16 workers)
sudo ./run-server.sh

# Server with custom workers
WORKERS=64 ./run-server.sh

# Client
./run-client.sh
```

### 3. Using Make Commands

```bash
# Build (if needed)
make build

# Run server
make run-server

# Run client
make run-client
```

## System Deployment

### Minimal Setup (Local Testing)

```bash
# Terminal 1 - Server
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 16

# Terminal 2 - Client
./bin/dnstt-client -s 127.0.0.1:53 -l 127.0.0.1:8888
```

### Network Deployment

```bash
# On SERVER machine
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 64

# On CLIENT machine (replace 192.168.1.100 with server IP)
./bin/dnstt-client -s 192.168.1.100:53 -l 127.0.0.1:8888
```

### High-Performance Setup

```bash
# Adjust system limits
ulimit -n 65536

# Server with 128 workers
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 128 -v

# Monitor in separate window
watch -n 1 'head -n 5 /proc/$(pgrep dnstt-server)/stat'
```

## Configuration

### Server Options

```
-l address:port      Listen address (default: 0.0.0.0:53)
-h workers          Worker threads 1-256 (default: 4)
-v                  Verbose logging
-log file           Log file path
-stats interval     Stats interval in seconds (default: 30)
```

### Client Options

```
-s address:port      Server address (default: 127.0.0.1:53)
-l address:port      Local listen (default: 127.0.0.1:8888)
-w workers          Worker threads (default: 4)
-v                  Verbose logging
```

## Docker Deployment

### Quick Start with Docker

```bash
# Build image
docker build -t dnstt .

# Run server
docker run -d --name dnstt-server -p 53:53/udp \
  dnstt dnstt-server -l 0.0.0.0:53 -h 64 -v

# Run client
docker run -d --name dnstt-client -p 8888:8888 \
  dnstt dnstt-client -s dnstt-server:53 -l 0.0.0.0:8888
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Performance Validation

### Basic Testing

```bash
# Check if server is responding
dig @127.0.0.1 example.com

# Monitor statistics
while true; do 
  tail -2 /var/log/dnstt.log
  sleep 5
done
```

### Load Testing

```bash
# Install load testing tool
sudo apt-get install -y dnsperf

# Generate load
dnsperf -s 127.0.0.1 -c 10 -n 1000 -l queryfile.txt
```

### Monitoring

```bash
# Real-time stats (if verbose enabled)
sudo tail -f /var/log/dnstt.log | grep STATS

# Network monitoring
nethogs

# Resource usage
top -p $(pgrep -f dnstt-server)
```

## Production Checklist

- [ ] Verify binaries are executable: `ls -la bin/`
- [ ] Test on local system first
- [ ] Configure firewall for port 53
- [ ] Set up log rotation if using log files
- [ ] Monitor resource usage during peak
- [ ] Document custom configurations
- [ ] Set up monitoring/alerting

## Troubleshooting

### Server won't start

```bash
# Check if port 53 is in use
sudo ss -ulnp | grep 53

# Try different port
./bin/dnstt-server -l 0.0.0.0:8053

# Ensure root for port 53
sudo ./bin/dnstt-server -l 0.0.0.0:53
```

### High memory usage

```bash
# Reduce workers
./bin/dnstt-server -l 0.0.0.0:53 -h 8

# Monitor memory
free -h && ps aux | grep dnstt-server
```

### No packets in statistics

```bash
# Enable verbose mode
./bin/dnstt-server -l 0.0.0.0:53 -h 16 -v

# Check if client is connecting
netstat -an | grep 53
```

## File Structure

```
SLDNS/
├── bin/
│   ├── dnstt-server          # Ready-to-run server binary
│   └── dnstt-client          # Ready-to-run client binary
├── cmd/
│   ├── server/main.go        # Server source code
│   └── client/main.go        # Client source code
├── config/
│   ├── server.conf           # Server configuration example
│   └── client.conf           # Client configuration example
├── build.sh                  # Build script
├── run-server.sh             # Server launcher script
├── run-client.sh             # Client launcher script
├── Makefile                  # Make build automation
├── Dockerfile                # Docker image definition
├── docker-compose.yml        # Docker Compose setup
├── README.md                 # Full documentation
├── QUICKSTART.md             # Quick start guide
└── go.mod / go.sum          # Go dependencies
```

## Support & Documentation

- **README.md**: Full feature documentation
- **QUICKSTART.md**: 5-minute setup guide
- **config/**: Configuration file examples
- **Makefile**: Build targets and commands

## Notes

- All binaries are statically linked for portability
- No runtime dependencies required
-Works on Linux x86_64 systems
- Recommended: Ubuntu 20.04 LTS or newer

---

**Version**: 1.0  
**Ready for Production**: ✓
