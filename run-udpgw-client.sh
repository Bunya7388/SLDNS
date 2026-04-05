#!/bin/bash

# UDPGW Client - UDP tunneling client

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
CLIENT="$BIN_DIR/udpgw-client"

LISTEN="${LISTEN:-127.0.0.1:5555}"
REMOTE="${REMOTE:-192.168.1.1:5555}"
VERBOSE="${VERBOSE:-false}"

if [ ! -f "$CLIENT" ]; then
    echo "[!] Error: udpgw-client not found at $CLIENT"
    exit 1
fi

echo "=== UDPGW UDP Gateway Client ==="
echo "Listen: $LISTEN"
echo "Remote: $REMOTE"

CMD="$CLIENT -l $LISTEN -r $REMOTE"

[ "$VERBOSE" = "true" ] || [ "$VERBOSE" = "1" ] && CMD="$CMD -v"

echo "[*] Starting..."
exec $CMD
