#!/bin/bash

# UDPGW Server - UDP Gateway with port forwarding

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
SERVER="$BIN_DIR/udpgw-server"

LISTEN="${LISTEN:-0.0.0.0:5555}"
REMOTE="${REMOTE:-127.0.0.1:53}"
WORKERS="${WORKERS:-8}"
VERBOSE="${VERBOSE:-false}"
TIMEOUT="${TIMEOUT:-300}"

if [ ! -f "$SERVER" ]; then
    echo "[!] Error: udpgw-server not found at $SERVER"
    exit 1
fi

echo "=== UDPGW UDP Gateway Server ==="
echo "Listen: $LISTEN"
echo "Remote: $REMOTE"
echo "Workers: $WORKERS"
echo "Timeout: $TIMEOUT seconds"

CMD="$SERVER -l $LISTEN -r $REMOTE -w $WORKERS -t $TIMEOUT"

[ "$VERBOSE" = "true" ] || [ "$VERBOSE" = "1" ] && CMD="$CMD -v"

echo "[*] Starting..."
exec $CMD
