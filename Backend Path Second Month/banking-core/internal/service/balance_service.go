package service

import (
	"context"
	"sync"
	"time"

	"banking-core/internal/domain"
)

// BalanceService provides thread-safe balance updates and historical lookups.
type BalanceService struct {
	repo domain.BalanceRepository
	mu   sync.RWMutex
}

func NewBalanceService(repo domain.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

func (s *BalanceService) GetCurrent(ctx context.Context, userID int64) (*domain.Balance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.repo.Get(ctx, userID)
}

func (s *BalanceService) ApplyDelta(ctx context.Context, userID int64, delta float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.repo.UpdateAmount(ctx, userID, delta)
	return err
}

func (s *BalanceService) GetHistorical(ctx context.Context, userID int64) ([]domain.BalanceSnapshot, error) {
	return s.repo.ListHistorical(ctx, userID)
}

func (s *BalanceService) GetAtTime(ctx context.Context, userID int64, at time.Time) (*domain.BalanceSnapshot, error) {
	return s.repo.GetAtTime(ctx, userID, at)
}

// OptimizeCurrent reads once under RLock — used for read-heavy balance checks.
func (s *BalanceService) OptimizeCurrent(ctx context.Context, userIDs []int64) (map[int64]float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[int64]float64, len(userIDs))
	for _, id := range userIDs {
		b, err := s.repo.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		out[id] = b.GetAmount()
	}
	return out, nil
}
