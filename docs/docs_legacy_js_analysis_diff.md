--- docs/legacy/js_analysis.md (原始)


+++ docs/legacy/js_analysis.md (修改后)
# Análisis Exhaustivo de JavaScript Legacy

## A. Catálogo de Funciones JS

### Modales

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `pushModal(id)` | 102 | Agrega modal al stack, oculta overflow del body | `hx-on::after-request`, clicks en botones | `MODAL_STACK` (array global) | index.html:51, 143, 156, 167, 181, 192, 203, 214 |
| `cerrarModal(id)` | 118 | Remueve modal del stack, restaura overflow | Click en overlay, botón cerrar | `MODAL_STACK` | index.html:156, 167, 171, 181, 192, 203, 214, 225, 252 |
| `mostrarFormulario(id, modulo)` | 128 | Muestra modal de formulario con título dinámico | `hx-on::after-request` en btn-nuevo, `hxGetFormulario` | Lee `window.PAGE_DATA.modulos`, `window.PAGE_DATA.ActiveModule` | index.html:45, app.js:524 |
| `cerrarFormulario()` | 140 | Cierra modal de formulario | Click botón cancelar/guardar | — | components.html:3, 11, 18 |
| `cerrarHistorial()` | 179 | Cierra modal de historial | Click botón cerrar | — | index.html:167 |
| `cerrarRuta()` | 180 | Cierra modal de ruta procesos | Click botón cerrar | — | index.html:181 |
| `cerrarPendientes()` | 181 | Cierra modal de pendientes | Click botón cerrar | — | index.html:192 |
| `abrirSumas()` | 639 | Abre modal de sumas, inicializa filas | Click btn-sumas | — | index.html:61, 252 |
| `cerrarSumas()` | 642 | Cierra modal de sumas | Click botón cerrar | — | index.html:265 |
| `abrirModalExportar()` | 489 | Abre modal de exportar, carga columnas | Click btn-exportar | — | index.html:55, 225 |
| `cerrarModalExportar()` | 492 | Cierra modal de exportar | Click cancelar | — | index.html:225, 263 |

### Pines / Fijados (Acceso Rápido)

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `toggleFrecuente(id, solped, modulo)` | 263 | Agrega/remueve expediente de fijados | Click en icono pin en tabla | `localStorage.sidebarFrecuentes`, actualiza color del botón | tabla_expedientes.html:33, app.js:321 |
| `abrirFrecuentes()` | 293 | Muestra modal con lista de fijados | Click btn-frecuentes | Lee `localStorage.sidebarFrecuentes`, renderiza HTML dinámico | index.html:58, 214 |
| `cerrarFrecuentes()` | 328 | Cierra modal de fijados | Click botón cerrar | — | index.html:214, app.js:316 |
| `inicializarPines()` | 330 | Colorea pines según estado en localStorage | `DOMContentLoaded`, `htmx:afterSwap` | Lee `localStorage.sidebarFrecuentes`, modifica DOM | app.js:447, 453 |

### BD Recientes

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `registrarReciente(nombre, path)` | 184 | Guarda BD en recientes (max 5) | `DOMContentLoaded` si hay DB abierta | `localStorage.baseaccess_recientes` | app.js:449 |
| `eliminarReciente(path)` | 197 | Elimina BD reciente por path | Click en botón eliminar | `localStorage.baseaccess_recientes` | app.js:254 |
| `eliminarRecienteIndex(index)` | 205 | Elimina BD reciente por índice | Click en botón X en modal | `localStorage.baseaccess_recientes` | app.js:254 |
| `abrirBaseDatosReciente(path)` | 215 | Abre BD desde recientes | Click en item de recientes | — | app.js:249 |
| `abrirRecientes()` | 237 | Muestra modal de recientes | Click btn-recientes | Lee `localStorage.baseaccess_recientes`, renderiza HTML | index.html:36, 203 |
| `cerrarRecientes()` | 260 | Cierra modal de recientes | Click botón cerrar | — | index.html:203, app.js:234 |

### Paginación Cliente

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `irPagina(pagina)` | 353 | Cambia página actual y aplica filtrado DOM | Click en botón de paginación | `currentPage` (global) | app.js:401-424 (render) |
| `aplicarPaginacionDOM()` | 358 | Oculta/muestra filas según página | `irPagina`, `DOMContentLoaded`, `htmx:afterSwap` | `currentPage`, `CONFIG.pageSize` | app.js:447, 452 |
| `renderPaginacionControles(totalPages)` | 387 | Genera HTML de controles de paginación | `aplicarPaginacionDOM` | `CONFIG.maxVisiblePages` | app.js:384 |

### Exportar Excel

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `cargarColumnasExportar()` | 501 | Fetch columnas del módulo, render checkboxes y filtros | `abrirModalExportar`, `onchange` en select módulo | `_expSuperCache` (closure), lee `window.PAGE_DATA.catalogs` | index.html:230, 225 |
| `filtrarSuperintendenciasExportar()` | 567 | Filtra opciones de superintendencia según gerencia | `onchange` en gerencia, `cargarColumnasExportar` | `_expSuperCache` | index.html:236 |
| `toggleTodasColumnas(sel)` | 593 | Marca/desmarca todos los checkboxes de columnas | Click botones seleccionar/limpiar | — | index.html:244, 245 |
| `ejecutarExportar()` | 597 | Fetch GET a `/api/exportar-excel`, descarga blob | Click botón descargar | — | index.html:263 |

### Sumas (Calculadora)

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `anyadirFilaSuma()` | 646 | Agrega input numérico al modal de sumas | Click "Añadir número" | — | index.html:258 |
| `calcularSumas()` | 650 | Suma todos los valores y muestra resultado | `oninput` en inputs suma | — | index.html:258, 262 |
| `limpiarSumas()` | 655 | Resetea filas de suma a una sola | Click "Limpiar todo" | — | index.html:265 |

### Campos Numéricos

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `inicializarCamposNumericos()` | 716 | Aplica formato numérico a inputs con `inputmode="decimal"` o `.num-field` | `DOMContentLoaded`, `htmx:afterSettle`, `abrirSumas` | `dataset.raw`, `dataset.numInited` en cada input | form_*.html (todos), app.js:173, 641 |
| `_initNumInput(input)` | 688 | Adjunta listeners de focus/blur/input a un input numérico | `inicializarCamposNumericos` | `dataset.raw` para valor sin formato | — |
| `_fmtNum(input)` | 678 | Formatea valor con separador de miles y 2 decimales | `blur` event | — | — |
| `_parseValue(input)` | 667 | Parsea string a número, maneja coma como punto | Todas las funciones numéricas | — | — |
| `_rawNum(input)` | 662 | Retorna valor crudo sin formato | `convertirMoneda` | — | — |

### Superintendencias

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `cargarSuperintendencias()` | 144 | Filtra opciones de superintendencia según gerencia seleccionada | `mostrarFormulario`, `htmx:afterSettle`, `onchange` en gerencia | `_superCache` (global) | components.html:65, form_*.html, app.js:137, 171-176 |

### Conversión USD/Bs

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `convertirMoneda(origen)` | 728 | Convierte entre USD y Bs usando tipo de cambio | `oninput` en campos de presupuesto/adjudicación | `_convLock` (lock para evitar recursión) | form_expedientes.html:22-24, 48-49 |

### Helpers / Utilidades

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `$(id)` | 1 | Shortcut para `document.getElementById` | Todas las funciones | — | Todo el código |
| `toast(msg, tipo)` | 18 | Muestra notificación toast temporal | Reemplaza `alert`, llamado manual | — | app.js:29, 53, 619, 632 |
| `esc(v)` | 13 | Escape HTML para prevenir XSS | Renderizado dinámico en JS | — | app.js:251, 319, 528 |
| `toggleDesplegable(id)` | 92 | Muestra/oculta subfila de detalles | Click en fila de tabla | — | tabla_expedientes.html:24 |
| `toggleSortDir()` | 457 | Alterna dirección de ordenamiento (ASC/DESC) | Click botón ordenar | Lee/escribe `#sort-dir-val` | index.html:83 |
| `hxGetFormulario(id, modulo)` | 468 | Carga formulario via HTMX y abre modal | Click en fijados | — | app.js:316 |

### Gantt / Ruta Procesos (IIFE en ruta_procesos.html)

| Nombre | Línea | Qué hace | Cómo se activa | Estado que maneja | Template(s) |
|--------|-------|----------|----------------|-------------------|-------------|
| `toggleModal(id)` | 175 | Muestra/oculta modal genérico | Clicks en botones de hoja/leyenda | — | ruta_procesos.html:15, 75, 100, 122 |
| `cambiarHoja()` | 179 | Cambia hoja del Gantt via HTMX | `onchange` en select hoja | — | ruta_procesos.html:6 |
| `eliminarHojaActual()` | 185 | Elimina hoja actual y sus procesos | Click botón eliminar hoja | — | ruta_procesos.html:18 |
| `crearHoja()` | 202 | Crea nueva hoja | Click botón crear en modal | — | ruta_procesos.html:90 |
| `editarLeyenda(id, nombre, color)` | 224 | Abre modal de edición de leyenda | Click en item de leyenda | — | ruta_procesos.html:528 |
| `guardarEditarLeyenda()` | 231 | Actualiza leyenda via fetch POST | Click botón actualizar | — | ruta_procesos.html:136 |
| `crearLeyenda()` | 254 | Crea nueva leyenda | Click botón guardar en modal | — | ruta_procesos.html:113 |
| `cargarRegistrosModulo()` | 281 | Fetch registros disponibles para proceso | `toggleFormProceso`, `onchange` en módulo | — | ruta_procesos.html:37 |
| `toggleFormProceso()` | 308 | Muestra/oculta form de agregar proceso | Click "Añadir Proceso" | `formVisible` (closure) | ruta_procesos.html:30 |
| `agregarProceso()` | 316 | Agrega proceso a la ruta | Click botón Agregar | — | ruta_procesos.html:57 |
| `toggleProceso(id, checked)` | 344 | Activa/desactiva proceso | Click checkbox | — | ruta_procesos.html:52 |
| `eliminarProceso(id)` | 357 | Elimina proceso de la ruta | Click botón eliminar | — | ruta_procesos.html:52 |
| `renderAll()` | 371 | Renderiza grid de procesos | `agregarProceso`, `toggleProceso`, `eliminarProceso`, init | `processes` (closure) | ruta_procesos.html:60 |
| `renderGantt()` | 385 | Renderiza cronograma Gantt y leyenda | `renderAll` | `ganttColumns`, `legend`, `processes` (closure) | ruta_procesos.html:62 |
| `abrirEditarCronograma(procId, fecha, statusName, note)` | 508 | Abre modal para editar día del cronograma | Click en celda del Gantt | — | ruta_procesos.html:528 |
| `cerrarCronoModal()` | 528 | Cierra modal de cronograma | Click cancelar | — | ruta_procesos.html:145 |
| `guardarCronoDia()` | 533 | Guarda día del cronograma via fetch POST | Click botón guardar | Actualiza `processes.timeline` local | ruta_procesos.html:164 |

---

## B. Estado Global y localStorage

### Variables Globales

| Variable | Tipo | Propósito | Línea |
|----------|------|-----------|-------|
| `$` | Function | Shortcut `document.getElementById` | 1 |
| `STORAGE_KEYS` | Object | Constantes para claves de localStorage | 4 |
| `CONFIG` | Object | Configuración (pageSize, maxVisiblePages, timeouts) | 6-10 |
| `MODAL_STACK` | Array | Stack de IDs de modales abiertos | 101 |
| `_superCache` | Array/null | Cache de opciones de superintendencia | 143 |
| `currentPage` | Number | Página actual de paginación cliente | 351 |
| `_expSuperCache` | Array/null | Cache de superintendencias para exportar | 565 |
| `_convLock` | Boolean | Lock para evitar recursión en conversión USD/Bs | 727 |
| `window.PAGE_DATA` | Object | Datos inyectados desde Go (hasDB, dbPath, catalogs, modulos, filas, etc.) | index.html:16-27 |
| `window.alert` | Function | Sobrescrita para redirigir a `toast()` | 28-30 |

### Closure Variables (IIFE Gantt en ruta_procesos.html)

| Variable | Tipo | Propósito | Línea |
|----------|------|-----------|-------|
| `data` | Object | Datos JSON inyectados desde Go | ruta_procesos.html:171 |
| `legend` | Array | Leyendas del Gantt (estatus + colores) | ruta_procesos.html:172 |
| `columns` (ganttColumns) | Array | Columnas de fechas del Gantt | ruta_procesos.html:172 |
| `processes` | Array | Procesos activos en la ruta | ruta_procesos.html:172 |
| `offset_weeks` | Number | Offset de semanas para navegación | ruta_procesos.html:172 |
| `formVisible` | Boolean | Estado de visibilidad del form de agregar proceso | ruta_procesos.html:173 |

### Claves de localStorage

| Clave | Estructura | Propósito | Funciones que usan |
|-------|------------|-----------|-------------------|
| `sidebarFrecuentes` | `[{id: number, solped: string, modulo: string}]` | Expedientes fijados en sidebar | `toggleFrecuente`, `abrirFrecuentes`, `inicializarPines` |
| `baseaccess_recientes` | `[{nombre: string, path: string, timestamp: number}]` | Bases de datos recientes (max 5) | `registrarReciente`, `abrirRecientes`, `eliminarReciente`, `eliminarRecienteIndex` |

---

## C. Mapa de Interacción HTML → JS

### index.html

```
index.html
├── onclick="abrirBaseDatos()"          → app.js:33
├── onclick="abrirRecientes()"           → app.js:237
├── hx-on::after-request="mostrarFormulario(null)" → app.js:128
├── hx-on::after-request="pushModal('modal-pendientes')" → app.js:102
├── onclick="abrirModalExportar()"       → app.js:489
├── onclick="abrirFrecuentes()"          → app.js:293
├── onclick="abrirSumas()"               → app.js:639
├── onclick="toggleSortDir()"            → app.js:457
├── hx-on::after-request="pushModal('modal-ruta')" → app.js:102
├── onclick="cerrarFormulario()"         → app.js:140
├── onclick="cerrarHistorial()"          → app.js:179
├── onclick="cerrarRuta()"               → app.js:180
├── onclick="cerrarPendientes()"         → app.js:181
├── onclick="cerrarRecientes()"          → app.js:260
├── onclick="cerrarFrecuentes()"         → app.js:328
├── onclick="cerrarModalExportar()"      → app.js:492
├── onchange="cargarColumnasExportar()"  → app.js:501
├── onclick="toggleTodasColumnas(true)"  → app.js:593
├── onclick="toggleTodasColumnas(false)" → app.js:593
├── onclick="ejecutarExportar()"         → app.js:597
├── onclick="cerrarSumas()"              → app.js:642
├── onclick="anyadirFilaSuma()"          → app.js:646
├── onclick="limpiarSumas()"             → app.js:655
└── [DOMContentLoaded]                   → app.js:57, 447-453
    ├── registrarReciente()              → app.js:184
    ├── aplicarPaginacionDOM()           → app.js:358
    └── inicializarPines()               → app.js:330
```

### components.html

```
components.html
├── onclick="cerrarFormulario()"         → app.js:140
├── hx-on::after-request="... cerrarFormulario() ..." → app.js:140
├── onclick="cargarSuperintendencias()"  → app.js:144 (via onchange en gerencia)
├── onclick="toggleDesplegable(...)"     → app.js:92
├── hx-on::after-request="mostrarFormulario(...)" → app.js:128
└── hx-on::before-request="pushModal('historial-modal')..." → app.js:102
```

### form_expedientes.html (y otros form_*.html)

```
form_*.html
├── oninput="convertirMoneda()"          → app.js:728 (campos presupuesto USD, tipo_cambio)
├── oninput="convertirMoneda('bs_presup')" → app.js:728 (campo presupuesto Bs)
├── oninput="convertirMoneda('usd_adj')" → app.js:728 (campo adjudicado USD)
├── oninput="convertirMoneda('bs_adj')"  → app.js:728 (campo adjudicado Bs)
└── [htmx:afterSettle]                   → app.js:171-177
    ├── _superCache = null
    ├── inicializarCamposNumericos()     → app.js:716
    └── cargarSuperintendencias()        → app.js:144 (si form-modal visible)
```

### tabla_expedientes.html (y otras tabla_*.html)

```
tabla_*.html
├── onclick="toggleDesplegable(...)"     → app.js:92
├── hx-on::after-request="mostrarFormulario(...)" → app.js:128
├── onclick="toggleFrecuente(...)"       → app.js:263
├── hx-on::before-request="pushModal('historial-modal')..." → app.js:102
└── [htmx:afterSwap en #vista-tabla]     → app.js:441-445
    ├── currentPage = 1
    ├── aplicarPaginacionDOM()           → app.js:358
    └── inicializarPines()               → app.js:330
```

### ruta_procesos.html (IIFE propia)

```
ruta_procesos.html (dentro de IIFE)
├── onchange="cambiarHoja()"             → línea 179
├── onclick="toggleModal('crear-hoja-modal')" → línea 175
├── onclick="eliminarHojaActual()"       → línea 185
├── onclick="toggleFormProceso()"        → línea 308
├── onchange="cargarRegistrosModulo()"   → línea 281
├── onclick="agregarProceso()"           → línea 316
├── onclick="toggleProceso(...)"         → línea 344
├── onclick="eliminarProceso(...)"       → línea 357
├── onclick="editarLeyenda(...)"         → línea 224
├── onclick="toggleModal('crear-leyenda-modal')" → línea 175
├── onclick="guardarEditarLeyenda()"     → línea 231
├── onclick="toggleModal('editar-leyenda-modal')" → línea 175
├── onclick="cerrarCronoModal()"         → línea 528
├── onclick="abrirEditarCronograma(...)" → línea 508
└── onclick="guardarCronoDia()"          → línea 533
```

### Eventos Globales (document/body)

```
Global
├── dragover / dragleave / drop          → app.js:79-89 (drag & drop de BD)
├── click en .modal (overlay)            → app.js:110-117 (cerrar modal al hacer click afuera)
├── htmx:afterSettle                     → app.js:171-177 (reset superCache, inicializar campos)
├── htmx:afterSwap (#vista-tabla)        → app.js:441-445 (reset paginación, pines)
└── DOMContentLoaded                     → app.js:57-76, 441-453
    ├── listener en #dbfile              → app.js:58-75
    ├── registrarReciente()              → app.js:449
    ├── aplicarPaginacionDOM()           → app.js:452
    └── inicializarPines()               → app.js:453
```

---

## D. Dependencias entre Funciones

```
abrirBaseDatos()
└── fetch('/api/abrir-bd') → location.reload()

mostrarFormulario(id, modulo)
├── pushModal('form-modal')
└── cargarSuperintendencias()

cargarSuperintendencias()
└── usa _superCache (cache global)

toggleFrecuente(id, solped, modulo)
├── localStorage.setItem()
├── abrirFrecuentes() (si modal abierto)
└── actualiza DOM de botón pin

abrirFrecuentes()
├── pushModal('modal-frecuentes')
└── renderiza HTML desde localStorage

registrarReciente(nombre, path)
└── localStorage.setItem()

abrirRecientes()
├── pushModal('modal-recientes')
└── renderiza HTML desde localStorage

irPagina(pagina)
└── aplicarPaginacionDOM()
    └── renderPaginacionControles()

cargarColumnasExportar()
├── fetch('/api/columnas-modulo')
└── filtrarSuperintendenciasExportar()

anyadirFilaSuma()
└── _initNumInput() → calcularSumas()

calcularSumas()
└── _parseValue() para cada input

convertirMoneda(origen)
├── _parseValue() para leer tipo_cambio
└── setVal() con _fmtNum()

inicializarCamposNumericos()
└── _initNumInput() para cada input

cargarSuperintendencias()
└── usa _superCache

[IIFE Gantt]
renderAll()
└── renderGantt()
```

---

## E. Patrones de Diseño Identificados

### 1. Modal Stack Pattern
```js
MODAL_STACK.push(id); // al abrir
MODAL_STACK.splice(idx, 1); // al cerrar
```
Permite múltiples modales apilados con cierre jerárquico.

### 2. localStorage como Fuente de Verdad UI
```js
let list = JSON.parse(localStorage.getItem(KEY) || '[]');
// modificar list
localStorage.setItem(KEY, JSON.stringify(list));
```
Usado en fijados y recientes.

### 3. Cache de Selectores (Superintendencias)
```js
if (_superCache === null) {
    _superCache = [];
    sel.querySelectorAll('option').forEach(...);
}
```
Evita reconsultar el DOM en cada filtro.

### 4. Formato Numérico con dataset.raw
```js
input.dataset.raw = n.toFixed(2); // valor crudo
input.value = n.toLocaleString('es-ES', ...); // valor formateado
```
Separa valor interno de representación visual.

### 5. Lock para Evitar Recursión
```js
if (_convLock) return;
_convLock = true;
try { ... } finally { _convLock = false; }
```
Usado en conversión USD/Bs para evitar bucle infinito.

### 6. IIFE para Encapsular Estado Complejo
```js
(function() {
    const data = {{jsonEncode .}};
    var formVisible = false;
    // funciones privadas y públicas (window.xxx)
})();
```
Usado en Gantt para aislar estado del cronograma.

---

## F. Problemas Potenciales Identificados

1. **Variables globales mutables**: `_superCache`, `_expSuperCache`, `currentPage` pueden causar race conditions si hay múltiples operaciones simultáneas.

2. **Dependencia de IDs hardcodeados**: Funciones como `$('f-id_gerencia')`, `$('sumas-resultado')` asumen estructura fija del DOM.

3. **Mezcla de estilos de evento**: Algunos usan `onclick`, otros `addEventListener`, otros `hx-on::`. Inconsistente.

4. **No hay cleanup de listeners**: Los listeners agregados en `_initNumInput` nunca se remueven (pérdida de memoria en SPA de larga duración).

5. **Error handling inconsistente**: Algunas funciones usan `try/catch`, otras no. `alert` sobrescrito pero aún usado en algunos lugares.

6. **Magic numbers**: `CONFIG.pageSize = 10`, `maxVisiblePages = 7`, `timeouts` hardcoded.

7. **Acoplamiento fuerte con estructura HTML**: `renderPaginacionControles` genera HTML inline con clases Tailwind específicas.