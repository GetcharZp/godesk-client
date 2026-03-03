package device

import (
	"context"
	"errors"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/common"
	pb "godesk-client/proto"
	"io/fs"
	"runtime"

	"github.com/up-zero/gotool/randomutil"
	"go.uber.org/zap"
)

func (in *Service) ClientInit() {
	ctx = context.Background()
	client = pb.NewDeviceServiceClient(define.GrpcConn)
}

// Info 设备信息
func (in *Service) Info() (*Info, error) {
	sysConfig, err := common.GetSysConfig()
	if (err != nil && errors.Is(err, fs.ErrNotExist)) ||
		sysConfig == nil || sysConfig.Uuid == "" {

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
	} else if err != nil {
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

// List 获取设备列表
func (in *Service) List() ([]*pb.DeviceListItem, error) {
	authCtx := common.WithAuthorization(ctx)

	response, err := client.GetDeviceList(authCtx, &pb.DeviceListRequest{
		Base: &pb.BaseRequest{},
	})
	if err != nil {
		logger.Error("[sys] get device list error.", zap.Error(err))
		return nil, err
	}

	return response.GetList(), nil
}

// GetDeviceList 获取设备列表（供其他包调用）
func GetDeviceList() ([]*pb.DeviceListItem, error) {
	if client == nil {
		(&Service{}).ClientInit()
	}
	return (&Service{}).List()
}

// Add 添加设备
func (in *Service) Add(req *pb.AddDeviceRequest) error {
	authCtx := common.WithAuthorization(ctx)

	_, err := client.AddDevice(authCtx, req)
	if err != nil {
		logger.Error("[sys] add device error.", zap.Error(err))
		return err
	}

	return nil
}

// Edit 编辑设备
func (in *Service) Edit(req *pb.EditDeviceRequest) error {
	authCtx := common.WithAuthorization(ctx)

	_, err := client.EditDevice(authCtx, req)
	if err != nil {
		logger.Error("[sys] edit device error.", zap.Error(err))
		return err
	}

	return nil
}

// Delete 删除设备
func (in *Service) Delete(req *pb.DeleteDeviceRequest) error {
	authCtx := common.WithAuthorization(ctx)

	_, err := client.DeleteDevice(authCtx, req)
	if err != nil {
		logger.Error("[sys] delete device error.", zap.Error(err))
		return err
	}

	return nil
}
