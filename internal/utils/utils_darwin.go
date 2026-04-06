//go:build darwin

package utils

import "github.com/go-vgo/robotgo"

// SetCursorPosAbsolute 鼠标移动
func SetCursorPosAbsolute(x, y int) error {
	robotgo.Move(x, y)
	return nil
}
