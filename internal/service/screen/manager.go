package screen

import (
	"encoding/base64"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/capture"
	"sync"
	"sync/atomic"
)

// ScreenManager 屏幕管理器
type ScreenManager struct {
	captureService *capture.Service
	onFrame        func(imageData string, width, height int)
	isCapturing    int32 // 0 = false, 1 = true
}

var (
	screenManager     *ScreenManager
	screenManagerOnce sync.Once
)

// GetScreenManager 获取屏幕管理器单例
func GetScreenManager() *ScreenManager {
	screenManagerOnce.Do(func() {
		screenManager = &ScreenManager{
			captureService: capture.NewService(),
		}
	})
	return screenManager
}

// StartCapture 开始屏幕捕获
func (m *ScreenManager) StartCapture(onFrame func(imageData string, width, height int)) {
	if !atomic.CompareAndSwapInt32(&m.isCapturing, 0, 1) {
		logger.Info("[screen] already capturing")
		return
	}

	m.onFrame = onFrame

	// 启动捕获服务
	m.captureService.Start(func(data []byte, width, height int) {
		// 将图片数据编码为Base64
		imageData := base64.StdEncoding.EncodeToString(data)

		// logger.Debug("[screen] frame captured.",
		// 	zap.Int("width", width),
		// 	zap.Int("height", height),
		// 	zap.Int("dataSize", len(data)),
		// 	zap.Int("base64Size", len(imageData)))

		// 调用回调发送数据
		if m.onFrame != nil {
			m.onFrame(imageData, width, height)
		}
	})

	logger.Info("[screen] capture started.")
}

// StopCapture 停止屏幕捕获
func (m *ScreenManager) StopCapture() {
	if !atomic.CompareAndSwapInt32(&m.isCapturing, 1, 0) {
		return
	}

	m.captureService.Stop()
	m.onFrame = nil

	logger.Info("[screen] capture stopped.")
}

// IsCapturing 是否正在捕获
func (m *ScreenManager) IsCapturing() bool {
	return atomic.LoadInt32(&m.isCapturing) == 1
}
