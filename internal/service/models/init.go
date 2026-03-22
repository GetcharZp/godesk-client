package models

import (
	"godesk-client/internal/define"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	DB          *gorm.DB
	dbOnce      sync.Once
	homePath, _ = os.UserHomeDir()
	dbPath      = filepath.Join(homePath, define.DefaultConfig.AppName, "config.db")
)

// InitDB 初始化数据库
func InitDB() error {
	var initErr error
	dbOnce.Do(func() {
		// 初始化文件夹
		err := os.MkdirAll(path.Dir(dbPath), os.ModePerm)
		if err != nil {
			panic(err)
		}
		// 连接数据库
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			panic(err)
		}
		// 数据库迁移
		err = db.AutoMigrate(&SysConfig{})
		if err != nil {
			panic(err)
		}
		// 配置连接池
		sqlDB, err := db.DB()
		if err != nil {
			panic(err)
		}
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(130)
		sqlDB.SetConnMaxLifetime(time.Hour)
		DB = db
		// 初始化数据
		initData()
	})
	return initErr
}
