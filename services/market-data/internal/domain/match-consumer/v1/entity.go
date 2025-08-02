package v1

import (
	"time"
)

// MatchEvent represents a match event from the matching service.
type MatchEvent struct {
	ID          string
	Timestamp   time.Time
	Symbol      string
	Price       float64
	Volume      int64
	BuyOrderID  string
	SellOrderID string
	Exchange    string
	TakerSide   string // "buy" or "sell"
}
