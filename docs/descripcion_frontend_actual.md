# Plan: Frontend HTML, CSS, JS — Documentación Completa

## Índice

1. [Arquitectura General](#1-arquitectura-general)
2. [Flujo de Datos](#2-flujo-de-datos)
3. [HTML Templates](#3-html-templates)
4. [CSS — styles.css](#4-css--stylescss)
5. [JavaScript — Alpine.js](#5-javascript--alpinejs)
6. [Glosario de Clases CSS](#6-glosario-de-clases-css)
7. [Glosario de Componentes Alpine](#7-glosario-de-componentes-alpine)
8. [Pseudocódigo General](#8-pseudocódigo-general)

---

## 1. Arquitectura General

```
┌─────────────────────────────────────────────────────────────────┐
│  Go backend (handler.go)                                        │
│  ├── preparePageData(r) → PageData (JSON para el frontend)      │
│  ├── ServeHTTP() → rutas /api/* + página principal              │
│  └── Go html/template renderiza HTML con PageData               │
├─────────────────────────────────────────────────────────────────┤
│  HTMX (comunicación servidor)                                   │
│  ├── hx-get / hx-post → peticiones HTTP a /api/*                │
│  ├── hx-target → inyecta fragmentos HTML en el DOM              │
│  └── hx-swap → reemplaza contenido del objetivo                 │
├─────────────────────────────────────────────────────────────────┤
│  Alpine.js (estado UI reactiva)                                 │
│  ├── Stores globales: modals, toast, fijados                    │
│  ├── Data components: appShell, formularioModulo, etc.          │
│  └── Directiva x-currency: formateo numérico automático         │
├─────────────────────────────────────────────────────────────────┤
│  CSS                                                            │
│  ├── tailwind.min.css → utilidades Tailwind (purged)            │
│  ├── styles.css → clases custom (838 líneas)                    │
│  └── fontawesome.min.css → iconos                               │
└─────────────────────────────────────────────────────────────────┘
```

### Stack Frontend

| Capa | Tecnología | Propósito |
|------|-----------|-----------|
| Estructura | Go `html/template` | Renderizado server-side |
| Reactividad | Alpine.js v3.14.8 | Estado UI local, modales, validación |
| Comunicación | HTMX | Partial HTML swaps, formularios |
| Estilos | Tailwind CSS (purged) + styles.css | Utility-first + custom components |
| Iconos | Font Awesome Free 7.3 | Iconos `fas fa-*` |
| Puente | alpine-htmx-bridge.js | Alpine.initTree() tras swap HTMX |

---

## 2. Flujo de Datos

### 2.1 Carga Inicial de Página

```
1. Browser → GET /
2. handler.go → preparePageData(r) → PageData {
     HasDB, Catalogs, Modulos, Filas, TotalPages, CurrentPage,
     PageSize, ActiveModule, CatalogFilters, DBPath, Error
   }
3. Go template renderiza index.html con PageData
4. <script> window.PAGE_DATA = {JSON de PageData} </script>
5. Alpine.js inicializa: appShell, modals, toast, fijados, bdRecientes
6. tabla.html se incluye inline para el módulo activo
```

### 2.2 Interacción HTMX

```
1. Usuario hace click → hx-get="/api/..."
2. HTMX envía request al servidor
3. handler.go renderiza fragmento HTML parcial
4. HTMX inyecta el fragmento en hx-target
5. alpine-htmx-bridge dispara Alpine.initTree() en el nuevo contenido
6. Alpine inicializa x-data, x-bind, etc. en los nuevos elementos
```

### 2.3 PageData Struct

```go
type PageData struct {
    Title          string
    HasDB          bool
    Catalogs       map[string][]CatalogoItem
    ActiveModule   string
    Modulos        map[string]ModuloConfig
    Filas          []Row
    Error          string
    PageSize       int
    TotalPages     int
    CurrentPage    int
    SortColumn     string
    SortDir        string
    DBPath         string
    Registro       Row
    CatalogFilters map[string]CatalogFilter
}
```

### 2.4 ModuloConfig

```go
type ModuloConfig struct {
    Nombre         string   // Nombre para mostrar
    Tabla          string   // Tabla SQLite
    Vista          string   // Vista SQLite (para consultas)
    IDColumna      string   // Columna PK
    HistorialTabla string   // Tabla de historial
    Columnas       []string // Columnas editables
    FechaColumna   string   // Columna de fecha para filtros
    GerenciasIDs   []int    // IDs de gerencias permitidas
    OrdenExcel     []string // Orden de columnas para exportar
}
```

9 módulos: `expedientes`, `requisiciones`, `memorandums`, `recobros`, `valuaciones`, `aprobacion_jd`, `certificacion_bdu`, `vacaciones`, `reposos_medicos`.

### 2.5 Row Type

```go
type Row map[string]interface{}  // Genérico: columna → valor
```

---

## 3. HTML Templates

### 3.1 `templates/new/index.html` (504 líneas) — Shell Principal

#### Estructura

```
<!DOCTYPE html>
├── <head>
│   ├── Meta: charset, viewport
│   ├── Title: {{.Title}}
│   ├── CSS: tailwind.min.css, fontawesome.min.css, styles.css
│   └── JS: htmx.min.js, alpine-app.js, alpine-directives.js,
│           alpine.min.js, alpine-htmx-bridge.js
├── <body x-data="appShell">
│   ├── window.PAGE_DATA {JSON server-rendered}
│   ├── <header> — Barra superior
│   │   ├── Izquierda: "Abrir Base de Datos" (Wails/HTML), "Recientes"
│   │   └── Derecha: "Nuevo Registro", "Pendientes", "Exportar",
│   │                "Fijados", "Sumas"
│   ├── Barra de búsqueda (sticky)
│   │   ├── Select de orden + botón dirección
│   │   └── Input búsqueda → hx-get="/api/filtrar-expedientes"
│   ├── #vista-tabla — Área principal (HTMX target)
│   │   ├── Sin DB: placeholder
│   │   ├── Con error: .Error
│   │   └── {{template "tabla.html" .}}
│   ├── #bottom-bar — Switcher de módulos (fixed bottom)
│   │   ├── Botones por módulo (range .Modulos)
│   │   └── Botón "Ruta Procesos"
│   ├── MODALES (8):
│   │   ├── #form-modal — Formulario de registro
│   │   ├── #historial-modal — Historial de cambios
│   │   ├── #modal-ruta — Gantt de ruta procesos
│   │   ├── #modal-pendientes — Documentos pendientes
│   │   ├── #modal-recientes — BD recientes
│   │   ├── #modal-frecuentes — Items fijados
│   │   ├── #export-modal — Exportar Excel
│   │   └── #sumas-modal — Calculadora de sumas
│   ├── #spinner-overlay — Indicador de carga
│   ├── #toast-container — Notificaciones
│   └── <script> — abrirBaseDatos() + file input handler
```

#### Alpine.js Usado

| Atributo | Componente | Ubicación |
|----------|-----------|-----------|
| `x-data="appShell"` | `Alpine.data('appShell')` | `<body>` |
| `x-data="bdRecientes"` | `Alpine.data('bdRecientes')` | Header + #modal-recientes |
| `x-data="fijados"` | `Alpine.data('fijados')` | Header + #modal-frecuentes |
| `x-data="exportarExcel"` | `Alpine.data('exportarExcel')` | #export-modal |
| `x-data="calculadoraSumas"` | `Alpine.data('calculadoraSumas')` | #sumas-modal |
| `$store.modals` | `Alpine.store('modals')` | Todas las modales |

#### HTMX Usado

| Elemento | Atributo | Valor | Propósito |
|----------|----------|-------|-----------|
| "Nuevo Registro" | `hx-get` | `/api/cargar-expediente` | Cargar formulario vacío |
| | `hx-target` | `#form-expediente` | Inyectar HTML |
| | `hx-include` | `#active-module-val` | Pasar módulo actual |
| "Pendientes" | `hx-get` | `/api/pendientes` | Cargar pendientes |
| | `hx-target` | `#pendientes-contenido` | Inyectar en modal |
| Select orden | `hx-get` | `/api/filtrar-expedientes` | Re-filtrar |
| | `hx-trigger` | `change` | Al cambiar |
| Input búsqueda | `hx-get` | `/api/filtrar-expedientes` | Búsqueda live |
| | `hx-trigger` | `input changed delay:200ms` | Debounce 200ms |
| Botones módulo | `hx-get` | `/api/cambiar-modulo?modulo=X` | Cambiar módulo |
| | `hx-target` | `#vista-tabla` | Reemplazar tabla |
| "Ruta Procesos" | `hx-get` | `/api/ruta-procesos` | Cargar Gantt |
| | `hx-target` | `#ruta-contenido` | Inyectar en modal |

#### Pseudocódigo: Comportamiento del Body

```
ON body load (x-data="appShell"):
  dragOver = false

  ON dragover.window.prevent → dragOver = true
  ON dragleave.window → dragOver = false
  ON drop.window.prevent → onDrop(e):
    archivo = e.dataTransfer.files[0]
    POST /api/abrir-bd {path: archivo.path}
    IF success → location.reload()
    ELSE → toast.error(msg)

  ON keydown.escape.window:
    cerrar última modal del stack

  PAGE_DATA = {JSON server-rendered con todos los datos}

  abrirBaseDatos():
    TRY:
      IF window.go.main.App → path = AbrirDialogoBD()  // Wails nativo
      ELSE IF window.runtime → runtime.OpenFileDialog()  // Wails runtime
      IF path → POST /api/abrir-bd → location.reload()
    CATCH:
      Fallback: #dbfile.click() → input[type=file]
```

---

### 3.2 `templates/new/tabla.html` (445 líneas) — Vista de Tabla

#### Estructura

```
<div x-data="fijados">
  <input type="hidden" id="active-module-val" value="{{.ActiveModule}}">
  <div class="overflow-x-auto">
    <table>
      <thead>
        <tr> — Columnas por módulo (9 variantes)
      </thead>
      <tbody id="tabla-cuerpo">
        {{if empty}} — Fila vacía
        {{else}} — Por cada fila:
          <tr> — Fila principal (click para expandir)
            <td> Botón +/- toggle
            <td> Botón Editar + Botón Fijar
            <td...> Columnas específicas del módulo
          </tr>
          <tr id="subfila-{{id}}" class="hidden"> — Subfila expandible
            <td colspan>
              IF expedientes: grid 3 columnas
              ELSE: {{template "tabla_subrow_trazabilidad"}}
              {{template "tabla_subrow_observaciones_notas"}}
          </tr>
      </tbody>
    </table>
  </div>
  PAGINACIÓN (if totalPages > 1):
    Primero, Anterior, Páginas (ventana=2), Siguiente, Último
</div>
<script> — Lógica de resize de columnas con localStorage
```

#### Columnas por Módulo

| Módulo | Columnas (además de Ver/Acción) |
|--------|--------------------------------|
| expedientes | SOLPED, Gerencia, Documento, Descripción, Estatus Detalle, Nro Proceso |
| requisiciones | Gerencia, Documento, Descripción, Equipo/Serial, Pase Sicesma, Estatus |
| memorandums | Gerencia, Documento, Asunto, Estatus, Recibido |
| recobros | Gerencia, Documento, Asunto, Estatus, Servicios |
| valuaciones | SOLPED, Gerencia, Documento, Descripción, Estatus, Monto Valuación |
| aprobacion_jd | SOLPED, Gerencia, Documento, Descripción, Estatus, Presupuesto (Bs) |
| certificacion_bdu | Gerencia, Documento, Presupuesto Total USD, Monto Contrato, Estatus |
| vacaciones | Gerencia, Documento, Año, Días, Desde, Hasta, Estatus |
| reposos_medicos | Gerencia, Documento, Días, Desde, Hasta, Estatus |

#### Pseudocódigo: Interacción de Filas

```
ON row click:
  Toggle visibilidad de subfila-{{id}}
  Toggle icono +/- en el botón

ON edit button click (stopPropagation):
  HTMX GET /api/cargar-expediente?modulo=X&id=Y
  → Response reemplaza #form-expediente
  → Al completar: abrir form-modal

ON pin button click (stopPropagation):
  toggle(id, solped, modulo)
  Guarda/quita de localStorage 'sidebarFrecuentes'
  Toggle color: emerald (fijado) vs blue (no fijado)
```

#### Pseudocódigo: Resize de Columnas

```
ON DOMContentLoaded:
  Leer localStorage 'col_widths_{module}'
  Para cada <th>:
    Si hay ancho guardado → aplicar a th + tds correspondientes
    Crear .col-resizer div adjunto al th
    ON mousedown en resizer:
      Track startX, startW
      ON mousemove → calcular diff, aplicar nuevo ancho
      ON mouseup → guardar en localStorage, fillLastColumn()
  fillLastColumn():
    Calcular ancho restante después de columnas fijas
    Aplicar a última columna si espacio > MIN_WIDTH
```

#### Pseudocódigo: Paginación

```
Paginación se renderiza cuando totalPages > 1:
  Ventana de 2 páginas alrededor de la actual
  Siempre mostrar página 1 y última
  Elipsis para gaps
  Cada botón: hx-get="/api/filtrar-expedientes" con hx-vals='{"pagina": "N"}'
  Página actual: disabled, fondo teal
  Botones activos: bg gray-700, hover teal-600
```

---

### 3.3 `templates/new/form.html` (276 líneas) — Formulario

#### Estructura

```
<div x-data="formularioModulo('{{.ActiveModule}}', {{jsonEncode .Registro}})">
  Input hidden ID
  Toggle "Orden Excel" botón
  <fieldset> "Información General" — Campos comunes
    Select Documento, Selects Gerencia/Superintendencia
    Campos específicos del módulo
  <fieldset> "Presupuesto" — (expedientes, aprobacion_jd)
    Fecha, monto (USD/Bs), tipo de cambio
  <fieldset> "Costos" — (recobros, valuaciones, certificacion_bdu)
    Varios campos de dinero
  <fieldset> "Gestión y Firmas" — (expedientes)
    Emisor, Receptor, Estatus, fechas, info contrato
  <fieldset> "Trazabilidad" — (no-expedientes)
    Emisor, Receptor, Estatus, fechas
  <fieldset> "Adjudicación" — (expedientes, valuaciones)
    Empresa, nros contrato, tiempo_ejecucion
  <fieldset> "Observaciones" — (todos)
    Observación auto-generada (readonly), observaciones manuales, notas
  Hidden #excel-order-container para modo orden Excel
</div>
{{template "form_buttons_alpine"}}
```

#### Sub-templates Usados

| Template | Propósito |
|----------|-----------|
| `form_hidden_id_alpine` | Input hidden para ID del registro |
| `form_input_text_alpine` | Input texto con x-model |
| `form_input_date_alpine` | Input fecha con @change handler |
| `form_input_number_alpine` | Input decimal con x-currency |
| `form_input_monto_alpine` | Input dinero con conversión USD/Bs |
| `form_input_spin_alpine` | Input numérico con botones +/- |
| `form_textarea_alpine` | Textarea con x-model |
| `form_select_alpine` | Select con opciones de catálogo |
| `form_gerencia_superintendencia_alpine` | Selects Gerencia→Superintendencia en cascada |
| `form_buttons_alpine` | Botones Guardar/Eliminar/Cancelar |
| `form_emisor_receptor_alpine` | Selects Emisor + Receptor |

#### Pseudocódigo: FormularioModulo

```
INIT:
  registro = registroInicial || {}
  IF no id_estatus → set to '1'
  Parse observaciones: split '\n---\n' → autoHistorial + manual
  _trackInitialState: {id_documento, fecha_recibido, fecha_devuelto, id_estatus}
  WATCH registro.tipo_cambio → _syncAll()
  WATCH registro.id_documento → _obsCambio('Documento')
  WATCH registro.fecha_recibido → _obsCambio('Fecha Recibido')
  WATCH registro.fecha_devuelto → _obsCambio('Fecha Devuelto')
  WATCH registro.id_estatus → _obsCambio('Estatus')

_obsCambio(label, oldVal, newVal):
  IF old == new → return
  Marcar label como pendiente
  Construir línea completa: "Fecha Devuelto: X, Documento: Y, Estatus: Z"
  Set autoObs = línea

prepararObservaciones():
  manual = registro.observaciones.trim()
  auto = autoObs
  IF auto AND manual → combined = auto + "\n---\n" + manual
  ELIF auto → combined = auto
  ELIF manual → combined = manual
  registro.observaciones = combined
  Update hidden input value
  Clear autoObs

_syncAll():
  tipo_cambio = parseNumber(registro.tipo_cambio)
  IF not valid → return
  _syncPair('presupuesto_base_bs', 'presupuesto_base_usd', tc)
  _syncPair('monto_adjudicado_bs', 'monto_adjudicado_usd', tc)

_syncPair(bsKey, usdKey, tc):
  bs = parseNumber(registro[bsKey])
  usd = parseNumber(registro[usdKey])
  source = lastSource[bsKey + '_' + usdKey]
  IF source == bsKey AND bs → usd = bs / tc
  ELIF source == usdKey AND usd → bs = usd * tc
  ELIF bs AND !usd → usd = bs / tc
  ELIF usd AND !bs → bs = usd * tc

onMontoInput(event, origen):
  Limpiar input: solo dígitos, puntos, comas, menos
  Detectar separador (última coma vs último punto)
  Limitar decimales a 2 lugares
  Guardar valor raw en dataset.raw
  Mapear origen a par:
    'bs_presup' → [presupuesto_base_bs, presupuesto_base_usd]
    'usd_presup' → [presupuesto_base_usd, presupuesto_base_bs]
    'bs_adj' → [monto_adjudicado_bs, monto_adjudicado_usd]
    'usd_adj' → [monto_adjudicado_usd, monto_adjudicado_bs]
  Llamar _conv(origenKey, destinoKey)

_conv(origenKey, destinoKey):
  tc = parseNumber(tipo_cambio)
  val = parseNumber(registro[origenKey])
  IF origenKey contiene '_usd' → result = val * tc
  ELSE → result = val / tc
  Set destinoKey value y update input display

validarFechas():
  Check pairs: [recibido/devuelto], [desde/hasta], [inicio/final]
  IF desde > hasta → toast error, clear hasta

validarAntesGuardar():
  Same date pair validation
  Return false si par inválido

toggleOrden():
  ordenExcel = !ordenExcel
  _reordenar():
    IF ordenExcel ON:
      Ocultar todos los fieldsets
      Buscar todos [data-orden-excel]
      Ordenar por valor data-orden-excel
      Mover a #excel-order-container
      Mostrar container
    ELSE:
      Ocultar container
      Restaurar elementos a padres originales
      Mostrar fieldsets

appendDias():
  IF tiempo_ejecucion no termina con "DIAS" → append " DIAS"

spinFrente(delta):
  cantidad_frentes = max(0, parseInt(cantidad_frentes) + delta)
```

---

### 3.4 `templates/new/components.html` (225 líneas) — Componentes Reutilizables

#### Templates Definidos

**`form_buttons_alpine`** (línea 1-21):
```
Barra sticky footer con:
  - Botón Cancelar → $store.modals.cerrar('form-modal')
  - Botón Eliminar (si .Registro existe):
      hx-post="/api/eliminar-expediente?modulo={{.Modulo}}"
      hx-include="#f-{{.IDColumna}}"
      hx-confirm="Confirmar eliminación..."
      On success → toast + cerrar modal + reload
  - Botón Guardar:
      hx-post="/api/guardar-expediente?modulo={{.Modulo}}"
      hx-include="#form-expediente"
      On before-request → prepararObservaciones(), validarAntesGuardar()
      On success → toast + cerrar modal + reload
```

**`form_input_text_alpine`** (línea 23-31):
```
<div data-orden-excel="{{.ExcelOrder}}">
  <label class="label">{{.Label}}</label>
  <input type="text" id="f-{{.Name}}" name="{{.Name}}"
         class="input" x-model="registro.{{.Name}}">
</div>
```

**`form_input_date_alpine`** (línea 33-41):
```
<div data-orden-excel="{{.ExcelOrder}}">
  <label class="label">{{.Label}}</label>
  <input type="date" id="f-{{.Name}}" name="{{.Name}}" class="input"
         :value="registro.{{.Name}}"
         @change="registro.{{.Name}} = $event.target.value; validarFechas()">
</div>
```

**`form_input_number_alpine`** (línea 43-52):
```
<div data-orden-excel="{{.ExcelOrder}}">
  <label class="label">{{.Label}}</label>
  <input type="text" inputmode="decimal" id="f-{{.Name}}" name="{{.Name}}"
         class="input" x-currency x-model="registro.{{.Name}}">
</div>
```

**`form_input_monto_alpine`** (línea 54-64):
```
<div data-orden-excel="{{.ExcelOrder}}">
  <label class="label">{{.Label}}</label>
  <input type="text" inputmode="decimal" id="f-{{.Name}}" name="{{.Name}}"
         class="input" x-currency :value="registro.{{.Name}}"
         @input="onMontoInput($event, '{{.Origen}}')">
</div>
```
Usa directiva `x-currency` para formateo y `onMontoInput` para conversión USD/Bs.

**`form_input_spin_alpine`** (línea 66-85):
```
<div data-orden-excel="{{.ExcelOrder}}">
  <label class="label">{{.Label}}</label>
  <div class="flex items-stretch">
    <input type="text" inputmode="numeric" class="input rounded-r-none"
           x-model="registro.{{.Name}}">
    <div class="flex flex-col">
      <button @click="spinFrente(1)">+</button>
      <button @click="spinFrente(-1)">-</button>
    </div>
  </div>
</div>
```

**`form_textarea_alpine`** (línea 87-94):
```
<div data-orden-excel="{{.ExcelOrder}}">
  <label class="label">{{.Label}}</label>
  <textarea id="f-{{.Name}}" name="{{.Name}}" rows="{{.Rows}}"
            class="input" x-model="registro.{{.Name}}"></textarea>
</div>
```

**`form_select_alpine`** (línea 96-112):
```
<div data-orden-excel="{{.ExcelOrder}}">
  <label class="label">{{.Label}}</label>
  <select id="f-{{.Name}}" name="{{.Name}}" class="input"
          x-model="registro.{{.Name}}">
    <option value=""></option>
    {{range catalog items}}
      <option value="{{.ID}}" data-id-gerencia="{{.IDGerencia}}">{{.Nombre}}</option>
    {{end}}
  </select>
</div>
```

**`form_gerencia_superintendencia_alpine`** (línea 114-134):
```
<div x-data="filtroSuperintendencias(catalogs.superintendencia, registro.id_gerencia)">
  {{template "form_select_alpine" for Gerencia con @change="gerenciaSeleccionada = $el.value"}}
  <div>
    <label>Superintendencia</label>
    <select x-model="superintendenciaSeleccionada"
            @change="registro.id_superintendencia = $el.value">
      <template x-for="sup in superintendenciasFiltradas">
        <option :value="sup.id" x-text="sup.nombre">
      </template>
    </select>
  </div>
</div>
```
Filtro en cascada: seleccionar Gerencia filtra opciones de Superintendencia por `id_gerencia`.

**`form_hidden_id_alpine`** (línea 141-145):
```
<input type="hidden" id="f-{{.IDColumna}}" name="{{.IDColumna}}"
       x-model="registro.{{.IDColumna}}">
```

**`tabla_empty_row`** (línea 147-149):
```
<tr><td colspan="{{.Colspan}}" class="p-12 text-center text-gray-500">
  No hay registros en la base de datos o que coincidan con la búsqueda.
</td></tr>
```

**`tabla_action_buttons_alpine`** (línea 151-168):
```
<td> Botón +/- toggle </td>
<td>
  Botón Editar: hx-get="/api/cargar-expediente?modulo=X&id=Y" → abrir form-modal
  Botón Fijar: :class basado en estaFijado(), @click=toggle()
</td>
```

**`tabla_subrow_trazabilidad`** (línea 170-191):
```
<div class="subrow-card">
  <h4>Trazabilidad y Fechas</h4>
  Grid: Receptor, Estatus, Fecha Recibido, Fecha Devuelto,
        Fecha Creación, Última Modificación
  Botón "Ver Historial":
    hx-get="/api/historial?modulo=X&id=Y"
    hx-target="#historial-cuerpo"
    Abre historial-modal
</div>
```

**`tabla_subrow_observaciones_notas`** (línea 193-225):
```
Línea de observación auto-generada (Fecha Devuelto/Recibido, Documento, Estatus)
Div clickable para abrir historial modal completo
Texto de última observación (parseado del campo observaciones)
Sección de notas (si existe)
```

---

### 3.5 `templates/new/ruta_procesos.html` (700 líneas) — Vista Gantt

#### Estructura

```
#gantt-chart-container
├── <fieldset> "Hoja" — Selector de hoja
│   ├── <select> para hoja (hojas de meses)
│   ├── Botón "Nueva Hoja" → toggleModal('crear-hoja-modal')
│   └── Botón "Eliminar Hoja" → eliminarHojaActual()
├── #juntas-container — Por cada Junta:
│   <fieldset class="junta-block">
│     <legend>Junta No. {{.Numero}}</legend>
│     ├── Campos editables: Junta Directiva (readonly), Número, Consecutiva, Fecha
│     ├── Botones Guardar/Eliminar
│     ├── TABLA GANTT:
│     │   <thead>:
│     │     Fila 1: fechas "desde" por semana
│     │     Fila 2: fechas "al" por semana
│     │     Fila 3: "SEMANA N" + botones agregar/quitar semana
│     │     Fila 4: Encabezados de días (L M X J V)
│     │   <tbody>:
│     │     Por cada Proceso:
│     │       Fila: número, nombre, 5 celdas de día por semana
│     │         Cada celda: onclick="abrirEditarCronograma(procId, fecha)"
│     │         Entradas coloreadas renderizadas por JS
│     │       Botón eliminar proceso
│     │     Fila "Añadir proceso"
│     ├── <fieldset> "Leyendas" (por junta)
│     │   Agregar/Editar/Eliminar/Reordenar/Bloquear items de leyenda
│     │   Cada uno: círculo de color, nombre, flechas arriba/abajo,
│     │             botones editar/eliminar
│   </fieldset>
├── Botón "NUEVA JUNTA"
└── MODALES (5):
    ├── #crear-hoja-modal — Crear nueva hoja
    ├── #crear-leyenda-modal — Crear/editar leyenda
    ├── #eliminar-semanas-modal — Eliminar semanas (checkboxes)
    ├── #agregar-semana-modal — Agregar semana (inputs fecha)
    └── #crono-modal — Editar cronograma de un día
        ├── Lista de entradas existentes con botón eliminar
        ├── Select leyenda + textarea nota
        └── Botón "Agregar"
```

#### JavaScript (Vanilla JS en IIFE)

Esta plantilla NO usa Alpine.js. Usa JavaScript vanilla en un IIFE.

```javascript
const data = {{jsonEncode .}};  // JSON server-rendered
const { hojas, current_hoja, juntas, current_junta, semanas,
        procesos, legend, junta_legend } = data;
var currentHojaId = current_hoja.id;
```

#### Estructuras de Datos

| Tipo | Campos |
|------|--------|
| `RutaProcesosGanttData` | Contenedor raíz |
| `RutaProcesosHoja` | id, nombre |
| `RutaProcesosJunta` | id, id_hoja, numero, consecutiva, fecha |
| `RutaProcesosJuntaSemana` | id, id_junta, numero, fecha_inicio, fecha_fin, dias[5] |
| `RutaProcesosJuntaProceso` | id, id_junta, numero, proceso, timeline |
| `RutaProcesosCronogramaEntry` | id, fecha, id_leyenda, nota, status_name, hex_color |
| `RutaProcesosLegend` | id, nombre, color, ambito, id_hoja, bloqueado |
| `RutaProcesosJuntaLeyenda` | id, id_junta, id_leyenda, orden |

#### Pseudocódigo: Comportamiento JavaScript Completo

```
IIFE():
  Parse data del servidor a variables locales
  currentHojaId = current_hoja.id || 0

  UTILIDADES:
    toggleModal(id): toggle class 'hidden' en elemento
    esc(v): HTML escape
    jsonPost(url, body): fetch POST con body form-encoded, return JSON
    reload(): htmx.ajax('GET', '/api/ruta-procesos?hoja=' + currentHojaId,
                         '#ruta-contenido')

  GESTIÓN DE HOJAS:
    cambiarHoja(): Leer select, update currentHojaId, reload()
    eliminarHojaActual(): Confirm → POST hoja-eliminar → reload()
    crearHoja(): Leer nombre → POST hoja-crear → cerrar modal, reload()

  GESTIÓN DE JUNTAS:
    guardarJunta(id): Leer campos del DOM → POST junta-actualizar → toast + reload()
    eliminarJunta(id): Confirm → POST junta-eliminar → reload()
    crearJunta(): POST junta-crear con numero auto-incrementado → reload()

  GESTIÓN DE SEMANAS:
    agregarSemana(idJunta): Pre-calcular próximo Lunes/Viernes → abrir modal
    guardarSemana(): Leer fechas → POST semana-agregar → cerrar modal, reload()
    abrirEliminarSemanas(idJunta): Renderizar checkboxes por semana → abrir modal
    eliminarSemanasConfirmar(): Recoger semanas marcadas → POST semana-eliminar → reload()

  GESTIÓN DE PROCESOS:
    agregarProceso(idJunta): prompt() para nombre → POST proceso-agregar → reload()
    eliminarProceso(id): Confirm → POST proceso-eliminar → reload()

  GESTIÓN DE LEYENDAS:
    editarLeyenda(id, nombre, color): Poblar campos del modal → abrir modal
    guardarLeyenda(): SI editando → POST update ELSE → POST create → cerrar, reload()
    eliminarLeyenda(id): Confirm → POST delete → reload()
    abrirCrearLeyenda(idJunta): Reset campos del modal → abrir modal
    moverLeyenda(idJunta, idLeyenda, dir): POST leyenda-reordenar → reload()
    toggleBloquearLeyenda(id): POST leyenda-bloquear → reload()

  GESTIÓN DE CRONOGRAMA:
    abrirEditarCronograma(procId, fecha):
      Set campos hidden (proc-id, fecha)
      Poblar select de leyendas desde array global
      Renderizar entradas existentes para este proc+fecha
      Abrir crono-modal
    guardarCronoDia(): Leer campos → POST cronograma-guardar → cerrar, reload()
    eliminarCronoEntry(id): Confirm → POST delete → reload()

  COLOREAR CELDAS GANTT:
    colorearCeldas():
      Por cada proceso:
        Por cada fecha en timeline:
          Buscar td[data-proc][data-fecha]
          Reemplazar 'gantt-cell-empty' con 'gantt-cell-active'
          Renderizar divs de entradas coloreadas dentro de la celda

  Ejecutar colorearCeldas() al inicio
```

---

### 3.6 `templates/historial.html` (63 líneas) — Historial de Cambios

#### Estructura

```
{{if no rows}} "Sin movimientos" mensaje
{{else}}
  <table>
    <thead>: (icono expandir), Fecha Recibido, Fecha Devuelto, Estatus,
             Documento, Emisor, Receptor (si no reposos), Observaciones, Notas
    <tbody>:
      Por cada fila:
        <tr> — Fila resumen (click para expandir)
        <tr id="hist-detalle-{{i}}" class="hidden"> — Fila detalle
          <td colspan="10">
            <div x-data="{ row: {{jsonEncode $row}} }">
              Grid de todos los campos (excluyendo id_*, observaciones, notas)
              Secciones separadas para observaciones y notas
            </div>
```

#### Pseudocódigo

```
Por cada fila del historial:
  Mostrar columnas resumen en una fila de tabla
  ON click en fila → toggle fila detalle oculta
  Fila detalle muestra:
    Grid de todos los campos no-ID, no-observación
    Texto completo de observaciones
    Texto completo de notas
```

---

### 3.7 `templates/pendientes.html` (34 líneas) — Documentos Pendientes

#### Estructura

```
{{if empty}} "No hay documentos pendientes" mensaje
{{else}}
  <table>
    <thead>: SOLPED, Descripción, Estatus, Receptor, Proceso, Empresa
    <tbody>:
      Por cada fila:
        <tr>: descripción truncada, badge de estatus con estatusClass(),
             receptor, nro_proceso, empresa
```

#### Template Functions Usadas

- `rowGetStr . "solped"` — Obtener valor string
- `truncate (rowGetStr . "descripcion_proceso") 50` — Truncar a 50 chars
- `estatusClass (rowGetStr . "estatus_detalle")` — Obtener clase de color de estatus
- `default (rowGetStr . "estatus_detalle") "-"` — Valor fallback

---

## 4. CSS — styles.css

### 4.1 Custom Properties (`:root`)

| Variable | Valor | Propósito |
|----------|-------|-----------|
| `color-scheme` | `dark` | Forzar modo oscuro |
| `--bg-body` | `#111827` (gray-900) | Fondo de página |
| `--bg-card` | `rgba(17,24,39,0.4)` | Fondo de tarjeta |
| `--bg-input` | `#1f2937` (gray-800) | Fondo de input |
| `--bg-hover` | `#374151` (gray-700) | Estado hover |
| `--bg-disabled` | `#374151` | Estado disabled |
| `--border-default` | `#374151` | Borde por defecto |
| `--border-card` | `rgba(55,65,81,0.6)` | Borde de tarjeta |
| `--border-focus` | `#14b8a6` (teal-500) | Anillo de foco |
| `--text-body` | `#f3f4f6` (gray-100) | Texto principal |
| `--text-muted` | `#9ca3af` (gray-400) | Texto secundario |
| `--text-dim` | `#6b7280` (gray-500) | Texto tenue |
| `--text-disabled` | `#6b7280` | Texto disabled |
| `--primary` | `#0d9488` (teal-600) | Color primario |
| `--primary-hover` | `#14b8a6` (teal-500) | Primario hover |
| `--primary-text` | `#2dd4bf` (teal-400) | Texto primario |
| `--danger` | `#b91c1c` (red-700) | Color peligro |
| `--danger-hover` | `#dc2626` (red-600) | Peligro hover |
| `--amber` | `#f59e0b` | Acento amber |
| `--blue` | `#60a5fa` | Acento azul |
| `--emerald` | `#34d399` | Acento esmeralda |
| `--purple` | `#a78bfa` | Acento púrpura |
| `--radius` | `0.5rem` | Radio por defecto |
| `--radius-lg` | `0.5rem` | Radio grande |
| `--bg-surface-800` | `#1f2937` | Fondo de superficie |
| `--bg-surface-900` | `#111827` | Superficie profunda |
| `--shadow-sm/lg/xl` | Varios | Sombras |
| `--transition-fast/normal` | `0.15s/0.3s ease` | Transiciones |
| `--z-sticky` | `10` | z-index sticky |
| `--z-tooltip` | `50` | z-index tooltip |

### 4.2 Paleta de Colores (Dark Mode)

```
Fondos:
  gray-900 (#111827) → body, superficies profundas
  gray-800 (#1f2937) → cards, inputs, superficies
  gray-700 (#374151) → bordes, hover states

Texto:
  gray-100 (#f3f4f6) → texto principal
  gray-400 (#9ca3af) → texto secundario
  gray-500 (#6b7280) → texto tenue

Acentos:
  teal-400 (#2dd4bf) → texto primario, encabezados
  teal-500 (#14b8a6) → foco, hover
  teal-600 (#0d9488) → botones primarios
  emerald-400 (#34d399) → estado adjudicado
  amber-400 (#fbbf24) → presupuesto, acentos
  red-700 (#b91c1c) → eliminar, peligro

Bordes:
  gray-600 (#4b5563) → bordes de inputs
  gray-700 (#374151) → bordes de tarjetas
```

---

## 5. JavaScript — Alpine.js

### 5.1 Stores Globales

**`Alpine.store('modals')`** — Gestión centralizada de modales
```
State:
  stack: []  — Array de IDs de modales abiertos
Propiedades:
  abierto: boolean  — true si hay alguna modal abierta
Métodos:
  abrir(id): Push id al stack, set body overflow hidden
  cerrar(id): Remover id del stack, restaurar overflow si vacío
  toggle(id): Toggle abrir/cerrar
  tiene(id): Verificar si id está en el stack
  cerrarClickFuera(e, id): Cerrar si click target === currentTarget
```

**`Alpine.store('toast')`** — Sistema de notificaciones
```
Métodos:
  mostrar(msg, tipo='info'): Crear div toast, animar entrada,
    auto-remover después de 3s
  error(msg): mostrar(msg, 'error')
  success(msg): mostrar(msg, 'success')
  info(msg): mostrar(msg, 'info')
```

**`Alpine.store('fijados')`** — Registros fijados (persistido en localStorage)
```
Storage key: 'sidebarFrecuentes'
State:
  lista: []  — Array de {id, solped, modulo}
Métodos:
  init(): Cargar de localStorage
  guardar(): Guardar en localStorage
  toggle(id, solped, modulo): Agregar si no presente, remover si presente
  estaFijado(id, modulo): Verificar si registro está fijado
  eliminar(idx): Remover por índice
```

### 5.2 Components (Alpine.data)

**`appShell`** — Componente raíz del body
```
Propósito: Drag-and-drop para abrir archivos .db
State: dragOver = false
Métodos: onDrop(e) → POST /api/abrir-bd con archivo arrastrado
```

**`fijados`** — Bridge al store fijados
```
Propósito: Proxy del store para uso en templates
Properties: lista → store.lista
Métodos: toggle(), estaFijado(), eliminar()
```

**`bdRecientes`** — BD recientes (persistido en localStorage)
```
Storage key: 'baseaccess_recientes'
State: lista: [{nombre, path, timestamp}]
Métodos:
  init(): Cargar, registrar BD actual, auto-abrir si solo 1
  registrar(nombre, path): Agregar/mover al tope, cap en 5
  async abrir(path): POST /api/abrir-bd, reload on success
  eliminarPorIndex(idx), eliminarPorPath(path)
```

**`calculadoraSumas`** — Calculadora de sumas
```
State:
  filas: [{valor: ''}]
  resultado: 0
  fijados: []
Métodos:
  onInput(idx, event): Limpiar input, detectar separador, recalcular
  onBlur(idx, event): Formatear con locale español
  calcular(): Sumar todas las filas
  añadirFila(), quitarFila(idx), limpiar()
  fijarResultado(), quitarFijado(idx)
  totalFijados: getter → suma de fijados
  formatearNum(v): Formatear con locale español (2 decimales)
```

**`exportarExcel`** — Exportar a Excel con selección de columnas
```
State:
  modulo: 'expedientes'
  columnas: [{nombre, seleccionada}]
  filtros: {}
  fechaDesde, fechaHasta: ''
  cargando: false
Métodos:
  init(): Watch modulo → cargarColumnas()
  async cargarColumnas(): Fetch /api/columnas-modulo, renderizar checkboxes
  validarFechas(): Validar desde <= hasta
  async exportar(): Fetch /api/exportar-excel, descargar blob XLSX
```

**`filtroSuperintendencias`** — Filtro en cascada Gerencia→Superintendencia
```
Props: supOpts[], gerInicial
State: gerenciaSeleccionada, superintendenciaSeleccionada
Getter: superintendenciasFiltradas → filtra por id_gerencia
```

**`formularioModulo`** — Formulario principal con conversión de moneda
```
Props: modulo, registroInicial
State:
  registro: {}  — Datos del formulario (two-way binding)
  autoObs: ''  — Observación auto-generada
  lastSource: {}  — Tracking de campo editado para conversión
  ordenExcel: false  — Modo orden Excel
Métodos: (ver pseudocódigo detallado en sección 3.3)
  init(), prepararObservaciones(), validarFechas()
  _syncAll(), _syncPair(), onMontoInput(), _conv()
  toggleOrden(), _reordenar(), appendDias(), spinFrente()
```

### 5.3 Directiva Custom

**`x-currency`** (alpine-directives.js — 48 líneas)
```
Propósito: Formateo automático de números para inputs monetarios
Uso: <input x-currency> o <input x-currency.es-ES>

Lifecycle:
  onInput:
    Reemplazar comas con puntos
    Remover chars no numéricos (excepto puntos)
    Mantener solo último punto
    Limitar a 2 decimales
    Guardar valor raw en el.dataset.raw
  onBlur:
    Formatear con toLocaleString(locale, {min/maxFractionDigits:2})
  onFocus:
    Restaurar valor raw para edición
```

### 5.4 Puente HTMX-Alpine

**`alpine-htmx-bridge.js`** (25 líneas)
```
Evento: htmx:afterSettle
  Handler: Alpine.initTree(evt.detail.target)
  Propósito: Cuando HTMX inyecta HTML, Alpine necesita inicializar
             los x-data, x-bind, etc. en los nuevos elementos

Evento: htmx:afterSwap
  Condición: evt.detail.target.id === 'vista-tabla'
  Handler:
    1. Re-inicializar store fijados (recargar de localStorage)
    2. Actualizar colores de botones de fijar:
       - Para cada [id^="pin-btn-"]:
         - Extraer ID del botón
         - Si ID está en fijados.lista → color emerald
         - Else → color blue
```

---

## 6. Glosario de Clases CSS

### Botones

| Clase | Propósito |
|-------|-----------|
| `.btn` | Botón base: font 0.9375rem, weight 600, radius var(--radius-lg), sin borde |
| `.btn:disabled` | Estado disabled: cursor not-allowed, opacity 0.5 |
| `.btn-primary` | Acción primaria: bg var(--primary), text white |
| `.btn-secondary` | Acción secundaria: bg var(--bg-hover), text gray-200 |
| `.btn-danger` | Acción destructiva: bg var(--danger), text white |
| `.btn-sm` | Botón pequeño: padding 0.375rem 0.75rem, font 0.75rem |
| `.btn-lg` | Botón grande: padding 0.5rem 2rem |
| `.btn-icon` | Solo icono: sin bg, sin borde, cursor pointer |
| `.btn-amber-outline` | Variante amber: borde 1px solid var(--amber) |

### Inputs

| Clase | Propósito |
|-------|-----------|
| `.input` | Input base: width 100%, bg var(--bg-input), borde, radius, padding |
| `.input:focus` | Estado foco: border-color var(--border-focus) |
| `.input-lg` | Input grande: padding más grande |
| `.input-auto` | Ancho automático: width auto !important |

### Tarjetas y Layout

| Clase | Propósito |
|-------|-----------|
| `.card` | Tarjeta: bg var(--bg-card), padding 1rem, radius, borde |
| `.card-amber` | Variante amber |
| `.legend` | Legend de fieldset: uppercase, bold, 0.875rem |
| `.legend-teal/amber/blue/emerald/purple` | Variantes de color |
| `.label` | Label de formulario: block, 0.8125rem, color muted |
| `.subrow-card` | Tarjeta de fila expandida: bg gray-800/80, padding, borde |
| `.subrow-card h4` | Encabezado de sección: uppercase, teal, border-bottom |

### Modales

| Clase | Propósito |
|-------|-----------|
| `.modal-content` | Contenido modal: bg var(--bg-surface-800), radius, borde, shadow |
| `.modal-body` | Cuerpo modal: padding 1rem |
| `.modal-panel-header` | Header sticky: flex between, borde inferior, radius superior |
| `.modal-panel-footer` | Footer sticky: flex end, borde superior, radius inferior |

### Texto

| Clase | Propósito |
|-------|-----------|
| `.text-link` | Texto teal, bold, hover más claro |
| `.text-edit` | Texto azul, hover más claro |

### Toasts

| Clase | Propósito |
|-------|-----------|
| `.toast` | Toast base: padding, radius, shadow, opacity 0, translateX(100%) |
| `.toast.show` | Visible: opacity 1, translateX(0) |
| `.toast.info` | Fondo slate, borde slate |
| `.toast.success` | Fondo verde, borde emerald |
| `.toast.error` | Fondo rojo, borde red |
| `.toast.warning` | Fondo amber, borde amber |

### Gantt Chart

| Clase | Propósito |
|-------|-----------|
| `.gantt-table` | table-layout fixed, border-collapse |
| `.gantt-col-num` | Columna sticky 44px, z-index 2 |
| `.gantt-col-day` | Columna de día 32px, altura 40px |
| `.gantt-week-header` | 9px, texto teal, bg semi-transparente |
| `.gantt-week-subheader` | 9px, texto más claro |
| `.gantt-day-header` | 10px bold |
| `.gantt-cell-empty` | bg transparente, cursor pointer |
| `.gantt-cell-active` | Cursor pointer, posición relative |
| `.gantt-cell-entries` | Flex column, gap 3px |
| `.gantt-cell-entry` | min-height 12px, radius 3px |
| `.gantt-tooltip` | Tooltip posicionado absolute, 220px ancho |
| `.gantt-legend-grid` | Grid 2 columnas (5 en desktop) |
| `.gantt-legend-item` | Fila flex, card bg, efecto hover |
| `.gantt-legend-circle` | Círculo de color 24px |

### Tarjetas de Proceso

| Clase | Propósito |
|-------|-----------|
| `.proceso-card` | Fila flex con hover border |
| `.proceso-card-num` | Círculo teal con número |
| `.proceso-card-desc` | Texto descriptivo del proceso |
| `.proceso-card-badge` | Badge de estado |
| `.proceso-card-edit-btn` | Botón de editar |

### Scrollbar

```css
* { scrollbar-width: thin; scrollbar-color: var(--bg-hover) transparent; }
*::-webkit-scrollbar { width/height: 8px; }
*::-webkit-scrollbar-track { transparent; }
*::-webkit-scrollbar-thumb { var(--bg-hover); border-radius: 4px; }
*::-webkit-scrollbar-thumb:hover { var(--border-focus); }
```

### Bottom Bar

```css
#bottom-bar: fixed bottom, z-index 20
.btn-mod: sin border-radius, padding/font más pequeño
.inner-bar .btn-mod:first-child: border-radius izquierdo
.inner-bar .btn-mod:last-child: border-radius derecho
.ruta-mod: border-radius, borde amber
body.has-db: padding-bottom 3rem (espacio para bottom bar)
```

### Column Resizer

```css
.col-resizer: absolute right, 10px ancho, cursor col-resize
.col-resizer::after: barra 3px ancho, 70% altura, gris sutil
.col-resizer:hover::after: resaltado teal
```

### Otros

| Clase | Propósito |
|-------|-----------|
| `.reciente-item` | Item de BD reciente con transición hover |
| `.drag-over` | Outline teal punteado al arrastrar archivo |
| `.campo-frecuente .label` | Borde izquierdo amber + indicador punto |
| `#spinner-overlay.htmx-request` | Mostrar spinner cuando HTMX carga |
| `.exp-checkbox-label` | Checkbox premium para modal de exportar |
| `.exp-col` | Checkbox custom: 18px, rounded, estado checked teal |
| `.suma-fila` | Espaciado de fila de calculadora |
| `.leyenda-color-wrap input[type="color"]` | Selector de color circular |

---

## 7. Glosario de Componentes Alpine

| Componente | Tipo | Propósito |
|------------|------|-----------|
| `appShell` | `Alpine.data` | Drag-and-drop para abrir .db |
| `fijados` | `Alpine.data` | Proxy del store fijados |
| `bdRecientes` | `Alpine.data` | BD recientes (localStorage) |
| `calculadoraSumas` | `Alpine.data` | Calculadora de sumas |
| `exportarExcel` | `Alpine.data` | Exportar Excel con filtros |
| `filtroSuperintendencias` | `Alpine.data` | Filtro cascada Gerencia→Superintendencia |
| `formularioModulo` | `Alpine.data` | Formulario principal con conversión moneda |
| `modals` | `Alpine.store` | Gestión de stack de modales |
| `toast` | `Alpine.store` | Notificaciones toast |
| `fijados` (store) | `Alpine.store` | Registros fijados (localStorage) |

---

## 8. Pseudocódigo General

### Flujo Completo de Interacción

```
USUARIO ABRE APP:
  1. Browser → GET /
  2. Go → preparePageData() → PageData
  3. Template renderiza index.html + tabla.html
  4. Alpine inicializa stores y componentes
  5. bdRecientes carga BDs recientes de localStorage

USUARIO ABRE BD (desde teléfono):
  1. Click "Abrir Base de Datos"
  2. Fallback HTML: input[type=file] se abre
  3. Usuario selecciona archivo .db
  4. POST /api/abrir-bd con archivo
  5. Go abre SQLite, inicializa schema
  6. location.reload() → página se recarga con datos

USUARIO NAVEGA MÓDULOS:
  1. Click botón módulo en bottom bar
  2. hx-get="/api/cambiar-modulo?modulo=X"
  3. Go consulta vista del módulo, renderiza tabla
  4. HTMX reemplaza #vista-tabla
  5. alpine-htmx-bridge re-inicializa Alpine
  6. Column widths se restauran de localStorage

USUARIO BUSCA/FILTRA:
  1. Escribe en input búsqueda
  2. Debounce 200ms → hx-get="/api/filtrar-expedientes"
  3. Go filtra filas, pagina, renderiza tabla
  4. HTMX reemplaza #vista-tabla

USUARIO CREA/EDITA REGISTRO:
  1. Click "Nuevo Registro" o botón editar
  2. hx-get="/api/cargar-expediente" → Go renderiza form.html
  3. HTMX inyecta en #form-expediente
  4. Alpine inicializa formularioModulo con datos
  5. Modal se abre
  6. Usuario edita (Alpine two-way binding)
  7. Click "Guardar"
  8. hx-post="/api/guardar-expediente" con datos del form
  9. Go guarda en SQLite, retorna JSON
  10. Toast success, cerrar modal, reload

USUARIO USA RUTA PROCESOS:
  1. Click "Ruta Procesos" en bottom bar
  2. hx-get="/api/ruta-procesos" → Go renderiza ruta_procesos.html
  3. HTMX inyecta en #ruta-contenido
  4. JS IIFE parsea data del servidor
  5. colorearCeldas() marca celdas Gantt
  6. Interacciones: fetch() POST a /api/ruta-procesos-*
  7. reload() → hx.ajax re-carga el contenido
```
