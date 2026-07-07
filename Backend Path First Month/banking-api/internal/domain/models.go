package domain

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type TransactionType string

const (
	TransactionTypeCredit   TransactionType = "credit"
	TransactionTypeDebit    TransactionType = "debit"
	TransactionTypeTransfer TransactionType = "transfer"
)

type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("username is required")
	}
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if !strings.Contains(u.Email, "@") {
		return errors.New("invalid email")
	}
	if u.Role == "" {
		u.Role = RoleUser
	}
	return nil
}

type Transaction struct {
	ID         int64             `json:"id"`
	FromUserID *int64            `json:"from_user_id,omitempty"`
	ToUserID   int64             `json:"to_user_id"`
	Amount     float64           `json:"amount"`
	Type       TransactionType   `json:"type"`
	Status     TransactionStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
}

func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	switch t.Type {
	case TransactionTypeCredit:
		if t.ToUserID == 0 {
			return errors.New("to_user_id is required for credit")
		}
	case TransactionTypeDebit:
		if t.FromUserID == nil || *t.FromUserID == 0 {
			return errors.New("from_user_id is required for debit")
		}
	case TransactionTypeTransfer:
		if t.FromUserID == nil || *t.FromUserID == 0 || t.ToUserID == 0 {
			return errors.New("from_user_id and to_user_id are required for transfer")
		}
		if *t.FromUserID == t.ToUserID {
			return errors.New("cannot transfer to the same account")
		}
	default:
		return errors.New("invalid transaction type")
	}
	return nil
}

func (t *Transaction) MarkCompleted() {
	t.Status = TransactionStatusCompleted
}

func (t *Transaction) MarkFailed() {
	t.Status = TransactionStatusFailed
}

func (t *Transaction) MarkRolledBack() {
	t.Status = TransactionStatusRolledBack
}

type Balance struct {
	UserID        int64     `json:"user_id"`
	Amount        float64   `json:"amount"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type BalanceSnapshot struct {
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	RecordedAt time.Time `json:"recorded_at"`
}

type AuditLog struct {
	ID         int64           `json:"id"`
	EntityType string          `json:"entity_type"`
	EntityID   int64           `json:"entity_id"`
	Action     string          `json:"action"`
	Details    json.RawMessage `json:"details"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Event struct {
	ID         int64           `json:"id"`
	Aggregate  string          `json:"aggregate"`
	AggregateID int64          `json:"aggregate_id"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
}

type TransactionJob struct {
	Transaction *Transaction
	ResultCh    chan error
	Apply       func(context.Context, *Transaction) error
}

type TransactionStats struct {
	Processed int64 `json:"processed"`
	Failed    int64 `json:"failed"`
	Pending   int64 `json:"pending"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type CreditRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type DebitRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type TransferRequest struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
}
