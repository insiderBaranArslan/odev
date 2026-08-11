package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]User, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id int64) (*Transaction, error)
	ListByUser(ctx context.Context, userID int64) ([]Transaction, error)
	UpdateStatus(ctx context.Context, id int64, status TransactionStatus) error
	CountByStatus(ctx context.Context, status TransactionStatus) (int64, error)
}

type BalanceRepository interface {
	Get(ctx context.Context, userID int64) (*Balance, error)
	Upsert(ctx context.Context, balance *Balance) error
	UpdateAmount(ctx context.Context, userID int64, delta float64) (float64, error)
	ListHistorical(ctx context.Context, userID int64) ([]BalanceSnapshot, error)
	GetAtTime(ctx context.Context, userID int64, at time.Time) (*BalanceSnapshot, error)
}

type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
}

type UserService interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Authenticate(ctx context.Context, req LoginRequest) (*User, error)
	Authorize(user *User, roles ...Role) error
	GetByID(ctx context.Context, id int64) (*User, error)
}

type TransactionService interface {
	Credit(ctx context.Context, req CreditRequest) (*Transaction, error)
	Debit(ctx context.Context, req DebitRequest) (*Transaction, error)
	Transfer(ctx context.Context, req TransferRequest) (*Transaction, error)
	Rollback(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Transaction, error)
	History(ctx context.Context, userID int64) ([]Transaction, error)
	Stats() TransactionStats
}

type BalanceService interface {
	GetCurrent(ctx context.Context, userID int64) (*Balance, error)
	ApplyDelta(ctx context.Context, userID int64, delta float64) error
	GetHistorical(ctx context.Context, userID int64) ([]BalanceSnapshot, error)
	GetAtTime(ctx context.Context, userID int64, at time.Time) (*BalanceSnapshot, error)
}
