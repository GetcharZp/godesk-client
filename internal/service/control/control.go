package control

import (
	"context"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/common"
	"godesk-client/internal/service/device"
	"godesk-client/internal/service/session"
	pb "godesk-client/proto"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var (
	client pb.ChannelServiceClient
	ctx    context.Context
)

func (in *Service) ClientInit(c pb.ChannelServiceClient) {
	ctx = common.WithAuthorization(context.Background())
	client = c
}

// SendControlRequest 发送控制请求
func (in *Service) SendControlRequest(targetDeviceCode uint64, targetPassword string, requestControl bool) (*pb.ControlResponse, error) {
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[control] get sys config error.", zap.Error(err))
		return nil, err
	}

	req := &pb.ControlRequest{
		RequestId:        generateRequestId(),
		ControllerUuid:   sysConfig.Uuid,
		ControllerCode:   sysConfig.Code,
		ControllerName:   sysConfig.Username,
		TargetDeviceCode: targetDeviceCode,
		TargetPassword:   targetPassword,
		Timestamp:        time.Now().Unix(),
		RequestControl:   requestControl,
	}

	resp, err := client.SendControlRequest(ctx, req)
	if err != nil {
		logger.Error("[control] send control request error.", zap.Error(err))
		return nil, err
	}
	logger.Info("[control] control send control request.",
		zap.Any("req", req),
		zap.Any("resp", resp))

	// 如果请求被接受，创建会话
	if resp.Accepted {
		deviceName := getDeviceDisplayName(targetDeviceCode)

		logger.Info("[control] creating session",
			zap.String("sessionId", resp.SessionId),
			zap.Uint64("deviceCode", targetDeviceCode),
			zap.String("deviceName", deviceName))

		sess := session.CreateSession(resp.SessionId, targetDeviceCode, deviceName, !requestControl)
		if sess != nil {
			sess.ScreenWidth = resp.TargetInfo.ScreenWidth
			sess.ScreenHeight = resp.TargetInfo.ScreenHeight
			sess.Status = "connected"

			// 启动屏幕流接收
			go in.startScreenStream(sess)
		}

		// 打印当前所有会话
		allSessions := session.GetAllSessions()
		logger.Info("[control] all sessions after create", zap.Int("count", len(allSessions)))
	}

	return resp, nil
}

// getDeviceDisplayName 获取设备显示名称
func getDeviceDisplayName(deviceCode uint64) string {
	devices, err := device.GetDeviceList()
	if err == nil {
		for _, d := range devices {
			if d.Code == deviceCode {
				if d.Remark != "" {
					return d.Remark
				}
				return strconv.FormatUint(deviceCode, 10)
			}
		}
	}
	return strconv.FormatUint(deviceCode, 10)
}

// SendDisconnectNotify 发送断开连接通知
func (in *Service) SendDisconnectNotify(sessionId string, targetDeviceCode uint64) error {
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[control] get sys config error.", zap.Error(err))
		return err
	}

	req := &pb.DisconnectNotify{
		NotifyId:         generateRequestId(),
		SessionId:        sessionId,
		ControllerUuid:   sysConfig.Uuid,
		TargetDeviceCode: targetDeviceCode,
		Reason:           0, // 主动断开
		ReasonDesc:       "user_initiated",
		Timestamp:        time.Now().Unix(),
	}

	_, err = client.SendDisconnectNotify(ctx, req)
	if err != nil {
		logger.Error("[control] send disconnect notify error.", zap.Error(err))
		return err
	}

	// 停止屏幕流
	in.StopScreenStream(sessionId)

	return nil
}

// startScreenStream 启动屏幕流接收
// 注意：当前通过 DataStream 接收屏幕数据，StartScreenStream 服务暂未实现
func (in *Service) startScreenStream(sess *session.Session) {
	logger.Info("[control] screen stream started via DataStream.", zap.String("sessionId", sess.SessionId))
	sess.Status = "connected"

	// 屏幕数据将通过 DataStream 的 screen_stream_data 消息接收
	// 由 channel.handleScreenStreamData 处理
}

// StopScreenStream 停止屏幕流
func (in *Service) StopScreenStream(sessionId string) error {
	sess := session.GetSession(sessionId)
	if sess == nil || sess.StreamClient == nil {
		return nil
	}

	req := &pb.ScreenStreamRequest{
		SessionId: sessionId,
		Action:    "stop",
	}

	return sess.StreamClient.Send(req)
}

// generateRequestId 生成请求ID
func generateRequestId() string {
	return time.Now().Format("20060102150405") + string(rune(time.Now().UnixNano()%1000))
}
