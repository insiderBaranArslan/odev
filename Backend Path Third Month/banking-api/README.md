# Banking API — Backend Path Third Month

Go banking service covering **sections 1–3**: Project Setup, Core Implementation, and **API Implementation**.

## Month Focus: API Layer

This month builds on Month 1–2 core and adds:

- Custom router (chi) with middleware stack
- CORS + security headers
- Rate limiting
- Request logging & request ID tracking
- Auth / role / validation / error / performance middleware
- Full REST API under `/api/v1`

## Structure

```
banking-api/
├── cmd/server/           # HTTP server + graceful shutdown
├── internal/
│   ├── config/           # Env configuration
│   ├── logger/           # slog
│   ├── domain/           # Models + interfaces
│   ├── database/         # Postgres + migrations
│   ├── repository/       # Data access
│   ├── service/          # Business logic
│   ├── worker/           # Concurrent transaction pool
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # Auth, CORS, rate limit, etc.
│   └── server/           # Router setup
├── migrations/
├── pkg/response/
├── Dockerfile
└── docker-compose.yml
```

## API Endpoints

### Auth
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

### Users (list/update/delete → admin)
- `GET /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`

### Transactions (credit → admin)
- `POST /api/v1/transactions/credit`
- `POST /api/v1/transactions/debit`
- `POST /api/v1/transactions/transfer`
- `GET /api/v1/transactions/history`
- `GET /api/v1/transactions/{id}`

### Balances
- `GET /api/v1/balances/current`
- `GET /api/v1/balances/historical?user_id=1`
- `GET /api/v1/balances/at-time?user_id=1&at=2026-01-01T00:00:00Z`

## Middleware (3.3)

| Middleware | File / location |
|------------|-----------------|
| Authentication | `middleware.Auth` |
| Role-based authorization | `middleware.RequireRole` |
| Request validation | request DTOs `Validate()` in handlers |
| Error handling | `Recoverer` + JSON error responses |
| Performance monitoring | `middleware.Performance` (slow request logs) |
| CORS / security / rate limit / logging | global stack in `server.NewRouter` |

## Run

```bash
cd "Backend Path Third Month/banking-api"
docker compose up --build
```

API: http://localhost:8080

Local (Postgres only on 5434):

```bash
docker compose up -d postgres
export DATABASE_URL="postgres://banking:banking@localhost:5434/banking?sslmode=disable"
go run ./cmd/server
```

## Example

```bash
# Register
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"password123"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}' | jq -r .access_token)

# Current balance
curl -s http://localhost:8080/api/v1/balances/current \
  -H "Authorization: Bearer $TOKEN"
```

## Notes

- Credit and user admin endpoints require `role=admin`.
- Debit/transfer are scoped to the authenticated user (unless admin).
- Core (worker pool, balances, services) comes from Month 2 patterns.
