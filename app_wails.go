//go:build wails

package main

import (
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) AbrirDialogoBD() (string, error) {
	return wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Seleccionar base de datos",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "SQLite DB", Pattern: "*.db;*.sqlite"},
		},
	})
}

func (a *App) GuardarDialogoBD(nombreDefault string) (string, error) {
	return wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Guardar copia de base de datos",
		DefaultFilename: nombreDefault,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "SQLite DB", Pattern: "*.db;*.sqlite"},
		},
	})
}
