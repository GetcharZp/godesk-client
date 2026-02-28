package internal

import (
	"godesk-client/internal/define"
	"godesk-client/internal/logger"
)

// Reconnect 重新连接
func Reconnect() {
	logger.Info("[sys] manual reconnect triggered.")
	if define.GrpcConn != nil {
		define.GrpcConn.Close()
		define.GrpcConn = nil
	}
	newRpcClient()
}
