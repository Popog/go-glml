// Copyright © 2012 Popog
package glml

// #include "helper_windows.h"
// extern BOOL CALLBACK (*pEnumMonitors)(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData);
// 
import "C"
import (
	"unsafe"
)

type monitorInternalFinder struct {
	i        int // counter of the current iteration
	monitors []C.HMONITOR
}

func findMonitors() *monitorInternalFinder {
	m := &monitorInternalFinder{}

	C.EnumDisplayMonitors(nil, nil, C.pEnumMonitors, C.LPARAM(uintptr(unsafe.Pointer(m))))

	return m
}

//export enumMonitors
func enumMonitors(hMonitor C.HMONITOR, hdcMonitor C.HDC, lprcMonitor C.LPRECT, dwData C.LPARAM) C.BOOL {
	mf := (*monitorInternalFinder)(unsafe.Pointer(uintptr(dwData)))
	mf.monitors = append(mf.monitors, hMonitor)
	return C.TRUE
}

// returns true if a monitor was found
func (mf *monitorInternalFinder) get(internal *monitorInternal) bool {
	if mf.i >= len(mf.monitors) {
		return false
	}

	internal.handle = mf.monitors[mf.i]
	info := (*C.MONITORINFO)(unsafe.Pointer(&internal.info))
	info.cbSize = C.MONITORINFOEXW_size

	mf.i++
	return internal.isValid()
}

type monitorInternal struct {
	handle C.HMONITOR
	info   C.MONITORINFOEXW
}

func (mi *monitorInternal) getDefaultMonitor() {
	mi.handle = C.MonitorFromRect(nil, C.MONITOR_DEFAULTTOPRIMARY)

	info := (*C.MONITORINFO)(unsafe.Pointer(&mi.info))
	info.cbSize = C.MONITORINFOEXW_size

	C.__GetMonitorInfoW(mi.handle, &mi.info)
}

func (mi *monitorInternal) isDefault() bool {
	info := (*C.MONITORINFO)(unsafe.Pointer(&mi.info))
	return info.dwFlags&C.MONITORINFOF_PRIMARY != 0
}

func (mi *monitorInternal) isValid() bool {
	if mi.handle == nil {
		return false
	}
	info := (*C.MONITORINFO)(unsafe.Pointer(&mi.info))
	info.cbSize = C.MONITORINFOEXW_size

	return C.__GetMonitorInfoW(mi.handle, &mi.info) != 0
}

// Get the list of all the supported fullscreen video modes
func (mi *monitorInternal) getFullscreenVideoModes() []VideoMode {
	// Enumerate all available video modes for the primary display adapter
	mode_set := make(map[VideoMode]bool)
	for win32Mode, count := (C.DEVMODEW{dmSize: C.DEVMODEW_size}), C.DWORD(0); C.__EnumDisplaySettingsW(&mi.info.szDevice[0], count, &win32Mode) != 0; count++ {
		vm := VideoMode{
			Width:        uint(win32Mode.dmPelsWidth),
			Height:       uint(win32Mode.dmPelsHeight),
			BitsPerPixel: uint(win32Mode.dmBitsPerPel),
		}
		mode_set[vm] = true
	}

	// add them all into the slice
	modes := make([]VideoMode, 0, len(mode_set))
	for mode, _ := range mode_set {
		modes = append(modes, mode)
	}
	return modes
}

// Returns whether or not a monitor supports a particular video mode
func (mi *monitorInternal) supportsMode(mode VideoMode) bool {
	devMode := C.DEVMODEW{
		dmSize:       C.DEVMODEW_size,
		dmPelsWidth:  C.DWORD(mode.Width),
		dmPelsHeight: C.DWORD(mode.Height),
		dmBitsPerPel: C.DWORD(mode.BitsPerPixel),
		dmFields:     C.DM_PELSWIDTH | C.DM_PELSHEIGHT | C.DM_BITSPERPEL,
	}

	err := ChangeDisplaySettingsExW(&mi.info.szDevice[0], &devMode, nil, C.CDS_TEST, nil)
	return err == nil
}

// Get the current desktop video mode
func (mi *monitorInternal) getDesktopMode() VideoMode {
	win32Mode := C.DEVMODEW{dmSize: C.DEVMODEW_size}
	C.__EnumDisplaySettingsW(&mi.info.szDevice[0], C.ENUM_CURRENT_SETTINGS, &win32Mode)
	return VideoMode{
		Width:        uint(win32Mode.dmPelsWidth),
		Height:       uint(win32Mode.dmPelsHeight),
		BitsPerPixel: uint(win32Mode.dmBitsPerPel),
	}
}
