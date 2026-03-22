package channel

import (
	"context"
	"encoding/json"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/cache"
	"godesk-client/internal/service/screen"
	"godesk-client/internal/service/session"
	"godesk-client/internal/utils"
	pb "godesk-client/proto"
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	stream         pb.ChannelService_DataStreamClient
	ctx            context.Context
	cancelFunc     context.CancelFunc
	heartbeatTimer *time.Timer
	myUUID         string // 本机UUID
)

func (in *Service) ClientInit(c pb.ChannelServiceClient) {
	var err error

	// 获取系统配置
	sysConfig := cache.GetSysConfig()

	myUUID = sysConfig.Uuid

	// 创建带有 AccessToken 的 context
	baseCtx := context.Background()
	if sysConfig.AccessToken != "" {
		md := metadata.New(map[string]string{
			"accesstoken": sysConfig.AccessToken,
		})
		baseCtx = metadata.NewOutgoingContext(baseCtx, md)
	}

	ctx, cancelFunc = context.WithCancel(baseCtx)
	stream, err = c.DataStream(ctx)
	if err != nil {
		logger.Error("[sys] stream init error.", zap.Error(err))
		return
	}

	// 发送设备注册消息
	in.sendRegister()

	// 启动接收数据协程
	go in.ReceiveDataHandle()

	// 启动心跳定时器
	go in.startHeartbeat()
}

// sendRegister 发送设备注册消息
func (in *Service) sendRegister() {
	sysConfig := cache.GetSysConfig()

	registerData := &pb.RegisterData{
		Os:         runtime.GOOS,
		DeviceName: sysConfig.Username,
	}

	data, err := json.Marshal(registerData)
	if err != nil {
		logger.Error("[sys] marshal register data error.", zap.Error(err))
		return
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   sysConfig.Uuid,
		TargetClientUuid: "", // 服务器
		Key:              "register",
		Data:             data,
	}

	in.SendMessage(req)
	logger.Info("[sys] device registered.", zap.String("uuid", sysConfig.Uuid))
}

// startHeartbeat 启动心跳定时器
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

// sendHeartbeat 发送心跳
func (in *Service) sendHeartbeat() {
	sysConfig := cache.GetSysConfig()

	heartbeatData := &pb.HeartbeatData{
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(heartbeatData)
	if err != nil {
		logger.Error("[sys] marshal heartbeat data error.", zap.Error(err))
		return
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   sysConfig.Uuid,
		TargetClientUuid: "",
		Key:              "heartbeat",
		Data:             data,
	}

	in.SendMessage(req)
	// logger.Debug("[sys] heartbeat sent.")
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
			in.handleMessage(req)
		}
	}
}

// handleMessage 处理接收到的消息
func (in *Service) handleMessage(req *pb.ChannelRequest) {
	switch req.Key {
	case "control_started_request":
		in.handleControlStartedRequest(req)
	case "control_started_response":
		in.handleControlStartedResponse(req)
	case "control_ended_request":
		in.handleControlEndedRequest(req)
	case "control_ended_response":
		in.handleControlEndedResponse(req)
	case "screen_stream_data":
		in.handleScreenStreamData(req)
	case "mouse_move":
		in.handleMouseMove(req)
	case "mouse_click":
		in.handleMouseClick(req)
	case "mouse_scroll":
		in.handleMouseScroll(req)
	case "key_down":
		in.handleKeyDown(req)
	case "key_up":
		in.handleKeyUp(req)
	default:
		logger.Info("[sys] unknown message key.", zap.String("key", req.Key))
	}
}

// handleControlStartedRequest 处理控制开始请求（被控端收到）
func (in *Service) handleControlStartedRequest(req *pb.ChannelRequest) {
	var data pb.ControlStartedRequestData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal control started request error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received control started request.",
		zap.Uint64("targetCode", data.TargetCode),
		zap.Bool("requestControl", data.RequestControl))

	// 验证密码
	sysConfig := cache.GetSysConfig()
	if data.TargetPassword != sysConfig.Password {
		logger.Warn("[sys] password verification failed.")
		// 发送拒绝响应
		in.sendControlStartedResponse(req.SendClientUuid, 2, "")
		return
	}

	// 启动屏幕捕获
	manager := screen.GetScreenManager()
	manager.StartCapture(func(frame *screen.FrameData) {
		in.sendScreenStreamData(req.SendClientUuid, frame)
	})

	// 发送接受响应
	in.sendControlStartedResponse(req.SendClientUuid, 0, sysConfig.Uuid)
}

// sendControlStartedResponse 发送控制开始响应
func (in *Service) sendControlStartedResponse(targetUUID string, code int32, uuid string) {
	resp := &pb.ControlStartedResponseData{
		Code:       code,
		Uuid:       uuid,
		TargetCode: 0,
	}

	data, _ := json.Marshal(resp)
	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "control_started_response",
		Data:             data,
	}

	in.SendMessage(req)
	logger.Info("[sys] control started response sent.", zap.Int32("code", code))
}

// handleControlStartedResponse 处理控制开始响应（控制端收到）
func (in *Service) handleControlStartedResponse(req *pb.ChannelRequest) {
	var data pb.ControlStartedResponseData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal control started response error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received control started response.", zap.Int32("code", data.Code), zap.String("uuid", data.Uuid))

	if data.Code == 0 {
		// 控制开始成功，更新会话状态
		// 根据 deviceCode 查找会话并设置 TargetUUID
		sess := session.GetSessionByDeviceCode(data.TargetCode)
		if sess != nil {
			sess.TargetUUID = data.Uuid
			sess.Status = "connected"
			logger.Info("[sys] session updated with target UUID.", zap.String("sessionId", sess.SessionId), zap.String("targetUUID", data.Uuid))
		} else {
			logger.Warn("[sys] no session found for device code.", zap.Uint64("deviceCode", data.TargetCode))
		}
	} else {
		logger.Warn("[sys] control started failed.", zap.Int32("code", data.Code))
	}
}

// handleControlEndedRequest 处理控制结束请求
func (in *Service) handleControlEndedRequest(req *pb.ChannelRequest) {
	logger.Info("[sys] received control ended request.")
	// 停止屏幕捕获
	manager := screen.GetScreenManager()
	manager.StopCapture()

	// 发送响应
	resp := &pb.ControlEndedResponseData{Code: 0}
	data, _ := json.Marshal(resp)
	in.SendMessage(&pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: req.SendClientUuid,
		Key:              "control_ended_response",
		Data:             data,
	})
}

// handleControlEndedResponse 处理控制结束响应
func (in *Service) handleControlEndedResponse(req *pb.ChannelRequest) {
	logger.Info("[sys] received control ended response.")
}

// sendScreenStreamData 发送屏幕流数据（被控端调用）
func (in *Service) sendScreenStreamData(targetUUID string, frame *screen.FrameData) {
	streamData := &pb.ScreenStreamData{
		SequenceId: frame.SequenceID,
		FrameData:  frame.FrameData,
		Codec:      frame.Codec,
		Width:      frame.Width,
		Height:     frame.Height,
		Timestamp:  frame.Timestamp,
		FrameType:  frame.FrameType,
		ExtraData:  frame.ExtraData,
	}

	data, err := json.Marshal(streamData)
	if err != nil {
		logger.Error("[sys] marshal screen stream data error.", zap.Error(err))
		return
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "screen_stream_data",
		Data:             data,
	}

	if err := in.SendMessage(req); err != nil {
		logger.Error("[sys] send screen stream data error.", zap.Error(err))
	} else {
		// logger.Debug("[sys] screen stream data sent.", zap.String("target", targetUUID), zap.Uint64("sequence", frame.SequenceID))
	}
}

// handleScreenStreamData 处理屏幕流数据（控制端收到）
func (in *Service) handleScreenStreamData(req *pb.ChannelRequest) {
	var data pb.ScreenStreamData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal screen stream data error.", zap.Error(err))
		return
	}

	// logger.Debug("[sys] received screen stream data.",
	// 	zap.Uint64("sequence", data.SequenceId),
	// 	zap.Int("width", int(data.Width)),
	// 	zap.Int("height", int(data.Height)),
	// 	zap.Int("dataSize", len(data.FrameData)))

	// 保存到会话（使用发送方UUID查找会话）
	sess := session.GetSessionByTargetUUID(req.SendClientUuid)
	if sess != nil {
		// 根据 codec 类型处理帧数据
		switch data.Codec {
		case "jpeg":
			// JPEG 直接保存帧数据
			sess.SetLastFrameData(&session.FrameData{
				SequenceID: data.SequenceId,
				FrameData:  data.FrameData,
				Codec:      data.Codec,
				Width:      data.Width,
				Height:     data.Height,
				Timestamp:  data.Timestamp,
				FrameType:  data.FrameType,
				ExtraData:  data.ExtraData,
			})
		case "h264", "h265":
			// H.264/H.265 视频流数据（需要解码器支持）
			// TODO: 实现视频解码
			logger.Warn("[sys] video codec not supported yet.", zap.String("codec", data.Codec))
		default:
			logger.Warn("[sys] unknown codec.", zap.String("codec", data.Codec))
		}
		sess.ScreenWidth = data.Width
		sess.ScreenHeight = data.Height
		// logger.Debug("[sys] screen data saved to session.", zap.String("targetUUID", req.SendClientUuid))
	} else {
		logger.Warn("[sys] no session found for target UUID.", zap.String("targetUUID", req.SendClientUuid))
	}
}

// handleMouseMove 处理鼠标移动事件（被控端收到）
func (in *Service) handleMouseMove(req *pb.ChannelRequest) {
	var data pb.MouseMoveData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal mouse move error.", zap.Error(err))
		return
	}

	logger.Debug("[sys] received mouse move.", zap.Int32("x", data.X), zap.Int32("y", data.Y))

	// 使用 Windows API 直接设置鼠标位置（避免 DPI 缩放问题）
	if runtime.GOOS == "windows" {
		utils.SetCursorPosAbsolute(int(data.X), int(data.Y))
	} else {
		// 其他平台使用 robotgo
		robotgo.Move(int(data.X), int(data.Y))
	}
}

// handleMouseClick 处理鼠标点击事件（被控端收到）
func (in *Service) handleMouseClick(req *pb.ChannelRequest) {
	var data pb.MouseClickData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal mouse click error.", zap.Error(err))
		return
	}

	logger.Debug("[sys] received mouse click.",
		zap.Int32("x", data.X),
		zap.Int32("y", data.Y),
		zap.Int32("button", data.Button),
		zap.String("action", data.Action))

	// 先移动鼠标到指定位置（使用 Windows API 避免 DPI 问题）
	if runtime.GOOS == "windows" {
		utils.SetCursorPosAbsolute(int(data.X), int(data.Y))
	} else {
		robotgo.Move(int(data.X), int(data.Y))
	}

	// 根据按钮和动作执行点击
	switch data.Button {
	case 0: // 左键
		if data.Action == "down" {
			robotgo.MouseDown("left")
		} else {
			robotgo.MouseUp("left")
		}
	case 1: // 中键
		if data.Action == "down" {
			robotgo.MouseDown("center")
		} else {
			robotgo.MouseUp("center")
		}
	case 2: // 右键
		if data.Action == "down" {
			robotgo.MouseDown("right")
		} else {
			robotgo.MouseUp("right")
		}
	}
}

// handleMouseScroll 处理鼠标滚轮事件（被控端收到）
func (in *Service) handleMouseScroll(req *pb.ChannelRequest) {
	var data pb.MouseScrollData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal mouse scroll error.", zap.Error(err))
		return
	}

	logger.Debug("[sys] received mouse scroll.",
		zap.Int32("deltaY", data.DeltaY))

	// 使用 robotgo 执行鼠标滚轮
	if data.DeltaY != 0 {
		robotgo.Scroll(int(data.DeltaY), int(data.X), int(data.Y))
	}
}

// handleKeyDown 处理键盘按下事件（被控端收到）
func (in *Service) handleKeyDown(req *pb.ChannelRequest) {
	var data pb.KeyDownData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal key down error.", zap.Error(err))
		return
	}

	logger.Debug("[sys] received key down.", zap.String("key", data.Key), zap.Strings("modifiers", data.Modifiers))

	// 使用 robotgo 执行键盘按下
	args := []interface{}{"down"}
	for _, m := range data.Modifiers {
		args = append(args, m)
	}
	robotgo.KeyToggle(data.Key, args...)
}

// handleKeyUp 处理键盘释放事件（被控端收到）
func (in *Service) handleKeyUp(req *pb.ChannelRequest) {
	var data pb.KeyUpData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal key up error.", zap.Error(err))
		return
	}

	logger.Debug("[sys] received key up.", zap.String("key", data.Key), zap.Strings("modifiers", data.Modifiers))

	// 使用 robotgo 执行键盘释放
	args := []interface{}{"up"}
	for _, m := range data.Modifiers {
		args = append(args, m)
	}
	robotgo.KeyToggle(data.Key, args...)
}

func (in *Service) SendMessage(req *pb.ChannelRequest) error {
	if stream == nil {
		logger.Error("[sys] stream is nil")
		return nil
	}
	if err := stream.Send(req); err != nil {
		logger.Error("[sys] stream send message error.", zap.Error(err))
		return err
	}
	return nil
}

// SendControlStartedRequest 发送控制开始请求（供control服务调用）
func SendControlStartedRequest(targetDeviceCode uint64, targetPassword string, requestControl bool) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send control request")
		return nil
	}

	reqData := &pb.ControlStartedRequestData{
		TargetCode:     targetDeviceCode,
		TargetPassword: targetPassword,
		RequestControl: requestControl,
		Timestamp:      time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal control started request error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: "", // 服务器会根据target_code查找被控端
		Key:              "control_started_request",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send control started request error.", zap.Error(err))
		return err
	}

	logger.Info("[sys] control started request sent.",
		zap.Uint64("targetCode", targetDeviceCode),
		zap.Bool("requestControl", requestControl))
	return nil
}

// SendControlEndedRequest 发送控制结束请求（供control服务调用）
func SendControlEndedRequest(targetDeviceCode uint64, targetUUID string) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send control ended request")
		return nil
	}

	reqData := &pb.ControlEndedRequestData{
		TargetCode: targetDeviceCode,
		Timestamp:  time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal control ended request error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "control_ended_request",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send control ended request error.", zap.Error(err))
		return err
	}

	logger.Info("[sys] control ended request sent.",
		zap.Uint64("targetCode", targetDeviceCode),
		zap.String("targetUUID", targetUUID))
	return nil
}

// SendMouseMove 发送鼠标移动事件
func SendMouseMove(targetUUID string, x, y int32) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send mouse move")
		return nil
	}

	reqData := &pb.MouseMoveData{
		X:         x,
		Y:         y,
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal mouse move error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "mouse_move",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send mouse move error.", zap.Error(err))
		return err
	}

	logger.Debug("[sys] mouse move sent.", zap.String("targetUUID", targetUUID), zap.Int32("x", x), zap.Int32("y", y))
	return nil
}

// SendMouseClick 发送鼠标点击事件
func SendMouseClick(targetUUID string, x, y, button int32, action string) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send mouse click")
		return nil
	}

	reqData := &pb.MouseClickData{
		X:         x,
		Y:         y,
		Button:    button,
		Action:    action,
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal mouse click error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "mouse_click",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send mouse click error.", zap.Error(err))
		return err
	}

	logger.Debug("[sys] mouse click sent.", zap.String("targetUUID", targetUUID), zap.Int32("x", x), zap.Int32("y", y), zap.String("action", action))
	return nil
}

// SendMouseScroll 发送鼠标滚轮事件
func SendMouseScroll(targetUUID string, x, y int32, deltaX, deltaY float64) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send mouse scroll")
		return nil
	}

	reqData := &pb.MouseScrollData{
		X:         x,
		Y:         y,
		DeltaY:    int32(deltaY),
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal mouse scroll error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "mouse_scroll",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send mouse scroll error.", zap.Error(err))
		return err
	}

	logger.Debug("[sys] mouse scroll sent.", zap.String("targetUUID", targetUUID), zap.Int32("deltaX", int32(deltaX)), zap.Int32("deltaY", int32(deltaY)))
	return nil
}

// SendKeyDown 发送键盘按下事件
func SendKeyDown(targetUUID string, key string, modifiers []string) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send key down")
		return nil
	}

	reqData := &pb.KeyDownData{
		Key:       key,
		Modifiers: modifiers,
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal key down error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "key_down",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send key down error.", zap.Error(err))
		return err
	}

	logger.Debug("[sys] key down sent.", zap.String("targetUUID", targetUUID), zap.String("key", key), zap.Strings("modifiers", modifiers))
	return nil
}

// SendKeyUp 发送键盘释放事件
func SendKeyUp(targetUUID string, key string, modifiers []string) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send key up")
		return nil
	}

	reqData := &pb.KeyUpData{
		Key:       key,
		Modifiers: modifiers,
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal key up error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "key_up",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send key up error.", zap.Error(err))
		return err
	}

	logger.Debug("[sys] key up sent.", zap.String("targetUUID", targetUUID), zap.String("key", key), zap.Strings("modifiers", modifiers))
	return nil
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
