#!/usr/bin/env bash
set -euo pipefail

gofmt -w .
go build -o ./api-new ./cmd/api

pkill -f './api-new' || true
nohup ./api-new > ./fetch.log 2>&1 & echo $! > ./tmp/api_new.pid
echo "Started api-new, pid: $(cat ./tmp/api_new.pid)"
