package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"banking-api/internal/domain"
)

type BalanceService struct {
	balances domain.BalanceRepository
	cache    domain.CacheStore
	mu       sync.RWMutex
}

func NewBalanceService(balances domain.BalanceRepository, cache domain.CacheStore) *BalanceService {
	return &BalanceService{balances: balances, cache: cache}
}

func (s *BalanceService) GetCurrent(ctx context.Context, userID int64) (*domain.Balance, error) {
	cacheKey := fmt.Sprintf("balance:%d", userID)
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			amount, _ := strconv.ParseFloat(cached, 64)
			return &domain.Balance{UserID: userID, Amount: amount, LastUpdatedAt: time.Now().UTC()}, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	balance, err := s.balances.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, fmt.Sprintf("%.2f", balance.Amount), 30*time.Second)
	}
	return balance, nil
}

func (s *BalanceService) GetHistorical(ctx context.Context, userID int64) ([]domain.BalanceSnapshot, error) {
	return s.balances.ListHistorical(ctx, userID)
}

func (s *BalanceService) GetAtTime(ctx context.Context, userID int64, at time.Time) (*domain.BalanceSnapshot, error) {
	return s.balances.GetAtTime(ctx, userID, at)
}

func (s *BalanceService) ApplyDelta(ctx context.Context, userID int64, delta float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.balances.UpdateAmount(ctx, userID, delta); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, fmt.Sprintf("balance:%d", userID))
	}
	return nil
}
