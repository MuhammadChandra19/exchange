package matchpublisher

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	pb "github.com/muhammadchandra19/exchange/proto/go/kafka/v1"
	matchpublisherv1 "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/match-publisher/v1"
	"github.com/muhammadchandra19/exchange/services/matching-engine/pkg/config"
	"github.com/segmentio/kafka-go"
)

// Publisher represents a Kafka Publisher for publishing match events.
type Publisher struct {
	kafkaWriter *kafka.Writer
	logger      logger.Logger
}

// NewPublisher creates a new Kafka publisher for publishing match events.
func NewPublisher(config config.MatchPublisherConfig, logger logger.Logger) *Publisher {
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: config.Brokers,
		Topic:   config.Topic,
	})

	return &Publisher{
		kafkaWriter: kafkaWriter,
		logger:      logger,
	}
}

// PublishMatchEvent publishes a match event to the Kafka topic.
func (p *Publisher) PublishMatchEvent(ctx context.Context, matchEvent *pb.MatchEventPayload) error {
	msg := kafka.Message{
		Value: matchpublisherv1.ToBytes(matchEvent),
	}

	if err := p.kafkaWriter.WriteMessages(ctx, msg); err != nil {
		p.logger.Error(err,
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "matchEvent", Value: matchEvent},
		)
		return errors.NewTracer("failed to publish match event")
	}
	return nil
}
