package worker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"banking-core/internal/domain"
)

// Pool processes transaction jobs concurrently via a buffered channel queue.
type Pool struct {
	queue     chan domain.TransactionJob
	workers   int
	wg        sync.WaitGroup
	cancel    context.CancelFunc
	processed atomic.Int64
	failed    atomic.Int64
}

func NewPool(workers, queueSize int) *Pool {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 1 {
		queueSize = 1
	}
	return &Pool{
		queue:   make(chan domain.TransactionJob, queueSize),
		workers: workers,
	}
}

func (p *Pool) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	p.cancel = cancel
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
	slog.Info("worker pool started", "workers", p.workers, "queue_size", cap(p.queue))
}

func (p *Pool) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	slog.Info("worker pool stopped",
		"processed", p.processed.Load(),
		"failed", p.failed.Load(),
	)
}

func (p *Pool) Submit(job domain.TransactionJob) {
	p.queue <- job
}

func (p *Pool) Processed() int64 { return p.processed.Load() }
func (p *Pool) Failed() int64    { return p.failed.Load() }

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.queue:
			if !ok {
				return
			}
			if job.Transaction != nil {
				job.Transaction.MarkProcessing()
			}
			var err error
			if job.Apply != nil {
				err = job.Apply()
			}
			if err != nil {
				p.failed.Add(1)
				if job.Transaction != nil {
					job.Transaction.MarkFailed()
				}
			} else {
				p.processed.Add(1)
				if job.Transaction != nil {
					job.Transaction.MarkCompleted()
				}
			}
			if job.ResultCh != nil {
				job.ResultCh <- err
				close(job.ResultCh)
			}
			_ = id
		}
	}
}

// BatchProcessor collects jobs and flushes them to the pool.
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

func (b *BatchProcessor) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.items)
}

func (b *BatchProcessor) Flush(pool *Pool) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(b.items)
	for _, job := range b.items {
		pool.Submit(job)
	}
	b.items = b.items[:0]
	return n
}
