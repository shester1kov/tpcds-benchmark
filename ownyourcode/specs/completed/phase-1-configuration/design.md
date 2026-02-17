# Technical Design: S3 Config — Фаза 1

> Основано на твоих решениях + официальная документация minio-go.

## Overview

Добавляем опциональную структуру `S3Config` в систему конфигурации проекта.
Структура хранит все параметры, необходимые для инициализации minio клиента.
S3 — необязательный модуль: если секция не указана или `enabled: false`, приложение работает как раньше.

## Структура S3Config

```
S3Config
├── endpoint     string   // хост:порт Ozone S3
├── access_key   string   // ключ доступа
├── secret_key   string   // секретный ключ
├── use_ssl      bool     // HTTPS или HTTP
├── bucket       string   // имя бакета
└── enabled      bool     // включить загрузку в S3
```

## Размещение в Config

```
Config (существующая структура)
├── Warehouses   []WarehouseConfig
├── ResultsPath  string
├── ...
└── S3           *S3Config   // НОВОЕ — указатель, nil если секция не указана
```

**Почему указатель:** позволяет отличить "не задано" (nil) от "задано с пустыми значениями" (пустая структура).

## Data Flow (только конфигурация)

1. **Запуск:** `LoadConfig("config/config.yaml")` читает файл
2. **Парсинг:** `yaml.Unmarshal` заполняет `Config`, включая `S3 *S3Config` если секция есть
3. **Валидация:** `Validate()` проверяет S3 поля только если `cfg.S3 != nil && cfg.S3.Enabled`
4. **Готово:** `Config` доступна остальным модулям через возвращаемый указатель

## Валидация

Добавить в `Validate()`:

```
if S3 != nil AND S3.Enabled:
    проверить endpoint не пустой → ошибка
    проверить access_key не пустой → ошибка
    проверить secret_key не пустой → ошибка
    проверить bucket не пустой → ошибка
```

Ошибки должны быть читаемыми: `"s3: endpoint не задан"`

## YAML-секция (пример)

```yaml
s3:
  endpoint: "ozone-s3.example.com:9000"
  access_key: "YOUR_ACCESS_KEY"
  secret_key: "YOUR_SECRET_KEY"
  use_ssl: false
  bucket: "benchmark-results"
  enabled: false
```

## Security Considerations

- `access_key` и `secret_key` — секретные данные
- В `config.yaml` (реальный конфиг) допустимо, если файл не в git
- `config.yaml.example` должен содержать **заглушки**, не реальные ключи
- Убедись, что `config.yaml` добавлен в `.gitignore`

## Паттерн из существующего кода

`WarehouseConfig.Enabled bool yaml:"enabled"` — тот же паттерн флага включения уже используется в проекте. Следуй ему.
