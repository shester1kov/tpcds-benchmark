package executor

import (
	"context"
	"time"
)

type QueryExecutor interface {
	Execute(ctx context.Context, query string, schema string) (*QueryResult, error)
	Name() string
	Close() error
}

type QueryResult struct {
	StartTimestamp time.Time
	EndTimestamp   time.Time
	Duration       time.Duration
	RowCount       int
	Success        bool
	Error          string
}
