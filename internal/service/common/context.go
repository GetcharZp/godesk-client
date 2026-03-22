package common

import (
	"context"
	"godesk-client/internal/service/cache"

	"google.golang.org/grpc/metadata"
)

// WithAuthorization 添加认证信息
func WithAuthorization(ctx context.Context) context.Context {
	sysConfig := cache.GetSysConfig()

	// 添加 authorization token
	if sysConfig.Token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", sysConfig.Token)
	}

	// 添加 access_token
	if sysConfig.AccessToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "accesstoken", sysConfig.AccessToken)
	}

	return ctx
}
