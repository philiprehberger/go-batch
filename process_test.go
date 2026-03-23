package batch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestProcess_Sequential(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	var order [][]int
	var mu sync.Mutex

	err := Process(context.Background(), items, 2, func(_ context.Context, batch []int) error {
		mu.Lock()
		defer mu.Unlock()
		cp := make([]int, len(batch))
		copy(cp, batch)
		order = append(order, cp)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(order))
	}
}

func TestProcess_Concurrent(t *testing.T) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	var count atomic.Int64

	err := Process(context.Background(), items, 10, func(_ context.Context, batch []int) error {
		count.Add(int64(len(batch)))
		return nil
	}, WithWorkers(4))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.Load() != 100 {
		t.Fatalf("expected 100 items processed, got %d", count.Load())
	}
}

func TestProcess_Error(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	errBoom := errors.New("boom")

	err := Process(context.Background(), items, 2, func(_ context.Context, batch []int) error {
		if batch[0] == 3 {
			return errBoom
		}
		return nil
	})

	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom in error chain, got: %v", err)
	}
}

func TestProcess_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	items := []int{1, 2, 3, 4, 5}

	err := Process(ctx, items, 1, func(ctx context.Context, batch []int) error {
		return nil
	})

	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestProcess_EmptyItems(t *testing.T) {
	called := false
	err := Process(context.Background(), []int{}, 5, func(_ context.Context, batch []int) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("fn should not be called for empty items")
	}
}

func TestProcessWithErrors_NoErrors(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	errs := ProcessWithErrors(items, 2, 2, func(batch []int) error {
		return nil
	})

	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
}

func TestProcessWithErrors_CollectsErrors(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6}
	errs := ProcessWithErrors(items, 2, 1, func(batch []int) error {
		if batch[0] == 3 {
			return errors.New("batch starting with 3 failed")
		}
		if batch[0] == 5 {
			return errors.New("batch starting with 5 failed")
		}
		return nil
	})

	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}

func TestProcessWithErrors_Empty(t *testing.T) {
	errs := ProcessWithErrors([]int{}, 5, 1, func(batch []int) error {
		return errors.New("should not be called")
	})

	if errs != nil {
		t.Fatalf("expected nil for empty items, got %v", errs)
	}
}

func TestProcessWithErrors_AllFail(t *testing.T) {
	items := []int{1, 2, 3}
	errs := ProcessWithErrors(items, 1, 2, func(batch []int) error {
		return errors.New("fail")
	})

	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errs))
	}
}
