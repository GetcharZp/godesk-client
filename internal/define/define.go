package define

import (
	"google.golang.org/grpc"
	"os"
	"path/filepath"
)

const DefaultAppName = "godesk-client"

var (
	DateTimeMilli = "2006-01-02 15:04:05.000"
	GrpcConn      *grpc.ClientConn
)

func AppDataDir() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", DefaultAppName)
	}
	return filepath.Join(homePath, DefaultAppName)
}

func ConfigDBPath() string {
	return filepath.Join(AppDataDir(), "config.db")
}

func DefaultLogPath() string {
	return filepath.Join(AppDataDir(), "logs", "godesk.log")
}

var DefaultConfig = Config{
	AppName:        DefaultAppName,
	LogPath:        DefaultLogPath(),
	ServiceAddress: "127.0.0.1:9620",
}
