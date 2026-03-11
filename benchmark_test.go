// Copyright Chrono Technologies LLC
// SPDX-License-Identifier: MIT

package ringslice

import (
	"testing"
)

func BenchmarkAdd(b *testing.B) {
	ring := New[int](1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ring.Add(i)
	}
}

func BenchmarkAddParallel(b *testing.B) {
	ring := New[int](1024)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ring.Add(i)
			i++
		}
	})
}
