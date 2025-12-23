package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

type BenchmarkResult struct {
	SaveResultTimestamp time.Time
	StartTimestamp      time.Time
	EndTimestamp        time.Time
	QueryID             string
	Warehouse           string
	Schema              string
	RunNumber           int
	ThreadID            int
	DurationMs          int
	Status              string
	ErrorMsg            string
	RowCount            int
}

type CSVStorage struct {
	filepath string
	mu       sync.Mutex
	writer   *csv.Writer
	file     os.File
}

func NewCSVStorage(filepath string) (*CSVStorage, error) {
	fileExists := false

	if _, err := os.Stat(filepath); err == nil {
		fileExists = true
	}

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла: %w", err)
	}

	writer := csv.NewWriter(file)

	storage := &CSVStorage{
		filepath: filepath,
		writer:   writer,
		file:     *file,
	}

	if !fileExists {
		if err := storage.writeHeader(); err != nil {
			file.Close()
			return nil, err
		}
	}

	return storage, nil
}

func (s *CSVStorage) writeHeader() error {
	header := []string{
		"save_result_timestamp",
		"start_timestamp",
		"end_timestamp",
		"query_id",
		"warehouse",
		"schema",
		"run_number",
		"thread_id",
		"duration_ms",
		"status",
		"error_message",
		"row_count",
	}

	if err := s.writer.Write(header); err != nil {
		return err
	}

	s.writer.Flush()
	return s.writer.Error()
}

func (s *CSVStorage) Save(result BenchmarkResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := []string{
		result.SaveResultTimestamp.Format(time.RFC3339),
		result.StartTimestamp.Format(time.RFC3339),
		result.EndTimestamp.Format(time.RFC3339),
		result.QueryID,
		result.Warehouse,
		result.Schema,
		fmt.Sprintf("%d", result.RunNumber),
		fmt.Sprintf("%d", result.ThreadID),
		fmt.Sprintf("%d", result.DurationMs),
		result.Status,
		result.ErrorMsg,
		fmt.Sprintf("%d", result.RowCount),
	}

	if err := s.writer.Write(record); err != nil {
		return err
	}

	s.writer.Flush()
	return s.writer.Error()
}

func (s *CSVStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.writer.Flush()
	return s.file.Close()
}
