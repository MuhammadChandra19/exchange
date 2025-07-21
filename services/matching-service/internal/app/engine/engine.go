package engine

import (
	"context"
	"sync"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	orderreaderv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/order-reader/v1"
	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/orderbook/v1"
	snapshotv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/snapshot/v1"
	"github.com/muhammadchandra19/exchange/services/matching-service/pkg/config"
	"go.uber.org/zap/zapcore"
)

// Engine is the simplified main engine for processing orders and managing the order book.
type Engine struct {
	// Core components
	orderbook     orderbookv1.Orderbook
	orderReader   orderreaderv1.OrderReader
	snapshotStore snapshotv1.Store
	logger        *logger.Logger
	config        *config.Config

	// Simple state management with mutex instead of atomics
	mu                 sync.RWMutex
	orderOffset        int64
	lastSnapshotOffset int64

	// Simple shutdown coordination
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Configuration
	snapshotInterval    time.Duration
	snapshotOffsetDelta int64

	// Match statistics
	totalMatches int64
	matchesMutex sync.RWMutex
}

// NewEngine creates a new instance of Engine with the provided dependencies.
func NewEngine(
	orderbook orderbookv1.Orderbook,
	orderReader orderreaderv1.OrderReader,
	snapshotStore snapshotv1.Store,
	logger *logger.Logger,
	config *config.Config,
) *Engine {
	return NewEngineWithOptions(orderbook, orderReader, snapshotStore, logger, config, DefaultEngineOptions())
}

// NewEngineWithOptions creates a new engine with custom options
func NewEngineWithOptions(
	orderbook orderbookv1.Orderbook,
	orderReader orderreaderv1.OrderReader,
	snapshotStore snapshotv1.Store,
	logger *logger.Logger,
	config *config.Config,
	options *Options,
) *Engine {
	e := &Engine{
		orderbook:     orderbook,
		orderReader:   orderReader,
		snapshotStore: snapshotStore,
		logger:        logger,
		config:        config,

		snapshotInterval:    options.SnapshotInterval,
		snapshotOffsetDelta: options.SnapshotOffsetDelta,
		orderOffset:         -1,
	}

	// Load snapshot during initialization
	if err := e.loadSnapshot(context.Background()); err != nil {
		e.logger.GetZap().Fatal("Failed to load snapshot", zapcore.Field{
			Key:       "error",
			Interface: err,
		})
	}

	return e
}

// Start initializes the engine and starts processing routines.
func (e *Engine) Start(ctx context.Context) error {
	// Create cancellable context
	e.ctx, e.cancel = context.WithCancel(ctx)

	e.wg.Add(2) // Reduced from 3 to 2 (no match consumer needed)
	go e.runOrderProcessor()
	go e.runSnapshotManager()

	e.logger.Info("Simplified engine started", logger.Field{
		Key:   "pair",
		Value: e.config.Pair,
	})

	return nil
}

// Stop gracefully shuts down the engine
func (e *Engine) Stop(ctx context.Context) error {
	if e.cancel != nil {
		e.cancel()
	}

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		e.logger.Info("Simplified engine stopped gracefully")
		return nil
	case <-ctx.Done():
		e.logger.Warn("Engine stop timeout exceeded")
		return ctx.Err()
	}
}

// runOrderProcessor combines order reading and processing in a single goroutine
func (e *Engine) runOrderProcessor() {
	defer e.wg.Done()

	e.logger.Info("Starting order processor", logger.Field{
		Key:   "pair",
		Value: e.config.Pair,
	})

	// Set the initial offset
	currentOffset := e.getOrderOffset()
	if currentOffset > 0 {
		currentOffset++
	}

	if err := e.orderReader.SetOffset(currentOffset); err != nil {
		e.logger.GetZap().Fatal("Failed to set offset for order reader", zapcore.Field{
			Key:       "error",
			Interface: err,
		})
	}

	for {
		select {
		case <-e.ctx.Done():
			e.logger.Info("Order processor shutting down")
			e.orderReader.Close()
			return
		default:
			// Read message directly
			msg, orderRequest, err := e.orderReader.ReadMessage(e.ctx)
			if err != nil {
				e.logger.ErrorContext(e.ctx, err, logger.Field{
					Key:   "action",
					Value: "read_order_message",
				})
				// Simple backoff
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Commit message
			if err := e.orderReader.CommitMessages(e.ctx, msg); err != nil {
				e.logger.ErrorContext(e.ctx, err, logger.Field{
					Key:   "action",
					Value: "commit_order_message",
				})
			}

			// Process order immediately
			if err := e.processOrder(&orderRequest); err != nil {
				e.logger.ErrorContext(e.ctx, err, logger.Field{
					Key:   "action",
					Value: "process_order",
				})
				continue
			}

			// Update offset
			e.setOrderOffset(msg.Offset)
		}
	}
}

// runSnapshotManager handles periodic snapshots
func (e *Engine) runSnapshotManager() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.snapshotInterval)
	defer ticker.Stop()

	e.logger.Info("Starting snapshot manager")

	for {
		select {
		case <-e.ctx.Done():
			e.logger.Info("Snapshot manager shutting down")
			return
		case <-ticker.C:
			if e.shouldCreateSnapshot() {
				e.createAndStoreSnapshot()
			}
		}
	}
}

// processOrder processes a single order request
func (e *Engine) processOrder(orderRequest *orderbookv1.PlaceOrderRequest) error {
	e.logger.Debug("Processing order",
		logger.Field{Key: "orderOffset", Value: orderRequest.Offset},
		logger.Field{Key: "userID", Value: orderRequest.UserID},
		logger.Field{Key: "bid", Value: orderRequest.Bid},
	)

	order := orderbookv1.NewOrder(orderRequest.UserID, orderRequest.Size, orderRequest.Bid)

	switch orderRequest.Type {
	case orderbookv1.OrderTypeLimit:
		return e.orderbook.PlaceLimitOrder(orderRequest.Price, order)
	case orderbookv1.OrderTypeMarket:
		matches, err := e.orderbook.PlaceMarketOrder(order)
		if err != nil {
			return err
		}

		// SIMPLIFIED: Just log matches directly instead of using channel
		if len(matches) > 0 {
			e.logMatches(matches)
		}
	case orderbookv1.OrderTypeCancel:
		err := e.orderbook.CancelOrder(orderRequest.OrderID)
		if err != nil {
			return err
		}
	}
	return nil
}

// logMatches logs the matches and updates statistics
func (e *Engine) logMatches(matches []orderbookv1.Match) {
	e.matchesMutex.Lock()
	e.totalMatches += int64(len(matches))
	currentTotal := e.totalMatches
	e.matchesMutex.Unlock()

	e.logger.Info("Matches executed",
		logger.Field{Key: "matchCount", Value: len(matches)},
		logger.Field{Key: "totalMatches", Value: currentTotal},
	)

	// Log each individual match
	for i, match := range matches {
		e.logger.Info("Trade executed",
			logger.Field{Key: "matchIndex", Value: i + 1},
			logger.Field{Key: "price", Value: match.Price},
			logger.Field{Key: "size", Value: match.SizeFilled},
			logger.Field{Key: "bidUser", Value: match.Bid.UserID},
			logger.Field{Key: "askUser", Value: match.Ask.UserID},
			logger.Field{Key: "bidOrderID", Value: match.Bid.ID},
			logger.Field{Key: "askOrderID", Value: match.Ask.ID},
			logger.Field{Key: "askIsFilled", Value: match.AskIsFilled()},
			logger.Field{Key: "bidIsFilled", Value: match.BidIsFilled()},
		)
	}
}

// shouldCreateSnapshot checks if a snapshot should be created
func (e *Engine) shouldCreateSnapshot() bool {
	e.mu.RLock()
	currentOffset := e.orderOffset
	lastSnapshotOffset := e.lastSnapshotOffset
	e.mu.RUnlock()

	if currentOffset <= 0 {
		return false
	}

	delta := currentOffset - lastSnapshotOffset
	return delta >= e.snapshotOffsetDelta
}

// createAndStoreSnapshot creates and stores a snapshot
func (e *Engine) createAndStoreSnapshot() {
	currentOffset := e.getOrderOffset()

	e.logger.Info("Creating snapshot", logger.Field{
		Key:   "currentOffset",
		Value: currentOffset,
	})

	snapshot := e.orderbook.CreateSnapshot()
	snapshot.OrderOffset = currentOffset

	if err := e.snapshotStore.Store(e.ctx, snapshot); err != nil {
		e.logger.ErrorContext(e.ctx, err, logger.Field{
			Key:   "action",
			Value: "store_snapshot",
		})
	} else {
		e.setLastSnapshotOffset(currentOffset)
		e.logger.Info("Snapshot stored successfully", logger.Field{
			Key:   "pair",
			Value: e.config.Pair,
		}, logger.Field{
			Key:   "offset",
			Value: currentOffset,
		})
	}
}

// Thread-safe getters and setters
func (e *Engine) getOrderOffset() int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.orderOffset
}

func (e *Engine) setOrderOffset(offset int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.orderOffset = offset
}

func (e *Engine) getLastSnapshotOffset() int64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastSnapshotOffset
}

func (e *Engine) setLastSnapshotOffset(offset int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastSnapshotOffset = offset
}

// loadSnapshot loads and restores the orderbook from snapshot
func (e *Engine) loadSnapshot(ctx context.Context) error {
	snapshot, err := e.snapshotStore.LoadStore(ctx)
	if err != nil {
		return err
	}

	if snapshot != nil {
		e.orderbook.RestoreOrderbook(snapshot)
		e.mu.Lock()
		e.orderOffset = snapshot.OrderOffset
		e.lastSnapshotOffset = snapshot.OrderOffset
		e.mu.Unlock()

		e.logger.Info("Orderbook restored from snapshot", logger.Field{
			Key:   "orderOffset",
			Value: snapshot.OrderOffset,
		})
	}

	return nil
}

// GetOrderOffset returns the current order offset
func (e *Engine) GetOrderOffset() int64 {
	return e.getOrderOffset()
}

// GetLastSnapshotOffset returns the last snapshot offset
func (e *Engine) GetLastSnapshotOffset() int64 {
	return e.getLastSnapshotOffset()
}

// GetTotalMatches returns the total number of matches processed
func (e *Engine) GetTotalMatches() int64 {
	e.matchesMutex.RLock()
	defer e.matchesMutex.RUnlock()
	return e.totalMatches
}
