// Package leakypool provides a generic, thread-safe non-blocking object pool implementation.
// The pool is "leaky" - when the pool is full, returned objects are discarded
// instead of blocking or growing the pool size.
package leakypool

import "errors"

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
	select {
	case obj := <-this.pool:
		return Ref[T]{Object: obj, pool: this.pool}, nil
	default:
		obj, err := this.factory()
		if err != nil {
			return Ref[T]{Object: nil, pool: this.pool}, err
		}
		return Ref[T]{Object: &obj, pool: this.pool}, nil
	}
}

// TryGet attempts to get an object from the pool without creating a new one.
// Returns true if an object was available, false if the pool was empty.
// The passed ref is populated with the pooled object on success.
//
// Parameters:
//   - ref *Ref[T]: A pointer to a Ref that will be populated with the pooled object if available.
//
// Returns:
//   - bool: True if an object was available from the pool, false otherwise.
func (this *LeakyPool[T]) TryGet() (Ref[T], bool) {
	select {
	case obj := <-this.pool:
		return Ref[T]{Object: obj, pool: this.pool}, true
	default:
		return Ref[T]{Object: nil, pool: this.pool}, false
	}
}

// Ref represents a reference to a pooled object. It provides a safe way
// to return objects to the pool and prevents use-after-return errors.
type Ref[T any] struct {
	// Object is the pooled object reference. May be nil if the Ref is invalid.
	Object *T
	// pool is the channel to return the object to. Set to nil after Return() is called.
	pool chan *T
}

// Return attempts to return the object to the pool. If the pool is full, the object is discarded.
// After calling Return(), the Ref should not be used again.
func (this *Ref[T]) Return() {
	if this.pool != nil && this.Object != nil {
		select {
		case this.pool <- this.Object:
			this.pool = nil
			this.Object = nil
		default:
			// Pool is full, discard the object
		}
	}
}

// Capacity returns the maximum number of objects that can be stored in the pool.
func (p *LeakyPool[T]) Capacity() int {
	return cap(p.pool)
}

// Size returns the current number of objects in the pool.
func (p *LeakyPool[T]) Size() int {
	return len(p.pool)
}
