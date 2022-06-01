package timap

import (
	"context"
	"runtime"
	"sync"
)

// CtxMap stores provided key/value pairs with context.
type CtxMap interface {
	Load(key interface{}) (value interface{}, ok bool)
	Store(ctx context.Context, key, value interface{})
	Delete(key interface{})
	Range(f func(key, value interface{}) bool)
}

type contextMap struct {
	// actual map
	*sync.Map
	// contains cancel funcs for watching contexts
	cancels *sync.Map
}

// NewCtxMap create a new map that stores key/value pairs with provided context.
// The pair is available until parent context is cancelled, or its deadline exceeds.
func NewCtxMap() CtxMap {
	return &contextMap{
		Map:     &sync.Map{},
		cancels: &sync.Map{},
	}
}

func (cm *contextMap) Store(parent context.Context, key, value interface{}) {
	ctx, cancel := context.WithCancel(parent)

	if cancelFunc, ok := cm.cancels.Load(key); ok {
		cancelFunc.(context.CancelFunc)() // stop previous watcher
	}

	cm.Map.Store(key, value)      // store actual key-value pair
	cm.cancels.Store(key, cancel) // store cancel func
	// add waiting goroutine for current pair
	go func() {
		<-ctx.Done()
		cm.Map.Delete(key)
		cm.cancels.Delete(key)
	}()
	// yields the processor, allowing other goroutines to run
	// it gives a chance to a goroutine above to be started before we exit Sore() func
	// which can be useful adding pairs with already cancelled context
	runtime.Gosched()
}

// Delete the value for a key calling cancel and not waiting for its context to be cancelled/exceeded.
func (cm *contextMap) Delete(key interface{}) {
	// check if there is a waiting context, do not remove the value explicitly
	// cancel the context and let its goroutine exit deleting the value
	if cancelFunc, ok := cm.cancels.Load(key); ok {
		cancelFunc.(context.CancelFunc)() // cancel previous context

		runtime.Gosched()
	}
}
