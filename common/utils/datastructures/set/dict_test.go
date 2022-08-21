package set

import (
	"testing"
	"strconv"
)

func TestExists(t *testing.T) {
	set := New()
	set.Add(`test`)

	if !set.Exists(`test`) {
		t.Errorf(`Correct existence not determined`)
	}

	if set.Exists(`test1`) {
		t.Errorf(`Correct nonexistence not determined.`)
	}
}

func TestExists_WithNewItems(t *testing.T) {
	set := New(`test`, `test1`)

	if !set.Exists(`test`) {
		t.Errorf(`Correct existence not determined`)
	}

	if !set.Exists(`test1`) {
		t.Errorf(`Correct existence not determined`)
	}

	if set.Exists(`test2`) {
		t.Errorf(`Correct nonexistence not determined.`)
	}
}

func TestLen(t *testing.T) {
	set := New()
	set.Add(`test`)

	if set.Len() != 1 {
		t.Errorf(`Expected len: %d, received: %d`, 1, set.Len())
	}

	set.Add(`test1`)
	if set.Len() != 2 {
		t.Errorf(`Expected len: %d, received: %d`, 2, set.Len())
	}
}

func TestClear(t *testing.T) {
	set := New()
	set.Add(`test`)

	set.Clear()

	if set.Len() != 0 {
		t.Errorf(`Expected len: %d, received: %d`, 0, set.Len())
	}
}

func BenchmarkLen(b *testing.B) {
	set := New()
	for i := 0; i < 50; i++ {
		item := strconv.Itoa(i)
		set.Add(item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Len()
	}
}

func BenchmarkExists(b *testing.B) {
	set := New()
	set.Add(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Exists(1)
	}
}

func BenchmarkClear(b *testing.B) {
	set := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Clear()
	}
}