package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tpcds_benchmark/pkg/config"
	"tpcds_benchmark/pkg/connection"
	"tpcds_benchmark/pkg/executor"
	"tpcds_benchmark/pkg/runner"
	"tpcds_benchmark/pkg/storage"
	"tpcds_benchmark/pkg/utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("ошибка при чтении конфига: %v", err)
	}

	log.Printf("загружена конфигурация с %d хранилищами\n", len(cfg.Warehouses))

	connectionTimeout, err := time.ParseDuration(cfg.ConnectionTimeout)
	if err != nil {
		log.Fatalf("неверный connection_timeout: %v", err)
	}

	retryDelay, err := time.ParseDuration(cfg.RetryDelay)
	if err != nil {
		log.Fatalf("неверный retry_delay: %v", err)
	}

	filename := utils.GetFileName(cfg)

	connMgr, err := connection.NewConnectionManager(
		cfg.CertPath,
		connectionTimeout,
		cfg.ConnectionRetries,
		retryDelay,
	)
	if err != nil {
		log.Fatalf("ошибка создания менеджера соединений: %v", err)
	}

	// if err := testConnections(cfg, connMgr); err != nil {
	// 	log.Fatalf("ошибка проверки соединений: %v", err)
	// }

	// log.Println("все соединения успешно проверены")

	if err := testExecutors(cfg, connMgr); err != nil {
		log.Fatalf("ошибка проверки экзекьютора: %v", err)
	}

	log.Println("все экзекьюторы проверены успешно")

	var s3 *storage.S3Storage

	if cfg.S3 != nil && cfg.S3.Enabled {

		s3, err = storage.NewS3Storage(cfg.S3, cfg.CertPath)
		if err != nil {
			log.Fatalf("ошибка при создании s3 клиента: %v", err)
		}
	}

	benchRunner, err := runner.NewBenchmarkRunner(cfg, connMgr, s3, filename)
	if err != nil {
		log.Fatalf("ошибка создания бенчмарка: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nполучен сигнал прерывания, корректное завершение работы...")
		benchRunner.Close()
		os.Exit(0)
	}()

	log.Println("старт tpcds бенчмарка...")
	if err := benchRunner.Run(); err != nil {
		log.Fatalf("ошибка бенчмарка: %v", err)
	}

	log.Println("бенчмарк завершен")

}

func testConnections(cfg *config.Config, connMgr *connection.ConnectionManager) error {
	for _, wh := range cfg.Warehouses {
		fmt.Println()

		if !wh.Enabled {
			log.Printf("пропуск хранилища: %s", wh.Name)
			continue
		}

		log.Printf("тест соединения с %s", wh.Name)

		switch wh.Type {
		case "trino":
			db, err := connMgr.ConnectTrino(wh.Connection, cfg.Schema)
			if err != nil {
				return fmt.Errorf("%s: %s", wh.Name, err)
			}
			db.Close()

		case "impala":
			db, err := connMgr.ConnectImpala(wh.Connection, cfg.Schema)
			if err != nil {
				return fmt.Errorf("%s: %s", wh.Name, err)
			}
			db.Close()

		case "vertica":

			db, err := connMgr.ConnectVertica(wh.Connection, cfg.Schema)
			if err != nil {
				return fmt.Errorf("%s: %w", wh.Name, err)
			}
			db.Close()

		case "hive":
			engineType := "HIVE_SQL"
			if hc, ok := wh.Connection.Properties["kyuubi.engine.type"]; ok {
				engineType = hc
			}

			conn, err := connMgr.ConnectHive(wh.Connection, engineType, cfg.Schema)
			if err != nil {
				return fmt.Errorf("%s: %w", wh.Name, err)
			}
			conn.Close()

		case "spark":
			engineType := ""
			if hc, ok := wh.Connection.Properties["kyuubi.engine.type"]; ok {
				engineType = hc
			}

			conn, err := connMgr.ConnectHive(wh.Connection, engineType, cfg.Schema)
			if err != nil {
				return fmt.Errorf("%s: %w", wh.Name, err)
			}

			conn.Close()

		default:
			return fmt.Errorf("неизвестный тип хранилища: %s", wh.Type)
		}

		log.Printf("успешная проверка соединения с %s", wh.Type)

		fmt.Println()

	}

	return nil

}

func testExecutors(cfg *config.Config, connMgr *connection.ConnectionManager) error {
	for _, wh := range cfg.Warehouses {

		fmt.Println()

		if !wh.Enabled {
			log.Printf("пропуск хранилища: %s", wh.Name)
			continue
		}

		schemaName := wh.GetSchemaName(cfg.Schema)

		log.Printf("тест %s с схемой %s...", wh.Name, schemaName)

		exec, err := executor.CreateExecutor(wh, connMgr, cfg.Schema)
		if err != nil {
			return fmt.Errorf("ошибка создания экзекьютора для %s: %w", wh.Name, err)
		}
		defer exec.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		result, err := exec.Execute(ctx, "SELECT 1", schemaName)
		if err != nil {
			return fmt.Errorf("ошибка при выполнении запроса в %s: %w", wh.Name, err)

		}

		if !result.Success {
			return fmt.Errorf("ошибка запроса в %s: %s", wh.Name, result.Error)
		}

		log.Printf("запрос выполнен за %v", result.Duration)

		fmt.Println()
	}

	return nil
}
