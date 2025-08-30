# syncsignal

A lightweight and efficient signaling mechanism for goroutines in Go.

This library provides a `Signal` type that allows multiple goroutines to wait for a signal to be emitted. It's similar in concept to a condition variable, but offers a simpler interface for scenarios where you only need to be notified of state changes, not necessarily every individual signal event.

## Features

* **Efficient:** Designed for minimal overhead, especially in scenarios with frequent signals.
* **Simple API:** Easy to use with `NewSignal()`, `Signal()`, `GetWaiter()`, and `Close()`.
* **State-based Signaling:** Optimized for signaling state changes, where only the latest state is relevant.
* **Multiple Waiters:** Supports multiple goroutines waiting on the same signal.

## API Reference

### Core Functions

* **`NewSignal()`:** Creates a new `Signal` instance ready for use.
* **`Signal()`:** Emits a signal, waking up all waiting goroutines. If the Signal has been closed, calling this method has no effect.
* **`Close()`:** Closes the Signal instance, preventing further signals from being processed. All waiting goroutines will be awakened and their wait functions will return `true` indicating the Signal has been closed.
* **`GetWaiter(signaled bool)`:** Returns a function (closure) that waits for a signal.
    * If `signaled` is `false`, the returned `wait()` function will block until a new signal is emitted.
    * If `signaled` is `true`, the returned `wait()` function starts signaled and will *not* block on its *first* call. It will proceed immediately, effectively "skipping" the first potential wait. This can be useful if you want a goroutine to check the current state without necessarily blocking initially. Subsequent calls to `wait()` *will* block until a new signal is emitted.
* **`wait()`:** The returned function from `GetWaiter()`. When called, it blocks until a new signal is emitted. The function returns `false` if the Signal was closed, `true` otherwise. Critically, `wait()` will always let a goroutine through *at least once* after `Signal()` has been called, even if there are multiple calls to `Signal()` in quick succession.

## Thread Safety

All methods in the `Signal` type are thread-safe and can be called from multiple goroutines concurrently:
- `NewSignal()`: Safe to call from multiple goroutines
- `Signal()`: Safe to call from multiple goroutines
- `Close()`: Safe to call from multiple goroutines (multiple calls are safe but have no additional effect)
- `GetWaiter()`: Safe to call from multiple goroutines (each call returns a new, independent wait function)

## Important Considerations

* **Skipping Intermediate Signals:** If `Signal()` is called multiple times before a goroutine calls `wait()`, the goroutine will only be awakened *once*, effectively *skipping* the intermediate signals. This is by design and makes the `syncsignal` package suitable for state-change signaling where only the most recent state is important. The `wait()` function ensures that a goroutine will always proceed after at least one call to `Signal()`. If you need to process every signal, consider using channels instead.

* **Multiple Waiters:** Multiple goroutines can call `GetWaiter()` to obtain their own `wait()` functions and concurrently wait on the same `Signal` instance. When `Signal()` is called, *all* waiting goroutines will be awakened.

* **Resource Cleanup:** Always call `Close()` when you're done with a Signal to ensure proper cleanup. This will wake up any remaining waiting goroutines and prevent memory leaks.

* **Closed Signal Behavior:** Once a Signal is closed, calling `Signal()` will have no effect, and all wait functions will return `false` indicating the Signal has been closed.

* **Signal and Close Sequence:** When `Signal()` is called followed by `Close()`, existing waiters will first be awakened and their wait functions will return `true` (indicating a signal was received). Any subsequent calls to wait functions will return `false` (indicating the Signal has been closed). This ensures that waiters don't miss the final signal before closure.

## Example Usage

### Config File Reloading

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
	signal := syncsignal.NewSignal()
	var wg sync.WaitGroup
	
	// Start two components that wait for config changes
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			wait := signal.GetWaiter(true)
			
			fmt.Printf("Component %d: waiting...\n", id)
			for wait() {
				time.Sleep(50 * time.Millisecond)
				fmt.Printf("Component %d: config updated\n", id)
			}
			fmt.Printf("Component %d: shutting down\n", id)
		}(i)
	}
	
	time.Sleep(100 * time.Millisecond)
	signal.Signal()
	
	time.Sleep(100 * time.Millisecond)
	signal.Signal()
	
	time.Sleep(100 * time.Millisecond)
	signal.Close()
	
	wg.Wait()
}
```

### Example Output

```
Component 1: waiting...
Component 2: waiting...
Component 1: config updated
Component 2: config updated
Component 1: config updated
Component 2: config updated
Component 1: shutting down
Component 2: shutting down
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
