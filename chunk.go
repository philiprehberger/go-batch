// Package batch provides batch processing and chunking utilities for Go.
//
// It includes generic functions for splitting slices into chunks, processing
// batches concurrently with bounded parallelism, and an auto-flushing
// accumulator with size and time-based triggers.
package batch

// Chunk splits items into sub-slices of the given size. The last chunk may
// contain fewer than size elements. Chunk panics if size is less than or
// equal to zero.
func Chunk[T any](items []T, size int) [][]T {
	if size <= 0 {
		panic("batch: chunk size must be greater than zero")
	}

	if len(items) == 0 {
		return nil
	}

	chunks := make([][]T, 0, (len(items)+size-1)/size)
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}

	return chunks
}

// ChunkBy groups items by the value returned from the key function. Items
// sharing the same key are collected into the same slice. The iteration order
// of the returned map is not guaranteed.
func ChunkBy[T any, K comparable](items []T, key func(T) K) map[K][]T {
	if len(items) == 0 {
		return nil
	}

	groups := make(map[K][]T)
	for _, item := range items {
		k := key(item)
		groups[k] = append(groups[k], item)
	}

	return groups
}
