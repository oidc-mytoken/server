#!/bin/bash
# Build the frontend and run the mytoken server for development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Parse arguments
SKIP_FRONTEND=false
CONFIG_FILE=""

print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -s, --skip-frontend    Skip frontend build (use existing build)"
    echo "  -c, --config FILE      Path to config file (default: uses server default)"
    echo "  -h, --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                     # Build frontend and run server"
    echo "  $0 -s                  # Skip frontend build, just run server"
    echo "  $0 -c /path/to/config  # Use specific config file"
    return 0
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--skip-frontend)
            SKIP_FRONTEND=true
            shift
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            print_usage
            exit 1
            ;;
    esac
done

# Build frontend unless skipped
if [[ "$SKIP_FRONTEND" = false ]]; then
    echo "=== Building frontend ==="
    "$SCRIPT_DIR/build-frontend.sh"
    echo ""
fi

# Build and run the server
echo "=== Building mytoken server ==="
cd "$ROOT_DIR"
go build -o "$ROOT_DIR/mytoken-server" ./cmd/mytoken-server

echo ""
echo "=== Starting mytoken server ==="

if [[ -n "$CONFIG_FILE" ]]; then
    exec "$ROOT_DIR/mytoken-server" --config "$CONFIG_FILE"
else
    exec "$ROOT_DIR/mytoken-server"
fi
