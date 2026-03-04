// Copyright Chrono Technologies LLC
// SPDX-License-Identifier: MIT

// Package ringslice provides a generic ring buffer backed by a Go slice,
// with iter.Seq iteration and callback hooks for production use cases.
package ringslice

import (
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
	onRotate    func()
}

// New returns a new Ring with the given capacity.
func New[T any](capacity int) *Ring[T] {
	return &Ring[T]{
		buf: make([]T, capacity),
	}
}

// SetOnBeforeAdd sets the callback invoked before a value is added.
// Returning false from fn prevents the value from being written.
func (r *Ring[T]) SetOnBeforeAdd(fn func(T) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onBeforeAdd = fn
}

// SetOnRotate sets the callback invoked when the write index wraps back to zero.
func (r *Ring[T]) SetOnRotate(fn func()) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.onRotate = fn
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
			r.onRotate()
		}
	}
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

func (r *Ring[T]) rotated() bool {
	return r.count == len(r.buf)
}

func (r *Ring[T]) startIdx() int {
	if r.rotated() {
		return r.idx
	}

	return 0
}
