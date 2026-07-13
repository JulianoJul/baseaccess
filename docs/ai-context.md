# AI Context — Gestión de Expedientes (Wails)

## Stack (no negociable)
- **Wails v2** (Go 1.21+): backend nativo, frontend web embebido
- **Go** + **mattn/go-sqlite3**: acceso SQLite directo al archivo .db
- **Frontend**: Go `html/template` renderiza el HTML desde `TemplateHandler`
- **Estáticos**: Tailwind CSS + Font Awesome (en `frontend/vendor/`)
- **Sin CDN, sin frameworks JS, sin backend externo**

## Líneas Rojas
- **Cero hardcodeo**: todo valor variable → constantes con nombre (`CONFIG.*`)
- **SPOT**: `frontend/schema-config.js` es la única fuente de verdad del schema
- **SoC**: separar Go (backend/BD) de JS (UI). Las funciones JS solo llaman `window.go.main.App.*`
- **KISS + YAGNI**: resolver solo lo pedido, sin features "por si acaso"
- **Sin efectos secundarios ocultos**: las funciones deben ser predecibles (Least Astonishment)
- **Makefile**: única fuente de automatización local

## Estado Actual (Julio 2026)
App con **Wails v2 + Go html/template + rutas API REST**. El HTML no es estático: lo renderiza `TemplateHandler` desde `templates/index.html` con datos inyectados (catálogos, expedientes). El handler expone 10 rutas `/api/*` (JSON) que el frontend consume con `fetch()`. JS reducido al mínimo: `fetch()`, toggle de modales, apertura de BD (único binding Wails: `AbrirDialogoBD`). Backend Go con 16 métodos + backup rotativo. WebView2 Fixed Runtime para Windows. Rama `wails-migration`, `master` intacto.

## Archivos Clave
| Archivo | Para qué |
|---------|----------|
| `main.go` | Entry point Wails (Handler en AssetServer, bind App) |
| `handler.go` | TemplateHandler: http.Handler que renderiza templates Go |
| `templates/index.html` | Go html/template (estructura HTML renderizada desde Go) |
| `app.go` | Backend Go: App struct, 12 métodos CRUD SQLite |
| `go.mod` | Dependencias Go (wails/v2 + go-sqlite3) |
| `wails.json` | Config proyecto Wails |
| `frontend/schema-config.js` | Config del schema (catálogos, columnas, etc.) |
| `frontend/ruta-procesos-data.js` | Datos Gantt para Ruta Procesos |
| `docs/doc.md` | Documentación + changelog |
| `docs/decisiones.md` | ADR: historial de decisiones técnicas |
| `docs/funciones.md` | Catálogo SPOT de funciones |
| `.github/workflows/build.yml` | CI: build Wails (Linux + Windows) |

## Regla de Oro
Antes de tocar código: leer `doc.md` + `decisiones.md` + `funciones.md` + `ai-context.md`.
