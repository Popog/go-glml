// Copyright © 2012 Popog
package glml

// #include "helper_windows.h"
import "C"

var mouse_vkeys = [MouseButtonCount]C.int{
	MouseLeftRH:   C.VK_LBUTTON,
	MouseRightRH:  C.VK_RBUTTON,
	MouseLeftLH:   C.VK_LBUTTON,
	MouseRightLH:  C.VK_RBUTTON,
	MouseMiddle:   C.VK_MBUTTON,
	MouseXButton1: C.VK_XBUTTON1,
	MouseXButton2: C.VK_XBUTTON2,

	MouseButtonLeft:  C.VK_LBUTTON,
	MouseButtonRight: C.VK_RBUTTON,
}

type mouseKey struct {
	message     C.UINT
	left_handed bool
}

var mouse_vkeys_handed_map = map[mouseKey]MouseButton{
	mouseKey{C.WM_LBUTTONDOWN, false}: MouseLeftRH,
	mouseKey{C.WM_LBUTTONDOWN, true}:  MouseLeftLH,
	mouseKey{C.WM_RBUTTONDOWN, false}: MouseRightRH,
	mouseKey{C.WM_RBUTTONDOWN, true}:  MouseRightLH,

	mouseKey{C.WM_LBUTTONUP, false}: MouseLeftRH,
	mouseKey{C.WM_LBUTTONUP, true}:  MouseLeftLH,
	mouseKey{C.WM_RBUTTONUP, false}: MouseRightRH,
	mouseKey{C.WM_RBUTTONUP, true}:  MouseRightLH,
}

func mouseIsLeft() bool {
	return C.GetSystemMetrics(C.SM_SWAPBUTTON) == C.TRUE
}

// Check if a mouse button is pressed
func isMouseButtonPressed(button MouseButton) bool {
	return uint16(C.GetAsyncKeyState(mouse_vkeys[button]))&0x8000 != 0
}

// Get the current position of the mouse in desktop coordinates
func getMousePosition() (x, y int) {
	x, y = -1, -1
	var point C.POINT
	if C.__GetCursorPos(&point) == 0 {
		C.GetLastError()
		return
	}
	return int(point.x), int(point.y)
}

// Set the current position of the mouse in desktop coordinates
func setMousePosition(x, y int) {
	if C.SetCursorPos(C.int(x), C.int(y)) == 0 {
		// C.GetLastError()
	}
}
