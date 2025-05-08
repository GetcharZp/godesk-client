package device

import (
	"context"
	pb "godesk-client/proto"
)

type Service struct {
}

var (
	ctx    context.Context
	client pb.DeviceServiceClient
)

type Info struct {
	Uuid     string `json:"uuid"`     // UUID
	Code     uint64 `json:"code"`     // 设备码
	Password string `json:"password"` // 密码
	Os       string `json:"os"`       // 操作系统, win, mac, linux
}
