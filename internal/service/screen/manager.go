package screen

import (
	"bytes"
	"godesk-client/internal/logger"
	"image/jpeg"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kbinani/screenshot"
	"go.uber.org/zap"
)

// FrameData 帧数据
type FrameData struct {
	SequenceID uint64
	FrameData  []byte
	Codec      string
	Width      int32
	Height     int32
	Timestamp  int64
	FrameType  int32
	ExtraData  []byte
}

// ScreenManager 屏幕管理器
type ScreenManager struct {
	onFrame     func(frame *FrameData)
	isCapturing int32 // 0 = false, 1 = true
	sequenceID  uint64
	quality     int
	width       int
	height      int
}

var (
	screenManager     *ScreenManager
	screenManagerOnce sync.Once
)

// GetScreenManager 获取屏幕管理器单例
func GetScreenManager() *ScreenManager {
	screenManagerOnce.Do(func() {
		screenManager = &ScreenManager{
			quality: 80, // JPEG 质量
		}
	})
	return screenManager
}

// StartCapture 开始屏幕捕获
func (m *ScreenManager) StartCapture(onFrame func(frame *FrameData)) {
	if !atomic.CompareAndSwapInt32(&m.isCapturing, 0, 1) {
		logger.Info("[screen] already capturing")
		return
	}

	m.onFrame = onFrame
	atomic.StoreUint64(&m.sequenceID, 0)

	// 获取屏幕尺寸
	bounds := screenshot.GetDisplayBounds(0)
	m.width = bounds.Dx()
	m.height = bounds.Dy()

	// 启动屏幕捕获循环
	go m.captureLoop()

	logger.Info("[screen] capture started.", zap.Int("width", m.width), zap.Int("height", m.height))
}

// captureLoop 屏幕捕获循环 - 直接捕获并编码，减少延迟
func (m *ScreenManager) captureLoop() {
	ticker := time.NewTicker(time.Second / 30) // 30fps
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.IsCapturing() {
				return
			}

			// 直接捕获屏幕
			img, err := screenshot.CaptureDisplay(0)
			if err != nil {
				logger.Error("[screen] capture display error", zap.Error(err))
				continue
			}

			// 直接编码为 JPEG
			var buf bytes.Buffer
			opt := &jpeg.Options{Quality: m.quality}
			if err := jpeg.Encode(&buf, img, opt); err != nil {
				logger.Error("[screen] encode jpeg error", zap.Error(err))
				continue
			}

			// 递增序列号并发送帧
			seqID := atomic.AddUint64(&m.sequenceID, 1)

			frame := &FrameData{
				SequenceID: seqID,
				FrameData:  buf.Bytes(),
				Codec:      "jpeg",
				Width:      int32(m.width),
				Height:     int32(m.height),
				Timestamp:  time.Now().UnixMilli(),
				FrameType:  1, // I-frame
			}

			if m.onFrame != nil {
				// 使用 goroutine 避免阻塞捕获循环
				go m.onFrame(frame)
			}
		}
	}
}

// StopCapture 停止屏幕捕获
func (m *ScreenManager) StopCapture() {
	if !atomic.CompareAndSwapInt32(&m.isCapturing, 1, 0) {
		return
	}

	m.onFrame = nil

	logger.Info("[screen] capture stopped.")
}

// IsCapturing 是否正在捕获
func (m *ScreenManager) IsCapturing() bool {
	return atomic.LoadInt32(&m.isCapturing) == 1
}

// GetSequenceID 获取当前序列号
func (m *ScreenManager) GetSequenceID() uint64 {
	return atomic.LoadUint64(&m.sequenceID)
}

// SetQuality 设置 JPEG 质量
func (m *ScreenManager) SetQuality(quality int) {
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	m.quality = quality
}
