package consumer

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
	v1 "github.com/muhammadchandra19/exchange/proto/go/kafka/v1"
	"github.com/muhammadchandra19/exchange/services/market-data/internal/domain/ohlc"
	"github.com/muhammadchandra19/exchange/services/market-data/internal/domain/tick"
	"github.com/muhammadchandra19/exchange/services/market-data/pkg/config"
	"github.com/muhammadchandra19/exchange/services/market-data/pkg/interval"

	ohlcInfra "github.com/muhammadchandra19/exchange/services/market-data/internal/infrastructure/questdb/ohlc"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data/internal/infrastructure/questdb/tick"
	"github.com/segmentio/kafka-go"
)

// MatchConsumer is a consumer for the match topic.
type MatchConsumer struct {
	kafkaReader *kafka.Reader
	logger      logger.Interface

	tickUsecase tick.Usecase
	ohlcUsecase ohlc.Usecase
	dbTx        questdb.Transaction

	msgChan chan kafka.Message

	ohlcMutex        sync.Mutex
	ohlcBuffers      map[string]map[string]*ohlc.Buffer
	enabledIntervals []interval.Interval
}

// NewMatchConsumer creates a new MatchConsumer instance.
func NewMatchConsumer(
	config config.MatchKafkaConfig,
	logger logger.Interface,
	tickUsecase tick.Usecase,
	ohlcUsecase ohlc.Usecase,
	dbTx questdb.Transaction,
) *MatchConsumer {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Brokers,
		Topic:       config.Topic,
		GroupID:     config.ConsumerGroup,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	// Initialize enabled intervals (you can make this configurable)
	enabledIntervals := []interval.Interval{
		interval.Interval1m,
		interval.Interval5m,
		interval.Interval15m,
		interval.Interval1h,
		interval.Interval4h,
		interval.Interval1d,
	}

	return &MatchConsumer{
		kafkaReader:      kafkaReader,
		logger:           logger,
		tickUsecase:      tickUsecase,
		ohlcUsecase:      ohlcUsecase,
		dbTx:             dbTx,
		msgChan:          make(chan kafka.Message),
		ohlcBuffers:      make(map[string]map[string]*ohlc.Buffer),
		enabledIntervals: enabledIntervals,
	}
}

// Start starts the MatchConsumer
func (c *MatchConsumer) Start(ctx context.Context) {
	c.logger.InfoContext(ctx, "starting match consumer with OHLC aggregation",
		logger.Field{Key: "action", Value: "match_consumer_start"},
		logger.Field{Key: "enabled_intervals", Value: len(c.enabledIntervals)},
	)

	go c.startReading(ctx)
	go c.startProcessing(ctx)
	go c.startOHLCAggregation(ctx)
}

// startReading reads messages from Kafka
func (c *MatchConsumer) startReading(ctx context.Context) {
	defer close(c.msgChan)

	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "match consumer reader stopped")
			return
		default:
			msg, err := c.kafkaReader.ReadMessage(ctx)
			if err != nil {
				c.logger.ErrorContext(ctx, err, logger.Field{
					Key:   "action",
					Value: "read_match_message",
				})
				continue
			}

			select {
			case c.msgChan <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}

// startProcessing processes match messages and converts them to ticks
func (c *MatchConsumer) startProcessing(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "match consumer processor stopped")
			return
		case msg, ok := <-c.msgChan:
			if !ok {
				return
			}

			if err := c.processMatchMessage(ctx, msg); err != nil {
				c.logger.ErrorContext(ctx, err, logger.Field{
					Key:   "action",
					Value: "process_match_message",
				})
				continue
			}
		}
	}
}

// processMatchMessage processes a single match message
func (c *MatchConsumer) processMatchMessage(ctx context.Context, msg kafka.Message) error {
	var matchEvent v1.MatchEventPayload
	if err := json.Unmarshal(msg.Value, &matchEvent); err != nil {
		return err
	}

	c.logger.InfoContext(ctx, "processing match event",
		logger.Field{Key: "matchID", Value: matchEvent.MatchID},
		logger.Field{Key: "symbol", Value: matchEvent.Symbol},
		logger.Field{Key: "price", Value: matchEvent.Price},
		logger.Field{Key: "volume", Value: matchEvent.Volume},
		logger.Field{Key: "takerSide", Value: matchEvent.TakerSide},
	)

	// Convert match to tick
	tick := c.matchEventToTick(&matchEvent)

	// Store the tick
	if err := c.tickUsecase.StoreTick(ctx, tick); err != nil {
		return err
	}

	c.addTickToOHLCBuffers(tick)

	c.logger.InfoContext(ctx, "tick stored and added to OHLC buffers",
		logger.Field{Key: "symbol", Value: tick.Symbol},
		logger.Field{Key: "price", Value: tick.Price},
		logger.Field{Key: "volume", Value: tick.Volume},
	)

	return nil
}

// addTickToOHLCBuffers adds a tick to all relevant OHLC interval buffers
func (c *MatchConsumer) addTickToOHLCBuffers(tick *tickInfra.Tick) {
	c.ohlcMutex.Lock()
	defer c.ohlcMutex.Unlock()

	// Initialize symbol map if not exists
	if c.ohlcBuffers[tick.Symbol] == nil {
		c.ohlcBuffers[tick.Symbol] = make(map[string]*ohlc.Buffer)
	}

	// Add tick to all enabled intervals
	for _, intervalConfig := range c.enabledIntervals {
		bucketTime := intervalConfig.CalculateBucketTime(tick.Timestamp)

		// Get or create buffer for this symbol+interval
		buffer := c.ohlcBuffers[tick.Symbol][intervalConfig.Name.String()]
		if buffer == nil || !buffer.BucketTime.Equal(bucketTime) {
			// Create new buffer for new bucket
			if buffer != nil {
				// Flush the old buffer asynchronously
				go c.flushOHLCBuffer(context.Background(), buffer)
			}

			buffer = &ohlc.Buffer{
				Symbol:     tick.Symbol,
				Interval:   intervalConfig.Name,
				BucketTime: bucketTime,
				Ticks:      make([]interval.TickData, 0),
				LastUpdate: time.Now(),
			}
			c.ohlcBuffers[tick.Symbol][intervalConfig.Name.String()] = buffer
		}

		// Add tick to buffer
		buffer.Mutex.Lock()
		buffer.Ticks = append(buffer.Ticks, interval.TickData{
			Timestamp: tick.Timestamp,
			Price:     tick.Price,
			Volume:    tick.Volume,
		})
		buffer.LastUpdate = time.Now()
		buffer.Mutex.Unlock()
	}
}

// startOHLCAggregation runs periodic OHLC aggregation
func (c *MatchConsumer) startOHLCAggregation(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Aggregate every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "OHLC aggregation stopped")
			return
		case <-ticker.C:
			c.aggregateOHLCBuffers(ctx)
		}
	}
}

// aggregateOHLCBuffers processes all OHLC buffers and creates OHLC records
func (c *MatchConsumer) aggregateOHLCBuffers(ctx context.Context) {
	c.ohlcMutex.Lock()
	defer c.ohlcMutex.Unlock()

	var buffersToFlush []*ohlc.Buffer

	// Find buffers that need to be flushed
	for symbol, intervals := range c.ohlcBuffers {
		for intervalName, buffer := range intervals {
			if buffer != nil {
				buffer.Mutex.Lock()

				// Check if buffer should be flushed (older than bucket + some threshold)
				intervalConfig, _ := interval.GetInterval(intervalName)
				currentBucket := intervalConfig.CalculateBucketTime(time.Now())

				// Flush if we've moved to a new bucket or buffer is old enough
				if buffer.BucketTime.Before(currentBucket) ||
					time.Since(buffer.LastUpdate) > time.Minute {
					if len(buffer.Ticks) > 0 {
						buffersToFlush = append(buffersToFlush, buffer)
						// Remove from active buffers
						delete(c.ohlcBuffers[symbol], intervalName)
					}
				}

				buffer.Mutex.Unlock()
			}
		}
	}

	// Flush buffers outside of the lock
	for _, buffer := range buffersToFlush {
		go c.flushOHLCBuffer(ctx, buffer)
	}
}

// flushOHLCBuffer aggregates ticks in a buffer and stores the OHLC record
func (c *MatchConsumer) flushOHLCBuffer(ctx context.Context, buffer *ohlc.Buffer) {
	buffer.Mutex.Lock()
	defer buffer.Mutex.Unlock()

	if len(buffer.Ticks) == 0 {
		return
	}

	// Get interval configuration
	intervalConfig, err := interval.GetInterval(buffer.Interval.String())
	if err != nil {
		c.logger.ErrorContext(ctx, err, logger.Field{
			Key: "action", Value: "get_interval_config",
		})
		return
	}

	// Aggregate ticks into OHLC
	ohlcData := intervalConfig.AggregateOHLC(buffer.Ticks, buffer.BucketTime)

	// Convert to infrastructure OHLC entity
	ohlcRecord := &ohlcInfra.OHLC{
		Timestamp:  ohlcData.Timestamp,
		Symbol:     buffer.Symbol,
		Interval:   buffer.Interval,
		Open:       ohlcData.Open,
		High:       ohlcData.High,
		Low:        ohlcData.Low,
		Close:      ohlcData.Close,
		Volume:     ohlcData.Volume,
		TradeCount: ohlcData.TradeCount,
	}

	// Store OHLC record
	if err := c.ohlcUsecase.StoreOHLC(ctx, ohlcRecord); err != nil {
		c.logger.ErrorContext(ctx, err,
			logger.Field{Key: "action", Value: "store_ohlc"},
			logger.Field{Key: "symbol", Value: buffer.Symbol},
			logger.Field{Key: "interval", Value: buffer.Interval},
		)
		return
	}

	c.logger.InfoContext(ctx, "OHLC record created",
		logger.Field{Key: "symbol", Value: buffer.Symbol},
		logger.Field{Key: "interval", Value: buffer.Interval},
		logger.Field{Key: "bucketTime", Value: buffer.BucketTime.Format(time.RFC3339)},
		logger.Field{Key: "tickCount", Value: len(buffer.Ticks)},
		logger.Field{Key: "open", Value: ohlcData.Open},
		logger.Field{Key: "high", Value: ohlcData.High},
		logger.Field{Key: "low", Value: ohlcData.Low},
		logger.Field{Key: "close", Value: ohlcData.Close},
		logger.Field{Key: "volume", Value: ohlcData.Volume},
	)
}

// matchEventToTick converts a match event to a tick
func (c *MatchConsumer) matchEventToTick(matchEvent *v1.MatchEventPayload) *tickInfra.Tick {
	timestamp := time.Now()
	if matchEvent.Timestamp != nil {
		timestamp = matchEvent.Timestamp.AsTime()
	}

	return &tickInfra.Tick{
		Timestamp: timestamp,
		Symbol:    matchEvent.Symbol,
		Price:     matchEvent.Price,
		Volume:    int64(matchEvent.Volume),
		Side:      matchEvent.TakerSide,
	}
}

// Stop stops the consumer
func (c *MatchConsumer) Stop(ctx context.Context) error {
	c.logger.InfoContext(ctx, "stopping match consumer")

	// Flush any remaining OHLC buffers
	c.aggregateOHLCBuffers(ctx)

	return c.kafkaReader.Close()
}
