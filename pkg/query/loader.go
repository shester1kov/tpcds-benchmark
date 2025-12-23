package query

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Query struct {
	ID   string
	SQL  string
	Path string
}

type QueryLoader struct {
	queriesPath string
}

func NewQueryLoader(queriesPath string) *QueryLoader {
	return &QueryLoader{
		queriesPath: queriesPath,
	}
}

func (ql *QueryLoader) LoadAll() ([]Query, error) {
	files, err := os.ReadDir(ql.queriesPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении директории: %w", err)
	}

	var queries []Query

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		query, err := ql.loadQuery(file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to load query %s: %s", file.Name(), err)
		}

		queries = append(queries, query)
	}

	sort.Slice(queries, func(i, j int) bool {
		return queries[i].ID < queries[j].ID
	})

	return queries, nil

}

func (ql *QueryLoader) loadQuery(filename string) (Query, error) {
	path := filepath.Join(ql.queriesPath, filename)

	content, err := os.ReadFile(path)

	if err != nil {
		return Query{}, err
	}

	id := strings.TrimSuffix(filename, ".sql")

	return Query{
		ID:   id,
		SQL:  string(content),
		Path: path,
	}, nil
}
