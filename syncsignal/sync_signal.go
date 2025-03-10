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
	cond  *sync.Cond
	value uint64
}

// NewSignal creates and returns a new Signal instance.
//
// Returns:
//   - *Signal: A new initialized Signal instance.
func NewSignal() *Signal {
	return &Signal{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

// Signal wakes up all goroutines that are waiting on this Signal instance.
func (this *Signal) Signal() {
	this.cond.L.Lock()
	this.value += 1
	this.cond.Broadcast()
	this.cond.L.Unlock()
}

// GetWaiter returns a wait function that blocks until a signal occurs.
//
// Parameters:
//   - signaled bool: If true, the returned wait function will not block on its first call.
//     If false, the wait function will block until a new signal is emitted.
//
// Returns:
//   - func(): A function that, when called, blocks until a new signal is received.
//     This function always proceeds at least once after Signal() has been called.
func (this *Signal) GetWaiter(signaled bool) func() {
	var value uint64
	this.cond.L.Lock()
	value = this.value
	this.cond.L.Unlock()
	if signaled {
		value -= 1
	}
	return func() {
		this.cond.L.Lock()
		for {
			if value != this.value {
				value = this.value
				break
			}
			this.cond.Wait()
		}
		this.cond.L.Unlock()
	}
}
