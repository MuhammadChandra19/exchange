package snapshotv1

// Snapshot represents a snapshot of the order book at a specific point in time.
type Snapshot struct {
	OrderOffset       int64             `json:"orderOffset"`
	OrderBookSnapshot OrderBookSnapshot `json:"orderBookSnapshot"`
}

// OrderBookSnapshot represents the state of the order book at a specific point in time.
type OrderBookSnapshot struct {
	Orders        []BookOrder `json:"orders"`
	TradeSequence int64       `json:"tradeSequence"`
	LogSequence   int64       `json:"logSequence"`
}

// BookOrder represents an order in the order book with its details.
type BookOrder struct {
	OrderID   string  `json:"orderID"`
	Size      float64 `json:"size"`
	Bid       bool    `json:"bid"`
	Price     float64 `json:"price"`
	UserID    string  `json:"userID"`
	Timestamp int64   `json:"timestamp"`
}
