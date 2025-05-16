package common

import (
	"github.com/up-zero/gotool/fileutil"
	"godesk-client/internal/define"
	"os"
	"path/filepath"
	"sync"
)

var (
	sysConfigMutex sync.Mutex
	homePath, _    = os.UserHomeDir()
	sysConfig      *define.SysConfig
	sysConfigPath  = filepath.Join(homePath, define.DefaultConfig.AppName, "sys_config.json")
)

// GetSysConfig 获取系统配置
func GetSysConfig() (*define.SysConfig, error) {
	sysConfigMutex.Lock()
	defer sysConfigMutex.Unlock()
	if sysConfig != nil {
		return sysConfig, nil
	}
	cfg := new(define.SysConfig)
	if err := fileutil.FileRead(sysConfigPath, cfg); err != nil {
		return nil, err
	}
	sysConfig = cfg
	return sysConfig, nil
}

// SaveSysConfig 保存系统配置
func SaveSysConfig(data *define.SysConfig) error {
	sysConfigMutex.Lock()
	defer sysConfigMutex.Unlock()
	sysConfig = data
	return fileutil.FileSave(sysConfigPath, data)
}
