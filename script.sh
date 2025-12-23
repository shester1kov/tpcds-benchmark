#!/bin/bash

# Переходим в папку tpcds
cd tpcds/modified || { echo "Папка tpcds не найдена"; exit 1; }

# Счетчик обработанных файлов
count=0

# Проходим по всем .sql файлам
for file in *.sql; do
    # Проверяем что файл существует
    if [ ! -f "$file" ]; then
        continue
    fi
    
    # Удаляем точку с запятой в конце файла
    # sed -i создает временный файл и заменяет оригинал
    # $ - конец файла, ;$ - точка с запятой в конце
    sed -i '' 's/;[[:space:]]*$//' "$file"
    
    count=$((count + 1))
    echo "Обработан: $file"
done

echo "Всего обработано файлов: $count"
