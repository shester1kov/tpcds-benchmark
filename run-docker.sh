#!/bin/bash

# Скрипт для запуска tpcds-benchmark в Docker контейнере
# Поддерживает гибкое монтирование конфигурации, запросов и результатов

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Значения по умолчанию
IMAGE_NAME="tpcds-benchmark:latest"
CONTAINER_NAME="tpcds-benchmark-$(date +%Y%m%d-%H%M%S)"
CONFIG_FILE="${CONFIG_FILE:-./config/config.yaml}"
QUERIES_DIR="${QUERIES_DIR:-./tpcds_simple_queries}"
RESULTS_DIR="${RESULTS_DIR:-./results}"
CERTS_FILE="${CERTS_FILE:-}"

# Функция помощи
usage() {
    cat << EOF
Использование: $0 [OPTIONS]

Опции:
    -c, --config FILE       Путь к config.yaml (по умолчанию: ./config/config.yaml)
    -q, --queries DIR       Путь к директории с запросами (по умолчанию: ./tpcds_simple_queries)
    -r, --results DIR       Путь к директории для результатов (по умолчанию: ./results)
    -t, --certs FILE        Путь к файлу сертификатов (опционально)
    -n, --name NAME         Имя контейнера (по умолчанию: tpcds-benchmark-TIMESTAMP)
    -b, --build             Пересобрать образ перед запуском
    -h, --help              Показать эту справку

Примеры:
    # Запуск с простыми запросами
    $0 -q ./tpcds_simple_queries -r ./results/simple

    # Запуск со сложными запросами и пересборкой образа
    $0 -b -q ./tpcds_hard_queries -r ./results/hard

    # Запуск с кастомным конфигом
    $0 -c ./config/production.yaml -q ./tpcds_good_queries -r ./results/production

    # Использование переменных окружения
    CONFIG_FILE=./my-config.yaml QUERIES_DIR=./my-queries $0

EOF
    exit 1
}

# Парсинг аргументов
BUILD=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -q|--queries)
            QUERIES_DIR="$2"
            shift 2
            ;;
        -r|--results)
            RESULTS_DIR="$2"
            shift 2
            ;;
        -t|--certs)
            CERTS_FILE="$2"
            shift 2
            ;;
        -n|--name)
            CONTAINER_NAME="$2"
            shift 2
            ;;
        -b|--build)
            BUILD=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Неизвестная опция: $1${NC}"
            usage
            ;;
    esac
done

# Проверка существования файлов и директорий
echo -e "${GREEN}Проверка параметров...${NC}"

if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}Ошибка: файл конфигурации не найден: $CONFIG_FILE${NC}"
    exit 1
fi

if [ ! -d "$QUERIES_DIR" ]; then
    echo -e "${RED}Ошибка: директория запросов не найдена: $QUERIES_DIR${NC}"
    exit 1
fi

# Создаем директорию для результатов если не существует
mkdir -p "$RESULTS_DIR"

# Получаем абсолютные пути
CONFIG_FILE=$(realpath "$CONFIG_FILE")
QUERIES_DIR=$(realpath "$QUERIES_DIR")
RESULTS_DIR=$(realpath "$RESULTS_DIR")

echo -e "${GREEN}Конфигурация:${NC}"
echo -e "  Config:  $CONFIG_FILE"
echo -e "  Queries: $QUERIES_DIR"
echo -e "  Results: $RESULTS_DIR"
if [ -n "$CERTS_FILE" ]; then
    CERTS_FILE=$(realpath "$CERTS_FILE")
    echo -e "  Certs:   $CERTS_FILE"
fi
echo ""

# Сборка образа если нужно
if [ "$BUILD" = true ]; then
    echo -e "${GREEN}Сборка Docker образа...${NC}"
    docker build -t "$IMAGE_NAME" .
    echo ""
fi

# Формируем команду docker run
DOCKER_CMD="docker run --rm"
DOCKER_CMD="$DOCKER_CMD --name $CONTAINER_NAME"
DOCKER_CMD="$DOCKER_CMD -v $CONFIG_FILE:/app/config/config.yaml:ro"
DOCKER_CMD="$DOCKER_CMD -v $QUERIES_DIR:/app/queries:ro"
DOCKER_CMD="$DOCKER_CMD -v $RESULTS_DIR:/app/results"

# Добавляем сертификаты если указаны
if [ -n "$CERTS_FILE" ]; then
    if [ -f "$CERTS_FILE" ]; then
        DOCKER_CMD="$DOCKER_CMD -v $CERTS_FILE:/app/certs/cacerts.pem:ro"
    else
        echo -e "${YELLOW}Предупреждение: файл сертификатов не найден: $CERTS_FILE${NC}"
    fi
fi

DOCKER_CMD="$DOCKER_CMD -e TZ=Europe/Moscow"
DOCKER_CMD="$DOCKER_CMD $IMAGE_NAME"

# Запуск контейнера
echo -e "${GREEN}Запуск контейнера: $CONTAINER_NAME${NC}"
echo -e "${YELLOW}Команда: $DOCKER_CMD${NC}"
echo ""

eval $DOCKER_CMD

echo ""
echo -e "${GREEN}Бенчмарк завершен!${NC}"
echo -e "Результаты сохранены в: ${GREEN}$RESULTS_DIR${NC}"
