package orderbookv1

// Match represents a match between an ask and a bid order.
type Match struct {
	Ask        *Order  `json:"ask"`
	Bid        *Order  `json:"bid"`
	SizeFilled float64 `json:"sizeFilled"`
	Price      float64 `json:"price"`
}
