package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type TransactionType string

const (
	TypeCredit   TransactionType = "credit"
	TypeDebit    TransactionType = "debit"
	TypeTransfer TransactionType = "transfer"
)

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "pending"
	StatusProcessing TransactionStatus = "processing"
	StatusCompleted  TransactionStatus = "completed"
	StatusFailed     TransactionStatus = "failed"
	StatusRolledBack TransactionStatus = "rolled_back"
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
	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if !strings.Contains(u.Email, "@") || !strings.Contains(u.Email, ".") {
		return errors.New("invalid email")
	}
	if u.Role == "" {
		u.Role = RoleUser
	}
	if u.Role != RoleUser && u.Role != RoleAdmin {
		return errors.New("invalid role")
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
	mu         sync.Mutex        `json:"-"`
}

func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	switch t.Type {
	case TypeCredit:
		if t.ToUserID == 0 {
			return errors.New("to_user_id is required for credit")
		}
	case TypeDebit:
		if t.FromUserID == nil || *t.FromUserID == 0 {
			return errors.New("from_user_id is required for debit")
		}
		t.ToUserID = *t.FromUserID
	case TypeTransfer:
		if t.FromUserID == nil || *t.FromUserID == 0 || t.ToUserID == 0 {
			return errors.New("from_user_id and to_user_id are required for transfer")
		}
		if *t.FromUserID == t.ToUserID {
			return errors.New("cannot transfer to the same account")
		}
	default:
		return errors.New("invalid transaction type")
	}
	if t.Status == "" {
		t.Status = StatusPending
	}
	return nil
}

func (t *Transaction) SetStatus(status TransactionStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
}

func (t *Transaction) GetStatus() TransactionStatus {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Status
}

func (t *Transaction) MarkProcessing() { t.SetStatus(StatusProcessing) }
func (t *Transaction) MarkCompleted()  { t.SetStatus(StatusCompleted) }
func (t *Transaction) MarkFailed()     { t.SetStatus(StatusFailed) }
func (t *Transaction) MarkRolledBack() { t.SetStatus(StatusRolledBack) }

type Balance struct {
	UserID        int64     `json:"user_id"`
	Amount        float64   `json:"amount"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	mu            sync.RWMutex
}

func NewBalance(userID int64, amount float64) *Balance {
	return &Balance{UserID: userID, Amount: amount, LastUpdatedAt: time.Now().UTC()}
}

func (b *Balance) GetAmount() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Amount
}

func (b *Balance) Apply(delta float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.Amount+delta < 0 {
		return errors.New("insufficient balance")
	}
	b.Amount += delta
	b.LastUpdatedAt = time.Now().UTC()
	return nil
}

func (b *Balance) Snapshot() Balance {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return Balance{UserID: b.UserID, Amount: b.Amount, LastUpdatedAt: b.LastUpdatedAt}
}

func (b Balance) MarshalJSON() ([]byte, error) {
	type Alias struct {
		UserID        int64     `json:"user_id"`
		Amount        float64   `json:"amount"`
		LastUpdatedAt time.Time `json:"last_updated_at"`
	}
	return json.Marshal(Alias{UserID: b.UserID, Amount: b.Amount, LastUpdatedAt: b.LastUpdatedAt})
}

type BalanceSnapshot struct {
	UserID     int64     `json:"user_id"`
	Amount     float64   `json:"amount"`
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

func (r RegisterRequest) Validate() error {
	if len(strings.TrimSpace(r.Username)) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if !strings.Contains(r.Email, "@") {
		return errors.New("invalid email")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" || r.Password == "" {
		return errors.New("email and password are required")
	}
	return nil
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

func (r CreditRequest) Validate() error {
	if r.UserID <= 0 || r.Amount <= 0 {
		return errors.New("user_id and positive amount are required")
	}
	return nil
}

type DebitRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

func (r DebitRequest) Validate() error {
	if r.UserID <= 0 || r.Amount <= 0 {
		return errors.New("user_id and positive amount are required")
	}
	return nil
}

type TransferRequest struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

func (r TransferRequest) Validate() error {
	if r.FromUserID <= 0 || r.ToUserID <= 0 || r.Amount <= 0 {
		return errors.New("from_user_id, to_user_id and positive amount are required")
	}
	if r.FromUserID == r.ToUserID {
		return errors.New("cannot transfer to the same account")
	}
	return nil
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     Role   `json:"role"`
}

type TransactionJob struct {
	Transaction *Transaction
	Apply       func() error
	ResultCh    chan error
}
