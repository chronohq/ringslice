// Copyright Chrono Technologies LLC
// SPDX-License-Identifier: MIT

// Package ringslice provides a generic ring buffer backed by a Go slice,
// with iter.Seq iteration and callback hooks for production use cases.
package ringslice

import (
	"encoding/json"
	"iter"
	"sync"
)

// Ring is a fixed-size ring buffer backed by a Go slice.
// When the buffer is full, new values overwrite the oldest.
type Ring[T any] struct {
	buf   []T
	count int
	idx   int
	mu    sync.RWMutex

	onBeforeAdd func(T) bool
	onFlush     func([]T)
	onRotate    func([]T)
}

// New returns a new Ring with the given capacity.
func New[T any](capacity int) *Ring[T] {
	return &Ring[T]{
		buf: make([]T, capacity),
	}
}

// OnBeforeAdd registers a callback invoked before a value is added.
// Returning false from fn prevents the value from being written.
// fn must not call any methods on the same Ring instance.
func (r *Ring[T]) OnBeforeAdd(fn func(T) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onBeforeAdd = fn
}

// OnRotate registers a callback invoked when the write index wraps back
// to zero. fn receives the underlying slice of the ring buffer, thus it
// must not mutate or retain the slice beyond the duration of the call.
// fn must not call any methods on the same Ring instance.
func (r *Ring[T]) OnRotate(fn func([]T)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onRotate = fn
}

// OnFlush registers a callback invoked by Flush() before the buffer is cleared.
// fn receives the underlying slice of the ring buffer, thus it must not mutate
// or retain the slice beyond the duration of the call. fn must not call any
// methods on the same Ring instance.
func (r *Ring[T]) OnFlush(fn func([]T)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onFlush = fn
}

// Add writes the given value to the ring.
func (r *Ring[T]) Add(val T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.onBeforeAdd != nil && !r.onBeforeAdd(val) {
		return
	}

	r.buf[r.idx] = val
	r.idx++

	if r.count < len(r.buf) {
		r.count++
	}

	// rotation
	if r.idx == len(r.buf) {
		r.idx = 0

		if r.onRotate != nil {
			r.onRotate(r.buf)
		}
	}
}

// Clear resets the ring buffer, discarding all elements.
func (r *Ring[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.resetBuffer()
}

// Flush resets the ring buffer, discarding all elements. If an OnFlush
// callback is registered, it is invoked before the buffer is cleared.
func (r *Ring[T]) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.onFlush != nil {
		r.onFlush(r.buf)
	}

	r.resetBuffer()
}

// Peek returns the most recently added item without consuming it.
// Returns the zero value and false if the ring buffer is empty.
func (r *Ring[T]) Peek() (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.count == 0 {
		return *new(T), false
	}

	// add len(r.buf) before the modulo to guard against when r.idx == 0
	idx := (r.idx - 1 + len(r.buf)) % len(r.buf)

	return r.buf[idx], true
}

// All returns an iterator that yields each element in chronological order.
func (r *Ring[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		start := r.startIdx()

		// use count to avoid yielding over empty slots
		for i := 0; i < r.count; i++ {
			idx := (start + i) % len(r.buf)

			if !yield(r.buf[idx]) {
				return
			}
		}
	}
}

// AllDesc returns an iterator that yields each element in reverse
// chronological order.
func (r *Ring[T]) AllDesc() iter.Seq[T] {
	return func(yield func(T) bool) {
		r.mu.RLock()
		defer r.mu.RUnlock()

		start := r.startIdx()

		for i := r.count - 1; i >= 0; i-- {
			idx := (start + i) % len(r.buf)
			if !yield(r.buf[idx]) {
				return
			}
		}
	}
}

// Len returns the number of elements in the ring.
func (r *Ring[T]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.count
}

// Cap returns the capacity of the ring.
func (r *Ring[T]) Cap() int {
	// read lock is not required but is held for consistency
	r.mu.RLock()
	defer r.mu.RUnlock()

	return cap(r.buf)
}

type ringJSON[T any] struct {
	Capacity int `json:"capacity"`
	Items    []T `json:"items"`
}

// MarshalJSON implements the json.Marshaler interface.
func (r *Ring[T]) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// determine the actual count of items in the ring
	n := min(r.count, len(r.buf))

	data := ringJSON[T]{
		Capacity: len(r.buf),
		Items:    make([]T, n),
	}

	start := r.startIdx()

	for i := range n {
		data.Items[i] = r.buf[(start+i)%len(r.buf)]
	}

	return json.Marshal(data)
}

func (r *Ring[T]) resetBuffer() {
	clear(r.buf)
	r.count = 0
	r.idx = 0
}

func (r *Ring[T]) rotated() bool {
	return r.count == len(r.buf)
}

func (r *Ring[T]) startIdx() int {
	if r.rotated() {
		return r.idx
	}

	return 0
}
