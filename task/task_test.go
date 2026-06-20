package task

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAsyncAwaitSuccess(t *testing.T) {
	tk := Async(func() (int, error) {
		return 42, nil
	})
	v, err := tk.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Errorf("got value %d, want 42", v)
	}
}

func TestAsyncAwaitError(t *testing.T) {
	want := errors.New("boom")
	tk := Async(func() (int, error) {
		return 0, want
	})
	_, err := tk.Await()
	if !errors.Is(err, want) {
		t.Fatalf("got error %v, want %v", err, want)
	}
}

func TestAsyncAwaitPanic(t *testing.T) {
	tk := Async(func() (int, error) {
		panic("oops")
	})
	_, err := tk.Await()
	if !errors.Is(err, ErrAsyncPanic) {
		t.Fatalf("errors.Is(err, ErrPanic) = false, got %v", err)
	}
	var pe *asyncPanicError
	if !errors.As(err, &pe) {
		t.Fatalf("errors.As(err, &PanicError) = false, got %v", err)
	}
	if pe.Value != "oops" {
		t.Errorf("PanicError.Value = %v, want %q", pe.Value, "oops")
	}
}

func TestAsyncAwaitPanicError(t *testing.T) {
	want := errors.New("inner")
	tk := Async(func() (int, error) {
		panic(want)
	})
	_, err := tk.Await()
	if !errors.Is(err, ErrAsyncPanic) {
		t.Fatalf("errors.Is(err, ErrPanic) = false, got %v", err)
	}
	if !errors.Is(err, want) {
		t.Fatalf("errors.Is(err, want) = false, got %v", err)
	}
}

func TestAwaitBlocksUntilComplete(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (string, error) {
		<-release
		return "done", nil
	})

	done := make(chan struct{})
	go func() {
		v, err := tk.Await()
		if err != nil || v != "done" {
			t.Errorf("Await returned %q, %v", v, err)
		}
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Await returned before task completed")
	case <-time.After(20 * time.Millisecond):
	}

	close(release)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Await did not return after task completed")
	}
}

func TestAwaitAfterAlreadyComplete(t *testing.T) {
	tk := Async(func() (int, error) {
		return 7, nil
	})

	v, err := tk.Await()
	if err != nil {
		t.Fatalf("first Await: unexpected error: %v", err)
	}
	if v != 7 {
		t.Fatalf("first Await: got value %d, want 7", v)
	}

	v, err = tk.Await()
	if err != nil {
		t.Fatalf("second Await: unexpected error: %v", err)
	}
	if v != 7 {
		t.Errorf("second Await: got value %d, want 7", v)
	}
}

func TestAwaitErrorAfterAlreadyComplete(t *testing.T) {
	want := errors.New("failed")
	tk := Async(func() (int, error) {
		return 0, want
	})
	_, err := tk.Await()
	if !errors.Is(err, want) {
		t.Fatalf("first Await: got error %v, want %v", err, want)
	}

	_, err = tk.Await()
	if !errors.Is(err, want) {
		t.Fatalf("second Await: got error %v, want %v", err, want)
	}
}

func TestWaitPollAndTimeout(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (int, error) {
		<-release
		return 1, nil
	})

	if tk.Wait(0) {
		t.Error("Wait(0) should return false for incomplete task")
	}

	waitDone := make(chan bool, 1)
	go func() {
		waitDone <- tk.Wait(200)
	}()

	select {
	case ok := <-waitDone:
		if ok {
			t.Error("Wait should return false on timeout")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Wait did not return on timeout")
	}

	close(release)
	tk.Await()

	if !tk.Wait(0) {
		t.Error("Wait(0) should return true after task completes")
	}
	if !tk.Wait(1000) {
		t.Error("Wait with positive timeout should return true for completed task")
	}
}

func TestLinkFiresOnCompletion(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (int, error) {
		<-release
		return 1, nil
	})

	var fired atomic.Bool
	tk.Link(func() { fired.Store(true) })

	if fired.Load() {
		t.Fatal("callback fired before completion")
	}

	close(release)
	tk.Await()

	if !fired.Load() {
		t.Error("callback did not fire on completion")
	}
}

func TestLinkImmediateWhenDone(t *testing.T) {
	tk := Async(func() (int, error) { return 1, nil })
	tk.Await()

	var fired bool
	h := tk.Link(func() { fired = true })
	if !fired {
		t.Error("callback should run immediately when task already done")
	}
	if h.id != 0 {
		t.Error("LinkHandle should be zero when callback runs immediately")
	}
}

func TestLinkUnlink(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (int, error) {
		<-release
		return 1, nil
	})

	var fired atomic.Bool
	h := tk.Link(func() { fired.Store(true) })
	h.Unlink()

	close(release)
	tk.Await()

	if fired.Load() {
		t.Error("unlinked callback should not fire")
	}
}

func TestLinkMultipleCallbacks(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (int, error) {
		<-release
		return 1, nil
	})

	var count atomic.Int32
	for range 3 {
		tk.Link(func() { count.Add(1) })
	}

	close(release)
	tk.Await()

	if count.Load() != 3 {
		t.Errorf("got %d callbacks, want 3", count.Load())
	}
}

func TestLinkCallbackPanicDoesNotBlockOthers(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (int, error) {
		<-release
		return 1, nil
	})

	var count atomic.Int32
	tk.Link(func() { panic("first") })
	tk.Link(func() { count.Add(1) })

	close(release)
	v, err := tk.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 1 {
		t.Fatalf("got value %d, want 1", v)
	}
	if count.Load() != 1 {
		t.Errorf("got %d callbacks after panic, want 1", count.Load())
	}
}

func TestWrapAwaitFault(t *testing.T) {
	old := WrapAwaitFault.Load()
	defer func() { WrapAwaitFault.Store(old) }()

	fn := func(err error) error {
		return fmt.Errorf("wrapped: %w", err)
	}
	WrapAwaitFault.Store(&fn)

	want := errors.New("inner")
	tk := Async(func() (int, error) { return 0, want })
	_, err := tk.Await()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.HasPrefix(err.Error(), "wrapped:") {
		t.Errorf("got %v, want wrapped error", err)
	}
	if !errors.Is(err, want) {
		t.Errorf("wrapped error should unwrap to %v", want)
	}

	tk.mu.Lock()
	stored := tk.err
	tk.mu.Unlock()
	if !errors.Is(stored, want) {
		t.Error("stored task error should remain unwrapped")
	}
}

func TestUnorderedIterateSliceCompletionOrder(t *testing.T) {
	s0 := make(chan struct{})
	s2 := make(chan struct{})
	s3 := make(chan struct{})

	tasks := []*Task[int]{
		Async(func() (int, error) { <-s0; return 0, nil }),
		nil,
		Async(func() (int, error) { <-s2; return 2, nil }),
		Async(func() (int, error) { <-s3; return 1, nil }),
	}

	gotCh := make(chan int, 3)
	go func() {
		for idx := range UnorderedIterateSlice(tasks) {
			gotCh <- idx
		}
		close(gotCh)
	}()

	for _, want := range []int{2, 3, 0} {
		switch want {
		case 0:
			close(s0)
		case 2:
			close(s2)
		case 3:
			close(s3)
		}
		select {
		case idx := <-gotCh:
			if idx != want {
				t.Fatalf("got index %d, want %d", idx, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for index %d", want)
		}
	}

	if _, ok := <-gotCh; ok {
		t.Fatal("iterator should be done")
	}
}

func TestUnorderedIterateSliceEarlyStop(t *testing.T) {
	tasks := []*Task[int]{
		Async(func() (int, error) { return 0, nil }),
		Async(func() (int, error) { return 1, nil }),
	}

	var count int
	for range UnorderedIterateSlice(tasks) {
		count++
		break
	}
	if count != 1 {
		t.Errorf("expected 1 yield, got %d", count)
	}
}

func TestUnorderedIterateAny(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (int, error) {
		<-release
		return 1, nil
	})

	seq, err := UnorderedIterateAny([]any{tk, nil})
	if err != nil {
		t.Fatal(err)
	}

	gotCh := make(chan int, 1)
	go func() {
		for idx := range seq {
			gotCh <- idx
		}
		close(gotCh)
	}()

	close(release)

	idx, ok := <-gotCh
	if !ok || idx != 0 {
		t.Errorf("got index %d (ok=%v), want 0", idx, ok)
	}
}

func TestUnorderedIterateAnyInvalidType(t *testing.T) {
	_, err := UnorderedIterateAny([]any{"not a task"})
	if err == nil {
		t.Fatal("expected error for non-TaskLinker")
	}
	if !strings.Contains(err.Error(), "tasks[0]") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUnorderedIterateLinkers(t *testing.T) {
	slow := make(chan struct{})
	fast := make(chan struct{})

	tasks := []TaskLinker{
		Async(func() (int, error) { <-slow; return 0, nil }),
		Async(func() (int, error) { <-fast; return 1, nil }),
	}

	gotCh := make(chan int, 2)
	go func() {
		for idx := range UnorderedIterateLinkers(tasks) {
			gotCh <- idx
		}
		close(gotCh)
	}()

	close(fast)
	if idx := <-gotCh; idx != 1 {
		t.Fatalf("fast task: got index %d, want 1", idx)
	}

	close(slow)
	if idx := <-gotCh; idx != 0 {
		t.Fatalf("slow task: got index %d, want 0", idx)
	}
}

func TestUnorderedIterateEmpty(t *testing.T) {
	var count int
	for range UnorderedIterateSlice([]*Task[int]{}) {
		count++
	}
	if count != 0 {
		t.Errorf("empty slice should yield nothing, got %d", count)
	}
}

func TestUnorderedIterateAllNil(t *testing.T) {
	tasks := []*Task[int]{nil, nil}
	var count int
	for range UnorderedIterateSlice(tasks) {
		count++
	}
	if count != 0 {
		t.Errorf("all-nil slice should yield nothing, got %d", count)
	}
}

func TestLinkHandleUnlinkNoOp(t *testing.T) {
	var h LinkHandle
	h.Unlink()

	tk := Async(func() (int, error) { return 1, nil })
	tk.Await()
	h2 := LinkHandle{reg: tk, id: 99}
	h2.Unlink()
}

func TestConcurrentAwait(t *testing.T) {
	release := make(chan struct{})
	tk := Async(func() (int, error) {
		<-release
		return 99, nil
	})

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := tk.Await()
			if err != nil || v != 99 {
				t.Errorf("Await returned %d, %v", v, err)
			}
		}()
	}

	close(release)
	wg.Wait()
}

func TestUnorderedIterateSliceCollectAll(t *testing.T) {
	n := 5
	release := make(chan struct{})
	tasks := make([]*Task[int], n)
	for i := range tasks {
		i := i
		tasks[i] = Async(func() (int, error) {
			<-release
			return i, nil
		})
	}

	var got []int
	done := make(chan struct{})
	go func() {
		for idx := range UnorderedIterateSlice(tasks) {
			got = append(got, idx)
		}
		close(done)
	}()

	close(release)
	<-done

	slices.Sort(got)
	want := []int{0, 1, 2, 3, 4}
	if !slices.Equal(got, want) {
		t.Errorf("got indices %v, want %v", got, want)
	}
}
