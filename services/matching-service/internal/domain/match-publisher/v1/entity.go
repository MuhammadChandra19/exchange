package matchpublisherv1

import (
	"encoding/json"
	"time"

	pb "github.com/muhammadchandra19/exchange/proto/go/kafka/v1"
	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/orderbook/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateFromMatch creates a match event from a match and an order.
func CreateFromMatch(match *orderbookv1.Match, order *orderbookv1.Order) *pb.MatchEventPayload {
	matchEvent := &pb.MatchEventPayload{
		MatchID:   order.ID,
		Timestamp: timestamppb.New(time.Unix(order.Timestamp, 0)),
	}

	if order.Bid {
		matchEvent.BuyOrderID = order.ID
		matchEvent.SellOrderID = match.Ask.ID
		matchEvent.TakerSide = "buy"
	} else {
		matchEvent.BuyOrderID = match.Bid.ID
		matchEvent.SellOrderID = order.ID
		matchEvent.TakerSide = "sell"
	}

	matchEvent.Volume = match.SizeFilled
	matchEvent.Price = match.Price

	return matchEvent
}

// ToBytes converts the match event to a byte array.
func ToBytes(matchEvent *pb.MatchEventPayload) []byte {
	json, err := json.Marshal(matchEvent)
	if err != nil {
		return nil
	}

	return json
}

// FromBytes converts a byte array to a match event.
func FromBytes(data []byte) *pb.MatchEventPayload {
	var matchEvent pb.MatchEventPayload
	err := json.Unmarshal(data, &matchEvent)
	if err != nil {
		return nil
	}
	return &matchEvent
}
