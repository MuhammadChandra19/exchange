-- Create ticks table
CREATE TABLE IF NOT EXISTS ticks (
    timestamp TIMESTAMP,
    symbol SYMBOL CAPACITY 1000 CACHE,
    price DOUBLE,
    volume LONG,
    exchange SYMBOL CAPACITY 50 CACHE,
    side SYMBOL CAPACITY 10 CACHE
) TIMESTAMP(timestamp) PARTITION BY HOUR WAL;

-- Create OHLC table (includes volume data)
CREATE TABLE IF NOT EXISTS ohlc (
    timestamp TIMESTAMP,
    symbol SYMBOL CAPACITY 1000 CACHE,
    interval SYMBOL CAPACITY 20 CACHE, -- '1m', '5m', '15m', '1h', '4h', '1d'
    open DOUBLE,
    high DOUBLE,
    low DOUBLE,
    close DOUBLE,
    volume LONG,
    trade_count LONG,
    exchange SYMBOL CAPACITY 50 CACHE
) TIMESTAMP(timestamp) PARTITION BY DAY WAL
DEDUP UPSERT KEYS (timestamp, symbol, interval, exchange);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    order_id STRING,
    timestamp TIMESTAMP,
    symbol SYMBOL CAPACITY 1000 CACHE,
    side SYMBOL CAPACITY 10 CACHE,
    price DOUBLE,
    quantity LONG,
    order_type SYMBOL CAPACITY 10 CACHE,
    status SYMBOL CAPACITY 10 CACHE,
    exchange SYMBOL CAPACITY 50 CACHE,
    user_id STRING
) TIMESTAMP(timestamp) PARTITION BY DAY WAL
DEDUP UPSERT KEYS (order_id);

-- Create order events table
CREATE TABLE IF NOT EXISTS order_events (
    event_id STRING,
    timestamp TIMESTAMP,
    order_id STRING,
    event_type SYMBOL CAPACITY 20 CACHE,
    symbol SYMBOL CAPACITY 1000 CACHE,
    side SYMBOL CAPACITY 10 CACHE,
    price DOUBLE,
    quantity LONG,
    exchange SYMBOL CAPACITY 50 CACHE,
    user_id STRING,
    new_price DOUBLE,
    new_quantity LONG
) TIMESTAMP(timestamp) PARTITION BY HOUR WAL; 