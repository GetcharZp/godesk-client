package cache

import (
	"godesk-client/internal/service/models"
	pb "godesk-client/proto"
	"os"
	"path/filepath"
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

// ========== 文件传输缓存 ==========

type fileTransferCache struct {
	info         *pb.FileTransferStartData
	tempFile     *os.File
	tempPath     string
	receivedSize int64
	totalSize    int64
	complete     bool
	errMsg       string
}

var (
	fileTransferCaches      = make(map[string]*fileTransferCache)
	fileTransferCachesMutex sync.Mutex
)

func SetFileTransfer(transferId string, data pb.FileTransferStartData) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	fileTransferCaches[transferId] = &fileTransferCache{
		info:      &data,
		totalSize: data.TotalSize,
	}
}

// InitUploadProgress 初始化上传进度（发送端调用）
func InitUploadProgress(transferId string, totalSize int64) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	fileTransferCaches[transferId] = &fileTransferCache{
		info:      &pb.FileTransferStartData{TransferId: transferId},
		totalSize: totalSize,
	}
}

// InitDownloadProgress 初始化下载进度（接收端调用）
func InitDownloadProgress(transferId string, targetPath string) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	fileTransferCaches[transferId] = &fileTransferCache{
		info:      &pb.FileTransferStartData{TransferId: transferId, TargetPath: targetPath},
		totalSize: 0,
	}
}

// UpdateUploadProgress 更新上传进度（发送端调用）
func UpdateUploadProgress(transferId string, sentSize int64) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	if cache, ok := fileTransferCaches[transferId]; ok {
		cache.receivedSize = sentSize
	}
}

// UpdateDownloadTotal 更新下载总大小
func UpdateDownloadTotal(transferId string, total int64) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	if cache, ok := fileTransferCaches[transferId]; ok {
		cache.totalSize = total
	}
}

func GetFileTransfer(transferId string) *pb.FileTransferStartData {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	if cache, ok := fileTransferCaches[transferId]; ok {
		return cache.info
	}
	return nil
}

func InitFileTransferTempFile(transferId string) error {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()

	cache, ok := fileTransferCaches[transferId]
	if !ok {
		return nil
	}

	dir := filepath.Dir(cache.info.TargetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tempPath := cache.info.TargetPath + ".tmp"
	f, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	cache.tempFile = f
	cache.tempPath = tempPath
	return nil
}

func WriteFileTransferChunk(transferId string, data []byte) error {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()

	cache, ok := fileTransferCaches[transferId]
	if !ok || cache.tempFile == nil {
		return nil
	}

	n, err := cache.tempFile.Write(data)
	if err != nil {
		return err
	}
	cache.receivedSize += int64(n)
	// 强制刷新到磁盘，确保文件大小实时更新
	cache.tempFile.Sync()
	return nil
}

func GetFileTransferProgress(transferId string) (int64, int64) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	if cache, ok := fileTransferCaches[transferId]; ok {
		return cache.receivedSize, cache.totalSize
	}
	return 0, 0
}

func SetFileTransferComplete(transferId string, success bool, errMsg string) error {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()

	cache, ok := fileTransferCaches[transferId]
	if !ok {
		return nil
	}

	cache.complete = true
	if !success {
		cache.errMsg = errMsg
		if cache.tempFile != nil {
			cache.tempFile.Close()
			os.Remove(cache.tempPath)
		}
		return nil
	}

	if cache.tempFile != nil {
		cache.tempFile.Close()
		if err := os.Rename(cache.tempPath, cache.info.TargetPath); err != nil {
			cache.errMsg = err.Error()
			return err
		}
	}
	return nil
}

func GetFileTransferStatus(transferId string) (bool, bool, string, int64, int64) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	if cache, ok := fileTransferCaches[transferId]; ok {
		return true, cache.complete, cache.errMsg, cache.receivedSize, cache.totalSize
	}
	return false, false, "", 0, 0
}

func ClearFileTransfer(transferId string) {
	fileTransferCachesMutex.Lock()
	defer fileTransferCachesMutex.Unlock()
	if cache, ok := fileTransferCaches[transferId]; ok {
		if cache.tempFile != nil {
			cache.tempFile.Close()
			os.Remove(cache.tempPath)
		}
	}
	delete(fileTransferCaches, transferId)
}

// ========== 文件重命名结果缓存 ==========

var (
	fileRenameResults      = make(map[string]*FileRenameResult)
	fileRenameResultsMutex sync.Mutex
)

type FileRenameResult struct {
	Code    int32
	Message string
	NewPath string
}

func SetFileRenameResult(requestId string, code int32, message string, newPath string) {
	fileRenameResultsMutex.Lock()
	defer fileRenameResultsMutex.Unlock()
	fileRenameResults[requestId] = &FileRenameResult{
		Code:    code,
		Message: message,
		NewPath: newPath,
	}
}

func GetFileRenameResult(requestId string) *FileRenameResult {
	fileRenameResultsMutex.Lock()
	defer fileRenameResultsMutex.Unlock()
	if result, ok := fileRenameResults[requestId]; ok {
		return result
	}
	return nil
}

func ClearFileRenameResult(requestId string) {
	fileRenameResultsMutex.Lock()
	defer fileRenameResultsMutex.Unlock()
	delete(fileRenameResults, requestId)
}

var (
	fileDeleteResults      = make(map[string]*FileDeleteResult)
	fileDeleteResultsMutex sync.Mutex
)

type FileDeleteResult struct {
	Code        int32
	Message     string
	DeletedPath string
}

func SetFileDeleteResult(requestId string, code int32, message string, deletedPath string) {
	fileDeleteResultsMutex.Lock()
	defer fileDeleteResultsMutex.Unlock()
	fileDeleteResults[requestId] = &FileDeleteResult{
		Code:        code,
		Message:     message,
		DeletedPath: deletedPath,
	}
}

func GetFileDeleteResult(requestId string) *FileDeleteResult {
	fileDeleteResultsMutex.Lock()
	defer fileDeleteResultsMutex.Unlock()
	if result, ok := fileDeleteResults[requestId]; ok {
		return result
	}
	return nil
}

func ClearFileDeleteResult(requestId string) {
	fileDeleteResultsMutex.Lock()
	defer fileDeleteResultsMutex.Unlock()
	delete(fileDeleteResults, requestId)
}
