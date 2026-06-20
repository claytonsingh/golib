package syncsignal

import "sync"

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
	trigger func()
}

// Signal sets the primitive to the signaled state and releases every goroutine
// blocked in Wait. If already signaled, Signal has no effect.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
func (this *OnceSignal) Signal() {
	if this.trigger != nil {
		this.trigger()
	}
}

// Wait blocks the calling goroutine until signaled. If already signaled, Wait
// returns immediately.
//
// Thread-safe: This method can be called from multiple goroutines concurrently.
func (this *OnceSignal) Wait() {
	this.mu.Lock()
	defer this.mu.Unlock()
}

// NewOnceSignal initializes a new OnceSignal in the nonsignaled state. Wait
// blocks until Signal is called.
func NewOnceSignal() *OnceSignal {
	this := &OnceSignal{}
	this.mu.Lock()
	this.trigger = sync.OnceFunc(this.mu.Unlock)
	return this
}
