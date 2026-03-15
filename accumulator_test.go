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
