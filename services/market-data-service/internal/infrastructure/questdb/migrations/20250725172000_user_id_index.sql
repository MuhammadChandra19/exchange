-- Migration: user_id_index
-- Created at: Fri Jul 25 17:20:00 WIB 2025

-- Add indexes on user_id columns (now that they are SYMBOL type)

-- Index on user_id for orders table
ALTER TABLE orders ALTER COLUMN user_id ADD INDEX;

-- Index on user_id for order_events table
ALTER TABLE order_events ALTER COLUMN user_id ADD INDEX; 