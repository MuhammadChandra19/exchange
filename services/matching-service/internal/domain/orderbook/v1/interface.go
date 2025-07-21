package orderbookv1

import snapshotv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/snapshot/v1"

// Orderbook defines the interface for an order book in a matching service.
type Orderbook interface {
	AskTotalVolume() float64
	Asks() []*Limit
	BidTotalVolume() float64
	Bids() []*Limit
	CancelOrder(orderID string) error
	PlaceLimitOrder(price float64, o *Order) error
	PlaceMarketOrder(o *Order) ([]Match, error)
	CreateSnapshot() *snapshotv1.Snapshot
	RestoreOrderbook(*snapshotv1.Snapshot) error
}
