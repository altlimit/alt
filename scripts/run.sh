#!/bin/sh
# alt — One-line run script for Unix (Linux & macOS)
# Installs alt (if needed) and runs a tool in one command.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.sh | sh -s -- user/repo [args...]
#
# Examples:
#   curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.sh | sh -s -- altlimit/sitegen
#   curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.sh | sh -s -- altlimit/sitegen -serve
#   curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.sh | sh -s -- altlimit/sitegen@v1.0.0

set -e

ALT_BIN="$HOME/.local/share/alt/bin/alt"
BIN_DIR="$HOME/.local/share/alt/bin"

# --- Main ---
if [ $# -eq 0 ]; then
    echo "Usage: curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.sh | sh -s -- user/repo [args...]"
    echo ""
    echo "Examples:"
    echo "  curl ... | sh -s -- altlimit/sitegen"
    echo "  curl ... | sh -s -- altlimit/sitegen -serve"
    echo "  curl ... | sh -s -- altlimit/sitegen@v1.0.0 --help"
    exit 1
fi

# Install alt if it doesn't exist
if [ ! -x "$ALT_BIN" ]; then
    curl -fsSL https://raw.githubusercontent.com/altlimit/alt/main/scripts/install.sh | sh
fi

# Ensure alt is in PATH for this session
export PATH="$BIN_DIR:$PATH"

# Run the tool — all arguments are passed through
exec "$ALT_BIN" run "$@"
