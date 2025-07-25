# Exchange Platform

A high-performance, microservices-based cryptocurrency exchange platform built with Go. This platform provides real-time order matching, market data processing, and comprehensive trading infrastructure designed for institutional and retail trading environments.

## Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Current Services](#current-services)
- [Shared Packages](#shared-packages)
- [Getting Started](#getting-started)
- [Development](#development)
- [Infrastructure](#infrastructure)
- [TODO & Roadmap](#todo--roadmap)
- [Contributing](#contributing)

## Overview

The Exchange Platform is a modular, scalable cryptocurrency exchange that handles:

- **High-frequency Trading**: Sub-millisecond order matching
- **Real-time Market Data**: Live price feeds, OHLC candles, and trading volumes
- **Event-driven Architecture**: Kafka-based event streaming between services
- **Time-series Analytics**: Historical data storage and analysis
- **Microservices Design**: Independent, scalable service components

### Key Features

- 🚀 **High Performance**: Optimized for low-latency, high-throughput trading
- 📊 **Real-time Data**: Live market data and order book management
- 🔄 **Event Streaming**: Asynchronous communication via Kafka
- 📈 **Time-series Storage**: QuestDB for efficient market data storage
- 🛠 **Developer Friendly**: Comprehensive tooling and documentation
- 🐳 **Container Ready**: Docker and Kubernetes deployment support

## System Architecture

### Service Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Exchange Platform                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  │  Matching       │    │  Market Data    │    │    Future       │
│  │  Service        │───▶│   Service       │    │   Services      │
│  │ (Order Matching)│    │ (Data Processing│    │ (Order Mgmt,    │
│  └─────────────────┘    └─────────────────┘    │  User Mgmt,     │
│           │                       │           │  Custodian)     │
│           ▼                       ▼           └─────────────────┘
│  ┌─────────────────┐    ┌─────────────────┐              │
│  │     Kafka       │    │    QuestDB      │              ▼
│  │ (Event Stream)  │    │ (Time-series)   │    ┌─────────────────┐
│  └─────────────────┘    └─────────────────┘    │   Shared Pkg    │
│           │                       │           │ (Logger, Redis,  │
│           └───────────┬───────────┘           │  QuestDB, etc.) │
│                       ▼                       └─────────────────┘
│           ┌─────────────────────────┐
│           │        Redis            │
│           │    (State & Cache)      │
│           └─────────────────────────┘
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Order Ingestion**: Orders flow into the matching service via Kafka
2. **Order Matching**: Matching service processes orders using in-memory orderbook
3. **Match Events**: Generated matches are published to Kafka
4. **Market Data**: Market data service consumes order/match events
5. **Data Storage**: Processed data stored in QuestDB for analytics
6. **State Management**: Redis used for caching and service state

## Current Services

### 🎯 [Matching Service](services/matching-service/)

High-performance order matching engine with price-time priority.

- **Technology**: Go, Kafka, Redis
- **Performance**: ~40,000 orders/second, sub-100μs latency  
- **Features**: FIFO matching, state recovery, real-time processing
- **Deployment**: One instance per trading pair

```bash
cd services/matching-service
go run cmd/main.go
```

### 📊 [Market Data Service](services/market-data-service/)

Real-time market data processing and API service.

- **Technology**: Go, QuestDB, Kafka, gRPC
- **Features**: OHLC generation, tick processing, historical data
- **APIs**: gRPC services for ticks, orders, and market data
- **Storage**: Time-series optimized with partitioning

```bash
cd services/market-data-service

# Start gRPC server
go run cmd/rpc/main.go

# Start order consumer  
go run cmd/order_consumer/main.go
```

## Shared Packages

The platform includes reusable packages in [`pkg/`](pkg/) for common functionality:

### 🗄 **Database & Storage**
- **[`pkg/questdb/`](pkg/questdb/)**: QuestDB client with connection pooling
- **[`pkg/redis/`](pkg/redis/)**: Redis client with clustering support
- **[`pkg/migration/`](pkg/migration/)**: Database migration system with UP/DOWN support

### 🔧 **Infrastructure**
- **[`pkg/logger/`](pkg/logger/)**: Structured logging with context support
- **[`pkg/errors/`](pkg/errors/)**: Error handling and tracing utilities
- **[`pkg/util/`](pkg/util/)**: Common utilities and helpers

### 🌐 **Communication**
- **[`pkg/grpclib/`](pkg/grpclib/)**: gRPC server utilities and health checks
- **[`pkg/httplib/`](pkg/httplib/)**: HTTP server utilities and middleware

### 📋 **Protocol Definitions**
- **[`proto/`](proto/)**: Protocol Buffer definitions for inter-service communication
- Generated Go code for type-safe service communication
- Kafka event schemas and gRPC service definitions

## Getting Started

### Prerequisites

- **Go 1.22+**
- **Docker & Docker Compose**
- **Apache Kafka 3.0+**
- **QuestDB 8.0+**
- **Redis 7.0+**

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd exchange
   ```

2. **Start infrastructure services**
   ```bash
   docker-compose up -d kafka questdb redis
   ```

3. **Run database migrations**
   ```bash
   cd services/market-data-service
   make migrate
   ```

4. **Start services**
   ```bash
   # Terminal 1: Matching Service (BTC/USD)
   cd services/matching-service
   PAIR=BTC/USD go run cmd/main.go

   # Terminal 2: Market Data gRPC Server
   cd services/market-data-service  
   go run cmd/rpc/main.go

   # Terminal 3: Market Data Consumer
   cd services/market-data-service
   go run cmd/order_consumer/main.go
   ```

### Environment Configuration

Create `.env` files in each service directory:

```env
# services/matching-service/.env
PAIR=BTC/USD
KAFKA_TOPIC=orders
KAFKA_BROKER=localhost:9092
REDIS_ADDRESS=localhost:6379

# services/market-data-service/.env  
APP_GRPC_PORT=9090
QUESTDB_HOST=localhost
QUESTDB_PORT=8812
ORDER_KAFKA_BROKERS=localhost:9092
ORDER_KAFKA_TOPIC=orders
```

## Development

### Project Structure

```
exchange/
├── services/                    # Microservices
│   ├── matching-service/       # Order matching engine
│   ├── market-data-service/    # Market data processing
│   └── [future-services]/      # Future service implementations
├── pkg/                        # Shared packages
│   ├── questdb/               # Database clients
│   ├── redis/                 # Cache clients  
│   ├── logger/                # Logging utilities
│   ├── migration/             # Database migrations
│   └── [other-packages]/      # Utility packages
├── proto/                      # Protocol definitions
│   ├── core/                  # Core message types
│   ├── kafka/                 # Kafka event schemas
│   └── modules/               # Service-specific APIs
├── dev/                       # Development tools
├── docker-compose.yaml        # Local development stack
└── README.md                  # This file
```

### Development Workflow

1. **Setup Development Environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Start local infrastructure
   docker-compose up -d
   ```

2. **Code Generation**
   ```bash
   # Generate protocol buffers
   cd proto && make generate
   
   # Generate mocks
   go generate ./...
   ```

3. **Testing**
   ```bash
   # Run all tests
   go test ./...
   
   # Run with coverage
   go test -coverprofile=coverage.out ./...
   ```

4. **Service Development**
   ```bash
   # Create new service
   mkdir services/new-service
   cd services/new-service
   go mod init github.com/muhammadchandra19/exchange/services/new-service
   ```

### Code Style & Standards

- **Go Modules**: Each service has its own `go.mod`
- **Clean Architecture**: Domain-driven design with clear layer separation
- **Error Handling**: Use `pkg/errors` for consistent error tracing
- **Logging**: Structured logging with `pkg/logger`
- **Testing**: Unit tests with >80% coverage requirement
- **Documentation**: README for each service and package

## Infrastructure

### Local Development

```bash
# Start all infrastructure services
docker-compose up -d

# Services available:
# - Kafka: localhost:9092
# - QuestDB: localhost:9000 (Web Console), localhost:8812 (PostgreSQL)
# - Redis: localhost:6379
# - Zookeeper: localhost:2181
```

### Production Deployment

- **Container Orchestration**: Kubernetes manifests in `/k8s`
- **Service Mesh**: Istio support for traffic management
- **Monitoring**: Prometheus metrics and Grafana dashboards
- **Logging**: Centralized logging with ELK stack
- **CI/CD**: GitHub Actions workflows for automated deployment

## TODO & Roadmap

### 🚧 **Immediate Priority**

- [ ] **Order Management Service**
  - Order lifecycle management (place, modify, cancel)
  - Order validation and risk checks
  - User order history and portfolio tracking
  - Integration with matching service

- [ ] **User Management Service**
  - User registration and authentication
  - KYC/AML verification workflows
  - User profile and preferences management
  - API key management for trading

- [ ] **Custodian Service**
  - Digital asset custody and wallet management
  - Multi-signature wallet support
  - Deposit and withdrawal processing
  - Cold storage integration

### 🔄 **Core Trading Infrastructure**

- [ ] **Risk Management Service**
  - Real-time risk assessment
  - Position limits and margin management
  - Circuit breakers and trading halts
  - Compliance monitoring

- [ ] **Settlement Service**
  - Trade settlement and clearing
  - Netting and position reconciliation
  - Settlement instruction management
  - Failed trade handling

- [ ] **Notification Service**
  - Real-time alerts and notifications
  - Email, SMS, and push notifications
  - Trading event notifications
  - System status alerts

### 📊 **Analytics & Reporting**

- [ ] **Analytics Service**
  - Advanced market analytics
  - Trading performance metrics
  - Risk analytics and reporting
  - Market microstructure analysis

- [ ] **Reporting Service**
  - Regulatory reporting (MiFID II, etc.)
  - Tax reporting and document generation
  - Audit trail and compliance reports
  - Custom report generation

### 🌐 **Client Interfaces**

- [ ] **REST API Gateway**
  - Unified API gateway for client access
  - Rate limiting and throttling
  - API versioning and documentation
  - Authentication and authorization

- [ ] **WebSocket API**
  - Real-time market data streaming
  - Order status updates
  - Account balance updates
  - Trading alerts

- [ ] **FIX Protocol Gateway**
  - FIX 4.2/4.4/5.0 support
  - Institutional client connectivity
  - Order routing and execution reports
  - Drop copy and trade reporting

### 🔐 **Security & Compliance**

- [ ] **Identity & Access Management**
  - Single sign-on (SSO) integration
  - Multi-factor authentication (MFA)
  - Role-based access control (RBAC)
  - API security and rate limiting

- [ ] **Compliance Service**
  - AML transaction monitoring
  - Sanctions screening
  - Suspicious activity reporting
  - Regulatory compliance automation

- [ ] **Audit Service**
  - Immutable audit trail
  - Compliance monitoring
  - Trade surveillance
  - Regulatory reporting

### 🏗 **Infrastructure & Operations**

- [ ] **Configuration Service**
  - Centralized configuration management
  - Dynamic configuration updates
  - Environment-specific settings
  - Feature flags and toggles

- [ ] **Monitoring & Observability**
  - Distributed tracing
  - Application performance monitoring
  - Custom metrics and dashboards
  - Health check aggregation

- [ ] **Backup & Recovery**
  - Automated backup systems
  - Disaster recovery procedures
  - Data replication strategies
  - Business continuity planning

### 💰 **Business Features**

- [ ] **Fee Management Service**
  - Dynamic fee calculation
  - Maker/taker fee models
  - Volume-based fee tiers
  - Fee distribution and accounting

- [ ] **Liquidity Management**
  - Market maker incentives
  - Liquidity provider programs
  - External liquidity aggregation
  - Smart order routing

- [ ] **Trading Tools**
  - Algorithmic trading support
  - Order types (stop-loss, iceberg, etc.)
  - Portfolio management tools
  - Trading strategy backtesting

## Contributing

We welcome contributions to the Exchange Platform! Here's how to get started:

### Development Setup

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/your-username/exchange.git
   cd exchange
   ```

3. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

4. **Make your changes**
   - Follow the code style guidelines
   - Add tests for new functionality
   - Update documentation as needed

5. **Test your changes**
   ```bash
   go test ./...
   go test -race ./...
   ```

6. **Submit a pull request**
   - Provide a clear description of the changes
   - Include any relevant issue numbers
   - Ensure all CI checks pass

### Code Review Process

- All changes require peer review
- Automated tests must pass
- Code coverage should not decrease
- Documentation must be updated for public APIs

### Community Guidelines

- Be respectful and inclusive
- Follow the project's code of conduct
- Provide constructive feedback in reviews
- Help others learn and grow

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: Check service-specific READMEs
- **Issues**: Report bugs and feature requests via GitHub Issues
- **Discussions**: Join technical discussions in GitHub Discussions
- **Contact**: Reach out to maintainers for critical issues

---

**Built with ❤️ for the future of decentralized finance**
