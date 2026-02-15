package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a new connection pool from a URL string.
func Connect(url string) (Pool, error) {
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return Pool{}, fmt.Errorf("pg connect: %w", err)
	}
	return Pool{pool: pool}, nil
}

// New starts a configuration builder with the given connection URL.
func New(url string) Config {
	return Config{url: url}
}

// MaxConns sets the maximum number of connections in the pool.
func MaxConns(cfg Config, n int32) Config {
	cfg.maxConns = n
	return cfg
}

// MinConns sets the minimum number of idle connections in the pool.
func MinConns(cfg Config, n int32) Config {
	cfg.minConns = n
	return cfg
}

// MaxConnLifetime sets the maximum lifetime of a connection (nanoseconds).
func MaxConnLifetime(cfg Config, d int64) Config {
	cfg.maxConnLifetimeNs = d
	return cfg
}

// MaxConnIdleTime sets the maximum idle time for a connection (nanoseconds).
func MaxConnIdleTime(cfg Config, d int64) Config {
	cfg.maxConnIdleTimeNs = d
	return cfg
}

// Open creates a connection pool from the builder configuration.
func Open(cfg Config) (Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.url)
	if err != nil {
		return Pool{}, fmt.Errorf("pg config: %w", err)
	}
	if cfg.maxConns > 0 {
		poolCfg.MaxConns = cfg.maxConns
	}
	if cfg.minConns > 0 {
		poolCfg.MinConns = cfg.minConns
	}
	if cfg.maxConnLifetimeNs > 0 {
		poolCfg.MaxConnLifetime = time.Duration(cfg.maxConnLifetimeNs)
	}
	if cfg.maxConnIdleTimeNs > 0 {
		poolCfg.MaxConnIdleTime = time.Duration(cfg.maxConnIdleTimeNs)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return Pool{}, fmt.Errorf("pg open: %w", err)
	}
	return Pool{pool: pool}, nil
}

// Query executes a query that returns multiple rows.
func Query(p Pool, sql string, args ...any) (Rows, error) {
	rows, err := p.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return Rows{}, fmt.Errorf("pg query: %w", err)
	}
	return Rows{rows: rows}, nil
}

// QueryRow executes a query that returns at most one row.
func QueryRow(p Pool, sql string, args ...any) (Row, error) {
	row := p.pool.QueryRow(context.Background(), sql, args...)
	return Row{scanFn: row}, nil
}

// Exec executes a query that doesn't return rows (INSERT, UPDATE, DELETE).
func Exec(p Pool, sql string, args ...any) (Result, error) {
	tag, err := p.pool.Exec(context.Background(), sql, args...)
	if err != nil {
		return Result{}, fmt.Errorf("pg exec: %w", err)
	}
	return Result{tag: tag}, nil
}

// Scan scans values from a Row into destination pointers.
func Scan(r Row, dest ...any) error {
	row := r.scanFn.(pgx.Row)
	if err := row.Scan(dest...); err != nil {
		return fmt.Errorf("pg scan: %w", err)
	}
	return nil
}

// ScanString scans a single string value from a Row.
func ScanString(r Row) (string, error) {
	var v string
	row := r.scanFn.(pgx.Row)
	if err := row.Scan(&v); err != nil {
		return "", fmt.Errorf("pg scan string: %w", err)
	}
	return v, nil
}

// ScanInt scans a single int value from a Row.
func ScanInt(r Row) (int, error) {
	var v int
	row := r.scanFn.(pgx.Row)
	if err := row.Scan(&v); err != nil {
		return 0, fmt.Errorf("pg scan int: %w", err)
	}
	return v, nil
}

// ScanInt64 scans a single int64 value from a Row.
func ScanInt64(r Row) (int64, error) {
	var v int64
	row := r.scanFn.(pgx.Row)
	if err := row.Scan(&v); err != nil {
		return 0, fmt.Errorf("pg scan int64: %w", err)
	}
	return v, nil
}

// ScanBool scans a single bool value from a Row.
func ScanBool(r Row) (bool, error) {
	var v bool
	row := r.scanFn.(pgx.Row)
	if err := row.Scan(&v); err != nil {
		return false, fmt.Errorf("pg scan bool: %w", err)
	}
	return v, nil
}

// ScanFloat64 scans a single float64 value from a Row.
func ScanFloat64(r Row) (float64, error) {
	var v float64
	row := r.scanFn.(pgx.Row)
	if err := row.Scan(&v); err != nil {
		return 0, fmt.Errorf("pg scan float64: %w", err)
	}
	return v, nil
}

// Next advances the Rows cursor to the next row.
func Next(r Rows) bool {
	rows := r.rows.(pgx.Rows)
	return rows.Next()
}

// ScanRow scans values from the current row in Rows into destination pointers.
func ScanRow(r Rows, dest ...any) error {
	rows := r.rows.(pgx.Rows)
	if err := rows.Scan(dest...); err != nil {
		return fmt.Errorf("pg scan row: %w", err)
	}
	return nil
}

// Close closes the Rows cursor, releasing resources.
func Close(r Rows) {
	rows := r.rows.(pgx.Rows)
	rows.Close()
}

// CollectRows reads all remaining rows into a list of maps.
func CollectRows(r Rows) ([]map[string]any, error) {
	rows := r.rows.(pgx.Rows)
	defer rows.Close()

	descs := rows.FieldDescriptions()
	var results []map[string]any

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("pg collect rows: %w", err)
		}
		row := make(map[string]any, len(descs))
		for i, desc := range descs {
			row[desc.Name] = values[i]
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pg collect rows: %w", err)
	}
	return results, nil
}

// Begin starts a new transaction.
func Begin(p Pool) (Tx, error) {
	tx, err := p.pool.Begin(context.Background())
	if err != nil {
		return Tx{}, fmt.Errorf("pg begin: %w", err)
	}
	return Tx{tx: tx}, nil
}

// TxQuery executes a query within a transaction that returns multiple rows.
func TxQuery(t Tx, sql string, args ...any) (Rows, error) {
	tx := t.tx.(pgx.Tx)
	rows, err := tx.Query(context.Background(), sql, args...)
	if err != nil {
		return Rows{}, fmt.Errorf("pg tx query: %w", err)
	}
	return Rows{rows: rows}, nil
}

// TxQueryRow executes a query within a transaction that returns at most one row.
func TxQueryRow(t Tx, sql string, args ...any) (Row, error) {
	tx := t.tx.(pgx.Tx)
	row := tx.QueryRow(context.Background(), sql, args...)
	return Row{scanFn: row}, nil
}

// TxExec executes a query within a transaction that doesn't return rows.
func TxExec(t Tx, sql string, args ...any) (Result, error) {
	tx := t.tx.(pgx.Tx)
	tag, err := tx.Exec(context.Background(), sql, args...)
	if err != nil {
		return Result{}, fmt.Errorf("pg tx exec: %w", err)
	}
	return Result{tag: tag}, nil
}

// Commit commits the transaction.
func Commit(t Tx) error {
	tx := t.tx.(pgx.Tx)
	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("pg commit: %w", err)
	}
	return nil
}

// Rollback aborts the transaction.
func Rollback(t Tx) error {
	tx := t.tx.(pgx.Tx)
	if err := tx.Rollback(context.Background()); err != nil {
		return fmt.Errorf("pg rollback: %w", err)
	}
	return nil
}

// RowsAffected returns the number of rows affected by an INSERT, UPDATE, or DELETE.
func RowsAffected(r Result) int64 {
	return r.tag.RowsAffected()
}

// ClosePool closes all connections in the pool.
func ClosePool(p Pool) {
	p.pool.Close()
}
