# Gestión de Expedientes con Historial — Documentación (Wails)

> **Ver también:** [`decisiones.md`](decisiones.md) — ADR con historial de decisiones técnicas.
> **Anchor IA:** [`ai-context.md`](ai-context.md) — stack, líneas rojas, estado actual (lee esto primero).
> **Catálogo:** [`funciones.md`](funciones.md) — SPOT de funciones (DRY: verificar antes de crear).

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
│   ├── decisiones.md       # ADR
│   ├── ai-context.md       # Anchor IA
│   └── funciones.md        # Catálogo SPOT
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
| 17 | `templates/index.html` | **Creado**: Go template con estructura HTML de la app, renderizado desde `html/template` con datos Go | Migración a Go html/template (DEC-011) |
| 18 | `main.go` | `Assets: assets` → `Handler: handler`. Eliminado `//go:embed all:frontend` (ahora en handler.go). Nuevo `NewTemplateHandler(app)` | AssetServer ahora usa Handler personalizado |
| 19 | `handler.go` | `PageData` con `Catalogs` + `Filas` precargados (multi-módulo). 12 rutas `/api/*` para CRUD, BD, historial, cambio de módulo, ruta procesos, pendientes, CSV, catálogos, VACUUM | Pasos 1-2 del roadmap completados |
| 20 | `templates/index.html` | Reescrito: tabla renderizada con `{{range .Filas}}`, `<select>` del formulario rellenados desde `{{range .Catalogs.*}}`. JS reducido a `fetch()` / `htmx`. Botonera inferior para cambiar de módulo. Título dinámico del modal. | Paso 3 del roadmap completado |
| 21 | `app.go` | `CatalogoItem` struct: añadido `IDGerencia int` para filtrar superintendencias por gerencia | Soporte template superintendencias |
| 22 | `templates/*`, `handler.go`, `index.html` | Migración completa a HTMX y plantillas fragmentadas | Remoción de gluecode JS para buscador, modales y formularios |
| 23 | `templates/ruta_procesos.html` | Gantt timeline restaurado utilizando `window.RUTA_PROCESOS_DATA` estático | Visualización correcta de Gantt en HTMX |
| 24 | `templates/index.html`, `templates/tabla_filas.html` | Panel de Fijados en modal superior, pins reactivos de color azul/verde en Acciones y bug de duplicados corregido | Acceso rápido premium |
| 25 | `templates/index.html`, `templates/pendientes.html` | Tabla configurada con `table-layout: fixed` y anchos proporcionales con reparto 50/50 para Documento/Descripción; badges con `whitespace-nowrap` | UX y diseño responsivo sin desbordamientos |
| 26 | `templates/index.html` | Paginación por bloques del lado del cliente acoplada con eventos de HTMX | Navegación de registros optimizada |
| 27 | `templates/index.html` | `schema-config.js` eliminado de imports; `STORAGE_KEYS` inlineado directamente en el template. Todas las references a `'baseaccess_recientes'` migradas a `STORAGE_KEYS.RECIENTES` | Eliminación total de dependencia externa de schema-config.js |
| 28 | `data/sql/01_master_control_docs_presidencia.sql`, `data/sql/02_modulos_adicionales.sql` | **Creados**: schema multi-módulo (master + 8 módulos adicionales con sus tablas hist_, vistas vw_reporte_*, y triggers de auditoría). `cat_gerencia` ampliada de 10 a 13 gerencias. | Soporte a 9 tipos de documentos en una sola BD |
| 29 | `app.go`, `handler.go`, `templates/*` | **Multi-módulo**: `Modulos` map (9 módulos), botonera inferior en index.html, `tabla_<key>.html`/`form_<key>.html` fragmentados, título dinámico del modal (`PAGE_DATA.modulos`), `historial.html` parametrizable (Receptor omitido en reposos_medicos, columna Notas añadida). API renombrada: `ObtenerFilas/GuardarFila/EliminarFila` con `moduloKey`. Wrappers legacy `*Expediente*` eliminados. | UI y API unificada multi-módulo |
| 30 | `templates/index.html` | **Bottom bar**: barra inferior fija tipo hojas de cálculo con pestañas de módulos. Ruta Procesos a la derecha en naranja. Oculta si no hay BD. | UX tipo spreadsheet |
| 31 | `templates/ruta_procesos.html`, `handler.go`, `app.go` | **Selector de expedientes existentes**: al añadir proceso en Ruta Procesos, se muestra un `<select>` con expedientes de la BD no agregados aún. Nuevo endpoint `/api/ruta-procesos-expedientes`. | Los procesos se agregan desde registros existentes |
| 32 | `data/sql/04_ruta_procesos_datos.sql` | **Eliminado**: seed data con IDs fijos (32-36) que podían no existir en la BD del usuario | Los procesos se agregan manualmente desde el selector |
| 33 | `Makefile` | Limpiado: eliminados targets legacy de Electron y Tauri | Proyecto Wails-only |
| 34 | `node_modules/`, `src/`, `src-tauri/`, `dist/`, `main.js`, `package.json`, `package-lock.json` | **Eliminados**: ~493 MB de archivos legacy de Electron/Tauri | Limpieza post-migración |
| 35 | `frontend/index.html`, `frontend/schema-config.js`, `frontend/ruta-procesos-data.js` | **Eliminados**: frontend legacy con bindings a métodos Go inexistentes, schema-config solo usado por index.html, Gantt hardcodeado (ahora server-side) | Limpieza de zombies post-migración |
| 36 | `app.go` | `backupMaxCopies` → `sync/atomic.Int64` | Race condition |
| 37 | `app.go` | WAL checkpoint (`PRAGMA wal_checkpoint(TRUNCATE)`) antes de backup | Consistencia en modo WAL |
| 38 | `app.go` | Backup verifica bytes copiados vs tamaño original | Detección de truncamiento |
| 39 | `app.go` | `ObtenerColumnasVista`: validación contra whitelist de vistas conocidas | SQL injection |
| 40 | `app.go` | `GuardarFila`: id `float64` → `int64`; UPDATE devuelve id real, no `LastInsertId()=0` | Precisión + lógica |
| 41 | `app.go` | `GuardarFila`: valores vacíos como `""` en vez de `nil` | Violación NOT NULL |
| 42 | `app.go` | `EliminarFila`: `defer Rollback` condicional post-commit | Transacción segura |
| 43 | `app.go` | `ObtenerCatalogos`: error loop antes de `rows.Close()` | Resource leak |
| 44 | `app.go` | `buildGanttColumns`: fechas dinámicas desde semana actual | Hardcode 2026 |
| 45 | `handler.go` | `handleEliminarExpediente`: `r.FormValue` → `r.PostFormValue` | Seguridad (solo POST) |
| 46 | `handler.go` | `handleCSV`: soporta `?modulo=...` | Hardcode a expedientes |
| 47 | `handler.go` | `handleExportarExcel`: `f.Close()` + log error escritura | Resource leak |
| 48 | `handler.go` | `handleCSV`: verifica error de `w.Write()` | Error silencioso |
| 49 | `templates/index.html` | `dbPath` escapado con `jsonEncode` | XSS |
| 50 | `templates/index.html` | `convertirMoneda`: `try/finally` para liberar lock | Lock infinito en error |
| 51 | `templates/index.html` | Referencia a `ruta-procesos-data.js` eliminada | Archivo eliminado |


## Auditorías de Código (Julio 2026)

Se recibieron 3 auditorías externas (Qwen, Kimi, GLM) con un total de ~70 hallazgos. Muchos "críticos" eran falsos positivos porque los auditores solo analizaron `01_master_control_docs_presidencia.sql` sin considerar `02_modulos_adicionales.sql` (8 módulos completos) ni `03_ruta_procesos.sql`.

### Correcciones aplicadas

| # | Hallazgo | Archivo | Fix |
|---|----------|---------|-----|
| 1 | `parseSpanishNumber` corrompía campos textuales (observaciones) | `handler.go` | Whitelist `columnasNumericas` — solo se aplica a columnas monetarias/numéricas |
| 2 | `sanitizarOrden` rechazaba `fecha_creacion`/`fecha_actualizacion` | `app.go` | `columnasOrdenValidas` incluidas en validación |
| 3 | `convertirMoneda` sobrescribía presupuesto al editar adjudicación | `templates/index.html` | Lógica de presupuesto vs adjudicación en bloques independientes |
| 4 | `toggleFrecuente` inyección JS por comillas en solped | `templates/tabla_expedientes.html`, `index.html` | Migrado a `data-df-*` attributes + dataset |
| 5 | `hxGetFormulario` no pasaba módulo | `templates/index.html` | Parámetro `modulo` en URL; items fijados guardan `{id, solped, modulo}` |
| 6 | `EliminarFila` no limpiaba `ruta_procesos_*` (FK constraint) | `app.go` | DELETE en cronograma + procesos antes del DELETE principal |
| 7 | Trigger sobrescribía `id_estatus` manual del usuario | `01_master...sql` | `AND OLD.fecha_firma_contrato IS NULL` en UPDATE a FIRMADO |
| 8 | `truncate` rompía UTF-8 multi-byte | `handler.go` | `[]rune(s)[:n]` en vez de `s[:n]` |
| 9 | `formatNumGo` error de precisión float (1,14 en vez de 1,15) | `handler.go` | `math.Round(f*100)/100` |
| 10 | `handleCSV` orden de columnas aleatorio (map iteration) | `handler.go` | `sort.Strings(headers)` |
| 11 | `fecha_actualizacion` inconsistente (CURRENT_DATE vs TIMESTAMP) | `02_modulos...sql` | Triggers cambiados a `CURRENT_TIMESTAMP` |
| 12 | `EliminarRutaProceso` sin transacción | `app.go` | Envuelto en `tx.Begin/Commit/Rollback` |
| 13 | `LastInsertId()` error ignorado | `app.go` | Error chequeado explícitamente |
| 14 | Errores scan silenciados en `ObtenerRutaProcesosData` | `app.go` | `log.Printf` en cada scan fallido |
| 15 | Items fijados perdían módulo al recargar | `templates/index.html` | `modulo` persistido en localStorage junto a id/solped |
| 16 | Label incorrecto "Asunto del Memorándum" en recobros | `templates/form_recobros.html` | Cambiado a "Asunto del Recobro" |
| 17 | `fecha_devuelto` consultado pero no mostrado en historial | `templates/historial.html` | Columna agregada al template |
| 18 | `dayNames` duplicado "M" para Lunes y Miércoles | `app.go` | "L", "M", **"X"**, "J", "V" |
| 19 | `pushModal` fuga de event listeners por apertura | `templates/index.html` | Listener único global con delegación de eventos |
| 20 | `crearBackup` corrupción de bak.1 si el sistema crashea durante la copia | `app.go` | Copia primero a `bak.tmp`, renombra solo si exitoso |
| — | — | — | — |
| 21 | `formatNumGo` pérdida de signo negativo en `-0.50` → `"0,50"` | `handler.go` | Signo capturado con `rounded < 0` antes de truncar a int64 |
| 22 | `safeHTML`/`safeJS`/`safeURL` desactivaban escape automático (XSS) | `handler.go` | Eliminadas del FuncMap |
| 23 | `id_estatus` se insertaba NULL al enviar vacío (bypasea DEFAULT 1) | `app.go` | UPDATE dinámico: solo incluye columnas con valor no-nulo |
| 24 | UPDATE sobrescribía campos vacíos con NULL (pérdida de datos) | `app.go` | SET dinámico solo para columnas no-nulas |
| 25 | `DescargarBD` sin WAL checkpoint → copia inconsistente | `app.go` | `PRAGMA wal_checkpoint(TRUNCATE)` antes de copiar |
| 26 | `parseSpanishNumber` convertía `1.23` → `123` | `handler.go` | Solo procesa si contiene coma (formato español) |
| 27 | `queryRows` devolvía `nil` en templates → `"<nil>"` visible | `app.go` | Default `nil` → `""` |
| 28 | `rows.Err()` no chequeado en `ObtenerCatalogos`, `ObtenerExpedientesDisponiblesRuta` | `app.go` | `rows.Err()` post-iteración |
| 29 | `window.PAGE_DATA.ActiveModule` no existía → JS siempre `'expedientes'` | `templates/index.html` | `ActiveModule` agregado al objeto JS |
| 30 | `handleCSV` construía CSV manualmente (comillas/saltos de línea frágiles) | `handler.go` | Migrado a `encoding/csv` |
| 31 | `columnasOrdenValidas` permitía `id_expediente` cross-module (error 500) | `app.go` | Reducido solo a `fecha_creacion`, `fecha_actualizacion` |
| 32 | Gantt 60 días calendario (~42 hábiles) vs 60 hábiles esperados | `app.go` | `bizTarget=60` iteración hasta alcanzar |
| 33 | `ObtenerRutaProcesosData` loop O(n×m) en match cronograma | `app.go` | `map[int]*Proceso` O(1) |
| 34 | Falta `_busy_timeout` en DSN → posibles `database is locked` | `app.go` | `_busy_timeout=5000` |
| 35 | `go 1.25.0` en go.mod (no existe como toolchain) | `go.mod` | `go 1.23.0` |
| 36 | Error ignorado en `handleCargarExpediente` | `handler.go` | Log agregado |

## Migración a Go html/template — Estado

| # | Paso | Estado | Detalle |
|---|------|--------|---------|
| 1 | **Datos precargados en PageData** | ✅ Hecho | `handler.go` — `PageData` inyecta catálogos y filas (multi-módulo). El template renderiza la tabla con `{{range}}`. |
| 2 | **Rutas API en el handler** | ✅ Hecho | `handler.go` — 11 rutas `/api/*` para CRUD, abrir BD, historial, ruta procesos, pendientes, CSV, catálogos, VACUUM. |
| 3 | **Reemplazar bindings JS** | ✅ Hecho | `templates/index.html` — `fetch()` y luego `htmx` reemplaza `window.go.main.App.*`. Solo queda 1 binding Wails: `AbrirDialogoBD`. |
| 4 | **HTMX** | ✅ Hecho | Integrado en plantillas y handler. Las vistas parciales renderizan HTML fragmentado reactivamente sin gluecode JS. |

### Tercera ronda (Julio 2026)

En esta ronda se recibieron 3 nuevas auditorías independientes (~70 hallazgos combinados). De ellos, solo estos eran válidos (el resto eran falsos positivos o ya corregidos):

| # | Hallazgo | Archivo | Fix |
|---|----------|---------|-----|
| 37 | JS destructuring con mayúsculas: `Legend`/`Processes` vs json tags `legend`/`processes` | `templates/ruta_procesos.html` | `const { legend, columns: ganttColumns, processes } = data;` |
| 38 | INSERT con `id_estatus = NULL` explícito bypassa `DEFAULT 1` (registros nuevos sin estatus) | `app.go` | INSERT dinámico: solo incluye columnas con valor no-nulo |
| 39 | `AgregarRutaProceso` inserta `db_id = 0` violando FK (no existe expediente 0) | `app.go` | `if dbID > 0 { dbIDVal = dbID }` — nil en caso contrario |
| 40 | Race condition: `a.db`/`a.dbPath` leídos sin `a.mu` en handlers | `handler.go` | Adquirir `a.mu.RLock()` antes de leer `db`/`dbPath` |
| 41 | `crearBackup` posible nil dereference: chequea `dbPath` pero no `db == nil` | `app.go` | `if a.dbPath == "" \|\| a.db == nil { return nil }` |
| 42 | Click fuera del modal no cierra: `closest('.modal')` nunca encuentra clase | `templates/index.html` | Clase `.modal` agregada a todos los contenedores de modal |
| 43 | `filterColMap` en exportar-excel solo cubre columnas de expedientes (otros módulos excluyen todas las filas) | `handler.go` | Si `row[rowKey]` no existe, saltar filtro |
| 44 | `PAGE_DATA.modulos` filtra `QueryHistorial` (SQL interno) al cliente | `handler.go` | Copia de Modulos sin `QueryHistorial` para frontend |
| 45 | `fecha_recibido` vacío comparado con fechas (`"" < "2024-01-01" = true`) | `handler.go` | `fr != ""` antes de comparar |
| 46 | Campos `nil` omitidos en UPDATE: usuario no puede limpiar/borrar campos | `app.go` | `col = NULL` incluido en SET cuando `vals[i] == nil` |
| 47 | `nota` NULL en cronograma causa `rows.Scan` fail → fila descartada (pérdida silenciosa) | `app.go` | `nota` cambiado de `string` a `sql.NullString` |
| 48 | `trg_exp_auditoria` no actualiza `fecha_actualizacion` (los otros 8 módulos sí) | `01_master_control_docs_presidencia.sql` | Agregado `UPDATE expedientes SET fecha_actualizacion = CURRENT_DATE` |
| 49 | `handleCSV` ignora filtros: exporta dataset completo sin respetar fecha/catálogo/gerencia | `handler.go` | Misma lógica de filtrado que `handleExportarExcel` |
| 50 | Templates legados `formulario.html` y `tabla_filas.html` cargados sin uso | `templates/`, `Makefile` | Eliminados |
| 51 | Gantt timeline keys RFC3339 vs "YYYY-MM-DD": todas las celdas del cronograma vacías | `app.go` | `strftime('%Y-%m-%d', c.fecha)` en la query |
| 52 | `convertirMoneda()` lee `dataset.raw` desactualizado (1 keystroke atrás) | `index.html` | `_parseValue()` lee desde `el.value` directamente |
| 53 | Doble fila de historial en INSERT de expedientes | `01_*.sql` | Temp table `_skip_audit` + `WHEN` en trigger UPDATE |
| 54 | Filas con `gerencia=""` se ven en pantalla pero desaparecen de CSV/Excel | `handler.go` | `gerName == "" \|\| permitidasNames[gerName]` en exports |
| 55 | Falta FK y UNIQUE en `ruta_procesos_cronograma` | `03_ruta_procesos.sql` | FK `id_proceso` + `UNIQUE(id_proceso, fecha)` |
| 56 | Estatus resuelto por `nombre` en triggers (vulnerable a rename) | `01_*.sql` | IDs literales: `1 = PENDIENTE`, `2 = FIRMADO` |
| 57 | 11 handlers con extracción de módulo duplicada | `handler.go` | Helper `moduloDesdeRequest(r)` + const `moduloDefault` |
| 58 | Export filter pipeline duplicado (-89 LOC) entre CSV y Excel | `handler.go` | Helper `filasParaExportar(r)` + `exportFilterColMap` |
| 59 | `fecha_actualizacion`: CURRENT_TIMESTAMP vs CURRENT_DATE inconsistente (9 módulos vs Go) | SQL + `app.go` | Unificado a `CURRENT_DATE` en todos |
| 60 | Scripts SQL no idempotentes; `Tablas8.sql` legado | `data/sql/` | `IF NOT EXISTS` en CREATEs; renombrado a `.legacy` |
| 61 | `withTx` helper DRY: boilerplate de transacción duplicado | `app.go` | `withTx(func(tx *sql.Tx) error) error` |
| 62 | `_skip_audit` TEMP TABLE: conexiones del pool no comparten tablas TEMP → INSERT falla en producción | `01_*.sql` | Cambiado de `CREATE TEMP TABLE` a `CREATE TABLE` regular |

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
| `/api/ruta-procesos-agregar` | POST | Agrega un proceso a la ruta (vinculado a un expediente existente) |
| `/api/ruta-procesos-toggle` | POST | Activa/desactiva un proceso en la ruta |
| `/api/ruta-procesos-eliminar` | POST | Elimina un proceso de la ruta |
| `/api/ruta-procesos-expedientes` | GET | Devuelve JSON con expedientes disponibles para agregar como procesos |
| `/api/pendientes` | GET | Devuelve fragmento HTML de documentos pendientes |
| `/api/guardar-catalogo` | POST | Agrega registro a un catálogo |
| `/api/optimizar-bd` | POST | Ejecuta VACUUM |
| `/api/csv` | GET | Descarga CSV del módulo indicado (`?modulo=...`) |
