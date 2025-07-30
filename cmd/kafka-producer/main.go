package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// Order represents the order structure
type Order struct {
	OrderID string  `json:"orderID"`
	UserID  string  `json:"userID"`
	Type    string  `json:"type"`
	Bid     bool    `json:"bid"`
	Size    float64 `json:"size"`
	Price   float64 `json:"price"`
	Offset  int64   `json:"offset"`
}

// generateRandomID creates a random alphanumeric ID
func generateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	var result strings.Builder
	for i := 0; i < length; i++ {
		result.WriteByte(charset[rand.Intn(len(charset))])
	}
	return result.String()
}

// generateOrders creates a specified number of realistic orders
func generateOrders(count int, basePrice float64, priceSpread float64) []Order {
	orders := make([]Order, count)

	for i := 0; i < count; i++ {
		// Order types: 70% limit, 30% market
		orderType := "limit"
		if rand.Float64() < 0.3 {
			orderType = "market"
		}

		// Order side: 50/50 buy/sell
		isBid := rand.Float64() < 0.5

		// Generate IDs
		orderID := generateRandomID(rand.Intn(4) + 5) // 5-8 characters
		userID := generateRandomID(rand.Intn(4) + 6)  // 6-9 characters

		// Order size between 0.01 and 10.0
		size := 0.01 + rand.Float64()*9.99
		size = float64(int(size*1000)) / 1000 // Round to 3 decimal places

		// Price calculation
		var price float64
		if orderType == "market" {
			// Market orders can have a price but it's often ignored
			price = basePrice + (rand.Float64()-0.5)*priceSpread
		} else {
			// Limit orders have specific prices
			if isBid { // Buy order - typically below market
				price = basePrice - rand.Float64()*priceSpread*0.8
			} else { // Sell order - typically above market
				price = basePrice + rand.Float64()*priceSpread*0.8
			}
		}
		price = float64(int(price*10)) / 10 // Round to 1 decimal place

		// Ensure price is positive
		if price <= 0 {
			price = basePrice
		}

		// Offset (sequential)
		offset := int64(i + 1)

		orders[i] = Order{
			OrderID: orderID,
			UserID:  userID,
			Type:    orderType,
			Bid:     isBid,
			Size:    size,
			Price:   price,
			Offset:  offset,
		}
	}

	return orders
}

func main() {
	var (
		brokers     = flag.String("brokers", "localhost:9092", "Kafka broker addresses (comma-separated)")
		topic       = flag.String("topic", "orders", "Kafka topic name")
		file        = flag.String("file", "", "JSON file with orders (optional, generates orders if not provided)")
		delay       = flag.Duration("delay", 100*time.Millisecond, "Delay between sending orders")
		count       = flag.Int("count", 1000, "Number of orders to generate")
		basePrice   = flag.Float64("base-price", 3945.5, "Base price for orders")
		priceSpread = flag.Float64("price-spread", 200.0, "Price spread range")
	)
	flag.Parse()

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Create Kafka writer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(*brokers),
		Topic:        *topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
	}
	defer writer.Close()

	ctx := context.Background()

	// Load orders
	var orders []Order
	if *file != "" {
		// Load from file
		data, err := os.ReadFile(*file)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", *file, err)
		}
		if err := json.Unmarshal(data, &orders); err != nil {
			log.Fatalf("Failed to parse JSON from file: %v", err)
		}
		log.Printf("Loaded %d orders from file: %s", len(orders), *file)
	} else {
		// Generate orders
		log.Printf("Generating %d orders...", *count)
		orders = generateOrders(*count, *basePrice, *priceSpread)
		log.Printf("Generated %d orders", len(orders))
	}

	log.Printf("Sending orders to Kafka broker: %s, topic: %s", *brokers, *topic)
	log.Printf("Delay between orders: %v", *delay)

	// Send orders
	for i, order := range orders {
		// Convert order to JSON
		orderJSON, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order %d: %v", i+1, err)
			continue
		}

		// Create Kafka message
		msg := kafka.Message{
			Key:   []byte(order.OrderID),
			Value: orderJSON,
			Time:  time.Now(),
		}

		// Send message
		if err := writer.WriteMessages(ctx, msg); err != nil {
			log.Printf("Failed to send order %d (%s): %v", i+1, order.OrderID, err)
			continue
		}

		// Log progress every 100 orders or for the last order
		if (i+1)%100 == 0 || i == len(orders)-1 {
			side := "SELL"
			if order.Bid {
				side = "BUY"
			}

			if order.Type == "market" {
				log.Printf("Sent order %d/%d: %s | %s | %s %s | Size: %.3f",
					i+1, len(orders), order.OrderID, order.UserID,
					order.Type, side, order.Size)
			} else {
				log.Printf("Sent order %d/%d: %s | %s | %s %s | Size: %.3f @ $%.1f",
					i+1, len(orders), order.OrderID, order.UserID,
					order.Type, side, order.Size, order.Price)
			}
		}

		// Wait before sending next order (except for the last one)
		if i < len(orders)-1 {
			time.Sleep(*delay)
		}
	}

	log.Printf("Successfully sent all %d orders!", len(orders))

	// Print summary
	marketOrders := 0
	limitOrders := 0
	buyOrders := 0
	sellOrders := 0

	for _, order := range orders {
		if order.Type == "market" {
			marketOrders++
		} else {
			limitOrders++
		}
		if order.Bid {
			buyOrders++
		} else {
			sellOrders++
		}
	}

	log.Printf("--- Summary ---")
	log.Printf("Total Orders: %d", len(orders))
	log.Printf("Market Orders: %d", marketOrders)
	log.Printf("Limit Orders: %d", limitOrders)
	log.Printf("Buy Orders: %d", buyOrders)
	log.Printf("Sell Orders: %d", sellOrders)
}
