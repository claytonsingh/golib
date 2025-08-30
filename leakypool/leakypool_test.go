package leakypool

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestObject is a simple test object that implements io.Closer
type TestObject struct {
	ID         int
	closed     bool
	closeError error // Optional error to return from Close()
}

func (t *TestObject) Close() error {
	t.closed = true
	return t.closeError
}

func testObjectFactory() func() (TestObject, error) {
	var (
		id int
		mu sync.Mutex
	)
	return func() (TestObject, error) {
		mu.Lock()
		id++
		newID := id
		mu.Unlock()
		return TestObject{ID: newID}, nil
	}
}

// testObjectWithErrorFactory returns a factory function that returns an error when the object is closed on odd ids.
func testObjectWithErrorFactory() func() (TestObject, error) {
	var (
		id int
		mu sync.Mutex
	)
	return func() (TestObject, error) {
		mu.Lock()
		newID := id
		id++
		mu.Unlock()
		if newID&1 == 1 {
			return TestObject{ID: newID, closeError: errors.New("close error")}, nil
		} else {
			return TestObject{ID: newID}, nil
		}
	}
}

func TestNewLeakyPool(t *testing.T) {
	t.Run("valid pool", func(t *testing.T) {
		pool, err := NewLeakyPool(10, testObjectFactory())
		if err != nil {
			t.Errorf("NewLeakyPool() error = %v", err)
		}
		if pool == nil {
			t.Error("NewLeakyPool() returned nil pool when no error expected")
		}
	})

	t.Run("zero size", func(t *testing.T) {
		pool, err := NewLeakyPool(0, testObjectFactory())
		if err == nil {
			t.Error("NewLeakyPool() expected error for zero size")
		}
		if pool != nil {
			t.Error("NewLeakyPool() returned non-nil pool when error expected")
		}
	})

	t.Run("negative size", func(t *testing.T) {
		pool, err := NewLeakyPool(-1, testObjectFactory())
		if err == nil {
			t.Error("NewLeakyPool() expected error for negative size")
		}
		if pool != nil {
			t.Error("NewLeakyPool() returned non-nil pool when error expected")
		}
	})

	t.Run("nil factory", func(t *testing.T) {
		pool, err := NewLeakyPool[TestObject](10, nil)
		if err == nil {
			t.Error("NewLeakyPool() expected error for nil factory")
		}
		if pool != nil {
			t.Error("NewLeakyPool() returned non-nil pool when error expected")
		}
	})
}

func TestLeakyPool_Get(t *testing.T) {
	pool, err := NewLeakyPool(2, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Test getting from empty pool (should create new object)
	ref, err := pool.Get()
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if ref.Object == nil {
		t.Error("Get() returned nil object")
	}
	if ref.Object.ID == 0 {
		t.Errorf("Get() returned object with ID %d, want non-zero", ref.Object.ID)
	}

	// Return object to pool
	ref.Return()

	// Test getting from non-empty pool (should reuse object)
	ref2, err := pool.Get()
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if ref2.Object == nil {
		t.Error("Get() returned nil object")
	}
	if ref2.Object.ID == 0 {
		t.Errorf("Get() returned object with ID %d, want non-zero", ref2.Object.ID)
	}
}

func TestLeakyPool_Get_WithError(t *testing.T) {
	expectedErr := errors.New("factory error")
	pool, err := NewLeakyPool(2, func() (TestObject, error) {
		return TestObject{}, expectedErr
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	ref, err := pool.Get()
	if err != expectedErr {
		t.Errorf("Get() error = %v, want %v", err, expectedErr)
	}
	if ref.Object != nil {
		t.Error("Get() returned non-nil object when factory failed")
	}
}

func TestLeakyPool_CapacityAndSize(t *testing.T) {
	pool, err := NewLeakyPool(5, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	if pool.Capacity() != 5 {
		t.Errorf("Capacity() = %d, want 5", pool.Capacity())
	}

	if pool.Size() != 0 {
		t.Errorf("Size() = %d, want 0", pool.Size())
	}

	// Add some objects
	ref1, _ := pool.Get()
	ref2, _ := pool.Get()
	ref1.Return()
	ref2.Return()

	if pool.Size() != 2 {
		t.Errorf("Size() = %d, want 2", pool.Size())
	}
}

func TestLeakyPool_Close(t *testing.T) {
	pool, err := NewLeakyPool(2, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Add objects to pool
	ref1, _ := pool.Get()
	ref2, _ := pool.Get()
	ref1.Return()
	ref2.Return()

	// Close pool
	err = pool.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify pool is closed
	if pool.Size() != 0 {
		t.Errorf("Size() after Close() = %d, want 0", pool.Size())
	}

	// Try to get from closed pool (should create new object)
	ref3, err := pool.Get()
	if err != nil {
		t.Errorf("Get() after Close() error = %v", err)
	}
	if ref3.Object == nil {
		t.Error("Get() after Close() returned nil object")
	}
}

func TestLeakyPool_Close_WithClosableObjects(t *testing.T) {
	pool, err := NewLeakyPool(2, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Add objects to pool
	ref1, _ := pool.Get()
	ref2, _ := pool.Get()

	// Store references to objects before returning them
	obj1 := ref1.Object
	obj2 := ref2.Object

	ref1.Return()
	ref2.Return()

	// Close pool
	err = pool.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify objects were closed
	if !obj1.closed {
		t.Error("Object 1 was not closed")
	}
	if !obj2.closed {
		t.Error("Object 2 was not closed")
	}
}

func TestLeakyPool_WorksAfterClose(t *testing.T) {
	pool, err := NewLeakyPool(2, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Close the pool first
	err = pool.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify pool size is 0 after close
	if pool.Size() != 0 {
		t.Errorf("Pool size after Close() = %d, want 0", pool.Size())
	}

	// Get an object after close (should create new object)
	ref1, err := pool.Get()
	if err != nil {
		t.Errorf("Get() after Close() error = %v", err)
	}
	if ref1.Object == nil {
		t.Error("Get() after Close() returned nil object")
	}

	// Get another object (should create another new object)
	ref2, err := pool.Get()
	if err != nil {
		t.Errorf("Get() after Close() error = %v", err)
	}
	if ref2.Object == nil {
		t.Error("Get() after Close() returned nil object")
	}

	// Verify these are different objects
	if ref1.Object == ref2.Object {
		t.Error("Get() after Close() returned same object twice")
	}

	// Store references to objects before returning them
	obj1 := ref1.Object
	obj2 := ref2.Object

	// Return first object (should call Close() since pool is closed)
	ref1.Return()

	// Verify first object was closed
	if !obj1.closed {
		t.Error("Object 1 was not closed when returned after Close()")
	}

	// Return second object (should call Close() since pool is closed)
	ref2.Return()

	// Verify second object was closed
	if !obj2.closed {
		t.Error("Object 2 was not closed when returned after Close()")
	}

	// Pool size should remain 0 since objects are discarded
	if pool.Size() != 0 {
		t.Errorf("Pool size after returns = %d, want 0", pool.Size())
	}
}

func TestLeakyPool_ConcurrentAccess(t *testing.T) {
	pool, err := NewLeakyPool(100, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	objectsPerGoroutine := 20

	// Start multiple goroutines that get and return objects
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < objectsPerGoroutine; j++ {
				ref, err := pool.Get()
				if err != nil {
					t.Errorf("Get() error in goroutine: %v", err)
					return
				}
				if ref.Object == nil {
					t.Error("Get() returned nil object in goroutine")
					return
				}
				ref.Return()
			}
		}()
	}

	wg.Wait()

	// Verify pool state is reasonable
	if pool.Size() > pool.Capacity() {
		t.Errorf("Pool size %d exceeds capacity %d", pool.Size(), pool.Capacity())
	}
}

func TestLeakyPool_LeakyBehavior(t *testing.T) {
	pool, err := NewLeakyPool(1, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Fill the pool
	ref1, _ := pool.Get()
	ref1.Return()

	// Try to return another object (should be discarded due to leaky behavior)
	ref2, _ := pool.Get()
	ref2.Return()

	// Pool should still only have 1 object
	if pool.Size() != 1 {
		t.Errorf("Pool size = %d, want 1", pool.Size())
	}
}

func TestRef_Return(t *testing.T) {
	pool, err := NewLeakyPool(1, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	ref, _ := pool.Get()

	// Return object
	ref.Return()

	// Verify ref is invalidated
	if ref.Object != nil {
		t.Error("Object not set to nil after Return()")
	}
	if ref.pool != nil {
		t.Error("Pool function not set to nil after Return()")
	}

	// Second return should be safe
	ref.Return()
}

func TestRef_Return_WithClosableObject(t *testing.T) {
	pool, err := NewLeakyPool(1, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// First, fill the pool with object 1
	ref1, _ := pool.Get()
	ref1.Return()

	// Verify pool has 1 object
	if pool.Size() != 1 {
		t.Errorf("Pool size after first return: %d, want 1", pool.Size())
	}

	// Now get object 1 from the pool
	ref2, _ := pool.Get()

	// Get another object (this creates a new object since pool is empty)
	ref3, _ := pool.Get()

	// Store reference to object 2 before returning it
	obj2 := ref3.Object

	// Return object 1 to the pool (should succeed)
	ref2.Return()

	// Pool should now have 1 object
	if pool.Size() != 1 {
		t.Errorf("Pool size after returning object 1: %d, want 1", pool.Size())
	}

	// Now try to return object 2 (pool is full, so it should be discarded)
	ref3.Return()

	// Verify object 2 was closed when discarded
	if !obj2.closed {
		t.Error("Object 2 was not closed when discarded")
	}

	// Pool should still have only 1 object
	if pool.Size() != 1 {
		t.Errorf("Pool size after discarding object 2: %d, want 1", pool.Size())
	}
}

func TestRef_Return_WithCloseError(t *testing.T) {
	pool, err := NewLeakyPool(1, testObjectWithErrorFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Fill the pool with first object (ID 0)
	ref1, _ := pool.Get()
	ref1.Return()

	// Verify pool has 1 object
	if pool.Size() != 1 {
		t.Errorf("Pool size after first return: %d, want 1", pool.Size())
	}

	// Get object from pool (should reuse existing object, ID 0)
	ref2, _ := pool.Get()

	// Get another object (this creates a new object since pool is empty, ID 1)
	ref3, _ := pool.Get()

	// Store reference to object before returning it
	obj3 := ref3.Object

	// Return object 2 to the pool (should succeed)
	ref2.Return()

	// Pool should now have 1 object
	if pool.Size() != 1 {
		t.Errorf("Pool size after returning object 2: %d, want 1", pool.Size())
	}

	// Now try to return object 3 (pool is full, so it should be discarded, ID 1)
	err = ref3.Return()

	// Check if this object should return an error based on its ID
	if obj3.ID != 1 {
		t.Errorf("Object 3 ID = %d, want 1", obj3.ID)
	}

	if err == nil {
		t.Error("Return() expected error for odd ID but got nil")
	}

	// Verify object was closed
	if !obj3.closed {
		t.Error("Object was not closed when discarded")
	}

	// Pool should still have only 1 object
	if pool.Size() != 1 {
		t.Errorf("Pool size after discarding object 3: %d, want 1", pool.Size())
	}
}

func TestLeakyPool_Close_WithCloseError(t *testing.T) {
	pool, err := NewLeakyPool(2, testObjectWithErrorFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Add objects to pool
	ref1, _ := pool.Get()
	ref2, _ := pool.Get()

	// Store references to objects before returning them
	obj1 := ref1.Object
	obj2 := ref2.Object

	ref1.Return()
	ref2.Return()

	// Close pool (should return combined errors if any objects have odd IDs)
	err = pool.Close()

	// Check if we have any odd IDs (which would return errors)
	hasOddIDs := (obj1.ID&1 == 1) || (obj2.ID&1 == 1)

	if hasOddIDs {
		if err == nil {
			t.Error("Close() expected error but got nil")
		}

		// Verify the error message contains the error message
		errStr := err.Error()
		if !strings.Contains(errStr, "close error") {
			t.Error("Close() error should contain 'close error'")
		}
	} else {
		// Even IDs should not return errors
		if err != nil {
			t.Errorf("Close() unexpected error: %v", err)
		}
	}

	// Verify objects were closed
	if !obj1.closed {
		t.Error("Object 1 was not closed")
	}
	if !obj2.closed {
		t.Error("Object 2 was not closed")
	}
}

func TestLeakyPool_StressTest(t *testing.T) {
	pool, err := NewLeakyPool(50, testObjectFactory())
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 20
	duration := 100 * time.Millisecond

	// Start stress test
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deadline := time.Now().Add(duration)
			for time.Now().Before(deadline) {
				ref, err := pool.Get()
				if err != nil {
					continue
				}
				ref.Return()
			}
		}()
	}

	wg.Wait()

	// Verify pool is still functional
	if pool.Capacity() != 50 {
		t.Errorf("Pool capacity changed during stress test: %d", pool.Capacity())
	}
}
