# task

A lightweight generic async task type for Go, similar to a future or promise.

This library provides a `Task[T]` handle for work started in a background goroutine. Callers can block for the result with `Await`, poll or time out with `Wait`, register completion callbacks with `Link`, and iterate over many tasks in completion order with `UnorderedIterate*`.

This package was written as a fun challenge to implement the full async task API without using channels. `Await` and `Wait` block with a mutex gate rather than a semaphore or `sync.Cond` — a lock is lighter weight for a one-shot wait (no channel or permit counter). `sync.Cond` is used only in `UnorderedIterate*`, where a single consumer must wake on many completion signals.

## Features

* **Generic:** Works with any result type using Go generics.
* **Simple API:** `Async` to start work, `Await` to get the result.
* **Completion callbacks:** `Link` registers handlers that run when a task finishes; handles can be cancelled with `Unlink`.
* **Non-blocking poll and timeout:** `Wait` returns immediately when `timeoutMs` is zero, or blocks up to a deadline otherwise.
* **Completion-order iteration:** `UnorderedIterateSlice`, `UnorderedIterateAny`, and `UnorderedIterateLinkers` yield indices as tasks finish, using Go's `iter.Seq`.
* **Optional error wrapping:** `WrapAwaitFault` can transform errors at the `Await` site for stack-trace or diagnostics tooling.

## API Reference

### Core Functions

* **`Async[T](fn func() (T, error)) *Task[T]`:** Runs `fn` in a new goroutine and returns a task handle immediately.
* **`Await() (T, error)`:** Blocks until the task completes and returns its value and error. Safe to call multiple times or from multiple goroutines. When `WrapAwaitFault` is set and the task failed, the returned error is passed through that hook; the stored task error is unchanged.
* **`Wait(timeoutMs int64) bool`:** Blocks until completion or timeout. Returns `true` if the task completed, `false` on timeout. A non-positive `timeoutMs` polls without blocking: `true` if already done, `false` otherwise.

### Completion Callbacks

* **`Link(cb func()) LinkHandle`:** Registers a callback to run when the task completes. If the task is already done, the callback runs immediately and returns a zero `LinkHandle`. Callbacks must not panic.
* **`LinkHandle.Unlink()`:** Cancels a pending registration before completion fires. After completion, unlink is a no-op.

### Iteration

* **`UnorderedIterateSlice[T](tasks []*Task[T]) iter.Seq[int]`:** Yields slice indices as tasks complete. Order is undefined. Nil entries are skipped.
* **`UnorderedIterateAny(tasks []any) (iter.Seq[int], error)`:** Same as above for a heterogeneous slice. Each non-nil item must implement `TaskLinker`.
* **`UnorderedIterateLinkers(tasks []TaskLinker) iter.Seq[int]`:** Same as above for a slice of `TaskLinker` values.

### Optional Hook

* **`WrapAwaitFault atomic.Pointer[func(error) error]`:** Optional hook applied when `Await` returns a non-nil error. Store with `WrapAwaitFault.Store(&fn)` (for example once during `init` in development or staging). Loaded on each `Await` call.

## Thread Safety

All methods on `Task[T]` are safe to call from multiple goroutines concurrently:

* `Async`: Each call returns an independent task.
* `Await`, `Wait`, `Link`, and `Unlink`: Safe to call concurrently on the same task.
* `UnorderedIterate*`: Safe to use while tasks are still running.

## Important Considerations

* **Link callbacks must not panic.** A panicking callback can leave internal state inconsistent.
* **`WrapAwaitFault` is await-site only.** It transforms the error returned from `Await`, not the error stored on the task. Pair with throw-site wrapping inside the `Async` function for full async fault chains. The hook is loaded atomically on each `Await` call.
* **`Wait` never returns an error.** Use `Await` when you need the result or failure reason.
* **Completion order is undefined.** `UnorderedIterate*` yields indices as tasks finish, not in slice order.
* **Nil task slots are skipped.** Typed-nil pointers and nil interfaces in a task slice are ignored by the iterate helpers.

## Example Usage

### Basic async work

```go
package main

import (
	"fmt"
	"time"

	"github.com/claytonsingh/golib/task"
)

func main() {
	tk := task.Async(func() (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "hello", nil
	})

	v, err := tk.Await()
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}
```

### Wait with timeout

```go
tk := task.Async(func() (int, error) {
	time.Sleep(2 * time.Second)
	return 1, nil
})

if tk.Wait(500) {
	fmt.Println("done")
} else {
	fmt.Println("timed out")
}
```

### Process tasks in slice order

All tasks still run concurrently once started with `Async`. A plain range over the slice awaits each handle in order, so you observe results in slice order rather than completion order.

```go
tasks := []*task.Task[int]{
	task.Async(func() (int, error) { return fetch(0) }),
	task.Async(func() (int, error) { return fetch(1) }),
	task.Async(func() (int, error) { return fetch(2) }),
}

for _, task := range tasks {
	v, err := task.Await()
	if err != nil {
		// handle error
		continue
	}
	fmt.Println(v)
}
```

### Process tasks as they complete

```go
tasks := []*task.Task[int]{
	task.Async(func() (int, error) { return fetch(0) }),
	task.Async(func() (int, error) { return fetch(1) }),
	task.Async(func() (int, error) { return fetch(2) }),
}

for idx := range task.UnorderedIterateSlice(tasks) {
	v, err := tasks[idx].Await()
	if err != nil {
		// handle error for tasks[idx]
		continue
	}
	fmt.Printf("task %d -> %d\n", idx, v)
}
```

### Completion callback

```go
tk := task.Async(func() (int, error) {
	return doWork()
})

hook := tk.Link(func() {
	fmt.Println("work finished")
})

// Cancel before completion if no longer needed:
// hook.Unlink()

_, _ = tk.Await()
_ = hook // zero handle if callback already ran
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
