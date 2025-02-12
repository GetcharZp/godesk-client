package channel

import (
	"context"
	"go.uber.org/zap"
	"godesk-client/internal/logger"
	pb "godesk-client/proto"
)

var stream pb.ChannelService_DataStreamClient

func (in *Service) ClientInit(c pb.ChannelServiceClient) {
	var err error
	stream, err = c.DataStream(context.Background())
	if err != nil {
		logger.Error("[sys] stream init error.", zap.Error(err))
		return
	}

	in.ReceiveDataHandle()
}

func (in *Service) ReceiveDataHandle() {
	for {
		req, err := stream.Recv()
		if err != nil {
			logger.Error("[sys] stream receive error.", zap.Error(err))
			return
		}
		logger.Info("[sys] stream receive message.", zap.Any("data", req))
	}
}

func (in *Service) SendMessage(req *pb.ChannelRequest) {
	if err := stream.Send(req); err != nil {
		logger.Error("[sys] stream send message error.", zap.Error(err))
		return
	}
}
