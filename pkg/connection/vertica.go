package connection

import (
	"context"
	"database/sql"
	"fmt"
	"tpcds_benchmark/pkg/config"

	_ "github.com/vertica/vertica-sql-go"
)

func (cm *ConnectionManager) ConnectVertica(cfg config.ConnectionConfig, schema string) (*sql.Conn, error) {
	var conn *sql.Conn

	err := cm.retry(fmt.Sprintf("Vertica(%s)", schema), func() error {
		connStr := fmt.Sprintf(
			"vertica://%s:%s@%s:%s/%s?connection_load_balance=1",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		)

		db, err := sql.Open("vertica", connStr)
		if err != nil {
			return fmt.Errorf("ошибка открытия соединения vertica: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			db.Close()
			return fmt.Errorf("ошибка проверки соединения vertica: %w", err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()

		c, err := db.Conn(ctx)
		if err != nil {
			db.Close()
			return fmt.Errorf("ошибка получения соединения vertica: %w", err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()

		_, err = c.ExecContext(ctx, fmt.Sprintf("SET SEARCH_PATH TO %s", schema))
		if err != nil {
			conn.Close()
			db.Close()
			return fmt.Errorf("vertica SET SEARCH_PATH failed: %w", err)
		}

		conn = c
		return nil

	})

	return conn, err
}
