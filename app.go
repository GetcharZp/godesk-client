package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"godesk-client/internal"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/channel"
	"godesk-client/internal/service/common"
	"godesk-client/internal/service/control"
	"godesk-client/internal/service/device"
	"godesk-client/internal/service/session"
	"godesk-client/internal/service/sys"
	"godesk-client/internal/service/user"
	pb "godesk-client/proto"

	"go.uber.org/zap"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	(&device.Service{}).ClientInit()
	(&user.Service{}).ClientInit()
	(&channel.Service{}).ClientInit(pb.NewChannelServiceClient(define.GrpcConn))
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func resp(data any, err error) any {
	if err == nil {
		return map[string]any{
			"code": 200,
			"msg":  "success",
			"data": data,
		}
	} else {
		return map[string]any{
			"code": -1,
			"msg":  err.Error(),
		}
	}
}

// DeviceInfo 设备信息
func (a *App) DeviceInfo() any {
	return resp((&device.Service{}).Info())
}

// GetUserInfo 获取用户信息
func (a *App) GetUserInfo() any {
	return resp((&user.Service{}).GetUserInfo())
}

// UserLogin 用户登录
func (a *App) UserLogin(req *pb.UserLoginRequest) any {
	return resp((&user.Service{}).Login(req))
}

// UserRegister 用户注册
func (a *App) UserRegister(req *pb.UserRegisterRequest) any {
	return resp((&user.Service{}).Register(req))
}

// UserLogout 退出登录
func (a *App) UserLogout() any {
	return resp((&user.Service{}).Logout())
}

// GetDeviceList 获取设备列表
func (a *App) GetDeviceList() any {
	return resp((&device.Service{}).List())
}

// AddDevice 添加设备
func (a *App) AddDevice(req *pb.AddDeviceRequest) any {
	return resp(nil, (&device.Service{}).Add(req))
}

// EditDevice 编辑设备
func (a *App) EditDevice(req *pb.EditDeviceRequest) any {
	return resp(nil, (&device.Service{}).Edit(req))
}

// DeleteDevice 删除设备
func (a *App) DeleteDevice(req *pb.DeleteDeviceRequest) any {
	return resp(nil, (&device.Service{}).Delete(req))
}

// GetSysConfig 获取系统配置
func (a *App) GetSysConfig() any {
	return resp(common.GetSysConfig())
}

// SaveSysConfig 保存系统配置
func (a *App) SaveSysConfig(cfg *define.SysConfig) any {
	return resp(nil, common.SaveSysConfig(cfg))
}

// Reconnect 重新连接服务
func (a *App) Reconnect() any {
	internal.Reconnect()
	return resp(nil, nil)
}

// GetConnectionStatus 获取连接状态
func (a *App) GetConnectionStatus() any {
	return resp((&sys.Service{}).GetConnectionStatus(), nil)
}

// SendControlRequest 发送控制请求
func (a *App) SendControlRequest(targetDeviceCode uint64, targetPassword string, requestControl bool) any {
	return resp((&control.Service{}).SendControlRequest(targetDeviceCode, targetPassword, requestControl))
}

// SendDisconnectNotify 发送断开连接通知
func (a *App) SendDisconnectNotify(sessionId string, targetDeviceCode uint64) any {
	// 获取会话，获取被控端UUID
	sess := session.GetSession(sessionId)
	if sess != nil && sess.TargetUUID != "" {
		// 发送控制结束请求给被控端
		if err := channel.SendControlEndedRequest(targetDeviceCode, sess.TargetUUID); err != nil {
			logger.Error("[app] send disconnect notify error.", zap.Error(err))
		}
	}
	return resp(nil, nil)
}

// StopScreenStream 停止屏幕流
func (a *App) StopScreenStream(sessionId string) any {
	// TODO: 实现停止屏幕流
	return resp(nil, nil)
}

// GetAllSessions 获取所有会话
func (a *App) GetAllSessions() any {
	return resp(session.GetAllSessions(), nil)
}

// GetSession 获取单个会话
func (a *App) GetSession(sessionId string) any {
	return resp(session.GetSession(sessionId), nil)
}

// GetSessionByDeviceCode 根据设备码获取会话
func (a *App) GetSessionByDeviceCode(deviceCode uint64) any {
	return resp(session.GetSessionByDeviceCode(deviceCode), nil)
}

// CreateSession 创建会话
func (a *App) CreateSession(sessionId string, deviceCode uint64, deviceName string, viewOnly bool) any {
	return resp(session.CreateSession(sessionId, deviceCode, deviceName, viewOnly), nil)
}

// RemoveSession 移除会话
func (a *App) RemoveSession(sessionId string) any {
	session.RemoveSession(sessionId)
	return resp(nil, nil)
}

// GetSessionImage 获取会话的最新图像数据
func (a *App) GetSessionImage(sessionId string) any {
	sess := session.GetSession(sessionId)
	if sess == nil {
		return resp(nil, nil)
	}

	imageData := sess.GetLastImageData()
	if imageData == nil {
		return resp(nil, nil)
	}

	// 返回 base64 编码的图像数据
	return resp(map[string]any{
		"sessionId": sess.SessionId,
		"imageData": base64.StdEncoding.EncodeToString(imageData),
		"sequence":  sess.UpdatedAt,
		"width":     sess.ScreenWidth,
		"height":    sess.ScreenHeight,
	}, nil)
}
