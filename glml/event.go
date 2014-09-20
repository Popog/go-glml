// Copyright © 2012 Popog
package glml

type Event interface {
}

// db   d8b   db d888888b d8b   db d8888b.  .d88b.  db   d8b   db 
// 88   I8I   88   `88'   888o  88 88  `8D .8P  Y8. 88   I8I   88 
// 88   I8I   88    88    88V8o 88 88   88 88    88 88   I8I   88 
// Y8   I8I   88    88    88 V8o88 88   88 88    88 Y8   I8I   88 
// `8b d8'8b d8'   .88.   88  V888 88  .8D `8b  d8' `8b d8'8b d8' 
//  `8b8' `8d8'  Y888888P VP   V8P Y8888D'  `Y88P'   `8b8' `8d8'  

// The window requested to be closed
type WindowClosedEvent struct {
}

// The window was resized
type WindowResizeEvent struct {
	Width, Height uint // New width and height, in pixels
}

// The window lost the focus
type WindowLostFocusEvent struct {
}

// The window gained the focus
type WindowGainedFocusEvent struct {
}

// db   dD d88888b db    db d8888b.  .d88b.   .d8b.  d8888b. d8888b. 
// 88 ,8P' 88'     `8b  d8' 88  `8D .8P  Y8. d8' `8b 88  `8D 88  `8D 
// 88,8P   88ooooo  `8bd8'  88oooY' 88    88 88ooo88 88oobY' 88   88 
// 88`8b   88~~~~~    88    88~~~b. 88    88 88~~~88 88`8b   88   88 
// 88 `88. 88.        88    88   8D `8b  d8' 88   88 88 `88. 88  .8D 
// YP   YD Y88888P    YP    Y8888P'  `Y88P'  YP   YP 88   YD Y8888D' 

// A character was entered
type TextEnteredEvent struct {
	Character rune // character
}

// A key was pressed
type KeyPressedEvent struct {
	Code                        Key  // Code of the key that has been pressed
	Alt, Control, Shift, System bool // Is a modifier key pressed?
}

// A key was released
type KeyReleasedEvent struct {
	Code                        Key  // Code of the key that has been released
	Alt, Control, Shift, System bool // Is a modifier key pressed?
}

// .88b  d88.  .d88b.  db    db .d8888. d88888b 
// 88'YbdP`88 .8P  Y8. 88    88 88'  YP 88'     
// 88  88  88 88    88 88    88 `8bo.   88ooooo 
// 88  88  88 88    88 88    88   `Y8b. 88~~~~~ 
// 88  88  88 `8b  d8' 88b  d88 db   8D 88.     
// YP  YP  YP  `Y88P'  ~Y8888P' `8888Y' Y88888P 

// The mouse cursor moved
type MouseMoveEvent struct {
	X, Y int // X and Y positions of the mouse pointer, relative to the top-left of the owner window
}

// A mouse button was pressed
type MouseButtonPressedEvent struct {
	Button MouseButton // Code of the concrete button that has been pressed
	X, Y   int         // X and Y position of the mouse pointer, relative to the top-left of the owner window
}

// A mouse button was released
type MouseButtonReleasedEvent struct {
	Button MouseButton // Code of the concrete button that has been released
	X, Y   int         // X and Y positions of the mouse pointer, relative to the top-left of the owner window
}

// The mouse wheel was scrolled
type MouseWheelEvent struct {
	Delta int // Number of ticks the wheel has moved (positive is up, negative is down)
	X, Y  int // X and Y position of the mouse pointer, relative to the top-left of the owner window
}

// The mouse cursor entered the area of the window
type MouseEnteredEvent struct {
}

// The mouse cursor left the area of the window
type MouseLeftEvent struct {
}
