package postgresql

import (
	"context"
	"fmt"
	"time"
)

// HealthCheck represents database health information
type HealthCheck struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	ActiveConns  int32         `json:"active_connections"`
	IdleConns    int32         `json:"idle_connections"`
	MaxConns     int32         `json:"max_connections"`
	DatabaseName string        `json:"database_name"`
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Error        string        `json:"error,omitempty"`
	Version      string        `json:"version,omitempty"`
}

// CheckHealth performs a comprehensive health check on the PostgreSQL connection
func (c *Client) CheckHealth(ctx context.Context) *HealthCheck {
	start := time.Now()

	health := &HealthCheck{
		DatabaseName: c.DatabaseName(),
		Host:         c.Host(),
		Port:         c.Port(),
	}

	// Get connection pool stats
	stats := c.Stats()
	health.ActiveConns = stats.AcquiredConns()
	health.IdleConns = stats.IdleConns()
	health.MaxConns = stats.MaxConns()

	// Test basic connectivity
	if err := c.Ping(ctx); err != nil {
		health.Status = "unhealthy"
		health.Error = fmt.Sprintf("ping failed: %v", err)
		health.ResponseTime = time.Since(start)
		return health
	}

	// Test query execution
	var version string
	err := c.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		health.Status = "unhealthy"
		health.Error = fmt.Sprintf("version query failed: %v", err)
		health.ResponseTime = time.Since(start)
		return health
	}

	health.Version = version
	health.Status = "healthy"
	health.ResponseTime = time.Since(start)

	return health
}

// IsHealthy returns true if the database is healthy
func (c *Client) IsHealthy(ctx context.Context) bool {
	health := c.CheckHealth(ctx)
	return health.Status == "healthy"
}
