// Copyright © 2012 Popog
package glml

import "errors"

// Enumeration of the window styles
type WindowStyle int

const (
	WindowStyleNone       WindowStyle = 0         // No border / title bar
	WindowStyleTitlebar   WindowStyle = 1 << iota // Title bar + border
	WindowStyleResize                             // Resizable border + maximize button (requires StyleTitlebar)
	WindowStyleClose                              // Close button (requires StyleTitlebar)
	WindowStyleFullscreen                         // Fullscreen mode (this flag and all others are mutually exclusive)

	WindowStyleDefault = WindowStyleTitlebar | WindowStyleResize | WindowStyleClose // Default window style
)

// return nil if there is no error
func (ws WindowStyle) Check() error {
	// StyleFullscreen is mutual exclusive with all other flags
	if ws&WindowStyleFullscreen != 0 && ws != WindowStyleFullscreen {
		return errors.New("WindowStyleFullscreen is mutual exclusive with all other flags")
	}

	// StyleResize and StyleClose require StyleTitlebar
	if ws&WindowStyleResize != 0 && ws&WindowStyleTitlebar == 0 {
		return errors.New("WindowStyleResize require WindowStyleTitlebar")
	}
	if ws&WindowStyleClose != 0 && ws&WindowStyleTitlebar == 0 {
		return errors.New("WindowStyleClose require WindowStyleTitlebar")
	}

	return nil
}
