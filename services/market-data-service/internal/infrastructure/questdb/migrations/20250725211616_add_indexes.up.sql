-- Migration: add_indexes
-- Created at: Sat Jul 25 21:16:16 WIB 2025

-- First convert user_id columns from STRING to SYMBOL for indexing capability
ALTER TABLE orders ALTER COLUMN user_id TYPE SYMBOL CAPACITY 128;
ALTER TABLE order_events ALTER COLUMN user_id TYPE SYMBOL CAPACITY 128;

-- Add indexes on symbol columns for all tables (trading pairs)
ALTER TABLE ticks ALTER COLUMN symbol ADD INDEX;
ALTER TABLE ohlc ALTER COLUMN symbol ADD INDEX;
ALTER TABLE orders ALTER COLUMN symbol ADD INDEX;
ALTER TABLE order_events ALTER COLUMN symbol ADD INDEX;

-- Add indexes on user_id columns (now that they are SYMBOL type)
ALTER TABLE orders ALTER COLUMN user_id ADD INDEX;
ALTER TABLE order_events ALTER COLUMN user_id ADD INDEX;
