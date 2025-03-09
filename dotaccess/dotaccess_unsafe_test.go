package dotaccess

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestUnexportedFields(t *testing.T) {
	// Test case 1: bytes.Buffer has unexported 'buf' field
	t.Run("bytes.Buffer", func(t *testing.T) {
		// Create a buffer with some content
		buf := bytes.NewBuffer([]byte("hello"))

		// Access the unexported 'buf' field
		accessor, err := UnsafeGetAccessorDot[[]byte](buf, "buf")
		if err != nil {
			t.Fatalf("Failed to get accessor for bytes.Buffer.buf: %v", err)
		}

		// Get the current value
		bufSlice := accessor.Get()
		if string(bufSlice[:5]) != "hello" {
			t.Errorf("Expected 'hello', got '%s'", string(bufSlice[:5]))
		}

		// Set a new value
		newBuf := []byte("world")
		err = accessor.Set(newBuf)
		if err != nil {
			t.Fatalf("Failed to set bytes.Buffer.buf: %v", err)
		}

		// Verify the change worked
		if buf.String() != "world" {
			t.Errorf("Expected 'world', got '%s'", buf.String())
		}
	})

	// Test case 2: time.Time has unexported fields like 'sec', 'nsec', etc.
	t.Run("time.Time", func(t *testing.T) {
		tm := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		// Access the unexported 'wall' field
		accessor, err := UnsafeGetAccessorDot[uint64](&tm, "wall")
		if err != nil {
			t.Fatalf("Failed to get accessor for time.Time.wall: %v", err)
		}

		// Get the current value
		wallValue := accessor.Get()
		t.Logf("time.Time.wall type: %T, value: %v", wallValue, wallValue)

		// We won't try to modify this as time.Time's internals are complex
		// and highly implementation-dependent
	})

	// Test case 3: sync.Mutex has unexported 'state' field
	t.Run("sync.Mutex", func(t *testing.T) {
		var mu sync.Mutex

		// Lock it so state isn't zero
		mu.Lock()

		// Access the unexported 'state' field
		accessor, err := UnsafeGetAccessorDot[int32](&mu, "state")
		if err != nil {
			t.Fatalf("Failed to get accessor for sync.Mutex.state: %v", err)
		}

		// Get the current value while locked
		lockedState := accessor.Get()
		t.Logf("Locked mutex state type: %T, value: %v", lockedState, lockedState)

		// Verify non-zero while locked
		if lockedState == 0 {
			t.Error("Expected non-zero state value while locked")
		}

		// Unlock and check state changed
		mu.Unlock()

		// Get state after unlock
		unlockedState := accessor.Get()
		t.Logf("Unlocked mutex state type: %T, value: %v", unlockedState, unlockedState)

		// Verify state changed after unlock
		if unlockedState == lockedState {
			t.Error("Expected mutex state to change after unlock")
		}
	})
}

func TestNestedUnexportedFields(t *testing.T) {
	// Testing with http.Transport's dialsInProgress.headPos field
	t.Run("http.Transport.dialsInProgress.headPos", func(t *testing.T) {
		// Create an http.Transport
		transport := &http.Transport{}

		// First, let's verify we can access the dialsInProgress field
		// Using any as the type since the internal structure varies across Go versions
		dialsAccessor, err := UnsafeGetAccessorDot[any](&transport, "dialsInProgress")
		if err != nil {
			t.Fatalf("Failed to get accessor for transport.dialsInProgress: %v", err)
		}

		dialsValue := dialsAccessor.Get()
		t.Logf("Initial dialsInProgress value: %v (type: %T)", dialsValue, dialsValue)

		// Now try to access the nested headPos field
		headPosAccessor, err := UnsafeGetAccessorDot[int](&transport, "dialsInProgress.headPos")
		if err != nil {
			t.Fatalf("Failed to get accessor for transport.dialsInProgress.headPos: %v", err)
		}

		// Get the current headPos value
		initialHeadPos := headPosAccessor.Get()
		t.Logf("Initial headPos value: %v (type: %T)", initialHeadPos, initialHeadPos)

		// Try to modify the headPos value
		err = headPosAccessor.Set(42)
		if err != nil {
			t.Fatalf("Failed to set transport.dialsInProgress.headPos: %v", err)
		}

		// Get the updated value to verify it changed
		updatedHeadPos := headPosAccessor.Get()
		t.Logf("Updated headPos value: %v (type: %T)", updatedHeadPos, updatedHeadPos)

		// Verify the value changed as expected
		if updatedHeadPos != 42 {
			t.Errorf("Expected headPos to be 42 after update, got %v", updatedHeadPos)
		}

		// Reset to original value for cleanliness
		err = headPosAccessor.Set(initialHeadPos)
		if err != nil {
			t.Fatalf("Failed to reset transport.dialsInProgress.headPos: %v", err)
		}
	})
}
