package session

import (
	"godesk-client/internal/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FrameData 帧数据（支持多种编码格式）
type FrameData struct {
	SequenceID uint64 `json:"sequenceId"`
	FrameData  []byte `json:"frameData"`
	Codec      string `json:"codec"`
	Width      int32  `json:"width"`
	Height     int32  `json:"height"`
	Timestamp  int64  `json:"timestamp"`
	FrameType  int32  `json:"frameType"`
	ExtraData  []byte `json:"extraData"`
}

type Session struct {
	SessionId    string `json:"sessionId"`
	DeviceCode   uint64 `json:"deviceCode"`
	DeviceName   string `json:"deviceName"`
	ViewOnly     bool   `json:"viewOnly"`
	Status       string `json:"status"`
	ScreenWidth  int32  `json:"screenWidth"`
	ScreenHeight int32  `json:"screenHeight"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	TargetUUID   string `json:"targetUuid"` // 被控端UUID，用于接收屏幕数据

	// 帧数据（支持视频流）
	lastFrameData *FrameData   `json:"-"`
	frameMux      sync.RWMutex `json:"-"`

	// 兼容旧版本的图像数据
	LastImageData []byte       `json:"-"`
	imageMux      sync.RWMutex `json:"-"`
}

var (
	sessions    = make(map[string]*Session)
	sessionsMux sync.RWMutex
)

func CreateSession(sessionId string, deviceCode uint64, deviceName string, viewOnly bool) *Session {
	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	// 检查是否已存在相同 deviceCode 的会话
	for _, existingSession := range sessions {
		if existingSession.DeviceCode == deviceCode {
			// 更新现有会话
			existingSession.SessionId = sessionId
			existingSession.DeviceName = deviceName
			existingSession.ViewOnly = viewOnly
			existingSession.Status = "connecting"
			existingSession.UpdatedAt = time.Now().Unix()
			saveSessions()
			logger.Info("[session] updated existing.", zap.String("sessionId", sessionId), zap.Uint64("deviceCode", deviceCode))
			return existingSession
		}
	}

	session := &Session{
		SessionId:  sessionId,
		DeviceCode: deviceCode,
		DeviceName: deviceName,
		ViewOnly:   viewOnly,
		Status:     "connecting",
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	sessions[sessionId] = session
	saveSessions()

	logger.Info("[session] created.", zap.String("sessionId", sessionId), zap.Uint64("deviceCode", deviceCode))

	return session
}

func GetSession(sessionId string) *Session {
	sessionsMux.RLock()
	defer sessionsMux.RUnlock()
	return sessions[sessionId]
}

func GetSessionByDeviceCode(deviceCode uint64) *Session {
	sessionsMux.RLock()
	defer sessionsMux.RUnlock()
	for _, session := range sessions {
		if session.DeviceCode == deviceCode {
			return session
		}
	}
	return nil
}

func GetSessionByTargetUUID(targetUUID string) *Session {
	sessionsMux.RLock()
	defer sessionsMux.RUnlock()
	for _, session := range sessions {
		if session.TargetUUID == targetUUID {
			return session
		}
	}
	return nil
}

func GetAllSessions() []*Session {
	sessionsMux.RLock()
	defer sessionsMux.RUnlock()

	result := make([]*Session, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, session)
	}
	return result
}

func RemoveSession(sessionId string) {
	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	delete(sessions, sessionId)
	saveSessions()

	logger.Info("[session] removed.", zap.String("sessionId", sessionId))
}

// SetLastImageData 设置最后的图像数据（兼容旧版本）
func (s *Session) SetLastImageData(data []byte) {
	s.imageMux.Lock()
	defer s.imageMux.Unlock()
	s.LastImageData = data
	s.UpdatedAt = time.Now().Unix()
}

// GetLastImageData 获取最后的图像数据（兼容旧版本）
func (s *Session) GetLastImageData() []byte {
	s.imageMux.RLock()
	defer s.imageMux.RUnlock()
	return s.LastImageData
}

// SetLastFrameData 设置最后的帧数据（支持视频流）
func (s *Session) SetLastFrameData(frame *FrameData) {
	s.frameMux.Lock()
	defer s.frameMux.Unlock()
	s.lastFrameData = frame
	s.UpdatedAt = time.Now().Unix()

	// 同时更新兼容字段（如果是 JPEG 格式）
	if frame != nil && frame.Codec == "jpeg" {
		s.imageMux.Lock()
		s.LastImageData = frame.FrameData
		s.imageMux.Unlock()
	}
}

// GetLastFrameData 获取最后的帧数据（支持视频流）
func (s *Session) GetLastFrameData() *FrameData {
	s.frameMux.RLock()
	defer s.frameMux.RUnlock()
	return s.lastFrameData
}

func saveSessions() {
	// 会话不持久化到配置文件，只保存在内存中
	// 应用重启后会话列表为空
}

func LoadSessions() {
	// 会话不持久化，启动时列表为空
	sessionsMux.Lock()
	defer sessionsMux.Unlock()
	sessions = make(map[string]*Session)
	logger.Info("[session] sessions not persisted, starting with empty list.")
}
