#!/usr/bin/env bash
set -euo pipefail

# Simple restart script for local API development
# - kills common running API binaries
# - builds from ./cmd/api
# - starts the binary in background logging to fetch.log

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

BIN=./api-new
BUILD_TARGET=./cmd/api
LOG=fetch.log

echo "Stopping existing API processes (if any)..."
pkill -f "\./api-new" || true
pkill -f "\./tmp" || true
pkill -f "\./api " || true

sleep 1

echo "Building $BUILD_TARGET -> $BIN"
go build -o "$BIN" "$BUILD_TARGET"

chmod +x "$BIN"

echo "Starting $BIN (logging -> $LOG)"
nohup "$BIN" > "$LOG" 2>&1 &
PID=$!

echo $PID > /tmp/begbot_api.pid
echo "Started $BIN with PID $PID"

exit 0
