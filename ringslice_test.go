package ringslice

import (
	"slices"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		numElems int
		want     []int
	}{
		{
			name:     "with less elements than capacity",
			capacity: 8,
			numElems: 6,
			want:     []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "with elements equal capacity",
			capacity: 8,
			numElems: 8,
			want:     []int{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:     "with more elements than capacity",
			capacity: 8,
			numElems: 12,
			want:     []int{4, 5, 6, 7, 8, 9, 10, 11},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ring := New[int](test.capacity)

			// populate
			for i := range test.numElems {
				ring.Add(i)
			}

			if test.numElems < test.capacity {
				// less elements than capacity, thus Len() equals numElems
				if ring.Len() != test.numElems {
					t.Errorf("got: %d, want: %d", ring.Len(), test.numElems)
					return
				}
			} else {
				// full capacity or more elements, thus Len() equals capacity
				if ring.Len() != test.capacity {
					t.Errorf("got: %d, want: %d", ring.Len(), test.capacity)
					return
				}
			}

			var got []int

			for v := range ring.All() {
				got = append(got, v)
			}

			if !slices.Equal(got, test.want) {
				t.Errorf("got: %v, want: %v", got, test.want)
				return
			}
		})
	}
}

func TestAllDesc(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		numElems int
		want     []int
	}{
		{
			name:     "with less elements than capacity",
			capacity: 8,
			numElems: 6,
			want:     []int{5, 4, 3, 2, 1, 0},
		},
		{
			name:     "with elements equal capacity",
			capacity: 8,
			numElems: 8,
			want:     []int{7, 6, 5, 4, 3, 2, 1, 0},
		},
		{
			name:     "with more elements than capacity",
			capacity: 8,
			numElems: 12,
			want:     []int{11, 10, 9, 8, 7, 6, 5, 4},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ring := New[int](test.capacity)

			for i := range test.numElems {
				ring.Add(i)
			}

			var got []int

			for v := range ring.AllDesc() {
				got = append(got, v)
			}

			if !slices.Equal(got, test.want) {
				t.Errorf("got: %v, want: %v", got, test.want)
				return
			}
		})
	}
}

func TestClear(t *testing.T) {
	ring := New[int](8)

	for i := range ring.Cap() {
		ring.Add(i)
	}

	if ring.Len() != ring.Cap() {
		t.Errorf("got: %v, want: %v", ring.Len(), ring.Cap())
		return
	}

	ring.Clear()

	if ring.Len() != 0 {
		t.Errorf("got: %v, want: %v", ring.Len(), 0)
		return
	}

	want := 121
	ring.Add(want)

	if ring.buf[0] != want {
		t.Errorf("got: %v, want: %v", ring.buf[0], want)
		return
	}
}
