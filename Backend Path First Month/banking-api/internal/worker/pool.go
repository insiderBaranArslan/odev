package worker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"banking-api/internal/domain"
)

type Pool struct {
	queue     chan domain.TransactionJob
	wg        sync.WaitGroup
	workers   int
	processed atomic.Int64
	failed    atomic.Int64
	cancel    context.CancelFunc
}

func NewPool(workers, queueSize int) *Pool {
	return &Pool{
		queue:   make(chan domain.TransactionJob, queueSize),
		workers: workers,
	}
}

func (p *Pool) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.runWorker(ctx)
	}
	slog.Info("worker pool started", "workers", p.workers)
}

func (p *Pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	slog.Info("worker pool stopped")
}

func (p *Pool) Submit(job domain.TransactionJob) {
	p.queue <- job
}

func (p *Pool) ProcessedCount() int64 {
	return p.processed.Load()
}

func (p *Pool) FailedCount() int64 {
	return p.failed.Load()
}

func (p *Pool) runWorker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.queue:
			if !ok {
				return
			}
			var err error
			if job.Apply != nil {
				err = job.Apply(ctx, job.Transaction)
			}
			if err != nil {
				p.failed.Add(1)
			} else {
				p.processed.Add(1)
			}
			if job.ResultCh != nil {
				job.ResultCh <- err
				close(job.ResultCh)
			}
		}
	}
}

type BatchProcessor struct {
	mu    sync.Mutex
	items []domain.TransactionJob
}

func NewBatchProcessor() *BatchProcessor {
	return &BatchProcessor{items: make([]domain.TransactionJob, 0)}
}

func (b *BatchProcessor) Add(job domain.TransactionJob) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = append(b.items, job)
}

func (b *BatchProcessor) Flush(pool *Pool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, job := range b.items {
		pool.Submit(job)
	}
	b.items = b.items[:0]
}
