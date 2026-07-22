# Análisis Ruta Procesos — Migración v1 → v2

## 1. Función Principal: `ObtenerRutaProcesosData` (app.go)

### 1.1 Estructura actual (v1)

```go
func (a *App) ObtenerRutaProcesosData(idHoja int, offsetWeeks int) (*RutaProcesosGanttData, error)
```

**Pseudocódigo v1:**
```
1. Bloquear lectura (RLock)
2. Verificar BD abierta
3. Cargar leyendas globales:
   SELECT id, status_name, hex_color, orden FROM ruta_procesos_leyenda
4. Cargar hojas:
   SELECT id, nombre, fecha_inicio, fecha_fin FROM ruta_procesos_hojas
5. Determinar currentHoja (primera o la que coincide con idHoja)
6. Si currentHoja existe:
   a. Construir columnas del Gantt (buildGanttColumns)
      - Calcula días hábiles entre fecha_inicio y fecha_fin
      - Genera semanas automáticamente (SEMANA 1, SEMANA 2...)
   b. Cargar procesos de la hoja desde ruta_procesos_procesos
      - JOIN con expedientes para traer solped, estatus, receptor, etc.
      - Campos: id, modulo, descripcion, db_id, activo, solped, estatus_detalle, receptor
   c. Para cada proceso, auto-popular timeline:
      - Si fecha_recibido → PENDIENTE
      - Si fecha_devuelto → DEVUELTO
      - Si fecha_firma_contrato → FIRMADO
   d. Cargar cronograma manual de ruta_procesos_cronograma
      - LEFT JOIN con leyenda para obtener status_name y hex_color
      - Timeline[fecha] = []map[string]string{entry}
7. Retornar RutaProcesosGanttData {Legend, Columns, Processes, Hojas, CurrentHoja, OffsetWeeks}
```

**Datos que retorna v1:**
```json
{
  "legend": [ {id, status_name, hex_color, orden} ],
  "columns": [ {day_name, week_label, date_str} ],
  "processes": [ {id, modulo, descripcion, db_id, activo, solped, estatus_detalle, receptor, timeline} ],
  "hojas": [ {id, nombre, fecha_inicio, fecha_fin} ],
  "current_hoja": {id, nombre, fecha_inicio, fecha_fin},
  "offset_weeks": 0
}
```

---

### 1.2 Nueva estructura (v2)

```go
func (a *App) ObtenerRutaProcesosData(idHoja int, idJunta int) (*RutaProcesosGanttData, error)
```

**Pseudocódigo v2:**
```
1. Bloquear lectura (RLock)
2. Verificar BD abierta
3. Cargar hojas:
   SELECT id, nombre FROM ruta_procesos_hoja
4. Determinar currentHoja
5. Cargar juntas de la hoja:
   SELECT id, id_hoja, numero, consecutiva, fecha FROM ruta_procesos_junta WHERE id_hoja = ?
6. Determinar currentJunta (primera o la que coincide con idJunta)
7. Si currentJunta existe:
   a. Cargar semanas de la junta:
      SELECT id, numero, fecha_inicio, fecha_fin FROM ruta_procesos_junta_semana WHERE id_junta = ? ORDER BY numero
   b. Construir columnas del Gantt desde las semanas (no más date range)
      - Cada semana tiene 5 columnas (L M X J V)
      - Construir "desde" y "al" de cada semana (fecha_inicio, fecha_fin)
   c. Cargar procesos de la junta:
      SELECT id, numero, proceso FROM ruta_procesos_junta_proceso WHERE id_junta = ? ORDER BY numero
   d. Para cada proceso, cargar cronograma manual:
      SELECT c.id, c.fecha, c.nota, l.nombre, l.color
      FROM ruta_procesos_cronograma c
      LEFT JOIN ruta_procesos_leyenda l ON c.id_leyenda = l.id
      WHERE c.id_junta_proceso = ?
   e. Timeline[fecha] = []RutaProcesosCronogramaEntry{...}
   f. Cargar leyendas activas de esta junta:
      - Leyendas globales (ambito='global')
      - Leyendas de esta hoja (ambito='hoja' AND id_hoja=?)
      - Leyendas de esta junta (ambito='junta' Y existen en ruta_procesos_junta_leyenda WHERE id_junta=?)
   g. Ordenar leyendas según ruta_procesos_junta_leyenda.orden
8. Retornar RutaProcesosGanttData {
   Hojas, CurrentHoja, Juntas, CurrentJunta,
   Semanas, Procesos, Legend, JuntaLegend
}
```

**Datos que retornará v2:**
```json
{
  "hojas": [ {id, nombre} ],
  "current_hoja": {id, nombre},
  "juntas": [ {id, id_hoja, numero, consecutiva, fecha} ],
  "current_junta": {id, id_hoja, numero, consecutiva, fecha},
  "semanas": [ {id, id_junta, numero, fecha_inicio, fecha_fin} ],
  "procesos": [ {id, id_junta, numero, proceso, timeline} ],
  "legend": [ {id, nombre, color, ambito, id_hoja, bloqueado} ],
  "junta_legend": [ {id, id_junta, id_leyenda, orden} ]
}
```

**Cambios clave:**
- Se elimina `offsetWeeks` (ya no hay scroll lateral de semanas, todas las semanas están en el Gantt)
- Se elimina `Columns` generadas por fecha (ahora las semanas son entidades DB)
- Los procesos ya no traen `modulo`, `db_id`, `solped`, `estatus` (desligados de expedientes)
- `CurrentJunta` es nuevo
- `Semanas` es nuevo

---

## 2. `buildGanttColumns` (app.go)

### 2.1 Actual v1

```go
func buildGanttColumns(inicioStr string, finStr string, offsetWeeks int) []map[string]string
```

- Recibe `fecha_inicio` y `fecha_fin` de la hoja
- Recibe `offsetWeeks` (desplazamiento de scroll)
- Genera columnas de días hábiles (L-V) con week labels
- Retorna: `[{day_name, week_label, date_str}, ...]`

### 2.2 Nueva v2

**Eliminada.** El Gantt ya no genera columnas desde un rango de fechas. En su lugar:
- Carga semanas desde `ruta_procesos_junta_semana`
- Cada semana tiene `numero`, `fecha_inicio`, `fecha_fin`
- Las columnas se construyen en el frontend por semana (L M X J V)

---

## 3. Funciones CRUD por entidad

### 3.1 Leyendas (ya existen, se modifica)

| Función v1 | Función v2 | Cambio |
|------------|------------|--------|
| `CrearRutaProcesosLeyenda(nombre, color)` | `CrearRutaProcesosLeyenda(nombre, color, ambito, idHoja, idJunta)` | Agrega `ambito` y opcionalmente `idHoja`. Si `ambito='global'` → inserta en todas las juntas. Si `ambito='hoja'` → inserta en todas las juntas de esa hoja. Si `ambito='junta'` → inserta solo en esa junta. |
| `ActualizarRutaProcesosLeyenda(id, nombre, color)` | `ActualizarRutaProcesosLeyenda(id, nombre, color)` | Igual. |
| `ReordenarRutaProcesosLeyenda(id, direction)` | `ReordenarRutaProcesosLeyenda(idJunta, idLeyenda, direction)` | Ahora actúa sobre `ruta_procesos_junta_leyenda.orden`. |

### 3.2 Hojas (ya existen, se modifica)

| Función v1 | Función v2 | Cambio |
|------------|------------|--------|
| `CrearRutaProcesosHoja(nombre, inicioStr, finStr)` | `CrearRutaProcesosHoja(nombre)` | Elimina `inicioStr` y `finStr`. Solo `nombre` (mes). |
| `EliminarRutaProcesosHoja(id)` | `EliminarRutaProcesosHoja(id)` | Igual. CASCADE borra juntas, semanas, procesos, cronograma. |

### 3.3 Nuevas entidades

| Entidad | Funciones nuevas |
|---------|------------------|
| **Junta** | `CrearRutaProcesosJunta(idHoja, numero, consecutiva, fecha)`, `ActualizarRutaProcesosJunta(id, numero, consecutiva, fecha)`, `EliminarRutaProcesosJunta(id)` |
| **Semana** | `AgregarRutaProcesosSemana(idJunta, numero, fechaInicio, fechaFin)`, `EliminarRutaProcesosSemanas(idJunta, numeros[])` (renumera al borrar) |
| **Proceso** | `AgregarRutaProcesosProceso(idJunta, numero, proceso)`, `EliminarRutaProcesosProceso(id)`, `ReordenarRutaProcesosProceso(idJunta, idProceso, direction)` |
| **Cronograma** | `GuardarCronogramaDia(idProceso, fecha, idLeyenda, nota)`, `EliminarCronogramaDia(id)` → ya existen, solo cambia `id_proceso` a `id_junta_proceso` |
| **Leyenda-bloqueo** | `ToggleBloquearLeyenda(id)` |

### 3.4 Funciones v1 que se eliminan

| Función | Motivo |
|---------|--------|
| `ToggleRutaProceso(id, activo)` | Ya no hay campo `activo` en procesos |
| `AgregarRutaProceso(idHoja, modulo, descripcion, dbID)` | Reemplazada por `AgregarRutaProcesosProceso(idJunta, numero, proceso)` |
| `EliminarRutaProceso(id)` | Reemplazada por `EliminarRutaProcesosProceso(id)` |
| `ObtenerExpedientesDisponiblesRuta()` | Ya no se vinculan a módulos |
| `ObtenerRegistrosDisponiblesRuta(modulo)` | Ya no se vinculan a módulos |
| `EliminarRutaCronogramaCelda(idProceso, fechaStr)` | Reemplazada por `EliminarCronogramaDia(id)` |

---

## 4. Frontend JS (`ruta_procesos.html`)

### 4.1 Estructura v1 (JS dentro del template)

```javascript
(function() {
    const data = {{jsonEncode .}};
    const { legend, columns, processes, offset_weeks, current_hoja } = data;
    
    // Funciones globales:
    window.toggleModal(id)
    window.cambiarHoja()
    window.eliminarHojaActual()
    window.crearHoja()
    window.editarLeyenda(id, nombre, color)
    window.guardarEditarLeyenda()
    window.crearLeyenda()
    window.moverLeyenda(id, direction)
    window.cargarRegistrosModulo()
    window.toggleFormProceso()
    window.agregarProceso()
    window.toggleProceso(id, checked)
    window.eliminarProceso(id)
    
    // Render:
    function renderAll() { ... }
    function renderGantt() { ... }
    
    // Cronograma:
    window.abrirEditarCronograma(procId, fecha)
    window.guardarCronoDia()
    window.eliminarCronoEntry(id)
})();
```

**Pseudocódigo renderGantt v1:**
```
1. Obtener semanas vistas de las columnas (c.week_label)
2. Generar encabezados de semana (row con colspan)
3. Generar labels de semana (row con colspan)
4. Generar fila de días (L M X J V)
5. Para cada proceso activo:
   a. Para cada columna:
      - Buscar entradas en timeline[fecha]
      - Si hay: renderizar celdas coloreadas apiladas
      - Si no: celda vacía clickeable
6. Renderizar leyendas con ▲▼
```

---

### 4.2 Nueva estructura v2 (JS dentro del template)

```javascript
(function() {
    const data = {{jsonEncode .}};
    const {
        hojas, current_hoja,
        juntas, current_junta,
        semanas, procesos, legend, junta_legend
    } = data;
    
    // Funciones globales:
    // -- Hoja --
    window.toggleModal(id)
    window.cambiarHoja()
    window.eliminarHojaActual()
    window.crearHoja()
    
    // -- Junta --
    window.editarJunta(id)           // Carga junta en formulario
    window.guardarJunta()            // POST /api/ruta-procesos-junta-actualizar
    window.crearJunta()              // POST /api/ruta-procesos-junta-crear
    window.eliminarJunta(id)         // POST /api/ruta-procesos-junta-eliminar
    
    // -- Semanas (Gantt header) --
    window.abrirAgregarSemana()      // Modal: desde lunes hasta viernes
    window.guardarSemana()           // POST /api/ruta-procesos-semana-agregar
    window.abrirEliminarSemanas()    // Modal con checkboxes
    window.eliminarSemanas()         // POST /api/ruta-procesos-semana-eliminar
    
    // -- Procesos --
    window.agregarProceso()          // POST /api/ruta-procesos-proceso-agregar
    window.eliminarProceso(id)
    window.moverProceso(id, direction)
    
    // -- Leyendas --
    window.editarLeyenda(id, nombre, color)
    window.guardarEditarLeyenda()
    window.crearLeyenda()
    window.moverLeyenda(id, direction)
    window.toggleBloquearLeyenda(id) // POST /api/ruta-procesos-leyenda-bloquear
    
    // -- Cronograma --
    window.abrirEditarCronograma(procId, fecha)
    window.guardarCronoDia()
    window.eliminarCronoEntry(id)
    
    // Render:
    function renderGantt() { ... }
    
    // Inicializar:
    renderAll();
})();
```

**Pseudocódigo renderGantt v2:**
```
1. Para cada semana (semanas de la junta):
   a. Fila "desde":  semana.fecha_inicio
   b. Fila "al":     semana.fecha_fin
   c. Fila "SEMANA N": semana.numero  (con [+] y [🗑] al final)
   d. Fila de días: L M X J V (5 celdas por semana)
2. Para cada proceso:
   a. Columnas N° y Proceso fijas
   b. Para cada semana, para cada día (L-V):
      - Buscar entradas en timeline[fecha]
      - Renderizar celdas coloreadas
3. Renderizar leyendas (específicas de la junta)
```

---

### 4.3 Principales cambios en el frontend

| Aspecto v1 | Aspecto v2 |
|-------------|-----------|
| Una hoja con múltiples procesos | Una hoja con múltiples juntas |
| Procesos de módulos (expedientes, etc.) | Procesos independientes (solo nombre) |
| Semanas generadas automáticamente del rango de la hoja | Semanas entidades DB agregadas/renumeradas manualmente |
| Leyendas globales únicas | Leyendas con 3 ámbitos (junta/hoja/global) |
| Columnas: solo días (L M X J V) | Columnas: N°, Proceso, días de cada semana |
| Fila "desde/al" global | Fila "desde/al" por semana |
| Botón "agregar proceso" carga de módulos | Botón "agregar proceso" crea un nombre libre |
| Botón "eliminar proceso" | Igual + botón "eliminar junta" |
| Scroll lateral de semanas con offset | Todas las semanas visibles (scroll vertical de juntas) |

---

## 5. Resumen de tablas eliminadas vs nuevas

### Tablas v1 (se mantienen en legacy)
- `ruta_procesos_leyenda` → **reemplazada** (nueva con ambito, id_hoja, bloqueado)
- `ruta_procesos_hojas` → **reemplazada** (ahora `ruta_procesos_hoja`, sin fechas)
- `ruta_procesos_procesos` → **eliminada** (reemplazada por `ruta_procesos_junta_proceso`)
- `ruta_procesos_cronograma` → **modificada** (nuevos FKs, sin id_expediente)

### Tablas v2 (nuevas)
- `ruta_procesos_hoja`
- `ruta_procesos_junta`
- `ruta_procesos_junta_semana`
- `ruta_procesos_junta_proceso`
- `ruta_procesos_cronograma` (modificada)
- `ruta_procesos_leyenda` (modificada)
- `ruta_procesos_junta_leyenda` (nueva)

---

## 6. Plan de implementación paso a paso

1. ✅ **SQL** — Crear nuevo `03_ruta_procesos.sql` con 7 tablas + leyendas base globales
2. **app.go** — Reemplazar `initRutaProcesosSchema`
3. **app.go** — Reemplazar structs v1 por v2
4. **app.go** — Reescribir `ObtenerRutaProcesosData`
5. **app.go** — Eliminar `buildGanttColumns`, `parseDateFlex`
6. **app.go** — Actualizar funciones CRUD de hojas y leyendas
7. **app.go** — Crear funciones CRUD de juntas, semanas, procesos
8. **app.go** — Eliminar funciones v1 obsoletas
9. **handler.go** — Actualizar handlers existentes
10. **handler.go** — Crear nuevos handlers
11. **handler.go** — Eliminar handlers v1 obsoletos
12. **ruta_procesos.html** — Reescribir template completo
13. **styles.css** — Ajustar estilos del Gantt (columnas N° y Proceso fijas)
14. Build y pruebas
