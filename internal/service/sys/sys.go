package sys

import (
	"godesk-client/internal/define"
	"godesk-client/internal/service/channel"
)

func (a *Service) GetConnectionStatus() bool {
	isConnected := false
	if define.GrpcConn != nil {
		isConnected = define.GrpcConn.GetState().String() == "READY" && channel.IsStreamConnected()
	}
	return isConnected
}
