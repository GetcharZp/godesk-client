//go:build windows
// +build windows

package utils

import (
	"godesk-client/internal/logger"
	"unsafe"

	"go.uber.org/zap"
	"golang.org/x/sys/windows"
)

var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	procSetCursorPos     = user32.NewProc("SetCursorPos")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procSendInput        = user32.NewProc("SendInput")
	procGetDC            = user32.NewProc("GetDC")
	procReleaseDC        = user32.NewProc("ReleaseDC")
	procGetDeviceCaps    = user32.NewProc("GetDeviceCaps")
	gdi32                = windows.NewLazySystemDLL("gdi32.dll")
	procGetDeviceCapsGDI = gdi32.NewProc("GetDeviceCaps")
)

const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
	LOGPIXELSX  = 88
	LOGPIXELSY  = 90
)

// POINT 结构体
type POINT struct {
	X, Y int32
}

// INPUT 结构体用于 SendInput
type INPUT struct {
	Type uint32
	Mi   MOUSEINPUT
}

// MOUSEINPUT 结构体
type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

const (
	INPUT_MOUSE          = 0
	MOUSEEVENTF_MOVE     = 0x0001
	MOUSEEVENTF_ABSOLUTE = 0x8000
)

// GetScreenSize 获取屏幕实际像素尺寸（考虑DPI缩放）
func GetScreenSize() (width, height int) {
	// 使用 GetSystemMetrics 获取逻辑像素
	cx, _, _ := procGetSystemMetrics.Call(uintptr(SM_CXSCREEN))
	cy, _, _ := procGetSystemMetrics.Call(uintptr(SM_CYSCREEN))
	return int(cx), int(cy)
}

// GetScreenSizePhysical 获取屏幕物理像素尺寸
func GetScreenSizePhysical() (width, height int) {
	// 获取主显示器 DC
	dc, _, _ := procGetDC.Call(0)
	if dc == 0 {
		return GetScreenSize()
	}
	defer procReleaseDC.Call(0, dc)

	// 获取水平方向的物理像素数
	widthPhys, _, _ := procGetDeviceCapsGDI.Call(dc, uintptr(8)) // HORZRES = 8
	// 获取垂直方向的物理像素数
	heightPhys, _, _ := procGetDeviceCapsGDI.Call(dc, uintptr(10)) // VERTRES = 10

	return int(widthPhys), int(heightPhys)
}

// GetDPI 获取屏幕 DPI
func GetDPI() (dpiX, dpiY int) {
	dc, _, _ := procGetDC.Call(0)
	if dc == 0 {
		return 96, 96 // 默认 DPI
	}
	defer procReleaseDC.Call(0, dc)

	dpiX64, _, _ := procGetDeviceCapsGDI.Call(dc, uintptr(LOGPIXELSX))
	dpiY64, _, _ := procGetDeviceCapsGDI.Call(dc, uintptr(LOGPIXELSY))

	return int(dpiX64), int(dpiY64)
}

// SetCursorPosAbsolute 设置鼠标绝对位置（使用物理像素坐标）
func SetCursorPosAbsolute(x, y int) error {
	// 直接使用 SetCursorPos，它使用物理像素
	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		logger.Error("[mouse] SetCursorPos failed", zap.Error(err), zap.Int("x", x), zap.Int("y", y))
		return err
	}
	return nil
}

// MoveMouseAbsolute 使用 SendInput 移动鼠标到绝对位置
func MoveMouseAbsolute(x, y int, screenWidth, screenHeight int) error {
	// SendInput 使用 0-65535 的归一化坐标
	// 将物理像素坐标转换为归一化坐标
	normalizedX := int32((x * 65535) / screenWidth)
	normalizedY := int32((y * 65535) / screenHeight)

	input := INPUT{
		Type: INPUT_MOUSE,
		Mi: MOUSEINPUT{
			Dx:      normalizedX,
			Dy:      normalizedY,
			DwFlags: MOUSEEVENTF_MOVE | MOUSEEVENTF_ABSOLUTE,
		},
	}

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		uintptr(unsafe.Sizeof(input)),
	)

	if ret == 0 {
		logger.Error("[mouse] SendInput failed", zap.Error(err))
		return err
	}

	return nil
}

// GetCursorPos 获取当前鼠标位置
func GetCursorPos() (x, y int, err error) {
	var pt POINT
	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return 0, 0, err
	}
	return int(pt.X), int(pt.Y), nil
}
