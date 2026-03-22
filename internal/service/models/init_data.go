package models

import (
	"godesk-client/internal/define"
)

func initData() {
	// 初始化系统配置
	initSysConfig()
}

func initSysConfig() {
	sysConfig := &SysConfig{
		ID:             defaultSysConfigID,
		ServiceAddress: define.DefaultConfig.ServiceAddress,
	}
	if err := DB.FirstOrCreate(sysConfig, SysConfig{ID: defaultSysConfigID}).Error; err != nil {
		panic(err)
	}
}
