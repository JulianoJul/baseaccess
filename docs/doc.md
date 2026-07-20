# Gestión de Expedientes con Historial — Documentación (Wails)

## Stack

| Capa | Tecnología |
|------|-----------|
| Backend | Go 1.21+, Wails v2 |
| SQLite | mattn/go-sqlite3 (driver nativo) |
| Frontend | Go `html/template` + HTMX + Tailwind CSS + Font Awesome |
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
│  │   ├── "/api/*" → renderiza fragmentos HTML      │
│  │   └── otro → sirve archivo estático             │
│  └── PageData struct (datos inyectados al template)│
├──────────────────────────────────────────────────┤
│  templates/ (Go html/template)                    │
│  ├── index.html          # Template principal      │
│  ├── tabla_<key>.html (9)# Listado por modulo      │
│  ├── form_<key>.html (9) # Formulario por modulo   │
│  ├── historial.html      # Historial (multi-modulo)│
│  ├── ruta_procesos.html  # Ruta de procesos Gantt  │
│  └── pendientes.html     # Docs pendientes         │
├──────────────────────────────────────────────────┤
│  app.go (backend Go nativo)                       │
│  ├── App struct { db *sql.DB, mu sync.Mutex }     │
│  ├── AbrirBaseDatos(filePath) → sql.Open          │
│  ├── ObtenerFilas(moduloKey, orden) → SELECT vista│
│  ├── GuardarFila(moduloKey, data) → INSERT/UPDATE │
│  ├── EliminarFila(moduloKey, id) → DELETE transacc│
│  └── ... Modulos map (9 modulos)                  │
├──────────────────────────────────────────────────┤
│  frontend/ (estáticos embebidos)                  │
│  └── vendor/ (Tailwind, FontAwesome, HTMX, styles)│
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

**Flujo de datos (interacción y actualización reactiva):**
```
Usuario → Click → HTMX realiza petición HTTP (hx-get / hx-post)
                             ↓
                     handler.go (Go)
                             ↓
                 Retorna fragmento HTML parcial
                             ↓
                 HTMX actualiza el DOM de forma reactiva
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
│   ├── index.html          # Template principal (estructura HTML)
│   ├── tabla_<key>.html (9)# Listado por modulo
│   ├── form_<key>.html (9) # Formulario por modulo
│   ├── historial.html      # Historial (multi-modulo)
│   ├── ruta_procesos.html  # Ruta de procesos Gantt
│   └── pendientes.html     # Docs pendientes
├── frontend/               # Estáticos embebidos (CSS, JS, fuentes)
│   └── vendor/             # Dependencias locales (sin CDN)
│       ├── tailwind.min.css    # Tailwind CSS build estático
│       ├── styles.css          # Estilos adicionales
│       ├── htmx.min.js         # HTMX
│       ├── fontawesome.min.css # Font Awesome Free
│       └── webfonts/           # Fuentes de iconos
├── data/                   # Archivos de datos
│   ├── sql/01_master_control_docs_presidencia.sql  # Schema: catalogos + expedientes
│   ├── sql/02_modulos_adicionales.sql               # Schema: 8 modulos adicionales
│   └── sql/03_ruta_procesos.sql                      # Schema: ruta procesos (Gantt)
├── docs/                   # Documentación
│   ├── doc.md              # Este archivo
│   ├── ai-context.md       # Anchor IA
│   ├── funciones.md        # Catálogo SPOT
│   └── legacy/             # Docs históricos (era pre-Wails)
│       ├── decisiones.md   # ADR completo
│       └── CHANGELOG.md    # Historial de cambios
├── Makefile                # Automatización
├── build/                  # Outputs de build (gitignored)
│   └── bin/                # Binarios + WebView2 runtime (Windows)
└── .github/workflows/      # CI/CD
```

## Tablas del Schema

| Tabla | Propósito |
|-------|-----------|
| `cat_gerencia` | Catálogo de gerencias (13 registros, IDs 1-13) |
| `cat_superintendencia` | Catálogo de superintendencias (FK → gerencia, 17 registros) |
| `cat_documento` | Tipos de documento (28 registros) |
| `cat_plan_contratacion` | Planes de contratación |
| `cat_modalidad` | Modalidades de contratación |
| `cat_art` | Artículos de normativa interna |
| `cat_tipo_contrato` | Tipos de contrato (PU, SG, MIXTO) |
| `cat_estatus_detalle` | Estatus (10 valores) |
| `cat_resultado_proceso` | Resultados (Adjudicado, Desierto...) |
| `cat_empresas` | Empresas adjudicadas |
| `cat_responsables` | Emisores/Receptores |
| `expedientes` | **Contrataciones**: ~30 columnas con fechas, montos, FK |
| `historial_movimientos` | Traza de cambios vía trigger |
| `vw_reporte_excel_contrataciones` | Vista JOIN completo para contrataciones |
| --- | **Módulos adicionales (02_modulos_adicionales.sql)** |
| `req_materiales` + `hist_req_materiales` | Requisición de Materiales |
| `memorandums` + `hist_memorandums` | Memorándums / Decisión de Gerencia |
| `recobros` + `hist_recobros` | Recobros |
| `valuaciones` + `hist_valuaciones` | Valuaciones |
| `aprobacion_jd` + `hist_aprobacion_jd` | Para Aprobación JD |
| `certificacion_bdu` + `hist_certificacion_bdu` | Certificación BDU |
| `vacaciones` + `hist_vacaciones` | Vacaciones |
| `reposos_medicos` + `hist_reposos_medicos` | Reposo Médico |
| `vw_reporte_req_materiales` … `vw_reporte_reposos_medicos` | 8 vistas JOIN por módulo |

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
make wails-build-linux      # Build Linux AMD64 (debug)
make wails-build-linux-prod # Build Linux AMD64 (produccion)
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

**Norma:** Antes de cada `GuardarFila()`, `EliminarFila()`, `GuardarNuevoCatalogo()` y `OptimizarBD()`, Go crea una copia de seguridad del .db actual con rotación de N backups (`.bak.1` más reciente, `.bak.N` más antiguo). Implementado en `app.go`.

**Importante modo WAL:** La BD opera en modo WAL (`_journal_mode=WAL`). Antes de copiar el archivo se ejecuta `PRAGMA wal_checkpoint(TRUNCATE)` para forzar el volcado del WAL al archivo principal, garantizando que el backup sea consistente.

### 2. SoC — Separation of Concerns

Separar estrictamente:
- **Go (app.go)**: acceso a datos SQLite, lógica de negocios
- **SPOT**: `app.go` + `handler.go` son la fuente de verdad del schema y datos
- **SoC**: separar Go (backend/BD) de JS (UI mínimo). JS solo controla modales y localStorage

### 3. SPOT — Single Point of Truth

- `app.go` es el SPOT para toda la lógica de BD
- `funciones.md` es el SPOT del catálogo de funciones

### 4. KISS — Keep It Simple, Stupid

### 5. YAGNI — You Aren't Gonna Need It

### 6. Principio de Menor Sorpresa (Least Astonishment)

### 7. Cohesión Alta, Acoplamiento Bajo

## CI/CD (GitHub Actions)

Workflow: `.github/workflows/build.yml`
- Push a `master` o `wails-migration` dispara build
- Jobs: `wails` (Linux + Windows)

Ver [`legacy/CHANGELOG.md`](legacy/CHANGELOG.md) para el historial completo de cambios.


### Rutas API del handler

| Ruta | Método | Descripción |
|------|--------|-------------|
| `/api/guardar-expediente` | POST | Guarda (INSERT/UPDATE) desde formulario |
| `/api/eliminar-expediente` | POST | Elimina expediente + historial por ID |
| `/api/cargar-expediente` | GET | Devuelve fragmento HTML del formulario de edición |
| `/api/filtrar-expedientes` | GET | Filtra, ordena y devuelve fragmento HTML de las filas de la tabla |
| `/api/cambiar-modulo` | GET | Cambia de módulo y devuelve fragmento HTML de la tabla correspondiente |
| `/api/historial` | GET | Devuelve fragmento HTML del historial de un registro (multi-módulo) |
| `/api/abrir-bd` | POST | Abre base de datos SQLite por ruta |
| `/api/ruta-procesos` | GET | Devuelve fragmento HTML de la vista Gantt de procesos |
| `/api/ruta-procesos-agregar` | POST | Agrega un proceso a la ruta (vinculado a un registro existente de cualquier módulo) |
| `/api/ruta-procesos-toggle` | POST | Activa/desactiva un proceso en la ruta |
| `/api/ruta-procesos-eliminar` | POST | Elimina un proceso de la ruta |
| `/api/ruta-procesos-registros` | GET | Devuelve JSON con registros disponibles para agregar como procesos (`?modulo=xxx`) |
| `/api/ruta-procesos-leyenda-crear` | POST | Crea una leyenda personalizada |
| `/api/ruta-procesos-leyenda-actualizar` | POST | Actualiza nombre y color de una leyenda existente |
| `/api/pendientes` | GET | Devuelve fragmento HTML de documentos pendientes |
| `/api/guardar-catalogo` | POST | Agrega registro a un catálogo |
| `/api/optimizar-bd` | POST | Ejecuta VACUUM |
| `/api/csv` | GET | Descarga CSV del módulo indicado (`?modulo=...`)

---

> **Ver también:** [`docs/legacy/decisiones.md`](legacy/decisiones.md) — ADR completo (historial de decisiones técnicas, incluyendo era sql.js legacy).
> **Anchor IA:** [`ai-context.md`](ai-context.md) — stack, líneas rojas, estado actual (lee esto primero).
> **Changelog:** [`docs/legacy/CHANGELOG.md`](legacy/CHANGELOG.md) — historial completo de cambios.
> **Catálogo:** [`funciones.md`](funciones.md) — SPOT de funciones (DRY: verificar antes de crear).
