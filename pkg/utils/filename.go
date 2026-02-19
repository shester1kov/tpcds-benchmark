package utils

import (
	"fmt"
	"time"
	"tpcds_benchmark/pkg/config"
)

func GetFileName(cfg *config.Config) string {
	now := time.Now()
	filename := fmt.Sprintf(
		"%d_runs_%d_concurrency_%s_schema_%s%s",
		cfg.Runs,
		cfg.Concurrency,
		cfg.Schema,
		now.Format("2006-01-02_15_04_05.000"),
		".csv",
	)

	return filename
}
