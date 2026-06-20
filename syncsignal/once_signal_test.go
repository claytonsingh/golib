package syncsignal

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestOnceSignalZeroValueSignaled(t *testing.T) {
	var ev OnceSignal
	done := make(chan struct{})
	go func() {
		ev.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Error("zero value OnceSignal should be signaled; Wait should return immediately")
	}
}

func TestNewOnceSignalBlocksUntilSignal(t *testing.T) {
	ev := NewOnceSignal()
	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		close(started)
		ev.Wait()
		close(done)
	}()

	<-started
	select {
	case <-done:
		t.Error("Wait should block until Signal is called")
	case <-time.After(50 * time.Millisecond):
	}

	ev.Signal()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Error("Wait should return after Signal")
	}
}

func TestOnceSignalReleasesAllWaiters(t *testing.T) {
	ev := NewOnceSignal()
	var counter atomic.Int32

	for i := 0; i < 10; i++ {
		go func() {
			counter.Add(1)
			ev.Wait()
			counter.Add(-1)
		}()
	}

	time.Sleep(10 * time.Millisecond)
	if counter.Load() != 10 {
		t.Fatalf("expected 10 goroutines waiting, got %d", counter.Load())
	}

	ev.Signal()
	time.Sleep(10 * time.Millisecond)
	if counter.Load() != 0 {
		t.Errorf("expected all goroutines to complete, got %d still running", counter.Load())
	}
}

func TestOnceSignalIdempotent(t *testing.T) {
	ev := NewOnceSignal()
	ev.Signal()
	ev.Signal()

	done := make(chan struct{})
	go func() {
		ev.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Error("Wait should return immediately after Signal")
	}
}

func TestOnceSignalFutureWaitReturnsImmediately(t *testing.T) {
	ev := NewOnceSignal()
	ev.Signal()

	for i := 0; i < 5; i++ {
		done := make(chan struct{})
		go func() {
			ev.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(50 * time.Millisecond):
			t.Fatalf("Wait %d should return immediately when already signaled", i)
		}
	}
}
