// Copyright © 2012 Popog
package glml

// #include "helper_windows.h"
// 
// void * __HWND_BOTTOM    = HWND_BOTTOM;
// void * __HWND_NOTOPMOST = HWND_NOTOPMOST;
// void * __HWND_TOP       = HWND_TOP;
// void * __HWND_TOPMOST   = HWND_TOPMOST;
import "C"
import (
	"errors"
	"fmt"
	"unicode/utf16"
)

var (
	HWND_BOTTOM    = C.HWND(C.__HWND_BOTTOM)
	HWND_NOTOPMOST = C.HWND(C.__HWND_NOTOPMOST)
	HWND_TOP       = C.HWND(C.__HWND_TOP)
	HWND_TOPMOST   = C.HWND(C.__HWND_TOPMOST)
)

func utf16Convert(s string) (ptr C.LPCWSTR, length int) {
	encoded := utf16.Encode([]rune(s + "\x00"))
	return C.LPCWSTR((*C.WCHAR)(&encoded[0])), len(encoded)
}

func utf16ConvertFrom(s []C.WCHAR) string {
	ret := make([]uint16, len(s))
	for i, v := range s {
		ret[i] = uint16(v)
	}
	return string(utf16.Decode(ret))
}

func ChangeDisplaySettingsExW(lpszDeviceName C.LPCWSTR, lpDevMode *C.DEVMODEW, hwnd C.HWND, dwflags C.DWORD, lParam C.LPVOID) error {
	result := C.__ChangeDisplaySettingsExW(lpszDeviceName, lpDevMode, hwnd, dwflags, lParam)

	switch result {
	case C.DISP_CHANGE_SUCCESSFUL:
		return nil
	case C.DISP_CHANGE_BADDUALVIEW:
		return errors.New("DISP_CHANGE_BADDUALVIEW")
	case C.DISP_CHANGE_BADFLAGS:
		return errors.New("DISP_CHANGE_BADFLAGS")
	case C.DISP_CHANGE_BADMODE:
		return errors.New("DISP_CHANGE_BADMODE")
	case C.DISP_CHANGE_BADPARAM:
		return errors.New("DISP_CHANGE_BADPARAM")
	case C.DISP_CHANGE_FAILED:
		return errors.New("DISP_CHANGE_FAILED")
	case C.DISP_CHANGE_NOTUPDATED:
		return errors.New("DISP_CHANGE_NOTUPDATED")
	case C.DISP_CHANGE_RESTART:
		return errors.New("DISP_CHANGE_RESTART")
	}

	return fmt.Errorf("ChangeDisplaySettingsExW failed (%d)", result)
}
