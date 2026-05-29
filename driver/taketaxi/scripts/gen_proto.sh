#!/bin/bash
# Wrapper: delegates to the thrift-based generator
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/gen_thrift.sh"
