package user

import (
	"context"
	"errors"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/cache"
	"godesk-client/internal/service/common"
	pb "godesk-client/proto"

	"go.uber.org/zap"
)

func (in *Service) ClientInit() {
	client = pb.NewUserServiceClient(define.GrpcConn)
}

// GetUserInfo 获取用户信息
func (in *Service) GetUserInfo() (*pb.UserInfoResponse, error) {
	reply := &pb.UserInfoResponse{}
	sysConfig := cache.GetSysConfig()

	if sysConfig.Token == "" {
		logger.Error("[sys] user token is empty.")
		return reply, errors.New("user token is empty")
	}
	response, err := client.GetUserInfo(common.WithAuthorization(context.Background()), &pb.EmptyRequest{})
	if err != nil {
		logger.Error("[sys] get user info error.", zap.Error(err))
		return nil, err
	}
	sysConfig.Token = response.Token
	if err := sysConfig.Updates(); err != nil {
		logger.Error("[sys] save sys config error.", zap.Error(err))
		return nil, err
	}
	return response, nil
}

// Login 用户登录
func (in *Service) Login(req *pb.UserLoginRequest) (*pb.UserInfoResponse, error) {
	response, err := client.UserLogin(common.WithAuthorization(context.Background()), req)
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
	response, err := client.UserRegister(common.WithAuthorization(context.Background()), req)
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
	sysConfig := cache.GetSysConfig()
	sysConfig.Username = data.Username
	sysConfig.Token = data.Token
	if err := sysConfig.Updates(); err != nil {
		logger.Error("[sys] save sys config error.", zap.Error(err))
		return err
	}
	cache.ClearSysConfig()
	return nil
}
