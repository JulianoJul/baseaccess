# AI Context — Gestión de Expedientes (Wails)

## Stack (no negociable)
- **Wails v2** (Go 1.21+): backend nativo, frontend web embebido
- **Go** + **mattn/go-sqlite3**: acceso SQLite directo al archivo .db
- **Frontend**: Go `html/template` renderiza el HTML desde `TemplateHandler`
- **Estáticos**: Tailwind CSS + Font Awesome (en `frontend/vendor/`)
- **Sin CDN, sin frameworks JS, sin backend externo**

## Líneas Rojas
- **Cero hardcodeo**: todo valor variable → constantes con nombre (`CONFIG.*`)
- **SPOT**: `app.go` + `handler.go` son la fuente de verdad del schema
- **SoC**: separar Go (backend/BD) de JS (UI mínimo). JS solo controla modales y localStorage
- **KISS + YAGNI**: resolver solo lo pedido, sin features "por si acaso"
- **Sin efectos secundarios ocultos**: las funciones deben ser predecibles (Least Astonishment)
- **Makefile**: única fuente de automatización local

## Estado Actual (Julio 2026)
App con **Wails v2 + Go html/template + HTMX**. El HTML y los fragmentos de la interfaz se renderizan en el backend de Go mediante `TemplateHandler` y plantillas HTML parciales en `templates/`. La comunicación es gestionada de manera declarativa con **htmx**. `schema-config.js` ya no se carga; sus constantes fueron inlineadas o migradas al backend. Únicos scripts JS restantes: helpers de modales, paginación DOM, localStorage (recientes/fijados) y el binding `AbrirDialogoBD` de Wails. Rama `wails-migration` activa.

## Archivos Clave
| Archivo | Para qué |
|---------|----------|
| `main.go` | Entry point Wails (Handler en AssetServer, bind App) |
| `handler.go` | TemplateHandler: http.Handler que renderiza templates Go |
| `templates/index.html` | Go html/template (estructura HTML renderizada desde Go) |
| `app.go` | Backend Go: App struct, 12 métodos CRUD SQLite |
| `go.mod` | Dependencias Go (wails/v2 + go-sqlite3) |
| `wails.json` | Config proyecto Wails |
| `frontend/schema-config.js` | Legacy — ya no se carga en templates (constantes inlineadas) |
| `frontend/ruta-procesos-data.js` | Datos Gantt para Ruta Procesos (único JS externo restante) |
| `docs/doc.md` | Documentación + changelog |
| `docs/decisiones.md` | ADR: historial de decisiones técnicas |
| `docs/funciones.md` | Catálogo SPOT de funciones |
| `.github/workflows/build.yml` | CI: build Wails (Linux + Windows) |

## Regla de Oro
Antes de tocar código: leer `doc.md` + `decisiones.md` + `funciones.md` + `ai-context.md`.
