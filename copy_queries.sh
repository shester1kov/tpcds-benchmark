#!/bin/bash

# Проверяем количество аргументов
if [ $# -ne 3 ]; then
    echo "Использование: $0 <исходная_папка> <целевая_папка> <список_запросов>"
    echo "Пример: $0 ./source ./target \"query1,query19,query26\""
    exit 1
fi

source_dir="$1"
target_dir="$2"
queries_string="$3"

# Проверяем существование исходной папки
if [ ! -d "$source_dir" ]; then
    echo "Ошибка: Папка '$source_dir' не существует"
    exit 1
fi

# Создаем целевую папку (если не существует)
mkdir -p "$target_dir"

# Разбиваем строку запросов по запятым в массив
IFS=',' read -ra queries <<< "$queries_string"

# Счетчики
total=${#queries[@]}
success=0
failed=0

echo "Копирование $total файлов..."
echo "============================="

# Проходим по всем запросам
for query in "${queries[@]}"; do
    # Убираем возможные пробелы
    query_clean=$(echo "$query" | xargs)
    
    # Исходный файл (предполагаем, что он существует с нужным расширением)
    source_file="$source_dir/$query_clean.sql"
    
    # Целевой файл
    target_file="$target_dir/$query_clean.sql"
    
    # Проверяем существование исходного файла
    if [ -f "$source_file" ]; then
        # Копируем файл
        cp "$source_file" "$target_file"
        echo "✓ $query_clean.sql -> $target_dir/"
        ((success++))
    else
        echo "✗ $query_clean.sql - не найден"
        ((failed++))
    fi
done

echo "============================="
echo "Готово!"
echo "Успешно: $success"
echo "Не найдено: $failed"
echo "Всего: $total"