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
