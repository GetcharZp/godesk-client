package models

// SysConfig 系统配置
type SysConfig struct {
	ID             uint   `gorm:"primaryKey"`      // 主键
	Username       string `json:"username"`        // 用户名
	Token          string `json:"token"`           // token
	Uuid           string `json:"uuid"`            // 设备唯一标识
	Password       string `json:"password"`        // 设备密码
	ServiceAddress string `json:"service_address"` // 服务地址
	AccessToken    string `json:"access_token"`    // 访问令牌
	Code           uint64 `json:"code"`            // 设备码
	Sessions       string `json:"sessions"`        // 会话信息（JSON格式）
}

func (s *SysConfig) TableName() string {
	return "sys_config"
}

const (
	defaultSysConfigID = 1
)

func (s *SysConfig) Get() (*SysConfig, error) {
	var sysConfig SysConfig
	err := DB.Where("id = ?", defaultSysConfigID).First(&sysConfig).Error
	if err != nil {
		return nil, err
	}
	return &sysConfig, nil
}

func (s *SysConfig) Updates() error {
	return DB.Model(s).Where("id = ?", defaultSysConfigID).Updates(s).Error
}
