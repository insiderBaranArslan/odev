package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
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
	UpdateAmount(ctx context.Context, userID int64, delta float64) error
	ListHistorical(ctx context.Context, userID int64) ([]BalanceSnapshot, error)
	GetAtTime(ctx context.Context, userID int64, at time.Time) (*BalanceSnapshot, error)
}

type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
}

type EventRepository interface {
	Append(ctx context.Context, event *Event) error
	ListByAggregate(ctx context.Context, aggregate string, aggregateID int64) ([]Event, error)
}

type UserService interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Authenticate(ctx context.Context, req LoginRequest) (*User, error)
	GetUser(ctx context.Context, id int64) (*User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, id int64, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id int64) error
	IssueTokens(ctx context.Context, user *User) (*AuthResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error)
}

type TransactionService interface {
	Credit(ctx context.Context, req CreditRequest) (*Transaction, error)
	Debit(ctx context.Context, req DebitRequest) (*Transaction, error)
	Transfer(ctx context.Context, req TransferRequest) (*Transaction, error)
	GetTransaction(ctx context.Context, id int64) (*Transaction, error)
	GetHistory(ctx context.Context, userID int64) ([]Transaction, error)
	Rollback(ctx context.Context, id int64) error
	Stats() TransactionStats
	Enqueue(job TransactionJob)
}

type BalanceService interface {
	GetCurrent(ctx context.Context, userID int64) (*Balance, error)
	GetHistorical(ctx context.Context, userID int64) ([]BalanceSnapshot, error)
	GetAtTime(ctx context.Context, userID int64, at time.Time) (*BalanceSnapshot, error)
}

type CacheStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
