package session

import (
	"encoding/base64"
	"godesk-client/internal/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

type FrameData struct {
	SequenceID  uint64 `json:"sequenceId"`
	FrameData   []byte `json:"frameData"`
	FrameData64 string `json:"frameData64"`
	Codec       string `json:"codec"`
	Width       int32  `json:"width"`
	Height      int32  `json:"height"`
	Timestamp   int64  `json:"timestamp"`
	FrameType   int32  `json:"frameType"`
	ExtraData   []byte `json:"extraData"`
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
	TargetUUID   string `json:"targetUuid"`
	SessionType  string `json:"sessionType"`

	lastFrameData *FrameData   `json:"-"`
	frameMux      sync.RWMutex `json:"-"`

	LastImageData []byte       `json:"-"`
	imageMux      sync.RWMutex `json:"-"`
}

var (
	sessions    = make(map[string]*Session)
	sessionsMux sync.RWMutex
)

func CreateSession(sessionId string, deviceCode uint64, deviceName string, viewOnly bool, sessionType string) *Session {
	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	for _, existingSession := range sessions {
		if existingSession.DeviceCode == deviceCode && existingSession.SessionType == sessionType {
			existingSession.SessionId = sessionId
			existingSession.DeviceName = deviceName
			existingSession.ViewOnly = viewOnly
			existingSession.Status = "connecting"
			existingSession.UpdatedAt = time.Now().Unix()
			logger.Info("[session] updated existing.", zap.String("sessionId", sessionId), zap.Uint64("deviceCode", deviceCode), zap.String("sessionType", sessionType))
			return existingSession
		}
	}

	session := &Session{
		SessionId:   sessionId,
		DeviceCode:  deviceCode,
		DeviceName:  deviceName,
		ViewOnly:    viewOnly,
		Status:      "connecting",
		SessionType: sessionType,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	sessions[sessionId] = session

	logger.Info("[session] created.", zap.String("sessionId", sessionId), zap.Uint64("deviceCode", deviceCode), zap.String("sessionType", sessionType))

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

func GetSessionByDeviceCodeAndType(deviceCode uint64, sessionType string) *Session {
	sessionsMux.RLock()
	defer sessionsMux.RUnlock()
	for _, session := range sessions {
		if session.DeviceCode == deviceCode && session.SessionType == sessionType {
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

func GetSessionsByType(sessionType string) []*Session {
	sessionsMux.RLock()
	defer sessionsMux.RUnlock()

	result := make([]*Session, 0)
	for _, session := range sessions {
		if session.SessionType == sessionType {
			result = append(result, session)
		}
	}
	return result
}

func RemoveSession(sessionId string) {
	sessionsMux.Lock()
	defer sessionsMux.Unlock()

	delete(sessions, sessionId)

	logger.Info("[session] removed.", zap.String("sessionId", sessionId))
}

func (s *Session) SetLastImageData(data []byte) {
	s.imageMux.Lock()
	defer s.imageMux.Unlock()
	s.LastImageData = data
	s.UpdatedAt = time.Now().Unix()
}

func (s *Session) GetLastImageData() []byte {
	s.imageMux.RLock()
	defer s.imageMux.RUnlock()
	return s.LastImageData
}

func (s *Session) SetLastFrameData(frame *FrameData) {
	s.frameMux.Lock()
	defer s.frameMux.Unlock()

	if frame != nil && frame.FrameData != nil && frame.FrameData64 == "" {
		frame.FrameData64 = base64.StdEncoding.EncodeToString(frame.FrameData)
	}

	s.lastFrameData = frame
	s.UpdatedAt = time.Now().Unix()

	if frame != nil && frame.Codec == "jpeg" {
		s.imageMux.Lock()
		s.LastImageData = frame.FrameData
		s.imageMux.Unlock()
	}
}

func (s *Session) GetLastFrameData() *FrameData {
	s.frameMux.RLock()
	defer s.frameMux.RUnlock()
	return s.lastFrameData
}

func LoadSessions() {
	sessionsMux.Lock()
	defer sessionsMux.Unlock()
	sessions = make(map[string]*Session)
	logger.Info("[session] sessions not persisted, starting with empty list.")
}
