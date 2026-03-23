package batch

import (
	"sync"
	"sync/atomic"
	"time"
)

// AccumulatorOption configures the behaviour of an Accumulator.
type AccumulatorOption[T any] func(*accumulatorConfig)

type accumulatorConfig struct {
	flushSize     int
	flushInterval time.Duration
	onFlush       any // stores func([]T) — type-asserted in flushLocked
}

// FlushSize configures the Accumulator to automatically flush when n items
// have been accumulated.
func FlushSize[T any](n int) AccumulatorOption[T] {
	return func(c *accumulatorConfig) {
		c.flushSize = n
	}
}

// FlushInterval configures the Accumulator to automatically flush on a
// recurring time interval. A background goroutine is started and runs until
// Stop is called.
func FlushInterval[T any](d time.Duration) AccumulatorOption[T] {
	return func(c *accumulatorConfig) {
		c.flushInterval = d
	}
}

// OnFlush registers a callback that is invoked after each flush with the
// flushed items. The callback runs after the main flush function.
func OnFlush[T any](fn func([]T)) AccumulatorOption[T] {
	return func(c *accumulatorConfig) {
		c.onFlush = fn
	}
}

// AccumulatorStats contains statistics about an Accumulator's lifetime.
type AccumulatorStats struct {
	// FlushCount is the total number of flushes performed.
	FlushCount int64
	// TotalItems is the total number of items flushed.
	TotalItems int64
	// Pending is the number of items currently buffered.
	Pending int
}

// Accumulator collects items and flushes them in batches. Flushing can be
// triggered manually, by reaching a size threshold, or on a time interval.
// All methods are safe for concurrent use.
type Accumulator[T any] struct {
	mu         sync.Mutex
	items      []T
	fn         func(items []T)
	cfg        accumulatorConfig
	stop       chan struct{}
	closed     bool
	flushCount atomic.Int64
	totalItems atomic.Int64
}

// NewAccumulator creates a new Accumulator that calls fn with the buffered
// items on each flush. Configure automatic flushing with FlushSize and
// FlushInterval options.
func NewAccumulator[T any](fn func(items []T), opts ...AccumulatorOption[T]) *Accumulator[T] {
	var cfg accumulatorConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	a := &Accumulator[T]{
		fn:   fn,
		cfg:  cfg,
		stop: make(chan struct{}),
	}

	if cfg.flushInterval > 0 {
		go a.runTicker()
	}

	return a
}

// Add appends an item to the accumulator. If a FlushSize has been configured
// and the number of buffered items reaches that threshold, a flush is
// triggered automatically.
func (a *Accumulator[T]) Add(item T) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.items = append(a.items, item)

	if a.cfg.flushSize > 0 && len(a.items) >= a.cfg.flushSize {
		a.flushLocked()
	}
}

// Flush sends all buffered items to the flush callback and resets the buffer.
// It is safe to call concurrently.
func (a *Accumulator[T]) Flush() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.flushLocked()
}

// Stop stops the interval timer (if running) and flushes any remaining
// buffered items. After Stop returns the accumulator should not be used.
func (a *Accumulator[T]) Stop() {
	a.mu.Lock()
	if !a.closed {
		a.closed = true
		close(a.stop)
	}
	a.mu.Unlock()

	// Flush remaining items outside the closed-check lock to avoid issues
	// with the ticker goroutine.
	a.Flush()
}

// Stats returns statistics about the accumulator's lifetime. It is safe to
// call concurrently.
func (a *Accumulator[T]) Stats() AccumulatorStats {
	a.mu.Lock()
	pending := len(a.items)
	a.mu.Unlock()

	return AccumulatorStats{
		FlushCount: a.flushCount.Load(),
		TotalItems: a.totalItems.Load(),
		Pending:    pending,
	}
}

// Peek returns a copy of the currently buffered items without flushing.
// It is safe to call concurrently.
func (a *Accumulator[T]) Peek() []T {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.items) == 0 {
		return nil
	}

	cp := make([]T, len(a.items))
	copy(cp, a.items)
	return cp
}

// flushLocked sends buffered items to the callback. Must be called with
// a.mu held.
func (a *Accumulator[T]) flushLocked() {
	if len(a.items) == 0 {
		return
	}
	items := a.items
	a.items = nil
	a.fn(items)

	a.flushCount.Add(1)
	a.totalItems.Add(int64(len(items)))

	if a.cfg.onFlush != nil {
		if cb, ok := a.cfg.onFlush.(func([]T)); ok {
			cb(items)
		}
	}
}

func (a *Accumulator[T]) runTicker() {
	ticker := time.NewTicker(a.cfg.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.Flush()
		case <-a.stop:
			return
		}
	}
}
