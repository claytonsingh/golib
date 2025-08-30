# leakypool

A generic, thread-safe, non-blocking object pool implementation in Go.

This library provides a `LeakyPool` type that manages a pool of reusable objects. The pool is "leaky" - when the pool is full, returned objects are discarded instead of blocking or growing the pool size. This design ensures the pool never blocks operations and maintains predictable memory usage.

## Features

* **Generic:** Works with any type using Go generics.
* **Thread-safe:** Safe for concurrent access from multiple goroutines.
* **Non-blocking:** Operations never block, ensuring predictable performance.
* **Simple API:** Easy to use with `NewLeakyPool()`, `Get()`, and `Return()`.
* **Resource Management:** Automatically closes discarded objects that implement `io.Closer`.

## How It Works

When you call `pool.Get()`:
- If the pool has available objects, one is returned immediately
- If the pool is empty, the factory function is called to create a new object

When you call `Ref.Return()` to return an object to the pool:
- If the pool has space, the object is added back to the pool for reuse.
- If the pool is full, the object is discarded. If the object implements the `io.Closer` interface, its `Close()` method is called to release any resources.
- After calling `Return()`, the reference should not be used again, as the object may have been reused or closed.
- If the object does not impliment `io.Closer` then `Ref.Return()` never returns an error.

When you call `pool.Close()`:
- The pool is closed and all objects currently in the pool are discarded. If the objects implement the `io.Closer` interface, their `Close()` method is called to release resources. If any `Close()` call returns an error, `pool.Close()` will forward that error (or a combined error if multiple objects fail to close).
- If the object does not impliment `io.Closer` then `pool.Close()` never returns an error.
- After calling `Close()` the capacity is set to zero. You can still call `Get()` but every call will invoke the facory and every `Ref.Return()` will discard the object.
- `Close()` may cause `Get()` to block while the internal state is reconfigured.

Thread Safety:
- The factory function should be thread-safe if the pool is used concurrently.
- `pool.Get()`, and `pool.Close()` are threadsafe.
- `Ref.Return()` is expected to only be called once.

## Usage Example

```go
package main

import (
    "fmt"
    "sync"
    "github.com/claytonsingh/golib/leakypool"
)

type MyObject struct {
    ID int
}

func main() {
    var mu sync.Mutex
    idCounter := 0
    
    // Create a pool with maximum 10 objects
    pool, err := leakypool.NewLeakyPool[MyObject](10, func() (MyObject, error) {
        mu.Lock()
        defer mu.Unlock()
        idCounter++
        return MyObject{ID: idCounter}, nil
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Pool created with capacity: %d\n", pool.Capacity())
    fmt.Printf("Initial pool size: %d\n", pool.Size())

    // Get an object (creates new if pool is empty)
    ref, err := pool.Get()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Got object: %+v\n", ref.Object)
    fmt.Printf("Pool size after Get: %d\n", pool.Size())

    // Return object to pool, ref is now invalid
    err = ref.Return()
    if err != nil {
        // Shown for example only
        // MyObject does not implement io.Closer, so ref.Return() would never return an error.
        fmt.Printf("Error returning object: %v\n", err)
    }
    fmt.Printf("Pool size after Return: %d\n", pool.Size())

    // Close the pool when done
    err = pool.Close()
    if err != nil {
        // Shown for example only
        // MyObject does not implement io.Closer, so pool.Close() would never return an error.
        fmt.Printf("Error closing pool: %v\n", err)
    }
    fmt.Printf("Pool size after Close: %d\n", pool.Size())
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
