package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"godesk-client/internal"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/cache"
	"godesk-client/internal/service/channel"
	"godesk-client/internal/service/common"
	"godesk-client/internal/service/control"
	"godesk-client/internal/service/device"
	"godesk-client/internal/service/file"
	"godesk-client/internal/service/models"
	"godesk-client/internal/service/session"
	"godesk-client/internal/service/sys"
	"godesk-client/internal/service/user"
	pb "godesk-client/proto"
	"os"
	"path/filepath"
	"time"

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
	return resp(cache.GetSysConfig(), nil)
}

// SaveSysConfig 保存系统配置
func (a *App) SaveSysConfig(cfg *models.SysConfig) any {
	return resp(nil, common.UpdateSysConfig(cfg))
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

// GetControlSessions 获取远程控制类型的会话
func (a *App) GetControlSessions() any {
	return resp(session.GetSessionsByType("control"), nil)
}

// GetSession 获取单个会话
func (a *App) GetSession(sessionId string) any {
	return resp(session.GetSession(sessionId), nil)
}

// GetSessionByDeviceCode 根据设备码获取会话
func (a *App) GetSessionByDeviceCode(deviceCode uint64) any {
	return resp(session.GetSessionByDeviceCode(deviceCode), nil)
}

// GetControlSessionByDeviceCode 根据设备码获取远程控制类型的会话
func (a *App) GetControlSessionByDeviceCode(deviceCode uint64) any {
	return resp(session.GetSessionByDeviceCodeAndType(deviceCode, "control"), nil)
}

// GetFileSessionByDeviceCode 根据设备码获取文件访问类型的会话
func (a *App) GetFileSessionByDeviceCode(deviceCode uint64) any {
	return resp(session.GetSessionByDeviceCodeAndType(deviceCode, "file"), nil)
}

// CreateSession 创建会话
func (a *App) CreateSession(sessionId string, deviceCode uint64, deviceName string, viewOnly bool, sessionType string) any {
	return resp(session.CreateSession(sessionId, deviceCode, deviceName, viewOnly, sessionType), nil)
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

	// 获取帧数据
	frameData := sess.GetLastFrameData()
	if frameData != nil {
		// 目前只支持 JPEG 格式
		// 后续支持其他格式时，在这里添加相应的处理逻辑
		return resp(map[string]any{
			"sessionId": sess.SessionId,
			"imageData": base64.StdEncoding.EncodeToString(frameData.FrameData),
			"sequence":  frameData.SequenceID,
			"width":     frameData.Width,
			"height":    frameData.Height,
			"codec":     frameData.Codec,
			"timestamp": frameData.Timestamp,
		}, nil)
	}

	// 兼容旧版本：获取图像数据
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
		"codec":     "jpeg",
	}, nil)
}

// SendMouseMove 发送鼠标移动事件
func (a *App) SendMouseMove(sessionId string, x, y int32) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, nil)
	}
	return resp(nil, channel.SendMouseMove(sess.TargetUUID, x, y))
}

// SendMouseClick 发送鼠标点击事件
func (a *App) SendMouseClick(sessionId string, x, y int32, button int32, action string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, nil)
	}
	return resp(nil, channel.SendMouseClick(sess.TargetUUID, x, y, button, action))
}

// SendMouseScroll 发送鼠标滚轮事件
func (a *App) SendMouseScroll(sessionId string, x, y int32, deltaX, deltaY float64) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, nil)
	}
	return resp(nil, channel.SendMouseScroll(sess.TargetUUID, x, y, deltaX, deltaY))
}

// SendKeyDown 发送键盘按下事件
func (a *App) SendKeyDown(sessionId string, key string, modifiers []string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, nil)
	}
	return resp(nil, channel.SendKeyDown(sess.TargetUUID, key, modifiers))
}

// SendKeyUp 发送键盘释放事件
func (a *App) SendKeyUp(sessionId string, key string, modifiers []string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, nil)
	}
	return resp(nil, channel.SendKeyUp(sess.TargetUUID, key, modifiers))
}

// ========== 文件管理相关 API ==========

// ListLocalFiles 获取本地文件列表
func (a *App) ListLocalFiles(path string) any {
	files, err := (&file.Service{}).ListLocalFiles(path)
	return resp(files, err)
}

// GetLocalDrives 获取本地驱动器列表
func (a *App) GetLocalDrives() any {
	drives, err := (&file.Service{}).GetLocalDrives()
	return resp(drives, err)
}

// GetRemoteFileList 获取远程文件列表（从缓存）
func (a *App) GetRemoteFileList(targetUUID string, path string) any {
	data, ok := cache.GetRemoteFileList(targetUUID, path)
	if !ok {
		logger.Info("[app] remote file list not found in cache.", zap.String("targetUUID", targetUUID), zap.String("path", path))
		return resp(nil, fmt.Errorf("no cached data"))
	}
	logger.Info("[app] remote file list found in cache.", zap.String("targetUUID", targetUUID), zap.String("path", path), zap.Int("fileCount", len(data.Files)))
	return resp(data, nil)
}

// RequestRemoteFileList 请求远程文件列表
func (a *App) RequestRemoteFileList(sessionId string, path string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, fmt.Errorf("session not found or targetUUID is empty"))
	}
	return resp(nil, channel.SendFileListRequest(sess.TargetUUID, sess.DeviceCode, path))
}

// UploadFile 上传文件到远程设备
func (a *App) UploadFile(sessionId string, localPath string, remotePath string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, fmt.Errorf("session not found or targetUUID is empty"))
	}

	fileInfo, err := file.GetFileInfo(localPath)
	if err != nil {
		return resp(nil, fmt.Errorf("failed to get file info: %v", err))
	}

	if fileInfo.IsDir {
		return resp(nil, fmt.Errorf("directory upload not supported"))
	}

	transferId := fmt.Sprintf("%d", time.Now().UnixMilli())
	chunkSize := int32(64 * 1024) // 64KB

	// 本地测试：直接复制文件
	if sess.TargetUUID == channel.GetMyUUID() {
		logger.Info("[app] local upload, copying file directly.", zap.String("localPath", localPath), zap.String("remotePath", remotePath))
		fileData, err := file.ReadFile(localPath)
		if err != nil {
			return resp(nil, fmt.Errorf("failed to read file: %v", err))
		}
		if err := file.WriteFile(remotePath, fileData); err != nil {
			return resp(nil, fmt.Errorf("failed to write file: %v", err))
		}
		return resp(map[string]any{
			"transferId": transferId,
			"totalSize":  fileInfo.Size,
			"complete":   true,
		}, nil)
	}

	// 初始化发送端进度缓存
	cache.InitUploadProgress(transferId, fileInfo.Size)

	// 远程上传：发送文件数据
	if err := channel.SendFileTransferStart(sess.TargetUUID, transferId, "upload", localPath, remotePath, fileInfo.Size, chunkSize); err != nil {
		cache.ClearFileTransfer(transferId)
		return resp(nil, err)
	}

	fileData, err := file.ReadFile(localPath)
	if err != nil {
		cache.ClearFileTransfer(transferId)
		return resp(nil, fmt.Errorf("failed to read file: %v", err))
	}

	go func() {
		totalChunks := int32(len(fileData) / int(chunkSize))
		if len(fileData)%int(chunkSize) != 0 {
			totalChunks++
		}

		logger.Info("[app] starting file upload.", zap.String("transferId", transferId), zap.Int32("totalChunks", totalChunks))

		for i := int32(0); i < totalChunks; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > int32(len(fileData)) {
				end = int32(len(fileData))
			}

			chunk := fileData[start:end]
			isLast := i == totalChunks-1

			if err := channel.SendFileTransferData(sess.TargetUUID, transferId, i, chunk, isLast, 0); err != nil {
				logger.Error("[app] send file transfer data error", zap.Error(err))
				cache.SetFileTransferComplete(transferId, false, err.Error())
				return
			}

			// 更新发送进度
			cache.UpdateUploadProgress(transferId, int64(end))

			time.Sleep(10 * time.Millisecond)
		}

		// 标记发送完成
		cache.SetFileTransferComplete(transferId, true, "")
		logger.Info("[app] file upload complete.", zap.String("transferId", transferId))
	}()

	return resp(map[string]any{
		"transferId": transferId,
		"totalSize":  fileInfo.Size,
		"chunkSize":  chunkSize,
	}, nil)
}

// DownloadFile 从远程设备下载文件
func (a *App) DownloadFile(sessionId string, remotePath string, localPath string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, fmt.Errorf("session not found or targetUUID is empty"))
	}

	transferId := fmt.Sprintf("%d", time.Now().UnixMilli())
	chunkSize := int32(64 * 1024)

	// 本地测试：直接复制文件
	if sess.TargetUUID == channel.GetMyUUID() {
		logger.Info("[app] local download, copying file directly.", zap.String("remotePath", remotePath), zap.String("localPath", localPath))
		fileData, err := file.ReadFile(remotePath)
		if err != nil {
			return resp(nil, fmt.Errorf("failed to read file: %v", err))
		}
		if err := file.WriteFile(localPath, fileData); err != nil {
			return resp(nil, fmt.Errorf("failed to write file: %v", err))
		}
		return resp(map[string]any{
			"transferId": transferId,
			"totalSize":  len(fileData),
			"complete":   true,
		}, nil)
	}

	// 初始化下载进度缓存（接收端会在收到数据时更新）
	cache.InitDownloadProgress(transferId, localPath)

	if err := channel.SendFileTransferStart(sess.TargetUUID, transferId, "download", remotePath, localPath, 0, chunkSize); err != nil {
		cache.ClearFileTransfer(transferId)
		return resp(nil, err)
	}

	return resp(map[string]any{
		"transferId": transferId,
		"chunkSize":  chunkSize,
	}, nil)
}

// GetFileTransferStatus 获取文件传输状态
func (a *App) GetFileTransferStatus(transferId string) any {
	exists, complete, errMsg, received, total := cache.GetFileTransferStatus(transferId)
	return resp(map[string]any{
		"exists":   exists,
		"complete": complete,
		"error":    errMsg,
		"received": received,
		"total":    total,
	}, nil)
}

// CheckFileExists 检查文件是否存在
func (a *App) CheckFileExists(path string) any {
	_, err := os.Stat(path)
	exists := !os.IsNotExist(err)
	return resp(map[string]any{
		"exists": exists,
	}, nil)
}

// RenameLocalFile 重命名本地文件
func (a *App) RenameLocalFile(oldPath string, newName string) any {
	if err := file.RenameFile(oldPath, newName); err != nil {
		return resp(nil, err)
	}
	dir := filepath.Dir(oldPath)
	newPath := filepath.Clean(filepath.Join(dir, newName))
	return resp(map[string]any{
		"oldPath": oldPath,
		"newPath": newPath,
		"newName": newName,
	}, nil)
}

// RenameRemoteFile 重命名远程文件
func (a *App) RenameRemoteFile(sessionId string, oldPath string, newName string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, fmt.Errorf("session not found or targetUUID is empty"))
	}

	requestId := fmt.Sprintf("%d", time.Now().UnixMilli())

	// 本地测试：直接重命名
	if sess.TargetUUID == channel.GetMyUUID() {
		if err := file.RenameFile(oldPath, newName); err != nil {
			return resp(nil, err)
		}
		dir := filepath.Dir(oldPath)
		newPath := filepath.Clean(filepath.Join(dir, newName))
		return resp(map[string]any{
			"requestId": requestId,
			"code":      0,
			"newPath":   newPath,
		}, nil)
	}

	// 发送重命名请求到远程
	if err := channel.SendFileRenameRequest(sess.TargetUUID, requestId, oldPath, newName); err != nil {
		return resp(nil, err)
	}

	return resp(map[string]any{
		"requestId": requestId,
	}, nil)
}

// GetFileRenameResult 获取文件重命名结果
func (a *App) GetFileRenameResult(requestId string) any {
	result := cache.GetFileRenameResult(requestId)
	if result == nil {
		return resp(map[string]any{
			"exists": false,
		}, nil)
	}
	return resp(map[string]any{
		"exists":   true,
		"code":     result.Code,
		"message":  result.Message,
		"newPath":  result.NewPath,
		"complete": true,
	}, nil)
}

// DeleteLocalFile 删除本地文件
func (a *App) DeleteLocalFile(path string) any {
	if err := file.DeleteFile(path); err != nil {
		return resp(nil, err)
	}
	return resp(map[string]any{
		"deletedPath": path,
	}, nil)
}

// DeleteRemoteFile 删除远程文件
func (a *App) DeleteRemoteFile(sessionId string, path string, force bool) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, fmt.Errorf("session not found or targetUUID is empty"))
	}

	requestId := fmt.Sprintf("%d", time.Now().UnixMilli())

	if sess.TargetUUID == channel.GetMyUUID() {
		if err := file.DeleteFile(path); err != nil {
			return resp(nil, err)
		}
		return resp(map[string]any{
			"requestId":   requestId,
			"code":        0,
			"deletedPath": path,
		}, nil)
	}

	if err := channel.SendFileDeleteRequest(sess.TargetUUID, requestId, path, force); err != nil {
		return resp(nil, err)
	}

	return resp(map[string]any{
		"requestId": requestId,
	}, nil)
}

// GetFileDeleteResult 获取文件删除结果
func (a *App) GetFileDeleteResult(requestId string) any {
	result := cache.GetFileDeleteResult(requestId)
	if result == nil {
		return resp(map[string]any{
			"exists": false,
		}, nil)
	}
	return resp(map[string]any{
		"exists":      true,
		"code":        result.Code,
		"message":     result.Message,
		"deletedPath": result.DeletedPath,
		"complete":    true,
	}, nil)
}

// CancelFileTransfer 取消文件传输
func (a *App) CancelFileTransfer(sessionId string, transferId string, reason string) any {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.TargetUUID == "" {
		return resp(nil, fmt.Errorf("session not found or targetUUID is empty"))
	}

	if err := channel.SendFileTransferCancel(sess.TargetUUID, transferId, reason); err != nil {
		return resp(nil, err)
	}

	cache.ClearFileTransfer(transferId)
	return resp(nil, nil)
}
