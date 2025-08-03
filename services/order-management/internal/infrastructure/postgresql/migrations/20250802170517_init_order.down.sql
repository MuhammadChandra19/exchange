-- Migration: init_order
-- Created at: Sat Aug  2 17:05:17 WIB 2025

-- Drop trigger
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_orders_user_timestamp;
DROP INDEX IF EXISTS idx_orders_symbol_side_price;
DROP INDEX IF EXISTS idx_orders_symbol_status;
DROP INDEX IF EXISTS idx_orders_user_symbol;
DROP INDEX IF EXISTS idx_orders_symbol_side;
DROP INDEX IF EXISTS idx_orders_timestamp;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_symbol;
DROP INDEX IF EXISTS idx_orders_user_id;

-- Drop orders table
DROP TABLE IF EXISTS orders;

-- Drop enums
DROP TYPE IF EXISTS order_type;
DROP TYPE IF EXISTS order_side;
DROP TYPE IF EXISTS order_status;
