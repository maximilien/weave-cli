// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/maximilien/weave-cli/src/pkg/ratelimit"
)

// Task represents a unit of work to be processed
type Task struct {
	ID       string
	FilePath string
	Process  func(context.Context) error
}

// Result contains the outcome of processing a task
type Result struct {
	TaskID   string
	FilePath string
	Error    error
	Success  bool
}

// Pool manages a pool of workers for parallel task processing
type Pool struct {
	workers     int
	tasks       chan Task
	results     chan Result
	rateLimiter *ratelimit.Limiter
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc

	// Metrics
	processed atomic.Int64
	failed    atomic.Int64
	succeeded atomic.Int64
}

// Config contains worker pool configuration
type Config struct {
	Workers        int    // Number of concurrent workers
	EmbeddingModel string // For rate limiting
	QueueSize      int    // Task queue buffer size (default: workers * 2)
}

// NewPool creates a new worker pool
func NewPool(config Config) *Pool {
	// Default queue size to workers * 2
	queueSize := config.QueueSize
	if queueSize == 0 {
		queueSize = config.Workers * 2
	}

	// Create rate limiter based on embedding model
	limiter := ratelimit.NewLimiter(config.EmbeddingModel)

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workers:     config.Workers,
		tasks:       make(chan Task, queueSize),
		results:     make(chan Result, queueSize),
		rateLimiter: limiter,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the worker pool
func (p *Pool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker processes tasks from the queue
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for task := range p.tasks {
		// Check if context is cancelled
		select {
		case <-p.ctx.Done():
			p.results <- Result{
				TaskID:   task.ID,
				FilePath: task.FilePath,
				Error:    fmt.Errorf("cancelled: %w", p.ctx.Err()),
				Success:  false,
			}
			continue
		default:
		}

		// Rate limit if needed
		if err := p.rateLimiter.Wait(p.ctx); err != nil {
			p.results <- Result{
				TaskID:   task.ID,
				FilePath: task.FilePath,
				Error:    fmt.Errorf("rate limit wait failed: %w", err),
				Success:  false,
			}
			p.failed.Add(1)
			p.processed.Add(1)
			continue
		}

		// Process the task
		err := task.Process(p.ctx)

		// Record result
		result := Result{
			TaskID:   task.ID,
			FilePath: task.FilePath,
			Error:    err,
			Success:  err == nil,
		}

		// Update metrics
		p.processed.Add(1)
		if err != nil {
			p.failed.Add(1)
		} else {
			p.succeeded.Add(1)
		}

		p.results <- result
	}
}

// Submit adds a task to the pool
func (p *Pool) Submit(task Task) {
	p.tasks <- task
}

// Wait closes the task queue and waits for all workers to finish
func (p *Pool) Wait() {
	close(p.tasks)
	p.wg.Wait()
	close(p.results)
}

// Stop cancels all workers and waits for them to finish
func (p *Pool) Stop() {
	p.cancel()
	p.Wait()
}

// Results returns the results channel
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Stats returns current processing statistics
type Stats struct {
	Processed int64
	Succeeded int64
	Failed    int64
}

// GetStats returns current processing statistics
func (p *Pool) GetStats() Stats {
	return Stats{
		Processed: p.processed.Load(),
		Succeeded: p.succeeded.Load(),
		Failed:    p.failed.Load(),
	}
}

// IsRateLimited returns true if rate limiting is active
func (p *Pool) IsRateLimited() bool {
	return p.rateLimiter.Enabled()
}

// Provider returns the detected embedding provider
func (p *Pool) Provider() string {
	return p.rateLimiter.Provider()
}
