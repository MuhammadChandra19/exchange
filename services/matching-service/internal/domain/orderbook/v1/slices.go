package orderbookv1

// Limits represents a slice of Limit pointers, representing multiple price levels.
type Limits []*Limit

// ByBestAsk sorts Limits by the best ask price (lowest price).
type ByBestAsk struct {
	Limits
}

func (a ByBestAsk) Len() int {
	return len(a.Limits)
}

func (a ByBestAsk) Less(i, j int) bool {
	return a.Limits[i].Price < a.Limits[j].Price
}

func (a ByBestAsk) Swap(i, j int) {
	a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i]
}

// ByBestBid sorts Limits by the best bid price (highest price).
type ByBestBid struct {
	Limits
}

func (a ByBestBid) Len() int {
	return len(a.Limits)
}
func (a ByBestBid) Less(i, j int) bool {
	return a.Limits[i].Price > a.Limits[j].Price
}
func (a ByBestBid) Swap(i, j int) {
	a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i]
}

// Orders is a slice of Order pointers, representing multiple orders.
type Orders []*Order

func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }
