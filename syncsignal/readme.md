# syncsignal

Synchronization primitives for coordinating goroutines in Go.

This package provides a state-based `BroadcastSignal` for change notifications and a one-shot `OnceSignal` for lifecycle milestones. All types are safe for concurrent use.

## Thread Safety

All methods on `BroadcastSignal` and `OnceSignal` are safe to call from multiple goroutines concurrently.

## Choosing a Primitive

| Type | Use when |
|------|----------|
| `BroadcastSignal` | Multiple goroutines need to react to state changes; intermediate signals can be coalesced. |
| `OnceSignal` | A lifecycle milestone happens exactly once (initialization complete, shutdown started). |

## BroadcastSignal

A lightweight signaling mechanism for goroutines. Similar in concept to a condition variable, but with a simpler interface for scenarios where only the latest state matters, not every individual signal.

### Features

* **Efficient:** Designed for minimal overhead, especially with frequent signals.
* **State-based:** Optimized for state-change notifications, where only the latest state is relevant.
* **Multiple waiters:** Supports many goroutines waiting on the same `BroadcastSignal`.
* **Graceful shutdown:** `Close()` releases waiters and stops further signaling.

### API

* **`NewBroadcastSignal()`:** Creates a new `BroadcastSignal` instance.
* **`Signal()`:** Wakes all waiting goroutines. Has no effect after `Close()`.
* **`Close()`:** Shuts down the `BroadcastSignal`. All blocked waiters are released. Idempotent.
* **`GetWaiter(signaled bool)`:** Returns a `wait()` function.
    * If `signaled` is `false`, the first `wait()` call blocks until a new signal is emitted.
    * If `signaled` is `true`, the first `wait()` call does not block. This is useful when a goroutine should do work immediately without waiting for the first signal. Subsequent `wait()` calls block until a new signal is emitted.
* **`wait()`:** The function returned by `GetWaiter()`.
    * Blocks until the `BroadcastSignal` is signaled or closed.
    * Returns `true` when a signal was received. The goroutine should react to the state change.
    * Returns `false` when the `BroadcastSignal` was closed. The goroutine should stop waiting, typically by exiting a loop.
    * **Guarantee:** After `Signal()` is called, `wait()` will return `true` at least once, even if `Signal()` was called multiple times in quick succession. Intermediate signals may be skipped; use channels if every signal must be processed individually.

### Important Considerations

* **Coalesced signals:** If `Signal()` is called multiple times before a goroutine calls `wait()`, intermediate signals are skipped. Each `wait()` still returns `true` at least once after signaling has occurred.

* **Multiple waiters:** Multiple goroutines can obtain their own `wait()` functions from the same `BroadcastSignal`. When `Signal()` is called, all blocked waiters are released.

* **Closed behavior:** Once a `BroadcastSignal` is closed, `Signal()` has no effect. A blocked `wait()` returns `false` when there is no outstanding signal to deliver.

* **Signal then Close:** When `Signal()` is followed by `Close()`, waiters observe the signal before closure: one `wait()` call returns `true`, and a subsequent `wait()` call returns `false`. Goroutines in a `for wait(); { ... }` loop therefore handle the final signal before shutting down.

* **Cleanup:** Call `Close()` when finished to release any remaining waiters.

### Example

A common use case for syncsignal is coordinating config file reloads across multiple goroutines:

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/claytonsingh/golib/syncsignal"
)

func main() {
	reload := syncsignal.NewBroadcastSignal()
	var wg sync.WaitGroup

	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Component %d: waiting...\n", id)
			for wait := reload.GetWaiter(true); wait(); {
				time.Sleep(50 * time.Millisecond)
				fmt.Printf("Component %d: config updated\n", id)
			}
			fmt.Printf("Component %d: shutting down\n", id)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	reload.Signal()

	time.Sleep(100 * time.Millisecond)
	reload.Close()

	wg.Wait()
}
```

## OnceSignal

A one-shot synchronization primitive. `Signal()` unblocks all current and future `Wait()` calls. Once signaled, further calls to `Signal()` have no effect and reset is not supported.

The zero value is already signaled. Use `NewOnceSignal()` to start in the nonsignaled state.

### API

* **`NewOnceSignal()`:** Creates a nonsignaled `OnceSignal`. `Wait()` blocks until `Signal()` is called.
* **`Signal()`:** Signals the primitive and releases every blocked waiter. Idempotent.
* **`Wait()`:** Blocks until signaled. Returns immediately if already signaled.

### Example

```go
ready := syncsignal.NewOnceSignal()

go func() {
	// ... initialize ...
	ready.Signal()
}()

ready.Wait()
// safe to proceed; all current and future Wait calls also return immediately
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
