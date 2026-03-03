package capture

import (
	"bytes"
	"godesk-client/internal/logger"
	"image"
	"image/jpeg"
	"sync"
	"time"

	"github.com/kbinani/screenshot"
	"go.uber.org/zap"
)

// SendFrameFunc 发送帧数据的函数类型
type SendFrameFunc func(data []byte, width, height int)

// Service 屏幕捕获服务
type Service struct {
	isRunning   bool
	stopChan    chan struct{}
	sendFrame   SendFrameFunc
	fps         int
	quality     int
	mu          sync.RWMutex
	lastFrame   []byte
	lastFrameMu sync.RWMutex
}

// NewService 创建屏幕捕获服务
func NewService() *Service {
	return &Service{
		stopChan: make(chan struct{}),
		fps:      30,
		quality:  80,
	}
}

// Start 开始屏幕捕获
func (s *Service) Start(sendFrame SendFrameFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return nil
	}

	s.sendFrame = sendFrame
	s.isRunning = true
	s.stopChan = make(chan struct{})

	go s.captureLoop()

	logger.Info("[capture] screen capture started.", zap.Int("fps", s.fps), zap.Int("quality", s.quality))
	return nil
}

// Stop 停止屏幕捕获
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}

	s.isRunning = false
	close(s.stopChan)

	logger.Info("[capture] screen capture stopped.")
}

// IsRunning 检查是否正在运行
func (s *Service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// SetFPS 设置帧率
func (s *Service) SetFPS(fps int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fps = fps
}

// SetQuality 设置图像质量
func (s *Service) SetQuality(quality int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	s.quality = quality
}

// captureLoop 捕获循环
func (s *Service) captureLoop() {
	ticker := time.NewTicker(time.Second / time.Duration(s.getFPS()))
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			logger.Debug("[capture] ticker triggered.", zap.Bool("isRunning", s.IsRunning()))
			if !s.IsRunning() {
				return
			}
			s.captureFrame()
		}
	}
}

// captureFrame 捕获一帧
func (s *Service) captureFrame() {
	logger.Debug("[capture] capturing frame...")

	// 获取主显示器截图
	bounds := screenshot.GetDisplayBounds(0)
	width, height := bounds.Dx(), bounds.Dy()
	logger.Debug("[capture] display bounds.", zap.Int("x", bounds.Min.X), zap.Int("y", bounds.Min.Y), zap.Int("width", width), zap.Int("height", height))

	img, err := screenshot.CaptureDisplay(0)
	if err != nil {
		logger.Error("[capture] capture display error.", zap.Error(err))
		return
	}

	logger.Debug("[capture] display captured.", zap.Int("width", img.Bounds().Dx()), zap.Int("height", img.Bounds().Dy()))

	// 编码为 JPEG
	data, err := s.encodeJPEG(img)
	if err != nil {
		logger.Error("[capture] encode jpeg error.", zap.Error(err))
		return
	}

	logger.Debug("[capture] frame encoded.", zap.Int("size", len(data)))

	// 保存最后一帧
	s.lastFrameMu.Lock()
	s.lastFrame = data
	s.lastFrameMu.Unlock()

	// 发送帧数据
	if s.sendFrame != nil {
		logger.Debug("[capture] sending frame.", zap.Int("dataSize", len(data)))
		go s.sendFrame(data, width, height) // 使用 goroutine 避免阻塞
	} else {
		logger.Warn("[capture] sendFrame is nil!")
	}
}

// encodeJPEG 编码为 JPEG
func (s *Service) encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	opt := &jpeg.Options{
		Quality: s.getQuality(),
	}
	if err := jpeg.Encode(&buf, img, opt); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetLastFrame 获取最后一帧
func (s *Service) GetLastFrame() []byte {
	s.lastFrameMu.RLock()
	defer s.lastFrameMu.RUnlock()
	return s.lastFrame
}

// getFPS 获取当前帧率
func (s *Service) getFPS() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fps
}

// getQuality 获取当前图像质量
func (s *Service) getQuality() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.quality
}

// GetScreenSize 获取屏幕尺寸
func GetScreenSize() (width, height int) {
	bounds := screenshot.GetDisplayBounds(0)
	return bounds.Dx(), bounds.Dy()
}
