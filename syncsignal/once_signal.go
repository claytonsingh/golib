package syncsignal

import (
	"sync"
	"sync/atomic"
)

// OnceSignal is a lightweight one-shot goroutine synchronization primitive.
// Signal unblocks all current and future Wait calls. Once signaled, Signal has
// no further effect. Reset is not supported.
//
// Remarks:
//
// OnceSignal is useful when a fixed lifecycle milestone must be observed exactly
// once, such as initialization complete or shutdown started. After Signal, every
// subsequent Wait returns immediately.
//
// The zero value is already signaled. Use NewOnceSignal to start nonsignaled.
//
// Thread Safety:
//
// This type is safe for concurrent use.
type OnceSignal struct {
	mu      sync.Mutex
	gate    sync.Mutex
	pending atomic.Bool
}

// Signal sets the primitive to the signaled state and releases every goroutine
// blocked in Wait. If already signaled, Signal has no effect.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
func (this *OnceSignal) Signal() {
	if this.pending.Load() {
		this.mu.Lock()
		defer this.mu.Unlock()
		if this.pending.Load() {
			// Make sure to unlock the gate before storing false. Otherwise signal
			// could return on another goroutine before the gate is unlocked.
			this.gate.Unlock()
			this.pending.Store(false)
		}
	}
}

// Wait blocks the calling goroutine until signaled. If already signaled, Wait
// returns immediately.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
func (this *OnceSignal) Wait() {
	if this.pending.Load() {
		this.gate.Lock()
		defer this.gate.Unlock()
	}
}

// NewOnceSignal initializes a new OnceSignal in the nonsignaled state. Wait
// blocks until Signal is called.
func NewOnceSignal() *OnceSignal {
	this := &OnceSignal{}
	this.gate.Lock()
	this.pending.Store(true)
	return this
}
