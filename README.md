# Subscription Aggregator Service

REST API сервис для агрегации данных об онлайн подписках пользователей.

## Технологии

- **Go 1.22+**
- **PostgreSQL 16**
- **Gin** - HTTP фреймворк
- **sqlx** - работа с базой данных
- **golang-migrate** - миграции
- **zerolog** - структурированное логирование
- **viper** - конфигурация
- **swag** - Swagger документация
- **Docker & Docker Compose**

## Структура проекта

```
.
├── cmd/
│   └── server/
│       └── main.go          # Точка входа
├── internal/
│   ├── config/              # Конфигурация
│   ├── handler/             # HTTP хэндлеры
│   ├── model/               # Модели данных
│   ├── repository/          # Слой работы с БД
│   ├── server/              # HTTP сервер
│   └── service/             # Бизнес-логика
├── migrations/              # SQL миграции
├── docs/                    # Swagger документация
├── config.yaml              # Конфигурационный файл
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Быстрый старт

### С помощью Docker Compose

```bash
# Запуск всех сервисов
docker compose up -d --build

# Просмотр логов
docker compose logs -f

# Остановка
docker compose down
```

### Локальный запуск

1. Установите зависимости:
```bash
go mod download
```

2. Запустите PostgreSQL и выполните миграции:
```bash
# Запуск PostgreSQL
docker run -d --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=subscriptions \
  -p 5432:5432 \
  postgres:16-alpine

# Миграции
make migrate-up
```

3. Запустите сервис:
```bash
make run
```

## API Endpoints

### Подписки (CRUDL)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/subscriptions` | Создание подписки |
| GET | `/api/v1/subscriptions` | Список подписок |
| GET | `/api/v1/subscriptions/:id` | Получение подписки |
| PUT | `/api/v1/subscriptions/:id` | Обновление подписки |
| DELETE | `/api/v1/subscriptions/:id` | Удаление подписки |

### Аналитика

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/subscriptions/cost` | Суммарная стоимость за период |

### Health Check

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/health` | Проверка состояния сервиса |

## Примеры запросов

### Создание подписки

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Получение списка подписок

```bash
# Все подписки
curl http://localhost:8080/api/v1/subscriptions

# С фильтрацией по пользователю
curl "http://localhost:8080/api/v1/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"

# С пагинацией
curl "http://localhost:8080/api/v1/subscriptions?limit=10&offset=0"
```

### Расчет стоимости за период

```bash
curl "http://localhost:8080/api/v1/subscriptions/cost?start_date=01-2025&end_date=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"
```

## Конфигурация

Конфигурация осуществляется через `config.yaml` или переменные окружения:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_HOST` | Хост PostgreSQL | localhost |
| `DB_PORT` | Порт PostgreSQL | 5432 |
| `DB_USER` | Пользователь БД | postgres |
| `DB_PASSWORD` | Пароль БД | postgres |
| `DB_NAME` | Имя базы данных | subscriptions |
| `DB_SSLMODE` | SSL режим | disable |
| `SERVER_PORT` | Порт сервера | 8080 |
| `LOG_LEVEL` | Уровень логирования | info |

## Swagger документация

После запуска сервиса документация доступна по адресу:
```
http://localhost:8080/swagger/index.html
```

Для генерации документации:
```bash
make swag
```

## Makefile команды

```bash
make build        # Сборка приложения
make run          # Локальный запуск
make test         # Запуск тестов
make docker-up    # Запуск в Docker
make docker-down  # Остановка Docker
make migrate-up   # Применение миграций
make migrate-down # Откат миграций
make swag         # Генерация Swagger
make lint         # Запуск линтера
```
