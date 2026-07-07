# Banking API

Go backend service for user management, balances, and concurrent transaction processing.

## Features

- Go modules with layered architecture (domain, repository, service, handler)
- Environment-based configuration
- Structured logging with `slog`
- Graceful shutdown
- PostgreSQL schema + migrations
- JWT authentication and role-based authorization
- Worker pool with channel-based transaction queue
- Thread-safe balance updates
- Audit logs and event store (event sourcing foundation)
- In-memory cache layer (Redis-ready via config)
- Prometheus metrics at `/metrics`
- Docker multi-stage build + docker-compose stack

## Project Structure

```
banking-api/
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── config/          # Environment configuration
│   ├── domain/          # Models and interfaces
│   ├── database/        # PostgreSQL connection
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # Auth, CORS, rate limit, logging
│   ├── worker/          # Concurrent transaction workers
│   ├── cache/           # Cache abstraction
│   └── server/          # Router setup
├── migrations/          # SQL migrations
├── deploy/              # Prometheus config
├── Dockerfile
└── docker-compose.yml
```

## Quick Start

### With Docker

```bash
docker compose up --build
```

API: `http://localhost:8080`

### Local Development

1. Start PostgreSQL and Redis (or use docker compose for infra only)
2. Copy env file:

```bash
cp .env.example .env
```

3. Run the app:

```bash
go run ./cmd/server
```

## API Endpoints

### Auth
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

### Users
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`

### Transactions
- `POST /api/v1/transactions/credit`
- `POST /api/v1/transactions/debit`
- `POST /api/v1/transactions/transfer`
- `GET /api/v1/transactions/history`
- `GET /api/v1/transactions/{id}`
- `GET /api/v1/transactions/stats`

### Balances
- `GET /api/v1/balances/current`
- `GET /api/v1/balances/historical?user_id=1`
- `GET /api/v1/balances/at-time?user_id=1&at=2026-01-01T00:00:00Z`

## Example Flow

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}'

# Use access_token from response for protected routes
curl http://localhost:8080/api/v1/balances/current \
  -H "Authorization: Bearer <access_token>"
```

## Monitoring

- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)
- App metrics: `http://localhost:8080/metrics`

## Database Tables

- `users`
- `transactions`
- `balances`
- `balance_history`
- `audit_logs`
- `events`

## Notes

- Admin role is required for user list/update/delete and credit operations.
- Credit/debit/transfer operations are processed through a worker pool.
- Rollback is supported for completed transactions via service layer.
- Event stream is stored in `events` table for replay/rebuild scenarios.
