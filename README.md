# Go Banking

A production-ready **banking backend** written in Go. It provides user management, account operations, fund transfers, email verification, and role-based access control. The API is exposed via **gRPC** and **REST** (gRPC-Gateway), with async background jobs for email delivery and full **Kubernetes** deployment on AWS EKS.

---

## Features

### User Management
- **Create user** — Registration with username, full name, email, and hashed password (bcrypt)
- **Login** — Returns access and refresh tokens (PASETO) and creates a session
- **Update user** — Protected endpoint; users can update their own profile; bankers can update any user (full name, email, password)
- **Email verification** — Post-registration verification via secret link; verification emails are sent asynchronously via Redis/Asynq

### Authentication & Authorization
- **PASETO** (v2) tokens for access (short-lived) and refresh (long-lived)
- **Sessions** — Stored in PostgreSQL (refresh token, user agent, client IP, expiry)
- **Role-based access** — `depositor` and `banker` roles; permission checks on protected RPCs
- **Metadata extraction** — User-Agent and Client IP from gRPC metadata / HTTP headers for logging and session tracking

### Banking (Data Model & API)
- **Accounts** — Owner, balance, currency; unique `(owner, currency)` per user
- **Entries** — Ledger entries (credits/debits) linked to accounts
- **Transfers** — Between accounts with amount validation
- **Legacy REST API** — Gin-based HTTP API in `api/` (accounts, transfers, auth) with unit tests; currently not wired in `main` (gRPC + Gateway are used instead)

### Infrastructure & DevOps
- **Docker Compose** — PostgreSQL 16, Redis 7, and the Go API; `wait-for` script ensures DB is ready before startup
- **Migrations** — `golang-migrate` for schema versioning
- **Background tasks** — **Asynq** + Redis for “send verify email” job (retries, queue priority)
- **Email** — Gmail SMTP sender for verification emails
- **API docs** — **Swagger UI** served at `/swagger/` (generated from Protobuf + gRPC-Gateway)
- **CORS** — Configurable allowed origins (e.g. `localhost:3000`, `https://golang-banking.xyz`)
- **Graceful shutdown** — `errgroup`-based orchestration for gRPC server, HTTP gateway, and task processor

### Deployment
- **Kubernetes (EKS)** — Deployment, Service, HTTP and gRPC Ingress (separate hosts), NGINX Ingress Controller
- **TLS** — **cert-manager** + **Let’s Encrypt** (ACME) with DNS-01 (Route53)
- **Secrets** — AWS Secrets Manager; values loaded into `app.env` and mounted as Kubernetes Secrets
- **CI/CD** — GitHub Actions: **test** workflow (Go tests + Postgres + migrations) and **deploy** workflow (build → ECR → EKS apply)

---

## Tech Stack

| Category        | Technology                                          |
|----------------|------------------------------------------------------|
| Language       | Go 1.25                                             |
| API            | gRPC, gRPC-Gateway (REST), Protobuf                 |
| Database       | PostgreSQL 16, **pgx/v5**, **sqlc**                 |
| Migrations     | **golang-migrate**                                  |
| Auth           | **PASETO** v2 (symmetric), bcrypt                   |
| Config         | **Viper** (env-based)                               |
| Logging        | **zerolog**                                         |
| Validation     | Custom **val** package + gRPC `BadRequest` details  |
| Background Jobs| **Asynq** + **Redis** 7                             |
| Email          | **jordan-wright/email** (Gmail SMTP)                |
| API Docs       | **Swagger** (OpenAPI v2) via **statik**             |
| CORS           | **rs/cors**                                         |
| Containers     | **Docker**, **Docker Compose**                      |
| Orchestration  | **Kubernetes** (EKS), **kubectl**                   |
| Ingress / TLS  | **NGINX Ingress**, **cert-manager**, **Let’s Encrypt** |
| CI/CD          | **GitHub Actions**                                  |
| Cloud          | **AWS** (ECR, EKS, Secrets Manager, Route53)        |

---

## Project Structure

```
.
├── api/                    # Legacy Gin REST API (accounts, transfers, auth)
├── db/
│   ├── migration/          # SQL migrations (golang-migrate)
│   ├── query/              # sqlc queries (accounts, entries, transfers, users, sessions, verify_email)
│   ├── sqlc/               # Generated Go code (sqlc)
│   └── mock/               # Mock store (mockgen)
├── doc/
│   ├── db.dbml             # DBML schema (optional doc)
│   ├── schema.sql          # Generated from DBML
│   ├── statik/             # Embedded Swagger UI (statik)
│   └── swagger/            # OpenAPI JSON + Swagger assets
├── eks/                    # Kubernetes manifests (EKS)
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress-http.yaml   # HTTP at api.golang-banking.xyz
│   ├── ingress-grpc.yaml   # gRPC at gapi.golang-banking.xyz
│   ├── ingress-nginx.yaml
│   ├── issuer.yaml         # cert-manager Let's Encrypt
│   └── aws-auth.yaml
├── gapi/                   # gRPC API implementation
│   ├── server.go
│   ├── rpc_create_user.go
│   ├── rpc_login_user.go
│   ├── rpc_update_user.go
│   ├── rpc_verify_email.go
│   ├── authorization.go
│   ├── metadata.go
│   ├── converter.go
│   ├── error.go
│   └── logger.go
├── mail/                   # Email sender (Gmail SMTP)
├── pb/                     # Generated Protobuf / gRPC / Gateway code
├── proto/                  # .proto definitions + Google API annotations
├── token/                  # PASETO token maker
├── util/                   # Config, password, random, roles, currency
├── val/                    # Validation helpers
├── worker/                 # Asynq task distributor + processor
├── main.go                 # Entrypoint: migrations, gRPC, HTTP gateway, task processor
├── Dockerfile
├── docker-compose.yaml
├── Makefile
├── app.env                 # Config template (see below)
├── start.sh                # Container entrypoint helper
├── wait-for.sh             # Wait for Postgres before starting app
└── sqlc.yaml               # sqlc config
```

---

## API Overview

### gRPC + REST (gRPC-Gateway)

| RPC            | HTTP Method | Path              | Auth   | Description                |
|----------------|-------------|-------------------|--------|----------------------------|
| `CreateUser`   | `POST`      | `/v1/create_user` | No     | Register; enqueues verify email |
| `LoginUser`    | `POST`      | `/v1/login_user`  | No     | Login; returns tokens + session |
| `UpdateUser`   | `PATCH`     | `/v1/update_user` | Bearer | Update profile (RBAC)      |
| `VerifyEmail`  | `GET`       | `/v1/verify_email`| No     | Verify email via link      |

- **gRPC** default: `0.0.0.0:9090`
- **HTTP** (Gateway + Swagger): `0.0.0.0:8080`
- **Swagger UI**: `http://localhost:8080/swagger/` (or your HTTP host)

Protected RPCs expect `Authorization: Bearer <access_token>`.

---

## Database Schema

- **users** — username (PK), role, hashed_password, full_name, email, is_email_verified, password_changed_at, created_at  
- **sessions** — id (UUID), username, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at  
- **verify_emails** — id, username, email, secret_code, is_used, created_at, expired_at  
- **accounts** — id, owner → users, balance, currency, created_at; unique (owner, currency)  
- **entries** — id, account_id → accounts, amount, created_at  
- **transfers** — id, from_account_id, to_account_id, amount, created_at  

Migrations live in `db/migration/` (`000001_init_schema` through `000005_add_role_to_users`).

---

## Configuration

Config is loaded from `app.env` (Viper) with `AutomaticEnv()`. Required variables:

| Variable               | Description                                  |
|------------------------|----------------------------------------------|
| `ENVIRONMENT`          | e.g. `development` / `production`            |
| `ALLOWED_ORIGINS`      | CORS origins (comma-separated)               |
| `DB_SOURCE`            | PostgreSQL DSN                               |
| `MIGRATION_URL`        | e.g. `file://db/migration`                   |
| `HTTP_SERVER_ADDRESS`  | e.g. `0.0.0.0:8080`                          |
| `GRPC_SERVER_ADDRESS`  | e.g. `0.0.0.0:9090`                          |
| `REDIS_ADDRESS`        | Redis host:port                              |
| `TOKEN_SYMMETRIC_KEY`  | 32-byte key for PASETO                       |
| `ACCESS_TOKEN_DURATION`| e.g. `15m`                                   |
| `REFRESH_TOKEN_DURATION` | e.g. `24h`                                |
| `EMAIL_SENDER_NAME`    | Sender display name                          |
| `EMAIL_SENDER_ADDRESS` | Gmail address                                |
| `EMAIL_SENDER_PASSWORD`| Gmail app password                           |

Use strong secrets and avoid committing real credentials. For EKS, these are stored in AWS Secrets Manager and injected via Kubernetes Secrets.

---

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [sqlc](https://sqlc.dev/) (for codegen)
- [protoc](https://grpc.io/docs/protoc-installation/) + Go plugins (for proto generation)
- [statik](https://github.com/rakyll/statik) (for embedding Swagger)

### Run with Docker Compose

1. Copy and adjust `app.env` (DB, Redis, tokens, email).
2. From project root:

   ```bash
   docker-compose up --build
   ```

   The API container uses `wait-for.sh` to wait for Postgres before running the app. HTTP server: `http://localhost:8080`, gRPC: `localhost:9090`, Swagger: `http://localhost:8080/swagger/`.

### Run locally (without Docker)

1. Start Postgres and Redis (e.g. via Makefile or manually):

   ```bash
   make postgres   # optional: make createdb
   make redis
   ```

2. Set `DB_SOURCE`, `REDIS_ADDRESS`, etc. (e.g. in `app.env`).
3. Run migrations:

   ```bash
   make migrateup
   ```

4. Start the server:

   ```bash
   make server
   ```

---

## Makefile Commands

| Command         | Description                                  |
|-----------------|----------------------------------------------|
| `make postgres` | Run Postgres 16 container                    |
| `make createdb` | Create `simple_bank` DB                      |
| `make dropdb`   | Drop `simple_bank` DB                        |
| `make new_migration name=...` | Create new migration              |
| `make migrateup`| Run all migrations up                        |
| `make migrateup1` | Run next migration up                      |
| `make migratedown` / `migratedown1` | Migrate down         |
| `make sqlc`     | Generate Go from `db/query`                  |
| `make db_schema`| Generate `doc/schema.sql` from DBML          |
| `make proto`    | Generate Go + Swagger from `proto/`, refresh statik |
| `make test`     | Run tests                                    |
| `make server`   | `go run main.go`                             |
| `make mock`     | (Re)generate mocks (store, task distributor) |
| `make redis`    | Run Redis 7 container                        |

---

## Testing

- Unit and integration tests live alongside packages (`*_test.go`).
- GitHub Actions **test** workflow:
  - Spins up Postgres 16
  - Runs `make migrateup` then `make test`.

Run locally:

```bash
make migrateup
make test
```

---

## Deployment (EKS)

1. **Secrets**: Store config in AWS Secrets Manager (e.g. secret `go_banking`). The deploy workflow reads it and writes `app.env`.
2. **ECR**: Repository `go-banking` in the deploy account/region.
3. **EKS**: Cluster `go-banking-cluster` in `eu-central-1`; `eks/aws-auth.yaml` applied as needed.
4. **Deploy workflow** (on push to `main`):
   - Configures AWS and kubectl.
   - Builds and pushes Docker image to ECR.
   - Creates/updates Kubernetes Secret `go-banking-secrets` from `app.env`.
   - Applies `eks/` manifests (deployment, service, ingresses, issuer).

**Endpoints** (when DNS and Ingress are configured):

- HTTP: `https://api.golang-banking.xyz`
- gRPC: `https://gapi.golang-banking.xyz`

TLS is managed by cert-manager and Let’s Encrypt (Route53 DNS-01).

---

## License

Application code: unspecified. The `wait-for.sh` script is under the [MIT License](https://opensource.org/licenses/MIT) (see script header).

---

## Contact

- **Go Simple Bank** — [GitHub](https://github.com/chuuch) · [email](mailto:daniel.chuchulev96@gmail.com)
