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
