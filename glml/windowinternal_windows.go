// Copyright © 2012 Popog
package glml

// #include "helper_windows.h"
// 
// extern LRESULT CALLBACK (*pGlobalOnEvent)(HWND handle, UINT message, WPARAM wParam, LPARAM lParam);
import "C"
import (
	"fmt"
	"image"
	"unsafe"
)

var className, classNameLength = utf16Convert("go-glml/window.Window")
var windowClass = C.WNDCLASSW{
	lpfnWndProc:   C.pGlobalOnEvent,
	style:         C.CS_OWNDC | C.CS_HREDRAW | C.CS_VREDRAW,
	hInstance:     C.HINSTANCE(C.GetModuleHandle(nil)),
	lpszClassName: className,
}

func init() {
	C.RegisterClassW(&windowClass)
}

type WindowHandle struct {
	Handle C.HWND
}

func (wh WindowHandle) IsValid() bool {
	return wh.Handle != nil
}

type windowInternal struct {
	events      []Event       // The events from polling
	eventErrors []ThreadError // the errors from polling

	devMode              *C.DEVMODEW    // The fullscreen settings
	monitor              *Monitor       // The monitor we're fullscreen on (nil if we're not a fullscreen window)
	window               WindowHandle   // Win32 handle of the window
	callback             unsafe.Pointer // Stores the original event callback function of the control
	cursor               C.HCURSOR      // The system cursor to display into the window
	icon                 C.HICON        // Custom icon assigned to the window
	keyRepeatEnabled     bool           // Automatic key-repeat state for keydown events
	isCursorIn           bool           // Is the mouse cursor in the window's area ?
	lastSizeX, lastSizeY uint           // The last handled size of the window
	resizing             bool           // Is the window being resized ?
	inactive, minimized  bool           // The current active or not state of the window

}

// Creates the window. This function expects not to be called on a ContextThread
func (wi *windowInternal) initialize(monitor *Monitor, mode VideoMode, title string, style WindowStyle) ThreadError {
	// Compute position and size
	screenDC := C.GetDC(nil)
	width := C.int(mode.Width)
	height := C.int(mode.Height)
	left := (C.GetDeviceCaps(screenDC, C.HORZRES) - width) / 2
	top := (C.GetDeviceCaps(screenDC, C.VERTRES) - height) / 2
	C.ReleaseDC(nil, screenDC)

	// Choose the window style according to the Style parameter
	win32Style := C.DWORD(C.WS_VISIBLE)
	if style == WindowStyleNone {
		win32Style |= C.WS_POPUP
	} else {
		if style&WindowStyleTitlebar != 0 {
			win32Style |= C.WS_CAPTION | C.WS_MINIMIZEBOX
		}
		if style&WindowStyleResize != 0 {
			win32Style |= C.WS_THICKFRAME | C.WS_MAXIMIZEBOX
		}
		if style&WindowStyleClose != 0 {
			win32Style |= C.WS_SYSMENU
		}
	}

	// In windowed mode, adjust width and height so that window will have the requested client area
	fullscreen := style&WindowStyleFullscreen != 0
	if !fullscreen {
		rectangle := C.RECT{
			left:   C.LONG(left),
			top:    C.LONG(top),
			right:  C.LONG(left + width),
			bottom: C.LONG(top + height),
		}
		C.__AdjustWindowRect(&rectangle, win32Style, C.FALSE)
		left = C.int(rectangle.left)
		top = C.int(rectangle.top)
		width = C.int(rectangle.right - rectangle.left)
		height = C.int(rectangle.bottom - rectangle.top)
	}
	wTitle, _ := utf16Convert(title)
	wi.window.Handle = C.CreateWindowExW(0, className, wTitle, win32Style, left, top, width, height, nil, nil, windowClass.hInstance, C.LPVOID(wi))

	// Switch to fullscreen if requested
	if fullscreen {
		wi.switchToFullscreen(monitor, mode)
	}

	return nil
}

func (wi *windowInternal) initializeFromExisting(window WindowHandle, settings ContextSettings) ThreadError {
	wi.window = window

	// We change the event procedure of the control (it is important to save the old one)
	C.__SetWindowLongPtr(wi.window.Handle, C.GWLP_USERDATA, unsafe.Pointer(wi))
	callback := C.__SetWindowLongPtr(wi.window.Handle, C.GWLP_WNDPROC, unsafe.Pointer(C.pGlobalOnEvent))
	wi.callback = unsafe.Pointer(uintptr(callback))

	return nil
}

func (wi *windowInternal) close() ThreadError {
	// Destroy the custom icon, if any
	if wi.icon != nil {
		C.DestroyIcon(wi.icon)
	}

	if wi.callback == nil {
		// Destroy the window
		if wi.window.IsValid() {
			C.DestroyWindow(wi.window.Handle)
		}
	} else {
		// The window is external : remove the hook on its message callback
		C.__SetWindowLongPtr(wi.window.Handle, C.GWLP_WNDPROC, wi.callback)
	}

	return nil
}

func (wi *windowInternal) getSystemHandle() WindowHandle {
	return wi.window
}

// Get the contents of the window's event queue and evacuate it.
func (wi *windowInternal) pollEvents(block bool) ([]Event, []ThreadError) {
	// clear the events before and after
	wi.events = nil
	defer func() {
		wi.events = nil
		wi.eventErrors = nil
	}()

	// We process the window events only if we own it
	if wi.callback != nil {
		return nil, nil
	}

	// Wait for messages if we're blocking
	if block {
		if C.WaitMessage() == 0 {
		// TODO error check
		}
	}

	for message := (C.MSG{}); C.__PeekMessage(&message, wi.window.Handle, 0, 0, C.PM_REMOVE) == C.TRUE; {
		C.TranslateMessage(&message)
		C.DispatchMessage(&message)

		// if the last one is fatal, return
		if len(wi.eventErrors) != 0 && wi.eventErrors[len(wi.eventErrors)-1].Fatal() {
			break
		}
	}

	return wi.events, wi.eventErrors
}

// Get the position of the window
func (wi *windowInternal) getPosition() (x, y int) {
	var rect C.RECT
	C.__GetWindowRect(wi.window.Handle, &rect)
	return int(rect.left), int(rect.top)
}

// Change the position of the window on screen
//
// This function only works for top-level windows
// (i.e. it will be ignored for windows created from
// the handle of a child window/control).
func (wi *windowInternal) setPosition(x, y int) {
	C.SetWindowPos(wi.window.Handle, nil, C.int(x), C.int(y), 0, 0, C.SWP_NOSIZE|C.SWP_NOZORDER)
}

// Get the size of the rendering region of the window
//
// The size doesn't include the titlebar and borders
// of the window.
func (wi *windowInternal) getSize() (x, y uint) {
	var rect C.RECT
	C.__GetWindowRect(wi.window.Handle, &rect)
	return uint(rect.right - rect.left), uint(rect.bottom - rect.top)
}

// Change the size of the rendering region of the window
func (wi *windowInternal) setSize(x, y uint) {
	// SetWindowPos wants the total size of the window (including title bar and borders),
	// so we have to compute it
	rect := C.RECT{0, 0, C.LONG(x), C.LONG(y)}
	style := C.GetWindowLong(wi.window.Handle, C.GWL_STYLE)
	C.__AdjustWindowRect(&rect, C.DWORD(style), C.FALSE)
	width := C.int(rect.right - rect.left)
	height := C.int(rect.bottom - rect.top)

	C.SetWindowPos(wi.window.Handle, nil, 0, 0, width, height, C.SWP_NOMOVE|C.SWP_NOZORDER)
}

// Change the title of the window
func (wi *windowInternal) setTitle(title string) {
	t, _ := utf16Convert(title)
	C.SetWindowTextW(wi.window.Handle, t)
}

// Change the window's icon
func (wi *windowInternal) setIcon(icon image.Image) error {
	// First destroy the previous one
	if wi.icon != nil {
		C.DestroyIcon(wi.icon)
	}

	type BGRA [4]uint8

	// Windows wants BGRA pixels: swap red and blue channels
	Rect := icon.Bounds()
	iconPixels := make([]BGRA, Rect.Dy()*Rect.Dx())
	for i, y := 0, 0; y < Rect.Dy(); y++ {
		for x := 0; x < Rect.Dx(); x++ {
			r, g, b, a := icon.At(x, y).RGBA()
			iconPixels[i][0] = uint8(b)
			iconPixels[i][1] = uint8(g)
			iconPixels[i][2] = uint8(r)
			iconPixels[i][3] = uint8(a)
		}
	}

	// Create the icon from the pixel array
	width, height := C.int(Rect.Dx()), C.int(Rect.Dy())
	wi.icon = C.CreateIcon(C.GetModuleHandle(nil), width, height, 1, 32, nil, (*C.BYTE)(&iconPixels[0][0]))

	if wi.icon == nil {
		return fmt.Errorf("could not create icon (%d)", C.GetLastError())
	}

	C.SendMessage(wi.window.Handle, C.WM_SETICON, C.ICON_BIG, C.LPARAM(uintptr(unsafe.Pointer(wi.icon))))
	C.SendMessage(wi.window.Handle, C.WM_SETICON, C.ICON_SMALL, C.LPARAM(uintptr(unsafe.Pointer(wi.icon))))

	return nil
}

// Show or hide the window
func (wi *windowInternal) setVisible(visible bool) ThreadError {
	nCmdShow := C.int(C.SW_HIDE)
	if visible {
		nCmdShow = C.SW_SHOW
	}

	C.ShowWindow(wi.window.Handle, nCmdShow)
	return nil
}

// Show or hide the mouse cursor
func (wi *windowInternal) setMouseCursorVisible(visible bool) ThreadError {
	wi.cursor = nil
	if visible {
		wi.cursor = C.LoadCursorW(nil, C.__IDC_ARROW)
	}

	C.SetCursor(wi.cursor)
	return nil
}

// Enable or disable automatic key-repeat
//
// If key repeat is enabled, you will receive repeated
// KeyPressed events while keeping a key pressed. If it is disabled,
// you will only get a single event when the key is pressed.
func (wi *windowInternal) setKeyRepeatEnabled(enabled bool) ThreadError {
	wi.keyRepeatEnabled = enabled
	return nil
}

// Get the current position of the mouse in window coordinates
func (wi *windowInternal) getMousePosition() (x, y int, err ThreadError) {
	if !wi.window.IsValid() {
		// TODO ERROR
		return
	}

	var point C.POINT
	if C.__GetCursorPos(&point) == 0 {
		err = NewThreadError(fmt.Errorf("GetCursorPos (%d)", C.GetLastError()), false)
		return
	}

	if C.__ScreenToClient(wi.window.Handle, &point) == 0 {
		err = NewThreadError(fmt.Errorf("ScreenToClient (%d)", C.GetLastError()), false)
		return
	}
	return int(point.x), int(point.y), nil
}

// Set the current position of the mouse in window coordinates
func (wi *windowInternal) setMousePosition(x, y int) ThreadError {
	if !wi.window.IsValid() {
		// TODO ERROR
		return nil
	}

	point := C.POINT{x: C.LONG(x), y: C.LONG(y)}
	if C.__ScreenToClient(wi.window.Handle, &point) == 0 {
		return NewThreadError(fmt.Errorf("ScreenToClient (%d)", C.GetLastError()), false)
	}

	if C.SetCursorPos(C.int(x), C.int(y)) == 0 {
		return NewThreadError(fmt.Errorf("SetCursorPos (%d)", C.GetLastError()), false)
	}

	return nil
}

func (wi *windowInternal) switchToFullscreen(monitor *Monitor, mode VideoMode) error {
	devMode := C.DEVMODEW{
		dmSize:       C.DEVMODEW_size,
		dmPelsWidth:  C.DWORD(mode.Width),
		dmPelsHeight: C.DWORD(mode.Height),
		dmBitsPerPel: C.DWORD(mode.BitsPerPixel),
		dmFields:     C.DM_PELSWIDTH | C.DM_PELSHEIGHT | C.DM_BITSPERPEL,
	}

	// Apply fullscreen mode
	if err := ChangeDisplaySettingsExW(&monitor.internal.info.szDevice[0], &devMode, nil, C.CDS_FULLSCREEN, nil); err != nil {
		return err
	}

	// Make the window flags compatible with fullscreen mode
	C.SetWindowULong(wi.window.Handle, C.GWL_STYLE, C.WS_POPUP|C.WS_CLIPCHILDREN|C.WS_CLIPSIBLINGS)
	C.SetWindowLong(wi.window.Handle, C.GWL_EXSTYLE, C.WS_EX_APPWINDOW)

	// Resize the window so that it fits the entire screen
	C.SetWindowPos(wi.window.Handle, HWND_TOP, 0, 0, C.int(mode.Width), C.int(mode.Height), C.SWP_FRAMECHANGED)
	C.ShowWindow(wi.window.Handle, C.SW_SHOW)

	// Set this as the current fullscreen window
	wi.monitor = monitor
	wi.devMode = &devMode

	return nil
}

func (wi *windowInternal) cleanup() {
	// Restore the previous video mode (in case we were running in fullscreen)
	// TODO
	if wi.monitor != nil && wi.monitor.IsValid() && !wi.minimized {
		if err := ChangeDisplaySettingsExW(&wi.monitor.internal.info.szDevice[0], nil, nil, 0, nil); err != nil {
			// TODO: Error handling
			panic(err)
		}
	}

	// Unhide the mouse cursor (in case it was hidden)
	wi.setMouseCursorVisible(true)
}

func (wi *windowInternal) processEvent(message C.UINT, wParam C.WPARAM, lParam C.LPARAM) (events []Event, eventErrors []ThreadError) {
	// Don't process any message until window is created
	if wi.window.Handle == nil {
		return
	}

	switch message {
	case C.WM_DESTROY: // Destroy event
		// Here we must cleanup resources !
		wi.cleanup()
	case C.WM_ACTIVATE:
		// TODO fullscreen handling
		if wi.monitor == nil || !wi.monitor.IsValid() {
			break
		}

		// Were we deactivated/iconified?
		wi.inactive = C.__LOWORD(C.DWORD(wParam)) == C.WA_INACTIVE
		minimized := C.__HIWORD(C.DWORD(wParam)) != 0

		if (wi.inactive || minimized) && !wi.minimized {
			// _glfwInputDeactivation();

			// If we are in fullscreen mode we need to iconify
			if wi.monitor != nil && wi.monitor.IsValid() {
				// Do we need to manually iconify?
				if err := ChangeDisplaySettingsExW(&wi.monitor.internal.info.szDevice[0], nil, nil, 0, nil); err != nil {
					// TODO: Error handling
					panic(err)
				}

				if !minimized {
					// Minimize window
					// C.SetWindowULong(wi.window.Handle, C.GWL_STYLE, C.WS_POPUP)
					// C.SetWindowLong(wi.window.Handle, C.GWL_EXSTYLE, 0)
					// C.SetWindowPos(wi.window.Handle, HWND_BOTTOM, 1, 1, 10, 10, C.SWP_HIDEWINDOW)
					C.ShowWindow(wi.window.Handle, C.SW_MINIMIZE)
					minimized = true
				}

				// Restore the original desktop resolution
				if err := ChangeDisplaySettingsExW(&wi.monitor.internal.info.szDevice[0], nil, nil, 0, nil); err != nil {
					// TODO: Error handling
					panic(err)
				}

			}

			// Unlock mouse if locked
			// if !_glfwWin.oldMouseLockValid {
			// 	_glfwWin.oldMouseLock = _glfwWin.mouseLock;
			// 	_glfwWin.oldMouseLockValid = GL_TRUE;
			// 	glfwEnable( GLFW_MOUSE_CURSOR );
			// }
		} else if !wi.inactive || !minimized {
			// If we are in fullscreen mode we need to maximize
			if wi.monitor != nil && wi.monitor.IsValid() && wi.minimized {
				// Change display settings to the user selected mode
				if err := ChangeDisplaySettingsExW(&wi.monitor.internal.info.szDevice[0], wi.devMode, nil, C.CDS_FULLSCREEN, nil); err != nil {
					// TODO error handling
					panic(err)
				}

				// Do we need to manually restore window?
				if minimized {
					// Restore window
					C.ShowWindow(wi.window.Handle, C.SW_RESTORE)
					minimized = false

					// Activate window
					C.ShowWindow(wi.window.Handle, C.SW_SHOW)
					// setForegroundWindow( _glfwWin.window );
					C.SetFocus(wi.window.Handle)
				}

				// Lock mouse, if necessary
				// if _glfwWin.oldMouseLockValid && _glfwWin.oldMouseLock {
				// 	glfwDisable( GLFW_MOUSE_CURSOR );
				// }
				// _glfwWin.oldMouseLockValid = GL_FALSE;
			}
		}

		wi.minimized = minimized

	case C.WM_SETCURSOR: // Set cursor event
		// The mouse has moved, if the cursor is in our window we must refresh the cursor
		if C.__LOWORD(C.DWORD(lParam)) == C.HTCLIENT {
			C.SetCursor(wi.cursor)
		}

	case C.WM_CLOSE: // Close event
		events = append(events, WindowClosedEvent{})

	case C.WM_SIZE: // Resize event
		// Consider only events triggered by a maximize or a un-maximize
		if wParam == C.SIZE_MINIMIZED || wi.resizing {
			break
		}

		// Ignore cases where the window has only been moved		
		if x, y := wi.getSize(); wi.lastSizeX == x && wi.lastSizeY == y {
			break
		}

		events = append(events, WindowResizeEvent{
			Width:  wi.lastSizeX,
			Height: wi.lastSizeY,
		})

	case C.WM_ENTERSIZEMOVE: // Start resizing
		wi.resizing = true

	case C.WM_EXITSIZEMOVE: // Stop resizing
		wi.resizing = false
		// Ignore cases where the window has only been moved
		if x, y := wi.getSize(); wi.lastSizeX == x && wi.lastSizeY == y {
			break
		} else {
			wi.lastSizeX, wi.lastSizeY = x, y
		}

		events = append(events, WindowResizeEvent{
			Width:  wi.lastSizeX,
			Height: wi.lastSizeY,
		})

	case C.WM_KILLFOCUS: // Lost focus event
		events = append(events, WindowGainedFocusEvent{})

	case C.WM_SETFOCUS: // Gain focus event
		events = append(events, WindowLostFocusEvent{})

	case C.WM_CHAR: // Text event
		if !wi.keyRepeatEnabled && lParam&(1<<30) != 0 {
			break
		}

		events = append(events, TextEnteredEvent{
			Character: rune(wParam),
		})

	case C.WM_KEYDOWN, C.WM_SYSKEYDOWN: // Keydown event
		if !wi.keyRepeatEnabled && C.__HIWORD(C.DWORD(lParam))&C.KF_REPEAT != 0 {
			break
		}

		events = append(events, KeyPressedEvent{
			Code:    virtualKeyCodeToSF(wParam, lParam),
			Alt:     C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_MENU))) != 0,
			Control: C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_CONTROL))) != 0,
			Shift:   C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_SHIFT))) != 0,
			System:  C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_LWIN))) != 0 || C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_RWIN))) != 0,
		})

	case C.WM_KEYUP, C.WM_SYSKEYUP: // Keyup event
		events = append(events, KeyReleasedEvent{
			Code:    virtualKeyCodeToSF(wParam, lParam),
			Alt:     C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_MENU))) != 0,
			Control: C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_CONTROL))) != 0,
			Shift:   C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_SHIFT))) != 0,
			System:  C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_LWIN))) != 0 || C.__HIWORD(C.DWORD(C.GetAsyncKeyState(C.VK_RWIN))) != 0,
		})

	case C.WM_MOUSEWHEEL: // Mouse wheel event
		// Mouse position is in screen coordinates, convert it to window coordinates
		position := C.POINT{
			x: C.LONG(C.__LOWORD(C.DWORD(lParam))),
			y: C.LONG(C.__HIWORD(C.DWORD(lParam))),
		}
		C.__ScreenToClient(wi.window.Handle, &position)

		events = append(events, MouseWheelEvent{
			Delta: int(int16(C.__HIWORD(C.DWORD(wParam))) / 120),
			X:     int(position.x),
			Y:     int(position.y),
		})

	case C.WM_LBUTTONDOWN, C.WM_RBUTTONDOWN: // Mouse left/right button down event
		button := mouse_vkeys_handed_map[mouseKey{message, C.GetSystemMetrics(C.SM_SWAPBUTTON) == C.TRUE}]
		events = append(events, MouseButtonPressedEvent{
			Button: button,
			X:      int(C.__LOWORD(C.DWORD(lParam))),
			Y:      int(C.__HIWORD(C.DWORD(lParam))),
		})

	case C.WM_LBUTTONUP, C.WM_RBUTTONUP: // Mouse left/right button up event
		button := mouse_vkeys_handed_map[mouseKey{message, C.GetSystemMetrics(C.SM_SWAPBUTTON) == C.TRUE}]
		events = append(events, MouseButtonReleasedEvent{
			Button: button,
			X:      int(C.__LOWORD(C.DWORD(lParam))),
			Y:      int(C.__HIWORD(C.DWORD(lParam))),
		})

	case C.WM_MBUTTONDOWN: // Mouse wheel button down event
		events = append(events, MouseButtonPressedEvent{
			Button: MouseMiddle,
			X:      int(C.__LOWORD(C.DWORD(lParam))),
			Y:      int(C.__HIWORD(C.DWORD(lParam))),
		})

	case C.WM_MBUTTONUP: // Mouse wheel button up event
		events = append(events, MouseButtonReleasedEvent{
			Button: MouseMiddle,
			X:      int(C.__LOWORD(C.DWORD(lParam))),
			Y:      int(C.__HIWORD(C.DWORD(lParam))),
		})

	case C.WM_XBUTTONDOWN: // Mouse X button down event
		event := MouseButtonPressedEvent{
			X: int(C.__LOWORD(C.DWORD(lParam))),
			Y: int(C.__HIWORD(C.DWORD(lParam))),
		}

		switch C.__HIWORD(C.DWORD(wParam)) {
		default:
			fallthrough
		case C.XBUTTON1:
			event.Button = MouseXButton1
		case C.XBUTTON2:
			event.Button = MouseXButton2

		}
		events = append(events, event)

	case C.WM_XBUTTONUP: // Mouse X button up event
		event := MouseButtonPressedEvent{
			X: int(C.__LOWORD(C.DWORD(lParam))),
			Y: int(C.__HIWORD(C.DWORD(lParam))),
		}

		switch C.__HIWORD(C.DWORD(wParam)) {
		default:
			fallthrough
		case C.XBUTTON1:
			event.Button = MouseXButton1
		case C.XBUTTON2:
			event.Button = MouseXButton2

		}
		events = append(events, event)

	case C.WM_MOUSEMOVE: // Mouse move event
		// Check if we need to generate a MouseEntered event
		if !wi.isCursorIn {
			mouseEvent := C.TRACKMOUSEEVENT{
				cbSize:    C.TRACKMOUSEEVENT_size,
				hwndTrack: wi.window.Handle,
				dwFlags:   C.TME_LEAVE,
			}
			C.__TrackMouseEvent(&mouseEvent)

			wi.isCursorIn = true

			events = append(events, MouseEnteredEvent{})
		}

		events = append(events, MouseMoveEvent{
			X: int(C.__LOWORD(C.DWORD(lParam))),
			Y: int(C.__HIWORD(C.DWORD(lParam))),
		})

	case C.WM_MOUSELEAVE: // Mouse leave event
		wi.isCursorIn = false
		events = append(events, MouseLeftEvent{})

	}

	return
}

//export globalOnEvent
func globalOnEvent(handle C.HWND, message C.UINT, wParam C.WPARAM, lParam C.LPARAM) C.LRESULT {
	// Associate handle and Window instance when the creation message is received
	if message == C.WM_CREATE {
		// Get WindowImplWin32 instance (it was passed as the last argument of CreateWindow)
		window := (*C.CREATESTRUCT)(unsafe.Pointer(uintptr(lParam))).lpCreateParams

		// Set as the "user data" parameter of the window
		C.__SetWindowLongPtr(handle, C.GWLP_USERDATA, unsafe.Pointer(window))
	}

	// Get the WindowImpl instance corresponding to the window handle
	// Forward the event to the appropriate function
	if wi := (*windowInternal)(C.__GetWindowLongPtr(handle, C.GWLP_USERDATA)); wi != nil {
		events, errors := wi.processEvent(message, wParam, lParam)
		wi.events = append(wi.events, events...)
		wi.eventErrors = append(wi.eventErrors, errors...)

		if callback := C.WNDPROC(wi.callback); callback != nil {
			return C.CallWindowProc(callback, handle, message, wParam, lParam)
		}
	}

	// We don't forward the WM_CLOSE message to prevent the OS from automatically destroying the window
	if message == C.WM_CLOSE {
		return 0
	}

	return C.DefWindowProcW(handle, message, wParam, lParam)
}
