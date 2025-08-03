-- Migration: init_order_event
-- Created at: Sun Aug  3 08:47:49 WIB 2025

-- Create order event type enum
CREATE TYPE order_event_type AS ENUM (
    'order_placed', 
    'order_cancelled', 
    'order_modified', 
    'order_filled',
    'order_partial_fill'
);

-- Create order_events table
CREATE TABLE order_events (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL REFERENCES orders(id),
    event_type order_event_type NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    side order_side NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    
    -- For modifications
    new_price DECIMAL(20, 8),
    new_quantity BIGINT,
    
    -- For fills
    filled_quantity BIGINT,
    remaining_quantity BIGINT,
    
    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for order events
CREATE INDEX idx_order_events_order_id ON order_events(order_id);
CREATE INDEX idx_order_events_user_id ON order_events(user_id);
CREATE INDEX idx_order_events_timestamp ON order_events(timestamp DESC);
CREATE INDEX idx_order_events_event_type ON order_events(event_type);
CREATE INDEX idx_order_events_symbol ON order_events(symbol);

-- Create composite indexes
CREATE INDEX idx_order_events_order_timestamp ON order_events(order_id, timestamp DESC);
CREATE INDEX idx_order_events_user_timestamp ON order_events(user_id, timestamp DESC);

-- Add table comments
COMMENT ON TABLE order_events IS 'Order event history for audit and tracking';
COMMENT ON COLUMN order_events.id IS 'Unique event identifier';
COMMENT ON COLUMN order_events.order_id IS 'Reference to the order';
COMMENT ON COLUMN order_events.event_type IS 'Type of order event';
COMMENT ON COLUMN order_events.new_price IS 'New price for modification events';
COMMENT ON COLUMN order_events.new_quantity IS 'New quantity for modification events';
COMMENT ON COLUMN order_events.filled_quantity IS 'Quantity filled in this event';
COMMENT ON COLUMN order_events.remaining_quantity IS 'Remaining quantity after fill';

