package connection

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"tpcds_benchmark/pkg/config"

	"github.com/sclgo/impala-go"
)

func (cm *ConnectionManager) ConnectImpala(cfg config.ConnectionConfig, database string) (*sql.Conn, error) {

	var conn *sql.Conn

	err := cm.retry(fmt.Sprintf("Impala(%s)", database), func() error {
		opts := impala.DefaultOptions
		opts.Host = cfg.Host
		opts.Port = cfg.Port
		opts.UseLDAP = true
		opts.Username = cfg.Username
		opts.Password = cfg.Password

		opts.UseTLS = true
		opts.CACertPath = cm.certPath

		// opts.ConnectTimeout = time.Duration(cm.connectionTimeout.Seconds())

		connector := impala.NewConnector(&opts)
		db := sql.OpenDB(connector)

		ctx, cancel := context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			db.Close()
			return fmt.Errorf("ошибка проверки соединения impala: %w", err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()

		c, err := db.Conn(ctx)
		if err != nil {
			db.Close()
			return fmt.Errorf("ошибка получения соединения impala: %w", err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()

		_, err = c.ExecContext(ctx, fmt.Sprintf("USE %s", database))
		if err != nil {
			conn.Close()
			db.Close()
			return fmt.Errorf("impala USE %s failed: %w", database, err)
		}

		for key, value := range cfg.Properties {
			ctx, cancel := context.WithTimeout(context.Background(), cm.connectionTimeout)
			_, err := c.ExecContext(ctx, fmt.Sprintf("SET %s=%s", key, value))
			cancel()

			if err != nil {
				log.Printf("WARNING: Failed to set %s=%s: %v", key, value, err)
			}
		}

		conn = c
		return nil

	})

	return conn, err
}
