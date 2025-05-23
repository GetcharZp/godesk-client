package main

import (
	"context"
	"fmt"
	"godesk-client/internal/service/device"
	"godesk-client/internal/service/user"
	pb "godesk-client/proto"
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

// GetUserInfo 获取用户信息
func (a *App) GetUserInfo() any {
	return resp((&user.Service{}).GetUserInfo())
}

// UserLogin 用户登录
func (a *App) UserLogin(req *pb.UserLoginRequest) any {
	return resp((&user.Service{}).Login(req))
}

// UserRegister 用户注册
func (a *App) UserRegister(req *pb.UserRegisterRequest) any {
	return resp((&user.Service{}).Register(req))
}

// UserLogout 退出登录
func (a *App) UserLogout() any {
	return resp((&user.Service{}).Logout())
}
