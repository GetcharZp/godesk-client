package common

import (
	"godesk-client/internal/service/cache"
	"godesk-client/internal/service/models"
)

func UpdateSysConfig(sysConfig *models.SysConfig) error {
	if err := sysConfig.Updates(); err != nil {
		return err
	}
	cache.ClearSysConfig()
	return nil
}

func UpdateSysConfigMap(mapData map[string]any) error {
	if err := (&models.SysConfig{}).UpdatesMap(mapData); err != nil {
		return err
	}
	cache.ClearSysConfig()
	return nil
}
