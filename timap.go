package timap

import (
	"sync"
	"time"
)

// Timap represents "time map", it is a wrapper over sync.Map that allows to store
// key-value pairs for certain period of time. Default duration can be defined for
// all key-value pairs in the constructor func, but it is also possible to define
// custom life time for each pair separately. Watcher will create a goroutine for
// every temporrary pair. Try not to use it with huge maps.
type Timap interface {
	Load(key interface{}) (value interface{}, ok bool)
	Store(key, value interface{}, lifeTime ...time.Duration)
	Delete(key interface{})
	Range(f func(key, value interface{}) bool)
}

type timap struct {
	// actual map
	*sync.Map
	// contains stop channels for stored timers (for temporrary pairs only)
	timers *sync.Map
	// default lifetime for key-value pair
	defaultLifetime time.Duration
}

// New is a Timap constructor func. If lifeTime == 0 key-value pairs will not have
// any time limits (by default), values still can be deleted by calling Delete().
func New(lifeTime time.Duration) Timap {
	return &timap{
		Map:             &sync.Map{},
		timers:          &sync.Map{},
		defaultLifetime: lifeTime,
	}
}

// Store sets the value for a key (for certain period of time).
func (tm *timap) Store(key, value interface{}, lifeTime ...time.Duration) {
	// check if watcher for current key was registered
	if stopChan, ok := tm.timers.Load(key); ok {
		// send stop singnal to the previous watcher, otherwise new value will be
		// deleted by the timer from previous watcher
		stopChan.(chan struct{}) <- struct{}{}
		// remove watcher since the next one can have zero duration (unlimited)
		tm.timers.Delete(key)
	}
	// store actual key-value pair
	tm.Map.Store(key, value)
	// total life time for current pair
	var totalTime time.Duration
	// calculate total life time
	for _, duration := range lifeTime {
		totalTime += duration
	}
	// if no arguments were provided use default life time
	if totalTime == 0 {
		totalTime = tm.defaultLifetime
	}
	// if default life time == 0 - do not add any watchers (persistent pair)
	if totalTime == 0 {
		return
	}
	// stop channel is needed in order to quit goroutines
	stop := make(chan struct{}, 1)
	// store stop channel for the pair
	tm.timers.Store(key, stop)
	// create new timer per key-value pair
	timer := time.NewTimer(totalTime)
	// add watching goroutine for current pair
	go func() {
		// wait for timer or stop signal to exit goroutine
		select {
		// value is expired
		case <-timer.C:
			// remove expired pair
			tm.Map.Delete(key)
			// remove watcher
			tm.timers.Delete(key)
		// value is being deleted manually by calling Delete()
		case <-stop:
			// stop the timer and exit goroutine
			timer.Stop()
		}
	}()
}

// Delete deletes the value for a key (and stops watcher's goroutine for temporrary
// key-value pairs).
func (tm *timap) Delete(key interface{}) {
	// check if there is a watcher registered for current key
	if stopChan, ok := tm.timers.Load(key); ok {
		// send stop singnal to watchers goroutine
		stopChan.(chan struct{}) <- struct{}{}
		// remove watcher from the list
		tm.timers.Delete(key)
	}
	// remove actual value
	tm.Map.Delete(key)
}
