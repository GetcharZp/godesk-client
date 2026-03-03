package screen

import (
	"godesk-client/internal/logger"
	"godesk-client/internal/service/capture"
	"godesk-client/internal/service/common"
	pb "godesk-client/proto"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager 屏幕共享管理器
// 负责管理被控端的屏幕共享，支持多控制端共享同一份截图数据
type Manager struct {
	captureService *capture.Service
	sessions       map[string]*ControlSession // session_id -> ControlSession
	sessionsMu     sync.RWMutex
	isCapturing    bool
	sequence       int64
}

// ControlSession 控制会话
type ControlSession struct {
	SessionID      string
	ControllerUUID string
	ControllerName string
	RequestControl bool
	StartedAt      int64
	SendChan       chan *pb.ScreenStreamData
}

var (
	manager     *Manager
	managerOnce sync.Once
)

// GetManager 获取屏幕共享管理器单例
func GetManager() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{
			sessions:       make(map[string]*ControlSession),
			captureService: capture.NewService(),
		}
	})
	return manager
}

// StartSharing 开始屏幕共享
func (m *Manager) StartSharing(sessionID, controllerUUID, controllerName string, requestControl bool) error {
	m.sessionsMu.Lock()
	defer m.sessionsMu.Unlock()

	// 检查会话是否已存在
	if _, exists := m.sessions[sessionID]; exists {
		logger.Info("[screen] session already exists.", zap.String("sessionId", sessionID))
		return nil
	}

	// 创建控制会话
	session := &ControlSession{
		SessionID:      sessionID,
		ControllerUUID: controllerUUID,
		ControllerName: controllerName,
		RequestControl: requestControl,
		StartedAt:      time.Now().Unix(),
		SendChan:       make(chan *pb.ScreenStreamData, 10),
	}
	m.sessions[sessionID] = session

	logger.Info("[screen] control session added.",
		zap.String("sessionId", sessionID),
		zap.String("controller", controllerName),
		zap.Bool("requestControl", requestControl))

	// 如果没有正在捕获，启动捕获
	if !m.isCapturing {
		m.startCapture()
	}

	return nil
}

// StopSharing 停止屏幕共享
func (m *Manager) StopSharing(sessionID string) {
	m.sessionsMu.Lock()
	defer m.sessionsMu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return
	}

	// 关闭发送通道
	close(session.SendChan)
	delete(m.sessions, sessionID)

	logger.Info("[screen] control session removed.", zap.String("sessionId", sessionID))

	// 如果没有会话了，停止捕获
	if len(m.sessions) == 0 && m.isCapturing {
		m.stopCapture()
	}
}

// GetSession 获取控制会话
func (m *Manager) GetSession(sessionID string) *ControlSession {
	m.sessionsMu.RLock()
	defer m.sessionsMu.RUnlock()
	return m.sessions[sessionID]
}

// GetAllSessions 获取所有控制会话
func (m *Manager) GetAllSessions() []*ControlSession {
	m.sessionsMu.RLock()
	defer m.sessionsMu.RUnlock()

	result := make([]*ControlSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		result = append(result, session)
	}
	return result
}

// startCapture 开始屏幕捕获
func (m *Manager) startCapture() {
	if m.isCapturing {
		return
	}

	m.isCapturing = true
	m.captureService.Start(func(data []byte, width, height int) {
		m.broadcastFrame(data, width, height)
	})

	// 发送状态通知
	m.sendStatusToAll("started", "屏幕共享已开始")

	logger.Info("[screen] capture started.")
}

// stopCapture 停止屏幕捕获
func (m *Manager) stopCapture() {
	if !m.isCapturing {
		return
	}

	m.isCapturing = false
	m.captureService.Stop()

	// 发送状态通知
	m.sendStatusToAll("stopped", "屏幕共享已停止")

	logger.Info("[screen] capture stopped.")
}

// broadcastFrame 广播帧数据到所有控制端
func (m *Manager) broadcastFrame(data []byte, width, height int) {
	m.sessionsMu.RLock()
	sessions := make([]*ControlSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	m.sessionsMu.RUnlock()

	if len(sessions) == 0 {
		logger.Debug("[screen] no sessions to broadcast")
		return
	}

	m.sequence++

	logger.Info("[screen] broadcasting frame.",
		zap.Int("sessions", len(sessions)),
		zap.Int("dataSize", len(data)),
		zap.Int("width", width),
		zap.Int("height", height))

	// 发送到所有控制端
	for _, session := range sessions {
		frame := &pb.ScreenStreamData{
			SessionId: session.SessionID,
			ImageData: data,
			Format:    "jpeg",
			Width:     int32(width),
			Height:    int32(height),
			Timestamp: time.Now().Unix(),
			Sequence:  m.sequence,
		}

		if sendScreenStreamData != nil {
			sendScreenStreamData(session.ControllerUUID, frame)
			logger.Debug("[screen] frame sent.", zap.String("sessionId", session.SessionID))
		}
	}
}

// sendStatusToAll 发送状态到所有控制端
func (m *Manager) sendStatusToAll(status, message string) {
	m.sessionsMu.RLock()
	sessions := make([]*ControlSession, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	m.sessionsMu.RUnlock()

	logger.Info("[screen] sending status to all.", zap.String("status", status), zap.String("message", message), zap.Int("sessions", len(sessions)))

	for _, session := range sessions {
		// 通过 channel 服务发送状态
		if sendScreenStreamData != nil {
			sendScreenStreamData(session.ControllerUUID, &pb.ScreenStreamData{
				SessionId: session.SessionID,
				Timestamp: time.Now().Unix(),
			})
		}
	}
}

// sendScreenStreamData 发送屏幕流数据（由 channel 服务实现）
var sendScreenStreamData func(controllerUUID string, data *pb.ScreenStreamData)

// SetSendScreenStreamDataFunc 设置发送数据函数
func SetSendScreenStreamDataFunc(f func(controllerUUID string, data *pb.ScreenStreamData)) {
	sendScreenStreamData = f
}

// VerifyPassword 验证控制密码
func VerifyPassword(inputPassword string) bool {
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[screen] get sys config error.", zap.Error(err))
		return false
	}
	return sysConfig.Password == inputPassword
}

// IsSharing 检查是否正在共享
func (m *Manager) IsSharing() bool {
	m.sessionsMu.RLock()
	defer m.sessionsMu.RUnlock()
	return m.isCapturing
}

// GetSessionCount 获取当前控制会话数量
func (m *Manager) GetSessionCount() int {
	m.sessionsMu.RLock()
	defer m.sessionsMu.RUnlock()
	return len(m.sessions)
}
