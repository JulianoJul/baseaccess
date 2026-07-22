# Catálogo de Funciones (SPOT)

Fuente única de verdad de la lógica existente. Antes de crear una nueva función, **revisar si ya existe** para evitar duplicación (DRY).

---

## Backend Go — Métodos exportados (app.go)

| Método | Parámetros | Descripción |
|--------|-----------|-------------|
| `Startup(ctx)` | `ctx`: context.Context | Inicializa el contexto de la app (Wails lifecycle) |
| `AbrirBaseDatos(filePath)` | `filePath`: ruta al .db | Abre BD SQLite con WAL + foreign_keys. Cierra la anterior si existe |
| `CerrarBaseDatos()` | — | Cierra la BD actual |
| `OptimizarBD()` | — | Ejecuta VACUUM con backup WAL checkpoint previo |
| `DescargarBD(destPath)` | `destPath`: string | Copia el .db actual a otra ruta |
| `AbrirDialogoBD()` | — | Abre diálogo nativo Wails (`runtime.OpenFileDialog`) para seleccionar .db |
| `GuardarDialogoBD(nombreDefault)` | `nombreDefault`: string | Abre diálogo nativo Wails (`runtime.SaveFileDialog`) para guardar copia |
| `SetBackupMaxCopies(n)` | `n`: int | Configura número de backups rotativos (1-20). Thread-safe via `atomic.Int64` |
| `GetBackupMaxCopies()` | — | Retorna número actual de backups. Thread-safe |
| `ObtenerFilas(moduloKey, orden)` | `moduloKey`: key de Modulos map; `orden`: columna DESC/ASC | SELECT * FROM cfg.Vista con sanitizacion whitelist. Soporta 9 modulos. |
| `ObtenerFilasPaginado(moduloKey, orden, pagina, pageSize)` | `moduloKey`: key de Modulos; `orden`: col; `pagina`: int; `pageSize`: int | SELECT * con LIMIT y OFFSET. Retorna `[]Row`, `totalPages` y `error` |
| `ObtenerFilaPorId(moduloKey, id)` | `moduloKey`, `id`: int | Retorna `Row` unica o error |
| `GuardarFila(moduloKey, data)` | `moduloKey`, `data`: map[string]interface{} | INSERT o UPDATE segun presencia de cfg.IDColumna. Backup WAL checkpoint automático |
| `EliminarFila(moduloKey, id)` | `moduloKey`, `id`: int64 | DELETE en transacción (historial + modulo) con Rollback condicional post-Commit |
| `ObtenerHistorialFila(moduloKey, id)` | `moduloKey`, `id`: int | SELECT cfg.QueryHistorial (JOIN multi-tabla segun modulo) |
| `ObtenerColumnasVista(vista)` | `vista`: string (validada contra whitelist) | Retorna nombres de columna de una vista SQL |
| `ObtenerCatalogos()` | — | Retorna map[string][]CatalogoItem (11 catalogos) |
| `GuardarNuevoCatalogo(tabla, nombre, extra)` | `extra`: map con col/val opcional | INSERT en tabla catalogo (whitelist tabla/columna) |
| `ObtenerDocumentosPendientes()` | — | WHERE estatus <> FIRMADO (especifico de expedientes) |
| `ObtenerRutaProcesos()` | — | JOIN completo para ruta de procesos (especifico de expedientes) |
| `ObtenerRutaProcesosData(idHoja, idJunta)` | `idHoja`: int; `idJunta`: int | Retorna datos completos del Gantt para una hoja/junta específica |
| `CrearRutaProcesosHoja(nombre)` | `nombre`: string | Crea una hoja nueva en el Gantt |
| `EliminarRutaProcesosHoja(id)` | `id`: int | Elimina una hoja y todos sus procesos |
| `CrearRutaProcesosJunta(idHoja, numero, consecutiva, fecha)` | `idHoja`, `numero`, `consecutiva`: int; `fecha`: string | Crea una junta dentro de una hoja |
| `ActualizarRutaProcesosJunta(id, numero, consecutiva, fecha)` | `id`, `numero`, `consecutiva`: int; `fecha`: string | Actualiza número, consecutiva y fecha de una junta |
| `EliminarRutaProcesosJunta(id)` | `id`: int | Elimina una junta |
| `AgregarRutaProcesosSemana(idJunta, numero, fechaInicio, fechaFin)` | `idJunta`, `numero`: int; `fechaInicio`, `fechaFin`: string | Agrega una semana a una junta |
| `EliminarRutaProcesosSemanas(idJunta, numeros)` | `idJunta`: int; `numeros`: []int | Elimina semanas de una junta por número |
| `AgregarRutaProcesosProceso(idJunta, unusedNumero, proceso)` | `idJunta`: int; `unusedNumero`: int; `proceso`: string | Agrega un proceso a una junta |
| `EliminarRutaProcesosProceso(id)` | `id`: int | Elimina un proceso |
| `ReordenarRutaProcesosProceso(idJunta, idProceso, direction)` | `idJunta`, `idProceso`, `direction`: int | Reordena procesos (1=arriba, -1=abajo) |
| `CrearRutaProcesosLeyenda(nombre, color, ambito, idHoja, idJunta)` | `nombre`, `color`, `ambito`: string; `idHoja`, `idJunta`: *int | Crea leyenda (global, por hoja o por junta) |
| `ActualizarRutaProcesosLeyenda(id, nombre, color)` | `id`: int; `nombre`, `color`: string | Actualiza nombre y color de una leyenda existente |
| `EliminarRutaProcesosLeyenda(id)` | `id`: int | Elimina una leyenda |
| `ReordenarRutaProcesosLeyenda(idJunta, idLeyenda, direction)` | `idJunta`, `idLeyenda`, `direction`: int | Reordena leyendas (1=arriba, -1=abajo) |
| `ToggleBloquearRutaProcesosLeyenda(id)` | `id`: int | Activa/desactiva una leyenda |
| `GuardarCronogramaDia(idProceso, fecha, idLeyenda, nota)` | `idProceso`: int; `fecha`: string; `idLeyenda`: int; `nota`: string | Guarda/actualiza/elimina un día en el cronograma Gantt |
| `EliminarCronogramaDia(id)` | `id`: int | Elimina un día del cronograma |

## Frontend Alpine.js (`frontend/new/vendor/`)

| Archivo | Propósito |
|---------|-----------|
| `alpine-app.js` | Stores + Alpine.data(): modales, fijados, recientes, sumas, exportar, formulario, appShell |
| `alpine-directives.js` | Directiva custom `x-currency` (formato numérico ES) |
| `alpine-htmx-bridge.js` | Puente: `Alpine.initTree()` tras HTMX swap + sincronización pines |

**Stores globales:** `modals` (abrir/cerrar/stack), `toast` (success/error/info)
**Componentes Alpine:** `appShell`, `bdRecientes`, `fijados`, `exportarExcel`, `calculadoraSumas`, `formularioModulo`, `filtroSuperintendencias`
**JS vainilla mínimo:** `abrirBaseDatos()` en `index.html` (Wails dialog, no reemplazable por Alpine)
**Helpers:** `toast()` vía `window.Alpine.store('toast')`
