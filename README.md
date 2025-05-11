# Calculator

Финальный проект годового курса Go от Я.Лицея

## Установка

 - Для установки нужно выбрать директорию проекта:
```bash
cd <your_dir>
```
 - Потом необходимо выполнить эту команду:
```bash
git clone https://github.com/xKARASb/Calculator
cd Calculator
```

## Использование

### Docker Compose запуск (рекомендуется)
```
# Запускаем сервисы через Docker Compose
docker-compose build
docker-compose up -d
```

Сервисы будут доступны по адресу `http://localhost:8080`

### Остановка сервисов

```bash
docker-compose down
```

### Запуск из исходников

1. Сначала необходимо открыть файл ```example.env``` и установить ваши параметры вместо дефолтных:
```env
PORT=8080
COMPOUNDING_POWER=10

POSTGRES_USER=postgres
POSTGRES_PASSWORD=123
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=users

REDIS_HOST=localhost
REDIS_PORT=6379

ORCHESTRATOR_HOST=localhost
```
2. Переименуйте ```example.env``` -> ```.env```

3. Проверьте, что PostgreSQL и Redis запущены согласно конфигу
4. Запустите оркестратор и агенты паралельно:

```bash
go run ./cmd/orchestrator/main.go

go run ./cmd/agent/main.go
```

> [!IMPORTANT]
> В Docker Compose окружении сервисы используют имена сервисов вместо `localhost`. При запуске напрямую из Go необходимо изменить адреса в `.env` файле для правильного соединения с базами данных и между сервисами.

## Использование API
#### Добавление вычисления арифметического выражения

##### Curl
### Расчёт выражения

#### Запрос

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```

#### Успешный ответ (HTTP 200)

```json
{
  "id": 1
}
```

### Примеры с различными выражениями

#### Простое сложение

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2"
}'
```

#### Сложное выражение со скобками

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "(10+2)*2"
}'
```

### Обработка ошибок

#### Деление на ноль

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "1/0"
}'
```

#### Ответ (HTTP 400)

```json
{
  "message": "division by zero"
}
```

#### Некорректный синтаксис

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2++2"
}'
```

#### Ответ (HTTP 400)

```json
{
  "message": "invalid expression"
}
```

## Структура проекта

```
.
├── cmd/                    # Точки входа
│   ├── agent/              # Запуск агента
│   └── orchestrator/       # Запуск оркестратора
├── internal/               # Внутренняя логика
│   ├── agent/              # Логика агента
│   ├── config/             # Конфигурация
│   └── orchestrator/       # Логика оркестратора
├── pkg/                    # Общие пакеты
│   ├── db/                 # Работа с базами данных
│   ├── models/             # Модели данных
│   ├── tests/              # Тесты
│   └── utils/              # Утилиты
├── dockerfiles/            # Докерфайлы
├── migrations/             # Миграции БД
└── web/                    # Веб-интерфейс
```

## Веб-интерфейс

Обыкновенный веб-интерфейс: `http://localhost:8080/`

## Тестирование

```bash
# Запуск unit-тестов
go test ./pkg/tests -v

# Запуск только интеграционных тестов
go test ./pkg/tests -v -run TestIntegration

# Пропуск интеграционных тестов
SKIP_INTEGRATION=true go test ./pkg/tests -v
```

## Дополнительные возможности

## Автор

[xKARASb](https://github.com/xkarasb)