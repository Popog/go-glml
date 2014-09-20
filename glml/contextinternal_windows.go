// Copyright © 2012 Popog
package glml

// #cgo windows LDFLAGS: -lopengl32 -lgdi32
// #include "helper_windows.h"
// #include <GL/gl.h>
// #include <GL/glu.h>
// #include "wglext.h"
//
// #define PROCLIST                                                    \
// PROC(PFNWGLSWAPINTERVALEXTPROC,         wglSwapIntervalEXT)         \
// PROC(PFNWGLCHOOSEPIXELFORMATARBPROC,    wglChoosePixelFormatARB)    \
// PROC(PFNWGLCREATECONTEXTATTRIBSARBPROC, wglCreateContextAttribsARB) \
//
// #define PROC(type, name)  \
// 	type p_##name;           \
// 	DWORD error_##name;     //
// typedef struct
// {
// PROCLIST
// } wglProcs;
// #undef PROC
//
// #define PROC(type, name)                                                \
// 	procs->p_##name     = (type)wglGetProcAddress(#name);                  \
// 	procs->error_##name = (procs->p_##name == NULL) ? GetLastError() : 0; //
// void wglLoadProcs(wglProcs * procs)
// {
// 	PROCLIST
// }
// #undef PROC
// 
// BOOL __wglSwapIntervalEXT(wglProcs const * procs, int interval)
// { return procs->p_wglSwapIntervalEXT(interval); }
// 
// BOOL __wglChoosePixelFormatARB(wglProcs const * procs, HDC hdc, const int *piAttribIList, const FLOAT *pfAttribFList, UINT nMaxFormats, int *piFormats, UINT *nNumFormats)
// { return procs->p_wglChoosePixelFormatARB(hdc, piAttribIList, pfAttribFList, nMaxFormats, piFormats, nNumFormats); }
// 
// HGLRC __wglCreateContextAttribsARB(wglProcs const * procs, HDC hDC, HGLRC hShareContext, const int *attribList)
// { return procs->p_wglCreateContextAttribsARB(hDC, hShareContext, attribList); }
//
import "C"
import (
	"errors"
	"fmt"
)

var contextInternal_className, _ = utf16Convert("STATIC")
var contextInternal_wTitle, _ = utf16Convert("")

func createHiddenWindow(width, height int) C.HWND {
	window := C.CreateWindowExW(0, contextInternal_className, contextInternal_wTitle, C.WS_POPUP|C.WS_DISABLED, 0, 0, C.int(width), C.int(height), nil, nil, C.GetModuleHandle(nil), nil)
	if window == nil {
		return nil
	}

	C.ShowWindow(window, C.SW_HIDE)
	return window
}

// Let's find a suitable pixel format -- first try with antialiasing
func BestwglChoosePixelFormatARB(procs *C.wglProcs, hdc C.HDC, bitsPerPixel uint, settings *ContextSettings) C.int {
	if settings.AntialiasingLevel <= 0 {
		return 0
	}
	if procs.p_wglChoosePixelFormatARB == nil {
		return 0
	}

	AttribIList := [...]C.int{
		C.WGL_DRAW_TO_WINDOW_ARB, C.GL_TRUE,
		C.WGL_SUPPORT_OPENGL_ARB, C.GL_TRUE,
		C.WGL_ACCELERATION_ARB, C.WGL_FULL_ACCELERATION_ARB,
		C.WGL_DOUBLE_BUFFER_ARB, C.GL_TRUE,
		C.WGL_SAMPLE_BUFFERS_ARB, C.GL_TRUE, // turn on antialiasing
		C.WGL_SAMPLES_ARB, C.int(settings.AntialiasingLevel),
		0, 0,
	}
	AttribFList := []C.FLOAT{0, 0}

	// Let's check how many formats are supporting our requirements
	const formats_size = 128
	var formats [formats_size]C.int
	var nbFormats C.UINT
	for settings.AntialiasingLevel > 0 {
		if C.__wglChoosePixelFormatARB(procs, hdc, &AttribIList[0], &AttribFList[0], formats_size, &formats[0], &nbFormats) == C.TRUE &&
			nbFormats > 0 {
			break
		}
		nbFormats = 0 // reset this

		// Decrease the antialiasing level until we find a valid one
		settings.AntialiasingLevel--
		AttribIList[11]--
	}

	bestScore := uint(1<<32 - 1)
	var bestFormat C.int
	for i := C.UINT(0); bestScore != 0 && i < nbFormats; i++ {
		// Get the current format's attributes
		attributes := C.PIXELFORMATDESCRIPTOR{
			nSize:    C.PIXELFORMATDESCRIPTOR_size,
			nVersion: 1,
		}
		if C.__DescribePixelFormat(hdc, formats[i], C.PIXELFORMATDESCRIPTOR_size, &attributes) == 0 {
			return 0
		}

		// Evaluate the current configuration
		color := uint(attributes.cRedBits + attributes.cGreenBits + attributes.cBlueBits + attributes.cAlphaBits)
		score := EvaluateFormat(bitsPerPixel, *settings, color, uint(attributes.cDepthBits), uint(attributes.cStencilBits), settings.AntialiasingLevel)

		// Keep it if it's better than the current best
		if score < bestScore {
			bestScore = score
			bestFormat = formats[i]
		}
	}

	return bestFormat
}

// Find a pixel format with no antialiasing, if not needed or not supported
func BestChoosePixelFormat(hdc C.HDC, bitsPerPixel uint, settings *ContextSettings) C.int {
	// Setup a pixel format descriptor from the rendering settings
	descriptor := C.PIXELFORMATDESCRIPTOR{
		nSize:        C.PIXELFORMATDESCRIPTOR_size,
		nVersion:     1,
		iLayerType:   C.PFD_MAIN_PLANE,
		dwFlags:      C.PFD_DRAW_TO_WINDOW | C.PFD_SUPPORT_OPENGL | C.PFD_DOUBLEBUFFER,
		iPixelType:   C.PFD_TYPE_RGBA,
		cColorBits:   C.BYTE(bitsPerPixel),
		cDepthBits:   C.BYTE(settings.DepthBits),
		cStencilBits: C.BYTE(settings.StencilBits),
	}
	if bitsPerPixel == 32 {
		descriptor.cAlphaBits = 8
	}

	// Get the pixel format that best matches our requirements
	return C.ChoosePixelFormat(hdc, &descriptor)
}

func createContext(procs *C.wglProcs, sharedContext C.HGLRC, hdc C.HDC, bitsPerPixel uint, settings *ContextSettings) C.HGLRC {
	bestFormat := BestwglChoosePixelFormatARB(procs, hdc, bitsPerPixel, settings)
	if bestFormat == 0 {
		bestFormat = BestChoosePixelFormat(hdc, bitsPerPixel, settings)
	}
	if bestFormat == 0 {
		return nil // Failed to find a suitable pixel format for device context -- cannot create OpenGL context
	}

	// Extract the depth and stencil bits from the chosen format
	actualFormat := C.PIXELFORMATDESCRIPTOR{
		nSize:    C.PIXELFORMATDESCRIPTOR_size,
		nVersion: 1,
	}
	if C.__DescribePixelFormat(hdc, bestFormat, C.PIXELFORMATDESCRIPTOR_size, &actualFormat) == 0 {
		return nil
	}
	settings.DepthBits = uint(actualFormat.cDepthBits)
	settings.StencilBits = uint(actualFormat.cStencilBits)

	// Set the chosen pixel format
	if C.SetPixelFormat(hdc, bestFormat, &actualFormat) == C.FALSE {
		return nil // Failed to set pixel format for device context -- cannot create OpenGL context
	}

	if procs.p_wglCreateContextAttribsARB != nil {
		for settings.MajorVersion >= 3 {
			attributes := [...]C.int{
				C.WGL_CONTEXT_MAJOR_VERSION_ARB, C.int(settings.MajorVersion),
				C.WGL_CONTEXT_MINOR_VERSION_ARB, C.int(settings.MinorVersion),
				C.WGL_CONTEXT_PROFILE_MASK_ARB, C.WGL_CONTEXT_COMPATIBILITY_PROFILE_BIT_ARB,
				0, 0,
			}

			if context := C.__wglCreateContextAttribsARB(procs, hdc, sharedContext, &attributes[0]); context != nil {
				return context
			}

			// If we couldn't create the context, lower the version number and try again -- stop at 3.0
			// Invalid version numbers will be generated by this algorithm (like 3.9), but we really don't care
			if settings.MinorVersion > 0 {
				// If the minor version is not 0, we decrease it and try again
				settings.MinorVersion--
			} else {
				// If the minor version is 0, we decrease the major version
				settings.MajorVersion--
				settings.MinorVersion = 9
			}
		}
	}

	// set the context version to 2.0 (arbitrary)
	settings.MajorVersion = 2
	settings.MinorVersion = 0

	context := C.wglCreateContext(hdc)
	if context == nil {
		return nil // Failed to create an OpenGL context for this window
	}

	// Share this context with others
	C.wglShareLists(sharedContext, context)
	return context
}

type contextInternal struct {
	deactivateSignal chan bool
	procs            C.wglProcs      // The function pointers for this context
	window           C.HWND          // Window to which the context is attached
	ownsWindow       bool            // Do we own the target window?
	hdc              C.HDC           // Device context associated to the context
	isSharedContext  bool            // whether or not this is the shared context
	context          C.HGLRC         // OpenGL context
	settings         ContextSettings // The settings for the context
}

func (ic *contextInternal) initialize() ThreadError {
	return ic.initializeFromSettings(ContextSettingsDefault, 1, 1)
}

func (ic *contextInternal) initializeFromOwner(settings ContextSettings, owner *windowInternal, bitsPerPixel uint) ThreadError {
	ic.deactivateSignal = make(chan bool)
	ic.settings = settings

	// Get the owner window and its device context
	ic.window = C.HWND(owner.window.Handle)
	ic.ownsWindow = false

	// get the device context
	ic.hdc = C.GetDC(ic.window)
	if ic.hdc == nil {
		return NewThreadError(errors.New("no device context"), true)
	}

	ic.context = createContext(&sharedContext.internal.procs, sharedContext.internal.context, ic.hdc, bitsPerPixel, &ic.settings)
	if ic.context == nil {
		return NewThreadError(fmt.Errorf("could not create context (%d)", C.GetLastError()), true)
	}

	// signal because we start out deactivated
	ic.signalDeactivation()
	return nil
}

func (ic *contextInternal) initializeFromSettings(settings ContextSettings, width, height int) ThreadError {
	ic.deactivateSignal = make(chan bool)
	ic.settings = settings

	ic.window = createHiddenWindow(width, height)
	if ic.window == nil {
		// C.GetLastError()
		return NewThreadError(fmt.Errorf("could not create window (%d)", C.GetLastError()), true)
	}
	ic.ownsWindow = true

	ic.hdc = C.GetDC(ic.window)
	if ic.hdc == nil {
		return NewThreadError(errors.New("no device context"), true)
	}

	if sharedContext == nil { // if there is no shared context, we are it
		ic.isSharedContext = true

		pfd := C.PIXELFORMATDESCRIPTOR{
			nSize:      C.PIXELFORMATDESCRIPTOR_size, // size of this pfd
			nVersion:   1,                            // version number
			iPixelType: C.PFD_TYPE_RGBA,              // RGBA type
			cColorBits: 24,                           // 24-bit color depth
			cDepthBits: 32,                           // 32-bit z-buffer
			iLayerType: C.PFD_MAIN_PLANE,             // main layer

			// support window | OpenGL | double buffer
			dwFlags: C.PFD_DRAW_TO_WINDOW | C.PFD_SUPPORT_OPENGL | C.PFD_DOUBLEBUFFER,
		}

		// get the best available match of pixel format for the device context
		// make that the pixel format of the device context  
		if iPixelFormat := C.ChoosePixelFormat(ic.hdc, &pfd); iPixelFormat == 0 {
			return NewThreadError(fmt.Errorf("sharedContext: ChoosePixelFormat failed (%d)", C.GetLastError()), true)
		} else if C.SetPixelFormat(ic.hdc, iPixelFormat, &pfd) == C.FALSE {
			return NewThreadError(fmt.Errorf("sharedContext: SetPixelFormat failed (%d)", C.GetLastError()), true)
		}

		ic.context = C.wglCreateContext(ic.hdc)
		if ic.context == nil {
			return NewThreadError(fmt.Errorf("sharedContext: wglCreateContext failed (%d)", C.GetLastError()), true)
		}
	} else { // otherwise we push the commands onto the shared context thread

		bitsPerPixel := GetDefaultMonitor().GetDesktopMode().BitsPerPixel
		ic.context = createContext(&sharedContext.internal.procs, sharedContext.internal.context, ic.hdc, bitsPerPixel, &ic.settings)
		if ic.context == nil {
			return NewThreadError(fmt.Errorf("could not create context (%d)", C.GetLastError()), true)
		}
	}
	// signal because we start out deactivated
	ic.signalDeactivation()
	return nil
}

func (ic *contextInternal) getSettings() (ContextSettings, ThreadError) {
	return ic.settings, nil
}

func (ic *contextInternal) setVerticalSyncEnabled(enabled bool) ThreadError {
	var interval C.int
	if enabled {
		interval = 1
	}

	if ic.procs.p_wglSwapIntervalEXT == nil {
		return NewThreadError(fmt.Errorf("wglSwapIntervalEXT == nil (%d)", ic.procs.error_wglSwapIntervalEXT), false)
	}

	if C.__wglSwapIntervalEXT(&ic.procs, interval) == C.FALSE {
		return NewThreadError(fmt.Errorf("wglSwapIntervalEXT failed (%d)", C.GetLastError()), false)
	}
	return nil
}

// Display what has been rendered to the context so far
func (ic *contextInternal) swapBuffers() ThreadError {
	if C.SwapBuffers(ic.hdc) == C.FALSE {
		return NewThreadError(fmt.Errorf("SwapBuffers (%d)", C.GetLastError()), true)
	}
	return nil
}

// Activate the context as the current target for rendering
func (ic *contextInternal) activate() ThreadError {
	// start by waiting for deactivation to finish
	<-ic.deactivateSignal

	// start up the context
	if C.wglMakeCurrent(ic.hdc, ic.context) == C.FALSE {
		return NewThreadError(fmt.Errorf("wglMakeCurrent failed (%d)", C.GetLastError()), true)
	}

	// Load all the functions and such
	C.wglLoadProcs(&ic.procs)

	return nil
}

// Deactivate the context as the current target for rendering
func (ic *contextInternal) deactivate() ThreadError {
	// disable the current context
	if C.wglMakeCurrent(ic.hdc, nil) == C.FALSE {
		return NewThreadError(fmt.Errorf("wglMakeCurrent failed (%d)", C.GetLastError()), true)
	}

	// end by signaling
	ic.signalDeactivation()
	return nil
}

// Deactivate, signal, wait for response, activate
func (ic *contextInternal) pause(signal chan bool) ThreadError {
	// disable the current context
	if C.wglMakeCurrent(ic.hdc, nil) == C.FALSE {
		return NewThreadError(fmt.Errorf("wglMakeCurrent failed (%d)", C.GetLastError()), true)
	}

	// let the other thread know and then wait for them
	signal <- true
	<-signal

	// start up the context
	if C.wglMakeCurrent(ic.hdc, ic.context) == C.FALSE {
		return NewThreadError(fmt.Errorf("wglMakeCurrent failed (%d)", C.GetLastError()), true)
	}
	return nil
}

// Temporary activate the context
func (ic *contextInternal) take() ThreadError {
	// start up the context
	if C.wglMakeCurrent(ic.hdc, ic.context) == C.FALSE {
		return NewThreadError(fmt.Errorf("wglMakeCurrent failed (%d)", C.GetLastError()), true)
	}
	return nil
}

// Temporary deactivate the context
func (ic *contextInternal) release() ThreadError {
	// disable the current context
	if C.wglMakeCurrent(ic.hdc, nil) == C.FALSE {
		return NewThreadError(fmt.Errorf("wglMakeCurrent failed (%d)", C.GetLastError()), true)
	}
	return nil
}

func (ic *contextInternal) signalDeactivation() {
	go func() { ic.deactivateSignal <- true }()
}

func (ic *contextInternal) close() ThreadError {
	// start by waiting for deactivation to finish
	<-ic.deactivateSignal

	// Destroy the OpenGL context
	if ic.context != nil {
		C.wglDeleteContext(ic.context)
	}

	// Destroy the device context
	if ic.hdc != nil {
		C.ReleaseDC(ic.window, ic.hdc)
	}

	// Destroy the window if we own it
	if ic.window != nil && ic.ownsWindow {
		C.DestroyWindow(ic.window)
	}

	return nil
}
