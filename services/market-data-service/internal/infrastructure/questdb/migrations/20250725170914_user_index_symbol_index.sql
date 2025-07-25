-- Migration: user_index_symbol_index
-- Created at: Fri Jul 25 17:09:14 WIB 2025

-- Add indexes for better query performance

-- Index on symbol for ticks table
ALTER TABLE ticks ALTER COLUMN symbol ADD INDEX;

-- Index on symbol for ohlc table  
ALTER TABLE ohlc ALTER COLUMN symbol ADD INDEX;

-- Index on symbol for orders table
ALTER TABLE orders ALTER COLUMN symbol ADD INDEX;

-- Index on symbol for order_events table
ALTER TABLE order_events ALTER COLUMN symbol ADD INDEX;

