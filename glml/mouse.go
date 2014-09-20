// Copyright © 2012 Popog
package glml

import "errors"

// Mouse buttons
type MouseButton uint

const (
	// Concrete buttons
	MouseLeftRH   MouseButton = iota // The left mouse button on a right-handed mouse (primary)
	MouseRightRH                     // The right mouse button on a right-handed mouse (secondary)
	MouseLeftLH                      // The left mouse button on a left-handed mouse (secondary)
	MouseRightLH                     // The right mouse button on a left-handed mouse (primary)
	MouseMiddle                      // The middle (wheel) mouse button
	MouseXButton1                    // The first extra mouse button
	MouseXButton2                    // The second extra mouse button

	// Abstract buttons. Will NOT be returned by API (e.g. Events)
	MouseButtonPrimary   // The primary mouse button regardless of handedness
	MouseButtonSecondary // The secondary mouse button regardless of handedness
	MouseButtonLeft      // The left mouse button regardless of handedness
	MouseButtonRight     // The right mouse button regardless of handedness

	MouseButtonCount // Keep last -- the total number of mouse buttons
)

// Check if a mouse button is pressed
func IsMouseButtonPressed(button MouseButton) bool {
	if button >= MouseButtonCount {
		panic(errors.New("button out of range"))
	}

	switch button {
	case MouseLeftRH, MouseRightRH:
		if mouseIsLeft() {
			return false
		}
	case MouseLeftLH, MouseRightLH:
		if !mouseIsLeft() {
			return false
		}
	case MouseButtonPrimary:
		if mouseIsLeft() {
			button = MouseButtonRight
		} else {
			button = MouseButtonLeft
		}
	case MouseButtonSecondary:
		if mouseIsLeft() {
			button = MouseButtonLeft
		} else {
			button = MouseButtonRight
		}
	}

	return isMouseButtonPressed(button)
}

// Get the current position of the mouse in desktop coordinates
func GetMousePosition() (x, y int) {
	return getMousePosition()
}

// Set the current position of the mouse in desktop coordinates
func SetMousePosition(x, y int) {
	setMousePosition(x, y)
}
