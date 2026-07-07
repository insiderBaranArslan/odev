package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"banking-api/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return user, err
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET username = $1, email = $2, role = $3, updated_at = NOW()
		WHERE id = $4 RETURNING updated_at`
	return r.pool.QueryRow(ctx, query, user.Username, user.Email, user.Role, user.ID).Scan(&user.UpdatedAt)
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

type PostgresTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTransactionRepository(pool *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{pool: pool}
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, tx.FromUserID, tx.ToUserID, tx.Amount, tx.Type, tx.Status).
		Scan(&tx.ID, &tx.CreatedAt)
}

func (r *PostgresTransactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	tx := &domain.Transaction{}
	query := `SELECT id, from_user_id, to_user_id, amount, type, status, created_at FROM transactions WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("transaction not found")
	}
	return tx, err
}

func (r *PostgresTransactionRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Transaction, error) {
	query := `
		SELECT id, from_user_id, to_user_id, amount, type, status, created_at
		FROM transactions
		WHERE from_user_id = $1 OR to_user_id = $1
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}

func (r *PostgresTransactionRepository) UpdateStatus(ctx context.Context, id int64, status domain.TransactionStatus) error {
	tag, err := r.pool.Exec(ctx, `UPDATE transactions SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found")
	}
	return nil
}

func (r *PostgresTransactionRepository) CountByStatus(ctx context.Context, status domain.TransactionStatus) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE status = $1`, status).Scan(&count)
	return count, err
}

type PostgresBalanceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBalanceRepository(pool *pgxpool.Pool) *PostgresBalanceRepository {
	return &PostgresBalanceRepository{pool: pool}
}

func (r *PostgresBalanceRepository) Get(ctx context.Context, userID int64) (*domain.Balance, error) {
	balance := &domain.Balance{UserID: userID}
	err := r.pool.QueryRow(ctx, `SELECT amount, last_updated_at FROM balances WHERE user_id = $1`, userID).
		Scan(&balance.Amount, &balance.LastUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &domain.Balance{UserID: userID, Amount: 0, LastUpdatedAt: time.Now().UTC()}, nil
	}
	return balance, err
}

func (r *PostgresBalanceRepository) Upsert(ctx context.Context, balance *domain.Balance) error {
	query := `
		INSERT INTO balances (user_id, amount, last_updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE SET amount = EXCLUDED.amount, last_updated_at = NOW()
		RETURNING last_updated_at`
	return r.pool.QueryRow(ctx, query, balance.UserID, balance.Amount).Scan(&balance.LastUpdatedAt)
}

func (r *PostgresBalanceRepository) UpdateAmount(ctx context.Context, userID int64, delta float64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO balances (user_id, amount, last_updated_at)
		VALUES ($1, 0, NOW())
		ON CONFLICT (user_id) DO NOTHING`, userID)
	if err != nil {
		return err
	}

	var amount float64
	err = tx.QueryRow(ctx, `
		UPDATE balances
		SET amount = amount + $2, last_updated_at = NOW()
		WHERE user_id = $1 AND amount + $2 >= 0
		RETURNING amount`, userID, delta).Scan(&amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("insufficient balance")
		}
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO balance_history (user_id, amount, recorded_at)
		VALUES ($1, $2, NOW())`, userID, amount); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresBalanceRepository) ListHistorical(ctx context.Context, userID int64) ([]domain.BalanceSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, amount, recorded_at FROM balance_history
		WHERE user_id = $1 ORDER BY recorded_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []domain.BalanceSnapshot
	for rows.Next() {
		var snap domain.BalanceSnapshot
		if err := rows.Scan(&snap.UserID, &snap.Amount, &snap.RecordedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}
	return snapshots, rows.Err()
}

func (r *PostgresBalanceRepository) GetAtTime(ctx context.Context, userID int64, at time.Time) (*domain.BalanceSnapshot, error) {
	snap := &domain.BalanceSnapshot{}
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, amount, recorded_at FROM balance_history
		WHERE user_id = $1 AND recorded_at <= $2
		ORDER BY recorded_at DESC LIMIT 1`, userID, at).
		Scan(&snap.UserID, &snap.Amount, &snap.RecordedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &domain.BalanceSnapshot{UserID: userID, Amount: 0, RecordedAt: at}, nil
	}
	return snap, err
}

type PostgresAuditRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAuditRepository(pool *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{pool: pool}
}

func (r *PostgresAuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if log.Details == nil {
		log.Details = json.RawMessage(`{}`)
	}
	query := `
		INSERT INTO audit_logs (entity_type, entity_id, action, details, created_at)
		VALUES ($1, $2, $3, $4, NOW()) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, log.EntityType, log.EntityID, log.Action, log.Details).
		Scan(&log.ID, &log.CreatedAt)
}

type PostgresEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEventRepository(pool *pgxpool.Pool) *PostgresEventRepository {
	return &PostgresEventRepository{pool: pool}
}

func (r *PostgresEventRepository) Append(ctx context.Context, event *domain.Event) error {
	if event.Payload == nil {
		event.Payload = json.RawMessage(`{}`)
	}
	query := `
		INSERT INTO events (aggregate, aggregate_id, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, NOW()) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, event.Aggregate, event.AggregateID, event.EventType, event.Payload).
		Scan(&event.ID, &event.CreatedAt)
}

func (r *PostgresEventRepository) ListByAggregate(ctx context.Context, aggregate string, aggregateID int64) ([]domain.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, aggregate, aggregate_id, event_type, payload, created_at
		FROM events WHERE aggregate = $1 AND aggregate_id = $2 ORDER BY created_at ASC`,
		aggregate, aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.ID, &event.Aggregate, &event.AggregateID, &event.EventType, &event.Payload, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
