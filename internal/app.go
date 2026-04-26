package internal

import (
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
	"godesk-client/internal/service/cache"
	"godesk-client/internal/service/channel"
	"godesk-client/internal/service/device"
	"godesk-client/internal/service/models"
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

var (
	rpcClientOnce sync.Once
	connMux       sync.Mutex
)

func NewService() {
	logger.NewLogger()
	models.InitDB()

	session.LoadSessions()

	rpcClientOnce.Do(initRpcClient)
	go handleReconnect()
}

func initRpcClient() {
	createRpcClient()
}

func createRpcClient() {
	connMux.Lock()
	defer connMux.Unlock()

	sysConfig := cache.GetSysConfig()
	var err error

	define.GrpcConn, err = grpc.NewClient(sysConfig.ServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("[sys] new client connection error.", zap.Error(err))
		return
	}

	(&device.Service{}).ClientInit()
	(&user.Service{}).ClientInit()
	(&channel.Service{}).ClientInit(pb.NewChannelServiceClient(define.GrpcConn))
}

func handleReconnect() {
	for {
		grpcReady := define.GrpcConn != nil && define.GrpcConn.GetState() == connectivity.Ready
		streamReady := channel.IsStreamConnected()

		if !grpcReady || !streamReady {
			if define.GrpcConn != nil {
				define.GrpcConn.Close()
			}
			logger.Info("[sys] reconnect to server.", zap.Bool("grpcReady", grpcReady), zap.Bool("streamReady", streamReady))
			createRpcClient()
		}
		time.Sleep(3 * time.Second)
	}
}
