package tick

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
)

// Repository represents the repository for tick data.
type Repository struct {
	client questdb.QuestDBClient // Using interface instead of concrete type
}

// NewRepository creates a new tick repository.
func NewRepository(client questdb.QuestDBClient) *Repository {
	return &Repository{
		client: client,
	}
}

// Store stores a tick data point.
func (r *Repository) Store(ctx context.Context, tick *Tick) error {
	query := `INSERT INTO ticks (timestamp, symbol, price, volume, side) 
			  VALUES ($1, $2, $3, $4, $5)`

	err := r.client.Exec(ctx, query,
		tick.Timestamp, tick.Symbol, tick.Price, tick.Volume, tick.Side)

	if err != nil {
		return fmt.Errorf("failed to store tick: %w", err)
	}

	return nil
}

// StoreBatch stores a batch of tick data points.
func (r *Repository) StoreBatch(ctx context.Context, ticks []*Tick) error {
	if len(ticks) == 0 {
		return nil
	}

	// Use CopyFrom for better performance with large batches
	copyCount, err := r.client.CopyFrom(
		ctx,
		pgx.Identifier{"ticks"},
		[]string{"timestamp", "symbol", "price", "volume", "exchange", "side"},
		pgx.CopyFromSlice(len(ticks), func(i int) ([]any, error) {
			tick := ticks[i]
			return []any{
				tick.Timestamp,
				tick.Symbol,
				tick.Price,
				tick.Volume,
				tick.Side,
			}, nil
		}),
	)

	if err != nil {
		return fmt.Errorf("failed to copy ticks: %w", err)
	}

	fmt.Printf("Inserted %d ticks\n", copyCount)
	return nil
}

// GetByFilter retrieves tick data points by filter.
func (r *Repository) GetByFilter(ctx context.Context, filter Filter) ([]*Tick, error) {
	query := "SELECT timestamp, symbol, price, volume, side FROM ticks WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.Symbol != "" {
		query += fmt.Sprintf(" AND symbol = $%d", argIndex)
		args = append(args, filter.Symbol)
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

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ticks: %w", err)
	}
	defer rows.Close()

	var ticks []*Tick
	for rows.Next() {
		tick := &Tick{}
		err := rows.Scan(&tick.Timestamp, &tick.Symbol, &tick.Price, &tick.Volume, &tick.Side)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tick: %w", err)
		}
		ticks = append(ticks, tick)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return ticks, nil
}

// GetLatestBySymbol retrieves the latest tick data point by symbol.
func (r *Repository) GetLatestBySymbol(ctx context.Context, symbol string) (*Tick, error) {
	query := `SELECT timestamp, symbol, price, volume, side 
			  FROM ticks 
			  WHERE symbol = $1 
			  ORDER BY timestamp DESC 
			  LIMIT 1`

	tick := &Tick{}
	err := r.client.QueryRow(ctx, query, symbol).Scan(
		&tick.Timestamp, &tick.Symbol, &tick.Price, &tick.Volume, &tick.Side)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest tick: %w", err)
	}

	fmt.Println("tick", tick)

	return tick, nil
}

// GetVolumeBySymbol retrieves the volume by symbol.
func (r *Repository) GetVolumeBySymbol(ctx context.Context, symbol string, from, to time.Time) (int64, error) {
	query := `SELECT COALESCE(SUM(volume), 0) FROM ticks 
			  WHERE symbol = $1 AND timestamp >= $2 AND timestamp <= $3`

	var totalVolume int64
	err := r.client.QueryRow(ctx, query, symbol, from, to).Scan(&totalVolume)
	if err != nil {
		return 0, fmt.Errorf("failed to get volume: %w", err)
	}

	return totalVolume, nil
}
