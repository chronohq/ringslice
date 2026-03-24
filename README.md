# Ringslice

[![go workflow](https://github.com/chronohq/ringslice/actions/workflows/go.yml/badge.svg)](https://github.com/chronohq/ringslice/actions/workflows/go.yml)
[![go reference](https://pkg.go.dev/badge/github.com/chronohq/ringslice.svg)](https://pkg.go.dev/github.com/chronohq/ringslice)
[![mit license](https://img.shields.io/badge/license-MIT-green)](/LICENSE)

Ringslice is a type-safe generic ring buffer backed by a Go slice, featuring `iter.Seq` iteration and zero-allocation performance. It includes lifecycle callbacks designed for production use cases.

For a traditional circular doubly-linked list, see Go's standard [container/ring](https://pkg.go.dev/container/ring) package.

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

## Concurrency Model

Ringslice uses a read-write lock to allow multiple concurrent readers while serializing writers.
This ensures thread-safety while maintaining high throughput for read-heavy workloads.
