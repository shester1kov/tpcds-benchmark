workspace "TPC-DS Benchmark System" "Architecture of the TPC-DS Benchmark Tool for multiple data warehouses" {

    model {
        # Внешние системы и пользователи
        user = person "Data Engineer / DBA" "Специалист, который запускает бенчмарки для оценки производительности хранилищ данных"

        # Внешние системы - хранилища данных
        trinoWarehouse = softwareSystem "Trino Data Warehouse" "Распределенное SQL хранилище данных Trino" "External System"
        hiveWarehouse = softwareSystem "Hive Data Warehouse" "Apache Hive хранилище данных с поддержкой HiveSQL и Spark" "External System"
        impalaWarehouse = softwareSystem "Impala Data Warehouse" "Cloudera Impala MPP SQL движок" "External System"
        verticaWarehouse = softwareSystem "Vertica Data Warehouse" "Vertica аналитическая СУБД" "External System"

        # Основная система
        tpcdsBenchmark = softwareSystem "TPC-DS Benchmark System" "Система для выполнения TPC-DS бенчмарков на различных хранилищах данных" {

            # Контейнеры
            cliApp = container "CLI Application" "Консольное приложение для запуска бенчмарков" "Go" {

                # Компоненты
                mainController = component "Main Controller" "Точка входа, инициализация системы и обработка сигналов" "Go package: main"

                configLoader = component "Config Loader" "Загрузка и валидация конфигурации из YAML файла" "Go package: config"

                connectionManager = component "Connection Manager" "Управление подключениями к БД с поддержкой retry и TLS" "Go package: connection" {
                    trinoConnector = component "Trino Connector" "Подключение к Trino через SQL драйвер" "Go"
                    hiveConnector = component "Hive Connector" "Подключение к Hive/Spark через Thrift" "Go"
                    impalaConnector = component "Impala Connector" "Подключение к Impala" "Go"
                    verticaConnector = component "Vertica Connector" "Подключение к Vertica" "Go"
                }

                executorFactory = component "Executor Factory" "Фабрика для создания исполнителей запросов" "Go package: executor"

                sqlExecutor = component "SQL Executor" "Исполнитель SQL запросов для стандартных БД (Trino, Vertica, Impala)" "Go package: executor"

                hiveExecutor = component "Hive Executor" "Специализированный исполнитель для Hive/Spark с поддержкой HiveServer2" "Go package: executor"

                queryLoader = component "Query Loader" "Загрузка TPC-DS SQL запросов из файловой системы" "Go package: query"

                benchmarkRunner = component "Benchmark Runner" "Оркестратор выполнения бенчмарков с поддержкой многопоточности" "Go package: runner"

                csvStorage = component "CSV Storage" "Потокобезопасное сохранение результатов в CSV" "Go package: storage"
            }

            configFile = container "Configuration File" "YAML конфигурация с параметрами подключений и бенчмарков" "YAML file" "Configuration"

            queriesStorage = container "Queries Storage" "Директории с TPC-DS SQL запросами разной сложности" "File System" "Storage"

            resultsStorage = container "Results Storage" "CSV файл с результатами бенчмарков" "CSV file" "Storage"
        }

        # Взаимодействия между людьми и системами
        user -> tpcdsBenchmark "Запускает бенчмарки, настраивает конфигурацию"
        tpcdsBenchmark -> trinoWarehouse "Выполняет TPC-DS запросы" "SQL/HTTPS"
        tpcdsBenchmark -> hiveWarehouse "Выполняет TPC-DS запросы" "HiveServer2/Thrift"
        tpcdsBenchmark -> impalaWarehouse "Выполняет TPC-DS запросы" "SQL"
        tpcdsBenchmark -> verticaWarehouse "Выполняет TPC-DS запросы" "SQL"

        # Взаимодействия между контейнерами
        cliApp -> configFile "Читает конфигурацию" "File I/O"
        cliApp -> queriesStorage "Загружает SQL запросы" "File I/O"
        cliApp -> resultsStorage "Записывает результаты" "File I/O"

        cliApp -> trinoWarehouse "Подключается и выполняет запросы" "SQL/HTTPS"
        cliApp -> hiveWarehouse "Подключается и выполняет запросы" "Thrift"
        cliApp -> impalaWarehouse "Подключается и выполняет запросы" "SQL"
        cliApp -> verticaWarehouse "Подключается и выполняет запросы" "SQL"

        # Взаимодействия между компонентами
        mainController -> configLoader "Загружает конфигурацию"
        mainController -> connectionManager "Создает менеджер подключений"
        mainController -> benchmarkRunner "Запускает бенчмарк"

        benchmarkRunner -> executorFactory "Создает исполнителей для каждого warehouse"
        benchmarkRunner -> queryLoader "Загружает TPC-DS запросы"
        benchmarkRunner -> csvStorage "Сохраняет результаты"
        benchmarkRunner -> sqlExecutor "Выполняет запросы (Trino, Vertica, Impala)"
        benchmarkRunner -> hiveExecutor "Выполняет запросы (Hive, Spark)"

        executorFactory -> sqlExecutor "Создает для стандартных SQL БД"
        executorFactory -> hiveExecutor "Создает для Hive/Spark"
        executorFactory -> connectionManager "Получает подключения"

        sqlExecutor -> connectionManager "Получает SQL соединения"
        hiveExecutor -> connectionManager "Получает Hive соединения"

        connectionManager -> trinoConnector "Делегирует подключение к Trino"
        connectionManager -> hiveConnector "Делегирует подключение к Hive/Spark"
        connectionManager -> impalaConnector "Делегирует подключение к Impala"
        connectionManager -> verticaConnector "Делегирует подключение к Vertica"

        trinoConnector -> trinoWarehouse "Устанавливает TCP подключение" "SQL/HTTPS"
        hiveConnector -> hiveWarehouse "Устанавливает Thrift подключение" "HiveServer2"
        impalaConnector -> impalaWarehouse "Устанавливает TCP подключение" "SQL"
        verticaConnector -> verticaWarehouse "Устанавливает TCP подключение" "SQL"

        queryLoader -> queriesStorage "Читает .sql файлы" "File I/O"
        csvStorage -> resultsStorage "Записывает результаты" "File I/O"
        configLoader -> configFile "Читает конфигурацию" "File I/O"
    }

    views {
        # 1. System Context диаграмма - общий обзор системы
        systemContext tpcdsBenchmark "SystemContext" {
            include *
            autoLayout lr
            description "Системный контекст TPC-DS Benchmark - взаимодействие с пользователями и внешними хранилищами данных"
        }

        # 2. Container диаграмма - основные контейнеры системы
        container tpcdsBenchmark "Containers" {
            include *
            autoLayout lr
            description "Контейнеры TPC-DS Benchmark: CLI приложение, конфигурация, хранилища запросов и результатов"
        }

        # 3. Component диаграмма - детальная архитектура CLI приложения
        component cliApp "Components" {
            include *
            autoLayout tb
            description "Компоненты CLI приложения: управление конфигурацией, подключениями, выполнением запросов и хранением результатов"
        }

        # 4. Component диаграмма - Connection Manager детально
        component connectionManager "ConnectionManagerComponents" {
            include *
            include cliApp
            include trinoWarehouse
            include hiveWarehouse
            include impalaWarehouse
            include verticaWarehouse
            autoLayout lr
            description "Детальная архитектура Connection Manager с коннекторами для различных БД"
        }

        # 5. Dynamic диаграмма - поток выполнения бенчмарка
        dynamic cliApp "BenchmarkExecutionFlow" "Поток выполнения бенчмарка" {
            user -> mainController "1. Запускает приложение"
            mainController -> configLoader "2. Загружает конфигурацию"
            configLoader -> configFile "3. Читает YAML"
            mainController -> connectionManager "4. Создает Connection Manager"
            mainController -> benchmarkRunner "5. Создает Benchmark Runner"
            benchmarkRunner -> queryLoader "6. Загружает запросы"
            queryLoader -> queriesStorage "7. Читает .sql файлы"
            benchmarkRunner -> executorFactory "8. Создает executors для каждого warehouse"
            executorFactory -> connectionManager "9. Получает подключения"
            connectionManager -> trinoConnector "10. Создает подключение к Trino"
            trinoConnector -> trinoWarehouse "11. Устанавливает соединение"
            benchmarkRunner -> sqlExecutor "12. Выполняет запросы"
            sqlExecutor -> trinoWarehouse "13. Отправляет SQL запрос"
            trinoWarehouse -> sqlExecutor "14. Возвращает результат"
            sqlExecutor -> benchmarkRunner "15. Возвращает QueryResult"
            benchmarkRunner -> csvStorage "16. Сохраняет результат"
            csvStorage -> resultsStorage "17. Записывает в CSV"
            autoLayout lr
        }

        # Deployment диаграмма
        deployment tpcdsBenchmark "Production" "Production" {
            deploymentNode "User Workstation" "" "Linux/macOS/Windows" {
                deploymentNode "Go Runtime" "" "Go 1.25+" {
                    containerInstance cliApp
                }
                deploymentNode "Local File System" "" "" {
                    containerInstance configFile
                    containerInstance queriesStorage
                    containerInstance resultsStorage
                }
            }

            deploymentNode "Data Warehouse Infrastructure" "" "Cloud/On-Premise" {
                deploymentNode "Trino Cluster" "" "Kubernetes/VMs" {
                    softwareSystemInstance trinoWarehouse
                }
                deploymentNode "Hive/Spark Cluster" "" "Hadoop/Kubernetes" {
                    softwareSystemInstance hiveWarehouse
                }
                deploymentNode "Impala Cluster" "" "Cloudera" {
                    softwareSystemInstance impalaWarehouse
                }
                deploymentNode "Vertica Cluster" "" "VMs" {
                    softwareSystemInstance verticaWarehouse
                }
            }
        }

        # Стили
        styles {
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "External System" {
                background #999999
                color #ffffff
            }
            element "Person" {
                shape person
                background #08427b
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Component" {
                background #85bbf0
                color #000000
            }
            element "Configuration" {
                shape Folder
                background #f4a460
                color #000000
            }
            element "Storage" {
                shape Cylinder
                background #90ee90
                color #000000
            }
        }
    }

    configuration {
        scope softwaresystem
    }
}
