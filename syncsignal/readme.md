# syncsignal

A lightweight and efficient signaling mechanism for goroutines in Go.

This library provides a `Signal` type that allows multiple goroutines to wait for a signal to be emitted. It's similar in concept to a condition variable, but offers a simpler interface for scenarios where you only need to be notified of state changes, not necessarily every individual signal event.

## Features

* **Efficient:** Designed for minimal overhead, especially in scenarios with frequent signals.
* **Simple API:** Easy to use with `NewSignal()`, `Signal()`, and `GetWaiter()`.
* **State-based Signaling:** Optimized for signaling state changes, where only the latest state is relevant.
* **Multiple Waiters:** Supports multiple goroutines waiting on the same signal.

## Explanation

* **`NewSignal()`:** Creates a new `Signal` instance.
* **`Signal()`:** Emits a signal, waking up waiting goroutines.
* **`GetWaiter(signaled bool)`:** Returns a function (closure) that waits for a signal.
    * If `signaled` is `false`, the returned `wait()` function will block until a new signal is emitted.
    * If `signaled` is `true`, the returned `wait()` function starts signaled and will *not* block on its *first* call.  It will proceed immediately, effectively "skipping" the first potential wait. This can be useful if you want a goroutine to check the current state without necessarily blocking initially. Subsequent calls to `wait()` *will* block until a new signal is emitted.
* **`wait()`:** The returned function from `GetWaiter()`. When called, it blocks until a new signal is emitted. Critically, `wait()` will always let a goroutine through *at least once* after `Signal()` has been called, even if there are multiple calls to `Signal()` in quick succession.

## Important Considerations

* **Skipping Intermediate Signals:** If `Signal()` is called multiple times before a goroutine calls `wait()`, the goroutine will only be awakened *once*, effectively *skipping* the intermediate signals.  This is by design and makes the `syncsignal` package suitable for state-change signaling where only the most recent state is important.  The `wait()` function ensures that a goroutine will always proceed after at least one call to `Signal()`. If you need to process every signal, consider using channels instead.
* **Multiple Waiters:**  Multiple goroutines can call `GetWaiter()` to obtain their own `wait()` functions and concurrently wait on the same `Signal` instance.  When `Signal()` is called, *all* waiting goroutines will be awakened.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
