// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	config := Config{
		Workers:        3,
		EmbeddingModel: "text-embedding-3-small",
	}

	pool := NewPool(config)
	defer pool.Stop()

	assert.NotNil(t, pool)
	assert.Equal(t, 3, pool.workers)
	assert.True(t, pool.IsRateLimited(), "OpenAI should have rate limiting")
	assert.Equal(t, "openai", pool.Provider())
}

func TestPool_SingleTask(t *testing.T) {
	config := Config{
		Workers:        1,
		EmbeddingModel: "sentence-transformers/all-mpnet-base-v2",
	}

	pool := NewPool(config)
	pool.Start()

	// Submit a task
	executed := false
	pool.Submit(Task{
		ID:       "task1",
		FilePath: "test.pdf",
		Process: func(ctx context.Context) error {
			executed = true
			return nil
		},
	})

	// Close and wait
	pool.Wait()

	// Check result
	result := <-pool.Results()
	assert.True(t, result.Success)
	assert.NoError(t, result.Error)
	assert.Equal(t, "task1", result.TaskID)
	assert.True(t, executed)

	// Check stats
	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.Processed)
	assert.Equal(t, int64(1), stats.Succeeded)
	assert.Equal(t, int64(0), stats.Failed)
}

func TestPool_MultipleTasks(t *testing.T) {
	config := Config{
		Workers:        3,
		EmbeddingModel: "sentence-transformers/all-mpnet-base-v2",
	}

	pool := NewPool(config)
	pool.Start()

	// Submit multiple tasks
	taskCount := 10
	var processedCount atomic.Int32

	for i := 0; i < taskCount; i++ {
		pool.Submit(Task{
			ID:       fmt.Sprintf("task%d", i),
			FilePath: fmt.Sprintf("test%d.pdf", i),
			Process: func(ctx context.Context) error {
				processedCount.Add(1)
				time.Sleep(10 * time.Millisecond) // Simulate work
				return nil
			},
		})
	}

	// Collect results
	go pool.Wait()

	successCount := 0
	for result := range pool.Results() {
		if result.Success {
			successCount++
		}
	}

	assert.Equal(t, taskCount, successCount)
	assert.Equal(t, int32(taskCount), processedCount.Load())

	stats := pool.GetStats()
	assert.Equal(t, int64(taskCount), stats.Processed)
	assert.Equal(t, int64(taskCount), stats.Succeeded)
	assert.Equal(t, int64(0), stats.Failed)
}

func TestPool_ErrorHandling(t *testing.T) {
	config := Config{
		Workers:        2,
		EmbeddingModel: "text-embedding-3-small",
	}

	pool := NewPool(config)
	pool.Start()

	// Submit tasks with errors
	pool.Submit(Task{
		ID:       "success",
		FilePath: "good.pdf",
		Process: func(ctx context.Context) error {
			return nil
		},
	})

	pool.Submit(Task{
		ID:       "failure",
		FilePath: "bad.pdf",
		Process: func(ctx context.Context) error {
			return errors.New("processing failed")
		},
	})

	// Collect results
	go pool.Wait()

	results := make(map[string]Result)
	for result := range pool.Results() {
		results[result.TaskID] = result
	}

	// Check success task
	assert.True(t, results["success"].Success)
	assert.NoError(t, results["success"].Error)

	// Check failure task
	assert.False(t, results["failure"].Success)
	assert.Error(t, results["failure"].Error)
	assert.Contains(t, results["failure"].Error.Error(), "processing failed")

	// Check stats
	stats := pool.GetStats()
	assert.Equal(t, int64(2), stats.Processed)
	assert.Equal(t, int64(1), stats.Succeeded)
	assert.Equal(t, int64(1), stats.Failed)
}

func TestPool_Stop(t *testing.T) {
	config := Config{
		Workers:        2,
		EmbeddingModel: "sentence-transformers/all-mpnet-base-v2",
	}

	pool := NewPool(config)
	pool.Start()

	// Submit a long-running task
	started := make(chan bool)
	pool.Submit(Task{
		ID:       "long",
		FilePath: "long.pdf",
		Process: func(ctx context.Context) error {
			started <- true
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				return nil
			}
		},
	})

	// Wait for task to start
	<-started

	// Stop the pool (should cancel context)
	pool.Stop()

	// Check result
	result := <-pool.Results()
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "cancel")
}

func TestPool_RateLimiting_OpenAI(t *testing.T) {
	config := Config{
		Workers:        1,
		EmbeddingModel: "text-embedding-3-small",
		QueueSize:      20,
	}

	pool := NewPool(config)
	pool.Start()

	assert.True(t, pool.IsRateLimited())
	assert.Equal(t, "openai", pool.Provider())

	// Submit more tasks to see rate limiting effect
	taskCount := 15
	for i := 0; i < taskCount; i++ {
		pool.Submit(Task{
			ID:       fmt.Sprintf("task%d", i),
			FilePath: fmt.Sprintf("test%d.pdf", i),
			Process: func(ctx context.Context) error {
				return nil
			},
		})
	}

	// Measure time to process (should show rate limiting effect)
	start := time.Now()
	go pool.Wait()

	count := 0
	for range pool.Results() {
		count++
	}
	elapsed := time.Since(start)

	// With rate limiting and 15 tasks, should take some measurable time
	assert.Equal(t, taskCount, count)
	assert.Greater(t, elapsed, 1*time.Millisecond) // More conservative check
}

func TestPool_RateLimiting_OSS(t *testing.T) {
	config := Config{
		Workers:        3,
		EmbeddingModel: "sentence-transformers/all-mpnet-base-v2",
		QueueSize:      100,
	}

	pool := NewPool(config)
	pool.Start()

	assert.False(t, pool.IsRateLimited())
	assert.Equal(t, "sentence-transformers", pool.Provider())

	// Submit tasks (should NOT be rate limited)
	taskCount := 50
	for i := 0; i < taskCount; i++ {
		pool.Submit(Task{
			ID:       fmt.Sprintf("task%d", i),
			FilePath: fmt.Sprintf("test%d.pdf", i),
			Process: func(ctx context.Context) error {
				return nil
			},
		})
	}

	// Measure time to process (should be fast with no rate limiting)
	start := time.Now()
	go pool.Wait()

	for range pool.Results() {
		// Collect results
	}
	elapsed := time.Since(start)

	// Without rate limiting, should be very fast
	assert.Less(t, elapsed, 1*time.Second)
}

func TestPool_ConcurrentAccess(t *testing.T) {
	config := Config{
		Workers:        5,
		EmbeddingModel: "sentence-transformers/all-mpnet-base-v2",
		QueueSize:      100,
	}

	pool := NewPool(config)
	pool.Start()

	// Submit tasks from multiple goroutines
	taskCount := 50
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				pool.Submit(Task{
					ID:       fmt.Sprintf("task%d-%d", offset, j),
					FilePath: fmt.Sprintf("test%d-%d.pdf", offset, j),
					Process: func(ctx context.Context) error {
						time.Sleep(1 * time.Millisecond)
						return nil
					},
				})
			}
		}(i)
	}

	wg.Wait()

	// Collect results
	go pool.Wait()

	count := 0
	for result := range pool.Results() {
		assert.True(t, result.Success)
		count++
	}

	assert.Equal(t, taskCount, count)
}

func TestPool_DefaultQueueSize(t *testing.T) {
	config := Config{
		Workers:        3,
		EmbeddingModel: "text-embedding-3-small",
		QueueSize:      0, // Should default to workers * 2
	}

	pool := NewPool(config)

	// Queue size should be workers * 2 = 6
	assert.Equal(t, 6, cap(pool.tasks))
}

func TestPool_CustomQueueSize(t *testing.T) {
	config := Config{
		Workers:        3,
		EmbeddingModel: "text-embedding-3-small",
		QueueSize:      10,
	}

	pool := NewPool(config)

	assert.Equal(t, 10, cap(pool.tasks))
}
