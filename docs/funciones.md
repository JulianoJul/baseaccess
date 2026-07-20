# Catálogo de Funciones (SPOT)

Fuente única de verdad de la lógica existente. Antes de crear una nueva función, **revisar si ya existe** para evitar duplicación (DRY).

---

## Backend Go — Métodos exportados (app.go)

| Método | Parámetros | Descripción |
|--------|-----------|-------------|
| `AbrirBaseDatos(filePath)` | `filePath`: ruta al .db | Abre BD SQLite con WAL + foreign_keys. Cierra la anterior si existe |
| `CerrarBaseDatos()` | — | Cierra la BD actual |
| `ObtenerFilas(moduloKey, orden)` | `moduloKey`: key de Modulos map; `orden`: columna DESC/ASC | SELECT * FROM cfg.Vista con sanitizacion whitelist. Soporta 9 modulos. |
| `ObtenerFilasPaginado(moduloKey, orden, pagina, pageSize)` | `moduloKey`: key de Modulos; `orden`: col; `pagina`: int; `pageSize`: int | SELECT * con LIMIT y OFFSET. Retorna `[]Row`, `totalPages` y `error` |
| `ObtenerFilaPorId(moduloKey, id)` | `moduloKey`, `id`: int | Retorna `Row` unica o error |
| `ObtenerRutaProcesos()` | — | JOIN completo para ruta de procesos (especifico de expedientes) |
| `ObtenerDocumentosPendientes()` | — | WHERE estatus <> FIRMADO (especifico de expedientes) |
| `ObtenerHistorialFila(moduloKey, id)` | `moduloKey`, `id`: int | SELECT cfg.QueryHistorial (JOIN multi-tabla segun modulo) |
| `GuardarFila(moduloKey, data)` | `moduloKey`, `data`: map[string]interface{} | INSERT o UPDATE segun presencia de cfg.IDColumna. UPDATE devuelve el id real (no `LastInsertId()=0`). Id como `int64`. Backup WAL checkpoint automático |
| `EliminarFila(moduloKey, id)` | `moduloKey`, `id`: int64 | DELETE en transacción (historial + modulo) con Rollback condicional post-Commit |
| `GuardarCronogramaDia(idProceso, fecha, idLeyenda, nota)` | `idProceso`: int; `fecha`: string; `idLeyenda`: int; `nota`: string | Guarda/actualiza/elimina un día en el cronograma Gantt |
| `ObtenerCatalogos()` | — | Retorna map[string][]CatalogoItem (11 catalogos) |
| `OptimizarBD()` | — | Ejecuta VACUUM con backup WAL checkpoint previo |
| `GuardarNuevoCatalogo(tabla, nombre, extra)` | `extra`: map con col/val opcional | INSERT en tabla catalogo (whitelist tabla/columna) |
| `AbrirDialogoBD()` | — | Abre diálogo nativo Wails (`runtime.OpenFileDialog`) para seleccionar .db |
| `GuardarDialogoBD(nombreDefault)` | `nombreDefault`: string | Abre diálogo nativo Wails (`runtime.SaveFileDialog`) para guardar copia |
| `ObtenerExpedientesDisponiblesRuta()` | — | Retorna JSON con expedientes no agregados aún a Ruta Procesos |
| `ObtenerRegistrosDisponiblesRuta(modulo)` | `modulo`: string | Retorna JSON con registros disponibles del módulo indicado para agregar como procesos |
| `ActualizarRutaProcesosLeyenda(id, nombre, color)` | `id`: int, `nombre`, `color`: string | Actualiza nombre y color de una leyenda existente |
| `ObtenerColumnasVista(vista)` | `vista`: string (validada contra whitelist) | Retorna nombres de columna de una vista SQL |
| `SetBackupMaxCopies(n)` | `n`: int | Configura número de backups rotativos (1-20). Thread-safe via `atomic.Int64` |
| `GetBackupMaxCopies()` | — | Retorna número actual de backups. Thread-safe |
| `DescargarBD(destPath)` | `destPath`: string | Copia el .db actual a otra ruta |

## Frontend JS (`frontend/vendor/app.js`)

**Modales:** `mostrarFormulario`, `cerrarFormulario`, `pushModal`, `cerrarModal`, `cerrarSiOverlay`
**Paginación (Servidor):** Migrada al servidor en la nueva estructura (usando el helper Go `pagRange` en `handler.go` y controles HTMX en `tabla.html`). Las funciones cliente `renderPaginacion`, `irPagina`, y `aplicarPaginacion` fueron eliminadas de la UI moderna.
**Exportar:** `abrirModalExportar`, `cerrarModalExportar`, `cargarColumnasExportar`, `ejecutarExportar`, `toggleTodasColumnas`, `filtrarSuperintendenciasExportar`
**Fijados (localStorage):** `toggleFrecuente`, `abrirFrecuentes`, `cerrarFrecuentes`
**BD Recientes:** `abrirRecientes`, `cerrarRecientes`, `registrarReciente`, `eliminarReciente`
**Sumas:** `abrirSumas`, `cerrarSumas`, `anyadirFilaSuma`, `calcularSumas`, `limpiarSumas`
**Conversión USD/Bs:** `calcularBs`
**BD:** `abrirBaseDatos`, `optimizarBD`, `descargarBDError`
**Helpers:** `toast`, `mostrarSpinner`, `ocultarSpinner`, `esc` (escapeHtml)
