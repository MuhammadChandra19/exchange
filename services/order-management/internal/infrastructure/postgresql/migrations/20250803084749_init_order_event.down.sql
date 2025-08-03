-- Migration: init_order_event
-- Created at: Sun Aug  3 08:47:49 WIB 2025

-- Drop indexes
DROP INDEX IF EXISTS idx_order_events_user_timestamp;
DROP INDEX IF EXISTS idx_order_events_order_timestamp;
DROP INDEX IF EXISTS idx_order_events_symbol;
DROP INDEX IF EXISTS idx_order_events_event_type;
DROP INDEX IF EXISTS idx_order_events_timestamp;
DROP INDEX IF EXISTS idx_order_events_user_id;
DROP INDEX IF EXISTS idx_order_events_order_id;

-- Drop table
DROP TABLE IF EXISTS order_events;

-- Drop enum
DROP TYPE IF EXISTS order_event_type;
