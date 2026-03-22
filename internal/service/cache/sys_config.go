package cache

import (
	"godesk-client/internal/service/models"
	"sync"
)

var (
	sysConfig      *models.SysConfig
	sysConfigMutex sync.Mutex
)

func GetSysConfig() *models.SysConfig {
	sysConfigMutex.Lock()
	defer sysConfigMutex.Unlock()
	if sysConfig == nil {
		sysConfig, _ = (&models.SysConfig{}).Get()
	}
	return sysConfig
}

func ClearSysConfig() {
	sysConfigMutex.Lock()
	defer sysConfigMutex.Unlock()
	sysConfig = nil
}
