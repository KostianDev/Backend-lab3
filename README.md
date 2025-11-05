# Backend Lab 3 – Income Tracking Service

This repository contains the third laboratory assignment for the "Технології серверного програмного забезпечення" course. The project is implemented in Go 1.25.3 with the Gin framework and PostgreSQL. Variant **3 (Облік доходів)** is selected, which introduces per-user income tracking and balance control.

## Contents
- [Features](#features)
- [Architecture Overview](#architecture-overview)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Configuration](#configuration)
  - [Running Locally](#running-locally)
  - [Running with Docker Compose](#running-with-docker-compose)
- [Project Structure](#project-structure)
- [Endpoints](#endpoints)
- [Testing](#testing)
- [Variant Justification](#variant-justification)
- [Render Deployment](#render-deployment)
- [Useful Commands](#useful-commands)

## Features
- Gin-based HTTP API with request validation and structured error handling.
- GORM ORM integration with PostgreSQL (runtime) and SQLite (tests).
- Authentication service creating users and default accounts with hashed passwords.
- User deletion endpoint with cascading cleanup of accounts, incomes, and expenses.
- Account service that credits incomes, debits expenses, enforces optional overdraft policy, and tracks balances in cents.
- Centralised error middleware translating domain errors to JSON envelopes.
- Migrations managed via dedicated package executed on startup.
- Comprehensive Go test suite covering services and HTTP handlers.

## Architecture Overview
```
src/
 ├─ cmd/app           # Application bootstrap
 ├─ internal/config   # Environment configuration loader
 ├─ internal/database # Database connection helper
 ├─ internal/models   # GORM entities (User, Account, Income, Expense)
 ├─ internal/migrations # Schema migrations executed at startup
 ├─ internal/storage  # Repositories and business services
 ├─ internal/services # Utilities (time provider abstraction)
 └─ internal/http     # Handlers, requests/DTOs, responses, middleware, router
```
The application starts from `src/cmd/app/main.go`, reading configuration, establishing the DB connection, running migrations, wiring services, and registering handlers via `router.New`.

## Getting Started

### Prerequisites
- Go 1.25.3
- Docker & Docker Compose

### Configuration
Two example environment files are provided:
- `.env.example` – configuration for running the Go binary locally (points to localhost PostgreSQL).
- `.env.docker.example` – configuration for Docker Compose deployment.

Copy one of the examples and adjust secrets/ports as needed:
```bash
cp .env.example .env
# or for Docker Compose
cp .env.docker.example .env
```

### Running Locally
1. Ensure PostgreSQL is running with credentials matching `.env`.
2. Install dependencies and build:
   ```bash
   go mod tidy
   go run ./src/cmd/app
   ```
3. The API listens on `HTTP_PORT` (default 8080); verify health:
   ```bash
   curl http://localhost:8080/healthz
   ```

### Running with Docker Compose
1. Copy `.env.docker.example` to `.env` and customise values.
2. Start the stack:
   ```bash
   docker compose up -d
   ```
3. Check service status:
   ```bash
   docker compose ps
   curl http://localhost:8080/healthz
   ```
4. Stop and remove containers:
   ```bash
   docker compose down
   # Add -v to remove the database volume if you need a clean start
   ```

## Project Structure
```
.
├─ Dockerfile
├─ docker-compose.yml
├─ .env.example
├─ .env.docker.example
└─ src/ ... (see Architecture Overview)
```

## Endpoints
| Method | Path                                  | Description                              |
|--------|---------------------------------------|------------------------------------------|
| GET    | `/healthz`                            | Liveness probe                           |
| POST   | `/api/v1/auth/register`               | Register a new user                      |
| POST   | `/api/v1/auth/login`                  | Authenticate user and return session token |
| DELETE | `/api/v1/auth/{userID}`               | Delete a user (requires email/password in body) |
| POST   | `/api/v1/accounts/{userID}/incomes`   | Credit an income to the user account     |
| POST   | `/api/v1/accounts/{userID}/expenses`  | Debit an expense from the user account   |
| GET    | `/api/v1/accounts/{userID}/balance`   | Retrieve current balance                 |
| GET    | `/api/v1/accounts/{userID}/incomes`   | List incomes (optional `limit` query)    |
| GET    | `/api/v1/accounts/{userID}/expenses`  | List expenses (optional `limit` query)   |

All payloads are documented in `internal/http/requests` and `internal/http/responses` packages.

Example login response:
```json
{
   "token": "<random-hex>",
   "user": {
      "id": 1,
      "email": "user@example.com",
      "default_currency": "USD"
   }
}
```

## Testing
Tests rely on SQLite and testify. Run the suite with:
```bash
go test ./...
```
Key coverage areas:
- `internal/storage/account_service_test.go`
- `internal/storage/auth_service_test.go`
- `internal/http/handlers/account_handler_test.go`

## Variant Justification
Group number modulo 3 equals 0, hence variant **3 - Облік доходів**. The implementation introduces per-user accounts (`models.Account`) and income tracking (`models.Income`), automatically debiting expenses while respecting an optional negative balance policy.

## Render Deployment
The service is deployed on [Render](https://render.com/) as a Web Service with an attached managed PostgreSQL instance. Supply the following environment variables in the Render dashboard:

| Variable | Suggested Value |
|----------|-----------------|
| `DATABASE_URL` | Copy the External Connection string from the Render PostgreSQL service (includes user, password, host, port, database, and `sslmode=require`). |
| `APP_NAME` | `backend-lab3` |
| `GIN_MODE` | `release` |
| `ALLOW_NEGATIVE_BALANCE` | `false` or `true` depending on desired policy |

> Render automatically exposes `PORT`; the application reads `HTTP_PORT`, so set `HTTP_PORT` to `$PORT` in the environment. Ensure the Render service build command runs `go build ./src/cmd/app` (or use this repository's Dockerfile) and the start command executes `./app`.

## Useful Commands
- `go mod tidy` - ensure dependencies are up to date.
- `docker compose logs -f app` - tail application logs.
- `docker compose exec db psql -U ${DB_USER} -d ${DB_NAME}` - connect to the PostgreSQL instance.
- `curl http://localhost:8080/api/v1/accounts/1/balance` - sample API call.
- `go test ./src/internal/storage -run TestAccountService` - run targeted tests.

---
Released under tag [`v3.0.0`](https://github.com/KostianDev/Backend-lab3/releases/tag/v3.0.0) for laboratory evaluation.
