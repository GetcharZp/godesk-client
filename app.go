package main

import (
	"context"
	"fmt"
	"godesk-client/internal/service/device"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func resp(data any, err error) any {
	if err == nil {
		return map[string]any{
			"code": 200,
			"msg":  "success",
			"data": data,
		}
	} else {
		return map[string]any{
			"code": -1,
			"msg":  err.Error(),
		}
	}
}

// DeviceInfo 设备信息
func (a *App) DeviceInfo() any {
	return resp((&device.Service{}).Info())
}
