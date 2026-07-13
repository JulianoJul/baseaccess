# Architecture Decision Records (ADR)

Registro cronológico de decisiones técnicas tomadas en el proyecto.

---

## DEC-001: Migración a Wails v2 (Go backend + Web frontend)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** La app original tenía 3 runtimes (Electron/sql.js, Tauri/Rust, navegador). Electron usaba sql.js (WASM) que obligaba a exportar la BD completa en cada guardado. Tauri/Rust era complejo para la IA. Se migró a Wails v2: backend Go nativo con mattn/go-sqlite3, frontend web embebido vía `go:embed`. Un solo runtime, escrituras directas al .db, bindings automáticos Go↔JS.
- **Alternativas evaluadas:**
  - Electron + better-sqlite3 — menos complejo que Wails, pero sigue requiriendo IPC, preload.js, contextBridge
  - Tauri (Rust) — el más complejo para la IA (lifetimes, genéricos, builds lentos)
  - Wails v2 (Go) — elegido: Go simple, bindings automáticos, sin IPC boilerplate
- **Impacto:**
  - Rama `wails-migration` creada (rama paralela, `master` intacto)
  - `main.go`: entry point Wails con `go:embed` para `frontend/`
  - `app.go`: App struct con 12 métodos exportados (AbrirBaseDatos, CRUD, catálogos, VACUUM)
  - `frontend/index.html`: copia de `src/index.html` adaptada — sql.js reemplazado por `window.go.main.App.*`
  - `frontend/schema-config.js`: idéntico al original
  - `go.mod` + `wails.json`: config proyecto Wails
  - `main.js`, `src/`, `src-tauri/`, `package.json`: legacy sin eliminar (master los conserva)
  - `.gitignore`: `build/bin/` añadido para outputs Wails
  - `Makefile`: targets `wails-*`
  - `.github/workflows/build.yml`: job `wails` (Linux + Windows)
  - Windows: WebView2 Fixed Version Runtime 150.0.4078.65 portable via `windows.Options.WebviewBrowserPath`

---

## DEC-002: Límite de 100MB en Drag & Drop

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Archivos SQLite grandes (>100MB) saturan recursos. Se mantuvo el límite del legacy.
- **Impacto:** Validación en `frontend/index.html`.

---

## DEC-003: WebView2 Fixed Version Runtime Portable (Windows)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Wails requiere WebView2 en Windows. Para 100% portabilidad (sin instalación, sin internet), se incluye el Fixed Version Runtime junto al .exe.
- **Alternativas evaluadas:**
  - `-webview2 embed` — incluye bootstrapper pero requiere internet para descargar runtime
  - `-webview2 download` — igual, requiere internet
  - Fixed Version Runtime — elegido: 100% offline, portátil en USB
- **Impacto:**
  - `main.go`: `windows.Options.WebviewBrowserPath` apunta al directorio del runtime
  - `build.yml`: download + extract del CAB (Microsoft), cacheado, copiado a `build/bin/`

---

## DEC-004: Escrituras Directas SQLite + Backup Rotativo en Go

- **Origen:** `[Derivado de DEC-001]`
- **Contexto y Causa:** En sql.js, cada guardado exportaba la BD completa (`db.export()`) y sobreescribía el archivo. En Wails, Go escribe directamente al .db vía `database/sql`. Sin embargo, hay cortes de energía frecuentes, por lo que se implementa backup rotativo en Go antes de cada escritura (`.bak.1` a `.bak.N` con N configurable).
- **Impacto:** `app.go`: función `crearBackup()` con rotación, llamada antes de GuardarExpediente, EliminarExpediente, GuardarNuevoCatalogo, OptimizarBD. No se necesita `guardarBD()` ni autosave.

---

## DEC-005: Descargar Copia de BD desde Error Boundary

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El error boundary tenía `descargarBDError()` como stub. Se implementó funcionalidad real para permitir al usuario guardar una copia del .db actual cuando ocurre un error crítico.
- **Alternativas evaluadas:**
  - Mantener stub — descartado: el usuario quería funcionalidad real
  - Diálogo nativo Wails para guardar archivo — elegido: usa `window.runtime.SaveFileDialog()`
- **Impacto:** `descargarBDError()` ahora abre diálogo de guardado y copia el .db vía Go.

---

## DEC-006: Sin Autoguardado (escrituras inmediatas)

- **Origen:** `[Derivado de DEC-001]`
- **Contexto y Causa:** En sql.js, los cambios estaban solo en RAM hasta que se exportaba la BD. En Wails, cada INSERT/UPDATE/DELETE escribe directamente al archivo .db. El autoguardado cada 30s ya no tiene sentido.
- **Impacto:** `AUTOSAVE_ENABLED` y `AUTOSAVE_INTERVAL_MS` eliminados de `schema-config.js`.

---

## DEC-007: Diálogos nativos desde Go (no desde JS)

- **Origen:** `[Bug reportado por usuario]`
- **Contexto y Causa:** En Wails v2, `window.runtime.OpenFileDialog`/`SaveFileDialog` NO existen en el runtime JS. El frontend llamaba a esos métodos inexistentes, caía al catch silenciosamente, y nunca abría el explorador de archivos.
- **Alternativas evaluadas:**
  - Usar `<input type="file">` con fallback — descartado: no hay APIs de archivo nativo en WebKit para rutas absolutas
  -Implementar diálogos nativos en Go — elegido: Wails expone `runtime.OpenFileDialog(ctx, opts)` en Go
- **Impacto:**
  - `app.go`: añadidos `AbrirDialogoBD()` y `GuardarDialogoBD(nombreDefault)` envolviendo `wailsRuntime.OpenFileDialog`/`SaveFileDialog`
  - `frontend/index.html`: detección `window.go.main.App` → usa bindings Go; fallback a `window.runtime` en navegador

---

## DEC-008: DevTools habilitados en builds debug

- **Origen:** `[Bug reportado por usuario]`
- **Contexto y Causa:** Para depurar el frontend en Wails (WebKitGTK), se necesita acceso a DevTools (F12). Por defecto Wails los deshabilita en producción.
- **Impacto:** `main.go`: `EnableDefaultContextMenu: true`, `Debug: options.Debug{OpenInspectorOnStartup: false}`. Makefile: `wails-build-linux` usa `-debug`, `wails-build-linux-prod` sin flag.

---

## DEC-009: Utilidades Tailwind emuladas en styles.css

- **Origen:** `[Bug reportado por usuario]`
- **Contexto y Causa:** `tailwind.min.css` es un build purgado que sólo incluye clases usadas en el HTML escaneado. Clases con opacidad (`bg-gray-700/40`, `border-gray-700/60`) y colores no escaneados (`border-gray-800`, `border-red-700`) NO existen. Resultado: bordes blancos visibles (preflight deja `border-color: #e5e7eb` por defecto).
- **Alternativas evaluadas:**
  - Regenerar tailwind.min.css con content scanning actualizado — descartado: requiere Node.js + configuración Tailwind
  - Migrar a clases existentes — descartado: muchas不知/a
  - Emular clases faltantes en styles.css — elegido: SoC, centraliza el fix
- **Impacto:** `frontend/vendor/styles.css`: añadidas ~20 utilidades (`.bg-gray-700\/10`, `.border-gray-800`, etc.) con valores `rgba()` equivalentes.

---

## DEC-010: Fechas de migración Excel trackeadas por solped

- **Origen:** `[Bug reportado por usuario]`
- **Contexto y Causa:** El trigger `trg_exp_auditoria` (Tablas8.sql) fuerza `fecha_actualizacion = CURRENT_DATE` en cada UPDATE. Durante la migración, los duplicados de solped disparaban UPDATEs que sobreescribían con la fecha de hoy. Además, `fecha_creacion` era pisada con la `fecha_recibido` más nueva en lugar de la más antigua.
- **Impacto:** `data/importar_datos.py`:
  - `DROP TRIGGER trg_exp_auditoria` al inicio (recreado al final)
  - Dict `solped_fechas` trackea `MIN(fecha_recibido)` y `MAX(fecha_devuelto or fecha_recibido)` por solped
  - UPDATE final aplica esos valores
  - `fecha_creacion` = nacimiento del expediente, `fecha_actualizacion` = último movimiento

---

## DEC-011: Go html/template via custom Handler en AssetServer

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El usuario quería un stack más unificado (solo Go) sin tecnologías frontend separadas. En lugar de servir `frontend/index.html` como archivo estático, se reemplazó `Assets: assets` por `Handler: handler` en el AssetServer de Wails. El handler (`TemplateHandler`) usa `html/template` de Go para renderizar el HTML con datos inyectados desde Go.
- **Alternativas evaluadas:**
  - Seguir con `Assets: assets` (frontend estático) — descartado: JS sigue siendo tecnología separada
  - `webview_go` + html/template — descartado: problema de cross-compile CGO en Linux→Windows
  - Wails + Handler + Go templates — elegido: Wails sigue dando la ventana nativa, el HTML se genera desde Go
- **Impacto:**
  - `handler.go` creado: `TemplateHandler` (http.Handler), `//go:embed all:frontend` (estáticos), `//go:embed templates/*` (templates Go)
  - `templates/index.html` creado: template Go de 332 líneas con la estructura HTML completa
  - `main.go`: `Assets: assets` → `Handler: handler`
  - Los bindings Wails (`window.go.main.App.*`) siguen funcionando (JS los llama igual)
  - Wails inyecta su runtime JS automáticamente en respuestas HTML
  - `frontend/index.html` se mantiene como estático (el handler lo ignora para `/`)

---

## DEC-012: Rutas API REST en el handler + JS mínimo

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Tras DEC-011, el template Go ya generaba el HTML, pero el JS seguía llamando a bindings Wails (`window.go.main.App.*`) para toda operación de datos. Se quería eliminar el "glue code" JS y que el handler sirviera rutas API REST (`/api/*`) que el frontend consume con `fetch()`.
- **Alternativas evaluadas:**
  - Mantener bindings Wails — descartado: requiere JS glue code
  - HTMX para interactividad declarativa — postergado: `fetch()` + JS mínimo es suficiente
  - Rutas API REST en handler + `fetch()` — elegido: elimina bindings, mantiene JS mínimo, la IA escribe rutas Go con facilidad
- **Impacto:**
  - `handler.go`: 10 rutas `/api/*` (JSON) para CRUD, BD, historial, ruta procesos, pendientes, CSV, catálogos, VACUUM
  - `handler.go`: `PageData` ahora inyecta `Catalogs` y `Expedientes` precargados al template
  - `handler.go`: Funciones template (`default`, `rowGet`, `rowGetStr`, `rowGetNum`, `estatusClass`, `formatNum`, `jsonEncode`, `truncate`, `isSelected`)
  - `templates/index.html`: Reescrito — tabla con `{{range}}`, `<select>` con catálogos Go, JS reducido a `fetch()` + modales + apertura BD
  - `app.go`: `CatalogoItem` con `IDGerencia` para filtrar superintendencias por gerencia
  - Único binding Wails restante: `AbrirDialogoBD` (diálogo nativo de archivos del SO)
