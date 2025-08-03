-- Migration: init_order
-- Created at: Sat Aug  2 17:05:17 WIB 2025

-- Create order status enum
CREATE TYPE order_status AS ENUM ('placed', 'cancelled', 'modified');

-- Create order side enum  
CREATE TYPE order_side AS ENUM ('buy', 'sell');

-- Create order type enum
CREATE TYPE order_type AS ENUM ('market', 'limit', 'stop', 'stop_limit');

-- Create orders table
CREATE TABLE orders (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    side order_side NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    quantity BIGINT NOT NULL,
    type order_type NOT NULL,
    status order_status NOT NULL DEFAULT 'placed',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_timestamp ON orders(timestamp DESC);
CREATE INDEX idx_orders_symbol_side ON orders(symbol, side);
CREATE INDEX idx_orders_user_symbol ON orders(user_id, symbol);
CREATE INDEX idx_orders_symbol_status ON orders(symbol, status);

-- Create composite index for order book queries
CREATE INDEX idx_orders_symbol_side_price ON orders(symbol, side, price) WHERE status = 'placed';

-- Create index for user order history
CREATE INDEX idx_orders_user_timestamp ON orders(user_id, timestamp DESC);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_orders_updated_at 
    BEFORE UPDATE ON orders 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add table comment
COMMENT ON TABLE orders IS 'Trading orders table for order management service';
COMMENT ON COLUMN orders.id IS 'Unique order identifier (ULID)';
COMMENT ON COLUMN orders.user_id IS 'User who placed the order';
COMMENT ON COLUMN orders.symbol IS 'Trading pair symbol (e.g., BTC/USDT)';
COMMENT ON COLUMN orders.side IS 'Order side: buy or sell';
COMMENT ON COLUMN orders.price IS 'Order price with 8 decimal precision';
COMMENT ON COLUMN orders.quantity IS 'Order quantity in base units';
COMMENT ON COLUMN orders.type IS 'Order type: market, limit, stop, stop_limit';
COMMENT ON COLUMN orders.status IS 'Current order status';
COMMENT ON COLUMN orders.timestamp IS 'When the order was placed (from event)';
COMMENT ON COLUMN orders.created_at IS 'When the record was created in database';
COMMENT ON COLUMN orders.updated_at IS 'When the record was last updated';
