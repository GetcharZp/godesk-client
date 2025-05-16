package user

import (
	"context"
	pb "godesk-client/proto"
)

type Service struct {
}

var (
	ctx    context.Context
	client pb.UserServiceClient
)
