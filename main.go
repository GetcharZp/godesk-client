package main

import (
	"embed"
	"godesk-client/internal"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 初始化服务
	go internal.NewService()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "GoDesk",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 14, B: 39, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			Theme:                             windows.SystemDefault,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:   windows.RGB(10, 14, 39),
				DarkModeTitleText:  windows.RGB(0, 212, 255),
				DarkModeBorder:     windows.RGB(45, 53, 97),
				LightModeTitleBar:  windows.RGB(248, 250, 252),
				LightModeTitleText: windows.RGB(0, 136, 204),
				LightModeBorder:    windows.RGB(226, 232, 240),
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
