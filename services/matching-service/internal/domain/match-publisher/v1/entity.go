package matchpublisherv1

import (
	"encoding/json"

	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/orderbook/v1"
)

// MatchEvent represents a match event.
type MatchEvent struct {
	ID          string  `json:"id"`
	Timestamp   int64   `json:"timestamp"`
	Symbol      string  `json:"symbol"`
	Price       float64 `json:"price"`
	Volume      float64 `json:"volume"`
	BuyOrderID  string  `json:"buyOrderID"`
	SellOrderID string  `json:"sellOrderID"`
	TakerSide   string  `json:"takerSide"` // "buy" or "sell"
}

// CreateFromMatch creates a match event from a match and an order.
func (m *MatchEvent) CreateFromMatch(match *orderbookv1.Match, order *orderbookv1.Order) {
	m.ID = order.ID
	m.Timestamp = order.Timestamp

	if order.Bid {
		m.BuyOrderID = order.ID
		m.SellOrderID = match.Ask.ID
		m.TakerSide = "buy"
	} else {
		m.BuyOrderID = match.Bid.ID
		m.SellOrderID = order.ID
		m.TakerSide = "sell"
	}

	m.Volume = match.SizeFilled
	m.Price = match.Price

}

// ToBytes converts the match event to a byte array.
func (m *MatchEvent) ToBytes() []byte {
	json, err := json.Marshal(m)
	if err != nil {
		return nil
	}

	return json
}
