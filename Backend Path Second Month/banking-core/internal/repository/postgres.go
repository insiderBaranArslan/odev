package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"banking-core/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct{ pool *pgxpool.Pool }

func NewUserRepo(pool *pgxpool.Pool) *UserRepo { return &UserRepo{pool: pool} }

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`,
		user.Username, user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	return r.pool.QueryRow(ctx, `
		UPDATE users SET username=$1, email=$2, role=$3, updated_at=NOW()
		WHERE id=$4 RETURNING updated_at`,
		user.Username, user.Email, user.Role, user.ID,
	).Scan(&user.UpdatedAt)
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

type TransactionRepo struct{ pool *pgxpool.Pool }

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pool: pool}
}

func (r *TransactionRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`,
		tx.FromUserID, tx.ToUserID, tx.Amount, tx.Type, tx.Status,
	).Scan(&tx.ID, &tx.CreatedAt)
}

func (r *TransactionRepo) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	tx := &domain.Transaction{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, from_user_id, to_user_id, amount, type, status, created_at
		FROM transactions WHERE id=$1`, id).
		Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("transaction not found")
	}
	return tx, err
}

func (r *TransactionRepo) ListByUser(ctx context.Context, userID int64) ([]domain.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, from_user_id, to_user_id, amount, type, status, created_at
		FROM transactions
		WHERE from_user_id=$1 OR to_user_id=$1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(&tx.ID, &tx.FromUserID, &tx.ToUserID, &tx.Amount, &tx.Type, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, tx)
	}
	return list, rows.Err()
}

func (r *TransactionRepo) UpdateStatus(ctx context.Context, id int64, status domain.TransactionStatus) error {
	tag, err := r.pool.Exec(ctx, `UPDATE transactions SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found")
	}
	return nil
}

func (r *TransactionRepo) CountByStatus(ctx context.Context, status domain.TransactionStatus) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE status=$1`, status).Scan(&n)
	return n, err
}

type BalanceRepo struct{ pool *pgxpool.Pool }

func NewBalanceRepo(pool *pgxpool.Pool) *BalanceRepo { return &BalanceRepo{pool: pool} }

func (r *BalanceRepo) Get(ctx context.Context, userID int64) (*domain.Balance, error) {
	b := domain.NewBalance(userID, 0)
	err := r.pool.QueryRow(ctx, `
		SELECT amount, last_updated_at FROM balances WHERE user_id=$1`, userID).
		Scan(&b.Amount, &b.LastUpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NewBalance(userID, 0), nil
	}
	return b, err
}

func (r *BalanceRepo) Upsert(ctx context.Context, balance *domain.Balance) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO balances (user_id, amount, last_updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET amount=EXCLUDED.amount, last_updated_at=NOW()
		RETURNING last_updated_at`,
		balance.UserID, balance.Amount,
	).Scan(&balance.LastUpdatedAt)
}

func (r *BalanceRepo) UpdateAmount(ctx context.Context, userID int64, delta float64) (float64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO balances (user_id, amount, last_updated_at)
		VALUES ($1, 0, NOW())
		ON CONFLICT (user_id) DO NOTHING`, userID); err != nil {
		return 0, err
	}

	var amount float64
	err = tx.QueryRow(ctx, `
		UPDATE balances
		SET amount = amount + $2, last_updated_at = NOW()
		WHERE user_id = $1 AND amount + $2 >= 0
		RETURNING amount`, userID, delta).Scan(&amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("insufficient balance")
		}
		return 0, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO balance_history (user_id, amount, recorded_at)
		VALUES ($1, $2, NOW())`, userID, amount); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return amount, nil
}

func (r *BalanceRepo) ListHistorical(ctx context.Context, userID int64) ([]domain.BalanceSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, amount, recorded_at
		FROM balance_history WHERE user_id=$1
		ORDER BY recorded_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snaps []domain.BalanceSnapshot
	for rows.Next() {
		var s domain.BalanceSnapshot
		if err := rows.Scan(&s.UserID, &s.Amount, &s.RecordedAt); err != nil {
			return nil, err
		}
		snaps = append(snaps, s)
	}
	return snaps, rows.Err()
}

func (r *BalanceRepo) GetAtTime(ctx context.Context, userID int64, at time.Time) (*domain.BalanceSnapshot, error) {
	s := &domain.BalanceSnapshot{}
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, amount, recorded_at FROM balance_history
		WHERE user_id=$1 AND recorded_at <= $2
		ORDER BY recorded_at DESC LIMIT 1`, userID, at).
		Scan(&s.UserID, &s.Amount, &s.RecordedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &domain.BalanceSnapshot{UserID: userID, Amount: 0, RecordedAt: at}, nil
	}
	return s, err
}

type AuditRepo struct{ pool *pgxpool.Pool }

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo { return &AuditRepo{pool: pool} }

func (r *AuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	if log.Details == nil {
		log.Details = json.RawMessage(`{}`)
	}
	return r.pool.QueryRow(ctx, `
		INSERT INTO audit_logs (entity_type, entity_id, action, details, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at`,
		log.EntityType, log.EntityID, log.Action, log.Details,
	).Scan(&log.ID, &log.CreatedAt)
}
