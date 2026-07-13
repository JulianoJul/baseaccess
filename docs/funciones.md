# Catálogo de Funciones (SPOT)

Fuente única de verdad de la lógica existente. Antes de crear una nueva función, **revisar si ya existe** para evitar duplicación (DRY).

---

## Backend Go — Métodos exportados (app.go)

| Método | Parámetros | Descripción |
|--------|-----------|-------------|
| `AbrirBaseDatos(filePath)` | `filePath`: ruta al .db | Abre BD SQLite con WAL + foreign_keys. Cierra la anterior si existe |
| `CerrarBaseDatos()` | — | Cierra la BD actual |
| `ObtenerFilas(moduloKey, orden)` | `moduloKey`: key de Modulos map; `orden`: columna DESC/ASC | SELECT * FROM cfg.Vista con sanitizacion whitelist. Soporta 9 modulos. |
| `ObtenerFilaPorId(moduloKey, id)` | `moduloKey`, `id`: int | Retorna `Row` unica o error |
| `ObtenerRutaProcesos()` | — | JOIN completo para ruta de procesos (especifico de expedientes) |
| `ObtenerDocumentosPendientes()` | — | WHERE estatus <> FIRMADO (especifico de expedientes) |
| `ObtenerHistorialFila(moduloKey, id)` | `moduloKey`, `id`: int | SELECT cfg.QueryHistorial (JOIN multi-tabla segun modulo) |
| `GuardarFila(moduloKey, data)` | `moduloKey`, `data`: map[string]interface{} | INSERT o UPDATE segun presencia de cfg.IDColumna |
| `EliminarFila(moduloKey, id)` | `moduloKey`, `id`: int64 | DELETE en transaccion (historial + modulo) |
| `ObtenerCatalogos()` | — | Retorna map[string][]CatalogoItem (11 catalogos) |
| `OptimizarBD()` | — | Ejecuta VACUUM |
| `GuardarNuevoCatalogo(tabla, nombre, extra)` | `extra`: map con col/val opcional | INSERT en tabla catalogo (whitelist tabla/columna) |
| `AbrirDialogoBD()` | — | Abre dialogo nativo Wails (`runtime.OpenFileDialog`) para seleccionar .db |
| `GuardarDialogoBD(nombreDefault)` | `nombreDefault`: string | Abre dialogo nativo Wails (`runtime.SaveFileDialog`) para guardar copia |
| `SetBackupMaxCopies(n)` | `n`: int | Configura numero de backups rotativos (1-20) |
| `GetBackupMaxCopies()` | — | Retorna numero actual de backups |
| `DescargarBD(destPath)` | `destPath`: string | Copia el .db actual a otra ruta |

## Data Layer — Frontend (HTMX + JS minimo)

La mayoria de las funciones JS previas (cargarCatalogos, obtenerExpedientes, guardarExpedienteEnBd, etc.) fueron reemplazadas por HTMX declarativo. JS actual minimo: helpers de modales (`mostrarFormulario`, `cerrarFormulario`), paginacion DOM del lado del cliente, localStorage (recientes/fijados), y `abrirBaseDatos()` (unica funcion que invoca el binding Wails `AbrirDialogoBD`).

## UI Layer — Tabla Principal (renderizada en Go, actualizada via HTMX)

| Funcion | Parametros | Descripcion |
|---------|-----------|-------------|
| `mostrarFormulario(id)` | `id`: opcional | Abre modal formulario (crear/editar) con titulo dinamico segun modulo |
| `cerrarFormulario()` | — | Cierra modal, limpia formulario |
| `toggleDesplegable(id)` | `id`: registro.id | Expande/colapsa fila desplegable |
| `renderPaginacion()` | — | Renderiza controles de paginacion |
| `irPagina(n)` | `n`: numero de pagina | Cambia pagina, refresca tabla |
| `aplicarPaginacion()` / `aplicarPaginacionDOM()` | — | Calcula paginas, renderiza slice actual |

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
