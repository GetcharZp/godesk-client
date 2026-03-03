package capture

// Frame 视频帧数据
type Frame struct {
	Data      []byte
	Width     int
	Height    int
	Timestamp int64
	Sequence  int64
}

// Config 捕获配置
type Config struct {
	FPS     int
	Quality int
	Scale   float64 // 缩放比例，1.0 表示原始大小
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		FPS:     30,
		Quality: 80,
		Scale:   1.0,
	}
}
