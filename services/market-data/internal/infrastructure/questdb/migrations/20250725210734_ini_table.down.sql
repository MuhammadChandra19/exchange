-- Migration: ini_table
-- Created at: Fri Jul 25 21:07:34 WIB 2025

-- Write your DOWN migration SQL here
DROP TABLE IF EXISTS order_events;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS ohlc;
DROP TABLE IF EXISTS ticks; 
