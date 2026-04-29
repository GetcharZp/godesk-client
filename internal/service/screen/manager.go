package screen

import (
	"bytes"
	"godesk-client/internal/logger"
	"image"
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

	lastFrameHash    uint64
	lastFrameTime    time.Time
	staticFrameCount int

	bufPool sync.Pool

	frameChan chan *FrameData
	stopChan  chan struct{}
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
			cap:       cap,
			quality:   75,
			bufPool:   sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
			frameChan: make(chan *FrameData, 2),
			stopChan:  make(chan struct{}),
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
	m.lastFrameHash = 0
	m.staticFrameCount = 0
	m.lastFrameTime = time.Now()

	img, err := m.cap.Capture()
	if err != nil {
		logger.Error("[screen] capture error", zap.Error(err))
		atomic.StoreInt32(&m.isCapturing, 0)
		return
	}
	m.width = img.Bounds().Dx()
	m.height = img.Bounds().Dy()

	go m.captureLoop()
	go m.sendLoop()

	logger.Info("[screen] capture started.", zap.Int("width", m.width), zap.Int("height", m.height))
}

func (m *ScreenManager) captureLoop() {
	baseInterval := time.Second / 30
	maxInterval := time.Second / 5

	currentInterval := baseInterval
	ticker := time.NewTicker(currentInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			if !m.IsCapturing() {
				return
			}

			img, err := m.cap.Capture()
			if err != nil {
				logger.Error("[screen] capture error", zap.Error(err))
				continue
			}

			hash := m.calculateImageHash(img)

			if hash == m.lastFrameHash {
				m.staticFrameCount++
				if m.staticFrameCount > 10 {
					currentInterval = maxInterval
				} else if m.staticFrameCount > 5 {
					currentInterval = time.Second / 10
				} else if m.staticFrameCount > 2 {
					currentInterval = time.Second / 20
				}
				ticker.Reset(currentInterval)
				continue
			}

			m.lastFrameHash = hash
			m.staticFrameCount = 0
			currentInterval = baseInterval
			ticker.Reset(currentInterval)

			frame, err := m.encodeFrame(img)
			if err != nil {
				logger.Error("[screen] encode frame error", zap.Error(err))
				continue
			}

			select {
			case m.frameChan <- frame:
			default:
			}
		}
	}
}

func (m *ScreenManager) sendLoop() {
	for {
		select {
		case <-m.stopChan:
			return
		case frame := <-m.frameChan:
			m.onFrameMux.RLock()
			onFrame := m.onFrame
			m.onFrameMux.RUnlock()
			if onFrame != nil {
				onFrame(frame)
			}
		}
	}
}

func (m *ScreenManager) calculateImageHash(img image.Image) uint64 {
	bounds := img.Bounds()
	width := bounds.Dx()

	var hash uint64 = 0
	sampleStep := 20

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleStep {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleStep {
			r, g, b, _ := img.At(x, y).RGBA()
			pixelHash := uint64(r>>8) + uint64(g>>8)<<8 + uint64(b>>8)<<16
			hash ^= pixelHash + uint64(x*width+y)
		}
	}

	return hash
}

func (m *ScreenManager) encodeFrame(img image.Image) (*FrameData, error) {
	buf := m.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer m.bufPool.Put(buf)

	opt := &jpeg.Options{Quality: m.quality}
	if err := jpeg.Encode(buf, img, opt); err != nil {
		return nil, err
	}

	seqID := atomic.AddUint64(&m.sequenceID, 1)

	frameData := make([]byte, buf.Len())
	copy(frameData, buf.Bytes())

	frame := &FrameData{
		SequenceID: seqID,
		FrameData:  frameData,
		Codec:      "jpeg",
		Width:      int32(m.width),
		Height:     int32(m.height),
		Timestamp:  time.Now().UnixMilli(),
		FrameType:  1,
	}

	return frame, nil
}

func (m *ScreenManager) StopCapture() {
	if !atomic.CompareAndSwapInt32(&m.isCapturing, 1, 0) {
		return
	}

	close(m.stopChan)

	m.onFrameMux.Lock()
	m.onFrame = nil
	m.onFrameMux.Unlock()

	m.stopChan = make(chan struct{})
	m.frameChan = make(chan *FrameData, 2)

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

func (m *ScreenManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"width":            m.width,
		"height":           m.height,
		"quality":          m.quality,
		"isCapturing":      m.IsCapturing(),
		"sequenceID":       m.GetSequenceID(),
		"staticFrameCount": m.staticFrameCount,
	}
}
