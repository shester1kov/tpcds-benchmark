package connection

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"tpcds_benchmark/pkg/config"

	"github.com/trinodb/trino-go-client/trino"
)

func (cm *ConnectionManager) ConnectTrino(cfg config.ConnectionConfig, schema string) (*sql.Conn, error) {
	var conn *sql.Conn

	err := cm.retry(fmt.Sprintf("Trino(%s.%s)", cfg.Database, schema), func() error {
		tlsConfig, err := loadTLSConfig(cm.certPath)
		if err != nil {
			return err
		}

		customClientName := "my-tls-client"
		trino.RegisterCustomClient(customClientName, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		})

		serverURI := fmt.Sprintf(
			"https://%s:%s@%s:%s",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
		)

		trinoConfig := trino.Config{
			ServerURI:         serverURI,
			Catalog:           cfg.Database,
			Schema:            schema,
			CustomClientName:  customClientName,
			SessionProperties: cfg.Properties,
		}

		dsn, err := trinoConfig.FormatDSN()
		// fmt.Println(dsn)
		if err != nil {
			return fmt.Errorf("ошибка формата строки подключения trino: %w", err)
		}

		db, err := sql.Open("trino", dsn)
		if err != nil {
			return fmt.Errorf("ошибка открытия соединения trino: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			db.Close()
			return fmt.Errorf("ошибка проверки соединения trino: %w", err)
		}

		ctx, cancel = context.WithTimeout(context.Background(), cm.connectionTimeout)
		defer cancel()
		c, err := db.Conn(ctx)

		if err != nil {
			db.Close()
			return fmt.Errorf("ошибка получения соединения trino: %w", err)
		}

		conn = c

		return nil

	})

	return conn, err

}
