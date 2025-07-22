#!/bin/bash

set -e

GO_DIR="go"

echo "🔧 Running post-generation steps..."

# Navigate to go directory
cd "$GO_DIR"

# Initialize go.mod if it doesn't exist
if [ ! -f "go.mod" ]; then
    echo "📦 Initializing go.mod..."
    go mod init github.com/muhammadchandra19/exchange/proto/go
fi


echo "✅ Post-generation steps complete!"

# Print statistics
echo ""
echo "📊 Generation Statistics:"
echo "  Go packages: $(find . -name "*.pb.go" | wc -l | tr -d ' ')"
echo "  gRPC services: $(find . -name "*_grpc.pb.go" | wc -l | tr -d ' ')"
echo "  Mock files: $(find . -name "*_mock.go" | wc -l | tr -d ' ')"
echo "  Gateway files: $(find . -name "*.gw.go" | wc -l | tr -d ' ')"
