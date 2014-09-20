// Copyright © 2012 Popog
package glml

import (
	"image"
)

// Windows contain an context, but note that fatal errors on the context will not close the window
// They will just make it mostly useless as you cannot spawn a new context.
type Window struct {
	initialize func(c *Window) ThreadError // The initialization function. Nil if already initialized

	thread   *Thread
	internal windowInternal
	context  *Context
}

// Construct a new window
//
// monitor  The monitor to create the window on. If nil, the default monitor is used.
// mode     Video mode to use (defines the width, height and depth of the rendering area of the window).
// title    Title of the window.
// style    Customize the look and behaviour of the window (borders, title bar, resizable, closable, ...)
//          If style is StyleFullscreen, then mode must be a valid video mode.
// settings Additional settings for the underlying OpenGL context.
func CreateWindow(monitor *Monitor, mode VideoMode, title string, style WindowStyle, settings ContextSettings) (*Window, error) {
	// Check the style
	if err := style.Check(); err != nil {
		return nil, err
	}

	// Get the default monitor if we need it
	if monitor == nil || !monitor.IsValid() {
		monitor = GetDefaultMonitor()
	}

	// initialize our window
	w := &Window{
		initialize: func(w *Window) ThreadError {
			return w.internal.initialize(monitor, mode, title, style)
		},
	}
	w.context = createFromOwner(settings, &w.internal, mode.BitsPerPixel)

	return w, nil
}

// The channel for input functions to run on this window.
func (w *Window) Commands() chan<- func(thread *Thread, t Threadable) ThreadError {
	return w.context.Commands()
}

// The error reporting channel
func (w *Window) Errors() <-chan ThreadError {
	return w.context.Errors()
}

// Literally GetThread() != nil
func (w *Window) IsActive() bool {
	return w.context.IsActive()
}

// Returns the thread the window is running on or nil if it is not currently
// running on a thread
func (w *Window) GetThread() *Thread {
	return w.context.GetThread()
}

// Expects to be called by Thread
// Sets the thread
func (w *Window) SetThread(thread *Thread) {
	w.context.SetThread(thread)
}

// Gets the thread that the window was initialized on.
// Returns nil if window has not been initialized
func (w *Window) InitialThread() *Thread {
	return w.context.InitialThread()
}

// Closing a window will deactivate it if necessary. A
// closed context is inactive and cannot be reactivated.
// all of window resources are freed, but not resources
// loaded by the user.
func (w *Window) Close() {
	if w.IsClosed() {
		return
	}

	// deactivate if need be
	if w.IsActive() {
		w.GetThread().SetActive(nil)
	}

	// if deactivating caused us to close, we're done
	if w.IsClosed() {
		return
	}

	// Close the threadable on the thread it was initialized on
	if w.ThreadIsInitialized() {
		w.InitialThread().CloseThreadable(w)
	}

	w.context.closed = true
}

// returns true if a context has been closed
func (w *Window) IsClosed() bool {
	return w.context.IsClosed()
}

// Expects to be called on a Thread
func (w *Window) ThreadIsInitialized() bool {
	return w.initialize == nil
}

// Perform some common internal initializations
func (w *Window) ThreadInitialize(thread *Thread) ThreadError {
	if w.ThreadIsInitialized() {
		panic("ThreadIsInitialized")
	}

	// initialize the window
	if err := w.initialize(w); err != nil {
		if err.Fatal() {
			return err
		}
		w.ThreadReportError(err)
	}

	// initialize the context
	if err := w.context.ThreadInitialize(thread); err != nil {
		if err.Fatal() {
			return err
		}
		w.ThreadReportError(err)
	}

	// Setup default behaviours (to get a consistent behaviour across different implementations)
	if err := w.ThreadSetVisible(thread, true); err != nil {
		if err.Fatal() {
			return err
		}
		w.ThreadReportError(err)
	}
	if err := w.ThreadSetMouseCursorVisible(thread, true); err != nil {
		if err.Fatal() {
			return err
		}
		w.ThreadReportError(err)
	}
	if err := w.ThreadSetVSyncEnabled(false); err != nil {
		if err.Fatal() {
			return err
		}
		w.ThreadReportError(err)
	}
	if err := w.ThreadSetKeyRepeatEnabled(thread, true); err != nil {
		if err.Fatal() {
			return err
		}
		w.ThreadReportError(err)
	}

	// Set ThreadInitialize to true
	w.initialize = nil
	return nil
}

// Expects to be called on a Thread
func (w *Window) ThreadClose(thread *Thread) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	// close the context
	w.context.ThreadClose(thread)

	// close the window
	if err := w.internal.close(); err != nil {
		w.ThreadReportError(err)
	}

	// TODO: Update the fullscreen window
}

// Expects to be called on a Thread
func (w *Window) ThreadActivate(thread *Thread) ThreadError {
	return w.context.ThreadActivate(thread)
}

// Expects to be called on a Thread
func (w *Window) ThreadDeactivate(thread *Thread) ThreadError {
	return w.context.ThreadDeactivate(thread)
}

// Expects to be called on a Thread
func (w *Window) ThreadCommands() <-chan func(thread *Thread, t Threadable) ThreadError {
	return w.context.ThreadCommands()
}

// Expects to be called on a Thread
// Sends an error to Window.Errors()
func (w *Window) ThreadReportError(err ThreadError) {
	w.context.ThreadReportError(err)
}

// Expects to be called on a Thread
// Swap the front and back buffers to display on screen what has been rendered
// so far.
//
// This function is typically called after all OpenGL rendering
// has been done for the current frame, in order to show
// it on screen.
func (w *Window) ThreadSwapBuffers() ThreadError {
	return w.context.ThreadSwapBuffers()
}

// A thread command wrapper for Window.ThreadSwapBuffers
func WindowThreadSwapBuffers(_ *Thread, t Threadable) ThreadError {
	return t.(*Window).ThreadSwapBuffers()
}

// Expects to be called on a Thread
// Retrieve the OpenGL context settings
//
// Note that these settings may be different from what was
// requested if one or more settings were not supported. In
// this case, the closest available match was chosen.
func (w *Window) ThreadGetSettings() (ContextSettings, ThreadError) {
	return w.context.ThreadGetSettings()
}

// A thread command helper for Window.ThreadGetSettings
func WindowThreadGetSettings(results chan<- ContextSettings) func(thread *Thread, t Threadable) ThreadError {
	return func(_ *Thread, t Threadable) ThreadError {
		s, err := t.(*Window).ThreadGetSettings()
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
func (w *Window) ThreadSetVSyncEnabled(enabled bool) ThreadError {
	return w.context.ThreadSetVSyncEnabled(enabled)
}

// A thread command helper for Window.ThreadSetVSyncEnabled
func WindowThreadSetVSyncEnabled(enabled bool) func(thread *Thread, t Threadable) ThreadError {
	return func(_ *Thread, t Threadable) ThreadError {
		return t.(*Window).ThreadSetVSyncEnabled(enabled)
	}
}

// Expects to be called on InitialThread()
// Get the position of the window
func (w *Window) ThreadGetPosition(thread *Thread) (x, y int) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.getPosition()
}

// Expects to be called on InitialThread()
// Change the position of the window on screen
//
// This function only works for top-level windows
// (i.e. it will be ignored for windows created from
// the handle of a child window/control).
func (w *Window) ThreadSetPosition(thread *Thread, x, y int) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	w.internal.setPosition(x, y)
}

// Expects to be called on InitialThread()
// Get the size of the rendering region of the window
//
// The size doesn't include the titlebar and borders
// of the window.
func (w *Window) ThreadGetSize(thread *Thread) (x, y uint) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.getSize()
}

// Expects to be called on InitialThread()
// Change the size of the rendering region of the window
func (w *Window) ThreadSetSize(thread *Thread, x, y uint) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	w.internal.setSize(x, y)
}

// Expects to be called on InitialThread()
// Change the title of the window
func (w *Window) ThreadSetTitle(thread *Thread, title string) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	w.internal.setTitle(title)
}

// Expects to be called on InitialThread()
// Change the window's icon
// The OS default icon is used by default.
func (w *Window) ThreadSetIcon(thread *Thread, icon image.Image) error {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.setIcon(icon)
}

// Expects to be called on InitialThread()
// Show or hide the window
//
// The window is shown by default.
func (w *Window) ThreadSetVisible(thread *Thread, visible bool) ThreadError {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.setVisible(visible)
}

// Expects to be called on InitialThread()
// Show or hide the mouse cursor
//
// The mouse cursor is visible by default.
func (w *Window) ThreadSetMouseCursorVisible(thread *Thread, visible bool) ThreadError {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.setMouseCursorVisible(visible)
}

// Expects to be called on InitialThread()
// Enable or disable automatic key-repeat
//
// If key repeat is enabled, you will receive repeated
// KeyPressed events while keeping a key pressed. If it is disabled,
// you will only get a single event when the key is pressed.
//
// Key repeat is enabled by default.
func (w *Window) ThreadSetKeyRepeatEnabled(thread *Thread, enabled bool) ThreadError {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.setKeyRepeatEnabled(enabled)
}

// Expects to be called on InitialThread()
// Get the contents of the window's event queue and evacuate it.
// 
// If block is true, this function will wait for an event. If block is false
// and there are no pending events then the return value is nil.
func (w *Window) ThreadPollEvents(thread *Thread, block bool) ([]Event, []ThreadError) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.pollEvents(block)
}

// Get the current position of the mouse in window coordinates
// Panics if window is nil or closed.
func (w *Window) ThreadGetMousePosition(thread *Thread) (x, y int, err ThreadError) {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.getMousePosition()
}

// Set the current position of the mouse in window coordinates
// Panics if window is nil or closed.
func (w *Window) ThreadSetMousePosition(thread *Thread, x, y int) ThreadError {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.setMousePosition(x, y)
}

// Expects to be called on InitialThread()
// Get the OS-specific handle of the window
//
// The type of the returned handle is WindowHandle,
// which is a typedef to the handle type defined by the OS.
// You shouldn't need to use this function, unless you have
// very specific stuff to implement that this library doesn't support,
// or implement a temporary workaround until a bug is fixed.
func (w *Window) ThreadGetSystemHandle(thread *Thread) WindowHandle {
	if w.InitialThread() != thread {
		panic("thread is not initialThread")
	}

	return w.internal.getSystemHandle()
}
