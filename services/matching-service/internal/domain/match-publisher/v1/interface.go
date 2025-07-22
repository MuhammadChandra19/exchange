package matchpublisherv1

import (
	"context"

	pb "github.com/muhammadchandra19/exchange/proto/go/kafka/v1"
)

// MatchPublisher defines the interface for publishing match events.
//
//go:generate mockgen -source interface.go -destination=mock/interface_mock.go -package=matchpublisherv1_mock
type MatchPublisher interface {
	// PublishMatchEvent publishes a match event to the Kafka topic.
	PublishMatchEvent(ctx context.Context, kafkaVayload *pb.MatchEventPayload) error
}
