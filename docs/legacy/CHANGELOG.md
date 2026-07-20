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
| 52 | `app.go`, `handler.go`, `templates/*` | **Multi-modulo en Ruta Procesos**: columna `modulo` en `ruta_procesos_procesos`. Nuevo endpoint `/api/ruta-procesos-registros?modulo=xxx`. Selector de módulo en "Añadir Proceso". | Se pueden agregar procesos de cualquier módulo al Gantt |
| 53 | `handler.go`, `index.html` | **Fix botón Nuevo Registro**: `hx-include` reemplazado por `hx-vals='js:{...}'` y `location.reload()` cambiado a `htmx.ajax` recargando solo la tabla del módulo activo | El formulario se carga y guarda en el módulo correcto sin recargar toda la página |
| 54 | `data/sql/01_master_control_docs_presidencia.sql` | Todos los `INSERT INTO` cambiados a `INSERT OR IGNORE INTO` | Idempotencia al reabrir BD desde Recientes |
| 55 | `data/sql/02_modulos_adicionales.sql` | Fix: `vw_reporte_recobros` faltaba `LEFT JOIN cat_documento` | Vista de recobros ahora funciona |
| 56 | `app.go` | SQL files embebidos via `//go:embed data/sql/*.sql` en vez de `os.ReadFile` | Portabilidad: no depende del directorio de trabajo |
| 57 | `templates/ruta_procesos.html`, `handler.go`, `app.go` | Leyendas clickeables: modal de edición de nombre y color. Nuevo endpoint `/api/ruta-procesos-leyenda-actualizar` | El usuario puede editar leyendas existentes |
| 58 | `app.go`, `data/sql/03_ruta_procesos.sql` | Leyendas ordenadas alfabéticamente con colores distintivos de alto contraste | Mejor legibilidad del Gantt |
| 59 | `handler.go` | Cache-Control headers añadidos a rutas `/` y `/api/*` | Prevenir caché de respuestas HTML/JSON |
| 60 | `templates/index.html` | `hx-indicator="#spinner-overlay"` y `history.replaceState(null, '', '?modulo={{$key}}')` en botones de módulo | Mostrar spinner al cambiar módulo + persistir módulo en URL |
| 61 | `frontend/vendor/styles.css` | Clases `.w-8`, `.h-8`, `.rounded-full`, `.border-gray-600`, `.shrink-0` añadidas | Preview de color en modales de leyenda |
| 62 | `templates/ruta_procesos.html` | Labels "Color HEX" → "Color" + preview circular con actualización en vivo | UX editor/creador de leyendas |
| 63 | `docs/doc.md`, `docs/ai-context.md` | Sección "Bugs Conocidos" documentada | Transparencia sobre bugs de frontend |
| 64 | `templates/components.html` | Revertido `htmx.ajax` → `location.reload()` para guardar/eliminar | `htmx.ajax` rompía persistencia de datos |


## Auditorías de Código (Julio 2026)
