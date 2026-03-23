package batch

import (
	"sync"
	"testing"
	"time"
)

func TestAccumulator_FlushSize(t *testing.T) {
	var mu sync.Mutex
	var flushed [][]int

	a := NewAccumulator(func(items []int) {
		mu.Lock()
		defer mu.Unlock()
		cp := make([]int, len(items))
		copy(cp, items)
		flushed = append(flushed, cp)
	}, FlushSize[int](3))

	a.Add(1)
	a.Add(2)
	a.Add(3) // triggers flush

	mu.Lock()
	defer mu.Unlock()

	if len(flushed) != 1 {
		t.Fatalf("expected 1 flush, got %d", len(flushed))
	}
	if len(flushed[0]) != 3 {
		t.Fatalf("expected 3 items in flush, got %d", len(flushed[0]))
	}
}

func TestAccumulator_ManualFlush(t *testing.T) {
	var flushed []int

	a := NewAccumulator(func(items []int) {
		flushed = append(flushed, items...)
	})

	a.Add(1)
	a.Add(2)
	a.Flush()

	if len(flushed) != 2 {
		t.Fatalf("expected 2 items, got %d", len(flushed))
	}

	// Second flush with no items should not call fn.
	before := len(flushed)
	a.Flush()
	if len(flushed) != before {
		t.Fatal("empty flush should not invoke callback")
	}
}

func TestAccumulator_FlushInterval(t *testing.T) {
	var mu sync.Mutex
	var flushed []int

	a := NewAccumulator(func(items []int) {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, items...)
	}, FlushInterval[int](50*time.Millisecond))

	a.Add(1)
	a.Add(2)

	// Wait for at least one interval tick.
	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	n := len(flushed)
	mu.Unlock()

	if n != 2 {
		t.Fatalf("expected 2 items flushed by interval, got %d", n)
	}

	a.Stop()
}

func TestAccumulator_Stop(t *testing.T) {
	var flushed []int

	a := NewAccumulator(func(items []int) {
		flushed = append(flushed, items...)
	})

	a.Add(1)
	a.Add(2)
	a.Stop()

	if len(flushed) != 2 {
		t.Fatalf("expected 2 items flushed on Stop, got %d", len(flushed))
	}
}

func TestAccumulator_ThreadSafety(t *testing.T) {
	var mu sync.Mutex
	var total int

	a := NewAccumulator(func(items []int) {
		mu.Lock()
		defer mu.Unlock()
		total += len(items)
	}, FlushSize[int](10))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			a.Add(v)
		}(i)
	}
	wg.Wait()
	a.Stop()

	mu.Lock()
	defer mu.Unlock()

	if total != 100 {
		t.Fatalf("expected 100 items total, got %d", total)
	}
}

func TestAccumulator_Stats(t *testing.T) {
	a := NewAccumulator(func(items []int) {}, FlushSize[int](3))

	a.Add(1)
	a.Add(2)

	stats := a.Stats()
	if stats.FlushCount != 0 {
		t.Fatalf("expected 0 flushes, got %d", stats.FlushCount)
	}
	if stats.TotalItems != 0 {
		t.Fatalf("expected 0 total items, got %d", stats.TotalItems)
	}
	if stats.Pending != 2 {
		t.Fatalf("expected 2 pending, got %d", stats.Pending)
	}

	a.Add(3) // triggers flush at size 3

	stats = a.Stats()
	if stats.FlushCount != 1 {
		t.Fatalf("expected 1 flush, got %d", stats.FlushCount)
	}
	if stats.TotalItems != 3 {
		t.Fatalf("expected 3 total items, got %d", stats.TotalItems)
	}
	if stats.Pending != 0 {
		t.Fatalf("expected 0 pending, got %d", stats.Pending)
	}

	a.Add(4)
	a.Add(5)
	a.Flush()

	stats = a.Stats()
	if stats.FlushCount != 2 {
		t.Fatalf("expected 2 flushes, got %d", stats.FlushCount)
	}
	if stats.TotalItems != 5 {
		t.Fatalf("expected 5 total items, got %d", stats.TotalItems)
	}
}

func TestAccumulator_Stats_ThreadSafety(t *testing.T) {
	a := NewAccumulator(func(items []int) {}, FlushSize[int](5))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			a.Add(v)
		}(i)
	}
	wg.Wait()
	a.Stop()

	stats := a.Stats()
	if stats.TotalItems != 50 {
		t.Fatalf("expected 50 total items, got %d", stats.TotalItems)
	}
}

func TestAccumulator_Peek(t *testing.T) {
	a := NewAccumulator[int](func(items []int) {})

	a.Add(1)
	a.Add(2)
	a.Add(3)

	peeked := a.Peek()
	if len(peeked) != 3 {
		t.Fatalf("expected 3 items, got %d", len(peeked))
	}
	if peeked[0] != 1 || peeked[1] != 2 || peeked[2] != 3 {
		t.Fatalf("unexpected peek values: %v", peeked)
	}

	// Peek should not flush — items should still be there.
	peeked2 := a.Peek()
	if len(peeked2) != 3 {
		t.Fatalf("expected 3 items after second peek, got %d", len(peeked2))
	}
}

func TestAccumulator_Peek_Empty(t *testing.T) {
	a := NewAccumulator[int](func(items []int) {})

	peeked := a.Peek()
	if peeked != nil {
		t.Fatalf("expected nil for empty peek, got %v", peeked)
	}
}

func TestAccumulator_Peek_ReturnsCopy(t *testing.T) {
	a := NewAccumulator[int](func(items []int) {})

	a.Add(1)
	a.Add(2)

	peeked := a.Peek()
	peeked[0] = 99 // mutate the copy

	peeked2 := a.Peek()
	if peeked2[0] != 1 {
		t.Fatal("Peek should return a copy; mutation affected original")
	}
}

func TestAccumulator_OnFlush(t *testing.T) {
	var mainFlushed []int
	var onFlushCalled []int

	a := NewAccumulator(func(items []int) {
		mainFlushed = append(mainFlushed, items...)
	}, FlushSize[int](2), OnFlush[int](func(items []int) {
		onFlushCalled = append(onFlushCalled, items...)
	}))

	a.Add(1)
	a.Add(2) // triggers flush

	if len(mainFlushed) != 2 {
		t.Fatalf("expected 2 main flushed items, got %d", len(mainFlushed))
	}
	if len(onFlushCalled) != 2 {
		t.Fatalf("expected 2 onFlush items, got %d", len(onFlushCalled))
	}

	a.Add(3)
	a.Stop()

	if len(onFlushCalled) != 3 {
		t.Fatalf("expected 3 onFlush items after stop, got %d", len(onFlushCalled))
	}
}

func TestAccumulator_OnFlush_CalledAfterMain(t *testing.T) {
	var order []string

	a := NewAccumulator(func(items []int) {
		order = append(order, "main")
	}, OnFlush[int](func(items []int) {
		order = append(order, "onFlush")
	}))

	a.Add(1)
	a.Flush()

	if len(order) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(order))
	}
	if order[0] != "main" {
		t.Fatal("expected main flush to be called first")
	}
	if order[1] != "onFlush" {
		t.Fatal("expected onFlush to be called second")
	}
}
