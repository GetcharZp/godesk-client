package cache

import (
	"godesk-client/internal/service/models"
	pb "godesk-client/proto"
	"sync"
	"time"
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

// ========== 远程文件列表缓存 ==========

type remoteFileListCacheItem struct {
	data      pb.FileListResponseData
	timestamp time.Time
}

var (
	remoteFileListCache      = make(map[string]*remoteFileListCacheItem)
	remoteFileListCacheMutex sync.Mutex
)

func SetRemoteFileList(targetUUID string, path string, data pb.FileListResponseData) {
	remoteFileListCacheMutex.Lock()
	defer remoteFileListCacheMutex.Unlock()
	key := targetUUID + ":" + path
	remoteFileListCache[key] = &remoteFileListCacheItem{
		data:      data,
		timestamp: time.Now(),
	}
}

func GetRemoteFileList(targetUUID string, path string) (pb.FileListResponseData, bool) {
	remoteFileListCacheMutex.Lock()
	defer remoteFileListCacheMutex.Unlock()
	key := targetUUID + ":" + path
	if item, ok := remoteFileListCache[key]; ok {
		if time.Since(item.timestamp) < 10*time.Second {
			return item.data, true
		}
		delete(remoteFileListCache, key)
	}
	return pb.FileListResponseData{}, false
}
