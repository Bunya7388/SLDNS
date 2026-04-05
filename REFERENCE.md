# SLDNS Quick Reference Card

## File Locations

| Component | Location | Type |
|-----------|----------|------|
| Server Binary | `bin/dnstt-server` | Executable (2.2 MB) |
| Client Binary | `bin/dnstt-client` | Executable (2.3 MB) |
| Source Code | `cmd/server/main.go`, `cmd/client/main.go` | Go Source |
| Documentation | `README.md`, `QUICKSTART.md`, `DEPLOY.md` | Markdown |

## Quick Commands

### Start Server (requires sudo)
```bash
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 16 -v
```

### Start Client
```bash
./bin/dnstt-client -s 127.0.0.1:53 -l 127.0.0.1:8888 -v
```

### Build from Source
```bash
./build.sh
# or
make build
```

## Common Configurations

### Local Testing
```bash
# Terminal 1
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 4

# Terminal 2
./bin/dnstt-client -s 127.0.0.1:53 -l 127.0.0.1:8888
```

### Production Server (Medium Load)
```bash
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 32 -log /var/log/dnstt.log
```

### Production Server (High Load)
```bash
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 128 -log /var/log/dnstt.log -v
```

### Remote Client
```bash
./bin/dnstt-client -s 192.168.1.100:53 -l 127.0.0.1:8888 -w 8
```

## Flag Reference

### Server (`dnstt-server`)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-l` | string | `0.0.0.0:53` | Listen address:port |
| `-h` | int | `4` | Worker threads (1-256) ⭐ |
| `-v` | bool | false | Verbose logging |
| `-log` | string | (none) | Log file path |
| `-stats` | int | 30 | Stats interval (seconds) |

### Client (`dnstt-client`)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-s` | string | `127.0.0.1:53` | Server address:port |
| `-l` | string | `127.0.0.1:8888` | Local listen address |
| `-w` | int | `4` | Worker threads |
| `-v` | bool | false | Verbose logging |

## Environment Variables

### Server
```bash
export LISTEN_ADDR="0.0.0.0:53"
export WORKERS=32
export VERBOSE=true
export LOG_FILE="/var/log/dnstt.log"
./run-server.sh
```

### Client
```bash
export SERVER_ADDR="192.168.1.100:53"
export LOCAL_ADDR="127.0.0.1:8888"
export WORKERS=8
export VERBOSE=true
./run-client.sh
```

## Make Targets

```bash
make build          # Build server & client
make server         # Build server only
make client         # Build client only
make clean          # Remove binaries
make deps           # Download dependencies
make run-server     # Build and run server
make run-client     # Build and run client
```

## Docker Commands

### Build Image
```bash
docker build -t dnstt .
```

### Run Server
```bash
docker run -d --name dnstt-server -p 53:53/udp \
  dnstt dnstt-server -l 0.0.0.0:53 -h 64
```

### Run Client
```bash
docker run -d --name dnstt-client -p 8888:8888 \
  dnstt dnstt-client -s dnstt-server:53 -l 0.0.0.0:8888
```

### Docker Compose
```bash
docker-compose up -d        # Start all
docker-compose down         # Stop all
docker-compose logs -f      # View logs
```

## Worker Count Guide

| Scenario | Workers | Notes |
|----------|---------|-------|
| Testing | 2-4 | Minimal resources |
| Small Site | 4-8 | Low traffic |
| Medium Site | 16-32 | Standard production |
| High Traffic | 64-128 | Heavy load |
| Very High | 128-256 | Maximum performance |

## Monitoring

### View Live Statistics
```bash
# If verbose enabled
tail -f /var/log/dnstt.log | grep STATS
```

### Check if Server Running
```bash
sudo ss -ulnp | grep 53
```

### Monitor Resource Usage
```bash
top -p $(pgrep -f dnstt-server)
```

### Network Traffic
```bash
nethogs
```

## Troubleshooting

### Server won't start - port in use
```bash
sudo lsof -i :53
sudo kill -9 <PID>
```

### Permission denied on port 53
```bash
sudo ./bin/dnstt-server -l 0.0.0.0:53 -h 16
```

### High memory usage
```bash
# Reduce workers
./bin/dnstt-server -h 8
```

### No data flowing
```bash
# Enable verbose mode
./bin/dnstt-server -h 16 -v
# Check stats output
```

## Performance Tuning

### Linux System Optimization
```bash
# Increase file descriptors
ulimit -n 65536

# Increase socket buffers
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728
```

### Server Tuning
- Start with `-h 16` or `-h 32`
- Increase if CPU < 80% usage
- Decrease if memory grows too fast
- Monitor PPS (packets/sec) in output

## Documentation Map

| Document | Purpose | Audience |
|----------|---------|----------|
| `README.md` | Complete feature guide | Everyone |
| `QUICKSTART.md` | 5-minute setup | New users |
| `DEPLOY.md` | Production deployment | DevOps/Admins |
| `SUMMARY.md` | Project overview | Project managers |
| `config/` | Configuration templates | Sysadmins |

## Binary Information

```
File: dnstt-server
Size: 2.2 MB
Type: Linux x86_64 ELF
Strip: Yes (optimized)
Architecture: 64-bit

File: dnstt-client  
Size: 2.3 MB
Type: Linux x86_64 ELF
Strip: Yes (optimized)
Architecture: 64-bit
```

## Features Implemented

✅ High-speed UDP tunneling  
✅ Configurable worker pools (1-256)  
✅ Session management & cleanup  
✅ Real-time statistics  
✅ Automatic error handling  
✅ Graceful shutdown  
✅ Docker support  
✅ Multiple deployment options  

## More Information

- See `README.md` for full documentation
- See `QUICKSTART.md` for step-by-step guide
- See `DEPLOY.md` for production setup
- Check `config/` for configuration examples
- Review `cmd/` for source code

---

**Version**: 1.0  
**Status**: Production Ready ✓
