#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"
kitex -module driver -gen-path taketaxi/common/kitexGen taketaxi/common/idl/driver.thrift
echo "Done"
