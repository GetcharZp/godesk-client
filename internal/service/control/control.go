package control

import (
	"godesk-client/internal/logger"
	"godesk-client/internal/service/channel"
	"godesk-client/internal/service/common"
	"godesk-client/internal/service/session"
	"time"

	"go.uber.org/zap"
)

// SendControlRequest 发送控制请求
func (in *Service) SendControlRequest(targetDeviceCode uint64, targetPassword string, requestControl bool) (string, error) {
	_, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[control] get sys config error.", zap.Error(err))
		return "", err
	}

	// 生成请求ID
	requestID := time.Now().Format("20060102150405") + string(rune(time.Now().UnixNano()%1000))

	// 创建会话
	sessionID := generateSessionID()
	sess := session.CreateSession(sessionID, targetDeviceCode, "", !requestControl)
	sess.Status = "connecting"

	// 发送控制开始请求（通过channel服务的DataStream）
	if err := channel.SendControlStartedRequest(targetDeviceCode, targetPassword, requestControl); err != nil {
		logger.Error("[control] send control started request error.", zap.Error(err))
		return "", err
	}

	logger.Info("[control] control request sent.",
		zap.String("requestID", requestID),
		zap.Uint64("targetDevice", targetDeviceCode))

	return sessionID, nil
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return time.Now().Format("20060102150405") + string(rune(time.Now().UnixNano()%1000))
}

// GetSessions 获取所有会话
func (in *Service) GetSessions() []*session.Session {
	return session.GetAllSessions()
}

// GetSession 获取指定会话
func (in *Service) GetSession(sessionId string) *session.Session {
	return session.GetSession(sessionId)
}

// RemoveSession 移除会话
func (in *Service) RemoveSession(sessionId string) {
	session.RemoveSession(sessionId)
}

// Service 控制服务
type Service struct{}

// NewService 创建控制服务
func NewService() *Service {
	return &Service{}
}
