// Package syncsignal provides a lightweight and efficient signaling mechanism
// for goroutines in Go. It's similar to a condition variable but with a simpler
// interface for scenarios focused on state change notifications.
package syncsignal

import (
	"sync"
)

// Signal is a state-based synchronization primitive that allows multiple goroutines
// to wait for notifications of state changes.
type Signal struct {
	cond   *sync.Cond
	value  uint64
	closed bool
}

// NewSignal creates and returns a new Signal instance.
//
// Returns:
//   - *Signal: A new initialized Signal instance.
func NewSignal() *Signal {
	return &Signal{
		cond:   sync.NewCond(&sync.Mutex{}),
		closed: false,
	}
}

// Signal wakes up all goroutines that are waiting on this Signal instance.
// If the Signal has been closed, calling Signal() will have no effect.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
func (this *Signal) Signal() {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()

	if !this.closed {
		this.value += 1
	}
	this.cond.Broadcast()
}

// Close closes the Signal instance, preventing any further signals from being processed.
// All waiting goroutines will be awakened and their wait functions will return true
// indicating the Signal has been closed. After closing, calling Signal() will have no effect.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
// Multiple calls to Close() are safe but have no additional effect.
func (this *Signal) Close() {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()

	if !this.closed {
		this.closed = true
		this.cond.Broadcast()
	}
}

// GetWaiter returns a wait function that blocks until a signal occurs.
// The returned function can be used to wait for the next signal or check if the
// Signal has been closed.
//
// Parameters:
//   - signaled bool: If true, the returned wait function will not block on its first call.
//     If false, the wait function will block until a new signal is emitted.
//
// Returns:
//   - func() bool: A function that, when called, blocks until a new signal is received.
//     This function always proceeds at least once after Signal() has been called.
//     Returns false if the Signal was closed, true otherwise.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
// Each call returns a new, independent wait function.
func (this *Signal) GetWaiter(signaled bool) func() bool {
	var value uint64
	this.cond.L.Lock()
	value = this.value
	this.cond.L.Unlock()

	if signaled {
		value -= 1
	}

	return func() bool {
		this.cond.L.Lock()
		defer this.cond.L.Unlock()
		for {
			if value != this.value {
				value = this.value
				return true
			}
			if this.closed {
				return false
			}
			this.cond.Wait()
		}
	}
}
