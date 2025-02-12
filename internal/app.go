package internal

import (
	"go.uber.org/zap"
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/channel"
	pb "godesk-client/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

var rpcClientOnce sync.Once

func NewService() {
	// 初始化日志
	logger.NewLogger()
	// 初始化配置

	// 初始化 RPC 客户端
	rpcClientOnce.Do(newRpcClient)
	// 重连监视
	go handleReconnect()
}

func newRpcClient() {
	var err error
	define.GrpcConn, err = grpc.NewClient(define.DefaultConfig.ServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("[sys] new client connection error.", zap.Error(err))
		return
	}

	(&channel.Service{}).ClientInit(pb.NewChannelServiceClient(define.GrpcConn))
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
