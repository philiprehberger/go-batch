# Changelog

## 0.2.0

- Add `ProcessWithErrors` for batch processing that returns individual errors as a slice
- Add `ChunkBy` for grouping items into a map by key function
- Add `Accumulator.Stats()` returning flush count, total items, and pending count
- Add `Accumulator.Peek()` to inspect buffered items without flushing
- Add `OnFlush` option for post-flush callback on Accumulator

## 0.1.3

- Consolidate README badges onto single line, fix CHANGELOG format

## 0.1.2

- Add Development section to README

## 0.1.0

- Initial release
- Generic chunk/batch slicing
- Concurrent batch processing with controlled parallelism
- Auto-flushing accumulator with size and time triggers
