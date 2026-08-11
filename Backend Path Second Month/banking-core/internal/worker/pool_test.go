package worker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"banking-core/internal/domain"
	"banking-core/internal/worker"
)

func TestWorkerPoolProcessesJobs(t *testing.T) {
	pool := worker.NewPool(3, 20)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)
	defer pool.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		result := make(chan error, 1)
		tx := &domain.Transaction{Amount: 1, Type: domain.TypeCredit, Status: domain.StatusPending, ToUserID: 1}
		pool.Submit(domain.TransactionJob{
			Transaction: tx,
			Apply: func() error {
				defer wg.Done()
				return nil
			},
			ResultCh: result,
		})
		if err := <-result; err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()

	if pool.Processed() < 10 {
		t.Fatalf("expected at least 10 processed, got %d", pool.Processed())
	}
}

func TestBatchProcessorFlush(t *testing.T) {
	pool := worker.NewPool(2, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)
	defer pool.Stop()

	batch := worker.NewBatchProcessor()
	for i := 0; i < 5; i++ {
		result := make(chan error, 1)
		batch.Add(domain.TransactionJob{
			Transaction: &domain.Transaction{ToUserID: 1, Amount: 1, Type: domain.TypeCredit},
			Apply:       func() error { return nil },
			ResultCh:    result,
		})
	}
	if n := batch.Flush(pool); n != 5 {
		t.Fatalf("expected flush 5, got %d", n)
	}

	deadline := time.After(2 * time.Second)
	for pool.Processed() < 5 {
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for batch, processed=%d", pool.Processed())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestWorkerPoolFailedCounter(t *testing.T) {
	pool := worker.NewPool(1, 5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)
	defer pool.Stop()

	result := make(chan error, 1)
	pool.Submit(domain.TransactionJob{
		Transaction: &domain.Transaction{ToUserID: 1, Amount: 1, Type: domain.TypeCredit},
		Apply:       func() error { return errors.New("boom") },
		ResultCh:    result,
	})
	if err := <-result; err == nil {
		t.Fatal("expected error")
	}
	if pool.Failed() < 1 {
		t.Fatalf("expected failed counter")
	}
}
