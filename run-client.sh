#!/bin/bash

# DNSTT Client Runner - High-speed DNS tunnel client

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
DNSTT_CLIENT="$BIN_DIR/dnstt-client"

# Configuration (override with environment variables)
SERVER_ADDR="${SERVER_ADDR:-127.0.0.1:53}"
LOCAL_ADDR="${LOCAL_ADDR:-127.0.0.1:8888}"
WORKERS="${WORKERS:-4}"
VERBOSE="${VERBOSE:-false}"

if [ ! -f "$DNSTT_CLIENT" ]; then
    echo "[!] Error: dnstt-client binary not found at $DNSTT_CLIENT"
    echo "[!] Please run: ./build.sh"
    exit 1
fi

echo "=== DNSTT UDP Tunnel Client ==="
echo "Server: $SERVER_ADDR"
echo "Local Listen: $LOCAL_ADDR"
echo "Workers: $WORKERS"
echo "Verbose: $VERBOSE"

# Build command
CMD="$DNSTT_CLIENT -s $SERVER_ADDR -l $LOCAL_ADDR"

if [ "$WORKERS" != "4" ]; then
    CMD="$CMD -w $WORKERS"
fi

if [ "$VERBOSE" = "true" ] || [ "$VERBOSE" = "1" ]; then
    CMD="$CMD -v"
fi

echo "[*] Starting client..."
exec $CMD
