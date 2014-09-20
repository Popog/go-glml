// Copyright © 2012 Popog

#define WIN32_LEAN_AND_MEAN 1
#include <windows.h>

#include "_cgo_export.h"

LRESULT CALLBACK __globalOnEvent(HWND handle, UINT message, WPARAM wParam, LPARAM lParam)
{
	return globalOnEvent(handle, message, wParam, lParam);
}
LRESULT CALLBACK (*pGlobalOnEvent)(HWND handle, UINT message, WPARAM wParam, LPARAM lParam) = &__globalOnEvent;
