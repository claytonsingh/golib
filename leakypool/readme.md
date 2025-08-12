# leakypool

A generic, thread-safe, non-blocking object pool implementation in Go.

This library provides a `LeakyPool` type that manages a pool of reusable objects. The pool is "leaky" - when the pool is full, returned objects are discarded instead of blocking or growing the pool size. This design ensures the pool never blocks operations and maintains predictable memory usage.

## Features

* **Generic:** Works with any type using Go generics.
* **Thread-safe:** Safe for concurrent access from multiple goroutines.
* **Non-blocking:** Operations never block, ensuring predictable performance.
* **Memory-bounded:** Pool max size is fixed, preventing unbounded memory growth.
* **Simple API:** Easy to use with `NewLeakyPool()`, `Get()`, `TryGet()`, and `Return()`.

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
    
    // Create a pool with a 10 object maximum
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

    // Try to get from pool without creating new
    if ref2, ok := pool.TryGet(); ok {
        fmt.Println("Got object from pool")
        ref2.Return()
    } else {
        fmt.Println("Pool was empty")
    }

    // Return object to pool, ref is now invalid
    ref.Return()
    fmt.Printf("Pool size after Return: %d\n", pool.Size())
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
