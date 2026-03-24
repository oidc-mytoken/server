#!/bin/bash
# Build the Svelte frontend and copy to the Go embed directory

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="$ROOT_DIR/frontend"
SPA_DIST_DIR="$ROOT_DIR/internal/server/spa/dist"

echo "Building Svelte frontend..."
cd "$FRONTEND_DIR"

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

# Build the frontend
echo "Running npm build..."
npm run build

# Clear the old dist directory (but keep .gitkeep)
echo "Clearing old dist directory..."
find "$SPA_DIST_DIR" -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true

# Copy the built files
echo "Copying built files to $SPA_DIST_DIR..."
cp -r "$FRONTEND_DIR/build/"* "$SPA_DIST_DIR/"

echo "Frontend build complete!"
echo "Files copied to: $SPA_DIST_DIR"
echo ""
echo "To use the SPA, add this to your config.yaml:"
echo "  features:"
echo "    web_interface:"
echo "      use_spa: true"
