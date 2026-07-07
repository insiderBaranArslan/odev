package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"banking-api/internal/domain"
	"banking-api/internal/worker"
)

type TransactionService struct {
	transactions domain.TransactionRepository
	balances     *BalanceService
	users        domain.UserRepository
	audit        domain.AuditRepository
	events       domain.EventRepository
	pool         *worker.Pool
	processed    atomic.Int64
	failed       atomic.Int64
}

func NewTransactionService(
	transactions domain.TransactionRepository,
	balances *BalanceService,
	users domain.UserRepository,
	audit domain.AuditRepository,
	events domain.EventRepository,
	pool *worker.Pool,
) *TransactionService {
	return &TransactionService{
		transactions: transactions,
		balances:     balances,
		users:        users,
		audit:        audit,
		events:       events,
		pool:         pool,
	}
}

func (s *TransactionService) Credit(ctx context.Context, req domain.CreditRequest) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		ToUserID: req.UserID,
		Amount:   req.Amount,
		Type:     domain.TransactionTypeCredit,
		Status:   domain.TransactionStatusPending,
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.users.GetByID(ctx, req.UserID); err != nil {
		return nil, err
	}
	return s.execute(ctx, tx, func(ctx context.Context, t *domain.Transaction) error {
		return s.balances.ApplyDelta(ctx, t.ToUserID, t.Amount)
	})
}

func (s *TransactionService) Debit(ctx context.Context, req domain.DebitRequest) (*domain.Transaction, error) {
	from := req.UserID
	tx := &domain.Transaction{
		FromUserID: &from,
		ToUserID:   req.UserID,
		Amount:     req.Amount,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}
	return s.execute(ctx, tx, func(ctx context.Context, t *domain.Transaction) error {
		return s.balances.ApplyDelta(ctx, *t.FromUserID, -t.Amount)
	})
}

func (s *TransactionService) Transfer(ctx context.Context, req domain.TransferRequest) (*domain.Transaction, error) {
	from := req.FromUserID
	tx := &domain.Transaction{
		FromUserID: &from,
		ToUserID:   req.ToUserID,
		Amount:     req.Amount,
		Type:       domain.TransactionTypeTransfer,
		Status:     domain.TransactionStatusPending,
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}
	return s.execute(ctx, tx, func(ctx context.Context, t *domain.Transaction) error {
		if err := s.balances.ApplyDelta(ctx, *t.FromUserID, -t.Amount); err != nil {
			return err
		}
		return s.balances.ApplyDelta(ctx, t.ToUserID, t.Amount)
	})
}

func (s *TransactionService) execute(ctx context.Context, tx *domain.Transaction, apply func(context.Context, *domain.Transaction) error) (*domain.Transaction, error) {
	if err := s.transactions.Create(ctx, tx); err != nil {
		return nil, err
	}

	resultCh := make(chan error, 1)
	s.pool.Submit(domain.TransactionJob{
		Transaction: tx,
		ResultCh:    resultCh,
		Apply:       apply,
	})

	select {
	case err := <-resultCh:
		if err != nil {
			s.failed.Add(1)
			tx.MarkFailed()
			_ = s.transactions.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed)
			return tx, err
		}
		s.processed.Add(1)
		tx.MarkCompleted()
		_ = s.transactions.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted)
		s.recordEvent(ctx, tx)
		_ = s.audit.Create(ctx, &domain.AuditLog{
			EntityType: "transaction",
			EntityID:   tx.ID,
			Action:     string(tx.Type),
			Details:    mustJSON(tx),
		})
		return tx, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *TransactionService) GetTransaction(ctx context.Context, id int64) (*domain.Transaction, error) {
	return s.transactions.GetByID(ctx, id)
}

func (s *TransactionService) GetHistory(ctx context.Context, userID int64) ([]domain.Transaction, error) {
	return s.transactions.ListByUser(ctx, userID)
}

func (s *TransactionService) Rollback(ctx context.Context, id int64) error {
	tx, err := s.transactions.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tx.Status != domain.TransactionStatusCompleted {
		return fmt.Errorf("only completed transactions can be rolled back")
	}

	switch tx.Type {
	case domain.TransactionTypeCredit:
		if err := s.balances.ApplyDelta(ctx, tx.ToUserID, -tx.Amount); err != nil {
			return err
		}
	case domain.TransactionTypeDebit:
		if err := s.balances.ApplyDelta(ctx, *tx.FromUserID, tx.Amount); err != nil {
			return err
		}
	case domain.TransactionTypeTransfer:
		if err := s.balances.ApplyDelta(ctx, *tx.FromUserID, tx.Amount); err != nil {
			return err
		}
		if err := s.balances.ApplyDelta(ctx, tx.ToUserID, -tx.Amount); err != nil {
			return err
		}
	}

	tx.MarkRolledBack()
	if err := s.transactions.UpdateStatus(ctx, tx.ID, domain.TransactionStatusRolledBack); err != nil {
		return err
	}
	_ = s.audit.Create(ctx, &domain.AuditLog{
		EntityType: "transaction",
		EntityID:   tx.ID,
		Action:     "rollback",
		Details:    mustJSON(map[string]any{"transaction_id": tx.ID}),
	})
	return nil
}

func (s *TransactionService) Stats() domain.TransactionStats {
	pending, _ := s.transactions.CountByStatus(context.Background(), domain.TransactionStatusPending)
	return domain.TransactionStats{
		Processed: s.processed.Load() + s.pool.ProcessedCount(),
		Failed:    s.failed.Load() + s.pool.FailedCount(),
		Pending:   pending,
	}
}

func (s *TransactionService) Enqueue(job domain.TransactionJob) {
	s.pool.Submit(job)
}

func (s *TransactionService) ReplayEvents(ctx context.Context, aggregateID int64) ([]domain.Event, error) {
	return s.events.ListByAggregate(ctx, "transaction", aggregateID)
}

func (s *TransactionService) recordEvent(ctx context.Context, tx *domain.Transaction) {
	payload, _ := json.Marshal(tx)
	_ = s.events.Append(ctx, &domain.Event{
		Aggregate:   "transaction",
		AggregateID: tx.ID,
		EventType:   string(tx.Type) + ".completed",
		Payload:     payload,
	})
}
