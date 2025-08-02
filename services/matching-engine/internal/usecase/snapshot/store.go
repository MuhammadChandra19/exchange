package snapshot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	logger "github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/redis"
	snapshotv1 "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/snapshot/v1"
)

// Store represents a snapshot of the order book.
type Store struct {
	pair        string
	logger      *logger.Logger
	redisclient redis.Client
}

// NewSnapshotStore creates a new Snapshot instance with the given Redis client and pair.
func NewSnapshotStore(redisclient redis.Client, pair string, logger *logger.Logger) *Store {
	return &Store{
		pair:        pair,
		redisclient: redisclient,
		logger:      logger,
	}
}

// Store stores the snapshot in Redis.
func (s *Store) Store(ctx context.Context, snapshot *snapshotv1.Snapshot) error {
	// Serialize the snapshot and store it in Redis.
	s.logger.InfoContext(ctx, fmt.Sprintf("Storing snapshot for pair %s", s.pair), logger.Field{
		Key:   "pair",
		Value: s.pair,
	})

	buf, err := json.Marshal(snapshot)
	if err != nil {
		s.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "pair",
			Value: s.pair,
		}, logger.Field{
			Key:   "snapshot",
			Value: snapshot,
		})
		return errors.NewTracer("snapshot_marshal_error").Wrap(err)
	}

	err = s.redisclient.Set(ctx, s.pair, buf, 0)
	if err != nil {
		s.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "pair",
			Value: s.pair,
		}, logger.Field{
			Key:   "snapshot",
			Value: snapshot,
		})

		return errors.NewTracer("snapshot_store_error").Wrap(err)
	}
	s.logger.InfoContext(ctx, fmt.Sprintf("Snapshot stored for pair %s", s.pair), logger.Field{
		Key:   "pair",
		Value: s.pair,
	}, logger.Field{
		Key:   "action",
		Value: "store snapshot",
	})
	return nil
}

// LoadStore loads the snapshot from Redis.
func (s *Store) LoadStore(ctx context.Context) (*snapshotv1.Snapshot, error) {
	s.logger.InfoContext(ctx, fmt.Sprintf("Loading snapshot for pair %s", s.pair), logger.Field{
		Key:   "pair",
		Value: s.pair,
	}, logger.Field{
		Key:   "action",
		Value: "load snapshot",
	})
	// Deserialize the snapshot from Redis.
	data, err := s.redisclient.Get(ctx, s.pair)
	if err != nil {
		s.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "pair",
			Value: s.pair,
		}, logger.Field{
			Key:   "action",
			Value: "load snapshot",
		})
		return nil, errors.NewTracer("snapshot_load_error").Wrap(err)
	}

	if data == "" {
		s.logger.WarnContext(ctx, fmt.Sprintf("No snapshot found for pair %s", s.pair), logger.Field{
			Key:   "pair",
			Value: s.pair,
		}, logger.Field{
			Key:   "action",
			Value: "load snapshot",
		})
		return nil, nil
	}

	// Assuming the data can be unmarshalled into a *snapshotv1.Snapshot.
	var snapshot snapshotv1.Snapshot
	if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
		s.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "pair",
			Value: s.pair,
		}, logger.Field{
			Key:   "action",
			Value: "unmarshal snapshot",
		})
		return nil, errors.NewTracer("snapshot_unmarshal_error").Wrap(err)
	}

	return &snapshot, nil
}
