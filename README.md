# Ringslice

[![go workflow](https://github.com/chronohq/ringslice/actions/workflows/go.yml/badge.svg)](https://github.com/chronohq/ringslice/actions/workflows/go.yml)
[![go reference](https://pkg.go.dev/badge/github.com/chronohq/ringslice.svg)](https://pkg.go.dev/github.com/chronohq/ringslice)
[![mit license](https://img.shields.io/badge/license-MIT-green)](/LICENSE)

Ringslice is a type-safe generic ring buffer backed by a Go slice, designed for
production use cases like sliding window and streaming workloads.

Most ring buffer implementations focus on storage. Ringslice adds lifecycle hooks
so you can react to what happens as data moves through the buffer.
It also features `iter.Seq` iteration for idiomatic Go range support.

For pointer-based circular lists and round-robin traversal, see Go’s standard
[container/ring](https://pkg.go.dev/container/ring).

## Basic Usage

```go
const maxRingCapacity = 128

ring := ringslice.New[string](maxRingCapacity)

ring.Add("generic")
ring.Add("ring")
ring.Add("buffer")

// prints: "generic", "ring", "buffer"
for v := range ring.All() {
    fmt.Println(v)
}

// prints: "buffer", "ring", "generic"
for v := range ring.AllDesc() {
    fmt.Println(v)
}
```

## Hooks

Ringslice provides callback hooks that let you react to internal lifecycle events.

### OnBeforeAdd

Called before the value is added to the ring. The value will be rejected if `false` is returned.
This is particularly useful if you want to exclude values with certain characteristics.

```go
// reject empty strings
ring.OnBeforeAdd(func(value string) bool {
    return len(value) > 0
})
```

### OnRotate

Called each time the write index wraps around the ring. Useful for logging, flushing, or instrumentation.

```go
ring.OnRotate(func(values []string) {
    fmt.Println("lap complete 🏁")
})
```

### OnFlush

Called by `Flush()` before the buffer is cleared. Useful for draining the buffer, persisting elements, or instrumentation.

```go
ring.OnFlush(func(values []string) {
    for _, v := range values {
        db.Insert(v)
    }
})
```

Note: `Clear()` discards all elements without invoking this callback.

## Serialization

Ringslice implements the `json.Marshaler` interface, allowing you to serialize the buffer directly with `json.Marshal()`.

```go
ring := ringslice.New[int](8)

ring.Add(1)
ring.Add(2)
ring.Add(3)

buf, err := json.Marshal(ring)
```

This functionality is useful when you want to persist the ring buffer,
even though ringslice is inherently volatile. The serialized output
preserves insertion order regardless of internal buffer rotation.

## Concurrency Model

Ringslice uses a read-write lock to allow multiple concurrent readers while serializing writers.
This ensures thread-safety while maintaining high throughput for read-heavy workloads.
