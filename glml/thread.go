// Copyright © 2012 Popog
package glml

import (
	"errors"
	"runtime"
)

type Threadable interface {
	// If threadables do not need thread-dependent setup and teardown
	// ThreadIsInitialized should always return true.
	// If ThreadIsInitialized returns false, ThreadInitialize is called
	// and the threadable is recorded in a teardown list.
	// ThreadClose will be called on the thread when either the Threadable
	// is requested to be closed on the thread on which it was created
	// 
	ThreadIsInitialized() bool
	ThreadInitialize(thread *Thread) ThreadError
	ThreadClose(thread *Thread)

	ThreadActivate(thread *Thread) ThreadError                             // The first function run on the thread. Any error returned is considered fatal.
	ThreadDeactivate(thread *Thread) ThreadError                           // The last function run on the thread
	ThreadCommands() <-chan func(thread *Thread, t Threadable) ThreadError // The channel from which more functions will be run

	// All errors which occur on the thread will be sent via
	// calls to item.ReportError()
	// it is recommended that the contents of ReportError be
	// non-blocking
	ThreadReportError(err ThreadError)

	// For recording which thread the threadable is running on
	SetThread(t *Thread)
	GetThread() *Thread

	// Threadables are closed if they encounter a fatal error
	// If ThreadInitialize(thread) was called, Close must call
	// thread.CloseThreadable to ensure ThreadClose is called properly.
	Close()
	IsClosed() bool
}

// For things which need to be run on a consistent thread
type Thread struct {
	threadables         chan Threadable
	closeThreadables    chan Threadable
	current             Threadable
	activateErrors      chan ThreadError
	deactivateErrors    chan ThreadError
	closed, forceClosed bool
}

func CreateThread() *Thread {
	thread := &Thread{
		threadables:      make(chan Threadable),
		closeThreadables: make(chan Threadable),
		activateErrors:   make(chan ThreadError),
		deactivateErrors: make(chan ThreadError),
	}
	go thread.run()
	return thread
}

// Closing a thread deactivates any active context
// The thread will remain until all threadables that were
// initialized on the thread are closed
func (thread *Thread) Close() {
	if thread.closed {
		return
	}

	thread.SetActive(nil)
	close(thread.threadables)
	thread.closed = true
}

// Forcibly close all initialized items.
// You probably don't want to do this.
func (thread *Thread) ForceClose() {
	if thread.forceClosed {
		return
	}

	thread.Close()
	close(thread.closeThreadables)
	thread.forceClosed = true
}

// Deactivates the current threaded item and activate the new threaded item.
// Passing the same value as the previous call does nothing.
// passing nil deactivates just deactivates the current running item
// Returns nil if the new context is activated without error.
// Otherwise the error that prevented the context from being
// activated is returned
func (thread *Thread) SetActive(item Threadable) error {
	// if the context is already on this thread, do nothing
	if thread.current == item {
		return nil
	}

	if item == nil { // Avoid null dereference
	} else if item.GetThread() != nil {
		// Error if the context is currently on a different thread
		return errors.New("Context active on another thread")
	} else if item.IsClosed() {
		// Error if the context is closed
		return errors.New("Context is closed")
	}

	// Record that the old context stopping
	if thread.current != nil {
		thread.current.SetThread(nil)
	}

	// store these for closing
	old := thread.current

	// Record the new thread running
	if item != nil {
		item.SetThread(thread)
	}
	thread.current = item

	// mail off the context to the runner
	thread.threadables <- item

	// check the old context
	if deactivate_err := <-thread.deactivateErrors; deactivate_err != nil {
		old.Close()
	}

	// check the new context
	if activate_err := <-thread.activateErrors; activate_err != nil {
		// Record the nil thread running
		item.SetThread(nil)
		thread.current = nil

		item.Close()
		return activate_err
	}

	return nil
}

// Get active returns the currently active context
func (thread *Thread) GetActive() Threadable {
	return thread.current
}

// Cause ThreadClose to be called
func (thread *Thread) CloseThreadable(item Threadable) {
	if item == nil {
		return
	}
	thread.closeThreadables <- item
}

func (thread *Thread) run() {
	// Lock the context to a thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	initialized := make(map[Threadable]bool)

	var deactivate_err ThreadError
threadables_loop:
	for {
		var current_item Threadable
		select {
		case v, ok := <-thread.threadables:
			if !ok {
				break threadables_loop
			}
			current_item = v

		case v := <-thread.closeThreadables:
			if !initialized[v] {
				break
			}
			delete(initialized, v)
			v.ThreadClose(thread)
			continue
		}
		// send out the last deactivate error and clear it
		thread.deactivateErrors <- deactivate_err
		deactivate_err = nil

	activate:
		if current_item == nil {
			panic("context should not be nil")
		}

		// Initialize the threadable and report errors
		if !current_item.ThreadIsInitialized() {
			if err := current_item.ThreadInitialize(thread); err != nil {
				current_item.ThreadReportError(err)
				thread.activateErrors <- err
				continue
			}
			initialized[current_item] = true
		}

		// Activate the threadable and report errors
		if err := current_item.ThreadActivate(thread); err != nil {
			current_item.ThreadReportError(err)
			thread.activateErrors <- err
			continue
		}

		// Send a successful activate
		thread.activateErrors <- nil

	run_commands:
		select {
		case v, ok := <-thread.threadables:
			// if the channel closes, we're done
			if !ok {
				break threadables_loop
			}

			// try to deactivate the current context
			if err := current_item.ThreadDeactivate(thread); err != nil {
				current_item.ThreadReportError(err)
				thread.deactivateErrors <- err
			} else {
				// send a successful deactivate
				thread.deactivateErrors <- nil
			}

			current_item = v

			// try to activate the context			
			if current_item == nil {
				// send a successful "activate"
				thread.activateErrors <- nil
				continue threadables_loop
			}
			goto activate

		case v := <-thread.closeThreadables:
			if !initialized[v] {
				break
			}
			delete(initialized, v)
			v.ThreadClose(thread)

			// wait for the next command
			goto run_commands

		case f := <-current_item.ThreadCommands():
			if err := f(thread, current_item); err != nil {
				current_item.ThreadReportError(err)
				if err.Fatal() {
					deactivate_err = err
					continue threadables_loop
				}
			}

			// wait for the next command
			goto run_commands
		}
	}

	// Loop over all the nicely closing threads
	for v := range thread.closeThreadables {
		if !initialized[v] {
			continue
		}

		delete(initialized, v)
		v.ThreadClose(thread)

		if len(initialized) != 0 {
			break
		}
	}

	// Force close anything else
	for v := range initialized {
		v.ThreadClose(thread)
	}
}
