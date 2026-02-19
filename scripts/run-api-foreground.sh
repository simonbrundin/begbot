#!/usr/bin/env bash
set -euo pipefail

gofmt -w .
go build -o ./api-new ./cmd/api

./api-new 2>&1 | tee ./fetch.log
