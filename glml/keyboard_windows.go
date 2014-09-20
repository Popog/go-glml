// Copyright © 2012 Popog
package glml

// #include "helper_windows.h"
import "C"

var keyboard_vkeys_map = make(map[C.int]Key)
var keyboard_vkeys = [KeyCount]C.int{
	KeyA: 'A',
	KeyB: 'B',
	KeyC: 'C',
	KeyD: 'D',
	KeyE: 'E',
	KeyF: 'F',
	KeyG: 'G',
	KeyH: 'H',
	KeyI: 'I',
	KeyJ: 'J',
	KeyK: 'K',
	KeyL: 'L',
	KeyM: 'M',
	KeyN: 'N',
	KeyO: 'O',
	KeyP: 'P',
	KeyQ: 'Q',
	KeyR: 'R',
	KeyS: 'S',
	KeyT: 'T',
	KeyU: 'U',
	KeyV: 'V',
	KeyW: 'W',
	KeyX: 'X',
	KeyY: 'Y',
	KeyZ: 'Z',

	KeyNum0: '0',
	KeyNum1: '1',
	KeyNum2: '2',
	KeyNum3: '3',
	KeyNum4: '4',
	KeyNum5: '5',
	KeyNum6: '6',
	KeyNum7: '7',
	KeyNum8: '8',
	KeyNum9: '9',

	KeyEscape:    C.VK_ESCAPE,
	KeyLControl:  C.VK_LCONTROL,
	KeyLShift:    C.VK_LSHIFT,
	KeyLAlt:      C.VK_LMENU,
	KeyLSystem:   C.VK_LWIN,
	KeyRControl:  C.VK_RCONTROL,
	KeyRShift:    C.VK_RSHIFT,
	KeyRAlt:      C.VK_RMENU,
	KeyRSystem:   C.VK_RWIN,
	KeyMenu:      C.VK_APPS,
	KeyLBracket:  C.VK_OEM_4,
	KeyRBracket:  C.VK_OEM_6,
	KeySemiColon: C.VK_OEM_1,
	KeyComma:     C.VK_OEM_COMMA,
	KeyPeriod:    C.VK_OEM_PERIOD,
	KeyQuote:     C.VK_OEM_7,
	KeySlash:     C.VK_OEM_2,
	KeyBackSlash: C.VK_OEM_5,
	KeyTilde:     C.VK_OEM_3,
	KeyEqual:     C.VK_OEM_PLUS,
	KeyDash:      C.VK_OEM_MINUS,
	KeySpace:     C.VK_SPACE,
	KeyReturn:    C.VK_RETURN,
	KeyBackSpace: C.VK_BACK,
	KeyTab:       C.VK_TAB,
	KeyPageUp:    C.VK_PRIOR,
	KeyPageDown:  C.VK_NEXT,
	KeyEnd:       C.VK_END,
	KeyHome:      C.VK_HOME,
	KeyInsert:    C.VK_INSERT,
	KeyDelete:    C.VK_DELETE,
	KeyAdd:       C.VK_ADD,
	KeySubtract:  C.VK_SUBTRACT,
	KeyMultiply:  C.VK_MULTIPLY,
	KeyDivide:    C.VK_DIVIDE,
	KeyLeft:      C.VK_LEFT,
	KeyRight:     C.VK_RIGHT,
	KeyUp:        C.VK_UP,
	KeyDown:      C.VK_DOWN,
	KeyNumpad0:   C.VK_NUMPAD0,
	KeyNumpad1:   C.VK_NUMPAD1,
	KeyNumpad2:   C.VK_NUMPAD2,
	KeyNumpad3:   C.VK_NUMPAD3,
	KeyNumpad4:   C.VK_NUMPAD4,
	KeyNumpad5:   C.VK_NUMPAD5,
	KeyNumpad6:   C.VK_NUMPAD6,
	KeyNumpad7:   C.VK_NUMPAD7,
	KeyNumpad8:   C.VK_NUMPAD8,
	KeyNumpad9:   C.VK_NUMPAD9,
	KeyF1:        C.VK_F1,
	KeyF2:        C.VK_F2,
	KeyF3:        C.VK_F3,
	KeyF4:        C.VK_F4,
	KeyF5:        C.VK_F5,
	KeyF6:        C.VK_F6,
	KeyF7:        C.VK_F7,
	KeyF8:        C.VK_F8,
	KeyF9:        C.VK_F9,
	KeyF10:       C.VK_F10,
	KeyF11:       C.VK_F11,
	KeyF12:       C.VK_F12,
	KeyF13:       C.VK_F13,
	KeyF14:       C.VK_F14,
	KeyF15:       C.VK_F16,
	KeyPause:     C.VK_PAUSE,
}

var lShift = C.MapVirtualKey(C.VK_LSHIFT, C.MAPVK_VK_TO_VSC)

func init() {
	for kk, vk := range keyboard_vkeys {
		keyboard_vkeys_map[vk] = Key(kk)
	}
}

// Check if a key is pressed
func IsKeyPressed(key Key) bool {
	if key < 0 || key >= KeyCount {
		return false
	}

	return uint16(C.GetAsyncKeyState(C.int(keyboard_vkeys[key])))&0x8000 != 0
}

func virtualKeyCodeToSF(vkey C.WPARAM, flags C.LPARAM) Key {
	if key, ok := keyboard_vkeys_map[C.int(vkey)]; ok {
		return key
	}

	switch vkey {
	// Check the scancode to distinguish between left and right shift
	case C.VK_SHIFT:
		scancode := C.UINT((flags & (0xFF << 16)) >> 16)
		if scancode == lShift {
			return KeyLShift
		}
		return KeyRShift

	// Check the "extended" flag to distinguish between left and right alt
	case C.VK_MENU:
		if C.__HIWORD(C.DWORD(flags))&C.KF_EXTENDED != 0 {
			return KeyRAlt
		}
		return KeyLAlt

	// Check the "extended" flag to distinguish between left and right control
	case C.VK_CONTROL:
		if C.__HIWORD(C.DWORD(flags))&C.KF_EXTENDED != 0 {
			return KeyRControl
		}
		return KeyLControl

	}

	return KeyUnknown
}
