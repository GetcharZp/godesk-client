package screen

import (
	"bytes"
	"godesk-client/internal/logger"
	"image/jpeg"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getcharzp/goscap"
	"go.uber.org/zap"
)

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

type ScreenManager struct {
	onFrame     func(frame *FrameData)
	onFrameMux  sync.RWMutex
	cap         goscap.Capturer
	isCapturing int32
	sequenceID  uint64
	quality     int
	width       int
	height      int
}

var (
	screenManager     *ScreenManager
	screenManagerOnce sync.Once
)

func GetScreenManager() *ScreenManager {
	screenManagerOnce.Do(func() {
		cap, err := goscap.NewCapturerForDisplay(0)
		if err != nil {
			logger.Error("[screen] create capturer error", zap.Error(err))
			return
		}
		screenManager = &ScreenManager{
			cap:     cap,
			quality: 80,
		}
	})
	return screenManager
}

func (m *ScreenManager) StartCapture(onFrame func(frame *FrameData)) {
	if !atomic.CompareAndSwapInt32(&m.isCapturing, 0, 1) {
		logger.Info("[screen] already capturing")
		return
	}

	m.onFrameMux.Lock()
	m.onFrame = onFrame
	m.onFrameMux.Unlock()
	atomic.StoreUint64(&m.sequenceID, 0)

	img, err := m.cap.Capture()
	if err != nil {
		logger.Error("[screen] capture error", zap.Error(err))
		atomic.StoreInt32(&m.isCapturing, 0)
		return
	}
	m.width = img.Bounds().Dx()
	m.height = img.Bounds().Dy()

	go m.captureLoop()

	logger.Info("[screen] capture started.", zap.Int("width", m.width), zap.Int("height", m.height))
}

func (m *ScreenManager) captureLoop() {
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.IsCapturing() {
				return
			}

			img, err := m.cap.Capture()
			if err != nil {
				logger.Error("[screen] capture error", zap.Error(err))
				continue
			}

			var buf bytes.Buffer
			opt := &jpeg.Options{Quality: m.quality}
			if err := jpeg.Encode(&buf, img, opt); err != nil {
				logger.Error("[screen] encode jpeg error", zap.Error(err))
				continue
			}

			seqID := atomic.AddUint64(&m.sequenceID, 1)

			frame := &FrameData{
				SequenceID: seqID,
				FrameData:  buf.Bytes(),
				Codec:      "jpeg",
				Width:      int32(m.width),
				Height:     int32(m.height),
				Timestamp:  time.Now().UnixMilli(),
				FrameType:  1,
			}

			m.onFrameMux.RLock()
			onFrame := m.onFrame
			m.onFrameMux.RUnlock()
			if onFrame != nil {
				go onFrame(frame)
			}
		}
	}
}

func (m *ScreenManager) StopCapture() {
	if !atomic.CompareAndSwapInt32(&m.isCapturing, 1, 0) {
		return
	}

	m.onFrameMux.Lock()
	m.onFrame = nil
	m.onFrameMux.Unlock()

	logger.Info("[screen] capture stopped.")
}

func (m *ScreenManager) IsCapturing() bool {
	return atomic.LoadInt32(&m.isCapturing) == 1
}

func (m *ScreenManager) GetSequenceID() uint64 {
	return atomic.LoadUint64(&m.sequenceID)
}

func (m *ScreenManager) SetQuality(quality int) {
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	m.quality = quality
}
