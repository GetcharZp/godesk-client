package user

import (
	"context"
	"go.uber.org/zap"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/common"
	pb "godesk-client/proto"
)

func (in *Service) ClientInit() {
	ctx = context.Background()
	client = pb.NewUserServiceClient(define.GrpcConn)
}

// GetUserInfo 获取用户信息
func (in *Service) GetUserInfo() (*pb.UserInfoResponse, error) {
	reply := &pb.UserInfoResponse{}
	sysConfig, err := common.GetSysConfig()
	if err != nil || sysConfig.Token == "" {
		logger.Error("[sys] read sys config error.", zap.Error(err))
		return reply, nil
	}
	response, err := client.GetUserInfo(common.WithAuthorization(ctx), &pb.EmptyRequest{})
	if err != nil {
		logger.Error("[sys] get user info error.", zap.Error(err))
		return nil, err
	}
	sysConfig.Token = response.Token
	if err := common.SaveSysConfig(sysConfig); err != nil {
		logger.Error("[sys] save sys config error.", zap.Error(err))
		return nil, err
	}
	return response, nil
}

// Login 用户登录
func (in *Service) Login(req *pb.UserLoginRequest) (*pb.UserInfoResponse, error) {
	response, err := client.UserLogin(ctx, req)
	if err != nil {
		logger.Error("[sys] user login error.", zap.Error(err))
		return nil, err
	}
	if err := in.updateSysConfig(response); err != nil {
		return nil, err
	}
	return response, nil
}

// Register 用户注册
func (in *Service) Register(req *pb.UserRegisterRequest) (*pb.UserInfoResponse, error) {
	response, err := client.UserRegister(ctx, req)
	if err != nil {
		logger.Error("[sys] user register error.", zap.Error(err))
		return nil, err
	}
	if err := in.updateSysConfig(response); err != nil {
		return nil, err
	}
	return response, nil
}

// Logout 退出登录
func (in *Service) Logout() (any, error) {
	if err := in.updateSysConfig(&pb.UserInfoResponse{}); err != nil {
		return nil, err
	}
	return nil, nil
}

// updateSysConfig 更新系统配置
func (in *Service) updateSysConfig(data *pb.UserInfoResponse) error {
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[sys] read sys config error.", zap.Error(err))
		return err
	}
	sysConfig.Username = data.Username
	sysConfig.Token = data.Token
	if err := common.SaveSysConfig(sysConfig); err != nil {
		logger.Error("[sys] save sys config error.", zap.Error(err))
		return err
	}
	return nil
}
