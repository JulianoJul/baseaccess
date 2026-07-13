# PLAN — wails-migration: integración multi-módulo, fixes de Gemini, docs y push

> Cambio de modelo: este plan es **auto-contenido** y reproducible por cualquier ORM/agente sin contexto previo. Todas las decisiones ya están tomadas. Solo queda ejecutar.

---

## 0. CONTEXTO DEL PROYECTO

- **Repo**: `/home/user/Documentos/proyecto/baseaccess`, rama **`wails-migration`** (rama paralela a `master`/`tauri-migration`).
- **App**: "Gestión de Expedientes con Historial" — desktop app Wails v2 (Go + WebView) para registrar documentos con historial de movimientos.
- **Stack**: Wails v2 + Go 1.25 + `mattn/go-sqlite3` + Go `html/template` + HTMX + Tailwind CSS (sin frameworks JS).
- **Importante**: `master` queda intacto; esta rama es un rewrite. `src/`, `src-tauri/`, `main.js`, `package.json` son **legacy** (Electron/Tauri) y NO se tocan.

### Estado actual de la rama (antes de ejecutar este plan)

`git status` muestra cambios sin commitear por "Gemini":
- **Modificados**: `app.go`, `handler.go`, `templates/index.html`, `frontend/wailsjs/go/main/App.d.ts`, `frontend/wailsjs/go/main/App.js`.
- **Sin seguimiento (nuevos)**: 18 templates — 9 `tabla_<key>.html` + 9 `form_<key>.html` (uno por módulo declarado en `Modulos` map de `app.go`).
- **SQL nuevos committed**: `data/sql/01_master_control_docs_presidencia.sql` y `data/sql/02_modulos_adicionales.sql` (commit `5ad809d`).
- **Sin commitear**: ningún archivo SQL nuevo.

El commit `5ad809d` tiene el mensaje "feat(sql): add master and separated modules for presidencia docs, include new gerencias and bind docs to cat_documento". La integración en código Go y el front quedó SIN commitear pero en el dir de trabajo.

---

## 1. MAPEO EXCEL → BDD (verificado leyendo `data/CONTROL DE DOCUMENTOS JUNIO 2026.xlsx`)

| Columnas Excel (por hoja) | BDD destino | Notas |
|---|---|---|
| `OBSERVACIONES HISTORIAL` (col 20) | `historial_movimientos.observaciones` (snapshot) | Una línea resumen del último movimiento |
| `HISTORIAL COMPLETO` (col 21) | (se concatena en el snapshot de `historial_movimientos.observaciones`) | Texto multi-línea con todos los movimientos |
| `OBSERVACIONES` (col 34, última) | `expedientes.notas` | Observación libre del expediente actual |
| Hojas de módulos (req, mem, rec, val, jd, bdu, vac, rep) | `<tabla>.notas` + `hist_<tabla>.observaciones` | Mismo patrón |

**Bug conocido en `data/importar_datos.py:119-136`**: llena `expedientes.observaciones` (columna central del expediente, no del historial) con el `OBSERVACIONES HISTORIAL` del Excel — eso es conceptualmente mezcla. Pero es **fuera de scope de este plan** (script de migración se actualizará aparte según el usuario: "luego se actualiza para poder ejecutarlo sobre el schema limpio").

### Conclusión para el front

- **`templates/historial.html`** ya muestra `observaciones` (snapshot del movimiento, columna OK). **ADEMÁS debe mostrar `notas`**, porque cada snapshot en `hist_<tabla>` captura ambos campos (`NEW.observaciones, NEW.notas` copiados por los triggers en INSERT y en UPDATE). Si el usuario escribió una nota libre en alguno de los N movimientos del expediente, esa nota se conserva en ese snapshot histórico y la UI debe reflejarla. Ver justificación extendida en §2.2.1 + bug #4(b).
- Por tanto, el bug "notas nunca se muestra en el historial" **SÍ es real** y se resuelve añadiendo una columna "Notas" al `historial.html` (ver §B.3).

---

## 2. HALLAZGOS (auditoría completada)

### 2.1 Bug funcional REAL — bloquea ejecución

| # | Archivo:line | Bug | Fix |
|---|---|---|---|
| 1 | `data/sql/02_modulos_adicionales.sql:71` | `trg_req_mat_auditoria` referencia `NEW.documento`, columna que NO existe en `req_materiales` (el campo correcto es `id_documento` FK + `descripcion_materiales`). Trigger `_inicial` (línea 65) está correcto; el `_auditoria` fue copy-paste mal hecho de `trg_mem_auditoria` (que sí usa `NEW.documento` porque `memorandums` sí tiene `documento TEXT`). **Todo UPDATE a `req_materiales` fallará con `no such column: documento`**. | Reemplazar en línea 71: `NEW.documento` → `NEW.descripcion_materiales`. Réplica exacta del patrón del trigger `_inicial` (línea 65). |

### 2.2 Bugs de integración front

| # | Archivo:line | Bug | Fix |
|---|---|---|---|
| 2 | `templates/index.html:176` | `{{template "formulario.html" .}}` legacy embebido estáticamente en `#form-expediente`. htmx lo reemplaza al abrir el modal, pero si la primera petición htmx falla el usuario ve un form de expedientes vacío en lugar de un error. `formulario.html` ya no es referenciado por handler.go (solo por esta línea). | Eliminar `{{template "formulario.html" .}}` de index.html. Dejar `<form id="form-expediente" class="space-y-6" onsubmit="return false;"></form>` vacío. |
| 3 | `templates/index.html:172` y `:345-350` | `mostrarFormulario()` siempre dice "Nuevo Expediente"/"Editar Expediente #N" sin importar el módulo activo. Todo el resto de la app respeta `#active-module-val`. | En `mostrarFormulario()`, leer `$('active-module-val').value` y traducirlo a `Modulos[key].Nombre` (ya expuesto vía `window.PAGE_DATA.modulos`). Build string: `id ? 'Editar ' + nombre + ' #' + id : 'Nuevo ' + nombre`. |
| 4 | `templates/historial.html` (template único) | Plantilla rígida de 6 columnas. **(a)** Columna "Receptor" aparece como "-" en `reposos_medicos` (su `QueryHistorial` no hace JOIN porque la tabla base no tiene `id_receptor`). **(b)** NO muestra `notas` — pese a que todos los `QueryHistorial` en `app.go` (líneas 52, 66, 79, 93, 109, 124, 138, 151, 164) proyectan `h.notas`, y todos los triggers SQL copian `NEW.notas` al snapshot histórico en cada INSERT/UPDATE. Esto significa que cuando un expediente pasa por varios cambios y en alguno el usuario escribió una nota libre, esa nota se conserva en el snapshot del historial pero la UI nunca la muestra. | **(a)** En `historial.html`, envolver la columna "Receptor" con `{{if ne .ActiveModule "reposos_medicos"}}…{{end}}`. **(b)** Añadir una columna "Notas" al template (entre "Observaciones" y el cierre). handler.go debe pasar `ActiveModule` al template (ver §B.3). |

#### 2.2.1 Justificación funcional de mostrar `notas` en el historial

Cada expediente/registro nace con un primer snapshot en `hist_<tabla>` (vía trigger `*_inicial` en INSERT) y acumula un snapshot nuevo en cada UPDATE (vía trigger `*_auditoria`). El usuario explicó el flujo así:

> "hay un expediente, el nace con ciertos datos, ese es su primer registro o entrada en el historial, luego cambiamos unas cosas y ese seria su segundo registro, y asi sucesivamente, en caso de que en uno de esos pasos se le ponga nota es importante que salga en el historial tambien"

Por tanto, el snapshot de cada movimiento histórico captura tanto `observaciones` (línea resumen del movimiento) como `notas` (observación libre del expediente en ese punto). Ambas columnas existen en `hist_<tabla>` porque los triggers las copian fielmente (`NEW.observaciones, NEW.notas` — verificado en `02_modulos_adicionales.sql:65, 71, 156, 162, 255, 261, 378, 384, 501, 507, 608, 614, 706, 712, 794, 800` y en `01_master_control_docs_presidencia.sql:184-185, 216-217`). La UI debe reflejar ambas columnas para que el historial sea fiel a la BDD.

### 2.3 Bugs cosméticos (FUERA DE SCOPE — se omite en este pase per user)

- `templates/form_recobros.html:27` — label "Asunto del Memorándum" copiado.
- `templates/form_vacaciones.html:26` — `anio` sin `min`/`max`.
- `templates/form_expedientes.html` — usa `.Expediente` (el resto usa `.Registro`).
- `templates/form_expedientes.html:150` — placeholder `ej: 120` en `tiempo_ejecucion` (TEXT).
- `templates/tabla_filas.html` y `templates/formulario.html` — legados huérfanos (solo `formulario.html` se referencia aún en `index.html:176`; ver bug #2).

**Decisión del user**: "Ninguno, solo bug crítico SQL + bugs 2-5 front + docs". Los cosméticos y la eliminación de `tabla_filas.html`+`formulario.html` (más allá del fix #2 que retira la referencia) se dejan para otro PR.

### 2.4 Bindings wailsjs (auto-generados)

`frontend/wailsjs/go/main/App.d.ts` y `App.js` ya fueron regenerados por `wails dev`/`wails build` e incluyen:
- Métodos nuevos: `EliminarFila`, `GuardarFila`, `ObtenerFilaPorId`, `ObtenerFilas`, `ObtenerHistorialFila`.
- Métodos legacy (`ObtenerExpedientes`, `GuardarExpediente`, etc.) siguen presentes como wrappers de `app.go:589-607`.

**Decisión del user**: "lo que sea más robusto y limpio para un futuro" → **eliminar los wrappers legacy de app.go y migrar `handler.go:675` (`handleCSV`) a `ObtenerFilas`**, así el catálogo de bindings se simplifica (aunque los archivos `.d.ts`/`.js` se regenerarán en el próximo `wails dev`).

### 2.5 Scripts SQL — idempotencia

**Decisión del user**: "pues yo tengo un script de migración, será que luego se actualiza para poder ejecutarlo sobre el schema limpio" → **NO se tocan los `.sql` más allá del bug #1.** El usuario actualizará `data/importar_datos.py` después para correrlo sobre schema limpio.

### 2.6 Docs desactualizados

| # | Archivo | ¿Qué corregir? |
|---|---|---|
| 5 | `docs/doc.md:39-50` (diagrama arquitectura) | Lista `tabla_filas.html` y `formulario.html` como vivos (ya no). Lista métodos `ObtenerExpedientes`/`GuardarExpediente`/`EliminarExpediente` como primarios (renombrados). |
| 6 | `docs/doc.md:121-135` (tabla del schema) | Solo lista `expedientes`+`historial_movimientos`+1 vista. Faltan 8 tablas + 8 vistas + 16 triggers de los módulos nuevos (requisiciones, memorandums, recobros, valuaciones, aprobacion_jd, certificacion_bdu, vacaciones, reposos_medicos). |
| 7 | `docs/doc.md:237-243` (changelog) | Cita `tabla_filas.html` y `formulario.html` como vivos. Añadir entradas #28+ para multi-módulo. |
| 8 | `docs/doc.md:248-255` (roadmap) | "PageData inyecta_expedientes" → cambiar a "Filas" / multi-módulo. |
| 9 | `docs/doc.md:261-271` (tabla rutas API) | Faltan: `/api/cambiar-modulo` (nueva ruta). |
| 10 | `docs/funciones.md:7-26` (catálogo Go) | Enumerar `ObtenerFilas`, `ObtenerFilaPorId`, `GuardarFila`, `EliminarFila`, `ObtenerHistorialFila` como primarios. Marcar `*Expediente*` como eliminados/legacy. |
| 11 | `docs/funciones.md:32-43` (capa JS) | Las funciones JS ahora en su mayor parte reemplazadas por HTMX. Se simplifica la tabla. |
| 12 | `docs/decisiones.md` | Añadir **DEC-014**: "Multi-módulo" — schema separado (`01_master` + `02_modulos`), `Modulos` map en `app.go`, botonera inferior, templates fragmentados `tabla_<key>.html`/`form_<key>.html`, migración del título dinámico del modal. |
| 13 | `docs/ai-context.md:18-19` | "Estado Actual Julio 2026" actualizado para mencionar multi-módulo. |
| 14 | `docs/ai-context.md:21-35` (tabla "Archivos Clave") | Añadir entradas: `data/sql/01_master_*.sql`, `data/sql/02_modulos_*.sql`, `templates/tabla_<key>.html`, `templates/form_<key>.html`. |
| 15 | `docs/doc.md` (sección schemas) | Mencionar `cat_gerencia` con 13 gerencias (IDs 11-13 nuevas: PROCURA, CONTROL DE DOCUMENTOS, ASUNTOS PÚBLICOS). |

---

## 3. ESTRUCTURA DE COMMITS (decisión del user: **tres commits separados**)

1. `fix(sql): corrige trg_req_mat_auditoria en 02_modulos_adicionales (NEW.documento → NEW.descripcion_materiales)`
2. `feat: integra multi-modulo — modal dinamico, historial sin receptor en reposos, elimina wrappers legacy de app.go/handler.go`
3. `docs: actualiza doc/funciones/decisiones/ai-context para multi-modulo`

Push final: `git push origin wails-migration` con **`[skip ci]`** en el cuerpo del mensaje de commit (el usuario compilará localmente desde Linux).

---

## 4. EJECUCIÓN PASO A PASO

### FASE A — Commit 1 (fix SQL)

#### A.1 Corregir `data/sql/02_modulos_adicionales.sql:71`

Edit en `trg_req_mat_auditoria` (línea 71). Palabra exacta a buscar y reemplazar:

```diff
-    VALUES (NEW.id_requisicion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.documento, NEW.serial_equipo, NEW.pase_sicesma, NEW.id_estatus, NEW.observaciones_entrega, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
+    VALUES (NEW.id_requisicion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.descripcion_materiales, NEW.serial_equipo, NEW.pase_sicesma, NEW.id_estatus, NEW.observaciones_entrega, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
```

**Verificación visual**: comparar contra trigger `_inicial` (línea 65) — deben ser idénticos salvo el `WHERE`/`SET fecha_actualizacion` final.

#### A.2 Commit 1

```bash
git add data/sql/02_modulos_adicionales.sql
git commit -m "$(cat <<'EOF'
fix(sql): corrige trg_req_mat_auditoria en 02_modulos_adicionales

El trigger trg_req_mat_auditoria (línea 71) referenciaba NEW.documento,
columna inexistente en req_materiales (copia-pestana de trg_mem_auditoria).
Causaba fallo "no such column: documento" en todo UPDATE a req_materiales.

Reemplazado por NEW.descripcion_materiales, replicando el patrón correcto
del trigger trg_req_mat_inicial (línea 65).

[skip ci]
EOF
)"
```

---

### FASE B — Commit 2 (front + app.go/handler.go)

#### B.1 Quitar referencia a `formulario.html` en `templates/index.html`

Buscar:
```
<form id="form-expediente" class="space-y-6" onsubmit="return false;">
    {{template "formulario.html" .}}
</form>
```
Reemplazar por:
```
<form id="form-expediente" class="space-y-6" onsubmit="return false;"></form>
```
(Location: aproximadamente línea 176; verificar que el `<form id="form-expediente">` quede vacío.)

**No se elimina `templates/formulario.html`** (cosmético fuera de scope).

#### B.2 Título dinámico del modal en `templates/index.html`

Encontrar el bloque `function mostrarFormulario(id)` (alrededor de línea 345-350):
```js
function mostrarFormulario(id) {
    const modal = $('form-modal');
    $('form-titulo').textContent = id ? 'Editar Expediente #' + id : 'Nuevo Expediente';
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
}
```

Reemplazar por:
```js
function mostrarFormulario(id) {
    const modal = $('form-modal');
    const moduloKey = (window.PAGE_DATA && window.PAGE_DATA.modulos && $('active-module-val'))
        ? $('active-module-val').value
        : 'expedientes';
    const nombreModulo = (window.PAGE_DATA && window.PAGE_DATA.modulos && window.PAGE_DATA.modulos[moduloKey])
        ? window.PAGE_DATA.modulos[moduloKey].Nombre
        : 'Registro';
    $('form-titulo').textContent = id ? 'Editar ' + nombreModulo + ' #' + id : 'Nuevo ' + nombreModulo;
    modal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
}
```

**Verificación**: `window.PAGE_DATA.modulos` debe existir — revisar `templates/index.html:42-50` (el bloque `<script>window.PAGE_DATA = {...}</script>`). Actualmente inyecta `hasDB`, `dbPath`, `catalogs`, `expedientes`, `totalPages`, `currentPage`, `pageSize`. **Falta `modulos`** — hay que añadirlo al script.

Inyectar `modulos: {{jsonEncode .Modulos}}` al objeto `window.PAGE_DATA` en `templates/index.html:42-50`. Usar el helper `jsonEncode` ya existente en `handler.go:42,49,182`.

**Resultado:**
```js
window.PAGE_DATA = {
    hasDB: {{.HasDB}},
    dbPath: "{{.DBPath}}",
    catalogs: {{jsonEncode .Catalogs}},
    modulos: {{jsonEncode .Modulos}},
    expedientes: {{jsonEncode .Expedientes}},   // (legacy — cambiar nombre opcional, pero no break)
    totalPages: {{.TotalPages}},
    currentPage: {{.CurrentPage}},
    pageSize: {{.PageSize}}
};
```

(`Expedientes` sigue siendo una key válida si PageData la tiene — verificar handler.go:212-225 que `PageData` define `Filas []Row` ya, no `Expedientes`. Si el index.html referencía `.Expedientes` ahora, eso ya está roto. Confirmar leyendo el diff — según el diff de handler.go, el struct fue renombrado `Expedientes` → `Filas`. El template viejo tenía `.Expedientes` y se dump-aba a JSON; ahora debe ser `.Filas`. Actualmente index.html:`expedientes: {{jsonEncode .Expedientes}}` está referenciando un campo inexistente ¿rompe Go? — NO, Go html/template solo deja vacío, no error. Pero para limpiar, debería ser `.Filas`. **Acción incluida más abajo**.)

##### B.2b Sincronizar `index.html:42-50` con PageData renombrada

`handler.go:212-225` define PageData con: `Title, HasDB, Catalogs, ActiveModule, Modulos, Filas, PageSize, TotalPages, CurrentPage, SortColumn, SortDir, DBPath, Registro`. NO tiene `Expedientes` ni `Expediente`.

Acción: en index.html:46-50, reescribir el script:
```js
window.PAGE_DATA = {
    hasDB: {{.HasDB}},
    dbPath: "{{.DBPath}}",
    catalogs: {{jsonEncode .Catalogs}},
    modulos: {{jsonEncode .Modulos}},
    filas: {{jsonEncode .Filas}},
    totalPages: {{.TotalPages}},
    currentPage: {{.CurrentPage}},
    pageSize: {{.PageSize}}
};
```

Si en cualquier parte del JS de index.html se referencia `window.PAGE_DATA.expedientes`, reemplazar por `window.PAGE_DATA.filas`. (Buscar antes de ejecutar.)

#### B.3 Hacer `historial.html` condicional para `reposos_medicos`

`handler.go:562-587` (`handleHistorial`) actualmente pasa solo `rows` al template. Necesitamos pasar `ActiveModule` también. Re-escribir el handler:

```go
func (h *TemplateHandler) handleHistorial(w http.ResponseWriter, r *http.Request) {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	rows, err := h.app.ObtenerHistorialFila(modulo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Rows         []Row
		ActiveModule string
	}{
		Rows:         rows,
		ActiveModule: modulo,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmplName := "historial.html"
	if err := h.tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		log.Printf("render error for %s: %v", tmplName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

**Nota crítica**: El template `historial.html` actualmente itera con `{{range .}}` directamente (toma `Rows` implícito). Si pasamos un struct `data`, `{{range .}}` no itera filas, hay que cambiar a `{{range .Rows}}` y todos los `{{rowGetStr . "..."}}` quedan igual (porque dentro del range el `.` es la fila).

Reescribir `templates/historial.html`:

```html
{{if eq (len .Rows) 0}}
    <p class="text-gray-500 italic">Sin movimientos registrados.</p>
{{else}}
    <table class="w-full text-left text-xs border-collapse">
        <thead>
            <tr class="bg-gray-700/60 text-teal-400 uppercase">
                <th class="p-2">Fecha Recibido</th>
                <th class="p-2">Estatus</th>
                <th class="p-2">Documento</th>
                <th class="p-2">Emisor</th>
                {{if ne .ActiveModule "reposos_medicos"}}<th class="p-2">Receptor</th>{{end}}
                <th class="p-2">Observaciones</th>
                <th class="p-2">Notas</th>
            </tr>
        </thead>
        <tbody class="divide-y divide-gray-700">
            {{range .Rows}}
            <tr class="hover:bg-gray-700/20">
                <td class="p-2 font-mono">{{default (rowGetStr . "fecha_recibido") "-"}}</td>
                <td class="p-2">{{default (rowGetStr . "estatus") "-"}}</td>
                <td class="p-2">{{default (rowGetStr . "documento") "-"}}</td>
                <td class="p-2">{{default (rowGetStr . "emisor") "-"}}</td>
                {{if ne $.ActiveModule "reposos_medicos"}}<td class="p-2">{{default (rowGetStr . "receptor") "-"}}</td>{{end}}
                <td class="p-2 text-gray-400">{{default (rowGetStr . "observaciones") "-"}}</td>
                <td class="p-2 text-gray-400">{{default (rowGetStr . "notas") "-"}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
{{end}}
```

**Notas críticas sobre este template**:
- `{{range .Rows}}` itera las filas; dentro del range, `.` es cada fila individual (pasa a `rowGetStr`).
- `$.ActiveModule` (con `$`) referencia el root data para acceder al `ActiveModule` fuera del range. Alternativamente `$.ActiveModule` puede usarse como `.ActiveModule` solo fuera del `{{range}}`; dentro del range hay que usar `$.ActiveModule` (Go html/template: `$` es el data original).
- El `{{if ne .ActiveModule "reposos_medicos"}}` del `<th>` se evalúa FUERA del range (en el `data` struct), entonces `.ActiveModule` funciona. El del `<td>` se evalúa DENTRO del range, por eso usa `$.ActiveModule`.
- `default` y `rowGetStr` son helpers ya registrados en `handler.go:38,104-113,201-207`.
- `COALESCE(h.notas, '') AS notas` en los QueryHistorial garantiza que `rowGetStr . "notas"` retorne `""` si la BDD tiene NULL; el helper `default` lo cambiará a `"-"` para mostrar un guion en la tabla en lugar de celda vacía.

**Verificar**: En `historial.html`, HTMX inserta el resultado dentro del contenedor del modal. El contenedor espera HTML en fila-sucesiva. Cambiar de `rows []Row` a `struct { Rows, ActiveModule }` es seguro — el `{{if eq (len .Rows) 0}}` tomará el conteo.

**Confirmación**: revisar templates/index.html para confirmar `hx-target="#historial-cuerpo"` (o similar) donde se inyecta el historial. La firma HTML resultante (tabla `<table>…</table>` o `<p>…</p>`) no cambia, así que el hx-target sigue funcionando.

#### B.4 Eliminar wrappers legacy de `app.go` y limpiar `handler.go`

`app.go:587-607` tiene:
```go
// --- legacy wrapper functions for backward compatibility ---

func (a *App) ObtenerExpedientes(orden string) ([]Row, error) {
	return a.ObtenerFilas("expedientes", orden)
}

func (a *App) ObtenerExpedientePorId(id int) (Row, error) {
	return a.ObtenerFilaPorId("expedientes", id)
}

func (a *App) GuardarExpediente(data map[string]interface{}) (int64, error) {
	return a.GuardarFila("expedientes", data)
}

func (a *App) EliminarExpediente(id int64) error {
	return a.EliminarFila("expedientes", id)
}

func (a *App) ObtenerHistorialCompleto(id int) ([]Row, error) {
	return a.ObtenerHistorialFila("expedientes", id)
}
```

**Acción**: Eliminar todo ese bloque (comentario incluido).

`handler.go:675` (en `handleCSV`) usa `h.app.ObtenerExpedientes("id_expediente DESC")`. Cambiar a:
```go
data, err := h.app.ObtenerFilas("expedientes", "id_expediente DESC")
```

**Verificación grep**: Buscar en todo el repo (archivos `.go`) referencias a estos métodos eliminados:
```bash
rg '\b(ObtenerExpedientes|GuardarExpediente|EliminarExpediente|ObtenerExpedientePorId|ObtenerHistorialCompleto)\b' --type go
```
Debe dar cero coincidencias en `.go`. (Los bindings `App.d.ts`/`App.js` se regeneran solos en el próximo `wails dev` — no editarlos a mano.)

#### B.5 Compilar y verificar

```bash
go vet ./...
go build ./...
```

`go vet` debe dar cero warnings. `go build` debe producir el binario sin errores. Si hay referencias residuales a wrappers eliminados, corregirlas.

**No ejecutar `wails dev` en este paso** — lo hará el usuario al final.

#### B.6 Commit 2

```bash
git add app.go handler.go templates/index.html templates/historial.html
git commit -m "$(cat <<'EOF'
feat: integra multi-modulo en front y limpia API legacy

- index.html: titulo de modal dinamico (Nuevo/Editar <Modulo.Nombre>)
  lee #active-module-val via window.PAGE_DATA.modulos. Elimina
  {{template "formulario.html"}} embebido (htmx lo reemplaza).
- index.html: PAGE_DATA sincronizado con PageData (Filas en vez de
  Expedientes; anyade Modulos).
- handler.go: handleHistorial pasa ActiveModule al template. handleCSV
  usa ObtenerFilas en vez del wrapper ObtenerExpedientes.
- historial.html: anyade columna Notas (snapshot del historial por
  trigger ya proyecta h.notas). Omite columna Receptor para
  reposos_medicos (tabla sin id_receptor). Recibe struct
  {Rows, ActiveModule}.
- app.go: eliminados wrappers legacy ObtenerExpedientes,
  ObtenerExpedientePorId, GuardarExpediente, EliminarExpediente,
  ObtenerHistorialCompleto. API primaria: ObtenerFila/GuardarFila/etc.

[skip ci]
EOF
)"
```

---

### FASE C — Commit 3 (docs)

#### C.1 `docs/doc.md`

**C.1.1** Diagrama arquitectura (líneas 18-56) — corregir:
- Línea 39: `tabla_filas.html` → eliminar (o cambiar a `tabla_<key>.html` (9 plantillas)`)
- Línea 43: `formulario.html` → `form_<key>.html` (9 plantillas)`
- Líneas 48-51: cambiar métodos a `ObtenerFilas`, `GuardarFila`, `EliminarFila` y mencionar "...y moduloKey como primer arg".

**C.1.2** Sección "Tablas del Schema" (líneas 121-135) — añadir las 16 tablas nuevas (8 principales + 8 históricas) y las 8 vistas nuevas. Mencionar `cat_gerencia` con 13 gerencias.

**C.1.3** Sección de routas (líneas 261-271) — añadir `/api/cambiar-modulo | GET | Devuelve fragmento HTML de la tabla del módulo solicitado`.

**C.1.4** Changelog (líneas 244-245) — añadir entradas:
```
| 28 | `data/sql/01_master_control_docs_presidencia.sql`, `data/sql/02_modulos_adicionales.sql` | Creados: schema multi-módulo (master + 8 módulos adicionales con sus tablas hist_, vistas vw_reporte_*, triggers) | Soporte a 9 módulos en una sola BD |
| 29 | `app.go`, `handler.go`, `templates/*` | Multi-módulo: `Modulos` map (9 módulos), botonera inferior, `tabla_<key>.html`/`form_<key>.html` fragmentados, título dinámico del modal, historial condicional (`reposos_medicos` sin receptor) | UI y API multi-módulo |
```

**C.1.5** Sección "Migración a Go html/template — Estado" — actualizar "PageData inyecta catálogos_y Filas (multi-módulo) precargados".

#### C.2 `docs/funciones.md`

**C.2.1** Tabla "Backend Go — Métodos exportados (app.go)" (líneas 7-26) — eliminar las 5 filas `*Expediente*` y reemplazar por:
```
| `ObtenerFilas(moduloKey, orden)` | `moduloKey`: key de Modulos map; `orden`: columna DESC/ASC | SELECT * FROM cfg.Vista con sanitización whitelist |
| `ObtenerFilaPorId(moduloKey, id)` | `moduloKey`, `id`: int | Retorna `Row` única o error |
| `GuardarFila(moduloKey, data)` | `moduloKey`, `data`: map[string]interface{} | INSERT o UPDATE según presencia de cfg.IDColumna |
| `EliminarFila(moduloKey, id)` | `moduloKey`, `id`: int64 | DELETE en transacción (historial + tabla módulo) |
| `ObtenerHistorialFila(moduloKey, id)` | `moduloKey`, `id`: int | SELECT cfg.QueryHistorial (JOIN multi-tabla) |
| `ObtenerRutaProcesos()` | — | (específico de expedientes) |
| `ObtenerDocumentosPendientes()` | — | (específico de expedientes) |
| `AbrirBaseDatos(filePath)` | filePath | Abre BD SQLite con WAL + foreign_keys |
| `CerrarBaseDatos()` | — | Cierra la BD |
| `ObtenerCatalogos()` | — | map[string][]CatalogoItem (11 catálogos) |
| `GuardarNuevoCatalogo(tabla, nombre, extra)` | — | INSERT whitelist |
| `OptimizarBD()` | — | VACUUM |
| `AbrirDialogoBD()`, `GuardarDialogoBD(...)`, `SetBackupMaxCopies(n)`, `GetBackupMaxCopies()`, `DescargarBD(destPath)` | — | (diálogos y backup nativos Wails) |
```

**C.2.2** Capa JS — simplificar, ya que la mayoría de funciones JS se eliminaron con HTMX:

Reemplazar las líneas 28-43 con nota indicativa:
> "La mayoría de las funciones JS previas (cargarCatalogos, obtenerExpedientes, guardarExpedienteEnBd, etc.) fueron reemplazadas por HTMX. JS actual mínimo: helpers de modales, paginación DOM del lado del cliente, localStorage (recientes/fijados), y `abrirBaseDatos()` (única función que invoca binding Wails `AbrirDialogoBD`)."

#### C.3 `docs/decisiones.md`

Añadir **DEC-014** al final (después de DEC-013):
```markdown
---

## DEC-014: Multi-módulo con schema separado y Modulos map

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** La app gestionaba originalmente un solo tipo de documento (expedientes de contrataciones). El usuario extendió el control a 9 tipos: expedientes, requisiciones, memorandums, recobros, valuaciones, aprobacion_jd, certificacion_bdu, vacaciones, reposos_medicos. Cada tipo tiene su propia tabla principal, tabla de historial, vista de reporte y trigger de auditoría. Se dividió el schema en dos archivos SQL limpios: `01_master_control_docs_presidencia.sql` (catálogos + expedientes) y `02_modulos_adicionales.sql` (8 módulos restantes). En código Go, se unificó la API con `var Modulos map[string]ModuloConfig` (app.go), cada entrada define `Tabla`, `Vista`, `IDColumna`, `HistorialTabla`, `Columnas`, `QueryHistorial`. Los handlers y templates se fragmentan en `tabla_<key>.html`/`form_<key>.html` y se despachan via `{{if eq .ActiveModule "<key>"}}`. Se añadió botonera inferior en index.html para cambiar de módulo sin recargar la página (HTMX swap de `#vista-tabla`).
- **Alternativas evaluadas:**
  - Una sola tabla polimórfica con `tipo_documento` — descartado: perdía integridad referencial y tipado de columnas.
  - 9 esquemas SQLite separados — descartado: backup rotativo y apertura de BD por usuario no lo justifican.
  - Multi-módulo con `Modulos` map + schema separado — elegido: DRY en la API Go, schemas limpios e independientes, UI unificada.
- **Impacto:**
  - `data/sql/01_master_control_docs_presidencia.sql` + `02_modulos_adicionales.sql` creados (Tablas8.sql queda obsoleto, conservado por histórico). `cat_gerencia` ampliada con 3 gerencias (IDs 11-13: PROCURA, CONTROL DE DOCUMENTOS, ASUNTOS PÚBLICOS).
  - `app.go`: `Modulos map[string]ModuloConfig` (9 entradas). API primaria renombrada: `ObtenerFilas/FilaPorId/GuardarFila/EliminarFila/ObtenerHistorialFila` con `moduloKey` como primer arg. Wrappers legacy `*Expediente*` eliminados.
  - `handler.go`: ruta nueva `/api/cambiar-modulo`. `handleCSV` migrado. `handleHistorial` pasa `ActiveModule` al template.
  - `templates/`: 18 nuevos templates (9 `tabla_<key>.html` + 9 `form_<key>.html`). `historial.html` condicional (`{{if ne .ActiveModule "reposos_medicos"}}` para columna Receptor). `index.html`: botonera inferior `{{range $key, $cfg := .Modulos}}`, título de modal dinamico via `window.PAGE_DATA.modulos`.
  - `templates/formulario.html` y `templates/tabla_filas.html`: legados sin uso (referencia a `formulario.html` en index.html removida). [su eliminación física queda para un futuro PR cosmético].
```

#### C.4 `docs/ai-context.md`

**C.4.1** Líneas 18-19 (Estado Actual Julio 2026) — reemplazar por:
```
## Estado Actual (Julio 2026)
App con **Wails v2 + Go html/template + HTMX**, ahora **multi-módulo** (9 tipos de documentos: expedientes, requisiciones, memorandums, recobros, valuaciones, aprobacion_jd, certificacion_bdu, vacaciones, reposos_medicos). Schema dividido en `data/sql/01_master_control_docs_presidencia.sql` + `data/sql/02_modulos_adicionales.sql`. API Go unificada vía `var Modulos map[string]ModuloConfig` en `app.go`. Botonera inferior en `index.html` para conmutar módulos sin recargar. `frontend/wailsjs/` se regenera con `wails dev`. Único binding Wails utilizado: `AbrirDialogoBD`. Rama `wails-migration` activa.
```

**C.4.2** Tabla "Archivos Clave" (líneas 21-35) — añadir:
```
| `data/sql/01_master_control_docs_presidencia.sql` | Schema: catálogos + expedientes + historial_movimientos + vistas + triggers |
| `data/sql/02_modulos_adicionales.sql` | Schema: 8 módulos adicionales con sus tablas hist_, vistas vw_reporte_*, triggers |
| `templates/tabla_<key>.html` (9) | Plantilla de listado por módulo |
| `templates/form_<key>.html` (9) | Plantilla de formulario por módulo |
```

#### C.5 Commit 3

```bash
git add docs/doc.md docs/funciones.md docs/decisiones.md docs/ai-context.md plan.md
git commit -m "$(cat <<'EOF'
docs: actualiza doc, funciones, decisiones y ai-context para multi-modulo

- doc.md: arquitectura lista tabla_<key>/form_<key>; schema incluye 16
  tablas+8 vistas nuevas; rutas API incluye /api/cambiar-modulo;
  changelog anyade entradas #28-29. cat_gerencia ahora con 13 IDs.
- funciones.md: API primaria ObtenerFila/FilaPorId/GuardarFila/EliminarFila/
  ObtenerHistorialFila. Wrappers legacy eliminados. Capa JS simplificada.
- decisiones.md: anyade DEC-014 (Multi-modulo con schema separado y
  Modulos map).
- ai-context.md: "Estado Actual Julio 2026" actualizado. Archivos Clave
  incluyen sql/01_master, sql/02_modulos, templates tabla_/form_<key>.
- plan.md: documento de planeacion auto-contenido para replicabilidad.

[skip ci]
EOF
)"
```

---

### FASE D — Push final

```bash
git push origin wails-migration
```

Si el remote rechaza (fast-forward), NO usar `--force` sin consultar. Hacer `git pull --rebase origin wails-migration` primero, resolver conflictos y luego push.

Tras el push:
- CI en GitHub Actions disparará job `wails` (Linux+Windows) — los 3 commits llevan `[skip ci]` así que NO debería dispararse (confirmar leyendo `.github/workflows/build.yml` para ver si respeta `[skip ci]`).
- El usuario compilará localmente desde Linux para probar: `make wails-build-linux` (debug) o `make wails-build-linux-prod`.

---

## 5. VERIFICACIÓN FINAL (smoke test manual tras push, por el user)

1. `make wails-build-linux` ( genera `build/bin/GestionExpedientes`).
2. Ejecutar el binario. Window 1400×900 abre.
3. Abrir `data/expedientes.db` ( debe cargar catálogos y filas ).
4. Click en botonera inferior → cambiar entre los 9 módulos. Cada uno muestra su propia tabla.
5. Click "Nuevo Registro" → modal abre con título "Nuevo <Modulo.Nombre>" (no "Expediente").
6. Llenar un form de `requisiciones` y guardar → toast "Registro guardado". Listado actualiza.
7. Editar esa fila (UPDATE) → trigger `trg_req_mat_auditoria` YA NO falla (fix #1).
8. Click "Ver Historial" en la fila → modal muestra tabla. Para `reposos_medicos`, columna Receptor no aparece. Para los demás, sí. La columna **Notas** aparece para todos los módulos; si el expediente tuvo una nota en algún snapshot (verificado con `SELECT id_movimiento, notas FROM historial_movimientos WHERE notas IS NOT NULL AND notas != '' LIMIT 5;` en SQLite CLI antes de probar), debe verse el texto en la celda correspondiente; si no hay nota en ese snapshot, debe verse "-".
9. Si algo de esto rompe, viene un nuevo pase de fix + commit.

---

## 6. RIESGOS Y NOTAS

- **Bindings wailsjs**: `frontend/wailsjs/go/main/App.d.ts` y `App.js` siguen teniendo los métodos legacy después de eliminar los wrappers de app.go. **No se editan a mano**; el próximo `wails dev` o `wails build` los regenerará sin esos métodos. Si el usuario ejecuta `wails dev` antes del push, los bindings se regeneran y se quedan en el dir de trabajo — en ese caso añadirlos al Commit 2 con `git add frontend/wailsjs/`.
- **`make combine`**: si se corre después de los cambios, `combined.txt` puede crecer significativamente. Verificar que `.gitignore` lo excluye (no lo hace — está committed como `combined.txt` en `git ls-files`. Pero el Makefile lo genera — es un build artifact; considerar añadirlo a `.gitignore` en el futuro. FUERA DE SCOPE ahora).
- **`docs/decisiones.md`** crece grande; el archivo sigue el patrón de los ADR previos sin refactor.
- **`data/expedientes.db`**: si el usuario ya tiene un .db populado con schema viejo, ejecutar los scripts `01_`+`02_` fallará (no tienen `DROP IF EXISTS` ni `INSERT OR IGNORE`). Confirmado: el usuario actualizará `importar_datos.py` después para recrear la BD limpia desde los nuevos scripts.
- **`PRAGMA user_version`** está comentado en `01_:397` y en `02_` no se setea. Recomendación (fuera de scope): descomentar y bump a `9` en `01_` (nueva versión de schema). El script de migración deberá respetar esto.

---

## 7. RESUMEN EJECUTIVO

| Fase | Archivos tocados | Líneas aprox. |
|---|---|---|
| A (SQL fix) | `data/sql/02_modulos_adicionales.sql` | 1 línea |
| B (front + Go) | `templates/index.html`, `templates/historial.html`, `handler.go`, `app.go` | ~90 líneas |
| C (docs) | `docs/doc.md`, `docs/funciones.md`, `docs/decisiones.md`, `docs/ai-context.md` | ~150 líneas añadidas |
| D (push) | — | git push |

**Total**: 3 commits, ~240 líneas, 1 push. Cero archivos eliminados (los cosméticos quedan para otro PR). Bug crítico del trigger resuelto. Multi-módulo operativo. Docs coherentes. Historial muestra `notas`.

---

## 8. ANTI-ALUCINACIÓN (verificación cruzada para el agente que ejecuta)

> Esta sección existe para reducir el riesgo de que el modelo que ejecute este plan (p.ej. DeepSeek V4 Flash) invente archivos, lineas o comportamientos. **Antes de cada edit, Siege verificar con el tool correspondiente (Read/Grep/Bash)**.

### 8.1 Archivos que SÍ existen (verificar con `ls`/Read antes de editar)

- `/home/user/Documentos/proyecto/baseaccess/data/sql/02_modulos_adicionales.sql` (827 lineas, commit committed)
- `/home/user/Documentos/proyecto/baseaccess/data/sql/01_master_control_docs_presidencia.sql` (398 lineas)
- `/home/user/Documentos/proyecto/baseaccess/app.go` (731 lineas, modified)
- `/home/user/Documentos/proyecto/baseaccess/handler.go` (728 lineas, modified)
- `/home/user/Documentos/proyecto/baseaccess/templates/index.html` (684 lineas, modified)
- `/home/user/Documentos/proyecto/baseaccess/templates/historial.html` (28 lineas)
- `/home/user/Documentos/proyecto/baseaccess/docs/doc.md` (271 lineas)
- `/home/user/Documentos/proyecto/baseaccess/docs/funciones.md` (148 lineas)
- `/home/user/Documentos/proyecto/baseaccess/docs/decisiones.md` (172 lineas)
- `/home/user/Documentos/proyecto/baseaccess/docs/ai-context.md` (38 lineas)

### 8.2 Archivos que NO existen o NO se tocan

- `templates/formulario.html` — existe pero NO se elimina fisicamente en este plan (su referencia en `index.html:176` SI se quita — eso es todo).
- `templates/tabla_filas.html` — existe pero NO se toca en este plan (cosmético fuera de scope).
- `data/sql/Tablas8.sql` — legado, NO se toca.
- `data/importar_datos.py` — NO se toca en este plan (el user lo actualizara aparte).
- `main.go`, `go.mod`, `wails.json` — NO se tocan.
- `frontend/wailsjs/go/main/App.d.ts`, `App.js` — NO se editan a mano (se regeneran con `wails dev`/`wails build`).
- `frontend/index.html` (legacy estático) — NO se toca.
- `data/expedientes.db` — NO se toca (gitignored).

### 8.3 Lineas exactas a editar (anchors verificadas)

| Archivo | Linea | Cadena exacta a buscar | Acción |
|---|---|---|---|
| `data/sql/02_modulos_adicionales.sql` | 71 | `VALUES (NEW.id_requisicion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.documento, NEW.serial_equipo, NEW.pase_sicesma, NEW.id_estatus, NEW.observaciones_entrega, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);` | Reemplazar `NEW.documento` por `NEW.descripcion_materiales` (solo esa ocurrencia en linea 71). |
| `templates/index.html` | ~176 | `<form id="form-expediente" class="space-y-6" onsubmit="return false;">\n    {{template "formulario.html" .}}\n</form>` | Eliminar la linea `{{template "formulario.html" .}}`. |
| `templates/index.html` | ~46-50 | bloque `window.PAGE_DATA = { ... }` | Sustituir por bloque con `modulos: {{jsonEncode .Modulos}},` y `filas: {{jsonEncode .Filas}},` (eliminando `expedientes: {{jsonEncode .Expedientes}}`). |
| `templates/index.html` | 345-350 | `function mostrarFormulario(id) { const modal = $('form-modal'); $('form-titulo').textContent = id ? 'Editar Expediente #' + id : 'Nuevo Expediente'; modal.classList.remove('hidden'); document.body.style.overflow = 'hidden'; }` | Sustituir por version dinamica que lee `window.PAGE_DATA.modulos[moduloKey].Nombre`. |
| `handler.go` | 562-587 (`handleHistorial`) | bloque `rows, err := h.app.ObtenerHistorialFila(modulo, id)` … `h.tmpl.ExecuteTemplate(w, "historial.html", rows)` | Cambiar para pasar `struct { Rows []Row; ActiveModule string }`. |
| `handler.go` | 675 | `data, err := h.app.ObtenerExpedientes("id_expediente DESC")` | Cambiar `ObtenerExpedientes` por `ObtenerFilas("expedientes", ...)`. |
| `app.go` | 587-607 | bloque `// --- legacy wrapper functions for backward compatibility ---` … `func (a *App) ObtenerHistorialCompleto(id int) ([]Row, error) { return a.ObtenerHistorialFila("expedientes", id) }` | Eliminar el bloque entero (5 funciones + comentario). |
| `templates/historial.html` | 1-28 (archivo completo) | plantilla actual con `{{if eq (len .) 0}}` … `{{range .}}` … `{{end}}` | Reescribir según §B.3 con `struct {Rows, ActiveModule}`, condicional Receptor, nueva columna Notas. |

**Regla de oro para el agente que ejecuta**: si una busqueda con `grep`/Read no encuentra la cadena exacta listada arriba, **NO inventar** el edit. Detenerse y reportar el desajuste al usuario. Es preferible abortar una fase que editar una linea distinta y romper el razon de gemini.

### 8.4 Verificaciones de compilación/sintaxis después de los edits

1. `go vet ./...` — debe dar cero warnings.
2. `go build ./...` — debe compilar sin errores (producira un binario `baseaccess` en la raiz — borrarlo despues, esta gitignored si o no `.gitignore` no lo cubre; usar `git status` para confirmar que no se agregara al commit).
3. Verificar con `git diff --stat` que los archivos tocados son solo los listados en cada commit. Si aparece cualquier otro archivo (p.ej. `app.go` en Commit 1, o `wailsjs/*` en Commit 2 sin runs previos de `wails dev`), abortar y reportar.

### 8.5 Verificación post-push (smoke mínimo sin abrir la app)

```bash
# 1. Confirmar 3 commits nuevos en wails-migration locales
git log --oneline -5

# 2. Confirmar que remote los recibio
git log --oneline origin/wails-migration -5

# 3. Confirmar que [skip ci] esta presente (para que GitHub Actions no corra)
git log -3 --format="%H %s%n%b" | rg -i 'skip ci'
# Debe devolver 3 lineas con '[skip ci]'

# 4. Verificar el diff final de los 3 commits
git diff origin/wails-migration~3 origin/wails-migration --stat
# Debe listar: app.go, handler.go, templates/index.html, templates/historial.html,
#              data/sql/02_modulos_adicionales.sql,
#              docs/doc.md, docs/funciones.md, docs/decisiones.md, docs/ai-context.md
# (9 archivos maximo)
```

### 8.6 Anti-alucinaciones específicas para DeepSeek V4 Flash

Modelos con contexto grande pero reasoning medio-bajo tienden a:

1. **Inventar archivos** que suenan plausible pero no existen. **Mitigación**: §8.1 lista los unicos archivos validos. Antes de tocar cualquiera, ejecutar `ls <ruta>` y verificar que devuelva el archivo.
2. **Mover lineas que no existen** en el archivo (`Edit` con `oldString` fabricado). **Mitigación**: §8.3 da los anchors exactos. Si `grep` no los encuentra, abortar.
3. **Agregar imports o constantes innecesarios** (p.ej. `import "strings"` si ya esta). **Mitigación**: NO agregar imports. Todos los edits se hacen con funciones/constantes ya existentes (`jsonEncode`, `rowGetStr`, `default`, `eq`, `ne`, `range`).
4. **Crear templates nuevos** (p.ej. `historial_<key>.html` por modulo). **Mitigación**: SOLO se reescribe `historial.html`. NO crear ningun template nuevo en este plan.
5. **Eliminar archivos legacy** (`formulario.html`, `tabla_filas.html`). **Mitigación**: NO borrar archivos en este plan, solo se quita la referencia `{{template "formulario.html" .}}` de `index.html`.
6. **Tocar el schema SQL** mas alla de la linea 71 de `02_modulos_adicionales.sql`. **Mitigación**: el UNICO edit SQL es esa unica linea. No agregar `DROP IF EXISTS`, no tocar `01_master`, no modificar triggers otros que `trg_req_mat_auditoria`.
7. **Regenerar bindings wailsjs a mano**. **Mitigación**: NO editar `frontend/wailsjs/*`. Se regeneran solos en el proximo `wails dev`/`wails build`.
8. **Agregar el `بیان` script de migracion de datos**. **Mitigación**: NO tocar `data/importar_datos.py`. El usuario lo actualizara aparte.
9. **Hacer push con `--force`**. **Mitigación**: NUNCA usar `--force`. Si el push falla por fast-forward, hacer `git pull --rebase origin wails-migration` primero.
10. **Commits mezclando fases**. **Mitigación**: 3 commits separados, cada uno con su `git add` explicito listado en §A.2, §B.6, §C.5. NO usar `git add -A`. NO usar `git add .`.

### 8.7 Comandos exactos a NO ejecutar

- ❌ `git add -A` o `git add .` (commitearia archivos no deseados como `data/expedientes.db` si no estuviera gitignored, o `combined.txt` si se corrio `make combine`).
- ❌ `wails dev` o `wails build` durante el plan (lo hara el user al final).
- ❌ `make combine` (no se quiere regenerar `combined.txt`).
- ❌ `git push --force` o `git push -f`.
- ❌ `rm templates/formulario.html` o `rm templates/tabla_filas.html` (cosmetico fuera de scope).
- ❌ `sqlite3 data/expedientes.db ...` (no se toca la BD).

### 8.8 Confirmaciones de estado pre-ejecución

El agente que ejecuta debe, antes de tocar nada, correr:

```bash
cd /home/user/Documentos/proyecto/baseaccess
git branch --show-current          # debe imprimir: wails-migration
git status --short                 # debe mostrar:
                                    #  M app.go
                                    #  M frontend/wailsjs/go/main/App.d.ts
                                    #  M frontend/wailsjs/go/main/App.js
                                    #  M handler.go
                                    #  M templates/index.html
                                    # ?? templates/form_aprobacion_jd.html
                                    # ?? templates/form_certificacion_bdu.html
                                    # ?? templates/form_expedientes.html
                                    # ?? templates/form_memorandums.html
                                    # ?? templates/form_recobros.html
                                    # ?? templates/form_reposos_medicos.html
                                    # ?? templates/form_requisiciones.html
                                    # ?? templates/form_vacaciones.html
                                    # ?? templates/form_valuaciones.html
                                    # ?? templates/tabla_aprobacion_jd.html
                                    # ?? templates/tabla_certificacion_bdu.html
                                    # ?? templates/tabla_expedientes.html
                                    # ?? templates/tabla_memorandums.html
                                    # ?? templates/tabla_recobros.html
                                    # ?? templates/tabla_reposos_medicos.html
                                    # ?? templates/tabla_requisiciones.html
                                    # ?? templates/tabla_vacaciones.html
                                    # ?? templates/tabla_valuaciones.html
```

Si el `git status` NO coincide con lo esperado (p.ej. hay mas archivos modified, o `plan.md` aparece como `??` porque se acaba de crear — eso es OK, `plan.md` se commitea en Commit 3 junto con docs), **abortar** y reportar al usuario.

**Importante sobre `plan.md`**: este archivo (`plan.md`) se crea como artefacto del planeo. **Se commitea junto con los docs en Commit 3**. Si se prefiere excluirlo, en Commit 3 hacer `git add docs/` en vez de listar archivos uno por uno, y dejar `plan.md` sin commitear. **Recomendación**: commitear `plan.md` tambien en Commit 3 (añadir `git add plan.md` al listado §C.5), porque documenta el racional de los cambios para futuros maintainers.
