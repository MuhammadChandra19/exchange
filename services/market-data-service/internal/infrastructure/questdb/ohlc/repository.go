package ohlc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
	"github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/shared"
)

// Repository represents the repository for OHLC data.
type Repository struct {
	client questdb.QuestDBClient // Using interface instead of concrete type
}

// NewRepository creates a new OHLC repository.
func NewRepository(client questdb.QuestDBClient) *Repository {
	return &Repository{
		client: client,
	}
}

// Store stores an OHLC data point.
func (r *Repository) Store(ctx context.Context, ohlc *OHLC) error {
	query := `INSERT INTO ohlc (timestamp, symbol, interval, open, high, low, close, volume, trade_count) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	err := r.client.Exec(ctx, query,
		ohlc.Timestamp, ohlc.Symbol, ohlc.Interval, ohlc.Open, ohlc.High,
		ohlc.Low, ohlc.Close, ohlc.Volume, ohlc.TradeCount)

	if err != nil {
		return fmt.Errorf("failed to store OHLC: %w", err)
	}

	return nil
}

// StoreBatch stores a batch of OHLC data points.
func (r *Repository) StoreBatch(ctx context.Context, ohlcs []*OHLC) error {
	if len(ohlcs) == 0 {
		return nil
	}

	// Use CopyFrom for better performance
	copyCount, err := r.client.CopyFrom(
		ctx,
		pgx.Identifier{"ohlc"},
		[]string{"timestamp", "symbol", "interval", "open", "high", "low", "close", "volume", "trade_count"},
		pgx.CopyFromSlice(len(ohlcs), func(i int) ([]any, error) {
			ohlc := ohlcs[i]
			return []any{
				ohlc.Timestamp,
				ohlc.Symbol,
				ohlc.Interval,
				ohlc.Open,
				ohlc.High,
				ohlc.Low,
				ohlc.Close,
				ohlc.Volume,
				ohlc.TradeCount,
			}, nil
		}),
	)

	if err != nil {
		return fmt.Errorf("failed to copy OHLC batch: %w", err)
	}

	fmt.Printf("Inserted %d OHLC records\n", copyCount)
	return nil
}

// GetByFilter retrieves OHLC data points by filter.
func (r *Repository) GetByFilter(ctx context.Context, filter OHLCFilter) ([]*OHLC, error) {
	query := "SELECT timestamp, symbol, interval, open, high, low, close, volume, trade_count FROM ohlc WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.Symbol != "" {
		query += fmt.Sprintf(" AND symbol = $%d", argIndex)
		args = append(args, filter.Symbol)
		argIndex++
	}

	if filter.Interval != shared.Interval_INTERVAL_UNDEFINED {
		query += fmt.Sprintf(" AND interval = $%d", argIndex)
		args = append(args, filter.Interval)
		argIndex++
	}

	if filter.From != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filter.From)
		argIndex++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filter.To)
		argIndex++
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
	}

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query OHLC: %w", err)
	}
	defer rows.Close()

	var ohlcs []*OHLC
	for rows.Next() {
		ohlc := &OHLC{}
		err := rows.Scan(&ohlc.Timestamp, &ohlc.Symbol, &ohlc.Interval, &ohlc.Open,
			&ohlc.High, &ohlc.Low, &ohlc.Close, &ohlc.Volume, &ohlc.TradeCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OHLC: %w", err)
		}
		ohlcs = append(ohlcs, ohlc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ohlcs, nil
}

// GetLatest retrieves the latest OHLC data point.
func (r *Repository) GetLatest(ctx context.Context, symbol, interval string) (*OHLC, error) {
	query := `SELECT timestamp, symbol, interval, open, high, low, close, volume, trade_count
			  FROM ohlc 
			  WHERE symbol = $1 AND interval = $2 
			  ORDER BY timestamp DESC 
			  LIMIT 1`

	ohlc := &OHLC{}
	err := r.client.QueryRow(ctx, query, symbol, interval).Scan(
		&ohlc.Timestamp, &ohlc.Symbol, &ohlc.Interval, &ohlc.Open, &ohlc.High,
		&ohlc.Low, &ohlc.Close, &ohlc.Volume, &ohlc.TradeCount)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest OHLC: %w", err)
	}

	return ohlc, nil
}

// GetIntradayData retrieves intraday OHLC data points.
func (r *Repository) GetIntradayData(ctx context.Context, symbol string, interval string, limit int) ([]*OHLC, error) {
	query := `SELECT timestamp, symbol, interval, open, high, low, close, volume, trade_count
			  FROM ohlc 
			  WHERE symbol = $1 AND interval = $2 
			  ORDER BY timestamp DESC 
			  LIMIT $3`

	rows, err := r.client.Query(ctx, query, symbol, interval, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query intraday OHLC: %w", err)
	}
	defer rows.Close()

	var ohlcs []*OHLC
	for rows.Next() {
		ohlc := &OHLC{}
		err := rows.Scan(&ohlc.Timestamp, &ohlc.Symbol, &ohlc.Interval, &ohlc.Open,
			&ohlc.High, &ohlc.Low, &ohlc.Close, &ohlc.Volume, &ohlc.TradeCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OHLC: %w", err)
		}
		ohlcs = append(ohlcs, ohlc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ohlcs, nil
}
