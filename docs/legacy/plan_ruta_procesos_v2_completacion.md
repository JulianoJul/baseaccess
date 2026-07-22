# Ruta Procesos v2 — Plan de Completación (Brechas v1 → v2)

> Fecha: 2026-07-22
> Contexto: La migración v1→v2 dejó el backend 100% completo pero el frontend simplificado.
> Este documento lista exactamente qué falta y cómo implementarlo.

---

## Estado Actual del Backend (100% completo)

### SQL Schema v2
- Archivo: `data/sql/03_ruta_procesos.sql`
- 7 tablas: `ruta_procesos_hoja`, `ruta_procesos_junta`, `ruta_procesos_junta_semana`, `ruta_procesos_junta_proceso`, `ruta_procesos_cronograma`, `ruta_procesos_leyenda`, `ruta_procesos_junta_leyenda`
- 10 leyendas base globales insertadas

### app.go — Structs v2
```go
type RutaProcesosLegend struct      // id, nombre, color, ambito, id_hoja, bloqueado
type RutaProcesosJunta struct       // id, id_hoja, numero, consecutiva, fecha
type RutaProcesosJuntaSemana struct  // id, id_junta, numero, fecha_inicio, fecha_fin, dias[]
type RutaProcesosJuntaProceso struct // id, id_junta, numero, proceso, timeline map[string][]Entry
type RutaProcesosCronogramaEntry struct // id, id_junta_proceso, fecha, id_leyenda, nota, status_name, hex_color
type RutaProcesosJuntaLeyenda struct // id, id_junta, id_leyenda, orden
type RutaProcesosHoja struct         // id, nombre
type RutaProcesosGanttData struct    // hojas, current_hoja, juntas, current_junta, semanas, procesos, legend, junta_legend
```

### app.go — Funciones CRUD v2 (todas implementadas)
| Función | Línea aprox. |
|---|---|
| `initRutaProcesosSchema()` | 438 |
| `ObtenerRutaProcesosData(idHoja, idJunta)` | 1059 |
| `calcDiasSemana(inicio)` | 1274 |
| `CrearRutaProcesosHoja(nombre)` | 1293 |
| `EliminarRutaProcesosHoja(id)` | 1309 |
| `CrearRutaProcesosJunta(idHoja, numero, consecutiva, fecha)` | 1324 |
| `ActualizarRutaProcesosJunta(id, numero, consecutiva, fecha)` | 1354 |
| `EliminarRutaProcesosJunta(id)` | 1367 |
| `AgregarRutaProcesosSemana(idJunta, numero, fechaInicio, fechaFin)` | 1382 |
| `EliminarRutaProcesosSemanas(idJunta, numeros[])` | 1395 |
| `AgregarRutaProcesosProceso(idJunta, numero, proceso)` | 1432 |
| `EliminarRutaProcesosProceso(id)` | 1445 |
| `ReordenarRutaProcesosProceso(idJunta, idProceso, direction)` | 1458 |
| `CrearRutaProcesosLeyenda(nombre, color, ambito, idHoja)` | 1482 |
| `ActualizarRutaProcesosLeyenda(id, nombre, color)` | 1534 |
| `EliminarRutaProcesosLeyenda(id)` | 1547 |
| `ReordenarRutaProcesosLeyenda(idJunta, idLeyenda, direction)` | 1560 |
| `ToggleBloquearRutaProcesosLeyenda(id)` | 1582 |
| `GuardarCronogramaDia(idProceso, fecha, idLeyenda, nota)` | 1597 |
| `EliminarCronogramaDia(id)` | 1613 |

### handler.go — Endpoints v2 (todos registrados, líneas 472-521)
| Endpoint | Handler |
|---|---|
| `GET /api/ruta-procesos` | `handleRutaProcesos` |
| `POST /api/ruta-procesos-cronograma-guardar` | `handleGuardarCronogramaDia` |
| `POST /api/ruta-procesos-cronograma-eliminar` | `handleEliminarCronogramaDia` |
| `POST /api/ruta-procesos-hoja-crear` | `handleCrearRutaProcesoHoja` |
| `POST /api/ruta-procesos-hoja-eliminar` | `handleEliminarRutaProcesoHoja` |
| `POST /api/ruta-procesos-junta-crear` | `handleCrearJunta` |
| `POST /api/ruta-procesos-junta-actualizar` | `handleActualizarJunta` |
| `POST /api/ruta-procesos-junta-eliminar` | `handleEliminarJunta` |
| `POST /api/ruta-procesos-semana-agregar` | `handleAgregarSemana` |
| `POST /api/ruta-procesos-semana-eliminar` | `handleEliminarSemanas` |
| `POST /api/ruta-procesos-proceso-agregar` | `handleAgregarProceso` |
| `POST /api/ruta-procesos-proceso-eliminar` | `handleEliminarProceso` |
| `POST /api/ruta-procesos-leyenda-crear` | `handleCrearLeyenda` |
| `POST /api/ruta-procesos-leyenda-actualizar` | `handleActualizarLeyenda` |
| `POST /api/ruta-procesos-leyenda-eliminar` | `handleEliminarLeyenda` |
| `POST /api/ruta-procesos-leyenda-reordenar` | `handleReordenarLeyenda` |
| `POST /api/ruta-procesos-leyenda-bloquear` | `handleToggleBloquearLeyenda` |

---

## Brechas del Frontend (lo que falta)

### Brecha 1 (CRÍTICA): Celdas coloreadas del Gantt

**Problema:** En v2, TODAS las celdas de día del Gantt se renderizan como vacías (`·`) sin importar si tienen cronograma guardado. El backend carga correctamente el `timeline` en cada proceso pero el template Go no lo utiliza para colorear.

**Código actual v2 (templates/new/ruta_procesos.html, líneas 105-109):**
```html
<!-- Siempre renderiza celda vacía -->
<td class="gantt-col-day gantt-cell-empty cursor-pointer"
    onclick="abrirEditarCronograma({{$p.ID}}, '{{index $s.Dias 0}}')">
    <span class="text-gray-600">·</span>
</td>
```

**Lo que hacía v1 (ruta_procesos_v1/ruta_procesos.html, líneas 465-484, JS dinámico):**
```javascript
if (entries && entries.length > 0) {
    html += '<td class="gantt-cell-active" onclick="abrirEditarCronograma(...)"><div class="gantt-cell-entries">';
    entries.forEach(function(e) {
        html += '<div class="gantt-cell-entry" style="background-color:' + e.hex_color + '" title="' + e.status_name + '"></div>';
    });
    html += '</div></td>';
} else {
    html += '<td class="gantt-cell-empty" onclick="abrirEditarCronograma(...)"><span>·</span></td>';
}
```

**Solución recomendada (Post-render JS):** Añadir un bloque `<script>` al final que recorra `data.procesos[].timeline` y coloree las celdas correspondientes. Para hacerlo, asignar un `data-*` attribute a cada celda:

1. Modificar cada `<td>` de día para incluir `data-proc="{{$p.ID}}" data-fecha="{{index $s.Dias N}}"`.
2. Añadir al `<script>` existente una función `colorearCeldas()` que:
   - Itere sobre `procesos`
   - Para cada proceso con `timeline`, busque las celdas por `data-proc` + `data-fecha`
   - Reemplace el contenido `·` por barras coloreadas `<div class="gantt-cell-entry" style="background-color:..."></div>`
   - Cambie la clase de `gantt-cell-empty` a `gantt-cell-active`
3. Llamar a `colorearCeldas()` al final del script.

---

### Brecha 2: Estilos CSS del Gantt

**Problema:** El CSS actual solo tiene `.gantt-cell-entry { margin: 0 1px; }` (línea 305-308 del template). Faltan todas las clases necesarias para que las celdas coloreadas se vean bien.

**Clases CSS a portar desde `ruta_procesos_v1/styles.css`:**

```css
/* Celda con actividad */
.gantt-cell-active {
    cursor: pointer;
    padding: 2px 0;
}

.gantt-cell-empty {
    cursor: pointer;
}
.gantt-cell-empty:hover {
    background-color: rgba(55, 65, 81, 0.4);
}

/* Contenedor de entradas apiladas dentro de una celda */
.gantt-cell-entries {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-height: 12px;
}

/* Barra coloreada individual */
.gantt-cell-entry {
    height: 4px;
    border-radius: 1px;
    margin: 0 1px;
    transition: opacity 0.15s;
}

.gantt-cell-active:hover {
    background-color: rgba(55, 65, 81, 0.3);
}

/* Tooltip en hover (opcional, ver Brecha 2b) */
.gantt-tooltip {
    display: none;
    position: absolute;
    bottom: calc(100% + 4px);
    left: 50%;
    transform: translateX(-50%);
    background: #1f2937;
    border: 1px solid #374151;
    border-radius: 8px;
    padding: 6px 10px;
    min-width: 140px;
    z-index: 50;
    pointer-events: none;
}
.gantt-cell-active:hover .gantt-tooltip,
.gantt-cell-empty:hover .gantt-tooltip {
    display: block;
}

/* Leyenda grid */
.gantt-legend-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
}
@media (min-width: 768px) {
    .gantt-legend-grid {
        grid-template-columns: repeat(3, 1fr);
    }
}

.gantt-legend-item {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    background: rgba(17, 24, 39, 0.6);
    border: 1px solid rgba(55, 65, 81, 0.6);
    border-radius: 8px;
    font-size: 0.75rem;
}

.gantt-legend-circle {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    flex-shrink: 0;
}

.gantt-legend-label {
    flex: 1;
    color: #d1d5db;
    cursor: pointer;
}

/* Flechas reorden leyenda */
.gantt-legend-arrows {
    display: flex;
    flex-direction: column;
    gap: 0;
}
.legend-arrow {
    background: none;
    border: none;
    color: #6b7280;
    font-size: 8px;
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
}
.legend-arrow:hover {
    color: #2dd4bf;
}
```

**Dónde poner:** En el bloque `<style>` existente en `ruta_procesos.html` (líneas 304-308) o en `frontend/vendor/styles.css`.

---

### Brecha 3: Botones ▲▼ para reordenar leyendas + funciones JS

**Problema:** La UI de leyendas (líneas 149-162 de ruta_procesos.html) muestra editar/eliminar pero no tiene flechas para reordenar. El endpoint backend existe (`/api/ruta-procesos-leyenda-reordenar`).

**Cambios en el template HTML** (dentro del loop `{{range $jl := $root.JuntaLegend}}`):

Agregar antes del botón editar:
```html
<div class="gantt-legend-arrows">
    <button onclick="moverLeyenda({{$j.ID}}, {{$l.ID}}, -1)" class="legend-arrow" title="Subir">&#9650;</button>
    <button onclick="moverLeyenda({{$j.ID}}, {{$l.ID}}, 1)" class="legend-arrow" title="Bajar">&#9660;</button>
</div>
```

**Función JS a agregar:**
```javascript
window.moverLeyenda = function(idJunta, idLeyenda, direction) {
    jsonPost('/api/ruta-procesos-leyenda-reordenar',
        'id_junta=' + idJunta + '&id_leyenda=' + idLeyenda + '&direction=' + direction
    ).then(function(res) {
        if (res.success) reload();
        else alert(res.error || 'Error');
    });
};
```

---

### Brecha 4: Toggle bloquear leyenda (función JS)

**Problema:** El backend tiene `ToggleBloquearRutaProcesosLeyenda` y el endpoint `/api/ruta-procesos-leyenda-bloquear`, pero no hay función JS que lo invoque. El ícono de candado se muestra pero no es interactivo.

**Función JS a agregar:**
```javascript
window.toggleBloquearLeyenda = function(id) {
    jsonPost('/api/ruta-procesos-leyenda-bloquear', 'id=' + id).then(function(res) {
        if (res.success) reload();
        else alert(res.error || 'Error');
    });
};
```

**Cambio en template:** Hacer el candado clickeable:
```html
{{if $l.Bloqueado}}
<button onclick="toggleBloquearLeyenda({{$l.ID}})" class="btn-icon text-yellow-500 text-[10px]" title="Desbloquear">
    <i class="fas fa-lock"></i>
</button>
{{else}}
<button onclick="toggleBloquearLeyenda({{$l.ID}})" class="btn-icon text-gray-500 hover:text-yellow-400 text-[10px]" title="Bloquear">
    <i class="fas fa-unlock"></i>
</button>
<button onclick="eliminarLeyenda({{$l.ID}})" class="btn-icon text-gray-500 hover:text-red-400 text-[10px]" title="Eliminar">
    <i class="fas fa-times"></i>
</button>
{{end}}
```

---

### Brecha 5: Botones reordenar procesos + función JS

**Problema:** La función `ReordenarRutaProcesosProceso` existe en app.go (línea 1458) pero no hay UI ni JS para invocarla. El endpoint no está registrado en handler.go (falta `/api/ruta-procesos-proceso-reordenar`).

**Cambios necesarios:**

1. **handler.go** — Agregar endpoint:
```go
case p == "/api/ruta-procesos-proceso-reordenar" && r.Method == http.MethodPost:
    h.handleReordenarProceso(w, r)
    return
```

2. **handler.go** — Agregar handler:
```go
func (h *TemplateHandler) handleReordenarProceso(w http.ResponseWriter, r *http.Request) {
    idJunta, _ := strconv.Atoi(r.FormValue("id_junta"))
    idProceso, _ := strconv.Atoi(r.FormValue("id_proceso"))
    direction, _ := strconv.Atoi(r.FormValue("direction"))
    if idJunta == 0 || idProceso == 0 || (direction != -1 && direction != 1) {
        writeJSONError(w, "Parámetros inválidos", http.StatusBadRequest)
        return
    }
    if err := h.app.ReordenarRutaProcesosProceso(idJunta, idProceso, direction); err != nil {
        writeJSONError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, map[string]interface{}{"success": true})
}
```

3. **ruta_procesos.html** — Agregar botones ▲▼ junto a cada proceso (en la fila del tbody).

4. **JS:**
```javascript
window.moverProceso = function(idJunta, idProceso, direction) {
    jsonPost('/api/ruta-procesos-proceso-reordenar',
        'id_junta=' + idJunta + '&id_proceso=' + idProceso + '&direction=' + direction
    ).then(function(res) {
        if (res.success) reload();
        else alert(res.error || 'Error');
    });
};
```

---

### Brecha 6: Reset del modal de leyenda al crear nueva

**Problema:** Al hacer clic en "Añadir" leyenda, el modal podría tener datos residuales de una edición previa.

**Cambio en el botón "Añadir" del template (línea 142):**
```html
<button onclick="abrirCrearLeyenda({{$j.ID}})" class="btn btn-secondary btn-sm text-xs">
    <i class="fas fa-plus mr-1"></i> Añadir
</button>
```

**Función JS nueva:**
```javascript
window.abrirCrearLeyenda = function(idJunta) {
    _editingLeyendaId = null;
    _editingJuntaId = idJunta;
    document.getElementById('leyenda-modal-title').textContent = 'Nueva Leyenda';
    document.getElementById('editar-leyenda-id').value = '';
    document.getElementById('leyenda-nombre').value = '';
    document.getElementById('leyenda-color').value = '#FFFFFF';
    document.getElementById('leyenda-ambito').disabled = false;
    document.getElementById('leyenda-ambito').value = 'junta';
    document.getElementById('leyenda-bloquear-wrap').classList.add('hidden');
    toggleModal('crear-leyenda-modal');
};
```

---

## Orden de Implementación

1. [ ] **CSS** — Agregar estilos del Gantt al `<style>` de ruta_procesos.html
2. [ ] **Template HTML** — Agregar `data-proc` y `data-fecha` a las celdas del Gantt
3. [ ] **JS — `colorearCeldas()`** — Post-render que colorea celdas con cronograma
4. [ ] **JS — `moverLeyenda()`** — Función reordenar leyendas
5. [ ] **JS — `toggleBloquearLeyenda()`** — Función toggle bloqueo
6. [ ] **Template — Leyendas** — Botones ▲▼ y toggle bloqueo clickeable
7. [ ] **handler.go** — Endpoint `/api/ruta-procesos-proceso-reordenar`
8. [ ] **JS — `moverProceso()`** — Función reordenar procesos
9. [ ] **Template — Procesos** — Botones ▲▼ junto a cada proceso
10. [ ] **JS — `abrirCrearLeyenda()`** — Reset modal al crear nueva leyenda
11. [ ] Build y pruebas

---

## Archivos a Modificar

| Archivo | Tipo de cambio |
|---|---|
| `templates/new/ruta_procesos.html` | CSS, HTML template, JS (mayor parte del trabajo) |
| `handler.go` | 1 endpoint nuevo + 1 handler nuevo |
| `frontend/vendor/styles.css` | Opcional (si preferimos CSS ahí en vez de inline) |

---

## Notas

- El backend está 100% listo. NO se necesitan cambios en `app.go` ni en `03_ruta_procesos.sql`.
- Los datos del cronograma ya se cargan correctamente en `ObtenerRutaProcesosData` → los procesos ya traen su `timeline` map poblado. Solo falta que el frontend lo renderice visualmente.
- El approach recomendado para colorear celdas es **post-render JS** (como hacía v1), identificando cada celda con `data-proc` + `data-fecha`.
