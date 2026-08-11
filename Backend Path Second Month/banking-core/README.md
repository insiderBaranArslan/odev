# Banking Core — Backend Path Second Month

Go module covering **Project Setup** and **Core Implementation** (sections 1–2).

## Scope

### 1. Project Setup
- Go modules + layered package structure
- Environment-based configuration
- Structured logging with `slog`
- Graceful shutdown (SIGINT / SIGTERM)

### 2. Core Implementation
- Domain models: `User`, `Transaction`, `Balance` (+ validation / state / RWMutex)
- Repository & service interfaces
- JSON marshaling for models
- Worker pool + channel queue
- `sync.RWMutex` on balance updates
- Atomic counters for transaction stats
- Batch processor for concurrent batch flush
- `UserService` — register (bcrypt), authenticate, authorize roles
- `TransactionService` — credit / debit / transfer / rollback
- `BalanceService` — thread-safe updates + historical tracking

## Structure

```
banking-core/
├── cmd/app/                 # Entrypoint + demo flow + graceful shutdown
├── internal/
│   ├── config/              # Env config
│   ├── logger/              # slog setup
│   ├── domain/              # Models + interfaces
│   ├── database/            # PostgreSQL pool + migrations runner
│   ├── repository/          # Postgres repositories
│   ├── service/             # Business logic
│   └── worker/              # Concurrent transaction workers
├── migrations/              # SQL up/down
├── docker-compose.yml       # Postgres only (port 5433)
└── README.md
```

## Database Tables

| Table | Purpose |
|-------|---------|
| `users` | Accounts with role |
| `transactions` | Credit / debit / transfer |
| `balances` | Current balance |
| `balance_history` | Historical snapshots |
| `audit_logs` | Action audit trail |

## Run

### 1. Start Postgres

```bash
cd "Backend Path Second Month/banking-core"
docker compose up -d
```

### 2. Install deps & test

```bash
go mod tidy
go test ./...
```

### 3. Run demo app

```bash
export DATABASE_URL="postgres://banking:banking@localhost:5433/banking?sslmode=disable"
go run ./cmd/app
```

The app:
1. Loads config from env
2. Connects to DB and runs migrations
3. Starts worker pool
4. Runs a demo: register alice/bob → credit → transfer → print balances/stats
5. Waits for Ctrl+C and shuts down gracefully

## Notes

- Month 5 (`Backend Path First Month`) already includes full HTTP API + Docker stack.
- This month focuses on **core domain + concurrency + services**.
- API layer (Month path section 3) can be built on top of these services next.
