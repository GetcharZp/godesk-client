package video

import (
	"bytes"
	"godesk-client/internal/logger"
	"image"
	"image/jpeg"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// CodecType 编码格式类型
type CodecType string

const (
	// CodecJPEG JPEG 编码
	CodecJPEG CodecType = "jpeg"
)

// Encoder 视频编码器接口
type Encoder interface {
	// Start 开始编码，onEncoded 回调返回编码后的数据
	Start(onEncoded func(data []byte, isKeyFrame bool))
	// Stop 停止编码
	Stop()
	// SendFrame 发送一帧图像进行编码
	SendFrame(img image.Image)
	// GetSequenceID 获取当前序列号
	GetSequenceID() uint64
	// Close 关闭编码器并释放资源
	Close()
	// GetCodec 获取编码格式类型
	GetCodec() CodecType
}

// EncoderConfig 编码器配置
type EncoderConfig struct {
	Width   int
	Height  int
	FPS     int
	Bitrate int64
	Codec   CodecType
	Quality int // JPEG 质量 (1-100)
}

// NewEncoder 创建视频编码器
func NewEncoder(config EncoderConfig) (Encoder, error) {
	return NewJPEGEncoder(config)
}

// JPEGEncoder JPEG 编码器实现
type JPEGEncoder struct {
	width   int
	height  int
	fps     int
	quality int
	codec   CodecType

	isRunning int32
	stopChan  chan struct{}
	inputChan chan image.Image
	onEncoded func(data []byte, isKeyFrame bool)

	sequenceID uint64
	mux        sync.Mutex
}

// NewJPEGEncoder 创建 JPEG 编码器
func NewJPEGEncoder(config EncoderConfig) (*JPEGEncoder, error) {
	quality := config.Quality
	if quality < 1 {
		quality = 80
	}
	if quality > 100 {
		quality = 100
	}

	e := &JPEGEncoder{
		width:     config.Width,
		height:    config.Height,
		fps:       config.FPS,
		quality:   quality,
		codec:     CodecJPEG,
		stopChan:  make(chan struct{}),
		inputChan: make(chan image.Image, 10),
	}

	logger.Info("[video] JPEG encoder created",
		zap.Int("width", e.width),
		zap.Int("height", e.height),
		zap.Int("fps", e.fps),
		zap.Int("quality", e.quality))

	return e, nil
}

// Start 开始编码
func (e *JPEGEncoder) Start(onEncoded func(data []byte, isKeyFrame bool)) {
	if !atomic.CompareAndSwapInt32(&e.isRunning, 0, 1) {
		return
	}
	e.onEncoded = onEncoded
	go e.encodeLoop()
	logger.Info("[video] JPEG encoder started")
}

// Stop 停止编码
func (e *JPEGEncoder) Stop() {
	if !atomic.CompareAndSwapInt32(&e.isRunning, 1, 0) {
		return
	}
	close(e.stopChan)
	close(e.inputChan)
	logger.Info("[video] JPEG encoder stopped")
}

// encodeLoop 编码主循环
func (e *JPEGEncoder) encodeLoop() {
	ticker := time.NewTicker(time.Second / time.Duration(e.fps))
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case img, ok := <-e.inputChan:
			if !ok {
				return
			}
			if err := e.encodeFrame(img); err != nil {
				logger.Error("[video] JPEG encode frame error", zap.Error(err))
			}
		case <-ticker.C:
			// 保持帧率
		}
	}
}

// encodeFrame 编码单帧
func (e *JPEGEncoder) encodeFrame(img image.Image) error {
	if img == nil {
		return nil
	}

	e.mux.Lock()
	defer e.mux.Unlock()

	var buf bytes.Buffer
	opt := &jpeg.Options{
		Quality: e.quality,
	}
	if err := jpeg.Encode(&buf, img, opt); err != nil {
		return err
	}

	data := buf.Bytes()
	atomic.AddUint64(&e.sequenceID, 1)

	// JPEG 所有帧都是关键帧 (I-frame)
	if e.onEncoded != nil {
		go e.onEncoded(data, true)
	}

	return nil
}

// SendFrame 发送帧进行编码
func (e *JPEGEncoder) SendFrame(img image.Image) {
	if atomic.LoadInt32(&e.isRunning) == 0 {
		return
	}
	select {
	case e.inputChan <- img:
	default:
		// 队列满，丢弃帧
		logger.Debug("[video] JPEG encoder input queue full, dropping frame")
	}
}

// GetSequenceID 获取当前序列号
func (e *JPEGEncoder) GetSequenceID() uint64 {
	return atomic.LoadUint64(&e.sequenceID)
}

// Close 关闭编码器
func (e *JPEGEncoder) Close() {
	e.Stop()
}

// GetCodec 获取编码格式
func (e *JPEGEncoder) GetCodec() CodecType {
	return e.codec
}
