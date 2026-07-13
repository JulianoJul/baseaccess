# Catálogo de Funciones (SPOT)

Fuente única de verdad de la lógica existente. Antes de crear una nueva función, **revisar si ya existe** para evitar duplicación (DRY).

---

## Backend Go — Métodos exportados (app.go)

| Método | Parámetros | Descripción |
|--------|-----------|-------------|
| `AbrirBaseDatos(filePath)` | `filePath`: ruta al .db | Abre BD SQLite con WAL + foreign_keys. Cierra la anterior si existe |
| `CerrarBaseDatos()` | — | Cierra la BD actual |
| `ObtenerExpedientes(orden)` | `orden`: columna DESC/ASC | SELECT con sanitización whitelist. Retorna `[]Row` |
| `ObtenerExpedientePorId(id)` | `id`: int | Retorna `Row` única o error |
| `ObtenerRutaProcesos()` | — | JOIN completo para ruta de procesos |
| `ObtenerDocumentosPendientes()` | — | WHERE estatus <> FIRMADO |
| `ObtenerHistorialCompleto(id)` | `id`: int | JOIN multi-tabla ordenado DESC |
| `GuardarExpediente(data)` | `data`: map[string]interface{} | INSERT o UPDATE según presencia de id_expediente |
| `EliminarExpediente(id)` | `id`: int64 | DELETE en transacción (historial + expediente) |
| `ObtenerCatalogos()` | — | Retorna map[string][]CatalogoItem (11 tablas) |
| `OptimizarBD()` | — | Ejecuta VACUUM |
| `GuardarNuevoCatalogo(tabla, nombre, extra)` | `extra`: map con col/val opcional | INSERT en tabla catálogo (whitelist tabla/columna) |
| `AbrirDialogoBD()` | — | Abre diálogo nativo Wails (`runtime.OpenFileDialog`) para seleccionar .db |
| `GuardarDialogoBD(nombreDefault)` | `nombreDefault`: string | Abre diálogo nativo Wails (`runtime.SaveFileDialog`) para guardar copia |
| `SetBackupMaxCopies(n)` | `n`: int | Configura número de backups rotativos (1-20) |
| `GetBackupMaxCopies()` | — | Retorna número actual de backups |

## Data Layer — Frontend JS (llama a Go)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `_cargarBaseDatosComun(filePath)` | `filePath`: string | Llama `AbrirBaseDatos`, luego `cargarCatalogos()` y `cargarDatos()` |
| `cargarCatalogos()` | — | Llama `ObtenerCatalogos()`, llena `catalogosCache`, repuebla selects |
| `cargarDatos()` | — | Llama `ObtenerExpedientes()`, renderiza tabla |
| `obtenerExpedientes(orden)` | `orden`: string | Wrapper async → `window.go.main.App.ObtenerExpedientes(orden)` |
| `obtenerExpedientePorId(id)` | `id`: int | Wrapper async → `window.go.main.App.ObtenerExpedientePorId(id)` |
| `obtenerHistorialPorId(id)` | `id`: int | Wrapper async → `window.go.main.App.ObtenerHistorialCompleto(id)` |
| `obtenerDatosReporteExcel()` | — | Llama `obtenerExpedientes('id_expediente DESC')` |
| `guardarExpedienteEnBd(id, data)` | `id`, `data` | Llama `GuardarExpediente(data)` |
| `eliminarExpedienteDeBd(id)` | `id`: int64 | Llama `EliminarExpediente(id)` |
| `guardarNuevoCatalogoEnBd(tabla, cols, vals)` | `tabla, cols[], vals[]` | Llama `GuardarNuevoCatalogo(...)` |

## UI Layer — Tabla Principal

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `renderizarTabla(lista)` | `lista[]`: array de objetos expediente | Renderiza tabla de 8 columnas + fila desplegable |
| `cambiarOrden()` | — | Lee selector de orden + dirección, recarga datos ordenados |
| `toggleSortDir()` | — | Alterna ASC/DESC, persiste en localStorage |
| `aplicarPaginacion()` | — | Calcula páginas, renderiza slice actual |
| `irPagina(n)` | `n`: número de página | Cambia página, refresca tabla |
| `renderPaginacion()` | — | Renderiza controles de paginación |
| `toggleDesplegable(id)` | `id`: expediente.id_expediente | Expande/colapsa fila desplegable |
| `toggleDetalle(prefix, id)` | `prefix`, `id` | Alterna visibilidad de detalle |
| `formatNum(v)` | `v`: número | Formatea con toLocaleString('es-VE') |
| `parseNum(v)` | `v`: string | Convierte a float |
| `validarFechas()` | — | Valida fechas del formulario |
| `validarFechasEntre(recibido, devuelto)` | fechas | Función pura, retorna `{valid, errorMsg}` |

## UI Layer — Formulario de Edición

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `mostrarFormulario(id?)` | `id`: opcional | Abre modal formulario (crear/editar) |
| `cancelarFormulario()` | — | Cierra modal, limpia formulario |
| `cargarExpediente(id)` | `id`: ID | Carga datos en formulario para edición |
| `calcularBs(origen?)` | `origen?`: 'usd' \| 'bs' | Conversión bidireccional USD↔Bs |
| `guardarExpediente()` | — | Valida, llama `guardarExpedienteEnBd`, recarga |
| `eliminarExpediente()` | — | Confirma, llama `eliminarExpedienteDeBd`, recarga |
| `marcarCamposEdicionFrecuente()` | — | Marca campos frecuentes con indicador |
| `generarObservacionAutomatica()` | — | Genera línea de observación automática |
| `previewObservacion()` | — | Combina parte automática + texto libre |
| `formatTiempoEjecucion(v)` | `v`: número/string | Aplica sufijo "DÍAS" |
| `parseTiempoEjecucionParaEdicion(v)` | `v`: string | Quita sufijo "DÍAS" |

## UI Layer — Modales Secundarios

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `abrirRutaProcesos()` | — | Modal Gantt-chart de ruta de procesos |
| `cerrarRutaProcesos()` | — | Cierra modal ruta |
| `abrirDocumentosPendientes()` | — | Modal con expedientes no FIRMADOS |
| `cerrarPendientes()` | — | Cierra modal pendientes |
| `abrirHistorialCompleto(id)` | `id`: ID | Modal historial de snapshots |
| `cerrarHistorialCompleto()` | — | Cierra modal historial |
| `cargarHistorialCompleto(id)` | `id`: ID | Consulta y renderiza historial |
| `getEstatusClass(estatus)` | `estatus`: string | Clase CSS del badge según estatus |

## UI Layer — Sidebar (Frecuentes)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `toggleFrecuente(id, solped)` | `id`, `solped` | Marca/desmarca frecuente en localStorage |
| `renderSidebar()` | — | Renderiza lista de frecuentes |
| `toggleSidebar()` | — | Colapsa/expande sidebar |
| `toggleModoOrdenForm()` | — | Alterna orden secciones/Excel en formulario |
| `aplicarOrdenExcel()` | — | Clona campos en grilla plana |
| `restaurarOrdenSecciones()` | — | Restaura orden agrupado |

## UI Layer — Catálogos

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `inicializarBotonesCatalogo()` | — | Asigna eventos a botones "+" |
| `abrirAgregarCatalogo(selectId)` | `selectId`: ID del `<select>` | Abre modal para nuevo registro |
| `cerrarAgregarCatalogo()` | — | Cierra modal agregar catálogo |
| `captureAndRestoreFormState(callback)` | `callback` | Captura valores, ejecuta callback, restaura |
| `guardarNuevoCatalogo()` | — | Inserta y actualiza selector |

## Utilidades y Mantenimiento

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `actualizarEstadoBD(msg)` | `msg`: string | Actualiza indicador visual de estado BD |
| `optimizarBD()` | — | Ejecuta VACUUM vía Go, reporta resultado |
| `exportarCSV()` | — | Exporta datos como CSV descargable |
| `descargarBDError()` | — | Abre diálogo para guardar copia del .db actual |
| `updateUIOnError()` | — | Deshabilita botones, añade badge solo-lectura |
| `abrirBaseDatos()` | — | Dispara `<input type="file">`, carga BD |
| `abrirBaseDatosReciente(path)` | `path`: string | Abre BD desde recientes |
| `mostrarMenuRecientes()` | — | Renderiza menú BD recientes |
| `registrarReciente(nombre, path)` | `nombre`, `path` | Guarda en localStorage |
| `eliminarReciente(path)` o `eliminarRecienteIndex(index)` | — | Elimina de recientes |
| `abrirRecientes()` | — | Modal con lista de BD recientes |

## Helpers

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `$(id)` | `id`: string | `document.getElementById(id)` |
| `toast(mensaje, tipo)` | `mensaje`, `tipo` | Notificación flotante auto-dismiss 3s |
| `mostrarSpinner(texto)` | `texto`: opcional | Overlay con spinner |
| `ocultarSpinner()` | — | Oculta overlay spinner |
| `validarForma()` | — | Retorna array de errores de validación |
| `renderCatalogSelect(selectId, catKey, selectValue)` | — | Puebla un select desde catálogo cacheado |
| `cerrarModalSiOverlay(e, closeFn)` | `e`, `closeFn` | Cierra modal si click fuera del contenido |
| `escapeHtml(text)` | `text`: string | Escapa HTML para prevenir XSS |

## Constantes Globales (schema-config.js)

| Constante | Descripción |
|-----------|-------------|
| `CONFIG` | `MAX_FILE_SIZE_BYTES`, `PAGE_SIZE`, `VACUUM_CONFIRM_THRESHOLD_MB`, etc. |
| `DEBUG` | Wrapper condicional de console: `DEBUG.log()`, `DEBUG.error()` |
| `MSG` | Mensajes de usuario centralizados |
| `STORAGE_KEYS` | Keys de localStorage |
| `SELECTORS` | IDs de elementos DOM |
| `MSG_EXTRA` | Mensajes de mantenimiento (VACUUM, errores, etc.) |
