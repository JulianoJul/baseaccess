# Gestión de Expedientes con Historial — Documentación (Wails)

> **Ver también:** [`decisiones.md`](decisiones.md) — ADR con historial de decisiones técnicas.
> **Anchor IA:** [`ai-context.md`](ai-context.md) — stack, líneas rojas, estado actual (lee esto primero).
> **Catálogo:** [`funciones.md`](funciones.md) — SPOT de funciones (DRY: verificar antes de crear).

## Stack

| Capa | Tecnología |
|------|-----------|
| Backend | Go 1.21+, Wails v2 |
| SQLite | mattn/go-sqlite3 (driver nativo) |
| Frontend | HTML + Tailwind CSS + Font Awesome |
| Empaquetado | Wails CLI (`wails build`) |
| Windows | WebView2 Fixed Version Runtime incluido (portable) |

## Arquitectura

```
┌──────────────────────────────────────────────────┐
│  main.go (entry point Wails)                      │
│  ├── go:embed all:frontend → assets embed.FS      │
│  ├── Bind: app (*App) → expuesto a JS             │
│  └── windows.Options.WebviewBrowserPath → runtime │
├──────────────────────────────────────────────────┤
│  app.go (backend Go nativo)                       │
│  ├── App struct { db *sql.DB, mu sync.Mutex }     │
│  ├── AbrirBaseDatos(filePath) → sql.Open          │
│  ├── ObtenerExpedientes(orden) → SELECT vista     │
│  ├── GuardarExpediente(data) → INSERT/UPDATE      │
│  ├── EliminarExpediente(id) → DELETE transacción  │
│  └── ...otros 8 métodos                           │
├──────────────────────────────────────────────────┤
│  frontend/ (embebido en binario)                  │
│  ├── index.html (HTML + CSS + JS)                 │
│  ├── schema-config.js (config del schema)         │
│  └── vendor/ (Tailwind, FontAwesome, styles)      │
└──────────────────────────────────────────────────┘
```

**Flujo de datos:**
```
Usuario → Click en UI → JS llama window.go.main.App.*
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
├── main.go                 # Entry point Wails
├── app.go                  # Backend Go (App struct, 12 métodos)
├── go.mod                  # Dependencias Go
├── go.sum                  # Checksums Go
├── wails.json              # Config proyecto Wails
├── frontend/               # Código fuente (embebido vía go:embed)
│   ├── index.html          # App completa (HTML + CSS + JS)
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
