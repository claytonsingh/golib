// Package syncsignal provides a lightweight and efficient signaling mechanism
// for goroutines in Go. It's similar to a condition variable but with a simpler
// interface for scenarios focused on state change notifications.
package syncsignal

import (
	"sync"
)

// BroadcastSignal is a state-based synchronization primitive that allows multiple goroutines
// to wait for notifications of state changes.
type BroadcastSignal struct {
	cond   *sync.Cond
	value  uint64
	closed bool
}

// Deprecated: Use BroadcastSignal instead.
type Signal = BroadcastSignal

// NewBroadcastSignal creates and returns a new BroadcastSignal instance.
//
// Returns:
//   - *BroadcastSignal: A new initialized BroadcastSignal instance.
func NewBroadcastSignal() *BroadcastSignal {
	return &BroadcastSignal{
		cond:   sync.NewCond(&sync.Mutex{}),
		closed: false,
	}
}

// NewSignal creates and returns a new Signal instance for backward compatibility.
//
// Deprecated: Use NewBroadcastSignal instead.
func NewSignal() *Signal {
	return NewBroadcastSignal()
}

// Signal wakes up all goroutines that are waiting on this BroadcastSignal instance.
// If the BroadcastSignal has been closed, calling Signal() will have no effect.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
func (this *BroadcastSignal) Signal() {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()

	if !this.closed {
		this.value += 1
	}
	this.cond.Broadcast()
}

// Close closes the BroadcastSignal instance, preventing any further signals from being processed.
// All waiting goroutines are released. A wait function returns false when the BroadcastSignal
// is closed, unless an outstanding signal still needs to be delivered. After closing,
// calling Signal() will have no effect.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
// Multiple calls to Close() are safe but have no additional effect.
func (this *BroadcastSignal) Close() {
	this.cond.L.Lock()
	defer this.cond.L.Unlock()

	if !this.closed {
		this.closed = true
		this.cond.Broadcast()
	}
}

// GetWaiter returns a wait function that blocks until a signal occurs.
// The returned function can be used to wait for the next signal or check if the
// BroadcastSignal has been closed.
//
// Parameters:
//   - signaled bool: If true, the returned wait function will not block on its first call.
//     If false, the wait function will block until a new signal is emitted.
//
// Returns:
//   - func() bool: A function that, when called, blocks until a new signal is received.
//     This function always proceeds at least once after Signal() has been called.
//     Returns false if the BroadcastSignal was closed, true otherwise.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
// Each call returns a new, independent wait function.
func (this *BroadcastSignal) GetWaiter(signaled bool) func() bool {
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
