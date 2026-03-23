# go-batch

[![CI](https://github.com/philiprehberger/go-batch/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-batch/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-batch.svg)](https://pkg.go.dev/github.com/philiprehberger/go-batch) [![License](https://img.shields.io/github/license/philiprehberger/go-batch)](LICENSE)

Batch processing and chunking utilities for Go. Generic, concurrent, zero dependencies

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

### Error Handling

```go
import "github.com/philiprehberger/go-batch"

errs := batch.ProcessWithErrors(records, 100, 4, func(batch []Record) error {
    return db.BulkInsert(batch)
})
for _, err := range errs {
    log.Printf("batch failed: %v", err)
}
```

### ChunkBy

```go
import "github.com/philiprehberger/go-batch"

type Order struct {
    Status string
    ID     int
}

orders := []Order{{Status: "pending", ID: 1}, {Status: "shipped", ID: 2}, {Status: "pending", ID: 3}}
grouped := batch.ChunkBy(orders, func(o Order) string { return o.Status })
// grouped["pending"] = [{pending 1} {pending 3}]
// grouped["shipped"] = [{shipped 2}]
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

### Stats

```go
acc := batch.NewAccumulator(func(events []Event) {
    publishEvents(events)
}, batch.FlushSize[Event](100))

acc.Add(event1)
acc.Add(event2)

stats := acc.Stats()
fmt.Printf("flushes: %d, total: %d, pending: %d\n", stats.FlushCount, stats.TotalItems, stats.Pending)
```

### Peek

```go
acc := batch.NewAccumulator[Event](func(events []Event) {
    publishEvents(events)
}, batch.FlushSize[Event](100))

acc.Add(event1)
acc.Add(event2)

buffered := acc.Peek() // returns copy of buffered items without flushing
fmt.Println(len(buffered)) // 2
```

## API

| Function | Description |
|----------|-------------|
| `Chunk[T](items, size)` | Split slice into batches |
| `ChunkBy[T, K](items, key)` | Group items into map by key function |
| `Process[T](ctx, items, size, fn, opts...)` | Concurrent batch processing |
| `ProcessWithErrors[T](items, size, workers, fn)` | Concurrent processing, returns error slice |
| `WithWorkers(n)` | Set number of concurrent workers |
| `NewAccumulator[T](fn, opts...)` | Auto-flushing item accumulator |
| `FlushSize[T](n)` | Flush when N items accumulated |
| `FlushInterval[T](d)` | Flush on time interval |
| `OnFlush[T](fn)` | Callback invoked after each flush |
| `Accumulator.Stats()` | Returns flush count, total items, pending |
| `Accumulator.Peek()` | Returns copy of buffered items without flushing |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
