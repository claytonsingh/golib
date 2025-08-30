// Package leakypool provides a generic, thread-safe non-blocking object pool implementation.
// The pool is "leaky" - when the pool is full, returned objects are discarded
// instead of blocking or growing the pool size.
package leakypool

import (
	"errors"
	"io"
	"sync"
)

var (
	// ErrInvalidSize is returned when attempting to create a pool with non-positive size.
	ErrInvalidSize = errors.New("pool size must be positive")
	// ErrNilFactory is returned when attempting to create a pool with a nil factory function.
	ErrNilFactory = errors.New("factory function cannot be nil")
)

// LeakyPool is a generic object pool that provides non-blocking access to pooled objects.
// When the pool is empty, Get() creates new objects using the factory function.
// When the pool is full, Return() discards objects instead of blocking.
type LeakyPool[T any] struct {
	pool    chan *T
	mu      sync.RWMutex // mu is used to protect the pool; it is only locked for methods that mutate pool.
	factory func() (T, error)
}

// NewLeakyPool creates and returns a new LeakyPool instance.
//
// Parameters:
//   - size int: The maximum number of objects that can be stored in the pool.
//   - factory func() (T, error): A function that creates new objects when the pool is empty.
//
// Returns:
//   - *LeakyPool[T]: A new initialized LeakyPool instance.
//   - error: An error if size is not positive or if factory is nil.
func NewLeakyPool[T any](size int, factory func() (T, error)) (*LeakyPool[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	if factory == nil {
		return nil, ErrNilFactory
	}

	var this LeakyPool[T]
	this.pool = make(chan *T, size)
	this.factory = factory
	return &this, nil
}

// Get retrieves an object from the pool. If the pool is empty, a new object is created using the factory function.
// This method never blocks.
//
// Returns:
//   - Ref[T]: A reference to the retrieved or newly created object.
//   - error: Forwards the error returned by the factory function if object creation fails.
func (this *LeakyPool[T]) Get() (Ref[T], error) {
	this.mu.RLock()
	pool := this.pool
	this.mu.RUnlock()

	select {
	case obj := <-pool:
		return Ref[T]{Object: obj, pool: this.storeorclose}, nil
	default:
		obj, err := this.factory()
		if err != nil {
			return Ref[T]{}, err
		}
		return Ref[T]{Object: &obj, pool: this.storeorclose}, nil
	}
}

// Close drains the pool and closes all objects.
// Any objects returned to the pool after Close() is called are discarded.
// If the objects implement io.Closer, they are closed.
//
// Returns:
//   - error: An error if any closable objects fail to close, or nil if all objects closed successfully.
func (this *LeakyPool[T]) Close() error {
	// Only Close mutates the pool, so we need exclusive access to it.
	this.mu.Lock()
	pool := this.pool
	this.pool = nil
	this.mu.Unlock()

	var errs []error

	if pool == nil {
		return nil
	}
	for {
		select {
		case obj := <-pool:
			if closer, ok := any(obj).(io.Closer); ok {
				err := closer.Close()
				if err != nil {
					errs = append(errs, err)
				}
			}
		default:
			close(pool)
			return errors.Join(errs...)
		}
	}
}

// storeorclose is a helper function to store an object in the pool or close/discard it if the pool is full.
//
// Returns:
//   - error: An error if the object is closable and Close() fails, or nil otherwise.
func (this *LeakyPool[T]) storeorclose(obj *T) error {
	this.mu.RLock()
	pool := this.pool
	this.mu.RUnlock()
	select {
	case pool <- obj:
		return nil
	default:
		// Pool is full, discard the object
		if closer, ok := any(obj).(io.Closer); ok {
			return closer.Close()
		}
		return nil
	}
}

// Capacity returns the maximum number of objects that can be stored in the pool.
func (this *LeakyPool[T]) Capacity() int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return cap(this.pool)
}

// Size returns the current number of objects in the pool.
func (this *LeakyPool[T]) Size() int {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return len(this.pool)
}

// Ref represents a reference to a pooled object. It provides a safe way
// to return objects to the pool and prevents use-after-return errors.
type Ref[T any] struct {
	// Object is the pooled object reference. May be nil if the Ref is invalid.
	Object *T
	// pool is the channel to return the object to. Set to nil after Return() is called.
	pool func(obj *T) error
}

// Return attempts to return the object to the pool. If the pool is full, the object is discarded.
// After calling Return(), the Ref should not be used again.
//
// Returns:
//   - error: An error if the object is closable and Close() fails when discarded, or nil otherwise.
func (this *Ref[T]) Return() error {
	var err error
	if this.pool != nil && this.Object != nil {
		err = this.pool(this.Object)
		this.pool = nil
		this.Object = nil
	}
	return err
}
