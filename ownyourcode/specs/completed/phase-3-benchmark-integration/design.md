# Technical Design: Интеграция S3-загрузки в бенчмарк

> Твои архитектурные решения, оформленные в виде технического дизайна.

## Обзор

`S3Storage` создаётся в `main` как опциональная зависимость и передаётся в `BenchmarkRunner` через конструктор. Runner не знает про флаг `enabled` — он просто проверяет `nil`. После завершения всех тестов CSV закрывается явно, затем вызывается upload. `defer Close()` остаётся как страховка при панике/ошибках.

## Поток данных

```
main()
  ├── cfg.S3Config.Enabled == true?
  │     ├── да  → NewS3Storage(&cfg.S3Config) → s3
  │     └── нет → s3 = nil
  │
  └── NewBenchmarkRunner(cfg, connMgr, s3)
        │
        └── Run()
              ├── defer storage.Close()       ← страховка
              ├── ... все тесты ...
              ├── storage.Close()             ← явный вызов, файл готов
              ├── if s3 != nil:
              │     s3.Upload(filePath)
              │       ├── успех → log: "файл загружен: X байт"
              │       └── ошибка → log: "ошибка загрузки в S3: ..."
              └── return nil
```

## Изменения в компонентах

| Компонент | Изменение | Тип |
|-----------|-----------|-----|
| `BenchmarkRunner` | Добавить поле `s3 *storage.S3Storage` | Модификация |
| `NewBenchmarkRunner` | Принять `s3 *storage.S3Storage` как параметр | Модификация |
| `Run()` | Явный `Close()` + conditional upload | Модификация |
| `main.go` | Создание `S3Storage` при `Enabled == true` | Модификация |

## Сигнатура конструктора (после изменений)

```go
func NewBenchmarkRunner(
    cfg     *config.Config,
    connMgr *connection.ConnectionManager,
    s3      *storage.S3Storage,   // nil если отключено
) (*BenchmarkRunner, error)
```

## Обработка ошибок

| Место | Ошибка | Действие |
|-------|--------|----------|
| `NewS3Storage` в main | Неверный endpoint/credentials | `log.Fatalf` — не можем продолжить |
| `storage.Close()` явный | Ошибка записи | Логируем, продолжаем (upload может не запускаться) |
| `s3.Upload()` | S3 недоступен | `log.Printf("WARNING: ...")`, return nil |
| `defer storage.Close()` | Файл уже закрыт | Ошибка игнорируется |

## Логирование (примеры)

```
// Успех:
log.Printf("файл успешно загружен в S3: %s (%d байт)", filename, size)

// Ошибка (не фатальная):
log.Printf("WARNING: ошибка загрузки в S3, файл остался локально: %v", err)

// S3 отключён (опционально, для отладки):
log.Printf("загрузка в S3 отключена")
```
