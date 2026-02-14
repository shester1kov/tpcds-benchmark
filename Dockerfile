# Multi-stage build для минимального размера образа

# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Установка необходимых инструментов для сборки
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник с оптимизацией
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o tpcds-benchmark \
    ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# Установка CA сертификатов для TLS соединений
RUN apk --no-cache add ca-certificates tzdata

# Создаем непривилегированного пользователя
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /build/tpcds-benchmark /app/tpcds-benchmark

# Создаем директории для монтирования
RUN mkdir -p /app/config /app/queries /app/results /app/certs && \
    chown -R appuser:appuser /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Точки монтирования:
# /app/config - для config.yaml
# /app/queries - для SQL запросов
# /app/results - для результатов бенчмарка
# /app/certs - для TLS сертификатов (опционально)
VOLUME ["/app/config", "/app/queries", "/app/results", "/app/certs"]

# Запуск приложения
ENTRYPOINT ["/app/tpcds-benchmark"]
