package postgres_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
)

// PostgresRepository implements the PaymentRepository port using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL payment repository
func NewPostgresRepository(connectionString string) (*PostgresRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create connection pool
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Check connection
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize repository
	repo := &PostgresRepository{pool: pool}

	return repo, nil
}

// Close closes the database connection pool
func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// Add records a payment in the specified channel
func (r *PostgresRepository) Add(channel domain.ProcessorChannel, amount float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx,
		"INSERT INTO payments (channel, amount) VALUES ($1, $2)",
		channel.String(), amount)

	if err != nil {
		return fmt.Errorf("failed to insert payment record: %w", err)
	}

	return nil
}

// GetSummary returns a summary of all payment channels for all time
func (r *PostgresRepository) GetSummary() (domain.PaymentsSummary, error) {
	return r.GetSummaryInRange(nil, nil)
}

// GetSummaryInRange returns a summary filtered by the provided time range
func (r *PostgresRepository) GetSummaryInRange(from, to *time.Time) (domain.PaymentsSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string
	var args []interface{}
	var argPosition int = 1

	// Base query
	query = `
		SELECT 
			channel, 
			COUNT(*) as total_requests, 
			SUM(amount) as total_amount 
		FROM payments
		WHERE 1=1
	`

	// Add time range filters if provided
	if from != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPosition)
		args = append(args, from.UTC())
		argPosition++
	}

	if to != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argPosition)
		args = append(args, to.UTC())
		argPosition++
	}

	// Group by channel
	query += " GROUP BY channel"

	// Execute the query
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return domain.PaymentsSummary{}, fmt.Errorf("failed to query payments summary: %w", err)
	}
	defer rows.Close()

	// Initialize summary with zero values
	summary := domain.PaymentsSummary{
		Default:  domain.PaymentsChannelStats{},
		Fallback: domain.PaymentsChannelStats{},
	}

	// Process results
	for rows.Next() {
		var channel string
		var totalRequests int
		var totalAmount float64

		if err := rows.Scan(&channel, &totalRequests, &totalAmount); err != nil {
			return domain.PaymentsSummary{}, fmt.Errorf("failed to scan row: %w", err)
		}

		// Update appropriate channel stats
		switch channel {
		case domain.DefaultProcessor.String():
			summary.Default.TotalRequests = totalRequests
			summary.Default.TotalAmount = totalAmount
		case domain.FallbackProcessor.String():
			summary.Fallback.TotalRequests = totalRequests
			summary.Fallback.TotalAmount = totalAmount
		}
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return domain.PaymentsSummary{}, fmt.Errorf("error iterating over rows: %w", err)
	}

	return summary, nil
}
