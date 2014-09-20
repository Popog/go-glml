#pragma once

#define WIN32_LEAN_AND_MEAN 1
#include <windows.h>

#define PIXELFORMATDESCRIPTOR_size sizeof(PIXELFORMATDESCRIPTOR)
#define DEVMODEW_size              sizeof(DEVMODEW)
#define TRACKMOUSEEVENT_size       sizeof(TRACKMOUSEEVENT)
#define MONITORINFOEXW_size        sizeof(MONITORINFOEXW)

extern LPCWSTR __IDC_ARROW;

BOOL __AdjustWindowRect(RECT *lpRect, DWORD dwStyle, BOOL bMenu);
void * __SetWindowLongPtr(HWND hWnd, int nIndex, void *dwNewLong);
void * __GetWindowLongPtr(HWND hWnd, int nIndex);
BOOL __PeekMessage(MSG *lpMsg, HWND hWnd, UINT wMsgFilterMin, UINT wMsgFilterMax, UINT wRemoveMsg);
BOOL __GetWindowRect(HWND hWnd, RECT *lpRect);
int __DescribePixelFormat(HDC hdc, int iPixelFormat, UINT nBytes, PIXELFORMATDESCRIPTOR *ppfd);
BOOL __EnumDisplaySettingsW(LPCWSTR lpszDeviceName, DWORD iModeNum, DEVMODEW *lpDevMode);
LONG __ChangeDisplaySettingsExW(LPCWSTR lpszDeviceName, DEVMODEW *lpDevMode, HWND hwnd, DWORD dwflags, LPVOID lParam);
//BOOL __EnumDisplayDevicesW(LPCWSTR lpDevice, DWORD iDevNum, DISPLAY_DEVICE *lpDisplayDevice, DWORD dwFlags);
BOOL __GetMonitorInfoW(HMONITOR hMonitor, MONITORINFOEXW *lpmi);
BOOL __ScreenToClient(HWND hWnd, POINT *lpPoint);
BOOL __GetCursorPos(POINT *lpPoint);
BOOL __TrackMouseEvent(TRACKMOUSEEVENT *lpEventTrack);
WORD __HIWORD(DWORD dwValue);
WORD __LOWORD(DWORD dwValue);

ULONG SetWindowULong(HWND hWnd, int nIndex, ULONG dwNewLong);