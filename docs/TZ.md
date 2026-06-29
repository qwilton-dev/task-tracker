# Техническое задание: Task Tracker (мини-Linear)

| Поле             | Значение                                        |
| ---------------- | ----------------------------------------------- |
| Версия документа | 2.0                                             |
| Дата             | 2026-06-16                                      |
| Статус           | Implemented                                     |
| Цель             | Pet-проект для портфолио: backend на Go с production-практиками |

---

## 1. Назначение и цели

### 1.1. Назначение

Веб-сервис для управления задачами в формате issue tracker: команды (workspace), проекты, kanban-доска, комментарии, история изменений, обновления в реальном времени.

### 1.2. Цели проекта

- Продемонстрировать в портфолио навыки backend-разработки на Go.
- Показать работу с PostgreSQL, миграциями, RBAC, JWT, SSE.
- Дать рекрутеру и интервьюеру понятный demo и README с архитектурными решениями.

### 1.3. Не входит в scope

- Микросервисная архитектура, Kubernetes.
- Спринты, roadmap, GitHub-интеграции, вложения файлов.
- Email/push-уведомления.
- Оплата, billing, multi-tenant SaaS.
- Полноценный клон Linear (cycles, insights, automations).
- Rate limiting, Prometheus метрики, structured logging (v2).
- WebSocket (presence, typing).

### 1.4. Целевая аудитория продукта

- Небольшие команды (2–20 человек) или личные проекты.
- Пользователь с аккаунтом создаёт workspace и приглашает участников.

---

## 2. Технологический стек

| Компонент           | Технология                     | Назначение                              |
| ------------------- | ------------------------------ | --------------------------------------- |
| Язык                | Go 1.26+                       | API-сервер                              |
| HTTP-роутер         | chi/v5                         | Маршрутизация, middleware               |
| СУБД                | PostgreSQL 16                  | Источник правды                         |
| Доступ к БД         | pgx/v5 (raw queries)           | Типобезопасные запросы                  |
| Миграции            | goose                          | Версионирование схемы                   |
| Аутентификация      | JWT (access) + refresh token   | Stateless API с возможностью отзыва     |
| Хеширование паролей | bcrypt                         | Хранение `password_hash`                |
| Контейнеризация     | Docker, Docker Compose         | Локальная и demo-среда                  |
| CI                  | GitHub Actions                 | lint, test, build                       |
| Тесты               | go test (unit)                 | Domain, service, handler тесты          |
| Frontend            | Vanilla JS                     | Kanban-доска, SSE, CRUD                 |

### 2.1. Инфраструктура локально

```yaml
# docker-compose: api, postgres
services:
  - api (порт 8080)
  - postgres (5432)
```

---

## 3. Архитектура

### 3.1. Стиль

Монолитное приложение, **слоистая архитектура**:

```
HTTP Handler → Service (use cases) → Repository → PostgreSQL
                      ↓
              Hub (in-memory) → SSE → клиенты
```

### 3.2. Принципы слоёв

| Слой        | Ответственность                                        | Запрещено                       |
| ----------- | ------------------------------------------------------ | ------------------------------- |
| `handler`   | Парсинг HTTP, валидация DTO, маппинг ошибок в codes   | Бизнес-логика, прямой SQL       |
| `service`   | Бизнес-правила, транзакции, RBAC-вызовы, pub events   | Знание `http.ResponseWriter`    |
| `repository`| SQL-запросы (pgx)                                      | Бизнес-правила                  |
| `domain`    | Типы, enum, доменные ошибки                            | Зависимости от инфраструктуры   |

### 3.3. Структура репозитория

```
task-tracker/
├── cmd/api/main.go
├── internal/
│   ├── config/
│   ├── domain/
│   ├── http/
│   │   ├── router.go
│   │   ├── middleware/
│   │   └── handler/
│   ├── service/
│   ├── repository/
│   │   └── postgres/
│   ├── auth/
│   ├── authz/
│   └── events/
├── db/
│   └── migrations/
├── web/
│   ├── index.html
│   ├── app.js
│   └── style.css
├── api/openapi.yaml
├── docs/
│   └── TZ.md
├── docker-compose.yaml
├── Dockerfile
├── Makefile
└── README.md
```

### 3.4. Транзакционные границы

Операции, требующие одной транзакции PostgreSQL:

- Создание issue: нумерация `MAX(number) + 1` + INSERT в одной транзакции (`CreateIssueTx` с `FOR UPDATE`).

Операции без транзакции (at-least-once через in-memory Hub):

- После успешного commit — публикация события в Hub → SSE клиенты.

### 3.5. Real-time

- **Протокол:** Server-Sent Events (SSE).
- **In-memory Hub:** события публикуются напрямую в Hub (map с buffered channels).
- **SSE endpoint:** `GET /projects/{id}/events?token=JWT`.

---

## 4. Доменная модель

### 4.1. Диаграмма сущностей

```
User ──< WorkspaceMember >── Workspace ──< Project ──< Issue
  │                              │              │
  │                              └──< Label      ├──< Comment
  │                                  │           ├──< IssueLabel >── Label
  └──< RefreshToken                └──< Invite   └──< ActivityEvent
```

### 4.2. Сущности

#### User

| Поле          | Тип         | Ограничения       |
| ------------- | ----------- | ----------------- |
| id            | UUID        | PK                |
| email         | string      | UNIQUE, NOT NULL  |
| password_hash | string      | NOT NULL          |
| name          | string      | NOT NULL          |
| created_at    | timestamptz | NOT NULL          |
| updated_at    | timestamptz | NOT NULL          |

#### Workspace

| Поле       | Тип         | Ограничения       |
| ---------- | ----------- | ----------------- |
| id         | UUID        | PK                |
| name       | string      | NOT NULL          |
| slug       | string      | UNIQUE, NOT NULL  |
| created_at | timestamptz | NOT NULL          |

#### WorkspaceMember

| Поле         | Тип         | Ограничения                        |
| ------------ | ----------- | ---------------------------------- |
| workspace_id | UUID        | FK, PK (composite)                 |
| user_id      | UUID        | FK, PK (composite)                 |
| role         | text        | CHECK: `owner`, `member`, `viewer` |

#### Project

| Поле         | Тип         | Ограничения                      |
| ------------ | ----------- | -------------------------------- |
| id           | UUID        | PK                               |
| workspace_id | UUID        | FK, NOT NULL                     |
| name         | string      | NOT NULL                         |
| key          | string      | 2–5 символов, A-Z, UNIQUE/ws    |
| created_at   | timestamptz | NOT NULL                         |

#### Issue

| Поле         | Тип         | Ограничения                         |
| ------------ | ----------- | ----------------------------------- |
| id           | UUID        | PK                                  |
| project_id   | UUID        | FK, NOT NULL                        |
| number       | int         | NOT NULL, UNIQUE (project_id, number) |
| title        | string      | NOT NULL, max 500                   |
| description  | text        | nullable                            |
| status       | text        | `backlog`, `todo`, `in_progress`, `review`, `done` |
| priority     | text        | `none`, `low`, `medium`, `high`, `urgent` |
| assignee_id  | UUID        | FK users, nullable                  |
| position     | numeric     | NOT NULL, для сортировки            |
| created_by   | UUID        | FK users, NOT NULL                  |
| created_at   | timestamptz | NOT NULL                            |
| updated_at   | timestamptz | NOT NULL                            |

Отображаемый идентификатор: `{project.key}-{number}` (например `BE-42`).

#### Label

| Поле         | Тип  | Ограничения                  |
| ------------ | ---- | ---------------------------- |
| id           | UUID | PK                           |
| workspace_id | UUID | FK                           |
| name         | text | UNIQUE (workspace_id, name)  |
| color        | text | hex `#RRGGBB`                |

#### IssueLabel

| Поле      | Тип         | Ограничения                     |
| --------- | ----------- | ------------------------------- |
| issue_id  | UUID        | FK ON DELETE CASCADE, PK part   |
| label_id  | UUID        | FK ON DELETE CASCADE, PK part   |
| created_at| timestamptz | NOT NULL                        |

#### Comment

| Поле      | Тип         | Ограничения   |
| --------- | ----------- | ------------- |
| id        | UUID        | PK            |
| issue_id  | UUID        | FK            |
| author_id | UUID        | FK users      |
| body      | text        | NOT NULL      |
| created_at| timestamptz | NOT NULL      |
| updated_at| timestamptz | nullable      |

#### ActivityEvent

| Поле      | Тип         | Ограничения   |
| --------- | ----------- | ------------- |
| id        | UUID        | PK            |
| issue_id  | UUID        | FK            |
| actor_id  | UUID        | FK users      |
| event_type| text        | NOT NULL      |
| payload   | JSONB       | NOT NULL      |
| created_at| timestamptz | NOT NULL      |

#### RefreshToken

| Поле       | Тип         | Ограничения |
| ---------- | ----------- | ----------- |
| id         | UUID        | PK          |
| user_id    | UUID        | FK          |
| token_hash | varchar     | NOT NULL    |
| expires_at | timestamptz | NOT NULL    |
| revoked_at | timestamptz | nullable    |
| created_at | timestamptz | NOT NULL    |

#### Invite

| Поле         | Тип         | Ограничения                        |
| ------------ | ----------- | ---------------------------------- |
| id           | UUID        | PK                                 |
| workspace_id | UUID        | FK                                 |
| email        | text        | NOT NULL                           |
| role         | text        | CHECK: `member`, `viewer`          |
| token        | text        | UNIQUE, NOT NULL                   |
| expires_at   | timestamptz | NOT NULL                           |
| accepted_at  | timestamptz | nullable                           |
| created_by   | UUID        | FK users                           |
| created_at   | timestamptz | NOT NULL                           |

### 4.3. Типы ActivityEvent

| type               | payload                          |
| ------------------ | -------------------------------- |
| `issue.created`    | `{ "title" }`                    |
| `issue.updated`    | `{ "title" }`                    |
| `issue.moved`      | `{ "title" }`                    |
| `issue.deleted`    | `{ "title" }`                    |
| `comment.added`    | `{ "comment_id" }`               |
| `issue.label_added`| `{ "label_id", "name" }`         |
| `issue.label_removed` | `{ "label_id", "name" }`      |

### 4.4. Индексы

- `issues (project_id, status)`
- `issues (project_id, number)` — UNIQUE
- `issues (assignee_id)` WHERE assignee_id IS NOT NULL
- `activity_events (issue_id, created_at DESC)`
- `comments (issue_id, created_at)`
- `workspace_member (user_id)`
- `project (workspace_id)`
- `issue_label (label_id)`
- `invites (workspace_id)`, `invites (token)`

---

## 5. RBAC

### 5.1. Область проверки

Права проверяются на уровне **workspace** (доступ к project/issue через membership в workspace проекта).

### 5.2. Матрица прав

| Действие                       | owner | member | viewer |
| ------------------------------ | :---: | :----: | :----: |
| Просмотр projects, issues      | ✓     | ✓      | ✓      |
| Создание/редактирование issue  | ✓     | ✓      | ✗      |
| Удаление issue                 | ✓     | ✓      | ✗      |
| Создание/редактирование project| ✓     | ✓      | ✗      |
| Удаление project               | ✓     | ✗      | ✗      |
| Управление labels              | ✓     | ✓      | ✗      |
| Создание invite                | ✓     | ✓      | ✗      |
| Изменение ролей members        | ✓     | ✗      | ✗      |
| Удаление workspace             | ✓     | ✗      | ✗      |

### 5.3. Реализация

- Пакет `internal/authz`: `Role` (owner/member/viewer), `Action` matrix, `Can()`, `AtLeast()`.
- HTTP middleware: `RequireAuth`, `RequireRole` с resolver'ами по workspaceID, workspaceSlug, projectID, issueID.
- Ошибка при отсутствии прав: HTTP 403.

---

## 6. Аутентификация и сессии

### 6.1. Регистрация и вход

- `POST /api/v1/auth/register` — email, password (≥8 символов), name.
- `POST /api/v1/auth/login` — email, password → access + refresh.
- `POST /api/v1/auth/refresh` — refresh token → новая пара токенов (ротация).
- `POST /api/v1/auth/logout` — отзыв refresh (revoked_at).
- `GET /api/v1/me` — текущий пользователь.

### 6.2. JWT Access Token

- Время жизни: **15 минут**.
- Claims: `sub` (user_id), `exp`, `iat`, `iss`, `aud`.
- Передача: заголовок `Authorization: Bearer <token>` или query `?token=`.

### 6.3. Refresh Token

- Время жизни: **7 дней**.
- Хранение: только хеш (SHA-256) в таблице `refresh_tokens`.
- Ротация при refresh (старый токен revoke, выдача нового).

---

## 7. API (REST v1)

Базовый префикс: `/api/v1`. Формат: JSON. Ошибки: `{ "error": { "code": "...", "message": "..." } }`.

### 7.1. Auth

| Method | Path                | Auth | Описание         |
|--------|---------------------|------|------------------|
| POST   | `/auth/register`    | —    | Регистрация      |
| POST   | `/auth/login`       | —    | Вход             |
| POST   | `/auth/refresh`     | —    | Обновление токенов |
| POST   | `/auth/logout`      | —    | Выход            |
| GET    | `/me`               | auth | Текущий пользователь |

### 7.2. Workspaces

| Method | Path                                | Min role | Описание                |
|--------|-------------------------------------|----------|-------------------------|
| GET    | `/workspaces`                       | auth     | Список workspace        |
| POST   | `/workspaces`                       | auth     | Создать (создатель = owner) |
| GET    | `/workspaces/{id}`                  | auth     | Детали                  |
| PATCH  | `/workspaces/{id}`                  | owner    | Обновить name/slug      |
| DELETE | `/workspaces/{id}`                  | owner    | Удалить                 |
| GET    | `/workspaces/{id}/members`          | auth     | Список участников       |
| POST   | `/workspaces/{id}/members`          | owner    | Добавить member         |
| PATCH  | `/workspaces/{id}/members/{userId}` | owner    | Сменить роль            |
| DELETE | `/workspaces/{id}/members/{userId}` | owner    | Удалить участника       |
| POST   | `/workspaces/{id}/invites`          | member+  | Создать invite          |
| GET    | `/workspaces/{id}/invites`          | auth     | Список invites          |
| POST   | `/invites/{token}/accept`           | auth     | Принять приглашение     |

### 7.3. Projects

| Method | Path                                | Min role | Описание        |
|--------|-------------------------------------|----------|-----------------|
| GET    | `/workspaces/{wsId}/projects`       | auth     | Список проектов |
| POST   | `/workspaces/{wsId}/projects`       | member+  | Создать проект  |
| GET    | `/projects/{id}`                    | auth     | Детали          |
| PATCH  | `/projects/{id}`                    | member+  | Обновить        |
| DELETE | `/projects/{id}`                    | owner    | Удалить         |

### 7.4. Issues

| Method | Path                        | Min role | Описание                    |
|--------|-----------------------------|----------|-----------------------------|
| GET    | `/projects/{id}/issues`     | auth     | Список (фильтры: status, assignee, q) |
| POST   | `/projects/{id}/issues`     | member+  | Создать issue               |
| GET    | `/issues/{id}`              | auth     | Детали                      |
| PATCH  | `/issues/{id}`              | member+  | Обновить поля               |
| DELETE | `/issues/{id}`              | member+  | Удалить                     |
| PATCH  | `/issues/{id}/move`         | member+  | Смена status и position     |

### 7.5. Comments и Activity

| Method | Path                    | Min role | Описание          |
|--------|-------------------------|----------|-------------------|
| GET    | `/issues/{id}/comments` | auth     | Список комментариев |
| POST   | `/issues/{id}/comments` | member+  | Добавить комментарий |
| GET    | `/issues/{id}/activity` | auth     | Лента activity     |

### 7.6. Labels

| Method | Path                                | Min role | Описание         |
|--------|-------------------------------------|----------|------------------|
| GET    | `/workspaces/{wsId}/labels`         | auth     | Список labels    |
| POST   | `/workspaces/{wsId}/labels`         | member+  | Создать label    |
| GET    | `/issues/{id}/labels`               | auth     | Labels issue     |
| POST   | `/issues/{id}/labels/{labelId}`     | member+  | Привязать label  |
| DELETE | `/issues/{id}/labels/{labelId}`     | member+  | Отвязать label   |

### 7.7. Real-time (SSE)

| Method | Path                      | Auth | Описание               |
|--------|---------------------------|------|------------------------|
| GET    | `/projects/{id}/events`   | auth | SSE-поток событий проекта |

Типы событий: `issue.created`, `issue.updated`, `issue.moved`, `issue.deleted`, `comment.added`.

### 7.8. Служебные endpoints

| Method | Path       | Описание  |
|--------|------------|-----------|
| GET    | `/healthz` | Liveness  |
| GET    | `/readyz`  | Readiness |

---

## 8. Бизнес-правила

### 8.1. Нумерация issue

- В транзакции: `number = COALESCE(MAX(number), 0) + 1` для `project_id` с `FOR UPDATE`.
- Уникальность: `(project_id, number)`.

### 8.2. Позиция на kanban (move)

- Клиент передаёт `status` и `position` (numeric).
- v1: простой numeric.

### 8.3. Удаление

- Hard delete с ON DELETE CASCADE.
- Удаление workspace каскадно удаляет projects, members, labels, invites.

### 8.4. Invites

- Срок действия: **7 дней**.
- Токен: 64 hex-символа (crypto/rand).
- Повторное принятие: 409 `INVITE_ALREADY_ACCEPTED`.
- Если пользователь уже участник — пропуск добавления.

### 8.5. Валидация

- Email: формат RFC, unique.
- Password: минимум 8 символов.
- Project key: `^[A-Z]{2,5}$`, unique в workspace.
- Title issue: 1–500 символов.
- Label color: hex `#RRGGBB`.
- Invite role: `member` или `viewer`.
- Member role: `owner`, `member` или `viewer`.

---

## 9. Observability

### 9.1. Логирование

- Стандартный `log` с chi middleware (RequestID, RealIP, Logger).

### 9.2. Health checks

- `/healthz` — процесс жив.
- `/readyz` — OK.

### 9.3. Graceful shutdown

- По SIGINT/SIGTERM: `srv.Shutdown()` с timeout 30s.

---

## 10. Тестирование

### 10.1. Unit-тесты (16 файлов)

- `internal/domain/`: user, project, issue, label, comment — валидация, нормализация.
- `internal/authz/`: матрица прав для всех ролей.
- `internal/service/`: auth, workspace, project, issue, label — бизнес-логика с моками.
- `internal/http/handler/`: auth, workspace, project, issue — HTTP handlers с моками.
- `internal/repository/postgres/`: error mapping.

### 10.2. CI (GitHub Actions)

1. `golangci-lint run`
2. `go test ./...`
3. `go build -o bin/api ./cmd/api`

---

## 11. Нефункциональные требования

| Требование            | Значение                     |
| --------------------- | ---------------------------- |
| Concurrent users      | до 50                        |
| Хранение данных       | PostgreSQL persistent volume |
| Безопасность          | secrets через env            |
| CORS                  | `CORS_ORIGINS`               |
| Пароли                | bcrypt, ≥8 символов          |
| Refresh token         | SHA-256 hash в БД            |

---

## 12. ADR (зафиксированные решения)

| ID     | Решение                          | Альтернатива       | Причина                          |
| ------ | -------------------------------- | ------------------ | -------------------------------- |
| ADR-001| pgx (raw queries)                | sqlc, GORM         | Прямой контроль над SQL          |
| ADR-002| SSE + in-memory Hub              | Redis pub/sub      | Single-instance, проще MVP       |
| ADR-003| RBAC на workspace                | Per-project roles  | Меньше сложности                 |
| ADR-004| UUID v4                          | bigint             | Удобство в API                   |
| ADR-005| Hard delete v1                   | soft delete        | Простота                         |
| ADR-006| Issue number в Tx с FOR UPDATE   | Глобальная sequence| Читаемый ключ per project        |
| ADR-007| Vanilla JS frontend              | React/Vue          | Без сборки, портативность        |

---

## 13. Фронтенд

Vanilla JS SPA в `web/`:

- Авторизация (login/register), JWT в localStorage.
- Kanban-доска с drag-and-drop перемещением issues.
- Создание/редактирование/удаление workspace, project, issue.
- Управление участниками и инвайтами.
- Управление labels, привязка к issues.
- Комментарии и activity лента с читаемыми именами.
- SSE: реалтайм обновления доски без перезагрузки.

---

## 14. Статус реализации

| Компонент                | Статус |
| ------------------------ | ------ |
| Auth (register/login/refresh/logout/me) | Done |
| Workspace CRUD          | Done |
| Workspace members       | Done |
| Workspace invites       | Done |
| Project CRUD            | Done |
| Issue CRUD + move       | Done |
| Comments                | Done |
| Activity events         | Done |
| Labels + attach/detach  | Done |
| SSE real-time           | Done |
| RBAC (3 роли)           | Done |
| 11 миграций             | Done |
| Graceful shutdown       | Done |
| Unit-тесты (16 файлов)  | Done |
| Frontend kanban         | Done |
| CI (GitHub Actions)     | Done |
| Docker                  | Done |

---

*Документ версии 2.0 — отражает реализованный scope v1.*
