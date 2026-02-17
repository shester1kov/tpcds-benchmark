# Feature: S3 Upload — Фаза 1: Конфигурация

> Сгенерировано OwnYourCode. Спеки отражают ТВОИ решения, доработанные в ходе обсуждения.

## User Story

Как разработчик, запускающий бенчмарк, я хочу задавать параметры S3 в конфиге,
чтобы приложение знало, куда загружать результаты.

## Acceptance Criteria

Фаза считается завершённой, когда выполнено всё это:

- [ ] Структура `S3Config` объявлена в `pkg/config/config.go` с полями: `endpoint`, `access_key`, `secret_key`, `use_ssl`, `bucket`, `enabled`
- [ ] `S3Config` встроена в `Config` как указатель: `S3 *S3Config yaml:"s3"`
- [ ] В `Validate()` добавлена проверка: если `enabled: true` — все обязательные поля заполнены
- [ ] Секция `s3:` добавлена в `config/config.yaml` с примером значений
- [ ] Файл `config/config.yaml.example` обновлён аналогично

## Edge Cases

| Сценарий | Ожидаемое поведение |
|----------|---------------------|
| Секция `s3:` не указана в YAML | `cfg.S3 == nil`, приложение работает без S3 |
| `enabled: false` | Загрузка в S3 пропускается, ошибки нет |
| `enabled: true`, `endpoint` пустой | `Validate()` возвращает ошибку с понятным сообщением |
| `enabled: true`, `access_key` пустой | `Validate()` возвращает ошибку |
| `enabled: true`, `secret_key` пустой | `Validate()` возвращает ошибку |
| `enabled: true`, `bucket` пустой | `Validate()` возвращает ошибку |
| `enabled: true`, все поля заполнены | Валидация проходит успешно |

> **Твои edge cases (добавь если пропустил):**
> - [ ] _..._

## Out of Scope

Эта фаза НЕ включает:

- Логику подключения к S3 (это Фаза 2)
- Загрузку файлов (это Фаза 2)
- Интеграцию в benchmark runner (это Фаза 3)

## Dependencies

- [ ] Нет внешних зависимостей для этой фазы — только изменения в существующих файлах

## Research References

- Context7 (minio-go): `minio.New(endpoint, &minio.Options{Creds, Secure})` — поля конфига отражают именно этот API
- Существующий паттерн: `WarehouseConfig.enabled` — та же логика флага включения
