.PHONY: help generate clean lint format breaking mocks setup post-generate dev ci

# Go module info (matching your script)
GO_MODULE = github.com/muhammadchandra19/exchange/proto/go
MOCK_TAG = mock

# Default target
help:
	@echo "Available targets:"
	@echo "  setup         - Initial setup and install dependencies"
	@echo "  generate      - Generate all protobuf code"
	@echo "  mocks         - Generate mocks for gRPC services"
	@echo "  clean         - Clean generated files"
	@echo "  lint          - Lint proto files"
	@echo "  format        - Format proto files"
	@echo "  breaking      - Check for breaking changes"
	@echo "  post-generate - Run post-generation steps (mod tidy, vendor)"
	@echo "  dev           - Full development workflow"
	@echo "  ci            - CI workflow"

# Setup and install dependencies
setup:
	@echo "Installing dependencies..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@go install github.com/golang/mock/mockgen@latest
	@echo "Dependencies installed!"

# Generate all protobuf code (matches your script workflow)
generate: clean copy-go-mod
	@echo "🚀 Generating protobuf code..."
	
	@echo "📦 Generating with buf..."
	buf generate
	
	@echo "🔧 Generating mocks..."
	@$(MAKE) mocks

	
	@echo "✅ Generation complete!"

# Copy go.mod and go.sum (matching your script)
copy-go-mod:
	@echo "📋 Copying go.mod and go.sum..."
	@if [ -f "../go.mod" ]; then cp ../go.mod go/; fi
	@if [ -f "../go.sum" ]; then cp ../go.sum go/; fi

# Generate mocks (matching your mockgen logic)
mocks:
	@echo "🎭 Generating mocks..."
	@chmod +x scripts/generate-mocks.sh
	@./scripts/generate-mocks.sh



# Clean generated files
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -rf go/core go/common go/modules go/openapi go/ts
	@rm -rf openapi/* ts/*
	@echo "Clean complete!"

# Lint proto files
lint:
	@echo "🔍 Linting proto files..."
	buf lint

# Format proto files
format:
	@echo "✨ Formatting proto files..."
	buf format -w

# Check for breaking changes
breaking:
	@echo "🔒 Checking for breaking changes..."
	buf breaking --against '.git#branch=main'

# Development workflow (matches your script)
dev: format lint generate

# CI workflow
ci: lint breaking generate

# Generate specific parts (like your original script)
generate-core:
	@echo "Generating core..."
	buf generate --path core

generate-common:
	@echo "Generating common..."
	buf generate --path common

generate-modules:
	@echo "Generating modules..."
	buf generate --path modules

# Watch mode for development
watch:
	@echo "👀 Watching for changes..."
	@which fswatch > /dev/null || (echo "Install fswatch for watch mode" && exit 1)
	fswatch -o . --exclude="go/.*" --exclude="openapi/.*" --exclude="ts/.*" | xargs -n1 -I{} $(MAKE) dev
