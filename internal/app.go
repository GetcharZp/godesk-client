package internal

import (
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/channel"
	"godesk-client/internal/service/common"
	"godesk-client/internal/service/control"
	"godesk-client/internal/service/device"
	"godesk-client/internal/service/session"
	"godesk-client/internal/service/user"
	pb "godesk-client/proto"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var rpcClientOnce sync.Once

func NewService() {
	// 初始化日志
	logger.NewLogger()
	// 初始化配置

	// 加载会话
	session.LoadSessions()

	// 初始化 RPC 客户端
	rpcClientOnce.Do(newRpcClient)
	// 重连监视
	go handleReconnect()
}

func newRpcClient() {
	// 从配置中获取服务地址
	sysConfig, err := common.GetSysConfig()
	if err != nil {
		logger.Error("[sys] get sys config error.", zap.Error(err))
		return
	}

	define.GrpcConn, err = grpc.NewClient(sysConfig.ServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("[sys] new client connection error.", zap.Error(err))
		return
	}

	(&device.Service{}).ClientInit()
	(&user.Service{}).ClientInit()
	(&channel.Service{}).ClientInit(pb.NewChannelServiceClient(define.GrpcConn))
	(&control.Service{}).ClientInit(pb.NewChannelServiceClient(define.GrpcConn))
}

func handleReconnect() {
	for {
		if define.GrpcConn == nil || define.GrpcConn.GetState() != connectivity.Ready {
			if define.GrpcConn != nil {
				define.GrpcConn.Close()
			}
			logger.Info("[sys] reconnect to server.")
			newRpcClient()
		}
		time.Sleep(3 * time.Second)
	}
}
