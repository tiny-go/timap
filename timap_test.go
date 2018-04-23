package timap

import (
	"testing"
	"time"
)

// the existence of this private method makes sense for testing only.
func (tm *timap) watchers() (total int) {
	tm.timers.Range(func(key interface{}, value interface{}) bool {
		total++
		return true
	})
	return
}

func Test_Timap(t *testing.T) {
	// create new Timap with unlimited life time for key-value pair by default
	m := New(0)

	t.Run("Add temporrary key-value pair that has to be deleted", func(t *testing.T) {
		m.Store("delete me", "as soon as possible", 10*time.Millisecond)
		if total := m.(*timap).watchers(); total != 1 {
			t.Errorf(`should have one watcher but has %d`, total)
		}
		d, ok := m.Load("delete me")
		if !ok {
			t.Error(`should already have key "delete me"`)
		}
		if d != "as soon as possible" {
			t.Error(`wrong value for key "delete me"`)
		}
	})

	t.Run("Delete temporrary key-value pair and its watcher", func(t *testing.T) {
		m.Delete("delete me")
		if total := m.(*timap).watchers(); total != 0 {
			t.Errorf(`should not have any watchers but has %d`, total)
		}
		d, ok := m.Load("delete me")
		if ok {
			t.Error(`value with key "delete me" should have been deleted`)
		}
		if d != nil {
			t.Errorf(`shoul be nil but has value "%v"`, d)
		}
	})

	t.Run("Add first temporrary key-value pair", func(t *testing.T) {
		m.Store("a", 42, 10*time.Millisecond)
		if total := m.(*timap).watchers(); total != 1 {
			t.Errorf(`should have one watcher but has %d`, total)
		}
		a, ok := m.Load("a")
		if !ok {
			t.Error(`should already have key "a"`)
		}
		if a != 42 {
			t.Error(`wrong value for "a"`)
		}
	})

	t.Run("Add second temporrary key-value pair", func(t *testing.T) {
		m.Store("b", "second", 80*time.Millisecond)
		if total := m.(*timap).watchers(); total != 2 {
			t.Errorf(`should have two watchers but has %d`, total)
		}
		b, ok := m.Load("b")
		if !ok {
			t.Error(`should already have key "b"`)
		}
		if b != "second" {
			t.Error(`wrong value for "b"`)
		}
	})

	t.Run("Add third persistent key-value pair", func(t *testing.T) {
		m.Store("c", [42]int{})
		if m.(*timap).watchers() != 2 {
			t.Error(`should not add any watchers`)
		}
		c, ok := m.Load("c")
		if !ok {
			t.Error(`should already have key "c"`)
		}
		if c != [42]int{} {
			t.Error(`wrong value for "c"`)
		}
	})

	t.Run("Replace first key-value pair with another tepmorrary pair", func(t *testing.T) {
		m.Store("a", "new A", 50*time.Millisecond)
		if total := m.(*timap).watchers(); total != 2 {
			t.Errorf(`should still have two values but has %d`, total)
		}
		d, ok := m.Load("a")
		if !ok {
			t.Error(`should already have key "a"`)
		}
		if d != "new A" {
			t.Error(`wrong new value for "a"`)
		}
	})

	t.Run("Check values after 30 milliseconds", func(t *testing.T) {
		time.Sleep(30 * time.Millisecond)
		if total := m.(*timap).watchers(); total != 2 {
			t.Errorf("should still have two watchers but has %d", total)
		}
		a, ok := m.Load("a")
		if !ok {
			t.Error(`should still have key "a" because its timer was redefined`)
		}
		if a != "new A" {
			t.Errorf(`"a" should have new value but has value "%v"`, a)
		}
		b, ok := m.Load("b")
		if !ok {
			t.Error(`should still have key "b"`)
		}
		if b != "second" {
			t.Error(`value of "b" should not be changed`)
		}
	})

	t.Run("Check values after 70 milliseconds", func(t *testing.T) {
		time.Sleep(40 * time.Millisecond)
		if total := m.(*timap).watchers(); total != 1 {
			t.Errorf("should have exactly one watcher but has %d", total)
		}
		a, ok := m.Load("a")
		if ok {
			t.Error(`"a" should have been expired and deleted`)
		}
		if a != nil {
			t.Errorf(`"a" should be nil but has value "%v"`, a)
		}
		b, ok := m.Load("b")
		if !ok {
			t.Error(`should still have key "b"`)
		}
		if b != "second" {
			t.Error(`value of "b" should not be changed`)
		}
	})

	t.Run("Check values after 100 milliseconds", func(t *testing.T) {
		time.Sleep(30 * time.Millisecond)
		if total := m.(*timap).watchers(); total != 0 {
			t.Errorf("should not have any watchers but has %d", total)
		}
		a, ok := m.Load("a")
		if ok {
			t.Error(`"a" should have never been expired or deleted`)
		}
		if a != nil {
			t.Errorf(`"a" should be nil but has value "%v"`, a)
		}
		b, ok := m.Load("b")
		if ok {
			t.Error(`"b" should have been expired and deleted`)
		}
		if b != nil {
			t.Errorf(`"b" should be nil but has value "%v"`, b)
		}
		c, ok := m.Load("c")
		if !ok {
			t.Error(`"c" should have never been expired`)
		}
		if c != [42]int{} {
			t.Error(`value of "c" should not be changed`)
		}
	})
}
