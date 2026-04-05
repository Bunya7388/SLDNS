#!/bin/bash

# DNSTT Build Script - Builds both server and client

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$PROJECT_DIR/bin"

echo "[*] Building DNSTT UDP Tunnel..."
echo "[*] Project directory: $PROJECT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Download dependencies
echo "[*] Downloading dependencies..."
cd "$PROJECT_DIR"
go mod download 2>/dev/null || true

# Build server
echo "[*] Building dnstt-server..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o "$OUTPUT_DIR/dnstt-server" \
    ./cmd/server

chmod +x "$OUTPUT_DIR/dnstt-server"
echo "[✓] Built: $OUTPUT_DIR/dnstt-server"

# Build client
echo "[*] Building dnstt-client..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o "$OUTPUT_DIR/dnstt-client" \
    ./cmd/client

chmod +x "$OUTPUT_DIR/dnstt-client"
echo "[✓] Built: $OUTPUT_DIR/dnstt-client"

echo ""
echo "[✓] Build completed successfully!"
echo ""
echo "Usage:"
echo "  Server: $OUTPUT_DIR/dnstt-server -l 0.0.0.0:53 -h 16 -v"
echo "  Client: $OUTPUT_DIR/dnstt-client -s SERVER_IP:53 -l 127.0.0.1:8888 -v"
