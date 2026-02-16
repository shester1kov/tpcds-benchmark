package executor

import (
	"fmt"
	"tpcds_benchmark/pkg/config"
	"tpcds_benchmark/pkg/connection"
)

func CreateExecutor(
	wh config.WarehouseConfig,
	connMgr *connection.ConnectionManager,
	baseSchema string,
) (QueryExecutor, error) {
	schema := wh.GetSchemaName(baseSchema)

	switch wh.Type {
	case "trino":
		conn, err := connMgr.ConnectTrino(wh.Connection, schema)
		if err != nil {
			return nil, err
		}
		return NewSQLExecutor(conn, wh.Name, wh.Type, wh.Connection.Database), nil

	case "impala":
		db, err := connMgr.ConnectImpala(wh.Connection, schema)
		if err != nil {
			return nil, err
		}

		return NewSQLExecutor(db, wh.Name, wh.Type, wh.Connection.Database), nil

	case "vertica":
		db, err := connMgr.ConnectVertica(wh.Connection, schema)
		if err != nil {
			return nil, err
		}
		return NewSQLExecutor(db, wh.Name, wh.Type, wh.Connection.Database), nil

	case "hive", "spark":
		engineType := ""
		if wh.Type == "hive" {
			engineType = "HIVE_SQL"
		}
		if hc, ok := wh.Connection.Properties["kyuubi.engine.type"]; ok {
			engineType = hc
		}

		conn, err := connMgr.ConnectHive(wh.Connection, engineType, schema)
		if err != nil {
			return nil, err
		}
		return NewHiveExecutor(conn, wh.Name, wh.Type), nil

	default:
		return nil, fmt.Errorf("неизвестный тип хранилища: %s", wh.Type)
	}
}
