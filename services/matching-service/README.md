# Matching Service

A high-performance order matching engine for cryptocurrency exchanges built in Go. This service processes limit and market orders, maintains an orderbook, and generates trade matches with proper price-time priority.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Performance](#performance)
- [Contributing](#contributing)

## Overview

The Matching Service is a core component of a cryptocurrency exchange that:

- **Processes Orders**: Handles limit and market orders from users
- **Maintains Orderbook**: Real-time orderbook with price-time priority
- **Generates Matches**: Creates trade matches when orders can be filled
- **Manages State**: Persistent state management with Redis snapshots
- **Streams Data**: Real-time order processing via Kafka

## Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Order Reader  │    │     Engine      │    │   Orderbook     │
│   (Kafka)       │───▶│   (Processor)   │───▶│   (In-Memory)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                    ┌─────────────────┐    ┌─────────────────┐
                    │ Snapshot Store  │    │   Match Chan    │
                    │   (Redis)       │    │  (Real-time)    │
                    └─────────────────┘    └─────────────────┘
```

### Directory Structure

```
services/matching-service/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── app/
│   │   └── engine/             # Main processing engine
│   ├── domain/
│   │   ├── orderbook/v1/       # Orderbook domain models
│   │   ├── order-reader/v1/    # Order reader interface
│   │   └── snapshot/v1/        # Snapshot domain models
│   └── usecase/
│       ├── orderbook/          # Orderbook implementation
│       ├── order-reader/       # Kafka order reader
│       └── snapshot/           # Redis snapshot store
├── pkg/
│   └── config/                 # Configuration management
└── README.md
```

## Features

### Order Processing
- **Limit Orders**: Orders with specific price levels
- **Market Orders**: Orders that execute immediately at best available price
- **Price-Time Priority**: FIFO matching within price levels
- **Partial Fills**: Orders can be partially filled across multiple matches

### Orderbook Management
- **Real-time Updates**: Live orderbook state
- **Price Levels**: Efficient limit management by price
- **Volume Tracking**: Accurate volume calculations
- **Thread Safety**: Concurrent access protection

### State Management
- **Snapshots**: Periodic Redis snapshots for recovery
- **Offset Tracking**: Kafka offset management
- **State Restoration**: Automatic recovery from snapshots

### Performance
- **In-Memory Processing**: Sub-millisecond order processing
- **Efficient Matching**: O(log n) price level operations
- **Minimal Allocations**: Optimized for low latency
- **Concurrent Safe**: Thread-safe operations

## Getting Started

### Prerequisites

- Go 1.21 or later
- Kafka (for order streaming)
- Redis (for state persistence)
- Docker (optional, for containerized deployment)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd exchange/services/matching-service
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the service**
   ```bash
   go build -o bin/matching-service cmd/main.go
   ```

### Running the Service

1. **Set up configuration**
   ```bash
   export PAIR="BTC-USD"
   export KAFKA_BROKERS="localhost:9092"
   export KAFKA_TOPIC="orders"
   export REDIS_ADDR="localhost:6379"
   ```

2. **Run the service**
   ```bash
   ./bin/matching-service
   ```

### Docker Deployment

```bash
# Build image
docker build -t matching-service .

# Run container
docker run -d \
  --name matching-service \
  -e PAIR=BTC-USD \
  -e KAFKA_BROKERS=kafka:9092 \
  -e REDIS_ADDR=redis:6379 \
  matching-service
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PAIR` | Trading pair (e.g., BTC-USD) | - | Yes |
| `KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` | Yes |
| `KAFKA_TOPIC` | Kafka topic for orders | `orders` | Yes |
| `REDIS_ADDR` | Redis server address | `localhost:6379` | Yes |
| `REDIS_PASSWORD` | Redis password | - | No |
| `REDIS_DB` | Redis database number | `0` | No |
| `SNAPSHOT_INTERVAL` | Snapshot creation interval | `5m` | No |
| `SNAPSHOT_OFFSET_DELTA` | Orders between snapshots | `1000` | No |

### Configuration File

```yaml
# config.yaml
pair: "BTC-USD"
kafka:
  brokers: ["localhost:9092"]
  topic: "orders"
  groupId: "matching-service"
redis:
  addrs: "localhost:6379"
  password: ""
  db: 0
snapshot:
  interval: "5m"
  offsetDelta: 1000
```

## API Reference

### Order Format

Orders are consumed from Kafka in JSON format:

```json
{
  "userID": "user123",
  "type": "limit",
  "bid": true,
  "size": 1.5,
  "price": 50000.0
}
```

#### Order Fields

- `userID` (string): User identifier
- `type` (string): Order type - `"limit"` or `"market"`
- `bid` (boolean): `true` for buy orders, `false` for sell orders
- `size` (float64): Order quantity
- `price` (float64): Order price (required for limit orders)

### Match Output

When orders are matched, the service produces matches:

```json
{
  "ask": {
    "id": "order456",
    "userID": "seller123",
    "size": 0.5,
    "bid": false
  },
  "bid": {
    "id": "order789",
    "userID": "buyer456", 
    "size": 0.0,
    "bid": true
  },
  "sizeFilled": 0.5,
  "price": 50000.0
}
```

### Orderbook State

The orderbook maintains the following state:

```go
type Orderbook struct {
    AskLimits map[float64]*Limit  // price -> ask limit
    BidLimits map[float64]*Limit  // price -> bid limit  
    Orders    map[string]*Order   // orderID -> order
}
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test -cover ./...
```

### Run Specific Tests

```bash
# Test orderbook functionality
go test ./internal/usecase/orderbook -v

# Test limit functionality  
go test ./internal/domain/orderbook/v1 -v

# Test engine
go test ./internal/app/engine -v
```

### Benchmark Tests

```bash
go test -bench=. ./internal/usecase/orderbook
```

## Performance

### Benchmarks

Typical performance on modern hardware:

- **Order Processing**: ~100,000 orders/second
- **Match Generation**: ~50,000 matches/second  
- **Memory Usage**: ~10MB for 100,000 orders
- **Latency**: <1ms for order processing

### Optimization Tips

1. **Batch Processing**: Process multiple orders together
2. **Memory Pools**: Reuse objects to reduce GC pressure
3. **Lock Optimization**: Minimize lock contention
4. **Snapshot Tuning**: Adjust snapshot frequency for your workload

## Monitoring

### Key Metrics

- **Orders Processed**: Total orders processed
- **Matches Generated**: Total matches created
- **Processing Latency**: Order processing time
- **Orderbook Depth**: Number of price levels
- **Memory Usage**: Current memory consumption

### Health Checks

The service exposes health endpoints:

- **Liveness**: `/health/live` - Service is running
- **Readiness**: `/health/ready` - Service is ready to process orders

## Troubleshooting

### Common Issues

1. **Kafka Connection Issues**
   ```
   Error: failed to connect to kafka
   Solution: Check KAFKA_BROKERS configuration
   ```

2. **Redis Connection Issues**
   ```
   Error: failed to connect to redis  
   Solution: Check REDIS_ADDR and credentials
   ```

3. **High Memory Usage**
   ```
   Issue: Memory usage growing continuously
   Solution: Tune snapshot frequency and GC settings
   ```

4. **Slow Processing**
   ```
   Issue: High order processing latency
   Solution: Check for lock contention and optimize orderbook operations
   ```

### Debugging

Enable debug logging:

```bash
export LOG_LEVEL=debug
./bin/matching-service
```

## Contributing

### Development Setup

1. **Install development tools**
   ```bash
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

2. **Run linting**
   ```bash
   golangci-lint run
   ```

3. **Format code**
   ```bash
   goimports -w .
   ```

### Code Style

- Follow standard Go conventions
- Use meaningful variable names
- Add comments for public interfaces
- Write tests for new functionality
- Keep functions small and focused

### Pull Request Process

1. Create feature branch from `main`
2. Make changes with tests
3. Run full test suite
4. Submit pull request with description
5. Address review feedback

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## Support

For questions and support:

- **Issues**: GitHub Issues
- **Documentation**: See `/docs` directory  
- **Examples**: See `/examples` directory

---

**Note**: This service is designed for high-frequency trading environments. Ensure proper testing and monitoring before production deployment. 