#!/bin/bash

# DNSTT Server Runner - High-performance DNS tunnel server

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
DNSTT_SERVER="$BIN_DIR/dnstt-server"

# Configuration (override with environment variables)
LISTEN_ADDR="${LISTEN_ADDR:-0.0.0.0:53}"
WORKERS="${WORKERS:-16}"
VERBOSE="${VERBOSE:-false}"
LOG_FILE="${LOG_FILE:-}"

if [ ! -f "$DNSTT_SERVER" ]; then
    echo "[!] Error: dnstt-server binary not found at $DNSTT_SERVER"
    echo "[!] Please run: ./build.sh"
    exit 1
fi

echo "=== DNSTT UDP Tunnel Server ==="
echo "Listen Address: $LISTEN_ADDR"
echo "Workers: $WORKERS"
echo "Verbose: $VERBOSE"
[ -n "$LOG_FILE" ] && echo "Log File: $LOG_FILE"

# Build command
CMD="$DNSTT_SERVER -l $LISTEN_ADDR -h $WORKERS"

if [ "$VERBOSE" = "true" ] || [ "$VERBOSE" = "1" ]; then
    CMD="$CMD -v"
fi

if [ -n "$LOG_FILE" ]; then
    CMD="$CMD -log $LOG_FILE"
fi

# Run server (requires root for port 53)
if [ "$LISTEN_ADDR" = "0.0.0.0:53" ] || [ "$LISTEN_ADDR" = "127.0.0.1:53" ]; then
    if [ $EUID -ne 0 ]; then
        echo "[!] Running on port 53 requires root privileges"
        echo "[*] Use: sudo $0"
        exit 1
    fi
fi

echo "[*] Starting service..."
exec $CMD
