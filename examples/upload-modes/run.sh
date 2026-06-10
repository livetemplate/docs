#!/bin/bash
# Run the upload-modes example.
cd "$(dirname "$0")"

echo "🚀 Starting Upload Modes example..."
echo "📍 http://localhost:8087"
echo ""

PORT=8087 go run main.go
