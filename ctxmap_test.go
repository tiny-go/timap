package timap

import (
	"context"
	"testing"
)

// the existence of this private method makes sense for testing only.
func (tm *contextMap) total() (total int) {
	tm.cancels.Range(func(key interface{}, value interface{}) bool {
		total++

		return true
	})

	return
}

func Test_CtxMap(t *testing.T) {
	t.Run("Add temporrary key-value pair with cancelled context", func(t *testing.T) {
		m := NewCtxMap()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel the context immediately

		m.Store(ctx, "foo", "bar")

		if total := m.(*contextMap).total(); total != 0 {
			t.Errorf(`should not have conntain any vaslues but has %d`, total)
		}

		_, ok := m.Load("delete me")
		if ok {
			t.Error(`key "delete me" should not exist`)
		}
	})

	t.Run("Add temporrary key-value pair that has to be checked", func(t *testing.T) {
		m := NewCtxMap()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		m.Store(ctx, "foo", "bar")

		if total := m.(*contextMap).total(); total != 1 {
			t.Errorf(`should have one watcher but has %d`, total)
		}

		d, ok := m.Load("foo")
		if !ok {
			t.Error(`expected key was not found`)
		}

		if d != "bar" {
			t.Error(`wrong value for a key"`)
		}
	})

	t.Run("Replace key-value pair with a new context", func(t *testing.T) {
		m := NewCtxMap()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		m.Store(context.Background(), "foo", "first pair")
		m.Store(context.Background(), "foo", "second pair")
		m.Store(ctx, "foo", "third pair")

		if total := m.(*contextMap).total(); total != 1 {
			t.Errorf(`should have one watcher but has %d`, total)
		}

		d, ok := m.Load("foo")
		if !ok {
			t.Error(`expected key was not found`)
		}

		if d != "third pair" {
			t.Error(`wrong value for a key`)
		}
	})

	t.Run("Manually delete key-value pair", func(t *testing.T) {
		m := NewCtxMap()

		m.Store(context.Background(), "foo", "bar")

		if total := m.(*contextMap).total(); total != 1 {
			t.Errorf(`should have one watcher but has %d`, total)
		}

		d, ok := m.Load("foo")
		if !ok {
			t.Error(`expected key was not found`)
		}

		if d != "bar" {
			t.Error(`wrong value for a key`)
		}

		m.Delete("foo")

		if total := m.(*contextMap).total(); total != 0 {
			t.Errorf(`should have no elements but got %d`, total)
		}
	})
}
