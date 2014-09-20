// Copyright © 2012 Popog

#define WIN32_LEAN_AND_MEAN 1
#include <windows.h>

#include "_cgo_export.h"

BOOL CALLBACK __EnumMonitors(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData)
{
	return enumMonitors(hMonitor, hdcMonitor, lprcMonitor, dwData);
}

BOOL CALLBACK (*pEnumMonitors)(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData) = &__EnumMonitors;
