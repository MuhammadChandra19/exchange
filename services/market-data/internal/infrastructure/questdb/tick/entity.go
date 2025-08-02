package tick

import (
	"time"

	"github.com/muhammadchandra19/exchange/proto/go/modules/market-data/v1/shared"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Tick represents a single tick (price and volume) data point.
type Tick struct {
	Timestamp time.Time
	Symbol    string
	Price     float64
	Volume    int64
	Side      string // "buy" or "sell"
}

// ToProto converts the Tick to a protobuf message.
func (t *Tick) ToProto() *shared.Tick {
	if t == nil {
		return nil
	}

	return &shared.Tick{
		Symbol:    t.Symbol,
		Price:     t.Price,
		Volume:    t.Volume,
		Side:      t.Side,
		Timestamp: timestamppb.New(t.Timestamp).AsTime().Format(time.RFC3339),
	}
}

// Filter represents the filter criteria for tick data.
type Filter struct {
	Symbol string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}
