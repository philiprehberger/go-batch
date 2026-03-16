# go-batch

[![CI](https://github.com/philiprehberger/go-batch/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-batch/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-batch.svg)](https://pkg.go.dev/github.com/philiprehberger/go-batch)
[![License](https://img.shields.io/github/license/philiprehberger/go-batch)](LICENSE)

Batch processing and chunking utilities for Go. Generic, concurrent, zero dependencies.

## Installation

```bash
go get github.com/philiprehberger/go-batch
```

## Usage

### Chunking

```go
import "github.com/philiprehberger/go-batch"

records := []Record{...} // 1000 records
for _, chunk := range batch.Chunk(records, 100) {
    db.BulkInsert(chunk) // insert 100 at a time
}
```

### Concurrent Processing

```go
import "github.com/philiprehberger/go-batch"

err := batch.Process(ctx, userIDs, 50, func(ctx context.Context, ids []int) error {
    return sendNotifications(ctx, ids)
}, batch.WithWorkers(4))
```

### Auto-flushing Accumulator

```go
import "github.com/philiprehberger/go-batch"

acc := batch.NewAccumulator(func(events []Event) {
    publishEvents(events)
}, batch.FlushSize[Event](100), batch.FlushInterval[Event](5*time.Second))

for event := range incomingEvents {
    acc.Add(event)
}
acc.Stop()
```

## API

| Function | Description |
|----------|-------------|
| `Chunk[T](items, size)` | Split slice into batches |
| `Process[T](ctx, items, size, fn, opts...)` | Concurrent batch processing |
| `WithWorkers(n)` | Set number of concurrent workers |
| `NewAccumulator[T](fn, opts...)` | Auto-flushing item accumulator |
| `FlushSize[T](n)` | Flush when N items accumulated |
| `FlushInterval[T](d)` | Flush on time interval |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
