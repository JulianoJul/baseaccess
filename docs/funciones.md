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
| `validarArchivoBD(file)` | `file`: File object del input/drop | Valida extensión (.db/.sqlite) y tamaño (&lt;= CONFIG.MAX_FILE_SIZE_BYTES). Retorna `true/false` |
| `obtenerRutaProcesos()` | — | Consulta `SCHEMA_CONFIG.viewName` con JOIN completo para ruta de procesos. Retorna `[{...}]` |
| `obtenerDocumentosPendientes()` | — | Consulta expedientes donde estatus ≠ FIRMADO. Retorna `[{...}]` |
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
| `optimizarBD()` | — async | Ejecuta `VACUUM` sobre la BD abierta. Reporta tamaño antes/después |
| `descargarBDError()` | — | Exporta BD actual como archivo `.db` descargable (uso desde error boundary modal) |
| `updateUIOnError()` | — | Deshabilita botones de modificación (nuevo, guardar, compactar) al ocurrir un error crítico |

## Electron (main.js) — Backup Rotativo

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `crearBackupRotativo(filePath)` | `filePath`: ruta del archivo .db actual | Rota hasta 5 backups (`.bak.1`..`.bak.5`), elimina el más antiguo, copia el actual como `.bak.1`. Llamado antes de cada `save-db` |

---

## Constantes Globales (schema-config.js)

| Constante | Descripción |
|-----------|-------------|
| `CONFIG` | Config numérica: `MAX_FILE_SIZE_BYTES`, `MAX_FILE_SIZE_MB`, `AUTOSAVE_INTERVAL_MS` |
| `DEBUG` | Wrapper condicional de console: `DEBUG.log()`, `DEBUG.error()` (controlado por `DEBUG.isEnabled`) |
| `MSG` | Mensajes de usuario centralizados: `ERROR_NO_DB`, `ERROR_TIPO_ARCHIVO`, `ERROR_TAMANO(sizeMB)`, `ERROR_LECTURA(err)`, `ERROR_CONSULTA(err)`, `ERROR_GUARDAR(err)`, `ERROR_ELIMINAR(err)`, `ERROR_NO_EXPEDIENTE`, `ERROR_ID_INVALIDO`, `ERROR_NO_BD_VALIDA`, `ERROR_NO_REABRIR(err)`, `ERROR_ABRIR_BD(err)`, `NOMBRE_OBLIGATORIO`, `EXITO_ACTUALIZADO`, `EXITO_CREADO`, `EXITO_ELIMINADO`, `FECHA_DEVUELTO_INVALIDA` |
| `STORAGE_KEYS` | Keys de localStorage: `FRECUENTES`, `RECIENTES`, `SIDEBAR_VISIBLE` |
| `SELECTORS` | IDs de elementos DOM: `TABLA_CUERPO`, `FORM_MODAL`, `SEARCH`, `SORT_ORDER`, `SIDEBAR`, `BODY`, `FILE_INPUT`, `MENU_RECIENTES`, `MODAL_RUTA`, `RUTA_CONTENIDO`, `MODAL_PENDIENTES`, `PENDIENTES_CONTENIDO`, `MODAL_HISTORIAL`, `HISTORIAL_CONTENIDO`, `MODAL_CATALOGO`, `AC_NOMBRE`, `F_OBSERVACIONES`, `GUARDAR_BD_BTN`, `BTN_VACUUM`, `MODAL_ERROR`, `ERROR_CONTENIDO`, `BTN_DESCARGAR_BD` |
| `MSG_EXTRA` | Mensajes de mantenimiento: `VACUUM_INICIADO`, `VACUUM_COMPLETADO(antes, despues)`, `VACUUM_ERROR(err)`, `ERROR_CRITICO`, `PROMESA_RECHAZADA`, `BD_DESCARGADA` |
| `BACKUP` | Config de backup rotativo: `MAX_COPIES: 5`, `SUFFIX: '.bak.'` |

## Helper (index.html)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `$(id)` | `id`: string ID del elemento | Atajo para `document.getElementById(id)` |

## SCHEMA_CONFIG (schema-config.js)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `generarObservacion(estatus, documento, fechaRecibido, fechaDevuelto)` | `estatus`: string, `documento`: string, `fechaRecibido`: string, `fechaDevuelto`: string | Retorna string auto-generado: `[ESTATUS] - [Documento] - ...` |
| `extraerTextoLibre(currentValue, autoLine)` | `currentValue`: texto completo del textarea, `autoLine`: línea automática | Resta `autoLine` de `currentValue` para aislar el texto libre del usuario |
| `estatusClass(estatus)` | `estatus`: string | Retorna clase CSS según estatus (verde=firmado, amarillo=en_proceso, rojo=devuelto, gris=pendiente) |
| `esEstatusFirmado(estatus)` | `estatus`: string | Retorna `true` si el estatus es FIRMADO o equivalente |

## SCHEMA_CONFIG (schema-config.js) — Nuevos campos

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `VERSION` | number | `8` — version del schema SQLite, validado contra `PRAGMA user_version` al cargar BD |
| `queries` | object | Queries SQL centralizadas: `rutaProcesos`, `documentosPendientes`, `reporteExcel`, `expedientesSelect`, `expedientePorId` |

---

## UI Layer — Nuevas funciones

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `renderBadgeEstatus(estatus)` | `estatus`: string | Retorna HTML string de badge `<span>` con clase de color según estatus. SPOT único para badges |
| `obtenerMaxBackups()` | — | Lee `localStorage[STORAGE_KEYS.BACKUP_MAX_COPIES]` con fallback a `BACKUP.MAX_COPIES`. Retorna número entre 1-20 |

## Constantes Globales (schema-config.js) — Actualizado

| Constante | Descripción |
|-----------|-------------|
| `CONFIG` | `MAX_FILE_SIZE_BYTES`, `MAX_FILE_SIZE_MB`, `BYTES_PER_MB`, `AUTOSAVE_INTERVAL_MS`, `AUTOSAVE_ENABLED` |
| `STORAGE_KEYS` | `FRECUENTES`, `RECIENTES`, `SIDEBAR_VISIBLE`, `BACKUP_MAX_COPIES` |
| `SELECTORS` | `TABLA_CUERPO`, `FORM_MODAL`, `SEARCH`, `SORT_ORDER`, `SIDEBAR`, `BODY`, `FILE_INPUT`, `MENU_RECIENTES`, `MODAL_RUTA`, `RUTA_CONTENIDO`, `MODAL_PENDIENTES`, `PENDIENTES_CONTENIDO`, `MODAL_HISTORIAL`, `HISTORIAL_CONTENIDO`, `MODAL_CATALOGO`, `AC_NOMBRE`, `F_OBSERVACIONES`, `GUARDAR_BD_BTN`, `BTN_VACUUM`, `MODAL_ERROR`, `ERROR_CONTENIDO`, `BTN_DESCARGAR_BD`, `ESTADO_BD` |

## Electron IPC (main.js)

| Handler | Parámetros | Descripción |
|---------|-----------|-------------|
| `save-db` | `dataBase64`: string base64 del buffer | Escribe buffer a la ruta actual de BD |
| `save-db-as` | `dataBase64`: string base64 | Abre diálogo "Guardar como" y escribe |
| `set-db-path` | `filePath`: string | Guarda la ruta del archivo actual |
| `get-db-path` | — | Retorna la ruta actual de BD |
| `open-db-file` | `filePath`: string | Lee archivo y retorna buffer como base64 |
| `open-db-dialog` | — | Abre diálogo nativo para seleccionar archivo .db |
| `set-backup-copies` | `n`: número | Configura cantidad de backups rotativos (1-20) |
| `get-backup-copies` | — | Retorna la cantidad actual de backups configurada |

## Electron Interna (main.js)

| Función | Parámetros | Descripción |
|---------|-----------|-------------|
| `crearBackupRotativo(filePath)` | `filePath`: string ruta de BD | Rota hasta `backupMaxCopies` backups (`.bak.1`..`.bak.N`), elimina más antiguo, copia el actual como `.bak.1`. Se llama antes de cada `save-db` |
| `setBackupMaxCopies(n)` | `n`: número | Actualiza el límite de backups (1-20), llamado vía IPC `set-backup-copies` |

## Electron Preload (preload.js)

| Exposición | Parámetros | Descripción |
|-----------|-----------|-------------|
| `saveDb` | `(dataBase64)` → `ipcRenderer.invoke('save-db', ...)` | Guarda BD en ruta actual |
| `saveDbAs` | `(dataBase64)` → `ipcRenderer.invoke('save-db-as', ...)` | Guarda BD con diálogo |
| `setDbPath` | `(filePath)` → `ipcRenderer.invoke('set-db-path', ...)` | Sincroniza ruta de BD |
| `getDbPath` | `()` → `ipcRenderer.invoke('get-db-path')` | Obtiene ruta actual |
| `openDbDialog` | `()` → `ipcRenderer.invoke('open-db-dialog')` | Abre selector de archivo nativo |
| `openDbFilePath` | `(filePath)` → `ipcRenderer.invoke('open-db-file', ...)` | Lee archivo por ruta |
| `setBackupCopies` | `(n)` → `ipcRenderer.invoke('set-backup-copies', n)` | Configura cantidad de backups |
| `getBackupCopies` | `()` → `ipcRenderer.invoke('get-backup-copies')` | Obtiene cantidad de backups |
