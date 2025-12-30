#!/bin/bash
# Helper script to view C4 architecture diagrams

set -e

echo "======================================"
echo "TPC-DS Benchmark - C4 Architecture Viewer"
echo "======================================"
echo ""

# Функция для проверки зависимостей
check_dependency() {
    if ! command -v "$1" &> /dev/null; then
        echo "❌ $1 не установлен"
        return 1
    else
        echo "✅ $1 установлен"
        return 0
    fi
}

# Меню выбора
echo "Выберите способ просмотра диаграмм:"
echo ""
echo "1) Structurizr Lite (Docker) - рекомендуется"
echo "2) Сгенерировать PNG из PlantUML файлов"
echo "3) Сгенерировать SVG из PlantUML файлов"
echo "4) Открыть PlantUML онлайн редактор"
echo "5) Выход"
echo ""
read -p "Ваш выбор (1-5): " choice

case $choice in
    1)
        echo ""
        echo "Запуск Structurizr Lite через Docker..."
        if check_dependency docker; then
            echo ""
            echo "📦 Запуск Structurizr Lite на http://localhost:8080"
            echo "⏸️  Нажмите Ctrl+C для остановки"
            echo ""
            docker run -it --rm -p 8080:8080 \
                -v "$(pwd):/usr/local/structurizr" \
                structurizr/lite
        else
            echo ""
            echo "Установите Docker: https://docs.docker.com/get-docker/"
        fi
        ;;

    2)
        echo ""
        echo "Генерация PNG изображений из PlantUML файлов..."
        if check_dependency plantuml; then
            mkdir -p output
            for file in *.puml; do
                if [ -f "$file" ]; then
                    echo "Генерация: $file -> output/${file%.puml}.png"
                    plantuml -o "$(pwd)/output" "$file"
                fi
            done
            echo ""
            echo "✅ PNG файлы созданы в директории output/"
            echo "Откройте их в любом просмотрщике изображений"
        else
            echo ""
            echo "Установите PlantUML:"
            echo "  macOS: brew install plantuml"
            echo "  Ubuntu/Debian: apt-get install plantuml"
            echo "  Или скачайте с: https://plantuml.com/download"
        fi
        ;;

    3)
        echo ""
        echo "Генерация SVG изображений из PlantUML файлов..."
        if check_dependency plantuml; then
            mkdir -p output
            for file in *.puml; do
                if [ -f "$file" ]; then
                    echo "Генерация: $file -> output/${file%.puml}.svg"
                    plantuml -tsvg -o "$(pwd)/output" "$file"
                fi
            done
            echo ""
            echo "✅ SVG файлы созданы в директории output/"
            echo "Откройте их в браузере или редакторе"
        else
            echo ""
            echo "Установите PlantUML:"
            echo "  macOS: brew install plantuml"
            echo "  Ubuntu/Debian: apt-get install plantuml"
            echo "  Или скачайте с: https://plantuml.com/download"
        fi
        ;;

    4)
        echo ""
        echo "Открытие PlantUML онлайн редактора..."
        echo ""
        echo "🌐 Откройте в браузере: https://www.plantuml.com/plantuml/uml/"
        echo ""
        echo "Затем скопируйте содержимое одного из файлов:"
        echo "  - c4-system-context.puml"
        echo "  - c4-containers.puml"
        echo "  - c4-components.puml"
        echo "  - sequence-benchmark-flow.puml"
        echo ""

        # Попытка открыть браузер
        if command -v xdg-open &> /dev/null; then
            xdg-open "https://www.plantuml.com/plantuml/uml/" 2>/dev/null || true
        elif command -v open &> /dev/null; then
            open "https://www.plantuml.com/plantuml/uml/" 2>/dev/null || true
        fi
        ;;

    5)
        echo "Выход"
        exit 0
        ;;

    *)
        echo "❌ Неверный выбор"
        exit 1
        ;;
esac
