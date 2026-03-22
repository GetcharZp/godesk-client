package device

import (
	"context"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/cache"
	"godesk-client/internal/service/common"
	"godesk-client/internal/service/models"
	pb "godesk-client/proto"
	"runtime"

	"github.com/up-zero/gotool/randomutil"
	"go.uber.org/zap"
)

func (in *Service) ClientInit() {
	client = pb.NewDeviceServiceClient(define.GrpcConn)
}

// Info 设备信息
func (in *Service) Info() (*Info, error) {
	sysConfig := cache.GetSysConfig()
	if sysConfig == nil || sysConfig.Uuid == "" {

		// 配置文件不存在
		password := randomutil.RandomAlphaNumber(8)
		response, err := client.CreateDevice(common.WithAuthorization(context.Background()), &pb.CreateDeviceRequest{
			Os: runtime.GOOS,
		})
		if err != nil {
			logger.Error("[sys] create device error.", zap.Error(err))
			return nil, err
		}
		// 保存配置文件
		if err := (&models.SysConfig{
			Uuid:     response.GetUuid(),
			Password: password,
		}).Updates(); err != nil {
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
	// 获取设备信息
	deviceInfo, err := client.GetDeviceInfo(common.WithAuthorization(context.Background()), &pb.DeviceInfoRequest{
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
	response, err := client.GetDeviceList(common.WithAuthorization(context.Background()), &pb.DeviceListRequest{
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
	_, err := client.AddDevice(common.WithAuthorization(context.Background()), req)
	if err != nil {
		logger.Error("[sys] add device error.", zap.Error(err))
		return err
	}

	return nil
}

// Edit 编辑设备
func (in *Service) Edit(req *pb.EditDeviceRequest) error {
	_, err := client.EditDevice(common.WithAuthorization(context.Background()), req)
	if err != nil {
		logger.Error("[sys] edit device error.", zap.Error(err))
		return err
	}

	return nil
}

// Delete 删除设备
func (in *Service) Delete(req *pb.DeleteDeviceRequest) error {
	_, err := client.DeleteDevice(common.WithAuthorization(context.Background()), req)
	if err != nil {
		logger.Error("[sys] delete device error.", zap.Error(err))
		return err
	}

	return nil
}
