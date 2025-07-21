# Build stage
FROM golang:1.24-alpine AS builder

# Install git and other build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.work go.work.sum ./
COPY services/market-data-service/go.mod services/market-data-service/go.sum ./services/market-data-service/
COPY pkg/errors/go.mod pkg/errors/go.sum ./pkg/errors/
COPY pkg/logger/go.mod pkg/logger/go.sum ./pkg/logger/
COPY pkg/util/go.mod pkg/util/go.sum ./pkg/util/

# Download dependencies
RUN cd services/market-data-service && go mod download

# Copy source code
COPY services/market-data-service/ ./services/market-data-service/
COPY pkg/ ./pkg/

# Build the application
RUN cd services/market-data-service && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/market-data-service cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/services/market-data-service/bin/market-data-service .

# Copy migrations
COPY --from=builder /app/services/market-data-service/internal/infrastructure/questdb/migrations ./migrations

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./market-data-service"] 