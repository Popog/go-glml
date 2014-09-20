// Copyright © 2012 Popog
package glml

// Monitors may expire randomly
type Monitor struct {
	internal monitorInternal
}

// Get the default monitor
func GetDefaultMonitor() *Monitor {
	m := &Monitor{}
	m.internal.getDefaultMonitor()
	return m
}

// Get all the monitors on the system
func GetMonitors() []*Monitor {
	var monitors []*Monitor

	for mf, m := findMonitors(), new(Monitor); mf.get(&m.internal); m = new(Monitor) {
		monitors = append(monitors, m)
	}
	return monitors
}

// Returns true if this monitor is the default monitor
func (m *Monitor) IsDefault() bool {
	return m.internal.isDefault()
}

// Returns true if the monitor is still a valid monitor
func (m *Monitor) IsValid() bool {
	return m.internal.isValid()
}

// Get the current desktop video mode
func (m *Monitor) GetDesktopMode() VideoMode {
	return m.internal.getDesktopMode()
}

// Get the list of all the supported fullscreen video modes
func (m *Monitor) GetFullscreenVideoModes() []VideoMode {
	return m.internal.getFullscreenVideoModes()
}

// Returns whether or not a monitor supports a particular video mode
func (m *Monitor) SupportsMode(mode VideoMode) bool {
	return m.internal.supportsMode(mode)
}

type VideoMode struct {
	Width, Height uint // Video mode width and height, in pixels
	BitsPerPixel  uint // Video mode pixel depth, in bits per pixels
}

// Compare two video modes. Pixel depth is considered more significant
// than dimensions
func (lhs VideoMode) Less(rhs VideoMode) bool {
	if lhs.BitsPerPixel != rhs.BitsPerPixel {
		return lhs.BitsPerPixel < rhs.BitsPerPixel
	}

	return lhs.Width*lhs.Height < rhs.Width*rhs.Height
}
