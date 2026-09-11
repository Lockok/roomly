# Roomly

Roomly — REST API для управления переговорными комнатами и бронированиями.

## Возможности

* управление переговорными комнатами;
* поиск свободных комнат по времени, вместимости, локации и оборудованию;
* создание, изменение и отмена бронирований;
* защита от пересечений бронирований на уровне PostgreSQL;
* календарь бронирований комнаты;
* регистрация и JWT-аутентификация;
* роли `employee` и `admin`;
* управление статусом пользователей;
* отчёт по загрузке переговорных;
* валидация входных данных;
* request ID и структурированное логирование;
* Docker Compose, миграции и health checks.

## Технологии

* Go 1.26
* PostgreSQL 16
* pgx/v5
* Goose
* Docker Compose
* JWT
* bcrypt
* `log/slog`
* UUID
* `testify`

## Быстрый запуск

1. Создайте локальный файл окружения:

```powershell
Copy-Item .env.example .env
```

Укажите значения в `.env`:

```env
POSTGRES_PASSWORD=your_secure_password
JWT_SECRET=your_long_random_secret
```

2. Запустите приложение:

```bash
docker compose up --build
```

3. Проверьте состояние API:

```powershell
Invoke-RestMethod http://localhost:8080/health/live
Invoke-RestMethod http://localhost:8080/health/ready
```

API доступен по адресу `http://localhost:8080`.

## Переменные окружения

| Переменная          | Описание                 | Пример                   |
| ------------------- | ------------------------ | ------------------------ |
| `APP_ENV`           | Окружение приложения     | `local`                  |
| `HTTP_PORT`         | HTTP-порт API            | `8080`                   |
| `POSTGRES_DB`       | Имя базы данных          | `roomly`                 |
| `POSTGRES_USER`     | Пользователь PostgreSQL  | `roomly`                 |
| `POSTGRES_PASSWORD` | Пароль PostgreSQL        | `change_me`              |
| `POSTGRES_PORT`     | Порт PostgreSQL на хосте | `5432`                   |
| `DATABASE_URL`      | URL подключения к БД     | `postgres://...`         |
| `JWT_SECRET`        | Секрет подписи JWT       | длинная случайная строка |
| `JWT_TTL`           | Время жизни access token | `24h`                    |

## Тесты

```bash
go test ./...
```

## Health checks

| Метод | Endpoint        | Описание                       |
| ----- | --------------- | ------------------------------ |
| `GET` | `/health/live`  | Приложение запущено            |
| `GET` | `/health/ready` | Приложение и PostgreSQL готовы |

## Аутентификация

После входа передавайте токен в заголовке:

```http
Authorization: Bearer <access_token>
```

### Регистрация сотрудника

```http
POST /api/v1/users
Content-Type: application/json
```

```json
{
  "email": "user@example.com",
  "password": "secure-password",
  "full_name": "User Name"
}
```

Публичная регистрация всегда создаёт пользователя с ролью `employee`.

### Вход

```http
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "email": "user@example.com",
  "password": "secure-password"
}
```

## Основные endpoints

| Метод   | Endpoint                          | Доступ             |
| ------- | --------------------------------- | ------------------ |
| `POST`  | `/api/v1/users`                   | публичный          |
| `POST`  | `/api/v1/auth/login`              | публичный          |
| `GET`   | `/api/v1/users`                   | публичный          |
| `GET`   | `/api/v1/rooms`                   | публичный          |
| `GET`   | `/api/v1/rooms/available`         | публичный          |
| `GET`   | `/api/v1/rooms/{id}`              | публичный          |
| `GET`   | `/api/v1/rooms/{id}/calendar`     | авторизованный     |
| `POST`  | `/api/v1/rooms`                   | admin              |
| `PATCH` | `/api/v1/rooms/{id}`              | admin              |
| `POST`  | `/api/v1/bookings`                | авторизованный     |
| `GET`   | `/api/v1/bookings`                | авторизованный     |
| `GET`   | `/api/v1/bookings/{id}`           | авторизованный     |
| `PATCH` | `/api/v1/bookings/{id}`           | владелец или admin |
| `POST`  | `/api/v1/bookings/{id}/cancel`    | владелец или admin |
| `POST`  | `/api/v1/admin/users`             | admin              |
| `PATCH` | `/api/v1/admin/users/{id}/status` | admin              |
| `GET`   | `/api/v1/reports/room-usage`      | admin              |
