# Market Data Service

A high-performance market data processing service for cryptocurrency exchanges built in Go. This service ingests order events, processes market data, generates OHLC (Open, High, Low, Close) candles, and provides real-time market data through gRPC APIs.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Database Schema](#database-schema)
- [Migration System](#migration-system)
- [Deployment](#deployment)
- [Development](#development)

## Overview

The Market Data Service is a core component of a cryptocurrency exchange that:

- **Processes Order Events**: Consumes order events from Kafka and stores them in QuestDB
- **Generates Market Data**: Creates ticks and processes trade matches for market analysis
- **OHLC Aggregation**: Generates candlestick data at multiple time intervals (1m, 5m, 15m, 1h, 4h, 1d)
- **Real-time APIs**: Provides gRPC APIs for market data queries
- **Historical Data**: Maintains comprehensive historical trading data
- **Time-series Storage**: Optimized for high-frequency time-series data using QuestDB

## Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Order Events  │    │  Order Consumer │    │    QuestDB      │
│    (Kafka)      │───▶│   (Processing)  │───▶│  (Time-series)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                    ┌─────────────────┐    ┌─────────────────┐
                    │ OHLC Aggregator │    │   gRPC APIs     │
                    │  (Intervals)    │    │  (Tick/Order)   │
                    └─────────────────┘    └─────────────────┘
```

### Service Architecture

- **gRPC Server**: Handles client requests for market data
- **Order Consumer**: Processes order events from the matching service
- **OHLC Engine**: Aggregates tick data into candlestick intervals
- **Repository Layer**: Manages data persistence with QuestDB
- **Migration System**: Database schema versioning and management

## Features

### 🔄 **Real-time Data Processing**
- **Order Event Consumption**: Processes order placed, modified, and cancelled events
- **Tick Generation**: Creates market ticks from trade matches
- **Live Updates**: Real-time market data streaming

### 📊 **Market Data Analytics**
- **OHLC Candlesticks**: Multiple timeframes (1m, 5m, 15m, 1h, 4h, 1d)
- **Volume Analysis**: Trading volume aggregation
- **Price Analysis**: Latest prices and historical data
- **Intraday Data**: Recent trading activity

### 🚀 **High Performance**
- **Time-series Optimized**: QuestDB for fast time-series operations
- **Concurrent Processing**: Goroutine-based order processing
- **Efficient Storage**: Partitioned tables for optimal query performance
- **Indexed Queries**: Symbol and user-based indexing

### 🛠 **Developer Tools**
- **Migration System**: Database schema versioning with UP/DOWN migrations
- **gRPC APIs**: Type-safe protocol buffer APIs
- **Configuration Management**: Environment-based configuration
- **Logging**: Structured logging with context

## Getting Started

### Prerequisites

- Go 1.22+
- QuestDB 8.0+
- Apache Kafka 3.0+
- Docker (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd exchange/services/market-data-service
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run database migrations**
   ```bash
   make migrate
   ```

5. **Start the service**
   ```bash
   # Start gRPC server
   go run cmd/rpc/main.go
   
   # Start order consumer (separate terminal)
   go run cmd/order_consumer/main.go
   ```

## Configuration

Configuration is managed through environment variables:

### Application Settings
```env
APP_NAME=market-data-service
APP_ENVIRONMENT=development
APP_PORT=8080
APP_GRPC_PORT=9090
APP_LOG_LEVEL=info
```

### QuestDB Configuration
```env
QUESTDB_HOST=localhost
QUESTDB_PORT=8812
QUESTDB_DATABASE=exchange
QUESTDB_USER=admin
QUESTDB_PASSWORD=quest
```

### Kafka Configuration
```env
# Order Events
ORDER_KAFKA_BROKERS=localhost:9092
ORDER_KAFKA_TOPIC=orders
ORDER_KAFKA_CONSUMER_GROUP=market-data-service
ORDER_KAFKA_CONSUMER_TIMEOUT=5
ORDER_KAFKA_MAX_RETRIES=3

# Match Events (Future)
MATCH_KAFKA_BROKERS=localhost:9092
MATCH_KAFKA_TOPIC=matches
MATCH_KAFKA_CONSUMER_GROUP=market-data-service
```

## API Reference

### gRPC Services

The service provides two main gRPC services:

#### 1. Tick Service
```protobuf
service TickService {
  rpc GetLatestTick(GetLatestTickRequest) returns (GetLatestTickResponse);
  rpc GetTickVolume(GetTickVolumeRequest) returns (GetTickVolumeResponse);
  rpc GetTicks(GetTicksRequest) returns (GetTicksResponse);
}
```

**Examples:**
- Get latest tick: `symbol: "BTC/USD"`
- Get volume: `symbol: "BTC/USD", from: "2024-01-01T00:00:00Z", to: "2024-01-02T00:00:00Z"`
- Get ticks: `symbol: "BTC/USD", limit: 100, offset: 0`

#### 2. Order Service
```protobuf
service OrderService {
  rpc GetPairActiveOrders(GetPairActiveOrdersRequest) returns (GetPairActiveOrdersResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
  rpc GetOrders(GetOrdersRequest) returns (GetOrdersResponse);
}
```

**Examples:**
- Get active orders: `symbol: "BTC/USD", side: "buy", limit: 50`
- Get order: `order_id: "order_123"`
- Get orders by filter: `symbol: "BTC/USD", user_id: "user_456"`

## Database Schema

### Core Tables

#### 1. **ticks**
```sql
CREATE TABLE ticks (
    timestamp TIMESTAMP,
    symbol SYMBOL CAPACITY 1000 CACHE,
    price DOUBLE,
    volume LONG,
    side SYMBOL CAPACITY 10 CACHE
) TIMESTAMP(timestamp) PARTITION BY HOUR WAL;
```

#### 2. **ohlc**
```sql
CREATE TABLE ohlc (
    timestamp TIMESTAMP,
    symbol SYMBOL CAPACITY 1000 CACHE,
    interval SYMBOL CAPACITY 20 CACHE, -- '1m', '5m', '15m', '1h', '4h', '1d'
    open DOUBLE,
    high DOUBLE,
    low DOUBLE,
    close DOUBLE,
    volume LONG,
    trade_count LONG
) TIMESTAMP(timestamp) PARTITION BY DAY WAL
DEDUP UPSERT KEYS (timestamp, symbol, interval);
```

#### 3. **orders**
```sql
CREATE TABLE orders (
    order_id STRING,
    timestamp TIMESTAMP,
    symbol SYMBOL CAPACITY 1000 CACHE,
    side SYMBOL CAPACITY 10 CACHE,
    price DOUBLE,
    quantity LONG,
    order_type SYMBOL CAPACITY 10 CACHE,
    status SYMBOL CAPACITY 10 CACHE,
    user_id SYMBOL CAPACITY 128
) TIMESTAMP(timestamp) PARTITION BY DAY WAL
DEDUP UPSERT KEYS (timestamp, order_id);
```

#### 4. **order_events**
```sql
CREATE TABLE order_events (
    event_id STRING,
    timestamp TIMESTAMP,
    order_id STRING,
    event_type SYMBOL CAPACITY 20 CACHE,
    symbol SYMBOL CAPACITY 1000 CACHE,
    side SYMBOL CAPACITY 10 CACHE,
    price DOUBLE,
    quantity LONG,
    user_id SYMBOL CAPACITY 128,
    new_price DOUBLE,
    new_quantity LONG
) TIMESTAMP(timestamp) PARTITION BY HOUR WAL;
```

### Performance Optimizations

- **Partitioning**: Tables are partitioned by time (HOUR/DAY) for optimal query performance
- **Symbol Indexing**: All symbol columns are indexed for fast filtering
- **User ID Indexing**: User-specific queries are optimized with SYMBOL type
- **WAL Enabled**: Write-Ahead Log for high-throughput ingestion
- **Deduplication**: Automatic deduplication for orders and OHLC data

## Migration System

The service uses a custom migration system with separate UP/DOWN files:

### Commands
```bash
# Apply all pending migrations
make migrate

# Apply specific number of migrations
make migrate-up steps=3

# Rollback migrations
make migrate-down steps=2

# Create new migration
make migration name=add_new_table
```

### Migration Files
```
migrations/
├── 20250125210734_ini_table.up.sql
├── 20250125210734_ini_table.down.sql
├── 20250125211616_add_indexes.up.sql
└── 20250125211616_add_indexes.down.sql
```

## Deployment

### Docker Deployment
```yaml
version: '3.8'
services:
  market-data-service:
    build: .
    environment:
      - QUESTDB_HOST=questdb
      - ORDER_KAFKA_BROKERS=kafka:9092
    depends_on:
      - questdb
      - kafka
```

### Health Checks
The service provides health check endpoints for monitoring deployment status.

### Scaling Considerations
- **Consumer Groups**: Multiple instances can share order consumption load
- **Database Connections**: Configure connection pooling for high concurrency
- **Memory Management**: Monitor memory usage for large historical datasets

## Development

### Project Structure
```
services/market-data-service/
├── cmd/                     # Application entry points
│   ├── rpc/                # gRPC server
│   ├── order_consumer/     # Order event consumer
│   └── migrate/            # Migration tool
├── internal/               # Internal packages
│   ├── bootstrap/          # Dependency injection
│   ├── consumer/           # Kafka consumers
│   ├── domain/             # Business domain
│   ├── infrastructure/     # External integrations
│   ├── rpc/               # gRPC handlers
│   └── usecase/           # Business logic
├── pkg/                   # Public packages
│   ├── config/           # Configuration
│   └── interval/         # OHLC aggregation
└── Makefile              # Development commands
```

### Testing
```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Generate test coverage
go test -coverprofile=coverage.out ./...
```

### Code Generation
```bash
# Generate protocol buffers
buf generate

# Generate mocks
go generate ./...
```

### Contributing
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

**Built with ❤️ using Go, QuestDB, and gRPC** 