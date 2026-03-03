package channel

import (
	"context"
	"encoding/json"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/common"
	"godesk-client/internal/service/screen"
	"godesk-client/internal/service/session"
	pb "godesk-client/proto"
	"runtime"
	"time"

	"go.uber.org/zap"
)

var (
	stream         pb.ChannelService_DataStreamClient
	ctx            context.Context
	cancelFunc     context.CancelFunc
	heartbeatTimer *time.Timer
)

func (in *Service) ClientInit(c pb.ChannelServiceClient) {
	var err error
	ctx, cancelFunc = context.WithCancel(context.Background())
	stream, err = c.DataStream(ctx)
	if err != nil {
		logger.Error("[sys] stream init error.", zap.Error(err))
		return
	}

	// 设置屏幕流数据发送函数
	screen.SetSendScreenStreamDataFunc(in.sendScreenStreamData)

	// 发送设备注册消息
	in.sendRegister()

	// 启动接收数据协程
	go in.ReceiveDataHandle()

	// 启动心跳定时器
	go in.startHeartbeat()
}

// sendScreenStreamData 发送屏幕流数据到指定控制端
func (in *Service) sendScreenStreamData(controllerUUID string, data *pb.ScreenStreamData) {
	logger.Debug("[sys] sendScreenStreamData called.",
		zap.String("controllerUUID", controllerUUID),
		zap.String("sessionId", data.SessionId),
		zap.Int("imageDataSize", len(data.ImageData)))

	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send screen data")
		return
	}

	logger.Debug("[sys] marshaling screen stream data...")
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.Error("[sys] marshal screen stream data error.", zap.Error(err))
		return
	}

	logger.Debug("[sys] screen stream data marshaled.", zap.Int("jsonSize", len(dataJSON)))

	req := &pb.ChannelRequest{
		ClientUuid: controllerUUID,
		Key:        "screen_stream_data",
		Data:       dataJSON,
	}

	logger.Debug("[sys] sending screen stream data to stream...")
	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send screen stream data error.", zap.Error(err))
		return
	}

	logger.Debug("[sys] screen stream data sent successfully.")
}

// sendRegister 发送设备注册消息
func (in *Service) sendRegister() {
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[sys] get sys config error.", zap.Error(err))
		return
	}

	registerData := &pb.RegisterData{
		Uuid:       sysConfig.Uuid,
		Os:         runtime.GOOS,
		DeviceName: sysConfig.Username,
	}

	data, err := json.Marshal(registerData)
	if err != nil {
		logger.Error("[sys] marshal register data error.", zap.Error(err))
		return
	}

	req := &pb.ChannelRequest{
		ClientUuid: sysConfig.Uuid,
		Key:        "register",
		Data:       data,
	}

	in.SendMessage(req)
	logger.Info("[sys] device register sent.", zap.String("uuid", sysConfig.Uuid))
}

// startHeartbeat 启动心跳定时发送
func (in *Service) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			in.sendHeartbeat()
		}
	}
}

// sendHeartbeat 发送心跳消息
func (in *Service) sendHeartbeat() {
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[sys] get sys config error.", zap.Error(err))
		return
	}

	heartbeatData := &pb.HeartbeatData{
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(heartbeatData)
	if err != nil {
		logger.Error("[sys] marshal heartbeat data error.", zap.Error(err))
		return
	}

	req := &pb.ChannelRequest{
		ClientUuid: sysConfig.Uuid,
		Key:        "heartbeat",
		Data:       data,
	}

	in.SendMessage(req)
	logger.Debug("[sys] heartbeat sent.")
}

func (in *Service) ReceiveDataHandle() {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := stream.Recv()
			if err != nil {
				logger.Error("[sys] stream receive error.", zap.Error(err))
				return
			}
			logger.Info("[sys] stream receive message.", zap.String("key", req.Key))
			in.handleMessage(req)
		}
	}
}

// handleMessage 处理接收到的消息
func (in *Service) handleMessage(req *pb.ChannelRequest) {
	switch req.Key {
	case "control_started":
		in.handleControlStarted(req.Data)
	case "control_ended":
		in.handleControlEnded(req.Data)
	case "start_screen_stream":
		in.handleStartScreenStream(req.Data)
	case "stop_screen_stream":
		in.handleStopScreenStream(req.Data)
	case "pause_screen_stream":
		in.handlePauseScreenStream(req.Data)
	case "resume_screen_stream":
		in.handleResumeScreenStream(req.Data)
	case "screen_stream_data":
		in.handleScreenStreamData(req.Data)
	default:
		logger.Info("[sys] unknown message key.", zap.String("key", req.Key))
	}
}

// handleControlStarted 处理控制开始通知
func (in *Service) handleControlStarted(data []byte) {
	var controlData pb.ControlStartedData
	if err := json.Unmarshal(data, &controlData); err != nil {
		logger.Error("[sys] unmarshal control started data error.", zap.Error(err))
		return
	}
	logger.Info("[sys] control started.", zap.String("session_id", controlData.SessionId))

	// 启动屏幕共享
	manager := screen.GetManager()
	if err := manager.StartSharing(
		controlData.SessionId,
		controlData.ControllerUuid,
		controlData.ControllerName,
		controlData.RequestControl,
	); err != nil {
		logger.Error("[sys] start screen sharing error.", zap.Error(err))
		return
	}

	// 发送屏幕流状态通知
	statusData := &pb.ScreenStreamStatusData{
		SessionId: controlData.SessionId,
		Status:    "started",
		Message:   "屏幕共享已开始",
		Timestamp: time.Now().Unix(),
	}
	statusJSON, _ := json.Marshal(statusData)
	in.SendMessage(&pb.ChannelRequest{
		ClientUuid: controlData.ControllerUuid,
		Key:        "screen_stream_status",
		Data:       statusJSON,
	})
}

// handleControlEnded 处理控制结束通知
func (in *Service) handleControlEnded(data []byte) {
	var controlData pb.ControlEndedData
	if err := json.Unmarshal(data, &controlData); err != nil {
		logger.Error("[sys] unmarshal control ended data error.", zap.Error(err))
		return
	}
	logger.Info("[sys] control ended.", zap.String("session_id", controlData.SessionId))

	// 停止屏幕共享
	manager := screen.GetManager()
	manager.StopSharing(controlData.SessionId)
}

// handleStartScreenStream 处理开始屏幕流请求
func (in *Service) handleStartScreenStream(data []byte) {
	var request pb.ScreenStreamRequest
	if err := json.Unmarshal(data, &request); err != nil {
		logger.Error("[sys] unmarshal screen stream request error.", zap.Error(err))
		return
	}
	logger.Info("[sys] start screen stream.", zap.String("session_id", request.SessionId))
	// TODO: 启动屏幕捕获和传输
}

// handleStopScreenStream 处理停止屏幕流请求
func (in *Service) handleStopScreenStream(data []byte) {
	logger.Info("[sys] stop screen stream.")
	// TODO: 停止屏幕捕获
}

// handlePauseScreenStream 处理暂停屏幕流请求
func (in *Service) handlePauseScreenStream(data []byte) {
	logger.Info("[sys] pause screen stream.")
	// TODO: 暂停屏幕捕获
}

// handleResumeScreenStream 处理恢复屏幕流请求
func (in *Service) handleResumeScreenStream(data []byte) {
	logger.Info("[sys] resume screen stream.")
	// TODO: 恢复屏幕捕获
}

// handleScreenStreamData 处理屏幕流数据（控制端接收被控端发送的数据）
func (in *Service) handleScreenStreamData(data []byte) {
	var streamData pb.ScreenStreamData
	if err := json.Unmarshal(data, &streamData); err != nil {
		logger.Error("[sys] unmarshal screen stream data error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received screen stream data.",
		zap.String("sessionId", streamData.SessionId),
		zap.Int("dataSize", len(streamData.ImageData)),
		zap.Int32("width", streamData.Width),
		zap.Int32("height", streamData.Height))

	// 保存图像数据到会话
	sess := session.GetSession(streamData.SessionId)
	if sess != nil {
		sess.SetLastImageData(streamData.ImageData)
		sess.ScreenWidth = streamData.Width
		sess.ScreenHeight = streamData.Height
		logger.Info("[sys] screen data saved to session.",
			zap.String("sessionId", streamData.SessionId),
			zap.Int("imageSize", len(streamData.ImageData)))
	} else {
		logger.Warn("[sys] session not found for screen data.", zap.String("sessionId", streamData.SessionId))
	}
}

func (in *Service) SendMessage(req *pb.ChannelRequest) {
	if stream == nil {
		logger.Error("[sys] stream is nil")
		return
	}
	if err := stream.Send(req); err != nil {
		logger.Error("[sys] stream send message error.", zap.Error(err))
		return
	}
}

// Close 关闭连接
func (in *Service) Close() {
	if cancelFunc != nil {
		cancelFunc()
	}
	if stream != nil {
		stream.CloseSend()
	}
}
