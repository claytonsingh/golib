package syncsignal

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewSignal verifies NewSignal creates a BroadcastSignal in the correct initial state.
func TestNewSignal(t *testing.T) {
	signal := NewSignal()
	if signal == nil {
		t.Fatal("NewSignal() returned nil")
	}
	if signal.closed {
		t.Error("BroadcastSignal should not be closed")
	}
	if signal.value != 0 {
		t.Errorf("BroadcastSignal should have value 0, got %d", signal.value)
	}
}

// TestWait tests multiple goroutines waiting on the same BroadcastSignal.
func TestWait(t *testing.T) {
	signal := NewSignal()
	var counter atomic.Int32

	// Start multiple goroutines waiting
	for i := 0; i < 10; i++ {
		go func(id int) {
			counter.Add(1)
			wait := signal.GetWaiter(false)
			if !wait() {
				t.Errorf("Goroutine %d: Wait should return true for signal", id)
			}
			counter.Add(-1)
		}(i)
	}

	// Give goroutines time to start and wait
	time.Sleep(10 * time.Millisecond)

	// Verify all goroutines are waiting
	c1 := counter.Load()
	if c1 != 10 {
		t.Errorf("Should have ten goroutines waiting, got %d", c1)
	}

	// Signal should wake up all waiting goroutines
	signal.Signal()

	// Give goroutines time to complete
	time.Sleep(10 * time.Millisecond)

	// Wait for all goroutines to complete
	c2 := counter.Load()
	if c2 != 0 {
		t.Errorf("Should have completed all goroutines, got %d", c2)
	}
}

// TestClose verifies Close() properly shuts down the BroadcastSignal and wakes all waiters.
func TestClose(t *testing.T) {
	signal := NewSignal()
	var counter atomic.Int32

	// Start two goroutines that wait for signals
	for i := 0; i < 10; i++ {
		go func(id int) {
			counter.Add(1)
			wait := signal.GetWaiter(false)
			if wait() {
				t.Errorf("Goroutine %d: Wait should return false for close", id)
			}
			counter.Add(-1)
		}(i)
	}

	// Give goroutines time to start and wait
	time.Sleep(10 * time.Millisecond)

	// Verify all goroutines are waiting
	c1 := counter.Load()
	if c1 != 10 {
		t.Errorf("Should have two goroutines waiting, got %d", c1)
	}

	// Close should wake up all waiting goroutines
	signal.Close()

	// Give goroutines time to complete
	time.Sleep(10 * time.Millisecond)

	// Verify all goroutines to complete
	c2 := counter.Load()
	if c2 != 0 {
		t.Errorf("Should have completed all goroutines, got %d", c2)
	}
}

// TestCloseIdempotent ensures multiple calls to Close() are safe
func TestCloseIdempotent(t *testing.T) {
	signal := NewSignal()

	// First close
	signal.Close()
	if !signal.closed {
		t.Error("BroadcastSignal should be closed after first Close() call")
	}

	// Second close should not cause issues
	signal.Close()
	if !signal.closed {
		t.Error("BroadcastSignal should remain closed after second Close() call")
	}
}

// TestGetWaiterMixed tests different blocking behavior between GetWaiter(false) and GetWaiter(true)
func TestGetWaiterMixed(t *testing.T) {
	signal := NewSignal()
	var wg sync.WaitGroup
	wg.Add(2)

	startTime := time.Now()
	var goroutine1Done, goroutine2Done bool
	var mu sync.Mutex

	// Goroutine 1: GetWaiter(false) - should block until signal
	go func() {
		defer wg.Done()
		wait := signal.GetWaiter(false)
		if !wait() {
			t.Error("GetWaiter(false) should return true for signal")
		}

		mu.Lock()
		goroutine1Done = true
		mu.Unlock()
	}()

	// Goroutine 2: GetWaiter(true) - should not block on first call
	go func() {
		defer wg.Done()
		wait := signal.GetWaiter(true)
		if !wait() {
			t.Error("GetWaiter(true) should return true for signal")
		}

		mu.Lock()
		goroutine2Done = true
		mu.Unlock()
	}()

	// Give goroutines time to start
	time.Sleep(10 * time.Millisecond)

	// Check state after first wait - GetWaiter(true) should be done, GetWaiter(false) should still be waiting
	mu.Lock()
	if !goroutine2Done {
		t.Error("GetWaiter(true) should complete immediately on first call")
	}
	if goroutine1Done {
		t.Error("GetWaiter(false) should still be waiting for signal")
	}
	mu.Unlock()

	// First signal to wake up GetWaiter(false)
	signal.Signal()

	// Give time for goroutine1 to process the signal
	time.Sleep(10 * time.Millisecond)

	// Check state after signal - both should be done
	mu.Lock()
	if !goroutine1Done {
		t.Error("GetWaiter(false) should complete after signal")
	}
	if !goroutine2Done {
		t.Error("GetWaiter(true) should still be done")
	}
	mu.Unlock()

	// Then close to wake up any remaining waiters
	signal.Close()

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify GetWaiter(true) didn't block initially
	if time.Since(startTime) > 50*time.Millisecond {
		t.Error("GetWaiter(true) should not block on first call")
	}
}

// TestSignalCloseSequence verifies proper behavior when Signal() is called then Close()
func TestSignalCloseSequence(t *testing.T) {
	signal := NewSignal()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		wait := signal.GetWaiter(false)

		// Should receive the signal first
		if !wait() {
			t.Error("Should receive signal first, not close")
		}

		// Then should receive close
		if wait() {
			t.Error("Should receive close after signal")
		}
	}()

	// Give goroutine time to start and wait
	time.Sleep(10 * time.Millisecond)

	// Send signal then close
	signal.Signal()
	signal.Close()

	wg.Wait()
}

// TestMultipleSignals tests that rapid successive signals only wake waiters once
func TestMultipleSignals(t *testing.T) {
	signal := NewSignal()
	var wg sync.WaitGroup
	wg.Add(1)

	signalCount := 0
	go func() {
		defer wg.Done()
		for wait := signal.GetWaiter(false); wait(); {
			time.Sleep(20 * time.Millisecond)
			signalCount++
		}
	}()

	// Give goroutine time to start and wait
	time.Sleep(10 * time.Millisecond)

	// Wake up the goroutine (signalCount = 1)
	signal.Signal()
	time.Sleep(10 * time.Millisecond)

	// Send multiple signals rapidly (signalCount = 2)
	for i := 0; i < 10; i++ {
		signal.Signal()
	}

	// Stop the goroutine from waiting
	time.Sleep(10 * time.Millisecond)
	signal.Close()

	// Wait for goroutine to complete
	wg.Wait()

	// Should only process two signals
	if signalCount != 2 {
		t.Errorf("Should only process two signals, processed %d", signalCount)
	}
}
