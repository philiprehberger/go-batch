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

func TestChunkBy_GroupsByKey(t *testing.T) {
	type item struct {
		Category string
		Value    int
	}

	items := []item{
		{"a", 1}, {"b", 2}, {"a", 3}, {"c", 4}, {"b", 5},
	}

	groups := ChunkBy(items, func(i item) string {
		return i.Category
	})

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if len(groups["a"]) != 2 {
		t.Fatalf("expected 2 items in group 'a', got %d", len(groups["a"]))
	}
	if len(groups["b"]) != 2 {
		t.Fatalf("expected 2 items in group 'b', got %d", len(groups["b"]))
	}
	if len(groups["c"]) != 1 {
		t.Fatalf("expected 1 item in group 'c', got %d", len(groups["c"]))
	}
}

func TestChunkBy_Empty(t *testing.T) {
	groups := ChunkBy([]int{}, func(i int) int { return i })

	if groups != nil {
		t.Fatalf("expected nil for empty input, got %v", groups)
	}
}

func TestChunkBy_SingleGroup(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	groups := ChunkBy(items, func(i int) string { return "all" })

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups["all"]) != 5 {
		t.Fatalf("expected 5 items, got %d", len(groups["all"]))
	}
}

func TestChunkBy_IntKeys(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6}
	groups := ChunkBy(items, func(i int) int { return i % 2 })

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups[0]) != 3 {
		t.Fatalf("expected 3 even items, got %d", len(groups[0]))
	}
	if len(groups[1]) != 3 {
		t.Fatalf("expected 3 odd items, got %d", len(groups[1]))
	}
}
