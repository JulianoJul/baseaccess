package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

const WebView2RuntimeDir = "Microsoft.WebView2.FixedVersionRuntime.150.0.4078.65.x64"

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Gestión de Expedientes con Historial",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
		EnableDefaultContextMenu: true,
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
		// Solo aplica en Windows (WebView2). En Linux/macOS se ignora.
		Windows: &windows.Options{
			WebviewBrowserPath: WebView2RuntimeDir,
		},
	})
	if err != nil {
		panic(err)
	}
}
