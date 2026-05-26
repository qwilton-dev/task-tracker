# Техническое задание: Task Tracker (мини-Linear)

| Поле | Значение |
|------|----------|
| Версия документа | 1.0 |
| Дата | 2026-05-25 |
| Статус | Draft → Ready for implementation |
| Цель | Pet-проект для портфолио: backend на Go с production-практиками |

---

## 1. Назначение и цели

### 1.1. Назначение

Веб-сервис для управления задачами в формате issue tracker: команды (workspace), проекты, kanban-доска, комментарии, история изменений, обновления в реальном времени.

### 1.2. Цели проекта

- Продемонстрировать в портфолио навыки backend-разработки на Go.
- Показать работу с PostgreSQL, миграциями, RBAC, JWT, Redis, observability.
- Дать рекрутеру и интервьюеру понятный demo и README с архитектурными решениями.

### 1.3. Не входит в scope (v1)

- Микросервисная архитектура, Kubernetes.
- Спринты, roadmap, GitHub-интеграции, вложения файлов.
- Email/push-уведомления (только in-app через SSE в v1).
- Оплата, billing, multi-tenant SaaS на уровне изоляции данных beyond workspace.
- Полноценный клон Linear (cycles, insights, automations).

### 1.4. Целевая аудитория продукта

- Небольшие команды (2–20 человек) или личные проекты.
- Пользователь с аккаунтом создаёт workspace и приглашает участников.

---

## 2. Технологический стек

| Компонент | Технология | Назначение |
|-----------|------------|------------|
| Язык | Go 1.23+ | API-сервер |
| HTTP-роутер | chi | Маршрутизация, middleware |
| СУБД | PostgreSQL 16 | Источник правды |
| Доступ к БД | pgx + sqlc | Типобезопасные запросы |
| Миграции | goose | Версионирование схемы |
| Кэш / bus | Redis 7 | Rate limit, pub/sub для SSE, опциональный кэш доски |
| Аутентификация | JWT (access) + refresh token в БД | Stateless API с возможностью отзыва сессий |
| Хеширование паролей | bcrypt или argon2id | Хранение `password_hash` |
| Валидация входа | go-playground/validator | DTO |
| Логирование | zap или slog | Structured JSON-логи |
| Метрики | prometheus/client_golang | Endpoint `/metrics` |
| Трейсинг (опционально) | OpenTelemetry | v1.1 или неделя 3 |
| Контейнеризация | Docker, Docker Compose | Локальная и demo-среда |
| CI | GitHub Actions | lint, test, build |
| Тесты | testify, testcontainers-go | Unit + интеграционные |
| API-документация | OpenAPI 3.x | `api/openapi.yaml` |

### 2.1. Инфраструктура локально

```yaml
# docker-compose: api, postgres, redis
services:
  - api (порт 8080)
  - postgres (5432)
  - redis (6379)
```

---

## 3. Архитектура

### 3.1. Стиль

Монолитное приложение, **слоистая архитектура**:

```
HTTP Handler → Service (use cases) → Repository → PostgreSQL
                      ↓
              Event Publisher → Redis Pub/Sub → SSE Hub → клиенты
                      ↓
              Cache (Redis) — опционально
```

### 3.2. Принципы слоёв

| Слой | Ответственность | Запрещено |
|------|-----------------|-----------|
| `handler` | Парсинг HTTP, валидация DTO, маппинг ошибок в status codes | Бизнес-логика, прямой SQL |
| `service` | Бизнес-правила, транзакции, RBAC-вызовы, публикация событий | Знание `http.ResponseWriter` |
| `repository` | SQL-запросы (sqlc) | Бизнес-правила |
| `domain` | Типы, enum, доменные ошибки | Зависимости от инфраструктуры |

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
│   ├── events/
│   └── cache/
├── db/
│   ├── migrations/
│   └── queries/
├── api/openapi.yaml
├── docs/
│   └── TZ.md
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```

### 3.4. Транзакционные границы

Операции, требующие одной транзакции PostgreSQL:

- Создание issue: INSERT issue + INSERT activity_event.
- Перемещение issue (move): UPDATE issue + INSERT activity_event.
- Добавление комментария: INSERT comment + INSERT activity_event.
- Принятие инвайта: INSERT workspace_member + UPDATE invite.

После успешного commit — публикация события в Redis (вне транзакции; at-least-once доставка в SSE).

### 3.5. Real-time

- **Протокол v1:** Server-Sent Events (SSE).
- **Канал Redis:** `channel:project:{project_id}`.
- **In-memory hub:** подписчики SSE на инстансе; при масштабировании — все инстансы слушают Redis.
- **WebSocket:** отложено на v2 (presence, typing).

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

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| email | string | UNIQUE, NOT NULL |
| password_hash | string | NOT NULL |
| name | string | NOT NULL |
| created_at | timestamptz | NOT NULL |
| updated_at | timestamptz | NOT NULL |

#### Workspace

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| name | string | NOT NULL |
| slug | string | UNIQUE, NOT NULL |
| created_at | timestamptz | NOT NULL |

#### WorkspaceMember

| Поле | Тип | Ограничения |
|------|-----|-------------|
| workspace_id | UUID | FK, PK (composite) |
| user_id | UUID | FK, PK (composite) |
| role | enum | `owner`, `member`, `viewer` |
| joined_at | timestamptz | NOT NULL |

#### Project

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| workspace_id | UUID | FK, NOT NULL |
| name | string | NOT NULL |
| key | string | 2–5 символов, A-Z, UNIQUE в workspace |
| created_at | timestamptz | NOT NULL |

#### Issue

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| project_id | UUID | FK, NOT NULL |
| number | int | NOT NULL, UNIQUE (project_id, number) |
| title | string | NOT NULL, max 500 |
| description | text | nullable, markdown |
| status | enum | см. 4.3 |
| priority | enum | см. 4.3 |
| assignee_id | UUID | FK users, nullable |
| position | numeric | NOT NULL, для сортировки в колонке |
| created_by | UUID | FK users, NOT NULL |
| created_at | timestamptz | NOT NULL |
| updated_at | timestamptz | NOT NULL |

Отображаемый идентификатор: `{project.key}-{number}` (например `BE-42`).

#### Label

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| workspace_id | UUID | FK |
| name | string | UNIQUE (workspace_id, name) |
| color | string | hex `#RRGGBB` |

#### IssueLabel (связь M:N Issue ↔ Label)

Junction-таблица: одна задача может иметь несколько меток, одна метка — на многих задачах в рамках workspace (через issues проектов этого workspace).

| Поле | Тип | Ограничения |
|------|-----|-------------|
| issue_id | UUID | FK → `issues(id)` ON DELETE CASCADE, часть PK |
| label_id | UUID | FK → `labels(id)` ON DELETE CASCADE, часть PK |
| created_at | timestamptz | NOT NULL, кто/когда привязал (опционально `created_by` в v1.1) |

**Первичный ключ:** `(issue_id, label_id)` — дубликаты привязки невозможны.

**Правила:**

- Label и Issue должны относиться к одному workspace (issue → project → workspace_id = label.workspace_id); иначе HTTP 400 `LABEL_WORKSPACE_MISMATCH`.
- При удалении issue или label строка в `issue_labels` удаляется каскадом.
- Привязка/отвязка: `POST/DELETE /issues/{id}/labels/{labelId}`; в activity — `issue.label_added` / `issue.label_removed`.
- В списке issues label подгружается JOIN или отдельным batch-запросом (избегать N+1).

**Пример SQL (создание таблицы):**

```sql
CREATE TABLE issue_labels (
    issue_id   UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id   UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (issue_id, label_id)
);
CREATE INDEX issue_labels_label_id_idx ON issue_labels (label_id);
```

#### Comment

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| issue_id | UUID | FK |
| author_id | UUID | FK users |
| body | text | NOT NULL |
| created_at | timestamptz | NOT NULL |
| updated_at | timestamptz | nullable |

#### ActivityEvent

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| issue_id | UUID | FK |
| actor_id | UUID | FK users |
| type | string | см. 4.4 |
| payload | JSONB | NOT NULL |
| created_at | timestamptz | NOT NULL |

#### RefreshToken

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| user_id | UUID | FK |
| token_hash | string | NOT NULL |
| expires_at | timestamptz | NOT NULL |
| revoked_at | timestamptz | nullable |
| created_at | timestamptz | NOT NULL |

#### Invite

| Поле | Тип | Ограничения |
|------|-----|-------------|
| id | UUID | PK |
| workspace_id | UUID | FK |
| email | string | NOT NULL |
| role | enum | `member`, `viewer` |
| token | string | UNIQUE, NOT NULL |
| expires_at | timestamptz | NOT NULL |
| accepted_at | timestamptz | nullable |
| created_by | UUID | FK users |

### 4.3. Перечисления

**Issue.status:** `backlog`, `todo`, `in_progress`, `review`, `done`

**Issue.priority:** `none`, `low`, `medium`, `high`, `urgent`

**WorkspaceMember.role:** `owner`, `member`, `viewer` (иерархия: owner > member > viewer)

### 4.4. Типы ActivityEvent

| type | payload (пример) |
|------|------------------|
| `issue.created` | `{ "issue_id", "title" }` |
| `issue.updated` | `{ "fields": ["title", "priority"] }` |
| `issue.moved` | `{ "old_status", "new_status", "position" }` |
| `issue.assigned` | `{ "assignee_id" }` |
| `comment.added` | `{ "comment_id" }` |
| `issue.label_added` | `{ "label_id", "name" }` |
| `issue.label_removed` | `{ "label_id", "name" }` |

### 4.5. Индексы (минимум)

- `issues (project_id, status)`
- `issues (project_id, number)` — UNIQUE
- `issues (assignee_id)` WHERE assignee_id IS NOT NULL
- `activity_events (issue_id, created_at DESC)`
- `comments (issue_id, created_at)`
- `workspace_members (user_id)`
- `projects (workspace_id)`
- `issue_labels (label_id)` — фильтр issues по label

---

## 5. RBAC

### 5.0. Что такое RBAC (простыми словами)

**RBAC** (Role-Based Access Control) — доступ к действиям определяется **ролью** пользователя, а не отдельным списком «можно/нельзя» на каждый endpoint.

В этом проекте:

1. Пользователь входит в **workspace** с ролью: `owner`, `member` или `viewer`.
2. Перед созданием issue, удалением project и т.д. сервис спрашивает: «роль пользователя в этом workspace ≥ требуемой?»
3. Если нет — **403 Forbidden**, данные не отдаём и не меняем.

Пример: `viewer` может `GET /issues`, но не может `POST /issues`. Это RBAC — не путать с **аутентификацией** (логин/JWT, «кто ты»); RBAC отвечает на вопрос «что тебе разрешено делать».

### 5.1. Область проверки

Права проверяются на уровне **workspace** (доступ к project/issue через membership в workspace проекта).

### 5.2. Матрица прав

| Действие | owner | member | viewer |
|----------|:-----:|:------:|:------:|
| Просмотр projects, issues, comments, activity | ✓ | ✓ | ✓ |
| Создание/редактирование issue | ✓ | ✓ | ✗ |
| Удаление issue | ✓ | ✓ | ✗ |
| Создание/редактирование project | ✓ | ✓ | ✗ |
| Удаление project | ✓ | ✗ | ✗ |
| Управление labels | ✓ | ✓ | ✗ |
| Создание invite | ✓ | ✓ | ✗ |
| Изменение ролей members | ✓ | ✗ | ✗ |
| Удаление workspace | ✓ | ✗ | ✗ |

### 5.3. Реализация

- Пакет `internal/authz`: `Can(ctx, userID, workspaceID, action Action) bool`.
- HTTP middleware: `RequireAuth`, `RequireWorkspaceRole(workspaceID, minRole)`.
- Ошибка при отсутствии прав: HTTP 403, код `FORBIDDEN`.

---

## 6. Аутентификация и сессии

### 6.1. Регистрация и вход

- `POST /api/v1/auth/register` — email, password, name.
- `POST /api/v1/auth/login` — email, password → access + refresh.
- `POST /api/v1/auth/refresh` — refresh token → новая пара токенов.
- `POST /api/v1/auth/logout` — отзыв refresh (запись `revoked_at`).

### 6.2. JWT Access Token

- Время жизни: **15 минут**.
- Claims: `sub` (user_id), `exp`, `iat`, опционально `jti`.
- Передача: заголовок `Authorization: Bearer <token>`.

### 6.3. Refresh Token

- Время жизни: **7 дней**.
- Хранение: только хеш в таблице `refresh_tokens`.
- Ротация при refresh (старый токен revoke, выдача нового).

### 6.4. Rate limiting (Redis)

| Endpoint | Лимит |
|----------|-------|
| `POST /auth/login` | 10 запросов / мин / IP |
| `POST /auth/register` | 5 запросов / мин / IP |

Ключ: `ratelimit:login:{ip}`, `ratelimit:register:{ip}`.

---

## 7. API (REST v1)

Базовый префикс: `/api/v1`. Формат: JSON. Ошибки: `{ "error": { "code": "...", "message": "..." } }`.

### 7.1. Auth

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| POST | `/auth/register` | — | Регистрация |
| POST | `/auth/login` | — | Вход |
| POST | `/auth/refresh` | — | Обновление токенов |
| POST | `/auth/logout` | refresh body | Выход |

### 7.2. Workspaces

| Method | Path | Min role | Описание |
|--------|------|----------|----------|
| GET | `/workspaces` | auth | Список workspace пользователя |
| POST | `/workspaces` | auth | Создать workspace (создатель = owner) |
| GET | `/workspaces/{id}` | viewer+ | Детали |
| PATCH | `/workspaces/{id}` | owner | Обновить name/slug |
| DELETE | `/workspaces/{id}` | owner | Удалить workspace |
| GET | `/workspaces/{id}/members` | viewer+ | Список участников |
| POST | `/workspaces/{id}/members` | owner | Добавить member (по user_id) |
| PATCH | `/workspaces/{id}/members/{userId}` | owner | Сменить роль |
| DELETE | `/workspaces/{id}/members/{userId}` | owner | Удалить участника |
| POST | `/workspaces/{id}/invites` | member+ | Создать invite |
| POST | `/invites/{token}/accept` | auth | Принять приглашение |

### 7.3. Projects

| Method | Path | Min role | Описание |
|--------|------|----------|----------|
| GET | `/workspaces/{wsId}/projects` | viewer+ | Список проектов |
| POST | `/workspaces/{wsId}/projects` | member+ | Создать проект |
| GET | `/projects/{id}` | viewer+ | Детали |
| PATCH | `/projects/{id}` | member+ | Обновить |
| DELETE | `/projects/{id}` | owner | Удалить |

### 7.4. Issues

| Method | Path | Min role | Описание |
|--------|------|----------|----------|
| GET | `/projects/{id}/issues` | viewer+ | Список (фильтры: status, assignee, label, q) |
| POST | `/projects/{id}/issues` | member+ | Создать issue |
| GET | `/issues/{id}` | viewer+ | Детали |
| PATCH | `/issues/{id}` | member+ | Обновить поля |
| DELETE | `/issues/{id}` | member+ | Удалить |
| PATCH | `/issues/{id}/move` | member+ | Смена status и position |

**Query для GET issues:** `?status=todo&assignee={userId}&label={labelId}&q=search`

### 7.5. Comments и Activity

| Method | Path | Min role | Описание |
|--------|------|----------|----------|
| GET | `/issues/{id}/comments` | viewer+ | Список комментариев |
| POST | `/issues/{id}/comments` | member+ | Добавить комментарий |
| GET | `/issues/{id}/activity` | viewer+ | Лента activity_events |

### 7.6. Labels

| Method | Path | Min role | Описание |
|--------|------|----------|----------|
| GET | `/workspaces/{wsId}/labels` | viewer+ | Список labels |
| POST | `/workspaces/{wsId}/labels` | member+ | Создать label |
| POST | `/issues/{id}/labels/{labelId}` | member+ | Привязать label |
| DELETE | `/issues/{id}/labels/{labelId}` | member+ | Отвязать label |

### 7.7. Real-time (SSE)

| Method | Path | Min role | Описание |
|--------|------|----------|----------|
| GET | `/projects/{id}/events` | viewer+ | SSE-поток событий проекта |

**Формат события SSE:**

```
event: issue.updated
data: {"type":"issue.updated","project_id":"...","payload":{...},"timestamp":"..."}

```

Типы событий в потоке: `issue.created`, `issue.updated`, `issue.moved`, `issue.deleted`, `comment.added`.

### 7.8. Служебные endpoints

| Method | Path | Описание |
|--------|------|----------|
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness (Postgres + Redis ping) |
| GET | `/metrics` | Prometheus metrics |

---

## 8. Бизнес-правила

### 8.1. Нумерация issue

- При создании issue в транзакции: `number = COALESCE(MAX(number), 0) + 1` для `project_id`.
- Уникальность: `(project_id, number)`.

### 8.2. Позиция на kanban (move)

- Клиент передаёт `status` и `position` (numeric).
- Стратегия: fractional indexing (вставка между соседними position).
- При конфликте/переполнении precision — фоновый rebalance (v1.1); в v1 допустим простой numeric.

### 8.3. Удаление

- Удаление project каскадно удаляет issues (или soft-delete — зафиксировать в миграции; **v1: hard delete** с ON DELETE CASCADE).
- Удаление workspace каскадно удаляет projects, members, labels.

### 8.4. Invites

- Срок действия invite: **7 дней**.
- Повторное принятие того же token: 409 `INVITE_ALREADY_ACCEPTED`.
- Email invite может не совпадать с email пользователя (опционально проверять в v1.1).

### 8.5. Валидация

- Email: формат RFC, unique.
- Password: минимум 8 символов.
- Project key: `^[A-Z]{2,5}$`, unique в workspace.
- Title issue: 1–500 символов.

---

## 9. Redis

| Ключ / канал | TTL | Назначение |
|--------------|-----|------------|
| `ratelimit:login:{ip}` | 1 min | Лимит попыток входа |
| `ratelimit:register:{ip}` | 1 min | Лимит регистраций |
| `board:{project_id}` | 60s | Кэш JSON списка issues (опционально) |
| `channel:project:{id}` | — | Pub/Sub для SSE |

Инвалидация кэша `board:{project_id}` при любом изменении issue в проекте.

---

## 10. Observability

### 10.1. Логирование

- Формат: JSON (zap/slog).
- Обязательные поля запроса: `request_id`, `method`, `path`, `status`, `duration_ms`, `user_id` (если есть).

### 10.2. Метрики Prometheus

| Метрика | Тип | Labels |
|---------|-----|--------|
| `http_requests_total` | counter | method, path, status |
| `http_request_duration_seconds` | histogram | method, path |
| `db_query_duration_seconds` | histogram | operation |
| `events_published_total` | counter | type |
| `sse_active_connections` | gauge | — |

Ограничить cardinality labels (не использовать `user_id` в labels).

### 10.3. Health checks

- `/healthz` — процесс жив.
- `/readyz` — успешный ping PostgreSQL и Redis.

### 10.4. Graceful shutdown

- По SIGINT/SIGTERM: прекратить приём новых запросов, дождаться in-flight (timeout 30s), закрыть SSE-подписчиков, закрыть пулы БД и Redis.

---

## 11. Нефункциональные требования

| Требование | Значение v1 |
|------------|-------------|
| Время ответа API (p95) | < 200 ms без SSE на локальной машине |
| Concurrent users (demo) | до 50 |
| Хранение данных | PostgreSQL persistent volume |
| Безопасность | HTTPS в production, secrets через env, не коммитить `.env` |
| CORS | Настраиваемый `CORS_ORIGINS` для фронта |

---

## 12. Тестирование

### 12.1. Unit-тесты

- `authz`: матрица прав для всех ролей.
- `service`: мок repository — create issue, move, forbidden для viewer.
- Валидация DTO.

### 12.2. Интеграционные тесты (testcontainers)

- Регистрация → login → create workspace → project → issue.
- Move issue меняет status и создаёт activity.
- Viewer не может создать issue (403).
- Rate limit на login (опционально).

### 12.3. CI (GitHub Actions)

1. `golangci-lint run`
2. `go test ./...`
3. `go build -o bin/api ./cmd/api`

---

## 13. Этапы разработки

### Неделя 1 — Фундамент

- [ ] Репозиторий, docker-compose, Makefile
- [ ] Миграции: users, workspaces, workspace_members, projects
- [ ] Auth: register, login, refresh, logout, JWT middleware
- [ ] CRUD workspace/project, RBAC middleware
- [ ] Интеграционные тесты auth + project

### Неделя 2 — Ядро

- [ ] Issues: CRUD, filters, move, activity_events
- [ ] Comments, labels
- [ ] SSE + Redis pub/sub
- [ ] OpenAPI черновик

### Неделя 3 — Production polish

- [ ] Invites
- [ ] Prometheus, structured logs, request_id
- [ ] golangci-lint + GitHub Actions
- [ ] README, ADR, деплой demo
- [ ] (Опционально) минимальный фронт kanban

---

## 14. Критерии приёмки MVP

1. Пользователь может зарегистрироваться, создать workspace и project.
2. Member создаёт issue, перемещает по статусам на доске; viewer только читает.
3. Комментарии и activity отображают историю изменений.
4. Два клиента с открытым SSE видят обновление доски без перезагрузки.
5. `/readyz` и `/metrics` работают; тесты проходят в CI.
6. README описывает запуск за ≤ 5 минут и архитектуру.

---

## 15. ADR (зафиксированные решения)

| ID | Решение | Альтернатива | Причина |
|----|---------|--------------|---------|
| ADR-001 | sqlc + pgx | GORM | Прозрачный SQL, типобезопасность |
| ADR-002 | SSE + Redis pub/sub | WebSocket | Проще MVP; WS в v2 |
| ADR-003 | RBAC на workspace | Per-project roles | Меньше сложности |
| ADR-004 | UUID v4/v7 | bigint | Удобство в API и распределённости |
| ADR-005 | Hard delete v1 | soft delete | Простота; soft в v2 |
| ADR-006 | Issue number в Tx | Глобальная sequence | Читаемый ключ per project |

---

## 16. Риски и ограничения

| Риск | Митигация |
|------|-----------|
| Переусложнение сроков | Строгий MVP по чеклисту §13 |
| Гонка при move position | Транзакции; rebalance в v1.1 |
| SSE при нескольких инстансах | Redis pub/sub обязателен при deploy >1 replica |
| Утечка goroutine на SSE | Context cancel при disconnect клиента |

---

## 17. Глоссарий

| Термин | Определение |
|--------|-------------|
| Workspace | Команда, контейнер для projects и members |
| Project | Набор issues с префиксом key (`BE-42`) |
| Issue | Задача на kanban-доске |
| Activity | Запись в audit-ленте issue |
| IssueLabel | Связь many-to-many: задача ↔ метка (таблица `issue_labels`) |
| Label | Метка на уровне workspace (например `bug`, `frontend`) |
| RBAC | Role-Based Access Control — права по роли в workspace |
| SSE | Server-Sent Events, однонаправленный поток сервер → клиент |

---

## 18. С чего начать разработку (порядок шагов)

Не пытайся сделать всё сразу. Один вертикальный срез за раз.

### День 1 — «проект дышит»

1. `go mod init` + `cmd/api/main.go` — сервер на chi, `GET /healthz` → `{"status":"ok"}`.
2. `docker-compose.yml` — postgres + redis (api пока с хоста или тоже в compose).
3. `Makefile`: `up`, `down`, `run`, `migrate`.
4. Первая миграция goose: только `users` (id, email, password_hash, name, timestamps).
5. Проверка: `make up && make migrate && make run` → curl healthz.

**Критерий:** контейнеры поднялись, миграция применилась, API отвечает.

### День 2–3 — auth

6. Миграция: `refresh_tokens`.
7. `POST /auth/register`, `POST /auth/login` — bcrypt, JWT access, refresh в БД.
8. Middleware `RequireAuth` — из заголовка `Authorization` достаёшь `user_id` в context.
9. 1–2 интеграционных теста: register → login → защищённый endpoint.

**Критерий:** без токена 401, с токеном — видишь свой user_id.

### День 4–5 — workspace + RBAC

10. Миграции: `workspaces`, `workspace_members`, `projects`.
11. `POST /workspaces` (создатель = owner), `GET /workspaces`, `POST .../projects`.
12. Пакет `authz` + middleware: для routes с `{workspaceId}` проверка роли.
13. Тест: viewer не может создать project.

**Критерий:** два пользователя, один workspace, роли работают.

### День 6–8 — issues (ядро)

14. Миграции: `issues`, `labels`, `issue_labels`, `comments`, `activity_events`.
15. CRUD issues + `PATCH .../move` + activity при каждом изменении.
16. Labels: CRUD на workspace + attach/detach `issue_labels`.
17. `GET /projects/{id}/issues` с фильтром `?label=`.

**Критерий:** в Insomnia/Bruno полный сценарий: login → workspace → project → issue → label.

### День 9–10 — real-time и polish

18. Redis pub/sub + `GET /projects/{id}/events` (SSE).
19. `/metrics`, `/readyz`, request_id в логах.
20. README: как запустить за 3 команды.

### Если теряешься прямо сейчас

Открой только **шаг 1–5 (День 1)**. Не читай SSE, не думай про фронт. Цель одного вечера: **Postgres в Docker + пустой API + одна таблица users**.

Структура папок на старте (достаточно):

```
cmd/api/main.go
internal/config/config.go
internal/http/router.go
db/migrations/001_users.sql
docker-compose.yml
Makefile
```

Остальные пакеты (`service`, `repository`) добавляй, когда появляется **вторая** ручка с бизнес-логикой — не раньше.

---

## 19. Ссылки на артефакты (заполнить по мере разработки)

| Артефакт | Путь |
|----------|------|
| OpenAPI | `api/openapi.yaml` |
| Миграции | `db/migrations/` |
| README | `README.md` |
| Demo URL | TBD |

---

*Документ является источником истины для scope v1. Изменения — через версионирование TZ (1.0.1 — IssueLabel, RBAC §5.0, §18 старт разработки).*
