// Copyright Chrono Technologies LLC
// SPDX-License-Identifier: MIT

// Package ringslice provides a generic ring buffer backed by a Go slice,
// with iter.Seq iteration and callback hooks for production use cases.
package ringslice

import (
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
