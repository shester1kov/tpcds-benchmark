package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/beltran/gohive"
)

type HiveExecutor struct {
	conn          *gohive.Connection
	name          string
	warehouseType string // trino, impala, vertica
}

func NewHiveExecutor(conn *gohive.Connection, name, warehouseType string) *HiveExecutor {
	return &HiveExecutor{
		conn:          conn,
		name:          name,
		warehouseType: warehouseType,
	}
}

func (e *HiveExecutor) Name() string {
	return e.name
}

func (e *HiveExecutor) Close() error {
	return e.conn.Close()
}

func (e *HiveExecutor) Execute(ctx context.Context, query string, schema string) (*QueryResult, error) {
	cursor := e.conn.Cursor()

	cursor.Exec(ctx, fmt.Sprintf("USE %s", schema))
	if cursor.Err != nil {
		return &QueryResult{
			Duration: 0,
			Success:  false,
			Error:    fmt.Sprintf("ошибка при выборе схемы: %v", cursor.Err),
		}, nil
	}

	start := time.Now()
	cursor.Exec(ctx, query)
	end := time.Now()
	duration := end.Sub(start)

	if cursor.Err != nil {
		return &QueryResult{
			StartTimestamp: start,
			EndTimestamp:   end,
			Duration:       duration,
			Success:        false,
			Error:          cursor.Err.Error(),
		}, nil
	}

	rowCount := 0

	return &QueryResult{
		StartTimestamp: start,
		EndTimestamp:   end,
		Duration:       duration,
		Success:        true,
		RowCount:       rowCount,
	}, nil

}
