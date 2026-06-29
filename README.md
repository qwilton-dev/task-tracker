# Task Tracker

[Русская версия](README.ru.md)

Issue tracker with kanban board, RBAC, real-time updates, and a built-in web UI. Built with Go, PostgreSQL, and Server-Sent Events.

## Features

- **Auth & Sessions** — register, login, JWT access tokens + refresh token rotation, logout
- **Workspaces** — create, update, delete workspaces; invite members with roles
- **RBAC** — three roles: `owner`, `member`, `viewer` with granular permissions
- **Projects** — CRUD, auto-generated key prefix (e.g. `BE-42`)
- **Issues** — CRUD + kanban move (status + position), filters by status/assignee/search
- **Labels** — create, attach/detach to issues
- **Comments** — add comments to issues
- **Activity Feed** — full audit trail of changes per issue
- **Real-time** — SSE stream per project, instant board updates
- **Health Checks** — `/healthz`, `/readyz`
- **Frontend** — vanilla JS SPA with kanban drag-and-drop, auth, SSE
- **CI/CD** — GitHub Actions: lint, test, build
- **Docker** — single `docker compose up` to run everything

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP Router | [chi/v5](https://github.com/go-chi/chi) |
| Database | PostgreSQL 16 |
| Driver | pgx/v5 |
| Migrations | [goose](https://github.com/pressly/goose) |
| Auth | JWT (access) + refresh tokens |
| Passwords | bcrypt |
| Real-time | Server-Sent Events (in-memory Hub) |
| Containerization | Docker, Docker Compose |
| CI | GitHub Actions |
| Frontend | Vanilla JS |

## Architecture

```
HTTP Handler → Service (use cases) → Repository → PostgreSQL
                      ↓
              Hub (in-memory) → SSE → clients
```

### Layers

| Layer | Responsibility |
|---|---|
| `handler` | HTTP parsing, DTO validation, error mapping |
| `service` | Business rules, transactions, RBAC, event publishing |
| `repository` | SQL queries via pgx |
| `domain` | Types, enums, domain errors |

### Project Structure

```
task-tracker/
├── cmd/api/main.go              # entrypoint
├── internal/
│   ├── auth/                    # JWT service
│   ├── authz/                   # RBAC roles & permissions
│   ├── cache/                   # caching layer
│   ├── config/                  # env config loader
│   ├── domain/                  # types, enums, validation
│   ├── events/                  # SSE hub
│   ├── http/
│   │   ├── router.go            # chi routes
│   │   ├── middleware/          # auth, CORS, RBAC middleware
│   │   └── handler/             # HTTP handlers
│   ├── repository/
│   │   └── postgres/            # pgx repositories
│   └── service/                 # business logic
├── db/migrations/               # 11 goose migrations
├── web/                         # frontend SPA
│   ├── index.html
│   ├── app.js
│   └── style.css
├── api/openapi.yaml             # API spec
├── docs/TZ.md                   # technical specification
├── docker-compose.yaml
├── Dockerfile
└── Makefile
```

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.26+ (for local dev)
- [goose](https://github.com/pressly/goose) (for migrations)

### Quick Start (Docker)

```bash
cp .env.example .env
# edit .env if needed
make up-build
```

The app will be available at `http://localhost:8080`.

### Local Development

```bash
# start postgres only
docker compose up -d postgres

# run migrations
make migrate-up

# start the API server
make run
```

### Make Commands

| Command | Description |
|---|---|
| `make up` | Start all services |
| `make up-build` | Build and start all services |
| `make down` | Stop all services |
| `make run` | Run API server locally |
| `make migrate-up` | Apply all migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-status` | Show migration status |
| `make logs` | Tail container logs |

## API

Base path: `/api/v1`. Format: JSON.

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | — | Register |
| POST | `/auth/login` | — | Login |
| POST | `/auth/refresh` | — | Refresh tokens |
| POST | `/auth/logout` | — | Logout (revoke refresh) |
| GET | `/me` | JWT | Current user |

### Workspaces

| Method | Path | Role | Description |
|---|---|---|---|
| GET | `/workspaces` | auth | List workspaces |
| POST | `/workspaces` | auth | Create workspace |
| GET | `/workspaces/{id}` | auth | Workspace details |
| PATCH | `/workspaces/{id}` | owner | Update workspace |
| DELETE | `/workspaces/{id}` | owner | Delete workspace |
| GET | `/workspaces/{id}/members` | auth | List members |
| POST | `/workspaces/{id}/members` | owner | Add member |
| PATCH | `/workspaces/{id}/members/{userId}` | owner | Change role |
| DELETE | `/workspaces/{id}/members/{userId}` | owner | Remove member |
| POST | `/workspaces/{id}/invites` | member+ | Create invite |
| GET | `/workspaces/{id}/invites` | auth | List invites |
| POST | `/invites/{token}/accept` | auth | Accept invite |

### Projects

| Method | Path | Role | Description |
|---|---|---|---|
| GET | `/workspaces/{wsId}/projects` | auth | List projects |
| POST | `/workspaces/{wsId}/projects` | member+ | Create project |
| GET | `/projects/{id}` | auth | Project details |
| PATCH | `/projects/{id}` | member+ | Update project |
| DELETE | `/projects/{id}` | owner | Delete project |

### Issues

| Method | Path | Role | Description |
|---|---|---|---|
| GET | `/projects/{id}/issues` | auth | List (filter: status, assignee, q) |
| POST | `/projects/{id}/issues` | member+ | Create issue |
| GET | `/issues/{id}` | auth | Issue details |
| PATCH | `/issues/{id}` | member+ | Update fields |
| DELETE | `/issues/{id}` | member+ | Delete issue |
| PATCH | `/issues/{id}/move` | member+ | Move (status + position) |

### Comments & Activity

| Method | Path | Role | Description |
|---|---|---|---|
| GET | `/issues/{id}/comments` | auth | List comments |
| POST | `/issues/{id}/comments` | member+ | Add comment |
| GET | `/issues/{id}/activity` | auth | Activity feed |

### Labels

| Method | Path | Role | Description |
|---|---|---|---|
| GET | `/workspaces/{wsId}/labels` | auth | List labels |
| POST | `/workspaces/{wsId}/labels` | member+ | Create label |
| GET | `/issues/{id}/labels` | auth | Labels on issue |
| POST | `/issues/{id}/labels/{labelId}` | member+ | Attach label |
| DELETE | `/issues/{id}/labels/{labelId}` | member+ | Detach label |

### Real-time (SSE)

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/projects/{id}/events` | JWT | SSE event stream |

Events: `issue.created`, `issue.updated`, `issue.moved`, `issue.deleted`, `comment.added`

### Health Checks

| Method | Path | Description |
|---|---|---|
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |

## RBAC

| Action | owner | member | viewer |
|---|:---:|:---:|:---:|
| View projects, issues | ✅ | ✅ | ✅ |
| Create/edit/delete issues | ✅ | ✅ | ❌ |
| Create/edit projects | ✅ | ✅ | ❌ |
| Delete projects | ✅ | ❌ | ❌ |
| Manage labels | ✅ | ✅ | ❌ |
| Create invites | ✅ | ✅ | ❌ |
| Change member roles | ✅ | ❌ | ❌ |
| Delete workspace | ✅ | ❌ | ❌ |

## Testing

```bash
# run all tests
go test ./...

# with race detection and coverage
go test -race -coverprofile=coverage.out ./...
```

24 test files covering domain validation, authz matrix, service logic, and HTTP handlers.

## License

MIT
