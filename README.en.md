# Campus Recruitment & Resume Platform

A three-sided campus recruitment system for **students, companies, and administrators**, covering student profiles and resumes, job posting, candidate search and bookmarking, resume applications, real-time chat, analytics dashboards, and bulk back-office management.

The repository uses a **single Go backend plus two front-end applications** (student/company portal and admin panel). The backend lives in `back/resume-go` and relies on MySQL + Redis + RabbitMQ, exposing REST, WebSocket chat, and an internal gRPC service.

> This document is derived **strictly from the source code and SQL files**. It covers the tech stack, architecture, every endpoint, authentication, the chat engine, database design, caching, messaging, gRPC, deployment, and known issues. After reading it you should be able to run, debug, and extend the project independently.
>
> For a file-by-file reading order, see [important.md](important.md) (must-read / optional / skip list and a minimal reading path).

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Feature Matrix](#2-feature-matrix)
3. [Technology Stack](#3-technology-stack)
4. [System Architecture](#4-system-architecture)
5. [Directory Layout](#5-directory-layout)
6. [Quick Start](#6-quick-start)
7. [Configuration](#7-configuration)
8. [Authentication & Authorization](#8-authentication--authorization)
9. [Response Envelope & Error Codes](#9-response-envelope--error-codes)
10. [Database Design](#10-database-design)
11. [API Reference (Complete)](#11-api-reference-complete)
12. [Chat System Deep Dive](#12-chat-system-deep-dive)
13. [Resume & File Storage](#13-resume--file-storage)
14. [Search, Pagination & Query Design](#14-search-pagination--query-design)
15. [Bookmark (Favorite) Module](#15-bookmark-favorite-module)
16. [Resume Application Module](#16-resume-application-module)
17. [Excel Bulk Import](#17-excel-bulk-import)
18. [Redis Cache Design](#18-redis-cache-design)
19. [Internal gRPC Service](#19-internal-grpc-service)
20. [RabbitMQ Offline Notifications](#20-rabbitmq-offline-notifications)
21. [Front-End Implementation](#21-front-end-implementation)
22. [Deployment](#22-deployment)
23. [Known Issues & Caveats](#23-known-issues--caveats)
24. [Development Guide](#24-development-guide)
25. [Appendix: Endpoint Cheat Sheet](#25-appendix-endpoint-cheat-sheet)

---

## 1. Project Overview

| Dimension | Description |
|---|---|
| Domain | Campus recruiting: students maintain profiles/resumes and apply; companies search students, post jobs, bookmark and chat; admins maintain master data |
| Roles | `student`, `company`, `admin` — stored in `user.user_type` |
| Student identity | Login name = student number (`user.username`); profile PK = `student.id`; student-number column = `student.stu_id` |
| Company identity | Login name = company name (`user.username`); profile PK = `company.id` |
| Chat identity | A person is always identified by the pair `(user_id, user_type)` |
| Backend | `back/resume-go` (Go 1.27 + Gin) on `8081`; gRPC on `9091` |
| Front ends | `front/yxl/recruit-project` (student + company, Vite 5173) and `front/ljk/adminPanel` (admin, Vite 5174 / `/admin/`) |
| Data stores | MySQL 8 (`resume` schema) + Redis 7 + RabbitMQ 3.13 |

### 1.1 Backend implementation

- A single backend (`back/resume-go`) layered as `api → service → data/cache/mq/rpc`, with raw SQL and no ORM.
- Backward compatible with the legacy API: requests accept both camelCase and snake_case, and responses emit both, so existing front ends need no changes.
- [start-all.bat](file:///d:/resume/start-all.bat) starts MySQL, the Go backend, and both front ends.

```
                  ┌──────────────────────────────┐
                  │  Front end A: student/company │
                  │  Vue3 + Element Plus + Pinia  │
                  └──────────────┬───────────────┘
                                 │ HTTP (REST) + WebSocket
                  ┌──────────────▼───────────────┐
                  │  Front end B: admin /admin/   │
                  │  Vue3 + Element Plus          │
                  └──────────────┬───────────────┘
                                 │
        ┌────────────────────────▼────────────────────────┐
        │  API layer :8081 (Go 1.27 / Gin)                 │
        │  includes /chat/ws and gRPC :9091                │
        └───┬──────────────┬──────────────┬───────────────┘
            │              │              │
       ┌────▼────┐   ┌─────▼─────┐  ┌─────▼──────┐
       │ MySQL 8 │   │  Redis    │  │ RabbitMQ   │
       │ source  │   │ cache/    │  │ offline    │
       │ of truth│   │ presence  │  │ notify     │
       └─────────┘   └───────────┘  └────────────┘
```

---

## 2. Feature Matrix

| Module | Student | Company | Admin | Backend |
|---|---|---|---|---|
| Login / refresh / validate token | ✅ | ✅ | ✅ | Go |
| View/update own account, rename | ✅ | ✅ | — | Go |
| Full student profile (9 sections + 6 child tables) | ✅ | read-only search | read-only search | Go |
| Resume upload / latest resume | ✅ | — | — | Go |
| Resume template editor + PDF export | ✅ (client-side) | — | — | Front end |
| Create / edit / activate / delete jobs | — | ✅ | — | Go |
| Job list / detail | ✅ | ✅ | — | Go |
| Multi-criteria student search (with details) | — | ✅ | ✅ | Go |
| Bookmark / unbookmark / check bookmark | — | ✅ | — | Go |
| Apply / withdraw / status workflow | ✅ | ✅ | — | Go |
| Real-time chat (text/image/resume) | ✅ | ✅ | — (blocked) | Go |
| Offline notifications / unread counts | ✅ | ✅ | — | Go |
| Analytics dashboard | public | public | — | Go |
| Excel import of students/companies | — | — | ✅ | Go |
| Bulk delete/update | — | — | ✅ | Go |
| Language/certificate/skill dictionaries | — | — | ✅ | Go |
| Reset student password | — | — | ✅ | Go |

---

## 3. Technology Stack

### 3.1 Front end A — student/company portal (`front/yxl/recruit-project`)

| Area | Technology | Version | Purpose |
|---|---|---|---|
| Framework | Vue | ^3.5.26 | Composition API + `<script setup>` |
| Build | Vite | ^7.3.1 | Dev server, HMR, proxy |
| Routing | Vue Router | ^4.6.4 | `createWebHistory`, lazy-loaded routes |
| State | Pinia | ^3.0.4 | User / filters / resume draft |
| Persistence | pinia-plugin-persistedstate | ^4.7.1 | Token persisted to `localStorage` |
| UI | Element Plus | ^2.13.2 | Components + dark theme |
| Charts | ECharts / vue-echarts | ^6.0.0 / ^8.0.1 | Five dashboard charts |
| HTTP | axios | ^1.13.4 | Shared instance + interceptors |
| PDF export | html2pdf.js | ^0.14.0 | Resume template export |
| Styles | Sass | ^1.97.3 | Component styles |
| Quality | ESLint + Oxlint + Prettier | — | `npm run lint` |
| Auto import | unplugin-auto-import / unplugin-vue-components | — | On-demand Element Plus |

### 3.2 Front end B — admin panel (`front/ljk/adminPanel`)

| Area | Technology | Version | Purpose |
|---|---|---|---|
| Framework | Vue | ^3.5.22 | `<script setup>` |
| Build | Vite | ^7.1.11 | `base: '/admin/'` |
| Routing | Vue Router | ^4.6.3 | `createWebHistory('/admin/')` |
| State | Pinia | ^3.0.3 | Sample store only (unused) |
| UI | Element Plus | ^2.11.8 | Tables, pagination, upload, drawer |
| HTTP | axios | ^1.8.4 | **Called directly inside the page, no wrapper** |
| Plugin | @vitejs/plugin-vue-jsx | ^5.1.1 | JSX support |

### 3.3 Backend — Go (`back/resume-go`)

| Area | Technology | Version | Purpose |
|---|---|---|---|
| Runtime | Go | 1.27 | — |
| HTTP | Gin | v1.12.0 | Groups, middleware, static files |
| Database | database/sql + go-sql-driver/mysql | v1.10.1 | Raw SQL, no ORM |
| WebSocket | gorilla/websocket | v1.5.3 | Chat long connections |
| JWT | stdlib `crypto/hmac` + `sha256` | — | **Hand-rolled HS256, no library** |
| Password | golang.org/x/crypto/bcrypt | v0.58.0 | bcrypt |
| Cache | redis/go-redis/v9 | v9.23.0 | Cache / favorites / presence / idempotency |
| Concurrency | golang.org/x/sync/singleflight | v0.24.0 | Cache stampede protection |
| MQ | rabbitmq/amqp091-go | v1.16.0 | Offline notifications |
| gRPC | google.golang.org/grpc | v1.84.0 | Internal service |
| Excel | xuri/excelize/v2 | v2.11.0 | `.xlsx` import |
| JSON | bytedance/sonic (Gin indirect) | — | Faster serialization |

### 3.4 Infrastructure

| Component | Version | Purpose |
|---|---|---|
| MySQL | 8.0 | Single source of truth; uses `ROW_NUMBER()` (8.0+) |
| Redis | 7 | Cache, favorite sets, presence, MQ idempotency |
| RabbitMQ | 3.13 (management) | Chat offline notifications |
| Nginx Ingress | — | K8s entry, `proxy-body-size: 12m` |

---

## 4. System Architecture

### 4.1 Layering

**Go backend**:

```
api (HTTP binding / response envelope / middleware)
  └── service (business rules, transactions, chat throttling)
        ├── data (SQL, transactions)
        ├── cache (Redis with in-memory fallback)
        ├── mq (RabbitMQ with no-op fallback)
        ├── rpc (gRPC client)
        └── chat.Hub (in-process connection registry)
```

- The `chat` package calls back into business logic via injected hooks (`Hub.OnMessage` / `OnRegister` / `OnUnregister`), avoiding a `chat → service → chat` import cycle.
- Dependency strength: MySQL is **hard** (exit on failure); Redis / RabbitMQ / gRPC are **soft** (auto-degrade).

### 4.2 Key data flows

| Scenario | Flow |
|---|---|
| Login | Front end → `/login` → `user` lookup → bcrypt check → sign access (2h) + refresh (7d) → Pinia/localStorage |
| Profile update | `/stu/completeInfo` → partial update of main row + delete-then-insert on 6 child tables → re-aggregate and return |
| Candidate search | `/com/detailSearch` → dynamic WHERE + COUNT/LIMIT → batch `IN` queries on child tables → assemble DTOs |
| Chat | WebSocket → Hub → service validation → `INSERT chat_message` → re-select → push to both sides → MQ offline notify |
| Application | `/stu/applications` → validate job/resume → `INSERT` (unique key blocks duplicates) → company status workflow |
| Dashboard | `/stu/statistics` (public) → five `GROUP BY` aggregates |

---

## 5. Directory Layout

```text
d:\resume
├─ back
│  ├─ resume.sql                 # base schema (user/student/company/...)
│  └─ resume-go                  # Go backend (Gin + WebSocket + gRPC + MQ)
│     ├─ main.go
│     ├─ internal
│     │  ├─ api                  # routes + handlers + middleware
│     │  ├─ auth                 # hand-rolled JWT + bcrypt
│     │  ├─ chat                 # Hub / Client (read/write pumps, heartbeat)
│     │  ├─ service              # business rules
│     │  ├─ data                 # SQL / transactions
│     │  ├─ cache                # Redis + in-memory fallback
│     │  ├─ mq                   # RabbitMQ publisher/consumer
│     │  ├─ rpc                  # gRPC server/client + generated pb
│     │  ├─ model                # DTO / entity
│     │  └─ config               # environment variables
│     ├─ proto/recruit/v1/recruit.proto
│     ├─ schema.sql              # extended schema (chat_message, ...)
│     ├─ seed_accounts.sql       # x1/q1/g1 demo accounts
│     ├─ seed_statistics.sql     # dashboard demo data
│     ├─ templates/*.xlsx        # Excel import templates
│     ├─ Dockerfile / docker-compose.yml
│     ├─ deploy/k8s/*.yaml
│     └─ uploads/{resume,chat}
├─ front
│  ├─ yxl/recruit-project        # student + company portal
│  │  └─ src/{api,components,router,stores,utils,views}
│  └─ ljk/adminPanel             # admin panel
│     └─ src/{components,router,views}
├─ start-all.bat                 # Windows one-click start (MySQL + Go + both UIs)
└─ CHAT_GUIDE.md                 # chat system tutorial
```

---

## 6. Quick Start

### 6.1 One-click start (Windows, recommended)

Double-click [start-all.bat](file:///d:/resume/start-all.bat). It performs:

1. **Database**: probe 3306, else 3307; if neither is listening and a local test data directory exists, launch `mysqld --datadir=... --port=3307`; then idempotently run `schema.sql`, `seed_accounts.sql`, `seed_statistics.sql`.
2. **Go backend**: `go run .` with `DB_HOST=127.0.0.1`, `DB_PORT=<detected>`; wait up to 90s for port 8081.
3. **Portal**: `npm install` on first run, then `npm run dev -- --host 127.0.0.1 --port 5173 --strictPort`, injecting `VITE_API_BASE_URL=http://127.0.0.1:8081` and `VITE_ADMIN_URL=http://127.0.0.1:5174`.
4. **Admin panel**: same, port 5174, injecting `VITE_LOGIN_URL=http://127.0.0.1:5173`.

| Entry | URL |
|---|---|
| Student/company portal | http://127.0.0.1:5173/ |
| Admin panel | http://127.0.0.1:5174/admin/ |
| Backend | http://localhost:8081/ |
| RabbitMQ console (if enabled) | http://localhost:15672/ |

Demo accounts (`seed_accounts.sql`):

| Role | Account | Password |
|---|---|---|
| Student | `x1` – `x10` | same as account |
| Company | `q1` – `q10` | same as account |
| Admin | `g1` | `g1` |

> To test real-time chat: log in as `x1` in a normal window and `q1` in a private window, then open **Messages** in both.

### 6.2 Run the Go backend manually

```bash
cd d:\resume\back\resume-go
# PowerShell example
$env:DB_HOST="127.0.0.1"; $env:DB_PORT="3306"; $env:DB_NAME="resume"
$env:DB_USER="root"; $env:DB_PASSWORD="<your password>"
$env:REDIS_ENABLED="false"; $env:RABBITMQ_ENABLED="false"; $env:GRPC_ENABLED="false"
go run .
```

With Redis/RabbitMQ/gRPC disabled, all core features still work; chat pushes only within this process and offline notifications are not generated.

### 6.3 Run the front ends

```bash
# Portal
cd d:\resume\front\yxl\recruit-project
npm install
npm run dev            # http://localhost:5173, /api-backend -> 127.0.0.1:8081

# Admin
cd d:\resume\front\ljk\adminPanel
npm install
npm run dev -- --port 5174
# open http://localhost:5174/admin/
```

Front-end environment variables (defaults apply when no `.env` exists):

| Variable | Purpose | Default |
|---|---|---|
| `VITE_API_BASE_URL` (portal) | Production backend URL | `http://182.92.210.198:8081` |
| `VITE_ADMIN_URL` (portal) | Redirect after admin login | `http://47.93.133.36` |
| `VITE_API_BASE_URL` (admin) | Backend URL | `http://81.70.199.174:8081` |
| `VITE_LOGIN_URL` (admin) | Login page when no token | `http://47.93.133.36` |

In development the portal uses the Vite proxy and does not need `VITE_API_BASE_URL`.

### 6.4 Docker Compose (Go)

```bash
cd d:\resume\back\resume-go
docker compose up -d --build
```

Services: `mysql:8.0` (3306), `redis:7-alpine` (6379), `rabbitmq:3.13-management` (5672/15672), `app` (8081).
MySQL initialisation mounts, in order: `../resume.sql` → `schema.sql` → `seed_accounts.sql` → `seed_statistics.sql`; `uploads` is mounted at `/app/uploads`.

### 6.5 Database initialisation order

```text
resume.sql           # base tables + legacy sample rows (user/student/company/stu_*)
schema.sql           # extra tables (job_postings/com_like_stu/chat_message/job_application/...)
seed_accounts.sql    # x1~x10 / q1~q10 / g1
seed_statistics.sql  # dashboard demo data
```

> `resume.sql` does not contain DDL for `job_postings`, `com_like_stu`, `admin_all_skills`, `language`, `certificate`; `resume-go/schema.sql` supplies them (as stated in its header comment).

---

## 7. Configuration

### 7.1 Go — environment variables

| Variable | Default | Notes |
|---|---|---|
| `SERVER_PORT` | `8081` | HTTP port |
| `DB_HOST` / `DB_PORT` | `localhost` / `3306` | MySQL |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `root` / see source / `resume` | MySQL |
| `UPLOAD_DIR` | `uploads/resume` | Resume dir; chat images go to sibling `chat/` |
| `JWT_SECRET` | fixed string (see `config.go`) | HS256 secret; externalise in production |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | `localhost:6379` / empty / `0` | Redis |
| `REDIS_ENABLED` | `true` | `false` = pure in-memory mode |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | AMQP |
| `RABBITMQ_ENABLED` | `true` | `false` skips the notification pipeline |
| `GRPC_PORT` | `9091` | gRPC port |
| `GRPC_ENABLED` | `true` | Enable gRPC |
| `GRPC_INTERNAL` | `127.0.0.1:<GRPC_PORT>` | Address the client uses |

DSN: `user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=true&loc=Local` (`loc=Local` avoids an 8-hour offset).

Pool: `MaxOpenConns=20`, `MaxIdleConns=10`, `ConnMaxLifetime=1h`.

### 7.2 Vite configuration

Portal `vite.config.js`:

```js
base: './',
alias: { '@': './src' },
server: {
  proxy: {
    '/api-backend': {
      target: 'http://127.0.0.1:8081',
      changeOrigin: true,
      ws: true,                       // proxies WebSocket too
      rewrite: p => p.replace(/^\/api-backend/, '')
    }
  }
}
```

Admin `vite.config.js`: `base` is `/admin/` in production, **no proxy** — it calls the backend cross-origin (CORS is enabled server-side).

---
## 8. Authentication & Authorization

### 8.1 Dual-token model

| Token | Lifetime | Transport | Purpose |
|---|---|---|---|
| Access Token | 2 hours | `Authorization: Bearer <token>` | Business APIs |
| Refresh Token | 7 days | `Refresh-Token: <token>` | Exchange for a new access token |

Claims:

```json
{
  "userId": 11,
  "username": "x1",
  "userType": "student",
  "tokenType": "access",   // access | refresh
  "exp": 1760000000
}
```

- Algorithm: `HS256` with the `JWT_SECRET` (default hard-coded in `config.go`; **externalise in production**).
- Implemented with stdlib `crypto/hmac`, `sha256`, and `base64.RawURLEncoding` — no third-party JWT library.
- Go verification order: split into three parts → verify signature (`hmac.Equal`, constant-time) → decode payload → validate `exp`.
- Refresh: verify signature/expiry → require `tokenType == "refresh"` → re-sign an access token from the original claims.

### 8.2 Request authentication

**Go (Gin middleware)**:

- Public: `POST /login`, `POST /refresh-token`, `POST /validateToken`, `GET /stu/statistics`, `GET /chat/ws`.
- Protected group `Group("/", authMiddleware)`: all remaining `/user/**`, `/stu/**`, `/com/**`, `/chat/**`.
- Admin group `Group("/admin", authMiddleware, adminMiddleware)`.
- `authMiddleware` also writes `New-Token` when the access token is near expiry (transparent renewal).
- Failures: `401 {"code":401,"message":"未提供Token"}` / `401 "Token无效或已过期"`; non-admin hitting the back office gets `403 "权限不足，需要管理员权限"`.

> **Note**: middleware verifies signature and expiry only — **`tokenType` is not enforced** on business routes, so a refresh token also works there (only `/refresh-token` requires the refresh type).

### 8.3 Passwords

- **BCrypt** (default cost 10); stored as `$2a$10$...`.
- Student initial password = student number; company initial password = company name; admin initial password is defined in `init_admin.sql` / `seed_accounts.sql`.
- `PATCH /stu/updatePwd` requires `old_pwd`, `new_pwd`, `re_pwd` (snake_case).

### 8.4 CORS

- Echoes `Origin` (`*` when absent), exposes `New-Token`, allows `Authorization, Content-Type, Refresh-Token, X-Requested-With, platform`.
- `OPTIONS` returns 204 immediately.

### 8.5 Front-end authentication

| App | Token storage | Guard logic |
|---|---|---|
| Portal | Pinia `rec-user` (persist → `localStorage`) | Route guard decodes the JWT locally and checks `exp`; on failure `clear()` and redirect to `/login` |
| Admin | `localStorage['token']` | Only checks presence, never validity |

Admin login chain: portal `/login` (`userType=admin`) → token → redirect to `${VITE_ADMIN_URL}/admin/?token=xxx` → admin guard stores it in `localStorage` and removes the query with `history.replaceState`.

---

## 9. Response Envelope & Error Codes

### 9.1 Envelope

```json
{ "code": 200, "message": "操作成功", "data": { } }
```

- `model.Result`; `respond()` **always uses HTTP 200** — only middleware and the WebSocket handshake can return non-200.
- Pagination shape:

```json
{
  "total": 128,
  "list": [],
  "pageNum": 1,
  "pageSize": 10,
  "pages": 13,
  "size": 10,
  "firstPage": 1,
  "lastPage": 13
}
```

### 9.2 Error codes

| code | Meaning | Typical case |
|---|---|---|
| 200 | Success | — |
| 201 | Created | Resume upload |
| 400 | Bad request / validation failure | Missing field, wrong old password, illegal status |
| 401 | Unauthenticated | Missing/invalid/expired token, role mismatch |
| 403 | Forbidden | Non-admin on admin routes, editing someone else's job, admin chatting |
| 404 | Not found | Job/student/company/user missing |
| 409 | Conflict | Duplicate application, re-sending resume before reply |
| 40005 | Wrong username or password | Login |
| 40009 | Empty Excel content | Import |
| 40101 | Insufficient permission (redundant check) | Admin password reset |
| 40402 / 40403 | Some students / companies missing | Bulk delete |
| 41301 | File too large | Resume > 10MB; image > 5MB |
| 41501 | Unsupported file type | Not pdf/doc/docx, not an image, not xlsx |
| 42201 | Age out of range (16–60) | Profile update |
| 42202 | Invalid graduation year | Profile update |
| 42205 / 42206 | Student / company validation failed | Bulk update |
| 500 | Internal server error | Uncaught exception, SQL error |
| 50001 | DB write failure / wrong password | Import, password reset |

Business errors are returned as `model.AppError{Code, Msg}`; `respond()` uses that code directly.

---

## 10. Database Design

### 10.1 Table inventory

| Table | Purpose | Users |
|---|---|---|
| `user` | Unified account (student/company/admin) | All |
| `student` | Student basic info | All |
| `company` | Company info | All |
| `stu_language` | Language skills | Profile |
| `stu_certificate` | Certificates | Profile |
| `stu_internship` | Internships | Profile |
| `stu_job` | Preferred roles | Profile |
| `stu_region` | Preferred province/city | Profile |
| `stu_speciality` | Skill tags | Profile |
| `stu_resume` | Resume file records | Profile/chat/application |
| `job_postings` | Job postings | Jobs |
| `job_application` | Applications | Applications (Go) |
| `com_like_stu` | Company bookmarks | Bookmarks |
| `admin_all_skills` | Skill dictionary | Admin |
| `language` | Language dictionary | Admin |
| `certificate` | Certificate dictionary | Admin |
| `chat_message` | Chat messages | Chat (Go) |
| `chat_notification` | Offline notifications | Chat (Go) |
| `mq_consumed_event` | MQ idempotency log | Chat (Go) |
| `com_collect` | Legacy bookmarks | **Unused** |
| `role` | Legacy roles | **Unused** |

### 10.2 Core tables

**user**

| Column | Type | Notes |
|---|---|---|
| `user_id` | int AI PK | Primary key |
| `username` | varchar(100) | Student number / company name / `admin` |
| `password` | varchar(100) | BCrypt hash |
| `user_type` | varchar(25) | `student` / `company` / `admin` (sole auth basis) |
| `create_time` / `update_time` | datetime | Timestamps |

**student**

| Column | Type | Notes |
|---|---|---|
| `id` | int AI PK | **All `stu_*` child tables reference this via `stu_id`** |
| `stu_id` | int | Student number (matches `user.username`) |
| `name` / `age` / `gender` / `nation` | — | Basic profile |
| `major` / `degree` / `graduation_year` | — | Major / degree / graduation year |
| `internship` | int | Number of internships |
| `user_id` | int | FK by convention to `user.user_id` (index `user_stu`, no constraint) |

**company**

| Column | Type | Notes |
|---|---|---|
| `id` | int unsigned ZEROFILL AI PK | Company ID |
| `com_name` | varchar(255) | Company name |
| `com_region` | varchar(255) | Region |
| `user_id` | int | Account link (index `user_com`) |

**Profile child tables** (all keyed by `stu_id`):

| Table | Key columns | Notes |
|---|---|---|
| `stu_language` | `language`, `level`, `cert_no`, `issue_date` | Languages |
| `stu_certificate` | `cert_name`, `level`, `cert_no`, `issue_date` | Certificates |
| `stu_internship` | `company`, `experience` | Internship company + description |
| `stu_job` | `post` | Preferred role |
| `stu_region` | `province`, `city` | One row per province or city (other column NULL) |
| `stu_speciality` | `speciality` | Skill tag |
| `stu_resume` | `resume` | Disk path (append-only versions) |

**job_postings**

| Column | Type | Notes |
|---|---|---|
| `id` | int AI PK | — |
| `com_id` | int | **The only real FK** → `company(id)`, `ON DELETE CASCADE` |
| `title` | varchar(255) NOT NULL | Job title |
| `description` | text | Description |
| `external_link` | varchar(500) | External URL |
| `status` | enum/varchar, default `active` | `active` / `inactive` |
| `created_at` / `updated_at` | datetime | Auto-maintained |
| Indexes | `idx_com_id`, `idx_status_created` | List/detail |

**job_application**

| Column | Type | Notes |
|---|---|---|
| `id` | bigint AI PK | — |
| `job_id` | bigint | Job |
| `com_id` | bigint | Company (denormalised for queries and authorization) |
| `student_id` | bigint | `student.id` |
| `resume_url` | varchar | Resume snapshot at apply time |
| `status` | varchar, default `pending` | `pending/viewed/interview/offer/rejected` |
| `created_at` / `updated_at` | datetime | — |
| Constraint | `UNIQUE uk_job_student(job_id, student_id)` | Blocks duplicates |
| Indexes | `idx_student(student_id,id)`, `idx_company_status(com_id,status,id)` | List queries |

**com_like_stu**

| Column | Type | Notes |
|---|---|---|
| `id` | int AI PK | — |
| `stu_id` / `com_id` | int | Student / company |
| `create_time` | datetime | Bookmark time |
| Constraint | `UNIQUE uk_stu_com(stu_id, com_id)` | Blocks duplicates |
| Index | `idx_com_time(com_id, create_time)` | Ordered bookmark list |

**chat_message**

| Column | Type | Notes |
|---|---|---|
| `id` | bigint AI PK | Auto-increment; also the pagination cursor |
| `sender_id` / `sender_type` | int / varchar | Sender `user_id` + role |
| `receiver_id` / `receiver_type` | int / varchar | Receiver |
| `content` | text | Text, image URL, or resume URL |
| `msg_type` | varchar, default `text` | `text` / `image` / `resume` |
| `is_read` | tinyint | 0/1 |
| `created_at` | datetime | Server time |
| Indexes | `idx_chat_pair(sender_id,sender_type,receiver_id,receiver_type,id)`, `idx_chat_inbox(receiver_id,receiver_type,is_read,id)` | History / unread |

**chat_notification**

| Column | Notes |
|---|---|
| `id` | Primary key |
| `message_id` | Related message |
| `receiver_id` / `receiver_type` | Receiver |
| `sender_id` / `sender_type` | Sender |
| `title` / `content` | Title and summary (`[图片]`, `[简历]`, text truncated to 60 chars) |
| `is_read` / `created_at` | State and time |
| Constraint | `UNIQUE uk_notification_message(message_id, receiver_id, receiver_type)` — idempotency fallback |

**mq_consumed_event**: `event_id` (PK), `consumed_at` — MQ deduplication.

**Dictionaries**: `admin_all_skills(skillName, create_time)`, `language(language_name PK)`, `certificate(certificate_name PK)`.

### 10.3 Relationships

```text
user 1──1 student 1──N stu_language
                   ├──N stu_certificate
                   ├──N stu_internship
                   ├──N stu_job
                   ├──N stu_region
                   ├──N stu_speciality
                   ├──N stu_resume
                   └──N job_application

user 1──1 company 1──N job_postings 1──N job_application
                   └──N com_like_stu N──1 student

chat_message / chat_notification: logically linked to user via (user_id, user_type), no FK
```

> Except for `job_postings.com_id`, all relations are enforced only by application code (indexes, no FKs), so deleting a student/company requires manual child-row cleanup.

---

## 11. API Reference (Complete)

### 11.1 Conventions

| Item | Notes |
|---|---|
| Base URL | `http://<host>:8081` |
| Body | `application/json; charset=utf-8` (except uploads) |
| Auth | `Authorization: Bearer <accessToken>` |
| Refresh | `Refresh-Token: <refreshToken>` (only `/refresh-token`) |
| Envelope | `{code, message, data}` |
| Pagination | Lists use `pageNum`/`pageSize`; admin uses `page`/`pageSize` |
| File fields | Resume `resumeFile`; chat image `image`; student Excel `stuFile`; company Excel `comFile` |
| JSON naming | Requests accept both camelCase and snake_case (`userId` / `user_id`); responses emit both to stay compatible with older clients |

**Login response example**:

```json
{
  "code": 200,
  "message": "操作成功",
  "data": {
    "userId": 11,
    "username": "x1",
    "userType": "student",
    "displayName": "王一",
    "token": "...",
    "refreshToken": "..."
  }
}
```

### 11.2 Authentication & user

| Method | Path | Description | Parameters | Returns | Backend |
|---|---|---|---|---|---|
| POST | `/login` | Login | JSON: `username`, `password`, `userType` (`student/company/admin`) | access + refresh | Go |
| POST | `/refresh-token` | Refresh access | Header `Refresh-Token` | `{token}` | Go |
| POST | `/validateToken` | Validate token | Body `{token}` (Go also accepts raw text) | `{valid,message,userId,userType,needRefresh}` | Go |
| GET | `/user/me` | Current user profile | none | `{userId,username,userType,displayName}` | Go |
| PUT | `/user/name` | Rename self | `{name}` (≤20 chars) | `UserProfile` | Go |

Login failures: `code=40005, message="用户名或密码错误"`; role mismatch: `code=401, message="用户身份不匹配"`.

### 11.3 Student endpoints

| Method | Path | Description | Parameters | Returns | Backend |
|---|---|---|---|---|---|
| GET | `/stu/getAllSkill` | Skill dictionary | none | Comma-separated string | Go |
| PUT | `/stu/completeInfo` | Save full profile | `StudentCompleteInfoDTO` (below) | `StudentDetailDTO` | Go |
| POST | `/stu/uploadResume` | Upload resume | `multipart`, field `resumeFile` | `{resumeId,resumeName,resumeUrl}` (code=201) | Go |
| GET | `/stu/statistics` | Profile statistics (public) | none | Five aggregates | Go |
| PATCH | `/stu/updatePwd` | Change password | `{old_pwd,new_pwd,re_pwd}` | null | Go |
| GET | `/stu/resume/latest` | Latest resume | none | `{url,name}`; 404 if none | Go |
| POST | `/stu/applications` | Apply to a job | `{jobId,resumeUrl?}` | `JobApplication` | Go |
| GET | `/stu/applications` | My applications | none | `{list:[...]}` | Go |
| DELETE | `/stu/applications/:id` | Withdraw | path id | null | Go |

`/stu/completeInfo` body:

```json
{
  "age": 21,
  "gender": "女",
  "nation": "汉族",
  "major": "软件工程",
  "degree": "本科生",
  "graduationYear": 2028,
  "internship": 1,
  "languages": [{"language": "英语", "level": "CET-6"}],
  "cities": ["大连"],
  "provinces": ["辽宁"],
  "certificates": [{"certName": "计算机二级", "level": "良好"}],
  "internships": [{"company": "某科技", "experience": "后端开发实习"}],
  "specialities": ["Java", "MySQL"],
  "jobs": ["后端开发"]
}
```

(The backend also accepts snake_case such as `graduation_year` and `certificates[].cert_name`.)

Validation: age 16–60 (`42201`); graduation year within `[currentYear-50, currentYear+10]` (`42202`); child tables use delete-then-insert; student number and name are immutable.

`/stu/statistics` response:

```json
{
  "languages": [{"value": 12, "language": "英语", "level": "CET-6"}],
  "cities": [{"value": 8, "name": "大连"}],
  "internships": [{"value": 5, "company": "某科技", "experience": "后端开发"}],
  "specialities": [{"value": 20, "name": "Java"}],
  "jobs": [{"value": 15, "name": "后端开发"}]
}
```

### 11.4 Company: jobs

| Method | Path | Description | Parameters | Returns | Backend |
|---|---|---|---|---|---|
| POST | `/com/jobPosting` | Create job | `{title,description,external_link,status}`; `com_id` overwritten from token | `JobPosting` | Go |
| PUT | `/com/jobPosting/{id}` | Update job | same (partial) | `JobPosting` | Go |
| PATCH | `/com/jobPosting/{id}/status` | Activate/deactivate | `{status:"active"}` or `{status:"inactive"}` | `JobPosting` | Go |
| DELETE | `/com/jobPosting/{id}` | Delete job | path id | `{success,message}` | Go |
| GET | `/com/jobPostings` | Paginated jobs | `pageNum`, `pageSize`, `status`, `keyword`, `comId` | `PageInfo<JobPosting>` | Go |
| GET | `/com/jobPosting/{id}` | Job detail | path id | `JobPosting` | Go |
| GET | `/com/myJobPostings` | My jobs (no paging) | `status?` | `JobPosting[]` | Go |
| GET | `/com/applications` | Received applications | `jobId?`, `status?` | `{list}` | Go |
| PATCH | `/com/applications/:id/status` | Advance status | `{status}` | `JobApplication` | Go |

Authorization: write operations resolve `company.id` from the token's `userId` and compare it with the job's `com_id`; mismatch → 403.

`JobPosting` fields: `id`, `comId`, `comName`, `title`, `description`, `externalLink`, `status`, `createdAt`, `updatedAt` (Go also emits `com_id`, `external_link`, `created_at`).

### 11.5 Company: bookmarks

| Method | Path | Description | Parameters | Returns | Backend |
|---|---|---|---|---|---|
| GET | `/com/QueryComIdByUserId` | Company id of current user | none | int | Go |
| POST | `/com/LikeAdd` | Bulk bookmark | `[{stuId,comId}]` | null; duplicates → 500 + message | Go |
| POST | `/com/LikeQuery` | Paginated bookmarks | `{com_id,comName?,pageNum,pageSize}` | `PageInfo<StuFavorite>` | Go |
| DELETE | `/com/LikeDelete` | Bulk unbookmark | `[{stuId,comId}]` | null | Go |
| GET | `/com/LikeIf` | Is bookmarked (DB) | `com_id`, `stu_id` | bool | Go |
| GET | `/com/LikeIfByRedis` | Bulk check (Redis) | `stuIdList=1,2,3` | `{stuId:true}` | Go |

`StuFavorite`: `stuId`, `studentName`, `gender`, `major`, `degree`, `graduationYear`, `languages`, `level`, `specialities`, `post` (aggregated by `GROUP_CONCAT`).

### 11.6 Search (company/admin)

| Method | Path | Description | Parameters | Returns | Backend |
|---|---|---|---|---|---|
| POST | `/com/search` | Student list (basic) | `StudentQueryDTO` | `PageInfo<Student>` | Go |
| POST | `/com/detailSearch` | Student list (full profile) | same | `PageInfo<StudentDetailDTO>` | Go |
| POST | `/admin/searchStuDetail` | Admin detail search | same | same | Go |

`StudentQueryDTO` fields:

| Type | Fields |
|---|---|
| Exact | `id`, `stu_id`, `age`, `gender`, `degree`, `graduation_year` |
| Fuzzy | `name`, `nation`, `major` |
| Range | `internship` (>=) |
| Lists | `cities[]`, `provinces[]`, `languages[{language,level}]`, `cert_names[]`, `specialities[]`, `jobs[]` |
| Other | `cert_level` (fuzzy), `internship_company` (fuzzy) |
| Paging | `page_num`, `page_size` (default 1 / 10) |

### 11.7 Admin

| Method | Path | Description | Parameters | Returns |
|---|---|---|---|---|
| PUT | `/admin/updateStuPwd/{stu_id}` | Reset student password | Query `pwd` | success message |
| POST | `/admin/addStu` | Excel import students | `stuFile` (.xlsx) | `{successCount,failureCount,errors[]}` |
| POST | `/admin/addCom` | Excel import companies | `comFile` (.xlsx) | same |
| DELETE | `/admin/delStu` | Bulk delete students | `{stuIds:[...]}` | result; partial → 40402 |
| DELETE | `/admin/delCom` | Bulk delete companies | `{comIds:[...]}` | result; partial → 40403 |
| PUT | `/admin/updateStu` | Bulk update students | `{students:[{id,...}]}` | partial → 42205 |
| PUT | `/admin/updateCom` | Bulk update companies | `{companies:[{id,...}]}` | partial → 42206 |
| GET | `/admin/searchStu` | Paginated students | `username` (name LIKE), `page`, `pageSize` | `{list,pagination}` |
| GET | `/admin/searchCom` | Paginated companies | `companyName`, `page`, `pageSize` | `{list,pagination}` |
| POST | `/admin/searchStuDetail` | Detail search | `StudentQueryDTO` | `PageInfo` |
| GET | `/admin/getAllLanguage` | Language dictionary | none | `[]string` |
| POST | `/admin/addLanguage` | Add language | form `language` | true |
| POST | `/admin/delLanguage` | Delete language | form `language` | true |
| GET | `/admin/getAllCertificate` | Certificate dictionary | none | `[]string` |
| POST | `/admin/addCertificate` | Add certificate | form `certificate` | true |
| POST | `/admin/delCertificate` | Delete certificate | form `certificate` | true |
| POST | `/admin/addSkillAll` | Bulk add skills | raw text `Java,Go` | message |
| GET | `/admin/getAllSkill` | Skill dictionary | none | comma string |
| DELETE | `/admin/deleteSkill` | Delete skills | Query `skills=Java,Go` | message |

Bulk result:

```json
{
  "successCount": 9,
  "failureCount": 1,
  "errors": [{"identifier": "20240001", "reason": "学号已存在"}]
}
```

### 11.8 Chat

| Method | Path | Description | Parameters | Returns |
|---|---|---|---|---|
| GET | `/chat/ws` | WebSocket handshake | Query `token` | 101 upgrade |
| GET | `/chat/peers` | Chat-capable peers | `keyword?` | `{list:[ChatPeer]}` |
| GET | `/chat/conversations` | Conversation list | none | `{list:[Conversation]}` |
| GET | `/chat/messages` | History | `peerId`, `peerType`, `beforeId?`, `limit?` (default 30, max 100) | `{messages,hasMore}` |
| POST | `/chat/messages` | Send via REST (fallback) | `{receiverId,receiverType,content,msgType}` | `ChatMessage` |
| POST | `/chat/uploadImage` | Chat image | `image` (≤5MB) | `{url:"/uploads/chat/xxx.png"}` |
| POST | `/chat/read` | Mark conversation read | `{peerId,peerType}` | null |
| GET | `/chat/resume-permission` | May I send a resume? | `peerId`, `peerType` | `{canSend,waitingReply}` |
| GET | `/chat/profile` | Peer profile | `peerId`, `peerType` | student profile / company info + active jobs |
| GET | `/chat/notifications` | Offline notifications | `limit?` (default 20, max 100) | `{list}` |
| POST | `/chat/notifications/read` | Mark all read | none | null |
| GET | `/chat/notifications/unread` | Unread count | none | `{count}` |

### 11.9 Static assets and fallback

| Path | Notes |
|---|---|
| `GET /uploads/resume/<file>` | Resume file |
| `GET /uploads/chat/<file>` | Chat image |
| `GET /resume/uploads/<file>` | Legacy prefix |
| Anything else | `NoRoute` → `{code:404,message:"接口不存在"}` (HTTP 200) |

### 11.10 Request examples

```bash
# Login
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"username":"x1","password":"x1","userType":"student"}'

# Job list with token
curl "http://localhost:8081/com/jobPostings?pageNum=1&pageSize=10&status=active" \
  -H "Authorization: Bearer <accessToken>"

# Resume upload
curl -X POST http://localhost:8081/stu/uploadResume \
  -H "Authorization: Bearer <accessToken>" \
  -F "resumeFile=@C:/tmp/resume.pdf"

# Admin bulk delete
curl -X DELETE http://localhost:8081/admin/delStu \
  -H "Authorization: Bearer <adminToken>" \
  -H "Content-Type: application/json" \
  -d '{"stu_ids":[1,2]}'
```

---
## 12. Chat System Deep Dive

Chat exists only in the **Go backend**; the UI is the portal's `/chat` page.

### 12.1 Design: REST + WebSocket

| Channel | Responsibility |
|---|---|
| HTTP REST | Login, peers, conversations, history, mark-read, fallback sending, image upload, notifications |
| WebSocket | Real-time new-message push, connection notice, heartbeat |

Why:

1. History is naturally paginated — HTTP is simpler.
2. WebSocket handles only "real time"; a disconnect never affects stored history.
3. When the socket is down, REST still delivers the message.

### 12.2 Connection and authentication

```text
ws://<host>:8081/chat/ws?token=<accessToken>
```

Browsers cannot set custom headers on `new WebSocket(url)`, so the token travels in the query string. Server flow:

1. Read `token`; empty → HTTP 401 `缺少 token`.
2. `auth.Manager.Parse(token)` verifies signature and `exp`; failure → 401 `token 无效或已过期`.
3. Role must be `student` or `company`; `admin` → 403 `只有学生和企业可以使用聊天`.
4. `websocket.Upgrader.Upgrade(...)` (`CheckOrigin` always true — tighten in production).
5. `chat.NewClient(hub, conn, userId, userType, username)` → `client.Start()`.
6. Immediately push `{"type":"connected","userId":11,"userType":"student"}`.

`/chat/ws` bypasses `authMiddleware`; every other `/chat/*` route requires `Authorization: Bearer`.

### 12.3 Hub: the connection registry

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]map[*Client]struct{} // "student:11" -> all connections of that user
    OnMessage    func(*Client, IncomingMessage)
    OnRegister   func(*Client)
    OnUnregister func(*Client)
}
```

| Method | Behaviour |
|---|---|
| `Register` | Create the set → add the connection → log "user online, connections N" → `go OnRegister(c)` async |
| `Unregister` | Remove only this connection; delete the outer key only when the set is empty (other tabs stay online) → `go OnUnregister(c)` async |
| `SendToUser(id,type,payload)` | Marshal once, `RLock`, iterate all connections and `push` |
| `IsOnline(id,type)` | Local process table only |

Key points:

- **A user may have multiple connections** (tabs/devices), hence a set per user.
- All map access is guarded (concurrent map read/write panics in Go).
- The Hub never imports `service`; injected callbacks break the cycle.

### 12.4 Client: two goroutines

Constants:

```go
writeWait      = 10 * time.Second
pongWait       = 60 * time.Second
pingPeriod     = 54 * time.Second   // pongWait * 9 / 10
maxMessageSize = 4 * 1024
sendBufferSize = 128
```

**readPump** (single reader)

1. `SetReadLimit(4KB)`, `SetReadDeadline(now+60s)`.
2. `SetPongHandler` refreshes the deadline by 60s.
3. Loop `ReadMessage()`:
   - error → exit;
   - any data refreshes the read deadline;
   - invalid JSON → push `{"type":"error","message":"消息格式错误"}` and continue;
   - `type=="ping"` → push `{"type":"pong"}`;
   - `type=="message"` → `Hub.OnMessage`;
   - anything else → push `{"type":"error","message":"未知的消息类型: X"}`.
4. `defer` cleanup order: `Hub.Unregister` → `close(send)` → `conn.Close()` (order matters).

**writePump** (single writer)

```text
business goroutine --push--> send chan (128) --> writePump --> conn.WriteMessage
```

- gorilla/websocket **forbids concurrent writes**; all writes are serialized here.
- Two `select` cases: `data, ok := <-send` (`ok=false` → send a Close frame and exit) and `<-ticker.C` (protocol Ping every 54s).
- Any write error returns; `conn.Close()` unblocks `ReadMessage`, which triggers readPump cleanup.

**Non-blocking enqueue (Hub protection)**

```go
select {
case c.send <- data:
default:
    // queue full: drop + log (message is already persisted; history reload recovers it)
}
```

### 12.5 Two heartbeat layers

| Layer | Frame | Initiator | Purpose |
|---|---|---|---|
| Protocol | WebSocket Ping/Pong control frames | Server pings every 54s, browser auto-pongs | Keep TCP/middleboxes alive; refreshes the 60s read deadline |
| Application | `{"type":"ping"}` / `{"type":"pong"}` | Client every 25s | Client-side health check; no pong for 35s → force reconnect |

### 12.6 Message protocol

Client → server:

```json
{"type":"message","receiverId":12,"receiverType":"company","content":"hello","msgType":"text"}
{"type":"ping"}
```

Server → client:

| type | Shape | When |
|---|---|---|
| `connected` | `{"type":"connected","userId":11,"userType":"student"}` | On connect |
| `pong` | `{"type":"pong"}` | Application ping received |
| `message` | `{"type":"message","data":ChatMessage}` | Business message (pushed to both sides) |
| `error` | `{"type":"error","message":"..."}` | Bad format / unknown type / validation failure |

`ChatMessage`:

```json
{
  "id": 101,
  "senderId": 11,
  "senderType": "student",
  "senderName": "王一",
  "receiverId": 12,
  "receiverType": "company",
  "receiverName": "企业一",
  "content": "hello",
  "msgType": "text",
  "isRead": false,
  "createdAt": "2026-10-23 21:43:27"
}
```

### 12.7 End-to-end sequence

```text
Sender browser                  Go backend                          Receiver browser
   │ 1. ws?token=xxx              │                                      │
   │─────────────────────────────►│ 2. verify, Upgrade, Hub.Register     │
   │                              │                                      │
   │ 3. {"type":"message",...}    │                                      │
   │─────────────────────────────►│ 4. service.SendChatMessage           │
   │                              │    - complementary-role check        │
   │                              │    - msgType/length/image/resume     │
   │                              │    - receiver existence (user table) │
   │                              │ 5. INSERT chat_message               │
   │                              │ 6. SELECT back (id/time/names)       │
   │                              │ 7. Hub.SendToUser(receiver) ────────►│ live
   │◄─────────────────────────────│ 8. Hub.SendToUser(sender)            │ multi-device
   │                              │ 9. MQ.Publish(chat.message.created)  │
   │                              ▼                                      │
   │                     RabbitMQ → HandleChatNotify → offline notice     │
```

**Persist first, push second**: an offline receiver never loses a message.

### 12.8 Validation before sending (`service.SendChatMessage`)

| Check | Failure code |
|---|---|
| Sender must be `student/company` | 403 |
| Receiver must be the complementary role | 400 |
| Cannot message yourself / `receiverId<=0` | 400 |
| `msgType` ∈ `text/image/resume` (empty → `text`) | 400 |
| Text non-empty after trimming; ≤800 **runes** | 400 |
| `image` content must start with `/uploads/chat/` (anti-spoofing) | 400 |
| `resume` only from students | 403 |
| `resume` must be a local URL (`/uploads/resume/` or legacy `/api/resume/download/`), ≤500 chars | 400 |
| Resume throttle: already sent one since the peer's last reply | 409 |
| Receiver must exist in `user` | 404 |

Resume-throttle SQL:

```sql
SELECT COUNT(*) FROM chat_message
WHERE sender_id=? AND sender_type=? AND receiver_id=? AND receiver_type=?
  AND msg_type='resume'
  AND id > COALESCE((
      SELECT MAX(id) FROM chat_message
      WHERE sender_id=? AND sender_type=? AND receiver_id=? AND receiver_type=?
  ), 0)
```

### 12.9 Conversation model and list SQL

**There is no `conversation` table** — a conversation is a view aggregated per peer:

```sql
SELECT t.peer_id, t.peer_type, t.unread,
       COALESCE(ss.name, cc.com_name, '') AS peer_name,
       m.content, m.msg_type, m.created_at, m.sender_id, m.sender_type
FROM (
  SELECT
    CASE WHEN sender_id=? AND sender_type=? THEN receiver_id ELSE sender_id END AS peer_id,
    CASE WHEN sender_id=? AND sender_type=? THEN receiver_type ELSE sender_type END AS peer_type,
    MAX(id) AS max_id,
    SUM(CASE WHEN receiver_id=? AND receiver_type=? AND is_read=0 THEN 1 ELSE 0 END) AS unread
  FROM chat_message
  WHERE (sender_id=? AND sender_type=?) OR (receiver_id=? AND receiver_type=?)
  GROUP BY peer_id, peer_type
) t
JOIN chat_message m ON m.id = t.max_id
LEFT JOIN student ss ON t.peer_type='student' AND ss.user_id=t.peer_id
LEFT JOIN company cc ON t.peer_type='company' AND cc.user_id=t.peer_id
ORDER BY t.max_id DESC
```

Notes:

- `CASE` normalises both directions into "the peer".
- `MAX(id)` is the last message (safer than `MAX(created_at)`, which collides at second granularity).
- `SUM(is_read=0)` is the total unread for that conversation.
- `LastFromMe` must compare both `sender_id` and `sender_type` (student/company ids can collide).

### 12.10 History: keyset pagination

```sql
SELECT m.*, sender_name, receiver_name
FROM chat_message m
LEFT JOIN student ss ON m.sender_type='student'  AND ss.user_id=m.sender_id
LEFT JOIN company cc ON m.sender_type='company'  AND cc.user_id=m.sender_id
LEFT JOIN student rs ON m.receiver_type='student' AND rs.user_id=m.receiver_id
LEFT JOIN company rc ON m.receiver_type='company' AND rc.user_id=m.receiver_id
WHERE ((me -> peer) OR (peer -> me))
  AND m.id < ?          -- beforeId cursor
ORDER BY m.id DESC
LIMIT ?
```

- `id < beforeId` instead of OFFSET avoids deep-pagination cost and duplicates/misses when new rows arrive.
- Results are reversed to oldest → newest in Go.
- `hasMore = len(messages) == limit`.
- The client passes `messages[0].id` as `beforeId` and restores scroll position with `scrollHeight - oldHeight + oldTop`.

### 12.11 Read / unread

```sql
UPDATE chat_message SET is_read = 1
WHERE sender_id=? AND sender_type=? AND receiver_id=? AND receiver_type=? AND is_read=0
```

Only messages received from the peer are marked read. The client calls `/chat/read` when a conversation is opened, clears the local unread badge, and the top navigation polls `/chat/conversations` every 15s to aggregate the badge (shows `99+` above 99).

### 12.12 Offline notifications: MQ + three idempotency layers

After a successful send, an event is published on a best-effort basis:

```json
{
  "eventId": "chat.message.created:101",
  "type": "chat.message.created",
  "messageId": 101,
  "senderId": 11,
  "senderType": "student",
  "receiverId": 12,
  "receiverType": "company",
  "content": "hello",
  "msgType": "text",
  "createdAt": "2026-10-23T21:43:27+08:00"
}
```

`HandleChatNotify`:

1. **MySQL idempotency**: `INSERT IGNORE INTO mq_consumed_event`; continue only when `RowsAffected==1`.
2. **Redis idempotency**: `SETNX resume:mq:consumed:{eventId}` (24h); returns true when Redis is down so MySQL is the fallback.
3. **Skip if online**: the receiver already got the WS push.
4. **Resolve sender name**: gRPC `GetUserBrief` → local DB → literal `"对方"`.
5. **Build a summary**: image → `[图片]`, resume → `[简历]`, text truncated to 60 runes.
6. **Insert notification**: `INSERT IGNORE INTO chat_notification` (unique key is the third guard).

> The order "DB first, Redis second" is deliberate: holding Redis first and then failing the DB write would mark the event consumed and permanently drop the notification.

### 12.13 Presence

| Event | Action |
|---|---|
| Register | `go OnRegister` → `Cache.SetOnline`: `INCR resume:online:{type}:{id}` + `EXPIRE 300s` |
| Unregister | `go OnUnregister` → `Cache.SetOffline`: `DECR`, `DEL` at ≤0 |
| Query | Redis `EXISTS` → fallback to local `Hub.IsOnline` |

- A counter (not a boolean) supports multiple tabs.
- TTL cleans up after a killed process.
- The client refreshes `online` via `/chat/conversations` or `/chat/peers` every 8s; **there is no presence push**.
- Known defect: the Redis TTL is set only at registration, so connections older than 5 minutes lose the Redis key (single instance falls back to the local Hub; multiple instances may misjudge).

### 12.14 Image and resume messages

**Image**: `POST /chat/uploadImage` (field `image`, ≤5MB, `jpg/jpeg/png/gif/webp`) → stored in `uploads/chat/` as `{userId}_{yyyyMMddHHmmss}_{8-hex}{ext}` → returns `/uploads/chat/xxx`, then sent as a `msgType=image` message.

**Resume**: the student fetches `GET /stu/resume/latest` and sends that URL with `msgType=resume`; the company can preview the PDF inline. The throttle allows one resume per peer reply.

### 12.15 Client WebSocket wrapper (`src/utils/chatSocket.js`)

```js
const socket = new ChatSocket({
  token: userStore.token,
  onMessage: handleSocketMessage,
  onStatus: s => (connectionStatus.value = s)
})
socket.connect()
```

Details:

- URL: take `API_BASE_URL` → prepend `window.location.origin` if relative → `replace(/^http/, 'ws')` → append `/chat/ws?token=...`.
  In development: `ws://localhost:5173/api-backend/chat/ws?token=...` (Vite `ws:true` proxies to 8081).
- Heartbeat: application ping every 25s; no pong for 35s → force `close()` to trigger reconnect.
- Reconnect: **exponential backoff 1s → 2s → 4s → 8s → 15s (cap)**, reset to 1s after success; disabled when `manualClose`.
- Messages: `type==='pong'` only updates `lastPongAt`; everything else goes to `onMessage`.
- Send: `sendMessage(receiverId, receiverType, content, msgType)`, returns true only when `readyState===OPEN`.

### 12.16 Chat page (`views/chat/index.vue`)

1. On mount: load conversations → load peers → students also load their resume → connect the socket.
2. Refresh presence every 8s; refresh immediately on window focus.
3. Select peer: clear messages → load 30 → refresh resume permission → mark read → clear local unread → `router.replace('/chat?peerId&peerType&name')`.
4. On message: de-duplicate by `id`; append if it belongs to the open conversation and scroll to bottom; otherwise debounce-refresh the conversation list (300ms).
5. Send: prefer WS; if it is not connected, fall back to `POST /chat/messages` and feed the response through the same de-duplicating handler.
6. Image: validate MIME and ≤5MB → upload → send `image`.
7. Resume: `getMyResume({silent:true})` → warn if none → send `resume` → optimistically set `canSendResume=false`.
8. Pinned resume card: students see their own latest resume; companies see the current peer's resume from `/chat/profile`.
9. Input: `Enter` sends, `Shift+Enter` newline, `Ctrl/Cmd+Enter` inserts a newline, IME composition never sends.
10. Unmount: close the socket and clear all timers/listeners.

### 12.17 Limits and extensions

| Limitation | Notes / extension |
|---|---|
| Single-instance Hub | No cross-instance push; use Redis Pub/Sub or MQ fan-out |
| MQ drops failures | `Nack(requeue=false)`, no retry, no DLQ (notifications are allowed to drop) |
| No recall / read receipts | Add `recalled_at` and a `{type:"read"}` push |
| No client message id | Add `clientMsgId` + unique index to survive retries |
| Token in URL | May leak into logs; use `wss://` and short-lived tokens |
| No presence broadcast | Clients must poll |

---

## 13. Resume & File Storage

### 13.1 Upload pipeline

| Step | Implementation |
|---|---|
| Field | `resumeFile` |
| Size | ≤10MB (`41301`) |
| Types | `.pdf/.doc/.docx` (`41501`) |
| Naming | `{studentNo}_{yyyyMMddHHmmss}_{8-hex}{ext}` |
| Storage | `UPLOAD_DIR` (default `uploads/resume`) |
| DB | `stu_resume(stu_id=student.id, resume=path)`, append-only |
| Response | `{resumeId,resumeName,resumeUrl}`, code=201 |
| Latest | `GET /stu/resume/latest` (`ORDER BY id DESC LIMIT 1`) |

Security: the **original filename is never reused**; extensions are lower-cased and whitelisted, preventing overwrite and `../` traversal.

### 13.2 Serving URLs

| Working URLs |
|---|
| `/uploads/resume/<file>`, `/uploads/chat/<file>`, legacy `/resume/uploads/<file>` |

**Known inconsistencies**:

- The upload endpoint returns `resumeUrl = /api/resume/download/<file>`, but **that route is not registered**.
- "Latest resume", applications, and chat resume messages go through `publicResumeURL()`, which converts to `/uploads/resume/<base>` and works.
- Legacy DB rows may contain `uploads\resume\xxx.pdf` (backslashes); the client `fileUrl()` normalises them.

---

## 14. Search, Pagination & Query Design

### 14.1 Dynamic WHERE

`data/studentWhere(q)`, for example:

| Condition | Implementation |
|---|---|
| Exact | `s.id=?`, `s.stu_id=?`, `s.age=?`, `s.gender=?`, `s.degree=?`, `s.graduation_year=?` |
| Fuzzy | `s.name LIKE CONCAT('%',?,'%')` (nation, major likewise) |
| Range | `s.internship >= ?` |
| Region | `sr.city IN (...)`, `sr.province IN (...)` |
| Languages | one `EXISTS (SELECT 1 FROM stu_language WHERE stu_id=s.id AND language=? [AND level=?])` per language, OR-ed |
| Certificates | `s.id IN (SELECT stu_id FROM stu_certificate WHERE cert_name IN (...))` |
| Certificate level | `level LIKE` |
| Internship company | `EXISTS (... company LIKE ...)` |
| Skills | `EXISTS (... speciality IN (...))` |
| Roles | `EXISTS (... post IN (...))` |

All values are bound with `?` — no SQL injection.

### 14.2 Pagination

- List pagination: explicit `SELECT COUNT(*) FROM (main query) t` + `LIMIT ? OFFSET ?`, returning the `PageInfo` shape the front end expects.
- Chat history: cursor pagination with `id < beforeId`.

### 14.3 Avoiding N+1

`detailSearch` uses two steps:

1. Page the main table (with region aggregation and a latest-resume window function).
2. Collect the page's `student.id` values and run one `WHERE stu_id IN (...)` per child table, then fill a `map[stuId]` in memory.

```sql
-- latest resume via MySQL 8 window function
SELECT * FROM (
  SELECT r.*, ROW_NUMBER() OVER (PARTITION BY stu_id ORDER BY id DESC) rn
  FROM stu_resume r
) t WHERE rn = 1
```

### 14.4 Statistics

`GET /stu/statistics` runs five `GROUP BY` queries:

```sql
SELECT COUNT(*) AS value, language, level FROM stu_language GROUP BY language, level ORDER BY value DESC;
SELECT COUNT(*) AS value, city AS name FROM stu_region WHERE city IS NOT NULL GROUP BY city ORDER BY value DESC;
SELECT COUNT(*) AS value, company, MAX(experience) AS experience FROM stu_internship
  WHERE company IS NOT NULL GROUP BY company ORDER BY value DESC;
SELECT COUNT(*) AS value, speciality AS name FROM stu_speciality GROUP BY speciality ORDER BY value DESC;
SELECT COUNT(*) AS value, post AS name FROM stu_job GROUP BY post ORDER BY value DESC;
```

Go closes each `rows` explicitly to avoid exhausting the pool across sequential queries.

---

## 15. Bookmark (Favorite) Module

### 15.1 Model

- Primary store: `com_like_stu(stu_id, com_id, create_time)` with unique key `(stu_id, com_id)`.
- Cache: Redis Set `resume:favorite:{userId}`, members are student-id strings.

### 15.2 Bookmark flow

1. Validate that both `student.id` and `company.id` exist.
2. Batch-check existing pairs with a row constructor: `WHERE (stu_id, com_id) IN ((?,?),(?,?))`.
3. Filter duplicates, then batch insert.
4. If inserted rows ≠ expected rows, throw and roll back.
5. If duplicates existed, return "failed to add N students: already bookmarked".

### 15.3 Unbookmark

`DELETE FROM com_like_stu WHERE (stu_id, com_id) IN ((?,?),...)`, returning the number of rows not deleted; if none existed, return "student does not exist". Successful deletes also run `SREM resume:favorite:{userId}`.

### 15.4 Bookmark list query

```sql
SELECT f.stu_id, s.name, s.gender, s.major, s.degree, s.graduation_year,
       (SELECT GROUP_CONCAT(language SEPARATOR '、') FROM stu_language WHERE stu_id=s.id) AS languages,
       (SELECT GROUP_CONCAT(speciality SEPARATOR '、') FROM stu_speciality WHERE stu_id=s.id) AS specialities,
       (SELECT GROUP_CONCAT(post SEPARATOR '、') FROM stu_job WHERE stu_id=s.id) AS post,
       f.create_time
FROM com_like_stu f
JOIN student s ON s.id = f.stu_id
WHERE f.com_id = ?
ORDER BY f.create_time DESC
LIMIT ? OFFSET ?
```

Correlated subqueries + `GROUP_CONCAT` avoid the Cartesian-product row inflation (and wrong `total`) caused by joining multiple one-to-many tables.

### 15.5 Two ways to check

| Endpoint | Source | Use case |
|---|---|---|
| `/com/LikeIf` | MySQL | Strong consistency, single check |
| `/com/LikeIfByRedis` | Redis Set | Bulk, low latency, tolerates short staleness |

---

## 16. Resume Application Module

Implemented in `internal/service/application.go` and `internal/data/application.go`.

### 16.1 State machine

```text
pending → viewed → interview → offer
                        └→ rejected
```

Constants: `pending`, `viewed`, `interview`, `offer`, `rejected`.

### 16.2 Student-side validation

1. Job must exist (404).
2. Job status must be `active` (400).
3. Empty `resumeUrl` falls back to the latest resume; none → 400.
4. `resumeUrl` must be a local URL (400).
5. `UNIQUE uk_job_student(job_id, student_id)` blocks duplicates; MySQL error **1062** is mapped to `ErrDuplicateApplication` → **409**.
6. `resume_url` is a **snapshot**; later resume changes do not affect submitted applications.

### 16.3 Company side

- List: `GET /com/applications?jobId=&status=`, joining job title, company name, and student name/major/degree.
- Status update: `UPDATE ... WHERE id=? AND com_id=?` (the `com_id` predicate prevents horizontal privilege escalation).
- Withdraw: `DELETE ... WHERE id=? AND student_id=?`.

---

## 17. Excel Bulk Import

### 17.1 Template columns

**Students** (row 1 is the header; data starts at row 2; columns 0–8):

| Column | Field |
|---|---|
| 0 | Student number `stu_id` |
| 1 | Name `name` |
| 2 | Age `age` |
| 3 | Gender `gender` |
| 4 | Nation `nation` |
| 5 | Major `major` |
| 6 | Degree `degree` |
| 7 | Graduation year `graduation_year` |
| 8 | Internship count `internship` |

**Companies** (columns 0–1): company name `com_name`, region `com_region`.

Templates: `back/resume-go/templates/学生数据导入模板.xlsx`, `企业数据导入模板.xlsx`.

### 17.2 Implementation

| Item | Implementation |
|---|---|
| Library | excelize v2 |
| Reading | First sheet only, skip row 1, read by column index |
| Cells | NUMERIC/STRING/BOOLEAN, numeric → int |
| Dedup | Students by `stu_id`, companies by `com_name` |
| Transaction | Insert inside a transaction; row-level errors collected without aborting |
| Default password | Student number / company name (BCrypt) |
| Result | `{successCount,failureCount,errors[{identifier,reason}]}` |
| Partial failure | `code=200, message="部分数据导入失败"` |

Insert order: `user` first (`useGeneratedKeys` fills `user_id`), then `student` / `company` with that `user_id`.

> Import only — **export is not implemented**.

---

## 18. Redis Cache Design

### 18.1 Keys

| Key | Type | TTL | Purpose |
|---|---|---|---|
| `resume:job:version` | string(int) | none | Generation counter for job caches |
| `resume:job:v{n}:list:{status}\|{keyword}\|{comId}\|{pageNum}\|{pageSize}` | string(JSON) | 30s + ≤10% jitter | Job lists |
| `resume:job:v{n}:detail:{id}` | string(JSON) | 5min + jitter | Job detail |
| `resume:favorite:{userId}` | Set | none | Company bookmark set |
| `resume:online:{type}:{id}` | string(counter) | 5min | Presence (multi-device counter) |
| `resume:mq:consumed:{eventId}` | string | 24h | Fast MQ idempotency |

### 18.2 Cache concerns

| Concern | Solution |
|---|---|
| Stampede | `singleflight.Group` — one reload per key under concurrency |
| Avalanche | TTL jitter of up to 10% |
| Consistency | Job writes `INCR resume:job:version`; readers build new keys; old keys expire naturally (no SCAN) |
| Penetration | Missing detail returns an error; **nulls are not cached** |
| Degradation | Redis down → in-memory map; `MarkOnce` returns true so MySQL's unique key is the guard |

---

## 19. Internal gRPC Service

### 19.1 Definition (`proto/recruit/v1/recruit.proto`)

```proto
service RecruitInternal {
  rpc GetUserBrief(GetUserBriefRequest) returns (UserBrief);
  rpc GetJobBrief(GetJobBriefRequest) returns (JobBrief);
}
```

| Method | Input | Output |
|---|---|---|
| `GetUserBrief` | `user_id`, `user_type` | `user_id`, `user_type`, `display_name`, `exists` |
| `GetJobBrief` | `job_id` | `job_id`, `title`, `company_name`, `status`, `exists` |

### 19.2 Implementation and callers

- Server: `rpc.Serve(ctx, ":9091", db)` with a logging interceptor (method, latency, error); `GracefulStop()` on shutdown with a 5s fallback to `Stop()`.
- Client: `grpc.NewClient` (lazy) with `insecure` credentials and 2s timeouts.
- **Only real caller**: `HandleChatNotify` for the sender's display name (falls back to DB).
- `JobBrief` has no callers yet (reserved).
- **Calls itself in-process only** (`127.0.0.1:9091`); it is not an external business API.

---

## 20. RabbitMQ Offline Notifications

| Item | Value |
|---|---|
| Queue | `resume.chat.notify` (durable) |
| Exchange | default `""`, routing key = queue name |
| Event type | `chat.message.created` |
| Properties | JSON, `Persistent`, `MessageId=eventId` |
| Producer | Only `SendChatMessage`; failures are logged |
| Consumer | Only `HandleChatNotify` |
| Prefetch | `Qos(20)` |
| ACK | Success → `Ack`; JSON error or handler error → `Nack(requeue=false)` (drop, no poison loop, no DLQ) |
| Reconnect | Consumer loop retries after 5s; publisher detects a closed connection and reconnects |
| Degradation | `RABBITMQ_ENABLED=false` makes `Publish` a no-op |

---

## 21. Front-End Implementation

### 21.1 Portal request layer (`src/utils/request.js`)

| Item | Implementation |
|---|---|
| baseURL | Dev `/api-backend`; prod `VITE_API_BASE_URL`, default `http://182.92.210.198:8081` |
| Timeout | 5000ms |
| Fixed header | `platform: H5` |
| Token | `useUserStore().token` → `Authorization: Bearer ...` |
| Loading | Global `ElLoading`, shared across concurrent requests |
| Success | `code === 200 || 201`, otherwise `ElMessage.error` + `reject` |
| Error sanitising | Regex strips `json/sql/mysql/dial tcp/panic/...`, showing "操作失败，请稍后重试" |
| `silent` | `config.silent === true` suppresses toasts (expected empties such as "no resume") |
| 401 | **No special handling** — no redirect or refresh (the route guard handles expiry locally) |
| `fileUrl(path)` | Backslashes → `/`; relative paths get `API_BASE_URL` |

### 21.2 State management

| Store | id | Contents | Persisted |
|---|---|---|---|
| `stores/user.js` | `rec-user` | `userId/username/displayName/userType/token` | ✅ localStorage |
| `stores/students.js` | `filter` | Degree/major/gender/language/job/skill/city filters | ✅ |
| `stores/resume.js` | `resume` | Resume draft | ✅ (unused) |
| `stores/chatUnread.js` | module-level `ref` | Global unread total | ❌ |

### 21.3 Routing and permissions

- `createWebHistory`, all pages lazy-loaded.
- Whitelist `/login`; everything else requires a token and a locally validated, unexpired JWT.
- **No role-based route guards**; differences appear only in menus and in-page conditionals:
  - Company: `/home`, `/job`;
  - Student/company: `/chat`;
  - Everyone: `/showData`.

### 21.4 ECharts dashboard

- On-demand registration of `CanvasRenderer`, `PieChart/BarChart/LineChart/RadarChart`, `Title/Tooltip/Legend/Grid`.
- Data from `/stu/statistics`, top 10 entries.
- `shallowRef` + `watchEffect` reduce reactivity overhead.
- A `MutationObserver` watches `<html class="dark">` to switch colours.

### 21.5 Resume template editor

`makeResume.vue` is **client-only**: editable fields, add/remove education/work/project/award entries, avatar upload (≤2MB, local preview), and `html2pdf.js` export of `${name}_简历.pdf` (`scale 1.8`, control panel excluded).

### 21.6 Admin panel

- No `src/api`, no axios wrapper — every request lives in `views/Home.vue`.
- `backendUrl = VITE_API_BASE_URL || 'http://81.70.199.174:8081'`.
- Token read from `localStorage['token']`, `Authorization` assembled manually.
- Features: student/company lists, pagination, search, bulk delete, Excel import; export, statistics, settings, and the edit drawer are placeholders.

---

## 22. Deployment

### 22.1 Windows local

See [6.1 One-click start](#61-one-click-start-windows-recommended).

### 22.2 Linux + systemd (Go)

```bash
cd back/resume-go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o resume-go .
# upload resume-go, schema.sql, and uploads/ to /opt/resume-go
```

Example `/etc/systemd/system/resume-go.service`:

```ini
[Unit]
Description=Resume Go Backend
After=network.target mysql.service redis.service rabbitmq-server.service

[Service]
WorkingDirectory=/opt/resume-go
EnvironmentFile=/opt/resume-go/.env
ExecStart=/opt/resume-go/resume-go
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Put `SERVER_PORT`, `DB_*`, `JWT_SECRET`, `REDIS_*`, `RABBITMQ_*`, `GRPC_*` in `.env` (see [7.1](#71-go--environment-variables)).

```bash
systemctl daemon-reload && systemctl enable --now resume-go
curl http://127.0.0.1:8081/stu/statistics   # health check
```

### 22.3 Docker Compose (Go)

`docker-compose.yml` defines mysql / redis / rabbitmq / app; `app` waits for all three to be `service_healthy`; `uploads` is a volume; every setting comes from environment variables.

### 22.4 Kubernetes (Go)

| Resource | Notes |
|---|---|
| `Namespace recruit` | Namespace |
| `ConfigMap resume-go-config` | Ports, DB/Redis/MQ addresses, `UPLOAD_DIR`, `TZ` |
| `Secret resume-go-secret` | `DB_USER`, `DB_PASSWORD`, `JWT_SECRET` |
| `Deployment resume-go` | 1 replica, RollingUpdate (`maxSurge 1` / `maxUnavailable 0`), ports 8081/9091, PVC at `/app/uploads`, readiness/liveness on `GET /stu/statistics`, `100m/128Mi`–`500m/512Mi` |
| `Service resume-go` | ClusterIP, 8081 http / 9091 grpc |
| `PVC resume-uploads-pvc` | `ReadWriteOnce`, 5Gi (scale-out needs object storage/NAS) |
| `Ingress` | `ingressClassName: nginx`, host `recruit.example.com`, `proxy-body-size: 12m` |

> Readiness probes hit the public `/stu/statistics`, so that endpoint must stay unauthenticated.

### 22.5 Nginx WebSocket proxy

```nginx
location /chat/ws {
    proxy_pass http://127.0.0.1:8081;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 3600s;
}
```

---

## 23. Known Issues & Caveats

> Everything below is observed in the code/SQL — not speculation.

### 23.1 Security

1. **Hard-coded secrets**: the JWT secret and DB password default to values in `config.go`; externalise them.
2. **Refresh token usable as access token**: `tokenType` is not enforced.
3. **Admin token passed in the URL** (`/admin/?token=...`) — leaks into history/logs; the admin app never validates it.
4. **CORS echoes any Origin** — restrict to an allow-list in production.
5. **WebSocket `CheckOrigin` always true**.

### 23.2 Functional / consistency

1. **`/api/resume/download/{name}` does not exist** — the upload endpoint returns it, but direct access 404s; use `/uploads/resume/<file>`.
2. **No resume parsing**: only type/size validation and storage; no PDF text extraction.
3. **`/stu/getAllSkill` returns a string, not an array**; clients must `split(',')`.
4. **Deleting students/companies leaves orphan rows** in `stu_*`, `stu_resume`, `com_like_stu`.
5. **Sample data inconsistencies**: `user_id=2` has a plaintext password (cannot pass BCrypt); `student.id=5` does not match some child rows (prefer `seed_accounts.sql`).

### 23.3 Front end

1. Menu index `/Job` vs route `/job` — **case mismatch breaks the "Jobs" menu**.
2. `myJobs.vue` delete/status methods are **neither imported nor bound** (dead code).
3. `makeResume.vue` preview/save/upload buttons have no handlers; only download works.
4. Admin export/statistics/settings/edit drawer are placeholders.
5. Both Vite apps default to 5173 (the batch script assigns 5173/5174).
6. Portal axios timeout is 5s (short for uploads) and has no unified 401 handling.

### 23.4 Operations

1. Chat presence TTL is set only at registration; long connections lose it after 5 minutes.
2. No cross-instance real-time push.
3. MQ consumption drops failures; no DLQ.
4. `mq_consumed_event` only grows — archive periodically.
5. K8s uses a `ReadWriteOnce` PVC with 1 replica; scaling needs object storage.

---

## 24. Development Guide

### 24.1 Add a Go endpoint

1. Define request/response structs in `internal/model` (implement `UnmarshalJSON` for compatibility).
2. Write SQL in `internal/data` (always bind with `?`; wrap multi-statement writes in `BeginTx`).
3. Add business rules in `internal/service`, returning `model.AppError{Code, Msg}`.
4. Add the handler in `internal/api` and register it under `authorized` or `admin`.
5. Add a function in `src/api/*.js` and call it from the page.

### 24.2 Add a front-end page

1. Create the `.vue` file under `views/`.
2. Add a lazy route in `router/index.js`.
3. Authentication is automatic (global guard).
4. To add a menu entry, edit `views/layout/index.vue` — **keep `index` casing identical to the route path**.

### 24.3 Add a database column

1. Add it in `back/resume.sql` (base tables) or `back/resume-go/schema.sql` (extended tables).
2. Update the `internal/model` struct and the SQL in `internal/data`.
3. For profile child tables, also update the delete-then-insert logic and the detail aggregation.

### 24.4 Troubleshooting

| Symptom | Check |
|---|---|
| 401 | Is the `Bearer ` prefix present? Has the token expired (`exp`)? |
| 403 | Admin-only route? Editing another company's job? |
| Upload fails | Field name must be `resumeFile` / `image` / `stuFile` / `comFile`; size and extension |
| No chat messages | Was the WS handshake 101? Valid token? Is the peer online? Does `/chat/messages` return history? |
| Duplicate chat bubbles | De-duplicate by message `id` |
| Wrong page total | Use `total`, not `size` |
| File 404 | Use `/uploads/...`, not `/api/resume/download/...` |

---

## 25. Appendix: Endpoint Cheat Sheet

### 25.1 By module

| Module | Endpoints | Prefixes |
|---|---|---|
| Auth/user | 5 | `/login`, `/refresh-token`, `/validateToken`, `/user/*` |
| Student | 9 | `/stu/*` |
| Jobs | 9 | `/com/jobPosting*`, `/com/applications*` |
| Bookmarks | 6 | `/com/Like*`, `/com/QueryComIdByUserId` |
| Search | 2 (+1 admin reuse) | `/com/search`, `/com/detailSearch` |
| Chat | 12 | `/chat/*` |
| Admin | 19 | `/admin/*` |
| Static | 3 | `/uploads/*`, `/resume/uploads/*` |

### 25.2 Public endpoints

```text
POST /login
POST /refresh-token
POST /validateToken
GET  /stu/statistics
GET  /chat/ws?token=...
```

### 25.3 Admin only

```text
/admin/**
```

### 25.4 Chat / application / user-centre endpoints

```text
/user/me, PUT /user/name
GET  /stu/resume/latest
/stu/applications*
/com/applications*
/chat/**
```

### 25.5 Source index

| Capability | File |
|---|---|
| Go route table | `back/resume-go/internal/api/api.go` |
| Go auth | `internal/auth/auth.go`, `internal/api/middleware.go` |
| Chat Hub | `internal/chat/hub.go` |
| Chat connection | `internal/chat/client.go` |
| Chat business rules | `internal/service/chat.go` |
| Chat SQL | `internal/data/chat.go` |
| Chat entry points | `internal/api/handlers_chat.go` |
| Cache | `internal/cache/cache.go` |
| MQ | `internal/mq/mq.go` |
| gRPC | `internal/rpc/server.go`, `client.go`, `proto/recruit/v1/recruit.proto` |
| Portal HTTP wrapper | `front/yxl/recruit-project/src/utils/request.js` |
| Portal WS wrapper | `front/yxl/recruit-project/src/utils/chatSocket.js` |
| Portal chat page | `front/yxl/recruit-project/src/views/chat/index.vue` |
| Portal route guard | `front/yxl/recruit-project/src/router/index.js` |
| Admin page | `front/ljk/adminPanel/src/views/Home.vue` |
| Chat tutorial | `CHAT_GUIDE.md` |

---

**This document covers every technology, endpoint, chat mechanism, and deployment path in the project.** For chat specifically, start at Chapter 12 and read it alongside `CHAT_GUIDE.md`.

