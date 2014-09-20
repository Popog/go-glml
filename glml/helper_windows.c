#include "helper_windows.h"

LPCWSTR __IDC_ARROW = MAKEINTRESOURCEW(32512);//IDC_ARROW

BOOL __AdjustWindowRect(RECT *lpRect, DWORD dwStyle, BOOL bMenu)
{ return AdjustWindowRect(lpRect, dwStyle, bMenu); }

void * __SetWindowLongPtr(HWND hWnd, int nIndex, void *dwNewLong)
{ return (void *)SetWindowLongPtr(hWnd, nIndex, (LONG_PTR)dwNewLong); }

void * __GetWindowLongPtr(HWND hWnd, int nIndex)
{ return (void *)GetWindowLongPtr(hWnd, nIndex); }

BOOL __PeekMessage(MSG *lpMsg, HWND hWnd, UINT wMsgFilterMin, UINT wMsgFilterMax, UINT wRemoveMsg)
{ return PeekMessage(lpMsg, hWnd, wMsgFilterMin, wMsgFilterMax, wRemoveMsg); }

BOOL __GetWindowRect(HWND hWnd, RECT *lpRect)
{ return GetWindowRect(hWnd, lpRect); }

int __DescribePixelFormat(HDC hdc, int iPixelFormat, UINT nBytes, PIXELFORMATDESCRIPTOR *ppfd)
{ return DescribePixelFormat(hdc, iPixelFormat, nBytes, ppfd); }

BOOL __EnumDisplaySettingsW(LPCWSTR lpszDeviceName, DWORD iModeNum, DEVMODEW *lpDevMode)
{ return EnumDisplaySettingsW(lpszDeviceName, iModeNum, lpDevMode); }

LONG __ChangeDisplaySettingsExW(LPCWSTR lpszDeviceName, DEVMODEW *lpDevMode, HWND hwnd, DWORD dwflags, LPVOID lParam)
{ return ChangeDisplaySettingsExW(lpszDeviceName, lpDevMode, hwnd, dwflags, lParam); }

//BOOL __EnumDisplayDevicesW(LPCWSTR lpDevice, DWORD iDevNum, DISPLAY_DEVICE *lpDisplayDevice, DWORD dwFlags)
//{ return EnumDisplayDevicesW(lpDevice, iDevNum, lpDisplayDevice, dwFlags); }

BOOL __GetMonitorInfoW(HMONITOR hMonitor, MONITORINFOEXW *lpmi)
{ return GetMonitorInfoW(hMonitor, (MONITORINFO *)lpmi); }

BOOL __ScreenToClient(HWND hWnd, POINT *lpPoint)
{ return ScreenToClient(hWnd, lpPoint); }

BOOL __GetCursorPos(POINT *lpPoint)
{ return GetCursorPos(lpPoint); }

BOOL __TrackMouseEvent(TRACKMOUSEEVENT *lpEventTrack)
{ return TrackMouseEvent(lpEventTrack); }

WORD __HIWORD(DWORD dwValue)
{ return HIWORD(dwValue); }

WORD __LOWORD(DWORD dwValue)
{ return LOWORD(dwValue); }

ULONG SetWindowULong(HWND hWnd, int nIndex, ULONG dwNewLong)
{ return (ULONG) SetWindowLong(hWnd, nIndex, (LONG)dwNewLong); }
