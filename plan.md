# PLAN — wails-migration: exportar Excel inteligente + mejoras Ruta Procesos (Gantt)

> Cambio de modelo: este plan es **auto-contenido** y reproducible por cualquier agente sin contexto previo. Todas las decisiones ya están tomadas. Solo queda ejecutar.

---

## 0. CONTEXTO DEL PROYECTO

- **Repo**: `/home/user/Documentos/proyecto/baseaccess`, rama **`wails-migration`** (rama paralela a `master`/`tauri-migration`).
- **App**: "Control de Documentos Presidencia" — desktop app Wails v2 (Go + WebView) para registrar documentos con historial de movimientos.
- **Stack**: Wails v2 + Go 1.25 + `mattn/go-sqlite3` + Go `html/template` + HTMX + Tailwind CSS (sin frameworks JS).
- **Multi-módulo (9)**: `expedientes`, `requisiciones`, `memorandums`, `recobros`, `valuaciones`, `aprobacion_jd`, `certificacion_bdu`, `vacaciones`, `reposos_medicos`. Cada módulo tiene su propia tabla, vista (`vw_reporte_*`), tabla de historial (`hist_*`) y triggers. Definidos en `var Modulos map[string]ModuloConfig` en `app.go`.
- **BD**: `data/expedientes.db` (gitignored). Ya tiene el schema completo (01_master + 02_modulos aplicados, 13 gerencias, 17 superintendencias). 36 registros en `expedientes`, 0 en módulos nuevos.

### Estado actual de la rama (antes de ejecutar este plan)

- **HEAD**: commit `cb05e50` "refactor: rename Contrataciones to Control Docs. Presidencia".
- **git status**: limpio (todo commiteado y pusheado a `origin/wails-migration`).
- **Tablas en BD**: `expedientes`, `req_materiales`, `memorandums`, `recobros`, `valuaciones`, `aprobacion_jd`, `certificacion_bdu`, `vacaciones`, `reposos_medicos` + sus `hist_*` + vistas `vw_reporte_*`.
- **Vistas `vw_reporte_*`**: devuelven columnas ya resueltas (JOIN con catálogos: gerencia, superintendencia, emisor, receptor, estatus, documento, empresa, resultado, etc.). Son la fuente de datos tanto para la UI como para el export.

---

## 1. OBJETIVOS (prioridades del user)

| # | Feature | Prioridad | Estado |
|---|---|---|---|
| 1 | **Exportar Excel (XLSX) inteligente** — módulo/hoja, filtro filas, columnas, rango fechas | **ALTA** (la más importante según el user) | Pendiente |
| 2 | **Ruta Procesos: elegir/eliminar procesos** + Gantt auto-update | MEDIA | Pendiente |
| 3 | **Ruta Procesos: añadir leyendas con colores** al Gantt | BAJA (dejada pendiente — "hay una más importante") | Posponerse |
| 4 | **Modales se cierran al clickear afuera** (con jerarquía: si A abre B, al cerrar B no se cierra A) | MEDIA | Pendiente |
| 5 | **Gerencias permitidas por módulo** en formularios (cada módulo solo muestra las gerencias que le corresponden) | ALTA | Pendiente |
| 6 | **Botonera inferior alineada al fondo** real de la ventana (no justo debajo de la tabla) | MEDIA | Pendiente |
| 7 | **Botón "Sumas"** — calculadora que sume varios números (solo 2 decimales) y muestre el resultado | MEDIA | Pendiente |

El user dijo: "ahora hagamos algunas nuevas funciones... pero esa [leyendas] dejala pendiente porque hay una mas importante, que al darle csv... que el excel final sea mas inteligente". Después: "que los menu flotante se cierren al clickear afuera, pero que por ejemplo si es un menu flotante que viene de otro menu flotante, que si cierro el mas nuevo el de antes siga abierto". Y después:dió las gerencias permitidas por módulo + alineación de la botonera inferior + botón "Sumas" con 2 decimales.

**Orden de ejecución** (revisado): Feature #1 (Excel) → Feature #4 (Modales) → Feature #5+#6+#7 (un commit, toca formularios y CSS) → Feature #2 (Ruta Procesos). Feature #3 queda documentado al final como "próximo PR".

---

## 2. FEATURE #1 — EXPORTAR EXCEL (XLSX) INTELIGENTE

### 2.1 Reemplazo del CSV actual por XLSX configurable

**Estado actual** (`handler.go:682-720`, `templates/index.html:81-82,677-678`):
- Ruta `/api/csv` (GET) que descarga SIEMPRE el módulo `expedientes` con TODAS las columnas y TODAS las filas, sin filtros. Devuelve `text/csv`.
- Botón "CSV" en `index.html` que llama `exportarCSV()` → `window.location.href = '/api/csv'`.

**Problemas**:
1. Solo exporta `expedientes`. El user quiere elegir la hoja/módulo (9 opciones).
2. No filtra filas (search, rango fechas).
3. No selecciona columnas (exporta todas, incluyendo IDs internos tipo `id_gerencia` que en la vista ya están resueltos como `gerencia` con nombre legible).
4. Formato CSV (no es .xls/.xlsx que el user pidió explícitamente).

**Objetivo**: reemplazar `/api/csv` por `/api/exportar-excel` (GET) que genere un `.xlsx` con:
- Una hoja por módulo seleccionado, O una sola hoja con el módulo activo.
- Filas filtradas por: texto de búsqueda (`q`), rango de fechas (`fecha_desde`/`fecha_hasta` aplicado a `fecha_recibido`).
- Columnas seleccionadas por el user (checkboxes en un modal de configuración).
- Cabecera con nombres legibles (los alias de las vistas `vw_reporte_*`, ej: `Gerencia` en vez de `id_gerencia`).

### 2.2 Librería Go para XLSX

**Decisión**: usar `github.com/xuri/excelize/v2` — la librería Go más popular y mantenida para XLSX. Soporta múltiples hojas, estilos, anchos de columna, freeze panes, filtros automáticos.

**Acción**: añadir a `go.mod` con `go get github.com/xuri/excelize/v2`. Es una dependencia directa (no indirecta).

### 2.3 Backend — nuevo endpoint `/api/exportar-excel`

**Archivo nuevo**: NO se crea un archivo nuevo. Se añade la función a `handler.go` (siguiendo el patrón de los handlers existentes).

**Ruta** (en `handler.go:287-324`, dentro del switch de `ServeHTTP`):
```go
case p == "/api/exportar-excel" && r.Method == http.MethodGet:
    h.handleExportarExcel(w, r)
    return
```

**Handler `handleExportarExcel`** (añadir al final de `handler.go`, antes del bloque de JSON helpers):
```go
func (h *TemplateHandler) handleExportarExcel(w http.ResponseWriter, r *http.Request) {
    modulo := r.URL.Query().Get("modulo")
    if modulo == "" {
        modulo = "expedientes"
    }
    cfg, ok := Modulos[modulo]
    if !ok {
        http.Error(w, "modulo invalido", http.StatusBadRequest)
        return
    }

    // Filtros
    q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
    fechaDesde := r.URL.Query().Get("fecha_desde")
    fechaHasta := r.URL.Query().Get("fecha_hasta")

    // Columnas seleccionadas (comma-separated). Si vacío, todas las de la vista.
    columnasParam := r.URL.Query().Get("columnas")
    var columnasSel []string
    if columnasParam != "" {
        columnasSel = strings.Split(columnasParam, ",")
    }

    // 1. Obtener filas (vista ya con JOINs resueltos)
    filas, err := h.app.ObtenerFilas(modulo, cfg.IDColumna+" DESC")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 2. Filtrar en Go (igual que handleFiltrarExpedientes)
    var filtered []Row
    for _, row := range filas {
        // Filtro texto
        if q != "" {
            matches := false
            for _, val := range row {
                if val != nil {
                    sVal := strings.ToLower(fmt.Sprintf("%v", val))
                    if strings.Contains(sVal, q) { matches = true; break }
                }
            }
            if !matches { continue }
        }
        // Filtro fecha_recibido
        fr, _ := row["fecha_recibido"].(string)
        if fechaDesde != "" && fr < fechaDesde { continue }
        if fechaHasta != "" && fr > fechaHasta { continue }
        filtered = append(filtered, row)
    }

    if len(filtered) == 0 {
        http.Error(w, "no hay datos para exportar con los filtros aplicados", http.StatusBadRequest)
        return
    }

    // 3. Determinar columnas finales
    keysOrdered := make([]string, 0, len(filtered[0]))
    for k := range filtered[0] {
        keysOrdered = append(keysOrdered, k)
    }
    sort.Strings(keysOrdered) // orden alfabético estable
    // Mover IDColumna al principio
    for i, k := range keysOrdered {
        if k == cfg.IDColumna {
            keysOrdered = append(keysOrdered[:i], keysOrdered[i+1:]...)
            keysOrdered = append([]string{cfg.IDColumna}, keysOrdered...)
            break
        }
    }
    // Filtrar a las columnasSel si se especificaron
    if len(columnasSel) > 0 {
        sel := map[string]bool{}
        for _, c := range columnasSel { sel[c] = true }
        filtered_keys := make([]string, 0, len(columnasSel))
        for _, k := range keysOrdered {
            if sel[k] { filtered_keys = append(filtered_keys, k) }
        }
        keysOrdered = filtered_keys
    }

    // 4. Mapear IDs de columnas a labels legibles (snake_case → Title Case)
    labelOf := func(k string) string {
        words := strings.Split(strings.ReplaceAll(k, "_", " "), " ")
        for i, w := range words {
            if len(w) > 0 { words[i] = strings.ToUpper(w[:1]) + w[1:] }
        }
        return strings.Join(words, " ")
    }

    // 5. Generar XLSX con excelize
    f := excelize.NewFile()
    sheetName := cfg.Nombre
    if len(sheetName) > 31 { sheetName = sheetName[:31] } // límite Excel
    f.SetSheetName(f.GetSheetName(0), sheetName)

    // Cabecera (fila 1)
    for i, k := range keysOrdered {
        cell, _ := excelize.CoordinatesToCellName(i+1, 1)
        f.SetCellValue(sheetName, cell, labelOf(k))
    }
    // Estilo cabecera (negrita + fondo teal)
    styleID, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
        Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#0F766E"}},
    })
    lastCell, _ := excelize.CoordinatesToCellName(len(keysOrdered), 1)
    f.SetCellStyle(sheetName, "A1", lastCell, styleID)

    // Filas de datos (desde fila 2)
    for ri, row := range filtered {
        for ci, k := range keysOrdered {
            cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
            v := row[k]
            if v == nil { continue }
            f.SetCellValue(sheetName, cell, v)
        }
    }

    // Auto-ancho aproximado
    for i, k := range keysOrdered {
        width := float64(len(labelOf(k))) + 4
        for _, row := range filtered {
            if v := row[k]; v != nil {
                s := fmt.Sprintf("%v", v)
                if len(s) > int(width) { width = float64(len(s)) + 2 }
            }
        }
        colName, _ := excelize.ColumnNumberToName(i + 1)
        f.SetColWidth(sheetName, colName, colName, width)
    }

    // Freeze panes (cabecera fija)
    f.SetPanes(sheetName, &excelize.Panes{
        Freeze: true, YSplit: 1,
        TopLeftCell: "A2", ActivePane: "bottomLeft",
    })

    // Filtros automáticos en cabecera
    f.AutoFilter(sheetName, "A1:"+lastCell, []excelize.AutoFilterOptions{})

    // 6. Escribir al Response
    filename := cfg.Nombre + "_" + time.Now().Format("2006-01-02") + ".xlsx"
    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
    if err := f.Write(w); err != nil {
        log.Printf("exportar-excel: error escribiendo: %v", err)
    }
}
```

**Imports a añadir en `handler.go`** (verificar antes que no estén ya):
- `"github.com/xuri/excelize/v2"`
- `"sort"`
- `"time"`

**Verificar**:IÓN con `go vet ./...` y `go build ./...` tras añadir el handler y los imports.

### 2.4 Backend — mantener `/api/csv` por compatibilidad

**Decisión**: NO eliminar `/api/csv`. Se mantiene por si algo lo referencia. Solo se añade `/api/exportar-excel` como nueva ruta paralela. El botón de la UI se migra al nuevo flujo (ver §2.5).

### 2.5 Frontend — modal de configuración de exportación

**Acción en `templates/index.html`**:

1. Reemplazar el botón "CSV" (línea 81-83) por "Exportar Excel" que abre un modal de configuración.
2. Añadir el modal al final del `<body>` (antes de los `<script>`s finales).

**Botón reemplazo** (línea 81-83):
```html
<button onclick="abrirModalExportar()" {{if not .HasDB}}disabled{{end}} class="btn btn-primary" id="btn-exportar">
    <i class="fas fa-file-excel mr-1"></i> Exportar Excel
</button>
```

**Modal** (añadir antes de `</body>`):
```html
<div id="export-modal" class="hidden fixed inset-0 z-50 flex items-start justify-center p-4 overflow-y-auto bg-black/70">
    <div class="relative modal-content w-full max-w-2xl my-8">
        <div class="sticky top-0 bg-gray-800 z-10 flex items-center justify-between p-4 border-b border-gray-700 rounded-t-xl">
            <h2 class="text-lg font-bold text-teal-400"><i class="fas fa-file-excel mr-2"></i>Exportar a Excel</h2>
            <button onclick="cerrarModalExportar()" class="btn-icon text-gray-400 hover:text-white"><i class="fas fa-times"></i></button>
        </div>
        <div class="p-6 space-y-5">
            <!-- Módulo/Hoja -->
            <div>
                <label class="label">Hoja a exportar</label>
                <select id="exp-modulo" class="input">
                    {{$first := true}}
                    {{range $key, $cfg := .Modulos}}
                    <option value="{{$key}}">{{$cfg.Nombre}}</option>
                    {{end}}
                </select>
            </div>
            <!-- Filtro texto -->
            <div>
                <label class="label">Filtro de búsqueda (opcional)</label>
                <input type="text" id="exp-q" class="input" placeholder="Filtrar filas que contengan...">
            </div>
            <!-- Rango fechas -->
            <div class="grid grid-cols-2 gap-4">
                <div><label class="label">Fecha Desde (recibido)</label><input type="date" id="exp-fecha-desde" class="input"></div>
                <div><label class="label">Fecha Hasta (recibido)</label><input type="date" id="exp-fecha-hasta" class="input"></div>
            </div>
            <!-- Columnas (se cargan dinámicamente según módulo) -->
            <div>
                <label class="label">Columnas a exportar (vacío = todas)</label>
                <div class="flex gap-2 mb-2">
                    <button type="button" onclick="toggleTodasColumnas(true)" class="btn btn-secondary text-xs">Seleccionar todas</button>
                    <button type="button" onclick="toggleTodasColumnas(false)" class="btn btn-secondary text-xs">Limpiar</button>
                </div>
                <div id="exp-columnas" class="grid grid-cols-2 md:grid-cols-3 gap-2 max-h-48 overflow-y-auto p-3 border border-gray-700 rounded-lg"></div>
            </div>
        </div>
        <div class="sticky bottom-0 bg-gray-800 border-t border-gray-700 p-4 flex justify-end gap-3 rounded-b-xl">
            <button onclick="cerrarModalExportar()" class="btn btn-secondary">Cancelar</button>
            <button onclick="ejecutarExportar()" class="btn btn-primary"><i class="fas fa-download mr-1"></i> Descargar XLSX</button>
        </div>
    </div>
</div>
```

**JS** (añadir al bloque `<script>` de `index.html`, después de `exportarCSV` actual):

```js
function abrirModalExportar() {
    $('export-modal').classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    cargarColumnasExportar();
}
function cerrarModalExportar() {
    $('export-modal').classList.add('hidden');
    document.body.style.overflow = '';
}
async function cargarColumnasExportar() {
    const modulo = $('exp-modulo').value;
    const cont = $('exp-columnas');
    cont.innerHTML = '<p class="text-gray-500 text-xs col-span-full">Cargando columnas...</p>';
    const res = await fetch('/api/columnas-modulo?modulo=' + encodeURIComponent(modulo));
    if (!res.ok) { cont.innerHTML = '<p class="text-red-400 text-xs col-span-full">Error al cargar columnas</p>'; return; }
    const cols = await res.json();
    cont.innerHTML = '';
    cols.forEach(c => {
        const lbl = document.createElement('label');
        lbl.className = 'flex items-center gap-2 text-xs text-gray-300 cursor-pointer';
        lbl.innerHTML = `<input type="checkbox" value="${c}" class="exp-col"> ${c.replace(/_/g, ' ')}`;
        cont.appendChild(lbl);
    });
}
function toggleTodasColumnas(sel) {
    document.querySelectorAll('.exp-col').forEach(cb => cb.checked = sel);
}
function ejecutarExportar() {
    const modulo = $('exp-modulo').value;
    const q = $('exp-q').value;
    const fd = $('exp-fecha-desde').value;
    const fh = $('exp-fecha-hasta').value;
    const cols = Array.from(document.querySelectorAll('.exp-col:checked')).map(cb => cb.value);
    const params = new URLSearchParams();
    params.set('modulo', modulo);
    if (q) params.set('q', q);
    if (fd) params.set('fecha_desde', fd);
    if (fh) params.set('fecha_hasta', fh);
    if (cols.length) params.set('columnas', cols.join(','));
    window.location.href = '/api/exportar-excel?' + params.toString();
    cerrarModalExportar();
}
$('exp-modulo').addEventListener('change', cargarColumnasExportar);
```

### 2.6 Backend — endpoint auxiliar `/api/columnas-modulo`

Para que el modal sepa qué columnas ofrece cada módulo (sin pedirselo al user a mano), se añade un endpoint mínimo que devuelve las columnas de la vista `cfg.Vista` consultando una fila vacía o usando `PRAGMA table_info`.

**Decisión**: usar `PRAGMA table_info(<vista>)` no funciona para vistas en SQLite (devuelve vacío). Mejor: hacer `SELECT * FROM <vista> LIMIT 0` y leer `rows.ColumnTypes()`.

**Handler en `handler.go`**:
```go
func (h *TemplateHandler) handleColumnasModulo(w http.ResponseWriter, r *http.Request) {
    modulo := r.URL.Query().Get("modulo")
    if modulo == "" { modulo = "expedientes" }
    cfg, ok := Modulos[modulo]
    if !ok {
        writeJSONError(w, "modulo invalido", http.StatusBadRequest)
        return
    }
    cols, err := h.app.ObtenerColumnasVista(cfg.Vista)
    if err != nil {
        writeJSONError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, cols)
}
```

**Método en `app.go`** (añadir dentro de `App`):
```go
func (a *App) ObtenerColumnasVista(vista string) ([]string, error) {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.db == nil { return nil, fmt.Errorf("no hay BD abierta") }
    rows, err := a.db.Query("SELECT * FROM " + vista + " LIMIT 0")
    if err != nil { return nil, err }
    defer rows.Close()
    return rows.Columns()
}
```

**Ruta en `ServeHTTP`** (añadir al switch):
```go
case p == "/api/columnas-modulo" && r.Method == http.MethodGet:
    h.handleColumnasModulo(w, r)
    return
```

### 2.7 Commits para Feature #1

**Un solo commit** (el user no pidió subdividir):
```bash
git add go.mod go.sum app.go handler.go templates/index.html
git commit -m "$(cat <<'EOF'
feat: exportar Excel XLSX inteligente con filtros y seleccion de columnas

- handler.go: nuevo endpoint /api/exportar-excel que genera .xlsx con
  github.com/xuri/excelize/v2. Soporta: modulo/hoja (9 opciones),
  filtro texto (q), rango fechas (fecha_recibido), columnas
  seleccionadas. Cabecera con estilo teal + freeze panes + autofilter.
- handler.go: endpoint auxiliar /api/columnas-modulo devuelve columnas
  de la vista activa (SELECT * LIMIT 0 + rows.Columns()).
- app.go: metodo ObtenerColumnasVista(vista).
- index.html: boton "Exportar Excel" reemplaza "CSV". Modal con
  selector de modulo, filtro texto, rango fechas, checkboxes de
  columnas (cargados dinamicamente por modulo). Mantiene /api/csv
  por compatibilidad.
- go.mod: anyade github.com/xuri/excelize/v2.

[skip ci]
EOF
)"
```

---

## 3. FEATURE #2 — RUTA PROCESOS: SELECCIÓN DE PROCESOS + AUTO-UPDATE

### 3.1 Bug actual — `RUTA_PROCESOS_DATA` roto en versión Wails

**Hallazgo**: `templates/ruta_procesos.html:16` referencia `window.RUTA_PROCESOS_DATA` (con `legend`, `columns`, `processes`) pero esa variable **NO está definida** en la nueva `templates/index.html` migrada a Wails. Solo existía en el `frontend/index.html` legacy (líneas 1906+). Esto significa que la Ruta Procesos actualmente está rota o muestra tabla vacía.

**Fix previo (debe ir antes o junto con Feature #2)**: definir `window.RUTA_PROCESOS_DATA` con datos reales. Requiere crear tablas en SQLite para:
- `ruta_procesos_leyenda` (id, status_name, hex_color)
- `ruta_procesos_cronograma` (id_proceso, id_expediente, fecha, id_leyenda, nota)
- `ruta_procesos_procesos` (id, descripcion, db_id)

**Decisión de scope**: este fix previo es necesario para que Feature #2 funcione. Se incluye en el mismo commit de Feature #2.

### 3.2 Schema nuevo para Ruta Procesos

**Archivo nuevo**: `data/sql/03_ruta_procesos.sql` — crea las tablas + inserts de la leyenda base + vista agrupada.

```sql
-- Schema Ruta Procesos (Gantt)
PRAGMA foreign_keys = ON;

CREATE TABLE ruta_procesos_leyenda (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    status_name TEXT NOT NULL UNIQUE,
    hex_color   TEXT NOT NULL
);

CREATE TABLE ruta_procesos_cronograma (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    id_proceso   INTEGER NOT NULL,
    id_expediente INTEGER,
    fecha        DATE NOT NULL,
    id_leyenda   INTEGER,
    nota         TEXT,
    CONSTRAINT fk_cron_ley FOREIGN KEY (id_leyenda) REFERENCES ruta_procesos_leyenda(id),
    CONSTRAINT fk_cron_exp FOREIGN KEY (id_expediente) REFERENCES expedientes(id_expediente)
);

CREATE TABLE ruta_procesos_procesos (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    descripcion TEXT NOT NULL,
    db_id       INTEGER,
    activo       INTEGER DEFAULT 1,
    CONSTRAINT fk_proc_exp FOREIGN KEY (db_id) REFERENCES expedientes(id_expediente)
);

-- Leyenda base (colores del Gantt actual del legacy)
INSERT INTO ruta_procesos_leyenda (status_name, hex_color) VALUES
('PENDIENTE', '#FFA500'),
('EN REVISIÓN', '#3B82F6'),
('FIRMADO', '#10B981'),
('DEVUELTO', '#EF4444'),
('SIN NOVEDAD', '#6B7280'),
('OTRO', '#000000');
```

### 3.3 Backend — nueva API para Ruta Procesos

**Endpoints nuevos en `handler.go`**:
- `GET /api/ruta-procesos` (ya existe, se reescribe): devuelve el HTML del Gantt Ya con datos de la BDD.
- `GET /api/ruta-procesos-json`: devuelve JSON con `{legend, columns, processes}` para que el JS del template renderice.
- `POST /api/ruta-procesos-toggle`: activa/desactiva un proceso (BOOL `activo`).
- `GET /api/ruta-procesos-leyenda`: devuelve JSON de la leyenda.
- `POST /api/ruta-procesos-leyenda`: añade/edita una leyenda (para Feature #3, pero el endpoint queda listo).

**Métodos nuevos en `app.go`**:
- `ObtenerRutaProcesosData()` → devuelve struct con `Legend`, `Columns`, `Processes` consultando las tablas nuevas + `vw_reporte_excel_contrataciones` para el `db_id`/`solped`/`estatus`.
- `ToggleRutaProceso(id int, activo bool)` → UPDATE `ruta_procesos_procesos.activo`.
- `ObtenerRutaProcesosLeyenda()` → SELECT * FROM `ruta_procesos_leyenda`.
- `GuardarRutaProcesosLeyenda(statusName, hexColor string)` → INSERT (USAR para Feature #3).

### 3.4 Frontend — modal de selección de procesos

**Acción en `templates/ruta_procesos.html`**:
- Reescribir el template para:
  1. Mostrar lista de procesos con checkboxes ( activo/inactivo ).
  2. Al togglear un checkbox → POST a `/api/ruta-procesos-toggle` → re-render del Gantt vía HTMX.
  3. El Gantt solo renderiza procesos con `activo = 1`.
  4. Auto-update: el Gantt se regenera desde la BDD cada vez que se cambia el estado, sin recargar la página.

**Botón nuevo en `index.html`** para abrir el gestor de procesos:
```html
<button onclick="abrirGestorProcesos()" class="btn btn-secondary">
    <i class="fas fa-cog mr-1"></i> Gestionar Procesos
</button>
```

Esto abre un modal	o panel lateral con la lista de procesos y checkboxes.

### 3.5 Commit para Feature #2

```bash
git add data/sql/03_ruta_procesos.sql app.go handler.go templates/ruta_procesos.html templates/index.html
git commit -m "$(cat <<'EOF'
feat: ruta procesos con seleccion de procesos y auto-update del Gantt

- data/sql/03_ruta_procesos.sql: tablas ruta_procesos_leyenda,
  _cronograma, _procesos + insert leyenda base.
- app.go: metodos ObtenerRutaProcesosData, ToggleRutaProceso,
  ObtenerRutaProcesosLeyenda, GuardarRutaProcesosLeyenda.
- handler.go: nuevos endpoints /api/ruta-procesos-json,
  /api/ruta-procesos-toggle, /api/ruta-procesos-leyenda.
- ruta_procesos.html: reescrito para renderizar desde BDD con
  checkboxes de activo/inactivo por proceso. Gantt auto-update
  via HTMX al togglear.
- index.html: boton "Gestionar Procesos" abre el modal.

[skip ci]
EOF
)"
```

---

## 3b. FEATURE #4 — MODALES SE CIERRAN AL CLICKEAR AFUERA (CON JERARQUÍA)

### 3b.1 Estado actual y comportamiento buscado

**Modales existentes en `templates/index.html`** (todos con backdrop `bg-black/70`, sin handler de cierre al clic afuera):

| ID del modal | Función de cierre | Origen (botón/acción) |
|---|---|---|
| `form-modal` (línea 170) | `cerrarFormulario()` (línea 357) | Botón "+ Nuevo Registro", `hxGetFormulario(id)`, fila de la tabla (editar) |
| `historial-modal` (línea 181) | `cerrarHistorial()` (línea 374) | Botón "Ver Historial" en cada fila |
| `modal-ruta` (línea 195) | `cerrarRuta()` (línea 375) | Botón "Ruta Procesos" |
| `modal-pendientes` (línea 206) | `cerrarPendientes()` (línea 376) | Botón "Pendientes" |
| `modal-recientes` (línea 217) | `cerrarRecientes()` (línea 456) | Botón "Recientes" |
| `modal-frecuentes` (línea 228) | `cerrarFrecuentes()` (línea ~528) | Botón "Frecuentes" |
| `export-modal` (añadido en Feature #1) | `cerrarModalExportar()` (definido en Feature #1) | Botón "Exportar Excel" |

**Comportamiento actual**: los modales solo se cierran con el botón "X" o con sus funciones `_cerrar*` explícitas. Click afuera NO cierra nada.

**Comportamiento deseado (del user)**:
1. Al clickear en el backdrop (overlay oscuro `bg-black/70`), el modal activo se cierra.
2. **Jerarquía nested**: si un modal A abre otro modal B, al clickear afuera de B solo se cierra B, NO A. Solo cuando se cierra B el siguiente clic afuera cierra A.
3. Click dentro del contenido del modal (`modal-content`) NO cierra nada (stopPropagation en elcontenido).

**Casos nested reales identificados en el código**:
- `ruta_procesos.html:47` — `cerrarRuta(); hxGetFormulario(${p.db_id})` cierra el Gantt (modal-ruta) y abre el formulario (form-modal). Actualmente es secuencial (uno se cierra antes de abrir el otro), pero si se cambia a no cerrar explícitamente, el form-modal quedaría encima de modal-ruta. La jerarquía debe respetarse.
- `toggleFrecuente` llamado desde `modal-recientes` puede invocar `abrirFrecuentes()` (línea 477) → abre `modal-frecuentes` mientras `modal-recientes` sigue visible.
- `hxGetFormulario` puede ser llamado desde cualquier modal → abre `form-modal` encima.

### 3b.2 Implementación — gestión de stack de modales

**Decisión**: Llevar un stack (`Array`) de IDs de modales abiertos en orden de apertura. Al clic afuera (en el backdrop), se cierra solo el tope del stack (el más reciente). El backdrop más antiguo permanece visible.

**Alternativas evaluadas**:
- Un solo handler global en `document.body` que cierre el último modal abierto — descartado: requiere saber ALWAYS cuál es el "último".
- Un backdrop invisible gigante que se redimensiona con cada modal nuevo — descartado: complica el CSS.
- Stack explícito — elegido: simple, robusto, sin depender de z-index ni de DOM order.

### 3b.3 Cambios en `templates/index.html`

**Bloque JS nuevo** (añadir al inicio del `<script>` principal, antes de las funciones de modal actuales):

```js
const MODAL_STACK = [];
function pushModal(id) {
    const el = $(id);
    if (!el || !el.classList.contains('hidden')) return;
    el.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    MODAL_STACK.push(id);
    // Listener del backdrop: solo el tope del stack reacciona
    el.addEventListener('click', (e) => {
        if (e.target === el) {
            // Click en el backdrop (no dentro de .modal-content)
            const top = MODAL_STACK[MODAL_STACK.length - 1];
            if (top === id) cerrarModal(id);
        }
    });
}
function cerrarModal(id) {
    const el = $(id);
    if (!el) return;
    el.classList.add('hidden');
    const idx = MODAL_STACK.lastIndexOf(id);
    if (idx !== -1) MODAL_STACK.splice(idx, 1);
    if (MODAL_STACK.length === 0) document.body.style.overflow = '';
}
function topModal() { return MODAL_STACK[MODAL_STACK.length - 1] || null; }
```

**Refactor de funciones existentes** (mantener nombres públicos por compatibilidad, delegar en el stack):

```js
function mostrarFormulario(id) {
    const titulo = $('form-titulo');
    const moduloKey = (window.PAGE_DATA?.modulos && $('active-module-val')) ? $('active-module-val').value : 'expedientes';
    const nombreModulo = window.PAGE_DATA?.modulos?.[moduloKey]?.Nombre || 'Registro';
    titulo.textContent = id ? 'Editar ' + nombreModulo + ' #' + id : 'Nuevo ' + nombreModulo;
    pushModal('form-modal');
}
function cerrarFormulario() { cerrarModal('form-modal'); }

function cerrarHistorial()   { cerrarModal('historial-modal'); }
function cerrarRuta()        { cerrarModal('modal-ruta'); }
function cerrarPendientes()  { cerrarModal('modal-pendientes'); }
function cerrarRecientes()   { cerrarModal('modal-recientes'); }
function cerrarFrecuentes()  { cerrarModal('modal-frecuentes'); }
function cerrarModalExportar() { cerrarModal('export-modal'); }
```

**Cambios en funciones de apertura**:

- `abrirRecientes()` (línea 432): sustituir `modal.classList.remove('hidden'); document.body.style.overflow = 'hidden';` por `pushModal('modal-recientes');` (resto del body se mantiene: construir el HTML interno, etc.).
- `abrirFrecuentes()` (línea 491): idem con `'modal-frecuentes'`.
- `abrirModalExportar()` (definido en Feature #1): sustituir `$('export-modal').classList.remove('hidden'); document.body.style.overflow = 'hidden';` por `pushModal('export-modal');`.
- El modal `historial-modal` se abre desde los `tabla_*.html` con `hx-on::before-request="$('historial-modal').classList.remove('hidden'); document.body.style.overflow = 'hidden'; ...`. Cambiar por `pushModal('historial-modal');` en cada `tabla_*.html` (9 archivos). El `hx-on::before-request` permite invocar JS, así que se reemplaza por `pushModal('historial-modal');` (el `document.body.style.overflow = 'hidden'` ya lo hace `pushModal`).
- El modal `modal-ruta` se abre desde index.html línea 71 con `hx-on::after-request="$('modal-ruta').classList.remove('hidden'); document.body.style.overflow = 'hidden';"`. Cambiar por `hx-on::after-request="pushModal('modal-ruta');"`.
- El modal `modal-pendientes` se abre desde index.html línea 77 con `hx-on::after-request="$('modal-pendientes').classList.remove('hidden'); document.body.style.overflow = 'hidden';"`. Cambiar por `hx-on::after-request="pushModal('modal-pendientes');"`.
- El modal `form-modal` ya se abre vía `mostrarFormulario` (refactoreado arriba) y vía `hxGetFormulario` (línea 671). No hay `hx-on::after-request` que abra el form-modal directamente — `hxGetFormulario` llama a `mostrarFormulario(id)` que ahora usa `pushModal`.

### 3b.4 Caso especial — `ruta_procesos.html:47`

Actualmente: `onclick="event.stopPropagation(); cerrarRuta(); hxGetFormulario(${p.db_id})"` cierra el Gantt y luego abre el formulario.

**Decisión**: dejarlo así. El comportamiento current es correcto para este caso: el user quiere editar un expediente y queda solo el form-modal. SI en el futuro se quiere mantener el Gantt abierto debajo (form-modal encima de modal-ruta), basta con quitar `cerrarRuta();` y reemplazar `hxGetFormulario(...)` por una variante que abra ambos con `pushModal('form-modal')` sin cerrar `'modal-ruta'`.

El stack soporta ambos modos automáticamente: si `cerrarRuta()` se llama, saca `modal-ruta` del stack; si se omite, queda y el clic afuera de `form-modal` solo cerrará form-modal, dejando `modal-ruta` visible para clic afuera en su backdrop.

### 3b.5 Verificación de jerarquía (caso de prueba)

1. Abrir app, abrir BD, abrir `modal-recientes` (stack: `[modal-recientes]`).
2. Desde `modal-recientes`, click en un item que abre `modal-frecuentes` (caso hypothetical: `toggleFrecuente` lo invoca). Stack: `[modal-recientes, modal-frecuentes]`.
3. Click en el backdrop oscuro de `modal-frecuentes` → se cierra solo `modal-frecuentes`. Stack: `[modal-recientes]`. `modal-recientes` sigue visible.
4. Click en el backdrop oscuro de `modal-recientes` → se cierra `modal-recientes`. Stack: `[]`. body.overflow restaurado.

### 3b.6 Commit para Feature #4

**Commit separado** (Feature #4 es independiente de #1 y #2, pero toca `index.html` que también se toca en #1 y #2 → requerirá merge si se hacen en paralelo; mejor做完 #1 antes de #4).

```bash
git add templates/index.html templates/tabla_*.html
git commit -m "$(cat <<'EOF'
feat: modales se cierran al clickear afuera del backdrop con jerarquia

- index.html: anyadido MODAL_STACK array + funciones pushModal/cerrarModal/
  topModal. Las funciones abrir*/cerrar* existentes delegan en el stack.
- El listener de click en cada backdrop solo cierra el modal si es el tope
  del stack (jerarquia: cerrar B no cierra A).
- tabla_*.html (9): hx-on::before-request migrado a pushModal('historial-modal').
- modal-ruta, modal-pendientes, modal-recientes, modal-frecuentes: aperturas
  migradas a pushModal.

[skip ci]
EOF
)"
```

---

## 3c. FEATURES #5, #6, #7 — GERENCIAS PERMITIDAS + BOTONERA AL FONDO + BOTÓN SUMAS

> Las 3 features se incluyen en **un solo commit** porque son pequeñas y/o comparten archivos (`app.go`, `templates/index.html`, `form_*.html`).

### 3c.1 Feature #5 — Gerencias permitidas por módulo en formularios

**Dato del user** — gerencias permitidas por módulo:

| Módulo (key) | Gerencias permitidas (IDs) | Note |
|---|---|---|
| `expedientes` (Control Docs. Presidencia) | 1-10 | SIAHO-A, TÉCNICA, OPERACIONES, SSGG, JURÍDICO, FINANZAS, CONTRATACIÓN, RRHH, ASUNTOS GUBERNAMENTALES, COMISIÓN |
| `requisiciones` | 1, 2, 3, 4, 8, 11 | + PROCURA |
| `memorandums` | 1-11 | 1-10 + PROCURA (11) |
| `recobros` | 1, 2, 3, 4, 8 | |
| `valuaciones` | 1, 2, 3, 4, 8 | |
| `aprobacion_jd` | 1, 2, 3, 4, 7, 8 | + CONTRATACIÓN (7) |
| `certificacion_bdu` | 7 | Solo CONTRATACIÓN |
| `vacaciones` | 1-10, 12 | + CONTROL DE DOCUMENTOS (12) |
| `reposos_medicos` | 1-10, 13 | + ASUNTOS PÚBLICOS (13) |

**Estado actual**: `app.go:34-167` define `Modulos` sin campo de gerencias. `handler.go:444` (`handleCargarExpediente`) pasa todos los catálogos completos al form. El `<select>` de `f-id_gerencia` en cada `form_*.html` renderiza `{{range .Catalogs.gerencia}}` — todas las 13 gerencias aparecen.

**Implementación**:

1. **`app.go`** — Añadir campo `GerenciasIDs []int` al struct `ModuloConfig`. Poblarlo en cada entrada de `Modulos`:

```go
type ModuloConfig struct {
    Nombre         string
    Tabla          string
    Vista          string
    IDColumna      string
    HistorialTabla string
    Columnas       []string
    QueryHistorial string
    GerenciasIDs   []int // IDs de gerencias permitidas en el form de este módulo
}
```

Ejemplos de poblado:
```go
"expedientes": { ..., GerenciasIDs: []int{1,2,3,4,5,6,7,8,9,10} },
"requisiciones": { ..., GerenciasIDs: []int{1,2,3,4,8,11} },
"memorandums": { ..., GerenciasIDs: []int{1,2,3,4,5,6,7,8,9,10,11} },
"recobros": { ..., GerenciasIDs: []int{1,2,3,4,8} },
"valuaciones": { ..., GerenciasIDs: []int{1,2,3,4,8} },
"aprobacion_jd": { ..., GerenciasIDs: []int{1,2,3,4,7,8} },
"certificacion_bdu": { ..., GerenciasIDs: []int{7} },
"vacaciones": { ..., GerenciasIDs: []int{1,2,3,4,5,6,7,8,9,10,12} },
"reposos_medicos": { ..., GerenciasIDs: []int{1,2,3,4,5,6,7,8,9,10,13} },
```

2. **`handler.go`** — En `handleCargarExpediente` (línea 425-463), después de obtener `catalogs`, filtrar la sub-lista `gerencia` por `cfg.GerenciasIDs`. Pasa el catálogo filtrado al template.

```go
cfg := Modulos[modulo]

catalogs, err2 := h.app.ObtenerCatalogos()
// ... error handling ya existente ...

// Filtrar gerencias permitidas para este módulo
if cfg.GerenciasIDs != nil {
    permitidas := map[int]bool{}
    for _, id := range cfg.GerenciasIDs { permitidas[id] = true }
    filtradas := make([]CatalogoItem, 0, len(cfg.GerenciasIDs))
    for _, g := range catalogs["gerencia"] {
        if permitidas[g.ID] { filtradas = append(filtradas, g) }
    }
    catalogs["gerencia"] = filtradas
}
// También filtrar superintendencias: solo las que su id_gerencia está en GerenciasIDs
if cfg.GerenciasIDs != nil {
    permitidas := map[int]bool{}
    for _, id := range cfg.GerenciasIDs { permitidas[id] = true }
    filtradas := []CatalogoItem{}
    for _, s := range catalogs["superintendencia"] {
        if permitidas[s.IDGerencia] { filtradas = append(filtradas, s) }
    }
    catalogs["superintendencia"] = filtradas
}
```

3. **`CatalogoItem`** — verificar que ya tiene campo `IDGerencia` para superintendencias (ver `handler.go` struct `CatalogoItem`). Si no, añadirlo y poblarlo en `ObtenerCatalogos`.

**Resultado**: los 9 `form_<key>.html` no requieren cambios — el `<select>` ya itera `.Catalogs.gerencia`, que ahora viene filtrado.

### 3c.2 Feature #6 — Botonera inferior alineada al fondo real de la ventana

**Estado actual** (`templates/index.html:149-164`): el `<footer>` con `id="modulo-botones"` está dentro del flujo normal, justo después de la tabla. Cuando hay pocos registros, queda " flotando" en el medio de la ventana.

**Comportamiento deseado**: la botonera SIEMPRE visible pegada al fondo de la ventana (sticky footer), incluso con la página vacía.

**Implementación** (CSS en `templates/index.html`, dentro del `<style>` existente en el `<head>`):

```css
#app-root { display: flex; flex-direction: column; min-height: 100vh; }
#vista-tabla { flex: 1 1 auto; }
#modulo-botones { position: sticky; bottom: 0; background: rgb(17, 24, 39); padding: 1rem 1rem 1.5rem; z-index: 10; border-top: 1px solid rgb(55, 65, 81); }
```

**Acción HTML**: envolver el `<div id="app">` (línea 54) con `<div id="app-root">` y mover el `<footer>` con `id="modulo-botones"` (líneas 150-164) para que sea hermano directo de `#app-root` nivel body (no dentro de `#app`). Alternativa más segura: aplicar `position: fixed; bottom: 0; left: 0; right: 0;` al footer y añadir `padding-bottom: 80px` al contenedor padre para evitar que el footer tape contenido.

**Decisión**: usar `position: sticky; bottom: 0` (alternativa A) sobre `position: fixed` para no tener que añadir `padding-bottom` extra y evitar que el footer tape el último registro.

### 3c.3 Feature #7 — Botón "Sumas" (calculadora con 2 decimales)

**Estado actual**: no existe botón de sumas. El botón superior (`index.html:60-78`) tiene: Nuevo Registro, Ruta Procesos, Pendientes, Recientes, Frecuentes, Exportar Excel (añadido en Feature #1), Optimizar BD, Abrir BD.

**Implementación**:

1. **Botón nuevo** en el header de `index.html` (al lado de "Exportar Excel"):
```html
<button onclick="abrirSumas()" class="btn btn-primary" id="btn-sumas">
    <i class="fas fa-calculator mr-1"></i> Sumas
</button>
```

2. **Modal nuevo** (añadir antes de `</body>`, junto con el `export-modal` de Feature #1):
```html
<div id="sumas-modal" class="hidden fixed inset-0 z-50 flex items-start justify-center p-4 overflow-y-auto bg-black/70">
    <div class="relative modal-content w-full max-w-md my-8">
        <div class="sticky top-0 bg-gray-800 z-10 flex items-center justify-between p-4 border-b border-gray-700 rounded-t-xl">
            <h2 class="text-lg font-bold text-teal-400"><i class="fas fa-calculator mr-2"></i>Sumas</h2>
            <button onclick="cerrarSumas()" class="btn-icon text-gray-400 hover:text-white"><i class="fas fa-times"></i></button>
        </div>
        <div class="p-6 space-y-3">
            <div id="sumas-filas" class="space-y-2"></div>
            <button onclick="anyadirFilaSuma()" class="btn btn-secondary w-full"><i class="fas fa-plus mr-1"></i> Añadir número</button>
            <div class="border-t border-gray-700 pt-4 mt-4">
                <div class="flex justify-between items-center text-lg font-bold">
                    <span class="text-gray-300">Resultado:</span>
                    <span id="sumas-resultado" class="text-teal-400">0.00</span>
                </div>
            </div>
            <button onclick="limpiarSumas()" class="btn btn-secondary w-full text-xs"><i class="fas fa-eraser mr-1"></i> Limpiar todo</button>
        </div>
    </div>
</div>
```

3. **JS nuevo** (al final del bloque `<script>` principal):
```js
function abrirSumas() {
    pushModal('sumas-modal');  // usa el stack de Feature #4
    if (document.querySelectorAll('.suma-fila').length === 0) anyadirFilaSuma();
    calcularSumas();
}
function cerrarSumas() { cerrarModal('sumas-modal'); }
function anyadirFilaSuma() {
    const div = document.createElement('div');
    div.className = 'suma-fila flex gap-2';
    div.innerHTML = `
        <input type="number" step="0.01" placeholder="0.00" class="input flex-1 suma-input" oninput="calcularSumas()">
        <button onclick="this.parentElement.remove(); calcularSumas();" class="btn-icon text-red-400 hover:text-red-300" title="Quitar"><i class="fas fa-times"></i></button>
    `;
    $('sumas-filas').appendChild(div);
    div.querySelector('input').focus();
}
function calcularSumas() {
    let total = 0;
    document.querySelectorAll('.suma-input').forEach(inp => {
        // Validar máximo 2 decimales: si tiene más, truncar (no redondear)
        let v = inp.value.trim();
        if (v === '') return;
        // Validar: si tiene más de 2 decimales, mostrar error y truncar
        const parts = v.split('.');
        if (parts.length > 1 && parts[1].length > 2) {
            inp.value = parseFloat(v).toFixed(2);
            v = inp.value;
        }
        const n = parseFloat(v);
        if (!isNaN(n)) total += n;
    });
    $('sumas-resultado').textContent = total.toFixed(2);
}
function limpiarSumas() {
    $('sumas-filas').innerHTML = '';
    anyadirFilaSuma();
    $('sumas-resultado').textContent = '0.00';
}
```

**Restricción de 2 decimales**: input `type="number"` con `step="0.01"`. JS trunca a 2 decimales si el user pega/teclea más. No permite negativos implícitamente (no se prohibe, pero el user dijo "varios números" sin especificar signo).

4. **Depende de Feature #4**: usa `pushModal`/`cerrarModal` del stack. Por eso el commit orden va después del de modales (Commit 2).

### 3c.4 Commit para Features #5, #6, #7

```bash
git add app.go handler.go templates/index.html
git commit -m "$(cat <<'EOF'
feat: gerencias por modulo en forms + botonera al fondo + boton Sumas

- app.go: anyade campo GerenciasIDs []int a ModuloConfig. Poblado por
  modulo con IDs 1-13 (PROCURA=11, CONTROL DE DOCUMENTOS=12,
  ASUNTOS PUBLICOS=13). Cada modulo solo permite sus gerencias.
- handler.go: handleCargarExpediente filtra Catalogs.gerencia y
  Catalogs.superintendencia por cfg.GerenciasIDs antes de pasar al
  form. Los form_*.html no cambian (siguen iterando .Catalogs.gerencia).
- index.html: CSS sticky footer para #modulo-botones (position: sticky
  bottom: 0). Boton "Sumas" en header. Modal de calculadora con inputs
  type=number step=0.01, JS trunca a 2 decimales si el user excede,
  muestra resultado en tiempo real. Usa pushModal del stack (#4).

[skip ci]
EOF
)"
```

---

## 4. FEATURE #3 (POSPUESTO) — LEYENDAS PERSONALIZADAS CON COLORES

> El user dijo: "esa dejala pendiente porque hay una mas importante". **No se ejecuta en este plan.** Queda documentado para un próximo PR.


**Lo que se hará cuando se active**:
- Reutilizar el endpoint `/api/ruta-procesos-leyenda` (POST) ya creado en Feature #2.
- UI: en el modal de gestión de procesos, añadir sección "Leyendas" con:
  - Listado de leyendas existentes (color + nombre).
  - Color picker (`<input type="color">`) + input de nombre + botón "Añadir".
  - Botón "Eliminar" por leyenda.
- El Gantt re-renderiza al cambiar la leyenda (mismo flujo de auto-update).

**No se añade nada al código en este plan para Feature #3**. Solo se deja el endpoint listo y se documenta.

---

## 5. ESTRUCTURA DE COMMITS (resumen)

1. **Commit 1** — Feature #1 (Exportar Excel XLSX inteligente) — archivos: `go.mod`, `go.sum`, `app.go`, `handler.go`, `templates/index.html`.
2. **Commit 2** — Feature #4 (Modales se cierran al clic afuera con jerarquía) — archivos: `templates/index.html`, `templates/tabla_*.html` (9). Se hace antes que #2 para no tocar `index.html` mientras se hace otra feature.
3. **Commit 3** — Features #5+#6+#7 (Gerencias por módulo, botonera al fondo, botón Sumas) — archivos: `app.go`, `handler.go`, `templates/index.html`.
4. **Commit 4** — Feature #2 (Ruta Procesos selección + auto-update) — archivos: `data/sql/03_ruta_procesos.sql`, `app.go`, `handler.go`, `templates/ruta_procesos.html`, `templates/index.html`.
5. **Push final**: `git push origin wails-migration` con `[skip ci]` en los 4 commits.

Feature #3 queda pospuesto y NO genera commit en este plan.

---

## 6. VERIFICACIÓN (smoke test manual tras push, por el user)

1. `make wails-build-linux` → genera `build/bin/GestionExpedientes`.
2. Ejecutar el binario. Abrir `data/expedientes.db`.
3. Click "Exportar Excel" → modal abre con:
   - Selector de módulo (9 opciones).
   - Filtro texto vacío.
   - Rango fechas vacío.
   - Lista de columnas del módulo (`expedientes` por defecto) con checkboxes.
4. Seleccionar "Requisiciones" en el selector → la lista de columnas cambia dinámicamente.
5. Marcar 3 columnas + escribir un filtro + rango fechas → Descargar XLSX → se descarga `Requisición de Materiales_2026-07-13.xlsx` con las filas filtradas y solo las 3 columnas seleccionadas, cabecera teal, freeze panes, autofilter.
6. Abrir el XLSX en Excel/LibreOffice → verificar que las filas coinciden con los filtros.
7. (Feature #2) Click "Gestionar Procesos" → lista de procesos con checkboxes → desactivar uno → el Gantt se re-renderiza sin ese proceso.
8. (Feature #4) Abrir "Recientes" → click afuera (en el backdrop oscuro) → el modal se cierra. body.overflow restaurado.
9. (Feature #4) Abrir "Ruta Procesos" → desde el Gantt, click "Ver/Editar Expediente #N" (abre form-modal encima de modal-ruta) → click afuera de form-modal → solo form-modal se cierra, modal-ruta sigue visible. Luego click afuera de modal-ruta → modal-ruta se cierra.
10. (Feature #5) Abrir "Nuevo Registro" con `expedientes` activo → el `<select>` de Gerencia muestra solo las 10 gerencias 1-10 (no aparecen 11 PROCURA, 12 CONTROL DE DOCUMENTOS, 13 ASUNTOS PÚBLICOS).
11. (Feature #5) Cambiar a `certificacion_bdu` → "Nuevo Registro" → el select de Gerencia muestra **solo** "CONTRATACIÓN" (ID 7).
12. (Feature #5) Cambiar a `vacaciones` → el select muestra 1-10 + 12 (CONTROL DE DOCUMENTOS), pero NO 11 (PROCURA) ni 13 (ASUNTOS PÚBLICOS).
13. (Feature #6) Con pocos registros (o vacío), la botonera inferior queda pegada al fondo de la ventana (sticky). No flota en el medio.
14. (Feature #7) Click "Sumas" → modal abre con un input vacío. Escribir "100.50" → añadir otra fila, escribir "200.25" → resultado muestra "300.75". Pegar "1.234" → se trunca a "1.23", resultado recalculado. Click "Añadir" → nueva fila en blanco, resultado no cambia.

---

## 7. NOTAS Y RIESGOS

- **`github.com/xuri/excelize/v2`**:.dependencia pura Go sin CGO, añade ~5 MB al binario. Aceptable.
- **`rows.Columns()` en vistas SQLite**: confirmado que funciona con `SELECT * FROM <vista> LIMIT 0` (devuelve nombres de columna de la vista).
- **`RUTA_PROCESOS_DATA` roto**: el bug ya existía en la rama antes de este plan. Feature #2 lo resuelve como efecto colateral.
- **BD**: `data/expedientes.db` ya tiene schema 01+02. Feature #2 añade schema 03 (aplicar con `sqlite3 data/expedientes.db < data/sql/03_ruta_procesos.sql` durante el desarrollo, NO se commitea el .db).
- **Anti-alucinación**: antes de cada edit, verificar con Read/Grep que el anchor existe. NO editar `frontend/wailsjs/*` a mano. NO tocar `src/`, `src-tauri/`, `main.js`, `package.json` (legacy).

---

## 8. ANTI-ALUCINACIÓN (verificación cruzada)

### 8.1 Archivos que SÍ existen y se editan

- `/home/user/Documentos/proyecto/baseaccess/app.go` (709 líneas) — añadir `ObtenerColumnasVista`, métodos de Ruta Procesos.
- `/home/user/Documentos/proyecto/baseaccess/handler.go` (736 líneas) — añadir `handleExportarExcel`, `handleColumnasModulo`, handlers Ruta Procesos.
- `/home/user/Documentos/proyecto/baseaccess/templates/index.html` (689 líneas) — reemplazar botón CSV, añadir modal exportar, botón gestionar procesos, añadir `MODAL_STACK` + `pushModal`/`cerrarModal` y refactor de `abrirRecientes`/`abrirFrecuentes`/`mostrarFormulario`/`cerrar*`.
- `/home/user/Documentos/proyecto/baseaccess/templates/tabla_*.html` (9: `tabla_expedientes`, `tabla_requisiciones`, `tabla_memorandums`, `tabla_recobros`, `tabla_valuaciones`, `tabla_aprobacion_jd`, `tabla_certificacion_bdu`, `tabla_vacaciones`, `tabla_reposos_medicos`) — migrar `hx-on::before-request` a `pushModal('historial-modal')`.
- `/home/user/Documentos/proyecto/baseaccess/templates/ruta_procesos.html` (180 líneas) — reescribir para datos desde BDD.
- `/home/user/Documentos/proyecto/baseaccess/go.mod` (39 líneas) — añadir `excelize/v2`.

### 8.2 Archivos NUEVOS

- `/home/user/Documentos/proyecto/baseaccess/data/sql/03_ruta_procesos.sql` — schema Ruta Procesos.

### 8.3 Archivos que NO existen o NO se tocan

- `frontend/index.html`, `src/index.html` — legacy, NO se tocan.
- `frontend/wailsjs/*` — se regeneran solos con `wails dev`/`wails build`.
- `data/expedientes.db` — gitignored, NO se commitea.
- `data/sql/01_master_*.sql`, `data/sql/02_modulos_*.sql` — NO se modifican.
- `main.go`, `wails.json` — NO se tocan.

### 8.4 Verificaciones post-edit

1. `go vet ./...` — cero warnings.
2. `go build ./...` — compila sin errores.
3. `make wails-build-linux` — genera binario correctamente.
4. `git diff --stat` — confirma archivos tocados son solo los listados por commit.
5. Aplicar `data/sql/03_ruta_procesos.sql` a `data/expedientes.db` antes de probar Feature #2.

---

## 9. RESUMEN EJECUTIVO

| Fase | Archivos tocados | Líneas aprox. | Commit |
|---|---|---|---|
| Feature #1 (Exportar Excel) | `go.mod`, `go.sum`, `handler.go`, `app.go`, `templates/index.html` | ~250 añadidas | 1 |
| Feature #4 (Modales click-afuera) | `templates/index.html`, `templates/tabla_*.html` (9) | ~80 añadidas | 1 |
| Features #5+#6+#7 (Gerencias + footer + Sumas) | `app.go`, `handler.go`, `templates/index.html` | ~120 añadidas | 1 |
| Feature #2 (Ruta Procesos) | `data/sql/03_ruta_procesos.sql`, `handler.go`, `app.go`, `templates/ruta_procesos.html`, `templates/index.html` | ~300 añadidas | 1 |
| Feature #3 (Leyendas custom) | — | — | POSPUESTO |

**Total**: 4 commits, ~750 líneas. 1 push. Exportar Excel XLSX con filtros. Modales se cierran al click afuera con jerarquía. Gerencias filtradas por módulo en forms. Botonera inferior sticky al fondo. Botón Sumas con 2 decimales. Ruta Procesos con selección y auto-update. Leyendas custom pospuestas.
