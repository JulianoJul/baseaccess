# Gestión de Expedientes con Historial — Documentación (Wails)

> **Ver también:** [`decisiones.md`](decisiones.md) — ADR con historial de decisiones técnicas.
> **Anchor IA:** [`ai-context.md`](ai-context.md) — stack, líneas rojas, estado actual (lee esto primero).
> **Catálogo:** [`funciones.md`](funciones.md) — SPOT de funciones (DRY: verificar antes de crear).

## Stack

| Capa | Tecnología |
|------|-----------|
| Backend | Go 1.21+, Wails v2 |
| SQLite | mattn/go-sqlite3 (driver nativo) |
| Frontend | Go `html/template` + Tailwind CSS + Font Awesome |
| Renderizado | `TemplateHandler` (http.Handler) intercepta AssetServer |
| Empaquetado | Wails CLI (`wails build`) |
| Windows | WebView2 Fixed Version Runtime incluido (portable) |

## Arquitectura

```
┌──────────────────────────────────────────────────┐
│  main.go (entry point Wails)                      │
│  ├── NewTemplateHandler(app) → http.Handler       │
│  ├── Handler en AssetServer (en vez de Assets)    │
│  ├── Bind: app (*App) → expuesto a JS             │
│  └── windows.Options.WebviewBrowserPath → runtime │
├──────────────────────────────────────────────────┤
│  handler.go (TemplateHandler)                     │
│  ├── go:embed all:frontend → estáticos (CSS/JS)   │
│  ├── go:embed templates/* → Go html/template      │
│  ├── ServeHTTP()                                   │
│  │   ├── "/" → renderiza template Go con datos     │
│  │   └── otro → sirve archivo estático             │
│  └── PageData struct (datos inyectados al template)│
├──────────────────────────────────────────────────┤
│  templates/ (Go html/template)                    │
│  └── index.html (estructura HTML renderizada Go)  │
├──────────────────────────────────────────────────┤
│  app.go (backend Go nativo)                       │
│  ├── App struct { db *sql.DB, mu sync.Mutex }     │
│  ├── AbrirBaseDatos(filePath) → sql.Open          │
│  ├── ObtenerExpedientes(orden) → SELECT vista     │
│  ├── GuardarExpediente(data) → INSERT/UPDATE      │
│  ├── EliminarExpediente(id) → DELETE transacción  │
│  └── ...otros 8 métodos                           │
├──────────────────────────────────────────────────┤
│  frontend/ (estáticos embebidos)                  │
│  ├── schema-config.js (config del schema)         │
│  ├── ruta-procesos-data.js (datos Gantt)          │
│  └── vendor/ (Tailwind, FontAwesome, styles)      │
└──────────────────────────────────────────────────┘
```

**Flujo de datos (render inicial):**
```
Wails webview → GET / → TemplateHandler.ServeHTTP()
                            ↓
                    templates/index.html (Go template)
                            ↓
                    HTML renderizado con datos Go
                            ↓
                    Wails inyecta runtime JS automáticamente
                            ↓
                    Navegador carga CSS/JS desde estáticos
```

**Flujo de datos (interacción):**
```
Usuario → Click → JS llama window.go.main.App.*
                          ↓
                  app.go (Go)
                    ↓
              database/sql + go-sqlite3
                    ↓
              Archivo .db (escritura directa)
```

## Estructura del Proyecto

```
baseaccess/
├── main.go                 # Entry point Wails (Handler en AssetServer)
├── handler.go              # TemplateHandler: http.Handler con templates Go
├── app.go                  # Backend Go (App struct, 12 métodos)
├── go.mod                  # Dependencias Go
├── go.sum                  # Checksums Go
├── wails.json              # Config proyecto Wails
├── templates/              # Go html/template (renderizados desde Go)
│   └── index.html          # Template principal (estructura HTML)
├── frontend/               # Estáticos embebidos (CSS, JS, fuentes)
│   ├── index.html          # Obsoleto (mantenido como fallback estático)
│   ├── schema-config.js    # Config del schema (catálogos, columnas, etc.)
│   ├── ruta-procesos-data.js  # Datos Gantt para Ruta Procesos
│   └── vendor/             # Dependencias locales (sin CDN)
│       ├── tailwind.min.css    # Tailwind CSS build estático
│       ├── styles.css          # Estilos adicionales
│       ├── fontawesome.min.css # Font Awesome Free
│       └── webfonts/           # Fuentes de iconos
├── data/                   # Archivos de datos
│   ├── sql/Tablas8.sql     # Schema SQLite v8
│   └── importar_datos.py   # Script de importación desde Excel
├── docs/                   # Documentación
│   ├── doc.md              # Este archivo
│   ├── decisiones.md       # ADR
│   ├── ai-context.md       # Anchor IA
│   └── funciones.md        # Catálogo SPOT
├── Makefile                # Automatización
├── build/                  # Outputs de build (gitignored)
│   └── bin/                # Binarios + WebView2 runtime (Windows)
└── .github/workflows/      # CI/CD
```

## Tablas del Schema (Tablas8.sql)

| Tabla | Propósito |
|-------|-----------|
| `cat_gerencia` | Catálogo de gerencias |
| `cat_superintendencia` | Catálogo de superintendencias (FK → gerencia) |
| `cat_documento` | Tipos de documento (28 registros) |
| `cat_plan_contratacion` | Planes de contratación |
| `cat_modalidad` | Modalidades de contratación |
| `cat_art` | Artículos de normativa interna |
| `cat_tipo_contrato` | Tipos de contrato (PU, SG, MIXTO) |
| `cat_estatus_detalle` | Estatus (Pendiente, Firmado, Devuelto...) |
| `cat_resultado_proceso` | Resultados (Adjudicado, Desierto...) |
| `cat_empresas` | Empresas adjudicadas |
| `cat_responsables` | Emisores/Receptores |
| `expedientes` | **Tabla principal**: ~30 columnas con fechas, montos, FK |
| `historial_movimientos` | Traza de cambios vía trigger |
| `vw_reporte_excel_contrataciones` | Vista JOIN completo para reportes |

## Esquema de Colores

Tailwind CSS (dark mode personalizado):
- Fondo: `bg-gray-900` | Superficie: `bg-gray-800` | Bordes: `border-gray-700`
- Texto: `text-gray-100` | Secundario: `text-gray-400`
- Acento: `teal-400` (botones, encabezados) | `teal-600` (botón primario)
- Estados: `emerald-400` (adjudicado) | `amber-400` (presupuesto) | `red-700` (eliminar)

## Makefile

```bash
make wails-install          # go install Wails CLI
make wails-dev              # wails dev (hot reload)
make wails-build-linux      # Build Linux AMD64
make wails-build-win        # Build Windows AMD64 (con WebView2 embed)
make wails-build            # Build Linux (default)
```

## Build (Wails)

**Linux:**
```bash
make wails-build-linux
# build/bin/GestionExpedientes
```

**Windows (desde Linux con MinGW o desde Windows):**
```bash
make wails-build-win
# build/bin/GestionExpedientes.exe + Microsoft.WebView2.FixedVersionRuntime.*/
```

El binario es 100% portable: copiar `build/bin/` a cualquier máquina y ejecutar.

## Principio Fundamental

**Cero assumptions del schema.** Todo se genera dinámicamente analizando la BD al cargarla:
- Catálogos → selectores poblados con `cargarCatalogos()`
- Vistas → tabla basada en `vw_reporte_excel_contrataciones`
- Historial → consulta JOIN bajo demanda al expandir fila

## Normas de Desarrollo / Buenas Prácticas

### 1. Backup Rotativo antes de cada escritura

**Riesgo:** Cortes de energía frecuentes pueden corromper el .db si ocurren durante una escritura física.

**Norma:** Antes de cada `GuardarExpediente()`, `EliminarExpediente()`, `GuardarNuevoCatalogo()` y `OptimizarBD()`, Go crea una copia de seguridad del .db actual con rotación de N backups (`.bak.1` más reciente, `.bak.N` más antiguo). Implementado en `app.go`.

### 2. SoC — Separation of Concerns

Separar estrictamente:
- **Go (app.go)**: acceso a datos SQLite, lógica de negocios
- **JS (frontend/index.html)**: UI, eventos, renderizado
- JS **nunca** construye queries SQL. Solo llama `window.go.main.App.*`

### 3. SPOT — Single Point of Truth

- `frontend/schema-config.js` es el SPOT para todo lo específico del schema
- `app.go` es el SPOT para toda la lógica de BD
- `funciones.md` es el SPOT del catálogo de funciones

### 4. KISS — Keep It Simple, Stupid

### 5. YAGNI — You Aren't Gonna Need It

### 6. Principio de Menor Sorpresa (Least Astonishment)

### 7. Cohesión Alta, Acoplamiento Bajo

## CI/CD (GitHub Actions)

Workflow: `.github/workflows/build.yml`
- Push a `master` o `wails-migration` dispara build
- Jobs: `tauri` (legacy), `wails`, `electron` (legacy)
- Wails: Linux (binary) + Windows (binary + WebView2 Fixed Runtime)

## Cambios Realizados

### Rama wails-migration (Julio 2026)

| # | Archivo | Cambio | Razón |
|---|---------|--------|-------|
| 1 | `main.go`, `app.go`, `go.mod`, `wails.json` | **Creados**: entry point Wails, backend Go con 12 métodos, dependencias | Migración de Electron/sql.js a Wails v2 |
| 2 | `frontend/` | **Creado** desde `src/`: index.html adaptado, sql.js reemplazado por `window.go.main.App.*` | Bindings Wails Go↔JS |
| 3 | `.github/workflows/build.yml` | Job `wails` añadido (Linux + Windows) | CI para nuevo runtime |
| 4 | `Makefile` | Targets `wails-*` añadidos | Automatización local |
| 5 | `.gitignore` | `build/bin/` añadido | Outputs Wails ignorados |
| 6 | `main.go`, `build.yml` | WebView2 Fixed Version Runtime portable para Windows | 100% portabilidad sin instalación |
| 7 | `app.go` | Backup rotativo implementado en Go antes de cada escritura | Cortes de energía frecuentes — riesgo de corrupción |
| 8 | `frontend/index.html` | Botón "Guardar BD" eliminado, `descargarBDError()` implementado (diálogo guardar copia), autosave eliminado | Escrituras directas Go, no necesitan exportación ni timer |
| 9 | `docs/*` | Actualizados: doc.md, ai-context.md, funciones.md, decisiones.md | Reflejar arquitectura Wails |
| 10 | `app.go` | Auditoría: `GuardarNuevoCatalogo` whitelist tabla/columna (SQL injection fix), `ObtenerCatalogos` rows.Close() por iteración (resource leak fix), queries a constantes `query*`, type assertions estrictas, `crearBackup()` con io.Copy, `SetBackupMaxCopies(n)` configurable. `AbrirDialogoBD()`/`GuardarDialogoBD()` para diálogos nativos Wails (vía `wailsRuntime.OpenFileDialog`). `queryRows` convierte `time.Time`→`"2006-01-02"` | Seguridad + SoC + portabilidad |
| 11 | `main.go` | `EnableDefaultContextMenu: true`, `Debug.OpenInspectorOnStartup` para F12 en WebKitGTK | Debugging en Linux |
| 12 | `frontend/index.html` | Detección `window.go.main.App` → bindings Wails (`AbrirDialogoBD`) vs fallback navegador. Fix bug `exportarCSV()` líneas duplicadas corruptas (`const` duplicado). Botón `sort-dir` toggle ASC/DESC con persistencia `localStorage`. Tooltip `title="Ordenar"` en select. Gap coherente `gap-3` global | UX + bugfix |
| 13 | `frontend/schema-config.js` | `MODAL_HISTORIAL`/`HISTORIAL_CONTENIDO` renombrados (`historial-modal`/`historial-cuerpo`). Eliminados selectores fantasma (`SIDEBAR_TOGGLE`, `MENU_RECIENTES`, `BTN_VACUUM`). Añadido `ORDEN_DIRECCION` a `STORAGE_KEYS` | Bugfix SMOKE test |
| 14 | `frontend/vendor/styles.css` | `color-scheme: dark` (WebKitGTK controles nativos oscuros). `select.input option` bg/color. ~20 utilidades Tailwind faltantes (`bg-gray-700/10`, `border-gray-800`, etc.) emuladas con `rgba()` — fix bordes blancos visibles | Bugfix Tailwind purgado |
| 15 | `data/importar_datos.py` | `DROP TRIGGER trg_exp_auditoria` durante migración. Tracking por solped: `fecha_creacion=MIN(fecha_recibido)`, `fecha_actualizacion=MAX(fecha_devuelto or fecha_recibido)`. Trigger recreado al final | Fix fechas migración Excel |
| 16 | `handler.go` | **Creado**: TemplateHandler con `http.Handler`, embebe `frontend/` y `templates/`, sirve templates Go para `/` y estáticos para el resto | Migración a Go html/template (DEC-011) |
| 17 | `templates/index.html` | **Creado**: Go template con estructura HTML de la app (332 líneas), renderizado desde `html/template` con datos Go | Migración a Go html/template (DEC-011) |
| 18 | `main.go` | `Assets: assets` → `Handler: handler`. Eliminado `//go:embed all:frontend` (ahora en handler.go). Nuevo `NewTemplateHandler(app)` | AssetServer ahora usa Handler personalizado |
| 19 | `handler.go` | `PageData` con `Catalogs` + `Expedientes` precargados. 10 rutas `/api/*` (JSON) para CRUD, BD, historial, ruta procesos, pendientes, CSV, catálogos, VACUUM. Funciones template: `default`, `rowGet`, `rowGetStr`, `rowGetNum`, `estatusClass`, `formatNum`, `jsonEncode`, `truncate`, `isSelected` | Pasos 1-2 del roadmap completados |
| 20 | `templates/index.html` | Reescrito: tabla renderizada con `{{range .Expedientes}}`, `<select>` del formulario rellenados desde `{{range .Catalogs.*}}`. JS reducido a `fetch()` a `/api/*` + toggle modales + apertura BD (único binding Wails restante). Eliminados: pagination JS, orden JS, cache JS, bindings Go directos | Paso 3 del roadmap completado |
| 21 | `app.go` | `CatalogoItem` struct: añadido `IDGerencia int` para filtrar superintendencias por gerencia. `ObtenerCatalogos` ahora popula `IDGerencia` | Soporte template superintendencias |

## Migración a Go html/template — Estado

| # | Paso | Estado | Detalle |
|---|------|--------|---------|
| 1 | **Datos precargados en PageData** | ✅ Hecho | `handler.go` — `PageData` inyecta catálogos y expedientes. El template renderiza la tabla con `{{range}}`. |
| 2 | **Rutas API en el handler** | ✅ Hecho | `handler.go` — 10 rutas `/api/*` (JSON) para CRUD, abrir BD, historial, ruta procesos, pendientes, CSV, catálogos, VACUUM. |
| 3 | **Reemplazar bindings JS** | ✅ Hecho | `templates/index.html` — `fetch()` a `/api/*` reemplaza `window.go.main.App.*`. Solo queda 1 binding Wails: `AbrirDialogoBD` (diálogo nativo de archivos). |
| 4 | **HTMX** | ⏸ Postergado | Se evaluó pero `fetch()` + JS mínimo es suficiente para el alcance actual. |

### Rutas API del handler

| Ruta | Método | Descripción |
|------|--------|-------------|
| `/api/guardar-expediente` | POST | Guarda (INSERT/UPDATE) desde formulario |
| `/api/eliminar-expediente` | POST | Elimina expediente + historial por ID |
| `/api/cargar-expediente` | GET | Devuelve JSON del expediente para edición |
| `/api/historial` | GET | Devuelve JSON del historial de un expediente |
| `/api/abrir-bd` | POST | Abre base de datos SQLite por ruta |
| `/api/ruta-procesos` | GET | Devuelve JSON de la ruta de procesos |
| `/api/pendientes` | GET | Devuelve JSON de documentos pendientes |
| `/api/guardar-catalogo` | POST | Agrega registro a un catálogo |
| `/api/optimizar-bd` | POST | Ejecuta VACUUM |
| `/api/csv` | GET | Descarga CSV de expedientes |
