#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../srvDriver" && go build -o ../../bin/srvDriver ./cmd/main.go
echo "Build done"
