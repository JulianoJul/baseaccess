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
App con **Wails v2 + Go html/template + HTMX**, **multi-modulo** (9 tipos de documentos). Schema dividido en `data/sql/01_master_control_docs_presidencia.sql` + `data/sql/02_modulos_adicionales.sql` + `data/sql/03_ruta_procesos.sql`. API Go unificada via `var Modulos map[string]ModuloConfig` en `app.go`. Bottom bar fija tipo hojas de cálculo con pestañas de módulos y Ruta Procesos a la derecha. Ruta Procesos ahora incluye soporte para **Hojas** (ventanas de tiempo persistentes con paginación independiente de semanas), una leyenda de colores estricta e independiente, y soporte para agregar procesos desde **cualquier módulo** (expedientes, memorandums, recobros, etc.) usando un selector de módulo. Archivos SQL embebidos via `//go:embed data/sql/*.sql` para portabilidad. Rama `wails-migration` activa.
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

## Bugs Conocidos (Frontend)
Ver `doc.md` → Bugs Conocidos. Resumen: `location.reload()` después de guardar/eliminar vuelve a expedientes, y el botón "Nuevo Registro" a veces no pasa el módulo correcto. El backend Go funciona correctamente.
