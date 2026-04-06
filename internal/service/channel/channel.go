package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/cache"
	"godesk-client/internal/service/file"
	"godesk-client/internal/service/screen"
	"godesk-client/internal/service/session"
	"godesk-client/internal/utils"
	pb "godesk-client/proto"
	"path/filepath"
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

// GetMyUUID 获取本机UUID
func GetMyUUID() string {
	return myUUID
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
	logger.Info("[sys] received message.", zap.String("key", req.Key), zap.String("from", req.SendClientUuid), zap.String("to", req.TargetClientUuid))

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
	case "file_list_request":
		in.handleFileListRequest(req)
	case "file_list_response":
		in.handleFileListResponse(req)
	case "file_rename_request":
		in.handleFileRenameRequest(req)
	case "file_rename_response":
		in.handleFileRenameResponse(req)
	case "file_transfer_start":
		in.handleFileTransferStart(req)
	case "file_transfer_data":
		in.handleFileTransferData(req)
	case "file_transfer_complete":
		in.handleFileTransferComplete(req)
	case "file_transfer_cancel":
		in.handleFileTransferCancel(req)
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
		in.sendControlStartedResponse(req.SendClientUuid, 2, "", 0)
		return
	}

	// 只有请求控制模式才启动屏幕捕获
	if data.RequestControl {
		manager := screen.GetScreenManager()
		manager.StartCapture(func(frame *screen.FrameData) {
			in.sendScreenStreamData(req.SendClientUuid, frame)
		})
		logger.Info("[sys] screen capture started for remote control.")
	} else {
		logger.Info("[sys] file access only, no screen capture.")
	}

	// 发送接受响应，带上 targetCode 以便控制端查找会话
	in.sendControlStartedResponse(req.SendClientUuid, 0, sysConfig.Uuid, data.TargetCode)
}

// sendControlStartedResponse 发送控制开始响应
func (in *Service) sendControlStartedResponse(targetUUID string, code int32, uuid string, targetCode uint64) {
	resp := &pb.ControlStartedResponseData{
		Code:       code,
		Uuid:       uuid,
		TargetCode: targetCode,
	}

	data, _ := json.Marshal(resp)
	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "control_started_response",
		Data:             data,
	}

	in.SendMessage(req)
	logger.Info("[sys] control started response sent.", zap.Int32("code", code), zap.Uint64("targetCode", targetCode))
}

// handleControlStartedResponse 处理控制开始响应（控制端收到）
func (in *Service) handleControlStartedResponse(req *pb.ChannelRequest) {
	var data pb.ControlStartedResponseData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal control started response error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received control started response.", zap.Int32("code", data.Code), zap.String("uuid", data.Uuid), zap.Uint64("targetCode", data.TargetCode))

	if data.Code == 0 {
		// 控制开始成功，更新会话状态
		// 先查找控制类型的会话，如果没有再查找文件类型
		sess := session.GetSessionByDeviceCodeAndType(data.TargetCode, "control")
		if sess == nil {
			sess = session.GetSessionByDeviceCodeAndType(data.TargetCode, "file")
		}
		if sess != nil {
			sess.TargetUUID = data.Uuid
			sess.Status = "connected"
			logger.Info("[sys] session updated with target UUID.", zap.String("sessionId", sess.SessionId), zap.String("targetUUID", data.Uuid), zap.String("sessionType", sess.SessionType))
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

	// 鼠标移动
	utils.SetCursorPosAbsolute(int(data.X), int(data.Y))
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

	// 鼠标移动
	utils.SetCursorPosAbsolute(int(data.X), int(data.Y))

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

// ========== 文件管理相关方法 ==========

// SendFileListRequest 发送文件列表请求（控制端调用）
func SendFileListRequest(targetUUID string, targetCode uint64, path string) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send file list request")
		return fmt.Errorf("stream is nil")
	}

	// 如果是本地请求（同一台机器测试），直接本地处理
	if targetUUID == myUUID {
		logger.Info("[sys] file list request is local, handling directly.", zap.String("path", path))
		files, err := file.ListFiles(path)
		respCode := int32(0)
		respMsg := "success"
		if err != nil {
			respCode = 3
			respMsg = err.Error()
		}

		var pbFiles []*pb.FileInfo
		for _, f := range files {
			pbFiles = append(pbFiles, &pb.FileInfo{
				Name:       f.Name,
				Path:       f.Path,
				Size:       f.Size,
				IsDir:      f.IsDir,
				ModifyTime: f.ModifyTime,
				Mode:       f.Mode,
			})
		}

		resp := &pb.FileListResponseData{
			Code:        respCode,
			Message:     respMsg,
			CurrentPath: path,
			Files:       pbFiles,
			Timestamp:   time.Now().UnixMilli(),
		}

		cache.SetRemoteFileList(myUUID, path, *resp)
		logger.Info("[sys] local file list handled.", zap.Int("fileCount", len(pbFiles)))
		return nil
	}

	reqData := &pb.FileListRequestData{
		TargetCode: targetCode,
		Path:       path,
		Timestamp:  time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal file list request error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "file_list_request",
		Data:             data,
	}

	logger.Info("[sys] sending file list request.", zap.String("targetUUID", targetUUID), zap.Uint64("targetCode", targetCode), zap.String("path", path), zap.Int("dataSize", len(data)))

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send file list request error.", zap.Error(err))
		return err
	}

	logger.Info("[sys] file list request sent.", zap.String("targetUUID", targetUUID), zap.Uint64("targetCode", targetCode), zap.String("path", path))
	return nil
}

// SendFileRenameRequest 发送文件重命名请求
func SendFileRenameRequest(targetUUID string, requestId string, oldPath string, newName string) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send file rename request")
		return fmt.Errorf("stream is nil")
	}

	reqData := &pb.FileRenameRequestData{
		RequestId: requestId,
		OldPath:   oldPath,
		NewName:   newName,
		Timestamp: time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal file rename request error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "file_rename_request",
		Data:             data,
	}

	logger.Info("[sys] sending file rename request.", zap.String("targetUUID", targetUUID), zap.String("requestId", requestId), zap.String("oldPath", oldPath), zap.String("newName", newName))

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send file rename request error.", zap.Error(err))
		return err
	}

	logger.Info("[sys] file rename request sent.", zap.String("targetUUID", targetUUID), zap.String("requestId", requestId))
	return nil
}

// handleFileListRequest 处理文件列表请求（被控端收到）
func (in *Service) handleFileListRequest(req *pb.ChannelRequest) {
	var data pb.FileListRequestData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file list request error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received file list request.", zap.String("path", data.Path))

	files, err := file.ListFiles(data.Path)
	respCode := int32(0)
	respMsg := "success"
	if err != nil {
		respCode = 3
		respMsg = err.Error()
	}

	var pbFiles []*pb.FileInfo
	for _, f := range files {
		pbFiles = append(pbFiles, &pb.FileInfo{
			Name:       f.Name,
			Path:       f.Path,
			Size:       f.Size,
			IsDir:      f.IsDir,
			ModifyTime: f.ModifyTime,
			Mode:       f.Mode,
		})
	}

	resp := &pb.FileListResponseData{
		Code:        respCode,
		Message:     respMsg,
		CurrentPath: data.Path,
		Files:       pbFiles,
		Timestamp:   time.Now().UnixMilli(),
	}

	respData, _ := json.Marshal(resp)
	in.SendMessage(&pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: req.SendClientUuid,
		Key:              "file_list_response",
		Data:             respData,
	})
	logger.Info("[sys] file list response sent.", zap.Int("fileCount", len(pbFiles)))
}

// handleFileListResponse 处理文件列表响应（控制端收到）
func (in *Service) handleFileListResponse(req *pb.ChannelRequest) {
	var data pb.FileListResponseData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file list response error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received file list response.", zap.Int32("code", data.Code), zap.Int("fileCount", len(data.Files)), zap.String("currentPath", data.CurrentPath), zap.String("fromUUID", req.SendClientUuid))
	cache.SetRemoteFileList(req.SendClientUuid, data.CurrentPath, data)
	logger.Info("[sys] remote file list cached.", zap.String("cacheKey", req.SendClientUuid+":"+data.CurrentPath))
}

// handleFileRenameRequest 处理文件重命名请求（被控端收到）
func (in *Service) handleFileRenameRequest(req *pb.ChannelRequest) {
	var data pb.FileRenameRequestData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file rename request error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received file rename request.", zap.String("requestId", data.RequestId), zap.String("oldPath", data.OldPath), zap.String("newName", data.NewName))

	respCode := int32(0)
	respMsg := "success"
	var newPath string

	err := file.RenameFile(data.OldPath, data.NewName)
	if err != nil {
		respCode = 4
		respMsg = err.Error()
		logger.Error("[sys] rename file error.", zap.Error(err))
	} else {
		dir := filepath.Dir(data.OldPath)
		newPath = filepath.Clean(filepath.Join(dir, data.NewName))
	}

	respData := &pb.FileRenameResponseData{
		RequestId: data.RequestId,
		Code:      respCode,
		Message:   respMsg,
		NewPath:   newPath,
		Timestamp: time.Now().UnixMilli(),
	}

	jsonData, err := json.Marshal(respData)
	if err != nil {
		logger.Error("[sys] marshal file rename response error.", zap.Error(err))
		return
	}

	resp := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: req.SendClientUuid,
		Key:              "file_rename_response",
		Data:             jsonData,
	}

	if err := stream.Send(resp); err != nil {
		logger.Error("[sys] send file rename response error.", zap.Error(err))
		return
	}

	logger.Info("[sys] file rename response sent.", zap.String("requestId", data.RequestId), zap.Int32("code", respCode), zap.String("newPath", newPath))
}

// handleFileRenameResponse 处理文件重命名响应（控制端收到）
func (in *Service) handleFileRenameResponse(req *pb.ChannelRequest) {
	var data pb.FileRenameResponseData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file rename response error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received file rename response.", zap.String("requestId", data.RequestId), zap.Int32("code", data.Code), zap.String("message", data.Message), zap.String("newPath", data.NewPath))

	cache.SetFileRenameResult(data.RequestId, data.Code, data.Message, data.NewPath)
}

// ========== 文件传输相关方法 ==========

// handleFileTransferStart 处理文件传输开始
func (in *Service) handleFileTransferStart(req *pb.ChannelRequest) {
	var data pb.FileTransferStartData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file transfer start error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received file transfer start.",
		zap.String("transferId", data.TransferId),
		zap.String("direction", data.Direction),
		zap.String("sourcePath", data.SourcePath),
		zap.String("targetPath", data.TargetPath),
		zap.Int64("totalSize", data.TotalSize))

	// 保存传输信息到缓存
	cache.SetFileTransfer(data.TransferId, data)

	// 如果是下载请求（被控端收到），读取文件并发送数据
	if data.Direction == "download" {
		go in.sendFileData(req.SendClientUuid, &data)
	}
	// 上传请求不需要响应，等待接收数据即可
}

// sendFileData 发送文件数据（被控端调用，响应下载请求）
func (in *Service) sendFileData(targetUUID string, transferInfo *pb.FileTransferStartData) {
	fileData, err := file.ReadFile(transferInfo.SourcePath)
	if err != nil {
		logger.Error("[sys] read file for download error.", zap.Error(err))
		in.sendFileTransferComplete(targetUUID, transferInfo.TransferId, 1, err.Error())
		return
	}

	totalSize := int64(len(fileData))
	chunkSize := transferInfo.ChunkSize
	if chunkSize == 0 {
		chunkSize = 64 * 1024
	}

	totalChunks := int32(totalSize / int64(chunkSize))
	if totalSize%int64(chunkSize) != 0 {
		totalChunks++
	}

	logger.Info("[sys] sending file data for download.",
		zap.String("transferId", transferInfo.TransferId),
		zap.Int64("totalSize", totalSize),
		zap.Int32("totalChunks", totalChunks))

	for i := int32(0); i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > int32(len(fileData)) {
			end = int32(len(fileData))
		}

		chunk := fileData[start:end]
		isLast := i == totalChunks-1

		// 第一个数据块时发送文件总大小
		var sendTotalSize int64
		if i == 0 {
			sendTotalSize = totalSize
		}

		if err := SendFileTransferData(targetUUID, transferInfo.TransferId, i, chunk, isLast, sendTotalSize); err != nil {
			logger.Error("[sys] send file data chunk error.", zap.Error(err))
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	logger.Info("[sys] file data send complete.", zap.String("transferId", transferInfo.TransferId))
}

// handleFileTransferData 处理文件传输数据块
func (in *Service) handleFileTransferData(req *pb.ChannelRequest) {
	var data pb.FileTransferData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file transfer data error.", zap.Error(err))
		return
	}

	transferInfo := cache.GetFileTransfer(data.TransferId)
	if transferInfo == nil {
		logger.Error("[sys] transfer not found.", zap.String("transferId", data.TransferId))
		return
	}

	// 第一个数据块时创建临时文件
	if data.ChunkIndex == 0 {
		// 如果有文件总大小，更新缓存
		if data.TotalSize > 0 {
			cache.UpdateDownloadTotal(data.TransferId, data.TotalSize)
		}

		if err := cache.InitFileTransferTempFile(data.TransferId); err != nil {
			logger.Error("[sys] create temp file error.", zap.Error(err))
			in.sendFileTransferComplete(req.SendClientUuid, data.TransferId, 1, err.Error())
			cache.ClearFileTransfer(data.TransferId)
			return
		}
	}

	// 写入数据块到临时文件
	if err := cache.WriteFileTransferChunk(data.TransferId, data.Data); err != nil {
		logger.Error("[sys] write file chunk error.", zap.Error(err))
		in.sendFileTransferComplete(req.SendClientUuid, data.TransferId, 1, err.Error())
		cache.ClearFileTransfer(data.TransferId)
		return
	}

	// 如果是最后一块，更新总大小（兼容没有 TotalSize 字段的情况）
	if data.IsLast {
		received, _ := cache.GetFileTransferProgress(data.TransferId)
		cache.UpdateDownloadTotal(data.TransferId, received)
	}

	received, total := cache.GetFileTransferProgress(data.TransferId)
	logger.Debug("[sys] received file transfer data.",
		zap.String("transferId", data.TransferId),
		zap.Int32("chunkIndex", data.ChunkIndex),
		zap.Bool("isLast", data.IsLast),
		zap.Int("dataSize", len(data.Data)),
		zap.Int64("received", received),
		zap.Int64("total", total))

	// 如果是最后一块，完成传输
	if data.IsLast {
		if err := cache.SetFileTransferComplete(data.TransferId, true, ""); err != nil {
			logger.Error("[sys] finalize file transfer error.", zap.Error(err))
			in.sendFileTransferComplete(req.SendClientUuid, data.TransferId, 1, err.Error())
		} else {
			logger.Info("[sys] file transfer complete.", zap.String("transferId", data.TransferId), zap.String("targetPath", transferInfo.TargetPath))
			in.sendFileTransferComplete(req.SendClientUuid, data.TransferId, 0, "success")
		}
	}
}

// handleFileTransferComplete 处理文件传输完成
func (in *Service) handleFileTransferComplete(req *pb.ChannelRequest) {
	var data pb.FileTransferCompleteData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file transfer complete error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received file transfer complete.",
		zap.String("transferId", data.TransferId),
		zap.Int32("code", data.Code),
		zap.String("message", data.Message))

	// 更新传输状态
	cache.SetFileTransferComplete(data.TransferId, data.Code == 0, data.Message)
}

// handleFileTransferCancel 处理文件传输取消
func (in *Service) handleFileTransferCancel(req *pb.ChannelRequest) {
	var data pb.FileTransferCancelData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[sys] unmarshal file transfer cancel error.", zap.Error(err))
		return
	}

	logger.Info("[sys] received file transfer cancel.",
		zap.String("transferId", data.TransferId),
		zap.String("reason", data.Reason))

	// 清理传输缓存
	cache.ClearFileTransfer(data.TransferId)
}

// sendFileTransferComplete 发送文件传输完成响应
func (in *Service) sendFileTransferComplete(targetUUID, transferId string, code int32, msg string) {
	resp := &pb.FileTransferCompleteData{
		TransferId: transferId,
		Code:       code,
		Message:    msg,
		Timestamp:  time.Now().UnixMilli(),
	}
	respData, _ := json.Marshal(resp)
	in.SendMessage(&pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "file_transfer_complete",
		Data:             respData,
	})
}

// SendFileTransferStart 发送文件传输开始请求
func SendFileTransferStart(targetUUID string, transferId, direction, sourcePath, targetPath string, totalSize int64, chunkSize int32) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send file transfer start")
		return fmt.Errorf("stream is nil")
	}

	reqData := &pb.FileTransferStartData{
		TransferId: transferId,
		Direction:  direction,
		SourcePath: sourcePath,
		TargetPath: targetPath,
		TotalSize:  totalSize,
		ChunkSize:  chunkSize,
		Timestamp:  time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal file transfer start error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "file_transfer_start",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send file transfer start error.", zap.Error(err))
		return err
	}

	logger.Info("[sys] file transfer start sent.", zap.String("transferId", transferId), zap.String("direction", direction))
	return nil
}

// SendFileTransferData 发送文件传输数据块
func SendFileTransferData(targetUUID string, transferId string, chunkIndex int32, data []byte, isLast bool, totalSize int64) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send file transfer data")
		return fmt.Errorf("stream is nil")
	}

	reqData := &pb.FileTransferData{
		TransferId: transferId,
		ChunkIndex: chunkIndex,
		IsLast:     isLast,
		Data:       data,
		DataSize:   int32(len(data)),
		Timestamp:  time.Now().UnixMilli(),
		TotalSize:  totalSize,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal file transfer data error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "file_transfer_data",
		Data:             jsonData,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send file transfer data error.", zap.Error(err))
		return err
	}

	return nil
}

// SendFileTransferCancel 发送文件传输取消
func SendFileTransferCancel(targetUUID string, transferId string, reason string) error {
	if stream == nil {
		logger.Error("[sys] stream is nil, cannot send file transfer cancel")
		return fmt.Errorf("stream is nil")
	}

	reqData := &pb.FileTransferCancelData{
		TransferId: transferId,
		Reason:     reason,
		Timestamp:  time.Now().UnixMilli(),
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		logger.Error("[sys] marshal file transfer cancel error.", zap.Error(err))
		return err
	}

	req := &pb.ChannelRequest{
		SendClientUuid:   myUUID,
		TargetClientUuid: targetUUID,
		Key:              "file_transfer_cancel",
		Data:             data,
	}

	if err := stream.Send(req); err != nil {
		logger.Error("[sys] send file transfer cancel error.", zap.Error(err))
		return err
	}

	logger.Info("[sys] file transfer cancel sent.", zap.String("transferId", transferId))
	return nil
}
