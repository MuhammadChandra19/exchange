-- Migration: alter_type
-- Created at: Fri Jul 25 17:18:38 WIB 2025

-- Convert user_id columns from STRING to SYMBOL for better performance and indexing

-- Convert user_id in orders table to SYMBOL
ALTER TABLE orders ALTER COLUMN user_id TYPE SYMBOL CAPACITY 128;

-- Convert user_id in order_events table to SYMBOL  
ALTER TABLE order_events ALTER COLUMN user_id TYPE SYMBOL CAPACITY 128;

