package task

import (
	"errors"
	"fmt"
	"iter"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/claytonsingh/golib/syncsignal"
)

// ErrAsyncPanic is returned when an async function panics.
var ErrAsyncPanic = errors.New("panic")

// asyncPanicError holds the value recovered from a panic in an async function.
type asyncPanicError struct {
	Value any
}

func (e *asyncPanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Value)
}

func (e *asyncPanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}

func (e *asyncPanicError) Is(target error) bool {
	return target == ErrAsyncPanic
}

// WrapAwaitFault is an optional hook applied when Await returns a non-nil error.
// It runs in the caller's goroutine at the await site, not in the Async worker,
// which makes it suitable for recording stack traces that cross the goroutine
// boundary (for example a stack-capturing error wrapper with Unwrap support).
//
// The default is nil (no transformation). Store a hook with WrapAwaitFault.Store,
// typically once during init and only in development or staging, so production
// pays a single nil check on the error path:
//
//	func init() {
//	    if isDev {
//	        fn := mypkg.WrapAwaitFault
//	        task.WrapAwaitFault.Store(&fn)
//	    }
//	}
//
// WrapAwaitFault is read with Load at each Await call. The stored task error is
// not modified; only the value returned from Await is wrapped. Pair with
// throw site wrapping inside the Async function for full async fault chains.
var WrapAwaitFault atomic.Pointer[func(error) error]

// Task holds the result of an async expression.
type Task[T any] struct {
	mu         sync.Mutex
	done       bool
	value      T
	err        error
	nextID     uint64
	onComplete map[uint64]func()
}

var _ TaskLinker = &Task[any]{}

type TaskLinker interface {
	Link(cb func()) LinkHandle
}

// LinkHandle is an opaque registration returned by Link. The zero value is invalid
// and Unlink is a no-op. Valid handles have id > 0 (nextID is pre-incremented).
type LinkHandle struct {
	reg linkRegistry
	id  uint64
}

type linkRegistry interface {
	unlinkHook(id uint64)
}

// Unlink cancels a pending registration before completion fires.
// After completion, the registry is cleared and Unlink is a no-op.
func (h LinkHandle) Unlink() {
	if h.id > 0 && h.reg != nil {
		h.reg.unlinkHook(h.id)
	}
}

// Async runs fn in a new goroutine and returns a Task handle.
func Async[T any](fn func() (T, error)) *Task[T] {
	t := &Task[T]{} // No initialization overhead
	go func() {
		var v T
		var err error
		defer func() {
			if r := recover(); r != nil {
				t.mu.Lock()
				t.err = &asyncPanicError{Value: r}
			} else {
				t.mu.Lock()
				t.value = v
				t.err = err
			}
			t.done = true

			// Fire: take every pending handler and clear the registry.
			// Handlers registered here are implicitly unlinked; no per-ID cleanup.
			callbacks := t.onComplete
			t.onComplete = nil
			t.mu.Unlock()

			for _, cb := range callbacks {
				safeCall(cb)
			}
		}()
		v, err = fn()
	}()
	return t
}

// Link registers a callback to be executed when the task completes.
// If the task is already complete, the callback is executed immediately and
// returns a zero LinkHandle (nothing to unlink).
//
// **Important:** callbacks passed to Link must never panic. Panicking inside a callback
// may crash the goroutine running the task or interfere with internal state.
func (t *Task[T]) Link(cb func()) LinkHandle {
	t.mu.Lock()
	if t.done {
		t.mu.Unlock()
		// Safe: The callback runs outside the lock, and because the task is
		// already done, this callback is never present in the onComplete map.
		safeCall(cb)
		return LinkHandle{}
	}

	if t.onComplete == nil {
		t.onComplete = make(map[uint64]func())
	}

	t.nextID++
	id := t.nextID
	t.onComplete[id] = cb
	t.mu.Unlock()
	return LinkHandle{reg: t, id: id}
}

func (t *Task[T]) unlinkHook(id uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.onComplete != nil {
		delete(t.onComplete, id)
	}
}

func safeCall(cb func()) {
	defer func() { recover() }()
	cb()
}

// Wait blocks until completion or timeout. It never returns an error.
// Non-positive timeout is a poll: true if already done, false otherwise.
//
// Wait and Await both block via syncsignal.OnceSignal rather than sync.Cond or a
// semaphore. Completion (or a timer in Wait) calls Signal exactly once.
func (t *Task[T]) Wait(timeoutMs int64) bool {
	t.mu.Lock()
	done := t.done
	t.mu.Unlock()

	if !done && timeoutMs > 0 {
		evt := syncsignal.NewOnceSignal()
		hook := t.Link(evt.Signal)
		timer := time.AfterFunc(time.Duration(timeoutMs)*time.Millisecond, evt.Signal)
		evt.Wait()
		timer.Stop()
		hook.Unlink()

		t.mu.Lock()
		done = t.done
		t.mu.Unlock()
	}

	return done
}

// Await blocks until completion and returns the result or error.
// When the task failed and WrapAwaitFault is set, the returned error is passed
// through that hook before being returned.
func (t *Task[T]) Await() (T, error) {
	t.mu.Lock()
	done := t.done
	v := t.value
	err := t.err
	t.mu.Unlock()

	if !done {
		evt := syncsignal.NewOnceSignal()
		hook := t.Link(evt.Signal)
		evt.Wait()
		hook.Unlink()

		t.mu.Lock()
		v = t.value
		err = t.err
		t.mu.Unlock()
	}

	if err != nil {
		if fn := WrapAwaitFault.Load(); fn != nil {
			err = (*fn)(err)
		}
	}

	return v, err
}

// UnorderedIterateSlice returns a Go iterator that yields task indices as they complete.
// Order is undefined but will only yield completed tasks.
// Nil entries are skipped and do not produce a yield.
func UnorderedIterateSlice[T any](tasks []*Task[T]) iter.Seq[int] {
	return unorderedIterate(tasks, func(t *Task[T]) TaskLinker { return t })
}

// UnorderedIterateAny returns a Go iterator that yields task indices as they complete.
// Order is undefined but will only yield completed tasks.
// Nil entries are skipped and do not produce a yield.
// All items must be castable to TaskLinker or nil.
func UnorderedIterateAny(tasks []any) (iter.Seq[int], error) {
	for i, item := range tasks {
		if item == nil {
			continue
		}
		if _, ok := item.(TaskLinker); !ok {
			return nil, fmt.Errorf("tasks[%d]: not a TaskLinker", i)
		}
	}
	return unorderedIterate(tasks, func(t any) TaskLinker { return t.(TaskLinker) }), nil
}

// UnorderedIterateLinkers returns a Go iterator that yields task indices as they complete.
// Order is undefined but will only yield completed tasks.
// Nil entries are skipped and do not produce a yield.
func UnorderedIterateLinkers(tasks []TaskLinker) iter.Seq[int] {
	return unorderedIterate(tasks, func(t TaskLinker) TaskLinker { return t })
}

// unorderedIterate returns a Go iterator that yields task indices as they complete.
// Order is undefined but will only yield completed tasks.
// Nil entries are skipped and do not produce a yield.
func unorderedIterate[T any](tasks []T, linkFn func(T) TaskLinker) iter.Seq[int] {
	return func(yield func(int) bool) {
		if len(tasks) == 0 {
			return
		}

		var localMu sync.Mutex
		localCond := sync.NewCond(&localMu)

		// Pre-allocated buffer matching the total task count.
		queue := make([]int, len(tasks))
		head := 0
		tail := 0

		hooks := make([]LinkHandle, 0, len(tasks))

		defer func() {
			for _, hook := range hooks {
				hook.Unlink()
			}
		}()

		// 1. Subscription Phase
		for i, t := range tasks {
			if isNil(t) {
				localMu.Lock()
				queue = queue[:len(queue)-1] // reduce the queue length by 1 for each nil task
				localMu.Unlock()
			} else {
				taskIdx := i // refers to the index of the task in the tasks slice
				hooks = append(hooks, linkFn(t).Link(func() {
					localMu.Lock()
					queue[tail] = taskIdx
					tail++
					localCond.Signal()
					localMu.Unlock()
				}))
			}
		}

		// 2. Consumption Phase
		for {
			localMu.Lock()

			// If the queue is dry, decide whether to exit or sleep.
			for head == tail {
				if head == len(queue) {
					localMu.Unlock()
					hooks = hooks[:0]
					return
				}
				localCond.Wait()
			}

			// When we get past the loop, head != tail is guaranteed.
			idx := queue[head]
			head++
			localMu.Unlock()

			if !yield(idx) {
				return
			}
		}
	}
}

// isNil reports skipped slots in the subscription loop: nil interfaces and typed-nil pointers.
// Used only by unorderedIterate
func isNil[T any](t T) bool {
	// Despite the use of reflect, this is just a lookup on the interface. So it's fast.
	v := reflect.ValueOf(t)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return !v.IsValid()
	}
}
