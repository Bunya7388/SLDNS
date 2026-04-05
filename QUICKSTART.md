# DNSTT Quick Start Guide

## 5-Minute Setup

### 1. Build the Project

```bash
cd /workspaces/SLDNS
chmod +x build.sh run-server.sh run-client.sh
./build.sh
```

Expected output:
```
[*] Building DNSTT UDP Tunnel...
[*] Building dnstt-server...
[✓] Built: bin/dnstt-server
[*] Building dnstt-client...
[✓] Built: bin/dnstt-client
[✓] Build completed successfully!
```

### 2. Start the Server

Open Terminal 1:

```bash
cd /workspaces/SLDNS
sudo ./run-server.sh
```

Expected output:
```
=== DNSTT UDP Tunnel Server ===
Listen Address: 0.0.0.0:53
Workers: 16
Verbose: false
[*] Starting service...
=== DNSTT UDP Tunnel Server v1.0 ===
Workers: 16
Listen: 0.0.0.0:53
Verbose: false
DNS Server listening on 0.0.0.0:53 with 16 workers
```

### 3. Start the Client

Open Terminal 2:

```bash
cd /workspaces/SLDNS
./run-client.sh
```

Expected output:
```
=== DNSTT UDP Tunnel Client ===
Server: 127.0.0.1:53
Local Listen: 127.0.0.1:8888
Workers: 4
=== DNSTT UDP Tunnel Client v1.0 ===
Server: 127.0.0.1:53
Local Listen: 127.0.0.1:8888
DNSTT Client listening on 127.0.0.1:8888
```

### 4. Test the Tunnel

Open Terminal 3 and use netcat to test:

```bash
# Create test data
echo "Hello DNSTT!" | nc 127.0.0.1 8888

# Monitor server stats
watch -n 1 "netstat -un | grep 53"
```

## Advanced Usage

### High-Performance Server Setup

For maximum throughput, use more workers:

```bash
export LISTEN_ADDR="0.0.0.0:53"
export WORKERS=64
export VERBOSE=true
./run-server.sh
```

### Custom Configuration

Edit configuration files in `config/`:

```bash
# Edit server config
nano config/server.conf

# Source and run
source config/server.conf
./run-server.sh
```

### Using Makefile

```bash
# Build everything
make build

# Run server with Makefile
make run-server

# Run client with Makefile
make run-client

# Clean builds
make clean
```

## Docker Deployment

### Single Container

```bash
# Build Docker image
docker build -t dnstt .

# Run server
docker run -d --name dnstt-server -p 53:53/udp \
  dnstt dnstt-server -l 0.0.0.0:53 -h 64 -v

# Run client  
docker run -d --name dnstt-client -p 8888:8888 \
  dnstt dnstt-client -s host.docker.internal:53 -l 0.0.0.0:8888 -v
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f dnstt-server
docker-compose logs -f dnstt-client

# Stop services
docker-compose down
```

## Performance Testing

### Generate Load

Use `dig` or `nslookup` to test:

```bash
# Single query
dig @127.0.0.1 example.com

# Rapid queries (needs dnsmasq or similar)
for i in {1..1000}; do
  dig @127.0.0.1 query$i.test >/dev/null 2>&1 &
done
wait
```

### Monitor Performance

Server automatically prints statistics every 30 seconds:

```
[STATS] Packets: 10000 | PPS: 333.0 | Bytes: 5242880 | BPS: 174.8 Mbps | Sessions: 42
```

### System Monitoring

```bash
# Monitor CPU and memory
top -p $(pgrep -f dnstt-server)

# Monitor network
iftop -n
nethogs

# Monitor disk I/O (if logging)
iotop
```

## Troubleshooting

### Server won't start on port 53

**Problem:** `bind: permission denied`

**Solution:**
```bash
sudo ./run-server.sh
# OR use higher port
./run-server.sh 8053
```

### Client can't connect to server

**Problem:** `connection refused` or `timeout`

**Solution:**
```bash
# Check server is running
sudo ss -ulnp | grep 53

# Verify firewall
sudo iptables -L -n | grep 53

# Test locally first
./run-client.sh -s 127.0.0.1:53
```

### High memory usage

**Problem:** Server uses too much RAM

**Solution:**
```bash
# Reduce workers
export WORKERS=8
./run-server.sh

# Monitor memory
free -h
```

### No packets showing in stats

**Problem:** Server shows 0 packets processed

**Solution:**
```bash
# Enable verbose logging
export VERBOSE=true
./run-server.sh

# Send test traffic
echo "test" | nc 127.0.0.1 8888
```

## Production Deployment

### Systemd Service

```bash
# Copy to system
sudo cp bin/dnstt-server /usr/local/bin/

# Create service
sudo tee /etc/systemd/system/dnstt.service > /dev/null <<EOF
[Unit]
Description=DNSTT UDP Tunnel Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/dnstt-server -l 0.0.0.0:53 -h 64
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable dnstt
sudo systemctl start dnstt
sudo systemctl status dnstt
```

### System Tuning

For maximum performance, adjust kernel parameters:

```bash
# Increase file descriptors
ulimit -n 65536

# Increase socket buffers
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728

# Increase UDP buffer
sudo sysctl -w net.ipv4.udp_mem="102400 873800 1677600"

# Persist changes
sudo tee -a /etc/sysctl.conf > /dev/null <<EOF
net.core.rmem_max=134217728
net.core.wmem_max=134217728
net.ipv4.udp_mem=102400 873800 1677600
EOF
```

## Next Steps

- Read [README.md](../README.md) for full documentation
- Review [cmd/server/main.go](../cmd/server/main.go) for server implementation
- Review [cmd/client/main.go](../cmd/client/main.go) for client implementation
- Check configuration examples in [config/](../config/)
