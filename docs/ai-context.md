# AI Context — Gestión de Expedientes (Wails)

## Stack (no negociable)
- **Wails v2** (Go 1.21+): backend nativo, frontend web embebido
- **Go** + **mattn/go-sqlite3**: acceso SQLite directo al archivo .db
- **Frontend**: HTML + Tailwind CSS + Font Awesome (en `frontend/vendor/`)
- **Sin CDN, sin frameworks JS, sin backend externo**

## Líneas Rojas
- **Cero hardcodeo**: todo valor variable → constantes con nombre (`CONFIG.*`)
- **SPOT**: `frontend/schema-config.js` es la única fuente de verdad del schema
- **SoC**: separar Go (backend/BD) de JS (UI). Las funciones JS solo llaman `window.go.main.App.*`
- **KISS + YAGNI**: resolver solo lo pedido, sin features "por si acaso"
- **Sin efectos secundarios ocultos**: las funciones deben ser predecibles (Least Astonishment)
- **Makefile**: única fuente de automatización local

## Estado Actual (Julio 2026)
App migrada de Electron/sql.js a **Wails v2**. Backend Go con 12 métodos exportados (`App.AbrirBaseDatos`, `ObtenerExpedientes`, `GuardarExpediente`, etc.) + backup rotativo antes de cada escritura. Frontend 100% adaptado a bindings `window.go.main.App.*`. WebView2 Fixed Runtime incluido para portabilidad Windows. Rama `wails-migration`, `master` intacto con Electron/Tauri original.

## Archivos Clave
| Archivo | Para qué |
|---------|----------|
| `main.go` | Entry point Wails (embed frontend, bind App) |
| `app.go` | Backend Go: App struct, 12 métodos CRUD SQLite |
| `go.mod` | Dependencias Go (wails/v2 + go-sqlite3) |
| `wails.json` | Config proyecto Wails |
| `frontend/index.html` | App completa (HTML + CSS + JS) |
| `frontend/schema-config.js` | Config del schema (catálogos, columnas, etc.) |
| `docs/doc.md` | Documentación + changelog |
| `docs/decisiones.md` | ADR: historial de decisiones técnicas |
| `docs/funciones.md` | Catálogo SPOT de funciones |
| `.github/workflows/build.yml` | CI: build Wails (Linux + Windows) |

## Regla de Oro
Antes de tocar código: leer `doc.md` + `decisiones.md` + `funciones.md` + `ai-context.md`.
