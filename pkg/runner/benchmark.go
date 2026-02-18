package runner

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"tpcds_benchmark/pkg/connection"
	"tpcds_benchmark/pkg/executor"
	"tpcds_benchmark/pkg/query"
	"tpcds_benchmark/pkg/storage"

	"tpcds_benchmark/pkg/config"
)

type BenchmarkRunner struct {
	cfg     *config.Config
	connMgr *connection.ConnectionManager
	storage *storage.CSVStorage
	queries []query.Query
	timeout time.Duration
	s3      *storage.S3Storage
}

func NewBenchmarkRunner(cfg *config.Config, connMgr *connection.ConnectionManager, s3 *storage.S3Storage, filename string) (*BenchmarkRunner, error) {

	st, err := storage.NewCSVStorage(cfg.ResultsPath, filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания хранилища: %w", err)
	}

	loader := query.NewQueryLoader(cfg.QueriesPath)
	queries, err := loader.LoadAll()
	if err != nil {
		st.Close()
		return nil, fmt.Errorf("ошибка загрузки запросов: %w", err)
	}

	log.Printf("загружено %d запросов", len(queries))

	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		st.Close()
		return nil, fmt.Errorf("неверный таймаут: %w", err)
	}

	return &BenchmarkRunner{
		cfg:     cfg,
		connMgr: connMgr,
		storage: st,
		queries: queries,
		timeout: timeout,
		s3:      s3,
	}, nil

}

func (br *BenchmarkRunner) Run() error {
	defer br.storage.Close()

	activeWarehouses := 0

	for _, wh := range br.cfg.Warehouses {
		if wh.Enabled {
			activeWarehouses++
		}
	}

	// log.Printf(
	// 	"запуск бенчмарка: %d хранилищ * %d запросов * %d повторений = %d тестов всего",
	// 	len(br.cfg.Warehouses),
	// 	len(br.queries),
	// 	totalTests,
	// )
	log.Printf("схема: %s", br.cfg.Schema)
	log.Printf("таймаут на запрос: %s", br.timeout)
	log.Printf("параллельность: %d потоков", br.cfg.Concurrency)
	log.Printf("активных хранилищ: %d", activeWarehouses)

	for _, wh := range br.cfg.Warehouses {
		if !wh.Enabled {
			log.Printf("пропуск неактивного хранилища: %s", wh.Name)
			continue
		}

		if err := br.runWarehouse(wh); err != nil {
			log.Printf("ERROR: %v хранилище %s", err, wh.Name)
		}

	}

	if err := br.storage.Close(); err != nil {
		return fmt.Errorf("ошибка при закрытии файла: %w", err)
	}
	log.Printf("тест завершен, результаты записаны в: %s", br.cfg.ResultsPath)

	filePath := br.storage.GetFilePath()

	if br.s3 != nil {
		if err := br.s3.Upload(filePath); err != nil {
			log.Printf("ошибка при загрузке файла в s3: %v", err)
			return nil
		}

		log.Printf("файл успешно загружен в s3")
	}
	return nil
}

func (br *BenchmarkRunner) runWarehouse(wh config.WarehouseConfig) error {
	schemaName := wh.GetSchemaName(br.cfg.Schema)

	log.Printf("=== хранилище %s (схема %s) ===", wh.Name, schemaName)

	executors := make([]executor.QueryExecutor, br.cfg.Concurrency)
	for i := 0; i < br.cfg.Concurrency; i++ {
		exec, err := executor.CreateExecutor(wh, br.connMgr, br.cfg.Schema)
		if err != nil {

			for j := 0; j < i; j++ {
				executors[j].Close()
			}
			return fmt.Errorf("ошибка создания экзекьютора: %w", err)
		}

		executors[i] = exec

	}

	defer func() {
		for _, exec := range executors {
			exec.Close()
		}
	}()

	tasksPerThread := len(br.queries) * br.cfg.Runs
	totalTasks := tasksPerThread * br.cfg.Concurrency

	log.Printf("запросов: %d, runs: %d, потоков: %d => всего задач: %d",
		len(br.queries),
		br.cfg.Runs,
		br.cfg.Concurrency,
		totalTasks,
	)

	var completedMu sync.Mutex
	completed := 0

	var wg sync.WaitGroup

	resultsChan := make(chan storage.BenchmarkResult, br.cfg.Concurrency*10)

	var writerWg sync.WaitGroup
	writerWg.Add(1)

	go func() {
		defer writerWg.Done()

		for result := range resultsChan {
			if err := br.storage.Save(result); err != nil {
				log.Printf(
					"[поток %d] WARNING: ошибка сохранения резульата (query=%s): %v",
					result.ThreadID,
					result.QueryID,
					err,
				)
			}
		}
	}()

	for threadID := 0; threadID < br.cfg.Concurrency; threadID++ {
		wg.Add(1)

		go func(threadID int, exec executor.QueryExecutor) {
			defer wg.Done()

			for _, q := range br.queries {

				for run := 1; run <= br.cfg.Runs; run++ {
					completedMu.Lock()
					completed++
					currentProgress := completed
					completedMu.Unlock()

					log.Printf(
						"[поток %d][%d/%d] запрос %s запуск %d/%d на %s",
						threadID,
						currentProgress,
						totalTasks,
						q.ID,
						run,
						br.cfg.Runs,
						wh.Name,
					)

					result := br.executeQuery(exec, q, schemaName, wh.Name, run, threadID)

					resultsChan <- result

					if result.Status == "success" {
						log.Printf("[поток %d] запрос завершен за %d ms (%d строк)",
							threadID,
							result.DurationMs,
							result.RowCount,
						)
					} else {
						log.Printf("[поток %d] * %s ошибка: %s",
							threadID,
							q.ID,
							result.ErrorMsg,
						)
					}
				}

			}
			log.Printf("[поток %d] завершил все свои задачи", threadID)
		}(threadID, executors[threadID])
	}

	wg.Wait()

	close(resultsChan)

	writerWg.Wait()

	log.Printf("=== завершены запросы в хранилище: %s ===", wh.Name)
	return nil

}

func (br *BenchmarkRunner) executeQuery(
	exec executor.QueryExecutor,
	q query.Query,
	schema string,
	warehouseName string,
	runNumber int,
	threadID int,
) storage.BenchmarkResult {
	ctx, cancel := context.WithTimeout(context.Background(), br.timeout)
	defer cancel()

	result := storage.BenchmarkResult{
		SaveResultTimestamp: time.Now(),
		QueryID:             q.ID,
		Warehouse:           warehouseName,
		Schema:              schema,
		RunNumber:           runNumber,
		ThreadID:            threadID,
	}

	queryResult, err := exec.Execute(ctx, q.SQL, schema)
	if err != nil {
		result.StartTimestamp = queryResult.StartTimestamp
		result.EndTimestamp = queryResult.EndTimestamp
		result.Status = "error"
		result.ErrorMsg = fmt.Sprintf("ошибка выполнения запроса: %v", err)
		return result
	}

	if !queryResult.Success {
		result.StartTimestamp = queryResult.StartTimestamp
		result.EndTimestamp = queryResult.EndTimestamp
		result.Status = "error"
		result.ErrorMsg = queryResult.Error
		result.DurationMs = int(queryResult.Duration.Milliseconds())
		return result
	}

	result.StartTimestamp = queryResult.StartTimestamp
	result.EndTimestamp = queryResult.EndTimestamp
	result.Status = "success"
	result.DurationMs = int(queryResult.Duration.Milliseconds())
	result.RowCount = queryResult.RowCount

	return result
}

func (br *BenchmarkRunner) Close() error {
	return br.storage.Close()
}
