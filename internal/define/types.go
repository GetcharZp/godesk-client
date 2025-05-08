package define

type Config struct {
	AppName        string `json:"app_name"`        // 应用名称
	LogPath        string `json:"log_path"`        // 日志路径
	ServiceAddress string `json:"service_address"` // 服务地址
}

type SysConfig struct {
	Uuid     string `json:"uuid"`     // 设备唯一标识
	Password string `json:"password"` // 设备密码
}
