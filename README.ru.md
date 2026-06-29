# Task Tracker

[English version](README.md)

Issue tracker с канбан-доской, RBAC, обновлениями в реальном времени и встроенным веб-интерфейсом. Написан на Go, PostgreSQL и Server-Sent Events.

## Возможности

- **Авторизация** — регистрация, вход, JWT access + refresh токены с ротацией, выход
- **Рабочие пространства** — создание, редактирование, удаление; приглашение участников с ролями
- **RBAC** — три роли: `owner`, `member`, `viewer` с гранулярными правами
- **Проекты** — CRUD, автогенерируемый ключ-префикс (например `BE-42`)
- **Задачи** — CRUD + перемещение по канбану (статус + позиция), фильтры по статусу/исполнителю/поиску
- **Метки** — создание, привязка/отвязка к задачам
- **Комментарии** — добавление комментариев к задачам
- **Лента активности** — полная история изменений по каждой задаче
- **Реалтайм** — SSE-поток для каждого проекта, мгновенные обновления доски
- **Health Checks** — `/healthz`, `/readyz`
- **Фронтенд** — vanilla JS SPA с drag-and-drop канбаном, авторизацией, SSE
- **CI/CD** — GitHub Actions: lint, test, build
- **Docker** — один `docker compose up` для запуска всего

## Технологический стек

| Компонент | Технология |
|---|---|
| Язык | Go 1.26 |
| HTTP-роутер | [chi/v5](https://github.com/go-chi/chi) |
| БД | PostgreSQL 16 |
| Драйвер | pgx/v5 (raw queries) |
| Миграции | [goose](https://github.com/pressly/goose) |
| Аутентификация | JWT (access) + refresh токены |
| Пароли | bcrypt |
| Реалтайм | Server-Sent Events (in-memory Hub) |
| Контейнеризация | Docker, Docker Compose |
| CI | GitHub Actions |
| Фронтенд | Vanilla JS |

## Архитектура

```
HTTP Handler → Service (use cases) → Repository → PostgreSQL
                      ↓
              Hub (in-memory) → SSE → клиенты
```

### Слои

| Слой | Ответственность |
|---|---|
| `handler` | Парсинг HTTP, валидация DTO, маппинг ошибок |
| `service` | Бизнес-логика, транзакции, RBAC, публикация событий |
| `repository` | SQL-запросы через pgx |
| `domain` | Типы, enum'ы, доменные ошибки |

### Структура проекта

```
task-tracker/
├── cmd/api/main.go              # точка входа
├── internal/
│   ├── auth/                    # JWT сервис
│   ├── authz/                   # RBAC роли и права
│   ├── cache/                   # слой кэширования
│   ├── config/                  # загрузчик конфигурации из env
│   ├── domain/                  # типы, enum'ы, валидация
│   ├── events/                  # SSE hub
│   ├── http/
│   │   ├── router.go            # chi маршруты
│   │   ├── middleware/          # auth, CORS, RBAC middleware
│   │   └── handler/             # HTTP-хендлеры
│   ├── repository/
│   │   └── postgres/            # pgx репозитории
│   └── service/                 # бизнес-логика
├── db/migrations/               # 11 миграций goose
├── web/                         # фронтенд SPA
│   ├── index.html
│   ├── app.js
│   └── style.css
├── api/openapi.yaml             # спецификация API
├── docs/TZ.md                   # техническое задание
├── docker-compose.yaml
├── Dockerfile
└── Makefile
```
```

## Быстрый старт

### Требования

- Docker & Docker Compose
- Go 1.26+ (для локальной разработки)
- [goose](https://github.com/pressly/goose) (для миграций)

### Запуск через Docker

```bash
cp .env.example .env
# отредактируйте .env при необходимости
make up-build
```

Приложение будет доступно на `http://localhost:8080`.

### Локальная разработка

```bash
# запустить только postgres
docker compose up -d postgres

# применить миграции
make migrate-up

# запустить API сервер
make run
```

### Команды Make

| Команда | Описание |
|---|---|
| `make up` | Запустить все сервисы |
| `make up-build` | Собрать и запустить все сервисы |
| `make down` | Остановить все сервисы |
| `make run` | Запустить API-сервер локально |
| `make migrate-up` | Применить все миграции |
| `make migrate-down` | Откатить последнюю миграцию |
| `make migrate-status` | Показать статус миграций |
| `make logs` | Логи контейнеров |

## API

Базовый путь: `/api/v1`. Формат: JSON.

### Auth

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| POST | `/auth/register` | — | Регистрация |
| POST | `/auth/login` | — | Вход |
| POST | `/auth/refresh` | — | Обновление токенов |
| POST | `/auth/logout` | — | Выход (отзыв refresh) |
| GET | `/me` | JWT | Текущий пользователь |

### Рабочие пространства

| Метод | Путь | Роль | Описание |
|---|---|---|---|
| GET | `/workspaces` | auth | Список workspace |
| POST | `/workspaces` | auth | Создать workspace |
| GET | `/workspaces/{id}` | auth | Детали workspace |
| PATCH | `/workspaces/{id}` | owner | Обновить workspace |
| DELETE | `/workspaces/{id}` | owner | Удалить workspace |
| GET | `/workspaces/{id}/members` | auth | Список участников |
| POST | `/workspaces/{id}/members` | owner | Добавить участника |
| PATCH | `/workspaces/{id}/members/{userId}` | owner | Сменить роль |
| DELETE | `/workspaces/{id}/members/{userId}` | owner | Удалить участника |
| POST | `/workspaces/{id}/invites` | member+ | Создать invite |
| GET | `/workspaces/{id}/invites` | auth | Список invites |
| POST | `/invites/{token}/accept` | auth | Принять invite |

### Проекты

| Метод | Путь | Роль | Описание |
|---|---|---|---|
| GET | `/workspaces/{wsId}/projects` | auth | Список проектов |
| POST | `/workspaces/{wsId}/projects` | member+ | Создать проект |
| GET | `/projects/{id}` | auth | Детали проекта |
| PATCH | `/projects/{id}` | member+ | Обновить проект |
| DELETE | `/projects/{id}` | owner | Удалить проект |

### Задачи

| Метод | Путь | Роль | Описание |
|---|---|---|---|
| GET | `/projects/{id}/issues` | auth | Список (фильтры: status, assignee, q) |
| POST | `/projects/{id}/issues` | member+ | Создать задачу |
| GET | `/issues/{id}` | auth | Детали задачи |
| PATCH | `/issues/{id}` | member+ | Обновить поля |
| DELETE | `/issues/{id}` | member+ | Удалить задачу |
| PATCH | `/issues/{id}/move` | member+ | Перемещение (статус + позиция) |

### Комментарии и активность

| Метод | Путь | Роль | Описание |
|---|---|---|---|
| GET | `/issues/{id}/comments` | auth | Список комментариев |
| POST | `/issues/{id}/comments` | member+ | Добавить комментарий |
| GET | `/issues/{id}/activity` | auth | Лента активности |

### Метки

| Метод | Путь | Роль | Описание |
|---|---|---|---|
| GET | `/workspaces/{wsId}/labels` | auth | Список меток |
| POST | `/workspaces/{wsId}/labels` | member+ | Создать метку |
| GET | `/issues/{id}/labels` | auth | Метки задачи |
| POST | `/issues/{id}/labels/{labelId}` | member+ | Привязать метку |
| DELETE | `/issues/{id}/labels/{labelId}` | member+ | Отвязать метку |

### Реалтайм (SSE)

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| GET | `/projects/{id}/events` | JWT | SSE-поток событий |

События: `issue.created`, `issue.updated`, `issue.moved`, `issue.deleted`, `comment.added`

### Health Checks

| Метод | Путь | Описание |
|---|---|---|
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |

## RBAC

| Действие | owner | member | viewer |
|---|:---:|:---:|:---:|
| Просмотр проектов и задач | ✅ | ✅ | ✅ |
| Создание/редактирование/удаление задач | ✅ | ✅ | ❌ |
| Создание/редактирование проектов | ✅ | ✅ | ❌ |
| Удаление проектов | ✅ | ❌ | ❌ |
| Управление метками | ✅ | ✅ | ❌ |
| Создание invites | ✅ | ✅ | ❌ |
| Изменение ролей участников | ✅ | ❌ | ❌ |
| Удаление workspace | ✅ | ❌ | ❌ |

## Тестирование

```bash
# запустить все тесты
go test ./...

# с детектированием гонок и покрытием
go test -race -coverprofile=coverage.out ./...
```

24 тестовых файла: валидация домена, матрица RBAC, бизнес-логика сервисов, HTTP-хендлеры.

## Лицензия

MIT
