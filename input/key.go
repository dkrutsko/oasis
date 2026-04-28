//go:build darwin || linux

package input

////////////////////////////////////////////////////////////////////////////////

// Key represents a Windows virtual key code. These values
// are sent to the target machine via the Moonlight shared
// memory input queue.
type Key uint16

////////////////////////////////////////////////////////////////////////////////

const (
	KeySpace     Key = 0x20 // VK_SPACE
	KeyEscape    Key = 0x1B // VK_ESCAPE

	KeyTab      Key = 0x09 // VK_TAB
	KeyAlt      Key = 0x12 // VK_MENU
	KeyLAlt     Key = 0xA4 // VK_LMENU
	KeyRAlt     Key = 0xA5 // VK_RMENU
	KeyControl  Key = 0x11 // VK_CONTROL
	KeyLControl Key = 0xA2 // VK_LCONTROL
	KeyRControl Key = 0xA3 // VK_RCONTROL
	KeyShift    Key = 0x10 // VK_SHIFT
	KeyLShift   Key = 0xA0 // VK_LSHIFT
	KeyRShift   Key = 0xA1 // VK_RSHIFT
	KeySystem   Key = 0x5B // VK_LWIN
	KeyLSystem  Key = 0x5B // VK_LWIN
	KeyRSystem  Key = 0x5C // VK_RWIN
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyF1  Key = 0x70 // VK_F1
	KeyF2  Key = 0x71 // VK_F2
	KeyF3  Key = 0x72 // VK_F3
	KeyF4  Key = 0x73 // VK_F4
	KeyF5  Key = 0x74 // VK_F5
	KeyF6  Key = 0x75 // VK_F6
	KeyF7  Key = 0x76 // VK_F7
	KeyF8  Key = 0x77 // VK_F8
	KeyF9  Key = 0x78 // VK_F9
	KeyF10 Key = 0x79 // VK_F10
	KeyF11 Key = 0x7A // VK_F11
	KeyF12 Key = 0x7B // VK_F12
)

////////////////////////////////////////////////////////////////////////////////

const (
	Key0 Key = 0x30 // VK_0
	Key1 Key = 0x31 // VK_1
	Key2 Key = 0x32 // VK_2
	Key3 Key = 0x33 // VK_3
	Key4 Key = 0x34 // VK_4
	Key5 Key = 0x35 // VK_5
	Key6 Key = 0x36 // VK_6
	Key7 Key = 0x37 // VK_7
	Key8 Key = 0x38 // VK_8
	Key9 Key = 0x39 // VK_9
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyA Key = 0x41 // VK_A
	KeyB Key = 0x42 // VK_B
	KeyC Key = 0x43 // VK_C
	KeyD Key = 0x44 // VK_D
	KeyE Key = 0x45 // VK_E
	KeyF Key = 0x46 // VK_F
	KeyG Key = 0x47 // VK_G
	KeyH Key = 0x48 // VK_H
	KeyI Key = 0x49 // VK_I
	KeyJ Key = 0x4A // VK_J
	KeyK Key = 0x4B // VK_K
	KeyL Key = 0x4C // VK_L
	KeyM Key = 0x4D // VK_M
	KeyN Key = 0x4E // VK_N
	KeyO Key = 0x4F // VK_O
	KeyP Key = 0x50 // VK_P
	KeyQ Key = 0x51 // VK_Q
	KeyR Key = 0x52 // VK_R
	KeyS Key = 0x53 // VK_S
	KeyT Key = 0x54 // VK_T
	KeyU Key = 0x55 // VK_U
	KeyV Key = 0x56 // VK_V
	KeyW Key = 0x57 // VK_W
	KeyX Key = 0x58 // VK_X
	KeyY Key = 0x59 // VK_Y
	KeyZ Key = 0x5A // VK_Z
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyGrave     Key = 0xC0 // VK_OEM_3
	KeyMinus     Key = 0xBD // VK_OEM_MINUS
	KeyEqual     Key = 0xBB // VK_OEM_PLUS
	KeyBackspace Key = 0x08 // VK_BACK
	KeyLBracket  Key = 0xDB // VK_OEM_4
	KeyRBracket  Key = 0xDD // VK_OEM_6
	KeyBackslash Key = 0xDC // VK_OEM_5
	KeySemicolon Key = 0xBA // VK_OEM_1
	KeyQuote     Key = 0xDE // VK_OEM_7
	KeyReturn    Key = 0x0D // VK_RETURN
	KeyComma     Key = 0xBC // VK_OEM_COMMA
	KeyPeriod    Key = 0xBE // VK_OEM_PERIOD
	KeySlash     Key = 0xBF // VK_OEM_2
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyLeft  Key = 0x25 // VK_LEFT
	KeyUp    Key = 0x26 // VK_UP
	KeyRight Key = 0x27 // VK_RIGHT
	KeyDown  Key = 0x28 // VK_DOWN
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyPrint    Key = 0x2C // VK_SNAPSHOT
	KeyPause    Key = 0x13 // VK_PAUSE
	KeyInsert   Key = 0x2D // VK_INSERT
	KeyDelete   Key = 0x2E // VK_DELETE
	KeyHome     Key = 0x24 // VK_HOME
	KeyEnd      Key = 0x23 // VK_END
	KeyPageUp   Key = 0x21 // VK_PRIOR
	KeyPageDown Key = 0x22 // VK_NEXT
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyAdd      Key = 0x6B // VK_ADD
	KeySubtract Key = 0x6D // VK_SUBTRACT
	KeyMultiply Key = 0x6A // VK_MULTIPLY
	KeyDivide   Key = 0x6F // VK_DIVIDE
	KeyDecimal  Key = 0x6E // VK_DECIMAL
	KeyEnter    Key = 0x0D // VK_RETURN
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyNum0 Key = 0x60 // VK_NUMPAD0
	KeyNum1 Key = 0x61 // VK_NUMPAD1
	KeyNum2 Key = 0x62 // VK_NUMPAD2
	KeyNum3 Key = 0x63 // VK_NUMPAD3
	KeyNum4 Key = 0x64 // VK_NUMPAD4
	KeyNum5 Key = 0x65 // VK_NUMPAD5
	KeyNum6 Key = 0x66 // VK_NUMPAD6
	KeyNum7 Key = 0x67 // VK_NUMPAD7
	KeyNum8 Key = 0x68 // VK_NUMPAD8
	KeyNum9 Key = 0x69 // VK_NUMPAD9
)

////////////////////////////////////////////////////////////////////////////////

const (
	KeyCapsLock   Key = 0x14 // VK_CAPITAL
	KeyScrollLock Key = 0x91 // VK_SCROLL
	KeyNumLock    Key = 0x90 // VK_NUMLOCK
)
