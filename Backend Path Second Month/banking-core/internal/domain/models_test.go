package domain_test

import (
	"encoding/json"
	"testing"

	"banking-core/internal/domain"
)

func TestUserValidate(t *testing.T) {
	u := &domain.User{Username: "ab", Email: "bad"}
	if err := u.Validate(); err == nil {
		t.Fatal("expected validation error")
	}

	u = &domain.User{Username: "alice", Email: "alice@example.com"}
	if err := u.Validate(); err != nil {
		t.Fatal(err)
	}
	if u.Role != domain.RoleUser {
		t.Fatalf("expected default role user, got %s", u.Role)
	}
}

func TestTransactionStateManagement(t *testing.T) {
	from := int64(1)
	tx := &domain.Transaction{
		FromUserID: &from,
		ToUserID:   2,
		Amount:     10,
		Type:       domain.TypeTransfer,
		Status:     domain.StatusPending,
	}
	if err := tx.Validate(); err != nil {
		t.Fatal(err)
	}

	tx.MarkProcessing()
	if tx.GetStatus() != domain.StatusProcessing {
		t.Fatalf("expected processing, got %s", tx.GetStatus())
	}
	tx.MarkCompleted()
	if tx.GetStatus() != domain.StatusCompleted {
		t.Fatalf("expected completed")
	}
}

func TestBalanceThreadSafeOps(t *testing.T) {
	b := domain.NewBalance(1, 50)
	if err := b.Apply(25); err != nil {
		t.Fatal(err)
	}
	if b.GetAmount() != 75 {
		t.Fatalf("expected 75, got %v", b.GetAmount())
	}
	if err := b.Apply(-100); err == nil {
		t.Fatal("expected insufficient balance")
	}

	data, err := json.Marshal(b.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	var decoded domain.Balance
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Amount != 75 {
		t.Fatalf("json roundtrip failed: %v", decoded.Amount)
	}
}
