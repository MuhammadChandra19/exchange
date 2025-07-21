package orderbookv1

// Match represents a match between an ask and a bid order.
type Match struct {
	Ask        *Order  `json:"ask"`
	Bid        *Order  `json:"bid"`
	SizeFilled float64 `json:"sizeFilled"`
	Price      float64 `json:"price"`
}

// AskIsFilled checks if the ask order is filled.
func (m *Match) AskIsFilled() bool {
	return m.Ask.Size <= 0
}

// BidIsFilled checks if the bid order is filled
func (m *Match) BidIsFilled() bool {
	return m.Bid.Size <= 0
}
