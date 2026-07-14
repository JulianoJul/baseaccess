# Archivos Legacy a eliminar

Proyecto migrado de Electron → Tauri → Wails. Todo lo que no usa Wails es candidato a borrar.

---

## Directorios completos

| Directorio | Tamaño | Descripción | Legacy de |
|------------|--------|-------------|-----------|
| `node_modules/` | ~133 MB | Dependencias npm (Electron, Tauri CLI) | Electron / Tauri |
| `src/` | ~1.2 MB | Frontend antiguo con sql.js WASM | Electron / Tauri |
| `src-tauri/` | ~170 KB | Backend Rust (Cargo.toml, lib.rs) | Tauri |
| `dist/` | ~359 MB | Build de Electron (GestionExpedientes.exe) | Electron |

## Archivos sueltos

| Archivo | Tamaño | Descripción | Legacy de |
|---------|--------|-------------|-----------|
| `main.js` | ~2 KB | Main process de Electron | Electron |
| `package.json` | ~1 KB | Config npm con scripts de Electron/Tauri | Electron / Tauri |
| `package-lock.json` | ~40 KB | Lockfile npm | Electron / Tauri |

## No se tocan

- `data/` — explícitamente excluido
- `docs/` — documentación del proyecto
- `app.go`, `handler.go`, `main.go` — fuente Go activo
- `go.mod`, `go.sum` — dependencias Go
- `templates/` — 24 plantillas HTML
- `frontend/` — assets embebidos (vendor/, wailsjs/, etc.)
- `Makefile` — targets de build Wails (contiene targets legacy pero se mantiene)
- `.github/` — CI/CD (contiene jobs legacy pero se mantiene)
- `Microsoft.WebView2.FixedVersionRuntime.150.0.4078.65.x64.cab` — runtime WebView2 para compilar en Windows sin admin
- `.gitignore`, `wails.json`, `plan.md`, `prompt`, `combined.txt`, `.clinerules` — config y docs

**Total a recuperar**: ~493 MB
