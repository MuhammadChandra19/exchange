#!/bin/bash

set -e

echo "🎭 Generating mocks for gRPC services..."

GO_DIR="go"
MOCK_PACKAGE="mock"

# Check if mockgen is installed
if ! which mockgen > /dev/null; then
    echo "❌ mockgen not found. Installing..."
    go install github.com/golang/mock/mockgen@latest
fi

# Function to generate mock for a gRPC service file
generate_mock_for_file() {
    local grpc_file="$1"
    local relative_path="${grpc_file#${GO_DIR}/}"
    local dir_path="$(dirname "$grpc_file")"
    local filename="$(basename "$grpc_file")"
    local mock_dir="${dir_path}/mock"
    
    # Extract base name without _grpc.pb.go suffix
    local base_name="${filename%_grpc.pb.go}"
    local mock_file="${mock_dir}/${base_name}_mock.go"
    
    echo "  📝 Generating mock for $relative_path..."
    
    # Create mock directory
    mkdir -p "$mock_dir"
    
    # Generate mock
    mockgen -source="$grpc_file" -package="$MOCK_PACKAGE" > "$mock_file"
    
    echo "    ✅ Created: ${mock_file#${GO_DIR}/}"
}

# Remove existing mocks
echo "🧹 Cleaning existing mocks..."
find "$GO_DIR" -name "mock" -type d -exec rm -rf {} + 2>/dev/null || true

echo "🔍 Scanning for gRPC service files..."

# Generate mocks for core
if [ -d "$GO_DIR/core" ]; then
    echo "📦 Processing core..."
    find "$GO_DIR/core" -name "*_grpc.pb.go" | while read -r grpc_file; do
        generate_mock_for_file "$grpc_file"
    done
fi

# Generate mocks for common  
if [ -d "$GO_DIR/common" ]; then
    echo "📦 Processing common..."
    find "$GO_DIR/common" -name "*_grpc.pb.go" | while read -r grpc_file; do
        generate_mock_for_file "$grpc_file"
    done
fi

# Generate mocks for modules (matches your nested loop structure)
if [ -d "$GO_DIR/modules" ]; then
    echo "📦 Processing modules..."
    
    for module_path in "$GO_DIR/modules"/*; do
        if [ -d "$module_path" ]; then
            module=$(basename "$module_path")
            echo "  🔧 Processing module: $module"
            
            for version_path in "$module_path"/*; do
                if [ -d "$version_path" ]; then
                    version=$(basename "$version_path")
                    echo "    📋 Processing version: $version"
                    
                    for domain_path in "$version_path"/*; do
                        if [ -d "$domain_path" ]; then
                            domain=$(basename "$domain_path")
                            echo "      🎯 Processing domain: $domain"
                            
                            # Find all gRPC files in this domain
                            find "$domain_path" -name "*_grpc.pb.go" | while read -r grpc_file; do
                                generate_mock_for_file "$grpc_file"
                            done
                        fi
                    done
                fi
            done
        fi
    done
fi

echo ""
echo "✅ Mock generation complete!"

# Count generated mocks
mock_count=$(find "$GO_DIR" -name "*_mock.go" | wc -l | tr -d ' ')
echo "📊 Generated $mock_count mock files"
