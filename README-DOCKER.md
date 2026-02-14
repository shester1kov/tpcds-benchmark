# TPC-DS Benchmark - Docker Руководство

## Обзор

Этот проект теперь поддерживает запуск в Docker контейнере, что позволяет:
- Запускать бенчмарк в изолированной среде
- Легко менять конфигурацию между запусками
- Использовать разные наборы запросов
- Запускать на Docker on YARN
- Повторять тесты с разными параметрами

## Быстрый старт

### 1. Сборка Docker образа

```bash
docker build -t tpcds-benchmark:latest .
```

### 2. Подготовка конфигурации

Создайте файл `config/config.yaml` на основе `config/config.yaml.example`:

```bash
cp config/config.yaml.example config/config.yaml
# Отредактируйте config.yaml под ваши нужды
```

### 3. Запуск

#### Вариант A: Использование скрипта (рекомендуется)

```bash
# Запуск с простыми запросами
./run-docker.sh -q ./tpcds_simple_queries -r ./results/simple

# Запуск со сложными запросами
./run-docker.sh -q ./tpcds_hard_queries -r ./results/hard

# Запуск с кастомным конфигом
./run-docker.sh -c ./config/production.yaml -q ./tpcds_good_queries -r ./results/prod
```

#### Вариант B: Прямой запуск через Docker

```bash
docker run --rm \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  -v $(pwd)/tpcds_simple_queries:/app/queries:ro \
  -v $(pwd)/results:/app/results \
  tpcds-benchmark:latest
```

#### Вариант C: Использование docker-compose

```bash
# Запуск с простыми запросами
docker-compose up benchmark-simple

# Запуск со средними запросами
docker-compose up benchmark-mid

# Запуск со сложными запросами
docker-compose up benchmark-hard

# Запуск с хорошими запросами
docker-compose up benchmark-good
```

## Монтирование директорий

### Обязательные монтирования

| Путь в контейнере | Назначение | Режим |
|-------------------|------------|-------|
| `/app/config/config.yaml` | Файл конфигурации | read-only |
| `/app/queries` | Директория с SQL запросами | read-only |
| `/app/results` | Директория для результатов | read-write |

### Опциональные монтирования

| Путь в контейнере | Назначение | Режим |
|-------------------|------------|-------|
| `/app/certs/cacerts.pem` | TLS сертификаты | read-only |

## Примеры использования

### Пример 1: Последовательное тестирование разных конфигураций

```bash
# Тест 1: Простые запросы, 1 поток
./run-docker.sh \
  -c ./config/config-1thread.yaml \
  -q ./tpcds_simple_queries \
  -r ./results/test1-simple-1thread

# Тест 2: Простые запросы, 4 потока
./run-docker.sh \
  -c ./config/config-4threads.yaml \
  -q ./tpcds_simple_queries \
  -r ./results/test2-simple-4threads

# Тест 3: Сложные запросы, 1 поток
./run-docker.sh \
  -c ./config/config-1thread.yaml \
  -q ./tpcds_hard_queries \
  -r ./results/test3-hard-1thread
```

### Пример 2: Использование переменных окружения

```bash
# Установка переменных
export CONFIG_FILE=./config/my-config.yaml
export QUERIES_DIR=./tpcds_good_queries
export RESULTS_DIR=./results/experiment1

# Запуск
./run-docker.sh
```

### Пример 3: Запуск с TLS сертификатами

```bash
./run-docker.sh \
  -c ./config/config.yaml \
  -q ./tpcds_simple_queries \
  -r ./results/secure-test \
  -t ./certs/cacerts.pem
```

### Пример 4: Пересборка образа и запуск

```bash
./run-docker.sh --build -q ./tpcds_hard_queries -r ./results/latest
```

## Запуск на Docker on YARN

### Подготовка образа

1. Соберите образ:
```bash
docker build -t your-registry.com/tpcds-benchmark:v1.0 .
```

2. Загрузите в registry:
```bash
docker push your-registry.com/tpcds-benchmark:v1.0
```

### Запуск через YARN

```bash
# Пример команды для Docker on YARN
yarn jar hadoop-yarn-applications-distributedshell.jar \
  -jar hadoop-yarn-applications-distributedshell.jar \
  -shell_command "docker run \
    -v /mnt/config/config.yaml:/app/config/config.yaml:ro \
    -v /mnt/queries:/app/queries:ro \
    -v /mnt/results:/app/results \
    your-registry.com/tpcds-benchmark:v1.0" \
  -container_memory 8192 \
  -container_vcores 4
```

## Структура файлов конфигурации

### Создание разных конфигураций

Создайте несколько файлов конфигурации для разных сценариев:

```
config/
├── config.yaml.example          # Шаблон
├── config-1thread.yaml          # Последовательное выполнение
├── config-4threads.yaml         # 4 параллельных потока
├── config-production.yaml       # Продакшн настройки
└── config-test.yaml             # Тестовые настройки
```

Пример `config-1thread.yaml`:
```yaml
queries_path: "./queries"
results_path: "./results/benchmark_results.csv"
runs: 3
concurrency: 1
timeout: "10m"
# ... остальные настройки
```

Пример `config-4threads.yaml`:
```yaml
queries_path: "./queries"
results_path: "./results/benchmark_results.csv"
runs: 3
concurrency: 4  # 4 параллельных потока
timeout: "10m"
# ... остальные настройки
```

## Организация результатов

### Рекомендуемая структура

```
results/
├── experiment1/
│   ├── simple-1thread/
│   │   └── benchmark_results.csv
│   ├── simple-4threads/
│   │   └── benchmark_results.csv
│   └── hard-1thread/
│       └── benchmark_results.csv
├── experiment2/
│   └── ...
└── production/
    └── ...
```

### Автоматическое именование с timestamp

Можно добавить timestamp в путь результатов:

```bash
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
./run-docker.sh \
  -q ./tpcds_simple_queries \
  -r ./results/run-$TIMESTAMP
```

## Параметры скрипта run-docker.sh

```
Опции:
    -c, --config FILE       Путь к config.yaml
    -q, --queries DIR       Путь к директории с запросами
    -r, --results DIR       Путь к директории для результатов
    -t, --certs FILE        Путь к файлу сертификатов (опционально)
    -n, --name NAME         Имя контейнера
    -b, --build             Пересобрать образ перед запуском
    -h, --help              Показать справку
```

## Отладка

### Просмотр логов

```bash
# Логи последнего запущенного контейнера
docker logs tpcds-benchmark-$(date +%Y%m%d)

# Логи конкретного контейнера
docker logs <container-name>
```

### Запуск с интерактивным shell

```bash
docker run -it --rm \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  -v $(pwd)/tpcds_simple_queries:/app/queries:ro \
  -v $(pwd)/results:/app/results \
  --entrypoint /bin/sh \
  tpcds-benchmark:latest
```

### Проверка смонтированных файлов

```bash
docker run --rm \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  -v $(pwd)/tpcds_simple_queries:/app/queries:ro \
  -v $(pwd)/results:/app/results \
  --entrypoint /bin/sh \
  tpcds-benchmark:latest \
  -c "ls -la /app/config && ls -la /app/queries && ls -la /app/results"
```

## Оптимизация

### Ограничение ресурсов

```bash
docker run --rm \
  --memory="8g" \
  --cpus="4.0" \
  -v $(pwd)/config/config.yaml:/app/config/config.yaml:ro \
  -v $(pwd)/tpcds_simple_queries:/app/queries:ro \
  -v $(pwd)/results:/app/results \
  tpcds-benchmark:latest
```

### Размер образа

Текущий образ использует multi-stage build для минимального размера:
- Builder stage: ~1.5GB (временный)
- Final stage: ~20MB (alpine + бинарник)

## Troubleshooting

### Проблема: Permission denied при записи результатов

**Решение**: Убедитесь что директория results имеет правильные права:
```bash
chmod -R 777 ./results
```

### Проблема: Config file not found

**Решение**: Проверьте абсолютные пути и убедитесь что файл существует:
```bash
ls -la $(pwd)/config/config.yaml
```

### Проблема: TLS certificate error

**Решение**: Убедитесь что путь к сертификатам правильный и файл смонтирован:
```bash
./run-docker.sh -t ./path/to/cacerts.pem ...
```

## Полезные команды

```bash
# Посмотреть размер образа
docker images tpcds-benchmark:latest

# Удалить старые контейнеры
docker container prune

# Удалить образ
docker rmi tpcds-benchmark:latest

# Посмотреть запущенные контейнеры
docker ps

# Остановить контейнер
docker stop <container-name>
```

## Интеграция в CI/CD

### Пример GitHub Actions

```yaml
name: Run TPC-DS Benchmark

on:
  schedule:
    - cron: '0 2 * * *'  # Каждый день в 2:00

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: docker build -t tpcds-benchmark:latest .

      - name: Run benchmark
        run: |
          ./run-docker.sh \
            -c ./config/config.yaml \
            -q ./tpcds_good_queries \
            -r ./results/$(date +%Y%m%d)

      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: results/
```

## Заключение

Теперь вы можете:
- ✅ Запускать бенчмарк в изолированном контейнере
- ✅ Легко менять конфигурацию между запусками
- ✅ Использовать разные наборы запросов
- ✅ Сохранять результаты в отдельные директории
- ✅ Повторять тесты с разными параметрами
- ✅ Запускать на Docker on YARN

Для вопросов и предложений создавайте Issues в репозитории.
