-- Migration: add_indexes
-- Created at: Sat Jul 25 21:16:16 WIB 2025

-- Remove indexes on user_id columns first
ALTER TABLE order_events ALTER COLUMN user_id DROP INDEX;
ALTER TABLE orders ALTER COLUMN user_id DROP INDEX;

-- Remove indexes on symbol columns
ALTER TABLE order_events ALTER COLUMN symbol DROP INDEX;
ALTER TABLE orders ALTER COLUMN symbol DROP INDEX;
ALTER TABLE ohlc ALTER COLUMN symbol DROP INDEX;
ALTER TABLE ticks ALTER COLUMN symbol DROP INDEX;

-- Revert user_id columns back to STRING type
ALTER TABLE order_events ALTER COLUMN user_id TYPE STRING;
ALTER TABLE orders ALTER COLUMN user_id TYPE STRING;
