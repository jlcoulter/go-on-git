package runner

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestMapPreservesOrder(t *testing.T) {
	items := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	got := Map(items, 4, func(i int) int {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		return i * 2
	})
	for i, v := range got {
		if v != i*2 {
			t.Errorf("got[%d] = %d, want %d", i, v, i*2)
		}
	}
}

func TestMapParallelismBounds(t *testing.T) {
	items := []int{1, 2, 3}
	if got := Map(items, 0, func(i int) int { return i }); len(got) != 3 {
		t.Errorf("parallelism 0: len = %d", len(got))
	}
	if got := Map(items, 100, func(i int) int { return i }); len(got) != 3 {
		t.Errorf("parallelism > len: len = %d", len(got))
	}
	if got := Map([]int{}, 4, func(i int) int { return i }); len(got) != 0 {
		t.Errorf("empty: len = %d", len(got))
	}
}

func TestMapOnDone(t *testing.T) {
	var calls atomic.Int32
	Map([]int{1, 2, 3}, 2, func(i int) int { return i }, func() { calls.Add(1) })
	if calls.Load() != 3 {
		t.Errorf("onDone calls = %d, want 3", calls.Load())
	}
}
