# AI Context — Gestión de Expedientes (Wails)

## Stack (no negociable)
- **Wails v2** (Go 1.21+): backend nativo, frontend web embebido
- **Go** + **mattn/go-sqlite3**: acceso SQLite directo al archivo .db
- **Frontend**: Go `html/template` renderiza el HTML desde `TemplateHandler`
- **Interactividad**: HTMX (declarativo, server-side)
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
App con **Wails v2 + Go html/template + HTMX**, **multi-modulo** (9 tipos de documentos). Schema en `data/sql/01_master_control_docs_presidencia.sql` + `02_modulos_adicionales.sql` + `03_ruta_procesos.sql`. API Go unificada via `var Modulos map[string]ModuloConfig` en `app.go`. Bottom bar tipo hojas de cálculo con pestañas de módulos. Ruta Procesos con Hojas (ventanas de tiempo persistentes), leyenda independiente, y procesos multi-módulo. SQL embebidos via `//go:embed`.
## Archivos Clave
| Archivo | Para qué |
|---------|----------|
| `main.go` | Entry point Wails (Handler en AssetServer, bind App) |
| `handler.go` | TemplateHandler: http.Handler que renderiza templates Go |
| `templates/index.html` | Go html/template (estructura HTML renderizada desde Go) |
| `app.go` | Backend Go: App struct, 12 métodos CRUD SQLite |
| `go.mod` | Dependencias Go (wails/v2 + go-sqlite3) |
| `wails.json` | Config proyecto Wails |
| `docs/doc.md` | Documentación + changelog |
| `docs/decisiones.md` | ADR: historial de decisiones técnicas |
| `docs/funciones.md` | Catálogo SPOT de funciones |
| `.github/workflows/build.yml` | CI: build Wails (Linux + Windows) |
| `data/sql/01_master_control_docs_presidencia.sql` | Schema: catalogos + expedientes + historial_movimientos + vistas + triggers |
| `data/sql/02_modulos_adicionales.sql` | Schema: 8 modulos adicionales con sus tablas hist_, vistas vw_reporte_*, triggers |
| `data/sql/03_ruta_procesos.sql` | Schema: ruta procesos (leyenda, cronograma, procesos) |

## Regla de Oro
Antes de tocar código: leer `doc.md` + `decisiones.md` + `funciones.md` + `ai-context.md`.


