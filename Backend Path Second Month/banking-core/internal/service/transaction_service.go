package service

import (
	"context"
	"fmt"
	"sync/atomic"

	"banking-core/internal/domain"
	"banking-core/internal/worker"
)

type TransactionService struct {
	txs      domain.TransactionRepository
	users    domain.UserRepository
	balances *BalanceService
	audit    domain.AuditRepository
	pool     *worker.Pool
	batch    *worker.BatchProcessor

	processed atomic.Int64
	failed    atomic.Int64
}

func NewTransactionService(
	txs domain.TransactionRepository,
	users domain.UserRepository,
	balances *BalanceService,
	audit domain.AuditRepository,
	pool *worker.Pool,
) *TransactionService {
	return &TransactionService{
		txs:      txs,
		users:    users,
		balances: balances,
		audit:    audit,
		pool:     pool,
		batch:    worker.NewBatchProcessor(),
	}
}

func (s *TransactionService) Credit(ctx context.Context, req domain.CreditRequest) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		ToUserID: req.UserID,
		Amount:   req.Amount,
		Type:     domain.TypeCredit,
		Status:   domain.StatusPending,
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.users.GetByID(ctx, req.UserID); err != nil {
		return nil, err
	}
	return s.execute(ctx, tx, func() error {
		return s.balances.ApplyDelta(ctx, tx.ToUserID, tx.Amount)
	})
}

func (s *TransactionService) Debit(ctx context.Context, req domain.DebitRequest) (*domain.Transaction, error) {
	from := req.UserID
	tx := &domain.Transaction{
		FromUserID: &from,
		ToUserID:   req.UserID,
		Amount:     req.Amount,
		Type:       domain.TypeDebit,
		Status:     domain.StatusPending,
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}
	return s.execute(ctx, tx, func() error {
		return s.balances.ApplyDelta(ctx, *tx.FromUserID, -tx.Amount)
	})
}

func (s *TransactionService) Transfer(ctx context.Context, req domain.TransferRequest) (*domain.Transaction, error) {
	from := req.FromUserID
	tx := &domain.Transaction{
		FromUserID: &from,
		ToUserID:   req.ToUserID,
		Amount:     req.Amount,
		Type:       domain.TypeTransfer,
		Status:     domain.StatusPending,
	}
	if err := tx.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.users.GetByID(ctx, req.FromUserID); err != nil {
		return nil, fmt.Errorf("from user: %w", err)
	}
	if _, err := s.users.GetByID(ctx, req.ToUserID); err != nil {
		return nil, fmt.Errorf("to user: %w", err)
	}
	return s.execute(ctx, tx, func() error {
		if err := s.balances.ApplyDelta(ctx, *tx.FromUserID, -tx.Amount); err != nil {
			return err
		}
		if err := s.balances.ApplyDelta(ctx, tx.ToUserID, tx.Amount); err != nil {
			_ = s.balances.ApplyDelta(ctx, *tx.FromUserID, tx.Amount) // compensate
			return err
		}
		return nil
	})
}

func (s *TransactionService) execute(ctx context.Context, tx *domain.Transaction, apply func() error) (*domain.Transaction, error) {
	if err := s.txs.Create(ctx, tx); err != nil {
		return nil, err
	}

	resultCh := make(chan error, 1)
	s.pool.Submit(domain.TransactionJob{
		Transaction: tx,
		Apply:       apply,
		ResultCh:    resultCh,
	})

	select {
	case err := <-resultCh:
		if err != nil {
			s.failed.Add(1)
			tx.MarkFailed()
			_ = s.txs.UpdateStatus(ctx, tx.ID, domain.StatusFailed)
			return tx, err
		}
		s.processed.Add(1)
		tx.MarkCompleted()
		_ = s.txs.UpdateStatus(ctx, tx.ID, domain.StatusCompleted)
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

func (s *TransactionService) Rollback(ctx context.Context, id int64) error {
	tx, err := s.txs.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tx.Status != domain.StatusCompleted {
		return fmt.Errorf("only completed transactions can be rolled back")
	}

	switch tx.Type {
	case domain.TypeCredit:
		if err := s.balances.ApplyDelta(ctx, tx.ToUserID, -tx.Amount); err != nil {
			return err
		}
	case domain.TypeDebit:
		if err := s.balances.ApplyDelta(ctx, *tx.FromUserID, tx.Amount); err != nil {
			return err
		}
	case domain.TypeTransfer:
		if err := s.balances.ApplyDelta(ctx, *tx.FromUserID, tx.Amount); err != nil {
			return err
		}
		if err := s.balances.ApplyDelta(ctx, tx.ToUserID, -tx.Amount); err != nil {
			_ = s.balances.ApplyDelta(ctx, *tx.FromUserID, -tx.Amount)
			return err
		}
	default:
		return fmt.Errorf("unsupported transaction type for rollback")
	}

	tx.MarkRolledBack()
	if err := s.txs.UpdateStatus(ctx, tx.ID, domain.StatusRolledBack); err != nil {
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

func (s *TransactionService) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return s.txs.GetByID(ctx, id)
}

func (s *TransactionService) History(ctx context.Context, userID int64) ([]domain.Transaction, error) {
	return s.txs.ListByUser(ctx, userID)
}

func (s *TransactionService) Stats() domain.TransactionStats {
	pending, _ := s.txs.CountByStatus(context.Background(), domain.StatusPending)
	return domain.TransactionStats{
		Processed: s.processed.Load() + s.pool.Processed(),
		Failed:    s.failed.Load() + s.pool.Failed(),
		Pending:   pending,
	}
}

// EnqueueBatch adds a job to the batch processor for concurrent batch flush.
func (s *TransactionService) EnqueueBatch(job domain.TransactionJob) {
	s.batch.Add(job)
}

func (s *TransactionService) FlushBatch() int {
	return s.batch.Flush(s.pool)
}
