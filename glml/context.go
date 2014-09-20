// Copyright © 2012 Popog
package glml

var (
	sharedContext       = CreateContext()
	sharedContextThread = CreateThread()
)

func init() {
	sharedContextThread.SetActive(sharedContext)
}

type Context struct {
	initialThread *Thread                                             // The thread this context was initialized on.
	thread        *Thread                                             // The thread this context is running on.
	commands      chan func(thread *Thread, t Threadable) ThreadError // The channel for input functions to run on this context.
	errors        chan ThreadError                                    // The error reporting channel
	initialize    func(c *Context) ThreadError                        // The initialization function. Nil if already initialized
	closed        bool                                                // Whether or not Close has already been called.
	internal      contextInternal                                     // The os specific context implementation. This should only be touched on threads.
}

// Create a context with default settings and dimensions
func CreateContext() *Context {
	return &Context{
		commands: make(chan func(thread *Thread, t Threadable) ThreadError),
		errors:   make(chan ThreadError),
		initialize: func(c *Context) ThreadError {
			return c.internal.initialize()
		},
	}
}

// A context with specific settings and back buffer dimensions
func CreateContextFromSettings(settings ContextSettings, width, height int) *Context {
	return &Context{
		commands: make(chan func(thread *Thread, t Threadable) ThreadError),
		errors:   make(chan ThreadError),
		initialize: func(c *Context) ThreadError {
			return c.internal.initializeFromSettings(settings, width, height)
		},
	}
}

// Initializes a context for an existing window
func createFromOwner(settings ContextSettings, owner *windowInternal, bitsPerPixel uint) *Context {
	return &Context{
		commands: make(chan func(thread *Thread, t Threadable) ThreadError),
		errors:   make(chan ThreadError),
		initialize: func(c *Context) ThreadError {
			return c.internal.initializeFromOwner(settings, owner, bitsPerPixel)
		},
	}
}

// The channel for input functions to run on this context.
func (c *Context) Commands() chan<- func(thread *Thread, t Threadable) ThreadError {
	return c.commands
}

// The error reporting channel
func (c *Context) Errors() <-chan ThreadError {
	return c.errors
}

// Literally GetThread() != nil
func (c *Context) IsActive() bool {
	return c.GetThread() != nil
}

// Returns the thread the context is running on or nil if it is not currently
// running on a thread
func (c *Context) GetThread() *Thread {
	return c.thread
}

// Expects to be called by Thread
// Sets the thread
func (c *Context) SetThread(thread *Thread) {
	c.thread = thread
}

// Gets the thread that the context was initialized on.
// Returns nil if context has not been initialized
func (c *Context) InitialThread() *Thread {
	return c.initialThread
}

// Closing a context will deactivate it if necessary. A
// closed context is inactive and cannot be reactivated.
// all of internal resources are freed, but not resources
// loaded by the user.
func (c *Context) Close() {
	if c.IsClosed() {
		return
	}

	if c == sharedContext {
		panic("sharedContext can never close")
	}

	// deactivate if need be
	if c.IsActive() {
		c.GetThread().SetActive(nil)
	}

	// if deactivating caused us to close, we're done
	if c.IsClosed() {
		return
	}

	// Close the threadable on the thread it was initialized on
	if c.ThreadIsInitialized() {
		c.InitialThread().CloseThreadable(c)
	}

	c.closed = true
}

// returns true if a context has been closed
func (c *Context) IsClosed() bool {
	return c.closed
}

// Expects to be called on a Thread
func (c *Context) ThreadIsInitialized() bool {
	return c.initialize == nil
}

// Expects to be called on a Thread
func (c *Context) ThreadInitialize(thread *Thread) ThreadError {
	if c.ThreadIsInitialized() {
		panic("ThreadIsInitialized")
	}

	if c != sharedContext {
		pause_signal := make(chan bool)
		sharedContext.Commands() <- func(_ *Thread, t Threadable) ThreadError {
			return t.(*Context).internal.pause(pause_signal)
		}

		<-pause_signal
		defer func() { pause_signal <- true }()

		// Try to take the shared context
		if err := sharedContext.internal.take(); err != nil {
			return err
		}
		defer sharedContext.internal.release()
	}

	if err := c.initialize(c); err != nil {
		return err
	}

	c.initialThread = thread
	c.initialize = nil

	return nil
}

// Expects to be called on a Thread
func (c *Context) ThreadClose(thread *Thread) {
	if c.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	if err := c.internal.close(); err != nil {
		c.ThreadReportError(err)
	}
}

// Expects to be called on a Thread
func (c *Context) ThreadActivate(*Thread) ThreadError {
	if c.closed {
		panic("context is closed")
	}
	return c.internal.activate()
}

// Expects to be called on a Thread
func (c *Context) ThreadDeactivate(*Thread) ThreadError {
	if c.closed {
		panic("context is closed")
	}
	return c.internal.deactivate()
}

// Expects to be called on a Thread
func (c *Context) ThreadCommands() <-chan func(thread *Thread, t Threadable) ThreadError {
	return c.commands
}

// Expects to be called on a Thread
// Sends an error to Context.Errors()
func (c *Context) ThreadReportError(err ThreadError) {
	go func() { c.errors <- err }()
}

// Expects to be called on a Thread
// Swap the front and back buffers to display on screen what has been rendered
// so far.
//
// This function is typically called after all OpenGL rendering
// has been done for the current frame, in order to show
// it on screen.
func (c *Context) ThreadSwapBuffers() ThreadError {
	return c.internal.swapBuffers()
}

// A thread command wrapper for Context.ThreadSwapBuffers
func ContextThreadSwapBuffers(_ *Thread, t Threadable) ThreadError {
	return t.(*Context).ThreadSwapBuffers()
}

// Expects to be called on a Thread
// Retrieve the OpenGL context settings
//
// Note that these settings may be different from what was
// requested if one or more settings were not supported. In
// this case, the closest available match was chosen.
func (c *Context) ThreadGetSettings() (ContextSettings, ThreadError) {
	return c.internal.getSettings()
}

// A thread command helper for Context.ThreadGetSettings
// If an error occurs, results will not be sent, so be sure to check Context.Errors()
func ContextThreadGetSettings(results chan<- ContextSettings) func(thread *Thread, t Threadable) ThreadError {
	return func(_ *Thread, t Threadable) ThreadError {
		s, err := t.(*Context).ThreadGetSettings()
		if err != nil {
			return err
		}

		results <- s
		return nil
	}
}

// Expects to be called on a Thread
// Enable or disable vertical synchronization.
//
// Activating vertical synchronization will limit the number
// of frames displayed to the refresh rate of the monitor.
// This can avoid some visual artifacts, and limit the framerate
// to a good value (but not constant across different computers).
func (c *Context) ThreadSetVSyncEnabled(enabled bool) ThreadError {
	return c.internal.setVerticalSyncEnabled(enabled)
}

// A thread command helper for Context.ThreadSetVSyncEnabled
func ContextThreadSetVSyncEnabled(enabled bool) func(thread *Thread, t Threadable) ThreadError {
	return func(_ *Thread, t Threadable) ThreadError {
		return t.(*Context).ThreadSetVSyncEnabled(enabled)
	}
}
