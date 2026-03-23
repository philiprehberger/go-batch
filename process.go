package batch

import (
	"context"
	"errors"
	"sync"
)

// ProcessOption configures the behaviour of Process.
type ProcessOption func(*processConfig)

type processConfig struct {
	workers int
}

// WithWorkers sets the number of concurrent workers used by Process.
// The default is 1 (sequential processing). If n is less than 1 it is
// treated as 1.
func WithWorkers(n int) ProcessOption {
	return func(c *processConfig) {
		if n > 0 {
			c.workers = n
		}
	}
}

// Process splits items into batches of batchSize and calls fn for each batch.
// Batches are processed concurrently up to the number of workers configured
// via WithWorkers (default 1). If any invocation of fn returns an error, all
// errors are collected and returned as a single combined error. Processing
// respects context cancellation — remaining batches are skipped when ctx is
// cancelled.
func Process[T any](ctx context.Context, items []T, batchSize int, fn func(ctx context.Context, batch []T) error, opts ...ProcessOption) error {
	cfg := processConfig{workers: 1}
	for _, opt := range opts {
		opt(&cfg)
	}

	batches := Chunk(items, batchSize)
	if len(batches) == 0 {
		return nil
	}

	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	// Semaphore channel to limit concurrency.
	sem := make(chan struct{}, cfg.workers)

	for _, b := range batches {
		// Check for context cancellation before starting a new batch.
		select {
		case <-ctx.Done():
			mu.Lock()
			errs = append(errs, ctx.Err())
			mu.Unlock()
			return errors.Join(errs...)
		default:
		}

		sem <- struct{}{} // acquire
		wg.Add(1)

		go func(batch []T) {
			defer wg.Done()
			defer func() { <-sem }() // release

			if err := fn(ctx, batch); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(b)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// ProcessWithErrors splits items into batches of batchSize and calls fn for
// each batch using the given number of concurrent workers. Unlike Process,
// it does not use a context and returns all non-nil errors as a slice rather
// than combining them.
func ProcessWithErrors[T any](items []T, size int, workers int, fn func([]T) error) []error {
	batches := Chunk(items, size)
	if len(batches) == 0 {
		return nil
	}

	if workers < 1 {
		workers = 1
	}

	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)

	sem := make(chan struct{}, workers)

	for _, b := range batches {
		sem <- struct{}{}
		wg.Add(1)

		go func(batch []T) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := fn(batch); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(b)
	}

	wg.Wait()

	return errs
}
