package sys

import "godesk-client/internal/define"

// GetConnectionStatus 获取连接状态
func (a *Service) GetConnectionStatus() bool {
	isConnected := false
	if define.GrpcConn != nil {
		isConnected = define.GrpcConn.GetState().String() == "READY"
	}
	return isConnected
}
