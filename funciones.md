# Catálogo de Funciones (SPOT)

Fuente única de verdad de la lógica existente en `index.html`, `schema-config.js`, `main.js` y `preload.js`. Antes de crear una nueva función, **revisar si ya existe** para evitar duplicación (DRY).

---

## Data Layer — SQLite/Catalogos

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `execSafe(sql, params?)` | `sql`: string SQL, `params[]`: valores opcionales para `?` | Ejecuta consulta SQL con escape manual. Retorna resultado de `db.exec()` o `null` si error |
| `toInt(v)` | `v`: valor a parsear | `parseInt(v,10)` con null-safe. Retorna `null` si no es número válido |
| `dbToObjects(res)` | `res`: resultado crudo de `db.exec()` | Convierte array de columnas+values a `[{col:val}]`. Retorna `[]` si vacío |
| `sanitizeNull(val)` | `val`: valor de la UI | Retorna `null` si es `null/undefined/''`, si no `val` tal cual |
| `cargarCatalogos()` | — | Carga todos los catálogos desde BD a `catalogosCache` según `SCHEMA_CONFIG.catalogoPorSelect`. Luego llama `poblarSelectores()` |
| `poblarSelectores()` | — | Llena todos los `<select>` con opciones desde `catalogosCache`. Al final llama `cargarSuperintendencias()` |
| `cargarSuperintendencias()` | — | Filtra superintendencias según gerencia seleccionada (FK dependiente) |
| `cargarDatos()` | — | Ejecuta `SELECT` sobre la vista (`SCHEMA_CONFIG.viewName`), llama `renderizarTabla()` |

---

## UI Layer — Tabla Principal

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `renderizarTabla(lista)` | `lista[]`: array de objetos expediente | Renderiza la tabla de 8 columnas + fila desplegable por expediente. Incluye botón de estrella (frecuentes) |
| `cambiarOrden()` | — | Lee el selector de orden (Reciente/Fecha creación/Fecha modificación) y recarga datos ordenados |
| `toggleDesplegable(id)` | `id`: `expediente.id_expediente` | Expande/colapsa fila desplegable con detalle completo, historial y notas |
| `toggleDetalle(prefix, id)` | `prefix`: string prefijo DOM, `id`: ID del expediente | Alterna visibilidad de detalle (reutilizable para distintos paneles) |
| `formatNum(v)` | `v`: número | Formatea con `toLocaleString('es-VE')` + 2 decimales |
| `parseNum(v)` | `v`: string de la UI | Convierte string numérico a `float` respetando separador `.` |
| `limitDecimals(v)` | `v`: número | Limita a 2 decimales sin redondeo excesivo |
| `validarFechas()` | — | Valida coherencia de fechas en el formulario (fecha_firma >= fecha_recibido, etc.). Retorna `true/false` |

---

## UI Layer — Formulario de Edición

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `mostrarFormulario(id?)` | `id`: opcional, si existe edita, si no crea nuevo | Abre modal de formulario, configura modo creación/edición |
| `cancelarFormulario()` | — | Cierra modal y limpia el formulario |
| `cargarExpediente(id)` | `id`: ID del expediente | Carga datos de un expediente existente en el formulario para edición |
| `calcularBs(origen?)` | `origen?`: `'usd'` o `'bs'` | Conversión bidireccional USD↔Bs. Si cambia tipo_cambio, recalcula ambos |
| `guardarExpediente()` | — | Valida campos, construye INSERT o UPDATE, ejecuta, actualiza BD y cierra modal |
| `eliminarExpediente()` | — | Pide confirmación, ejecuta DELETE, recarga tabla |
| `marcarCamposEdicionFrecuente()` | — | Marca con indicador amarillo los campos definidos en `SCHEMA_CONFIG.camposEdicionFrecuente` |
| `generarObservacionAutomatica()` | — | Genera línea de observación automática delegando en `SCHEMA_CONFIG.generarObservacion()` |
| `previewObservacion()` | — | Extrae texto libre, regenera observación combinando parte automática + texto libre del usuario |
| `aplicarTriggerFirma()` | — | Cambia lógica de firma según estatus (fecha_firma, etc.) |
| `formatTiempoEjecucion(v)` | `v`: número o string | Aplica sufijo "DÍAS" si es numérico |
| `parseTiempoEjecucionParaEdicion(v)` | `v`: string | Quita sufijo "DÍAS" para edición |

---

## UI Layer — Modales Secundarios

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `abrirRutaProcesos()` | — | Abre modal con tabla de ruteo: emisor, receptor, estatus, fechas |
| `cerrarRutaProcesos()` | — | Cierra modal de ruta de procesos |
| `abrirDocumentosPendientes()` | — | Abre modal con listado de expedientes donde estatus ≠ FIRMADO |
| `cerrarPendientes()` | — | Cierra modal de documentos pendientes |
| `abrirHistorialCompleto(id)` | `id`: ID del expediente | Abre modal con historial completo de snapshots |
| `cerrarHistorialCompleto()` | — | Cierra modal de historial |
| `cargarHistorialCompleto(id)` | `id`: ID del expediente | Consulta historial_movimientos y renderiza tabla de snapshots |
| `campoHistorial(label, valor, color)` | `label`: string, `valor`: string, `color`: clase CSS | Renderiza un campo individual en el detalle del historial |
| `getEstatusClass(estatus)` | `estatus`: string | Delega en `SCHEMA_CONFIG.estatusClass()` para obtener clase CSS del badge |

---

## UI Layer — Sidebar (Frecuentes)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `toggleFrecuente(id, solped)` | `id`: ID, `solped`: string | Marca/desmarca expediente como frecuente en localStorage. Actualiza estrella y sidebar |
| `renderSidebar()` | — | Renderiza la lista de expedientes frecuentes desde localStorage |
| `toggleSidebar()` | — | Colapsa/expande sidebar y persiste estado en localStorage |
| `toggleModoOrdenForm()` | — | Alterna entre orden por secciones y orden Excel en el formulario de edición |
| `aplicarOrdenExcel()` | — | Clona campos en grilla plana siguiendo `SCHEMA_CONFIG.ordenExcel` |
| `restaurarOrdenSecciones()` | — | Restaura el orden agrupado por secciones en el formulario |

---

## UI Layer — Catálogos (Gestión de opciones)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `inicializarBotonesCatalogo()` | — | Asigna eventos a botones "+" de cada selector de catálogo |
| `abrirAgregarCatalogo(selectId)` | `selectId`: ID del `<select>` | Abre modal para agregar nuevo registro al catálogo correspondiente |
| `cerrarAgregarCatalogo()` | — | Cierra modal de agregar catálogo |
| `captureAndRestoreFormState(callback)` | `callback`: función a ejecutar tras restaurar | Captura valores de todos los inputs/selects, ejecuta callback, restaura valores |
| `guardarNuevoCatalogo()` | — | Inserta nuevo registro en tabla catálogo y actualiza el selector |

---

## BD Layer — Archivos y Persistencia (Electron + Navegador)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `_cargarBaseDatosComun(bytes, fileName, filePath)` | `bytes`: Uint8Array, `fileName`: string, `filePath`: string | Inicializa BD desde buffer. Si Electron, sincroniza ruta vía IPC |
| `registrarReciente(nombre, path)` | `nombre`: string, `path`: string | Guarda BD en lista de recientes (localStorage) |
| `eliminarReciente(path)` | `path`: string | Elimina una BD de recientes por ruta |
| `eliminarRecienteIndex(index)` | `index`: número | Elimina una BD de recientes por índice |
| `escapeHtml(text)` | `text`: string | Escapa caracteres HTML para prevenir XSS en el DOM |
| `abrirBaseDatosReciente(path)` | `path`: string | Abre BD desde la lista de recientes |
| `mostrarMenuRecientes()` | — | Renderiza menú desplegable con BD recientes |
| `abrirBaseDatos()` | — async | Dispara `<input type="file">`, lee archivo, llama `_cargarBaseDatosComun()` |
| `guardarBD()` | — async | Exporta buffer sql.js y escribe a disco (Electron: IPC, navegador: download) |
| `marcarModificado()` | — | Marca la BD como modificada (habilita botón guardar si aplica) |
| `iniciarAutoguardado()` | — | Inicia intervalo de autoguardado cada 30s |
| `actualizarEstadoBD(msg)` | `msg`: string | Actualiza indicador visual de estado de BD en la UI |

---

## SCHEMA_CONFIG (schema-config.js)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `generarObservacion()` | (usa datos del formulario global) | Retorna string auto-generado: `[ESTATUS] Doc: [documento] [fechas]` |
| `extraerTextoLibre(currentValue, autoLine)` | `currentValue`: texto completo del textarea, `autoLine`: línea automática | Resta `autoLine` de `currentValue` para aislar el texto libre del usuario |
| `estatusClass(estatus)` | `estatus`: string | Retorna clase CSS según estatus (verde=firmado, amarillo=en_proceso, rojo=devuelto, gris=pendiente) |
| `esEstatusFirmado(estatus)` | `estatus`: string | Retorna `true` si el estatus es FIRMADO o equivalente |

---

## Electron IPC (main.js)

| Handler | Parámetros | Descripción |
|---------|-----------|-------------|
| `save-db` | `dataBase64`: string base64 del buffer | Escribe buffer a la ruta actual de BD |
| `save-db-as` | `dataBase64`: string base64 | Abre diálogo "Guardar como" y escribe |
| `set-db-path` | `filePath`: string | Guarda la ruta del archivo actual |
| `get-db-path` | — | Retorna la ruta actual de BD |
| `open-db-file` | `filePath`: string | Lee archivo y retorna buffer como base64 |
| `open-db-dialog` | — | Abre diálogo nativo para seleccionar archivo .db |

## Electron Preload (preload.js)

| Exposición | Parámetros | Descripción |
|-----------|-----------|-------------|
| `saveDb` | `(dataBase64)` → `ipcRenderer.invoke('save-db', ...)` | Guarda BD en ruta actual |
| `saveDbAs` | `(dataBase64)` → `ipcRenderer.invoke('save-db-as', ...)` | Guarda BD con diálogo |
| `setDbPath` | `(filePath)` → `ipcRenderer.invoke('set-db-path', ...)` | Sincroniza ruta de BD |
| `getDbPath` | `()` → `ipcRenderer.invoke('get-db-path')` | Obtiene ruta actual |
| `openDbDialog` | `()` → `ipcRenderer.invoke('open-db-dialog')` | Abre selector de archivo nativo |
| `openDbFilePath` | `(filePath)` → `ipcRenderer.invoke('open-db-file', ...)` | Lee archivo por ruta |
