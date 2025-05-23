package device

import (
	"context"
	"errors"
	"github.com/up-zero/gotool/randomutil"
	"go.uber.org/zap"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/common"
	pb "godesk-client/proto"
	"io/fs"
	"runtime"
)

func (in *Service) ClientInit() {
	ctx = context.Background()
	client = pb.NewDeviceServiceClient(define.GrpcConn)
}

// Info 设备信息
func (in *Service) Info() (*Info, error) {
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// 配置文件不存在
			password := randomutil.RandomAlphaNumber(8)
			response, err := client.CreateDevice(ctx, &pb.CreateDeviceRequest{
				Os: runtime.GOOS,
			})
			if err != nil {
				logger.Error("[sys] create device error.", zap.Error(err))
				return nil, err
			}
			// 保存配置文件
			if err := common.SaveSysConfig(&define.SysConfig{
				Uuid:     response.GetUuid(),
				Password: password,
			}); err != nil {
				logger.Error("[sys] save sys config error.", zap.Error(err))
				return nil, err
			}
			// 返回结果
			return &Info{
				Code:     response.GetCode(),
				Os:       runtime.GOOS,
				Password: password,
				Uuid:     response.GetUuid(),
			}, nil
		}
		logger.Error("[sys] read sys config error.", zap.Error(err))
		return nil, err
	}
	// 获取设备信息
	deviceInfo, err := client.GetDeviceInfo(ctx, &pb.DeviceInfoRequest{
		Uuid: sysConfig.Uuid,
	})
	if err != nil {
		logger.Error("[sys] get device info error.", zap.Error(err))
		return nil, err
	}

	return &Info{
		Code:     deviceInfo.GetCode(),
		Os:       deviceInfo.GetOs(),
		Password: sysConfig.Password,
		Uuid:     sysConfig.Uuid,
	}, nil
}
