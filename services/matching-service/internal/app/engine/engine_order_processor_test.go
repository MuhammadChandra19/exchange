package engine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/orderbook/v1"
)

// Test helper to capture what happens in runOrderProcessor
type orderProcessorTestHelper struct {
	messages []kafka.Message
	orders   []orderbookv1.PlaceOrderRequest
	errors   []error
	mu       sync.Mutex
}

func (h *orderProcessorTestHelper) addMessage(msg kafka.Message, order orderbookv1.PlaceOrderRequest, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = append(h.messages, msg)
	h.orders = append(h.orders, order)
	h.errors = append(h.errors, err)
}

func (h *orderProcessorTestHelper) getCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.messages)
}

func TestEngine_RunOrderProcessor_Basic(t *testing.T) {
	testCases := []struct {
		name             string
		initialOffset    int64
		setupMocks       func(*testFixture, *orderProcessorTestHelper, context.CancelFunc)
		expectedMessages int
		expectedOffset   int64
		expectedOrders   int
		expectedMatches  int64
		waitTime         time.Duration
	}{
		{
			name:          "process single limit order",
			initialOffset: -1,
			setupMocks: func(f *testFixture, helper *orderProcessorTestHelper, cancel context.CancelFunc) {
				// Snapshot loading
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)

				// SetOffset expectation
				f.mockOrderReader.EXPECT().
					SetOffset(int64(-1)).
					Return(nil).
					Times(1)

				// One successful message
				msg := kafka.Message{Offset: 1}
				order := createTestOrderRequest("user1", orderbookv1.OrderTypeLimit, false, 10.0, 50000.0, 1)

				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					DoAndReturn(func(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error) {
						helper.addMessage(msg, order, nil)
						return msg, order, nil
					}).
					Times(1)

				f.mockOrderReader.EXPECT().
					CommitMessages(gomock.Any(), msg).
					Return(nil).
					Times(1)

				// Second call will be cancelled
				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					DoAndReturn(func(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error) {
						<-ctx.Done()
						return kafka.Message{}, orderbookv1.PlaceOrderRequest{}, ctx.Err()
					}).
					Times(1)

				f.mockOrderReader.EXPECT().
					Close().
					Times(1)

				// Cancel after a short delay
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			},
			expectedMessages: 1,
			expectedOffset:   1,
			expectedOrders:   1,
			expectedMatches:  0,
			waitTime:         200 * time.Millisecond,
		},
		{
			name:          "process market order with match",
			initialOffset: -1,
			setupMocks: func(f *testFixture, helper *orderProcessorTestHelper, cancel context.CancelFunc) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					SetOffset(int64(-1)).
					Return(nil).
					Times(1)

				// First message - limit order
				msg1 := kafka.Message{Offset: 1}
				order1 := createTestOrderRequest("seller", orderbookv1.OrderTypeLimit, false, 10.0, 50000.0, 1)

				// Second message - market order
				msg2 := kafka.Message{Offset: 2}
				order2 := createTestOrderRequest("buyer", orderbookv1.OrderTypeMarket, true, 5.0, 0.0, 2)

				callCount := 0
				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					DoAndReturn(func(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error) {
						callCount++
						if callCount == 1 {
							helper.addMessage(msg1, order1, nil)
							return msg1, order1, nil
						} else if callCount == 2 {
							helper.addMessage(msg2, order2, nil)
							return msg2, order2, nil
						} else {
							<-ctx.Done()
							return kafka.Message{}, orderbookv1.PlaceOrderRequest{}, ctx.Err()
						}
					}).
					Times(3)

				f.mockOrderReader.EXPECT().
					CommitMessages(gomock.Any(), msg1).
					Return(nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					CommitMessages(gomock.Any(), msg2).
					Return(nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					Close().
					Times(1)

				go func() {
					time.Sleep(100 * time.Millisecond)
					cancel()
				}()
			},
			expectedMessages: 2,
			expectedOffset:   2,
			expectedOrders:   1, // One order remains after match
			expectedMatches:  1,
			waitTime:         250 * time.Millisecond,
		},
		{
			name:          "handle read error with backoff",
			initialOffset: -1,
			setupMocks: func(f *testFixture, helper *orderProcessorTestHelper, cancel context.CancelFunc) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					SetOffset(int64(-1)).
					Return(nil).
					Times(1)

				callCount := 0
				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					DoAndReturn(func(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error) {
						callCount++
						if callCount == 1 {
							// First call returns error
							helper.addMessage(kafka.Message{}, orderbookv1.PlaceOrderRequest{}, errors.New("kafka error"))
							return kafka.Message{}, orderbookv1.PlaceOrderRequest{}, errors.New("kafka error")
						} else {
							<-ctx.Done()
							return kafka.Message{}, orderbookv1.PlaceOrderRequest{}, ctx.Err()
						}
					}).
					Times(2)

				f.mockOrderReader.EXPECT().
					Close().
					Times(1)

				go func() {
					time.Sleep(150 * time.Millisecond) // Allow time for backoff
					cancel()
				}()
			},
			expectedMessages: 1,  // One error message
			expectedOffset:   -1, // No successful processing
			expectedOrders:   0,
			expectedMatches:  0,
			waitTime:         250 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupTestFixture(t)
			defer fixture.teardown()
			helper := &orderProcessorTestHelper{}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tc.setupMocks(fixture, helper, cancel)

			engine := createTestEngine(fixture)

			if tc.initialOffset > 0 {
				engine.setOrderOffset(tc.initialOffset)
			}

			// Start the engine
			err := engine.Start(ctx)
			require.NoError(t, err)

			// Wait for processing
			time.Sleep(tc.waitTime)

			// Stop the engine
			stopCtx, stopCancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer stopCancel()

			err = engine.Stop(stopCtx)
			assert.NoError(t, err)

			// Verify results
			assert.Equal(t, tc.expectedMessages, helper.getCount())
			assert.Equal(t, tc.expectedOffset, engine.GetOrderOffset())
			assert.Equal(t, tc.expectedOrders, len(fixture.orderbook.Orders))
			assert.Equal(t, tc.expectedMatches, engine.GetTotalMatches())
		})
	}
}

func TestEngine_RunOrderProcessor_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		setupMocks  func(*testFixture, context.CancelFunc)
		expectPanic bool
	}{
		{
			name: "commit error should not stop processing",
			setupMocks: func(f *testFixture, cancel context.CancelFunc) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					SetOffset(int64(-1)).
					Return(nil).
					Times(1)

				msg := kafka.Message{Offset: 1}
				order := createTestOrderRequest("user1", orderbookv1.OrderTypeLimit, false, 10.0, 50000.0, 1)

				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					Return(msg, order, nil).
					Times(1)

				// Commit fails
				f.mockOrderReader.EXPECT().
					CommitMessages(gomock.Any(), msg).
					Return(errors.New("commit failed")).
					Times(1)

				// Should continue reading
				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					DoAndReturn(func(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error) {
						<-ctx.Done()
						return kafka.Message{}, orderbookv1.PlaceOrderRequest{}, ctx.Err()
					}).
					Times(1)

				f.mockOrderReader.EXPECT().
					Close().
					Times(1)

				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			},
			expectPanic: false,
		},
		{
			name: "processing error should not stop engine",
			setupMocks: func(f *testFixture, cancel context.CancelFunc) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					SetOffset(int64(-1)).
					Return(nil).
					Times(1)

				// Invalid order (negative price)
				msg := kafka.Message{Offset: 1}
				order := createTestOrderRequest("user1", orderbookv1.OrderTypeLimit, false, 10.0, -1.0, 1)

				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					Return(msg, order, nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					CommitMessages(gomock.Any(), msg).
					Return(nil).
					Times(1)

				f.mockOrderReader.EXPECT().
					ReadMessage(gomock.Any()).
					DoAndReturn(func(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error) {
						<-ctx.Done()
						return kafka.Message{}, orderbookv1.PlaceOrderRequest{}, ctx.Err()
					}).
					Times(1)

				f.mockOrderReader.EXPECT().
					Close().
					Times(1)

				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			},
			expectPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupTestFixture(t)
			defer fixture.teardown()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tc.setupMocks(fixture, cancel)

			engine := createTestEngine(fixture)

			if tc.expectPanic {
				assert.Panics(t, func() {
					engine.Start(ctx)
					time.Sleep(100 * time.Millisecond)
				})
			} else {
				err := engine.Start(ctx)
				require.NoError(t, err)

				time.Sleep(100 * time.Millisecond)

				stopCtx, stopCancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer stopCancel()

				err = engine.Stop(stopCtx)
				assert.NoError(t, err)
			}
		})
	}
}

// Test that we can't easily unit test the SetOffset fatal error
// because it calls log.Fatal which exits the process
func TestEngine_RunOrderProcessor_SetOffsetError_Documentation(t *testing.T) {
	// This test documents that SetOffset errors cause Fatal logs
	// In a real application, you might want to:
	// 1. Return the error instead of calling Fatal
	// 2. Use dependency injection for the logger to make it testable
	// 3. Have a separate initialization phase that can be tested

	t.Log("SetOffset errors currently call log.Fatal which cannot be easily unit tested")
	t.Log("Consider refactoring to return errors instead of calling Fatal for better testability")
}

// Integration test with realistic message flow
func TestEngine_RunOrderProcessor_Integration(t *testing.T) {
	fixture := setupTestFixture(t)
	defer fixture.teardown()

	// Simple integration test with controlled message flow
	fixture.mockSnapshotStore.EXPECT().
		LoadStore(gomock.Any()).
		Return(nil, nil).
		Times(1)

	fixture.mockOrderReader.EXPECT().
		SetOffset(int64(-1)).
		Return(nil).
		Times(1)

	// Create a realistic sequence of messages
	messages := []struct {
		msg   kafka.Message
		order orderbookv1.PlaceOrderRequest
	}{
		{
			msg:   kafka.Message{Offset: 1},
			order: createTestOrderRequest("seller1", orderbookv1.OrderTypeLimit, false, 10.0, 50000.0, 1),
		},
		{
			msg:   kafka.Message{Offset: 2},
			order: createTestOrderRequest("buyer1", orderbookv1.OrderTypeLimit, true, 8.0, 49900.0, 2),
		},
		{
			msg:   kafka.Message{Offset: 3},
			order: createTestOrderRequest("buyer2", orderbookv1.OrderTypeMarket, true, 5.0, 0.0, 3),
		},
	}

	messageIndex := 0
	fixture.mockOrderReader.EXPECT().
		ReadMessage(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error) {
			if messageIndex < len(messages) {
				msg := messages[messageIndex]
				messageIndex++
				return msg.msg, msg.order, nil
			}
			// Block until cancelled after all messages
			<-ctx.Done()
			return kafka.Message{}, orderbookv1.PlaceOrderRequest{}, ctx.Err()
		}).
		Times(len(messages) + 1) // +1 for the final cancelled call

	// Expect commits for all messages
	for _, msg := range messages {
		fixture.mockOrderReader.EXPECT().
			CommitMessages(gomock.Any(), msg.msg).
			Return(nil).
			Times(1)
	}

	fixture.mockOrderReader.EXPECT().
		Close().
		Times(1)

	engine := createTestEngine(fixture)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := engine.Start(ctx)
	require.NoError(t, err)

	// Wait for all messages to be processed
	time.Sleep(100 * time.Millisecond)
	cancel()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer stopCancel()

	err = engine.Stop(stopCtx)
	assert.NoError(t, err)

	// Verify final state
	assert.Equal(t, int64(3), engine.GetOrderOffset())
	assert.Equal(t, 2, len(fixture.orderbook.Orders))   // 2 limit orders remain after market order match
	assert.Equal(t, int64(1), engine.GetTotalMatches()) // Market order generated 1 match
}
