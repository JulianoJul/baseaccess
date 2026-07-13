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
App con **Wails v2 + Go html/template + HTMX**, ahora **multi-modulo** (9 tipos de documentos: expedientes, requisiciones, memorandums, recobros, valuaciones, aprobacion_jd, certificacion_bdu, vacaciones, reposos_medicos). Schema dividido en `data/sql/01_master_control_docs_presidencia.sql` + `data/sql/02_modulos_adicionales.sql`. API Go unificada via `var Modulos map[string]ModuloConfig` en `app.go`. Botonera inferior en `index.html` para conmutar modulos sin recargar. `frontend/wailsjs/` se regenera con `wails dev`. Unico binding Wails utilizado: `AbrirDialogoBD`. Rama `wails-migration` activa.

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
| `data/sql/01_master_control_docs_presidencia.sql` | Schema: catalogos + expedientes + historial_movimientos + vistas + triggers |
| `data/sql/02_modulos_adicionales.sql` | Schema: 8 modulos adicionales con sus tablas hist_, vistas vw_reporte_*, triggers |
| `templates/tabla_<key>.html` (9) | Plantilla de listado por modulo |
| `templates/form_<key>.html` (9) | Plantilla de formulario por modulo |

## Regla de Oro
Antes de tocar código: leer `doc.md` + `decisiones.md` + `funciones.md` + `ai-context.md`.
