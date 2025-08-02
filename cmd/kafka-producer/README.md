# Simple Kafka Order Producer

A simple Go script to send trading orders to Kafka. This tool generates realistic orders dynamically with configurable parameters.

## Usage

### Quick Start

```bash
# Generate and send 1000 orders with default settings
go run main.go

# Custom number of orders
go run main.go -count 2000

# Custom Kafka settings
go run main.go -brokers localhost:9092 -topic orders

# Custom price settings
go run main.go -base-price 4000 -price-spread 300

# Faster sending (50ms delay between orders)
go run main.go -delay 50ms

# Send orders from a JSON file
go run main.go -file orders.json
```

### Build and Run

```bash
# Build the binary
go build -o kafka-producer main.go

# Run the binary
./kafka-producer -brokers localhost:9092 -topic orders
```

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-brokers` | `localhost:9092` | Kafka broker addresses |
| `-topic` | `orders` | Kafka topic name |
| `-count` | `1000` | Number of orders to generate |
| `-base-price` | `3945.5` | Base price for orders |
| `-price-spread` | `200.0` | Price spread range (±) |
| `-delay` | `100ms` | Delay between sending orders |
| `-file` | - | JSON file with orders (optional) |

## Sample Output

```
2024/07/29 15:45:00 Generating 1000 orders...
2024/07/29 15:45:00 Generated 1000 orders
2024/07/29 15:45:00 Sending orders to Kafka broker: localhost:9092, topic: orders
2024/07/29 15:45:00 Delay between orders: 100ms
2024/07/29 15:45:10 Sent order 100/1000: k7m3p | x9j2k8m | limit BUY | Size: 2.456 @ $3890.0
2024/07/29 15:45:20 Sent order 200/1000: n4q8z | p5k9j2n | market SELL | Size: 0.834
2024/07/29 15:45:30 Sent order 300/1000: r2t5w | m7k1j4p | limit SELL | Size: 1.267 @ $4020.5
...
2024/07/29 15:47:40 Sent order 1000/1000: z9x3k | q8m5n2j | market BUY | Size: 3.721
2024/07/29 15:47:40 Successfully sent all 1000 orders!
2024/07/29 15:47:40 --- Summary ---
2024/07/29 15:47:40 Total Orders: 1000
2024/07/29 15:47:40 Market Orders: 298
2024/07/29 15:47:40 Limit Orders: 702
2024/07/29 15:47:40 Buy Orders: 487
2024/07/29 15:47:40 Sell Orders: 513
```

## JSON File Format

If you want to use your own orders, create a JSON file like this:

```json
[
  {
    "orderID": "abc123",
    "userID": "user001",
    "type": "limit",
    "bid": true,
    "size": 1.5,
    "price": 4000.0,
    "offset": 1
  },
  {
    "orderID": "def456",
    "userID": "user002",
    "type": "market",
    "bid": false,
    "size": 2.0,
    "price": 3945.5,
    "offset": 2
  }
]
```

Then run: `go run main.go -file your-orders.json`

## Integration

This tool works perfectly with your matching service:

1. **Start your matching service:**
```bash
cd services/matching-engine
export KAFKA_TOPIC=orders
export KAFKA_BROKER=localhost:9092  
export PAIR=BTC-USD
go run cmd/main.go
```

2. **Send orders:**
```bash
cd cmd/kafka-producer
go run main.go -topic orders
```

3. **Watch the matching happen in the matching service logs!**

## Features

- ✅ **Generate 1000+ orders dynamically** (configurable count)
- ✅ **Realistic order distribution** (70% limit, 30% market, 50/50 buy/sell)
- ✅ **Configurable pricing** (base price and spread range)
- ✅ **Configurable delay** between orders  
- ✅ **Load from JSON file** option
- ✅ **Progress logging** every 100 orders
- ✅ **Summary statistics** showing distribution
- ✅ **Simple and lightweight** - just one Go file
- ✅ **Error handling** for failed sends
- ✅ **Ready to use** with your existing services 