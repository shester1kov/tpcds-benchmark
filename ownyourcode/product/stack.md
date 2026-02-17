# Technology Stack

## Detected Stack (existing project)

| Layer          | Technology     | Version   | Source      | Purpose                              |
|----------------|----------------|-----------|-------------|--------------------------------------|
| Language       | Go             | 1.25.0    | go.mod      | Основной язык приложения             |
| S3 Client      | minio-go       | —         | Verify at github.com/minio/minio-go | Загрузка файлов в Ozone S3          |
| Config         | gopkg.in/yaml.v3 | v3     | go.mod      | Парсинг YAML-конфигурации            |
| DB: Trino      | trinodb/trino-go-client | v0.333.0 | go.mod | Подключение к Trino        |
| DB: Impala     | sclgo/impala-go | v1.3.0   | go.mod      | Подключение к Impala                 |
| DB: Vertica    | vertica/vertica-sql-go | v1.3.4 | go.mod | Подключение к Vertica          |
| DB: Hive/Spark | beltran/gohive  | v1.8.1   | go.mod      | Подключение к Hive/Kyuubi            |
| Container      | Docker (Alpine) | —         | Dockerfile  | Контейнеризация                      |
| Analysis       | Python + pandas/matplotlib | 3.10 | packages/ | Анализ результатов          |

**Source Legend:**
- `go.mod` — версия из установленных зависимостей (источник истины)
- `—` — версия не верифицирована; проверь на официальном сайте

## New Dependency: minio-go

**Выбор:** `minio-go` вместо `aws-sdk-go`

**Почему:** minio-go явно позиционируется как клиент для любого S3-совместимого API,
что делает его предпочтительным выбором для Ozone S3 (не AWS). Меньше лишних абстракций.

Для установки:
```
go get github.com/minio/minio-go/v7
```

## Package Manager

**Using:** Go modules (`go mod`)

Go использует встроенную систему управления зависимостями. `go.sum` — файл с хешами,
`vendor/` — локальная копия зависимостей (используется в этом проекте).

## Key Files

| File                                  | Purpose                                      |
|---------------------------------------|----------------------------------------------|
| `cmd/main.go`                         | Точка входа                                  |
| `pkg/config/config.go`                | Структуры конфигурации + загрузка YAML       |
| `config/config.yaml`                  | Конфигурация (warehouses, paths, timeouts)   |
| `pkg/storage/csv_storage.go`          | Сохранение результатов в CSV                 |
| `pkg/storage/s3_storage.go`           | **NEW** Загрузка результатов в Ozone S3      |
| `pkg/runner/benchmark.go`             | Оркестрация запуска бенчмарка                |

## Version Freshness

⚠️ **Generated**: 2026-02-17

Версии зависимостей меняются. Если документу более 30 дней — перепроверь.
