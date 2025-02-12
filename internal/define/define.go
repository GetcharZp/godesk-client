package define

import "google.golang.org/grpc"

var (
	DateTimeMilli = "2006-01-02 15:04:05.000"
	GrpcConn      *grpc.ClientConn
)

var DefaultConfig = Config{
	AppName:        "godesk-client",
	LogPath:        "/var/log/godesk-client/godesk.log",
	ServiceAddress: "127.0.0.1:9620",
}
