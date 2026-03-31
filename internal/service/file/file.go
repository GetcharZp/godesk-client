package file

import (
	"godesk-client/internal/logger"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"
)

type Service struct{}

type FileInfo struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	IsDir      bool   `json:"isDir"`
	ModifyTime int64  `json:"modifyTime"`
	Mode       int32  `json:"mode"`
}

func (s *Service) ListLocalFiles(path string) ([]FileInfo, error) {
	return ListFiles(path)
}

func (s *Service) GetLocalDrives() ([]FileInfo, error) {
	return ListDrives()
}

func ListFiles(path string) ([]FileInfo, error) {
	if runtime.GOOS == "windows" && (path == "/" || path == "" || path == "\\") {
		return ListDrives()
	}

	if runtime.GOOS == "windows" {
		if len(path) >= 2 && path[1] == ':' {
			if len(path) == 2 {
				path = path + "\\"
			} else if path[2] == '/' {
				path = path[:2] + "\\"
			}
		} else {
			path = filepath.Clean(path)
		}
	} else {
		path = filepath.Clean(path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		logger.Error("[file] read dir error.", zap.String("path", path), zap.Error(err))
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		files = append(files, FileInfo{
			Name:       entry.Name(),
			Path:       fullPath,
			Size:       info.Size(),
			IsDir:      entry.IsDir(),
			ModifyTime: info.ModTime().UnixMilli(),
			Mode:       int32(info.Mode()),
		})
	}

	return files, nil
}

func ListDrives() ([]FileInfo, error) {
	if runtime.GOOS != "windows" {
		return []FileInfo{{Name: "/", Path: "/", Size: 0, IsDir: true, ModifyTime: time.Now().UnixMilli()}}, nil
	}

	var drives []FileInfo
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drivePath := string(drive) + ":\\"
		if _, err := os.Stat(drivePath); err == nil {
			drives = append(drives, FileInfo{
				Name:       string(drive) + ":",
				Path:       drivePath,
				Size:       0,
				IsDir:      true,
				ModifyTime: time.Now().UnixMilli(),
			})
		}
	}
	return drives, nil
}

func CreateFolder(parentPath, name string) error {
	fullPath := filepath.Join(parentPath, name)
	return os.MkdirAll(fullPath, 0755)
}

func DeleteFile(path string) error {
	return os.RemoveAll(path)
}

func RenameFile(oldPath, newName string) error {
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)
	return os.Rename(oldPath, newPath)
}

func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}

func GetDesktopDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return filepath.Join(home, "Desktop")
}

func GetDownloadsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return filepath.Join(home, "Downloads")
}

func GetDocumentsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return filepath.Join(home, "Documents")
}

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func GetFileInfo(path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Name:       filepath.Base(path),
		Path:       path,
		Size:       info.Size(),
		IsDir:      info.IsDir(),
		ModifyTime: info.ModTime().UnixMilli(),
		Mode:       int32(info.Mode()),
	}, nil
}
