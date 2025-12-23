package executor

import (
	"context"
	"database/sql"
	"time"
)

type SQLExecutor struct {
	conn          *sql.Conn
	name          string
	warehouseType string // trino, impala, vertica
	catalog       string // trino
}

func NewSQLExecutor(conn *sql.Conn, name, warehouseType, catalog string) *SQLExecutor {
	return &SQLExecutor{
		conn:          conn,
		name:          name,
		warehouseType: warehouseType,
		catalog:       catalog,
	}
}

func (e *SQLExecutor) Name() string {
	return e.name
}

func (e *SQLExecutor) Close() error {
	return e.conn.Close()
}

func (e *SQLExecutor) Execute(ctx context.Context, query string, schema string) (*QueryResult, error) {

	start := time.Now()
	rows, err := e.conn.QueryContext(ctx, query)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		return &QueryResult{
			StartTimestamp: start,
			EndTimestamp:   end,
			Duration:       duration,
			Success:        false,
			Error:          err.Error(),
		}, nil
	}

	defer rows.Close()

	rowCount := 0

	for rows.Next() {
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return &QueryResult{
			StartTimestamp: start,
			EndTimestamp:   end,
			Duration:       duration,
			Success:        false,
			Error:          err.Error(),
		}, nil
	}

	return &QueryResult{
		StartTimestamp: start,
		EndTimestamp:   end,
		Duration:       duration,
		Success:        true,
		RowCount:       rowCount,
	}, nil

}
