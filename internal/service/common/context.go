package common

import (
	"context"
	"go.uber.org/zap"
	"godesk-client/internal/logger"
	"google.golang.org/grpc/metadata"
)

// WithAuthorization 添加认证信息
func WithAuthorization(ctx context.Context) context.Context {
	sysConfig, err := GetSysConfig()
	if err != nil || sysConfig.Token == "" {
		logger.Error("[sys] read sys config error.", zap.Error(err))
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", sysConfig.Token)
}
