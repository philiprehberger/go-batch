package batch

import (
	"testing"
)

func TestChunk_EvenSplit(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	chunks := Chunk(items, 5)

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if len(chunks[0]) != 5 {
		t.Fatalf("expected first chunk length 5, got %d", len(chunks[0]))
	}
	if len(chunks[1]) != 5 {
		t.Fatalf("expected second chunk length 5, got %d", len(chunks[1]))
	}
}

func TestChunk_UnevenSplit(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	chunks := Chunk(items, 3)

	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}
	if len(chunks[3]) != 1 {
		t.Fatalf("expected last chunk length 1, got %d", len(chunks[3]))
	}
}

func TestChunk_SingleChunk(t *testing.T) {
	items := []int{1, 2, 3}
	chunks := Chunk(items, 10)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if len(chunks[0]) != 3 {
		t.Fatalf("expected chunk length 3, got %d", len(chunks[0]))
	}
}

func TestChunk_Empty(t *testing.T) {
	chunks := Chunk([]int{}, 5)

	if chunks != nil {
		t.Fatalf("expected nil, got %v", chunks)
	}
}

func TestChunk_SizeOne(t *testing.T) {
	items := []int{1, 2, 3, 4}
	chunks := Chunk(items, 1)

	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if len(c) != 1 {
			t.Fatalf("chunk %d: expected length 1, got %d", i, len(c))
		}
		if c[0] != items[i] {
			t.Fatalf("chunk %d: expected %d, got %d", i, items[i], c[0])
		}
	}
}

func TestChunk_PanicsOnZeroSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for size 0")
		}
	}()
	Chunk([]int{1}, 0)
}
