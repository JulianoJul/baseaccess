# Plan de Correcciones — baseaccess

Revisión exhaustiva del código (4 frentes: `app.go`/`main.go`, `handler.go`, templates HTML/JS, SQL). Hallazgos verificados contra el código real; dos confirmados empíricamente (driver sqlite3 v1.14.22 y jsdom). No se incluye nada de `docs/falsos_positivos.md`.

**Leyenda:** 🔴 P0 crítico · 🟠 P1 lógico · 🟡 P2 DRY · 🟣 P3 hardcode · ⚪ P4 menor

---

## 🔴 P0 — Bugs críticos (funcionalidad rota hoy)

### P0-1. Gantt: las keys del timeline nunca coinciden con las columnas — el cronograma renderiza vacío
- **Refs:** `app.go:817` (scan), `app.go:836` (map key) vs `app.go:880` (`date_str`), `templates/ruta_procesos.html:168`
- **Categoría:** LOGIC
- **Problema:** `ruta_procesos_cronograma.fecha` es `DATE` (`03_ruta_procesos.sql:14`). El driver convierte DATE a `time.Time`, y al escanear a `string` queda como RFC3339 (`"2026-06-01T00:00:00Z"` — verificado empíricamente). Pero `buildGanttColumns` genera `date_str` como `"2026-06-01"` y el frontend busca `p.timeline[col.date_str]`. Las keys jamás matchean → **todas las celdas del Gantt caen al `|| {}` y se ven vacías**. Los 95 registros migrados del cronograma son invisibles.
- **Fix:** `SELECT strftime('%Y-%m-%d', c.fecha) AS fecha ...` en la query, o escanear a `time.Time` y usar `t.Format("2006-01-02")` como key.

### P0-2. `convertirMoneda()` lee `dataset.raw` desactualizado — conversión USD/Bs un keystroke atrás
- **Refs:** `index.html:1078-1115` (esp. 1088) + `_initNumInput` `index.html:1054-1062` + `form_expedientes.html:83-85,151-152`, `form_aprobacion_jd.html:56-58`
- **Categoría:** LOGIC
- **Problema:** Los inputs llaman `oninput="convertirMoneda(...)"` (handler inline, registrado en parse-time), pero `dataset.raw` se actualiza en un `addEventListener('input')` que se registra después (en `htmx:afterSettle`). Verificado con jsdom: el inline dispara **antes** → `getRaw()` usa el valor del keystroke anterior. El último dígito nunca se convierte, y **al guardar se persiste el Bs/USD desactualizado** (tecleas "12" USD → se guarda la conversión de "1").
- **Fix:** Que `getRaw()` calcule desde `el.value` al momento de la llamada (misma limpieza del listener), o mover la invocación de `convertirMoneda` dentro del listener de `_initNumInput` tras actualizar `raw`.

### P0-3. Doble fila de historial en cada INSERT de expedientes
- **Refs:** `01_master_control_docs_presidencia.sql:187-195` (dentro de `trg_exp_snapshot_inicial`)
- **Categoría:** LOGIC
- **Problema:** Verificado: un INSERT produce **2 filas** en `historial_movimientos` — el snapshot, y otra porque el `UPDATE expedientes SET id_estatus=...` interno del trigger dispara `trg_exp_auditoria`. `recursive_triggers=OFF` solo bloquea auto-recursión; disparar un trigger *distinto* sí está permitido. Toda creación queda doble-registrada en la auditoría. *(Bug introducido en el fix reciente del trigger INSERT.)*
- **Fix:** Computar el estatus corregido dentro del propio INSERT del snapshot (un `CASE` sobre `NEW.fecha_firma_contrato`) en vez de hacer UPDATEs que disparan el otro trigger; o hacer el UPDATE solo si el estatus realmente cambia.

### P0-4. Filas con `gerencia` vacía se ven en pantalla pero desaparecen de CSV/Excel
- **Refs:** `handler.go:884-889` (tabla) vs `handler.go:955-959` (CSV) y `handler.go:1118-1122` (Excel)
- **Categoría:** LOGIC
- **Problema:** `filtrarPorGerencias` conserva filas con `gerencia == ""` (todas las vistas usan `LEFT JOIN cat_gerencia`, así que es común). Los handlers de exportación re-implementan el filtro **sin** esa cláusula → esas filas se excluyen silenciosamente. El usuario ve N filas, exporta y obtiene menos sin aviso — pérdida silenciosa en la función de reporte principal.
- **Fix:** Unificar en un solo helper con una sola semántica (`gerName == "" || permitidasNames[gerName]`), usado por los 3 call sites.

### P0-5. Falta FK en la relación principal de `ruta_procesos_cronograma`
- **Refs:** `03_ruta_procesos.sql:12` (hay FK en `id_leyenda`/`id_expediente` :17-18, pero no en `id_proceso`)
- **Categoría:** LOGIC
- **Problema:** Verificado: `INSERT ... id_proceso=99999` (inexistente) es aceptado con `foreign_keys=ON`. Las filas huérfanas son invisibles en el Gantt (se descartan en `procMap`, `app.go:823`). Como ningún código Go inserta en esta tabla (todo es SQL manual), el schema es la única defensa.
- **Fix:** `CONSTRAINT fk_cron_proc FOREIGN KEY (id_proceso) REFERENCES ruta_procesos_procesos(id) ON DELETE CASCADE`.

### P0-6. Falta `UNIQUE(id_proceso, fecha)` en cronograma
- **Refs:** `03_ruta_procesos.sql:10-19`; colapso en `app.go:836` (`p.Timeline[fecha] = ...` last-wins)
- **Categoría:** LOGIC
- **Problema:** Verificado: duplicados `(id_proceso, fecha)` insertan sin error. La query de lectura no tiene `ORDER BY` y el map colapsa con last-wins → un día duplicado muestra estatus no determinístico.
- **Fix:** `UNIQUE(id_proceso, fecha)` + usar `INSERT ... ON CONFLICT DO UPDATE` para escrituras de celda.

---

## 🟠 P1 — Errores de lógica

### P1-1. Renombrar catálogo de estatus rompe los triggers → `id_estatus = NULL` silencioso
- **Refs:** `01_master_control_docs_presidencia.sql:188, 193, 231, 237`
- **Categoría:** HARDCODE
- **Problema:** Los triggers resuelven `'PENDIENTE'`/`'FIRMADO'` por subquery de `nombre`. Verificado: tras renombrar la fila del catálogo, setear `fecha_firma_contrato` deja `id_estatus = NULL` — corrupción silenciosa (la FK no atrapa NULLs).
- **Fix:** Usar los IDs seed literales (`1`/`2`) con comentario, o guard que impida renombrar/borrar esas dos filas.

### P1-2. `calcularSumas()` — mismo bug de lectura stale que convertirMoneda
- **Refs:** `index.html:1006, 1011-1020, 1036-1037`
- **Categoría:** LOGIC
- **Problema:** El `oninput` inline dispara antes de que `_initNumInput` actualice `dataset.raw`. El primer keystroke es correcto por casualidad (fallback a `input.value`); del segundo en adelante el total usa el raw anterior.
- **Fix:** Mismo que P0-2: `_rawNum` debe recalcular desde `input.value` siempre.

### P1-3. Gantt: dos filas de header con contenido idéntico
- **Refs:** `ruta_procesos.html:153-154`
- **Categoría:** LOGIC
- **Problema:** `weekHeaders` y `weekSubs` emiten ambos `esc(c.date_str)` — dos `<tr>` con el mismo contenido. El CSS define 3 estilos distintos (`styles.css:300-321`), evidencia de que el subheader debía mostrar otra cosa. Copy-paste bug.
- **Fix:** Eliminar `weekSubs` (y ajustar `rowspan="4"`→`"3"` en :160) o renderizar contenido distinto (rango "13/07 – 17/07").

### P1-4. Modal de exportación bypasea `MODAL_STACK` — no cierra con backdrop
- **Refs:** `index.html:812-820` vs `index.html:437-460`
- **Categoría:** LOGIC
- **Problema:** `abrirModalExportar`/`cerrarModalExportar` manipulan clases directamente en vez de usar `pushModal`/`cerrarModal` como los otros 7 modales. El handler de click-fuera solo cierra el tope del stack → el export-modal nunca cierra con backdrop.
- **Fix:** Usar `pushModal('export-modal')` / `cerrarModal('export-modal')`.

### P1-5. Título del form usa el módulo de la tabla visible, no del form cargado
- **Refs:** `index.html:463-470` + `index.html:804-809`
- **Categoría:** LOGIC
- **Problema:** Al abrir un fijado de otro módulo desde Fijados, el form correcto se carga pero el título se deriva de `active-module-val` (la tabla visible) → "Editar Vacaciones #5" sobre un form de expedientes. El guardado es correcto; solo el título miente.
- **Fix:** Pasar el módulo como parámetro a `mostrarFormulario(id, modulo)` desde `hxGetFormulario` y los `hx-on`.

### P1-6. Semánticas de fallo opuestas ante error de catálogos: pantalla fail-open, export fail-closed
- **Refs:** `handler.go:859-867` vs `handler.go:935-962` / `1098-1125`
- **Categoría:** LOGIC
- **Problema:** Si `ObtenerCatalogos()` falla, `filtrarPorGerencias` devuelve **todas** las filas, pero los exports dejan `permitidasNames` vacío → descartan **todas** las filas y muestran el error engañoso "no hay datos con los filtros aplicados" (el error real `cerr` se traga en :937).
- **Fix:** Propagar el error de catálogos (500 con causa real) en ambos caminos. Cae gratis al extraer el helper de P2-1.

### P1-7. Filtros de export inaplicables ignorados silenciosamente → export sobre-inclusivo
- **Refs:** `handler.go:984-991` y `1147-1153`
- **Categoría:** LOGIC
- **Problema:** `if expectedName == "" { continue }` (ID de filtro no existe en catálogo) y `if !exists { continue }` (el módulo no tiene esa columna, ej. `id_empresa` en reposos_medicos) ignoran el filtro en vez de rechazarlo. El usuario cree que filtró; el export contiene todo.
- **Fix:** Tratar filtro inaplicable como no-match (`match=false; break`) o validar params upfront con 400 nombrando el filtro.

### P1-8. `ExpedientesDisponiblesRuta`: expedientes con `descripcion_proceso` NULL desaparecen sin log
- **Refs:** `app.go:934-939`
- **Categoría:** LOGIC
- **Problema:** La vista emite `descripcion_proceso` cruda (nullable); scan a `string` falla → `continue` **sin log** (a diferencia de :758, :785, :820 que sí loguean). Esos expedientes desaparecen del selector "agregar proceso".
- **Fix:** `sql.NullString` para `desc` (o `COALESCE(e.descripcion_proceso, '')` en la query) + `log.Printf` en la rama de error.

### P1-9. `sanitizarOrden`: whitelist valida columnas de tabla, pero ORDER BY corre contra la vista (latente)
- **Refs:** `app.go:439-465`, llamado desde `app.go:476-477`
- **Categoría:** LOGIC (latente)
- **Problema:** ~11 columnas whitelisted (`id_gerencia`, `id_estatus`…) no existen en `vw_reporte_excel_contrataciones` → error SQL si se usan; columnas útiles de la vista (`gerencia`, `emisor`) son rechazadas → fallback silencioso al sort default. Hoy inalcanzable (la UI solo envía `fecha_creacion`/`fecha_actualizacion`).
- **Fix:** Whitelist de columnas reales de la vista (vía `ObtenerColumnasVista`) o lista sortable explícita por módulo.

### P1-10. `GuardarNuevoCatalogo` acepta `nombre` vacío/whitespace
- **Refs:** `app.go:1054-1094`; el handler tampoco valida (`handler.go:815-829`)
- **Categoría:** LOGIC
- **Problema:** Sin `TrimSpace`/empty check antes del INSERT. El DDL es `nombre TEXT UNIQUE` sin `NOT NULL` → entradas vacías en los dropdowns.
- **Fix:** `nombre = strings.TrimSpace(nombre)`; error si vacío.

### P1-11. Triggers de auditoría insertan snapshot en UPDATEs no-op
- **Refs:** `01:197-245`; `02:68, 159, 258, 381, 504, 611, 709, 797`
- **Categoría:** LOGIC
- **Problema:** Verificado: un UPDATE que setea una columna a su propio valor igual agrega fila de historial. Guardar-sin-cambios infla las 9 tablas hist.
- **Fix:** Cláusula `WHEN` comparando OLD vs NEW en las columnas rastreadas.

### P1-12. `documento` como TEXT libre en 5 módulos vs FK en 4
- **Refs:** TEXT: `02:114, 201, 306, 658, 754`; FK `id_documento`: `01:36/70`, `02:15/33`, `02:439/463`, `02:556/575`
- **Categoría:** LOGIC / CODE
- **Problema:** Mismo concepto de negocio, dos modelos: texto libre sin integridad de catálogo (typos, duplicados) y rompe el patrón uniforme de módulos (`documento` vs `id_documento`).
- **Fix:** Migrar los 5 TEXT a `id_documento INTEGER REFERENCES cat_documento(id)`, o documentar por qué esos módulos necesitan texto libre.

### P1-13. `dir` en minúscula se fuerza a DESC (sort invertido)
- **Refs:** `handler.go:583-585`
- **Categoría:** LOGIC
- **Problema:** `?dir=asc` es tratado como inválido y forzado a `DESC` — produce el sort opuesto al pedido. (`sanitizarOrden` en app.go sí hace uppercase; este handler pre-valida sin hacerlo.)
- **Fix:** `dir = strings.ToUpper(dir)` antes del check (o dejar que `sanitizarOrden` normalice).

### P1-14. `handleHistorial` no valida `modulo` a nivel handler
- **Refs:** `handler.go:676-692`
- **Categoría:** LOGIC
- **Problema:** Todos los demás handlers validan y devuelven 400; aquí un módulo inválido cae a `ObtenerHistorialFila` y sale como 500 "modulo no soportado".
- **Fix:** Usar el helper compartido `moduloDesdeRequest` (P2-2) y devolver 400.

### P1-15. `db_id` parse error tragado en `handleAgregarRutaProceso`
- **Refs:** `handler.go:765-769`
- **Categoría:** LOGIC
- **Problema:** `dbID, _ = strconv.Atoi(dbIDStr)` — un `db_id` malformado se vuelve `0` silenciosamente, vinculando el proceso al expediente 0.
- **Fix:** `writeJSONError(w, "db_id invalido", 400)` en parse error.

### P1-16. `preparePageData` traga error de `ObtenerFilas` → tabla vacía silenciosa
- **Refs:** `handler.go:325-329`
- **Categoría:** LOGIC
- **Problema:** Un error de BD a mitad de sesión (archivo movido/bloqueado) solo se loguea; la página renderiza con tabla vacía, indistinguible de "sin registros". Handlers hermanos devuelven 500 por el mismo fallo.
- **Fix:** Superficiar el error (fragmento de error o `http.Error`).

### P1-17. `ObtenerCatalogos`: error de iteración logueado pero retorna éxito con datos parciales
- **Refs:** `app.go:1033-1035`
- **Categoría:** LOGIC
- **Problema:** Un fallo mid-iteration produce catálogo truncado reportado como éxito, inconsistente con errores de scan que sí abortan (`app.go:1030-1032`).
- **Fix:** Retornar el error (o al menos dropear la key).

### P1-18. `Modulos` crudo (con `QueryHistorial`) pasado a templates, bypass del stripping deliberado
- **Refs:** stripping en `handler.go:284-289`; pase crudo en `handler.go:628` y `:663`
- **Categoría:** LOGIC (latente)
- **Problema:** `preparePageData` blanquea `QueryHistorial` antes de exponer `Modulos` (index.html:64 lo serializa a JS). `handleFiltrarExpedientes` y `handleCambiarModulo` pasan el map global con el SQL completo. Hoy `tabla_*.html` no serializan `.Modulos` (latente), pero un `jsonEncode .Modulos` copy-pasteado filtraría todo el SQL de historial al cliente.
- **Fix:** Extraer el stripping a `func modulosSinQueries() map[string]ModuloConfig` y usarlo en los 3 handlers.

### P1-19. Pines identificados solo por id (colisión cross-módulo latente)
- **Refs:** `index.html:603, 608, 656`
- **Categoría:** LOGIC (latente)
- **Problema:** Los fijados guardan `{id, solped, modulo}` pero toggle/unpin comparan solo `id`. `id_expediente=5` e `id_recobro=5` colisionarían. Latente hoy (solo `tabla_expedientes.html:36` renderiza pins), pero el código ya está diseñado multi-módulo.
- **Fix:** Comparar `(id, modulo)` y pasar `modulo` en el unpin del modal.

### P1-20. Scripts 01/02 no idempotentes vs 03 idempotente
- **Refs:** CREATEs sin `IF NOT EXISTS` en 01/02; `IF NOT EXISTS` + `INSERT OR IGNORE` en `03:4,10,21,30`
- **Categoría:** CODE
- **Problema:** Re-correr 01 o 02 sobre una BD existente falla en el primer statement; 03 es re-ejecutable. Robustez de deployment inconsistente para scripts ejecutados manualmente.
- **Fix:** `IF NOT EXISTS` en todos los CREATEs de 01/02.

---

## 🟡 P2 — Violaciones DRY

### P2-1. Pipeline de filtros de export duplicado ~120 líneas entre `handleCSV` y `handleExportarExcel`
- **Refs:** `handler.go:894-1003` ≡ `1048-1166`; header-ordering `1014-1025` ≡ `1173-1184`
- **Categoría:** DRY
- **Problema:** Módulo default+validación, parseo fechas, captura `id_*`, `filterColMap` (12 entradas), indexado de catálogos, filtro GerenciasIDs, loop fecha+catálogo, y ordenación ID-first — todo copia verbatim. Esta divergencia ya produjo P0-4 (gerencia vacía) y la deriva de formato de errores (P4-2).
- **Fix:**
  ```go
  var exportFilterColMap = map[string][2]string{ /* única fuente */ }
  func filtrosIDDesdeQuery(q url.Values) map[string]string
  func indexarCatalogos(cs map[string][]CatalogoItem) map[string]map[string]string
  func (h *TemplateHandler) filasParaExportar(r *http.Request) (ModuloConfig, []Row, error)
  func columnasOrdenadas(r Row, idCol string) []string
  ```
  Los handlers conservan solo la serialización. Además, considerar eliminar `/api/csv` — **no tiene caller en la UI** (verificado: solo `/api/exportar-excel` se invoca, index.html:947), borrarlo elimina la mitad de la duplicación.

### P2-2. Extracción+default+validación de módulo copy-pasteada en 11 lugares
- **Refs:** `handler.go:279-282, 317-321, 430-438, 474-482, 504-511, 569-577, 640-648, 676-679, 894-902, 1048-1056, 1274-1282`
- **Categoría:** DRY
- **Problema:** Las copias ya divergieron: `handleHistorial` omite validación (P1-14) y el formato de error alterna JSON/texto plano (P4-2).
- **Fix:** `func moduloDesdeRequest(r *http.Request) (string, ModuloConfig, bool)`; cada handler queda en 3 líneas.

### P2-3. 9 forms ≈72% similitud, 9 tablas ≈71% (medido con difflib)
- **Refs:** `form_*.html` (966 líneas no-blancas), `tabla_*.html` (959 líneas)
- **Categoría:** DRY
- **Problema (medido):** forms: 72% pairwise promedio, **392 líneas presentes en ≥8 de 9 archivos**; tablas: 71%, **523 líneas en ≥8 de 9**. Bloques idénticos ×9: footer Cancelar/Eliminar/Guardar, fieldset Observaciones, selects Gerencia/Superintendencia, fieldset Trazabilidad (×6-8), wrapper tabla + columnas Ver/Acción, fila de estado vacío. JS: `cargarSuperintendencias` vs `filtrarSuperintendenciasExportar` = **98.4% idénticas** (`index.html:479-505` vs `898-925`); `esc()` duplicada (`index.html:353`, `ruta_procesos.html:33`).
- **Fix (Go templates):**
  1. `{{define "form_footer"}}` con `dict "modulo" . "idcol" . "registro" .` — elimina además los `?modulo=X` hardcodeados.
  2. `{{define "select_catalogo"}}` — resuelve ~40 selects.
  3. `{{define "fieldset_trazabilidad"}}` y `{{define "fieldset_obs"}}`.
  4. `{{define "tabla_shell"}}` con `{{block "columnas"}}`/`{{block "detalle"}}`.
  5. JS: una sola `filtrarSelectDependiente(gerEl, selEl, cache)`.
  Estimado: cada form/tabla queda en ~25-40 líneas específicas (~-60%).

### P2-4. Diez `QueryHistorial` casi idénticos
- **Refs:** `app.go:56, 71, 85, 100, 117, 133, 148, 162, 176`
- **Categoría:** DRY
- **Problema:** Mismo skeleton `COALESCE(x.nombre,'-') + LEFT JOIN cat_* ×5 + WHERE h.<id>=? ORDER BY h.id_movimiento DESC`, variando solo nombres. Un cambio cross-cutting requiere editar 10 strings en sync perfecto.
- **Fix:** Generar la query desde un spec por módulo (columnas planas + joins de catálogo) con un query-builder compartido.

### P2-5. Boilerplate de módulos en `02_modulos_adicionales.sql` (~62% del archivo)
- **Refs:** triggers `02:62-73…791-802`; hist tables `02:36-60…770-789`; views `02:75-100…804-824`
- **Categoría:** DRY (medido)
- **Problema:** 16 triggers = 88 líneas (un solo template estructural cubre los 16 tras normalizar identificadores — verificado programáticamente); 8 hist tables ≈210 líneas (espejo 1:1 del padre); 8 views ≈218 líneas con el mismo join skeleton. **~516 de 827 líneas son boilerplate mecánico**; solo ~200 líneas de columnas base llevan información real por módulo.
- **Fix:** SQLite no tiene macros. Opciones: generar `02_*.sql` desde un manifiesto de módulos (Python / Go `text/template`), o aceptar la duplicación y agregar un lint de simetría.

### P2-6. Boilerplate de transacciones duplicado
- **Refs:** `app.go:625-634, 657-661` (`EliminarFila`) ≡ `app.go:961-970, 977-981` (`EliminarRutaProceso`)
- **Categoría:** DRY
- **Fix:** Helper `withTx(func(tx *sql.Tx) error) error`.

### P2-7. Construcción de paths de backup y copia con checkpoint duplicados
- **Refs:** `app.go:302, 327, 332-333, 340` (patrón `.bak.N` ×4); `app.go:294-325` ≡ `355-373`
- **Categoría:** DRY
- **Fix:** `backupPath(i int) string` y `copyDBCheckpointed(dstPath string) error` compartidos.

---

## 🟣 P3 — Hardcode

### P3-1. `'expedientes'` como módulo default literal en 12+ lugares
- **Refs:** `handler.go:281, 319, 432, 476, 506, 571, 642, 678, 896, 1050, 1276`; `app.go:643`; `index.html:466, 607, 805`
- **Categoría:** HARDCODE
- **Fix:** `const moduloDefault = "expedientes"` usado por el helper de P2-2.

### P3-2. `'FIRMADO'` literal en query de pendientes
- **Refs:** `app.go:183`
- **Categoría:** HARDCODE
- **Problema:** El estatus es dato de catálogo; si se renombra en la BD, la vista de pendientes incluye firmados silenciosamente. Relacionado con P1-1 (mismo patrón en triggers).
- **Fix:** `const EstatusFirmado = "FIRMADO"` compartida, o resolver por ID de catálogo.

### P3-3. `CURRENT_TIMESTAMP` vs `CURRENT_DATE` inconsistente en `fecha_actualizacion`
- **Refs:** TIMESTAMP en los 8 triggers de `02` (:72, 163, 262, 385, 508, 615, 713, 801); DATE en expedientes (`01:243`); todos los DEFAULTs son `CURRENT_DATE`
- **Categoría:** HARDCODE
- **Problema:** Verificado: tras update, `memorandums.fecha_actualizacion = '2026-07-18 19:29:24'` vs `expedientes = '2026-07-18'` — dos formatos en la columna del mismo nombre, e incluso dentro de cada tabla (insert = date-only, update = datetime).
- **Fix:** Elegir uno (recomendado `CURRENT_TIMESTAMP`) en los 9 triggers y todos los DEFAULTs.

### P3-4. `id_estatus INTEGER DEFAULT 1` — magic number acoplado al seed en 9 tablas
- **Refs:** `01:49`; `02:19, 116, 209, 311, 448, 562, 663, 758`; seed `(1,'PENDIENTE')` en `01:371`
- **Categoría:** HARDCODE
- **Problema:** El default asume que seed id 1 = PENDIENTE; una reconstrucción del catálogo que cambie IDs hace que filas nuevas apunten al estatus equivocado. SQLite no permite DEFAULT desde subquery — solo es documentable.
- **Fix:** Comentario junto a cada DEFAULT declarando el contrato del seed, o enforcement vía trigger.

### P3-5. `row["fecha_recibido"]` como columna de filtro universal
- **Refs:** `handler.go:966, 1129`
- **Categoría:** HARDCODE
- **Problema:** Cualquier módulo futuro cuya vista carezca de esa columna devolverá "no hay datos" siempre que haya filtro de fecha (`fr` siempre `""` → `continue`).
- **Fix:** `const colFechaFiltro = "fecha_recibido"` o campo `FechaColumna string` en `ModuloConfig`.

### P3-6. Magic numbers en JS y duplicación de fuentes de verdad
- **Refs:** `index.html:687` (`pageSize = 10` — duplica `PageSize: 10` Go, `handler.go:295`), `:740` (`maxVisible = 7`), `:364` (toast 3000/300ms), `:472` (`setTimeout(...,50)`), `:528` (recientes cap 5), `:141` (delay:200ms); `ruta_procesos.html:55` (`substring(0,60)`); `pendientes.html:20` (`truncate 50`); `app.go:863` (`bizTarget := 60`); `app.go:225-229` (bounds 1/20 de backups); **76 inline `style="width: Npx"`** en las 9 tablas
- **Categoría:** HARDCODE
- **Fix:** `const CONFIG = {...}` centralizado, o inyectar vía `PAGE_DATA` los que ya existen Go-side (pageSize). Los widths → clases CSS.

### P3-7. Tres fuentes de verdad para el mapeo columna→catálogo
- **Refs:** `index.html:847-860` (`catalogKeys` JS) ≡ `handler.go:1082-1095` (`filterColMap` Go) ≡ `app.go:187-199` (`catalogosValidos` Go)
- **Categoría:** HARDCODE / DRY
- **Problema:** Agregar un módulo/catálogo requiere tocar 3 lugares.
- **Fix:** Servir el mapeo desde Go (parte de `PAGE_DATA` o del response de `/api/columnas-modulo`).

### P3-8. Placeholders COALESCE inconsistentes en la vista principal
- **Refs:** `01:253` (`'SIN_SOLPED'`), `01:276-277`
- **Categoría:** HARDCODE
- **Problema:** Tres convenciones (`SIN_SOLPED`, `NO POSEE`, `NO APLICA`) y `nro_contrato_sap` (:277) sin COALESCE mientras su hermano `nro_contrato_sicac` (:276) sí lo tiene — el Excel mezcla convenciones y muestra blanco para sap.
- **Fix:** Estandarizar un placeholder y aplicar (u omitir) COALESCE consistentemente en columnas hermanas.

### P3-9. Magic strings varios
- **Refs:** `app.go:263` (DSN params `?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000` — además rompe con `?` en el nombre de archivo); `app.go:771` (`'SIN SOLPED'`, `'N/A'`); `app.go:414, 880` (`"2006-01-02"` ×2 — y fue la causa raíz de P0-1)
- **Categoría:** HARDCODE
- **Fix:** `const fechaLayout = "2006-01-02"` usada en todo productor de keys; const para DSN; rechazar/escapar filenames con `?`.

### P3-10. Dos vocabularios de estatus paralelos sin mapeo
- **Refs:** `03:30-36` (leyenda Gantt) vs `01:370-380` (`cat_estatus_detalle`)
- **Categoría:** DRY / HARDCODE
- **Problema:** Se solapan (`PENDIENTE`, `FIRMADO`) pero divergen (`DEVUELTO` vs dos DEVUELTO*), sin FK ni mapeo — mismos nombres hardcodeados en dos seeds que pueden driftear.
- **Fix:** Documentar que son dominios separados, o derivar la leyenda del Gantt desde `cat_estatus_detalle`.

---

## ⚪ P4 — Menores / code errors

| # | Ref | Problema | Fix |
|---|-----|----------|-----|
| P4-1 | `handler.go:406-409, 562-565, 633-636, 669-672, 704-707, 737-740, 809-812` | `http.Error` después de `ExecuteTemplate` — doble write, 500 perdido si ya se escribió | Render a `bytes.Buffer`; headers+copy solo en éxito |
| P4-2 | `handler.go` (varios) | Formato de error inconsistente: JSON vs texto plano para fallos idénticos (`"modulo invalido"` JSON en :436 vs plain en :509; export vacío JSON en :1006 vs plain en :1169). El frontend hace `res.text()` → mostraría `{"success":false,...}` crudo | Una convención por familia de endpoints |
| P4-3 | `index.html:770` | Rama muerta: `evt.detail.target.id === 'tabla-cuerpo'` nunca se cumple (swaps van a `#vista-tabla`) | Eliminar la condición |
| P4-4 | `ruta_procesos.html:78` | Proceso recién agregado muestra "SIN SOLPED" y descripción redundante hasta recargar (el push usa el texto completo del `<option>`) | Guardar `solped`/`descripcion` como `data-*` en cada option |
| P4-5 | `index.html:1100` | Valores muertos `'usd_presup'`/`''` en `convertirMoneda` (ningún caller los pasa) | Simplificar a `origen === undefined` |
| P4-6 | `app.go:263` | DSN: `filePath+"?..."` rompe si el archivo contiene `?` en el nombre | Rechazar/escapar filenames con `?` (rel. P3-9) |
| P4-7 | SQL (varios) | Sin CHECK constraints: nada previene `fecha_devuelto < fecha_recibido`, `cantidad_dias <= 0`, montos negativos (`02:203-204, 322-323, 659-662, 756-757`, `01:50-51`) | `CHECK (fecha_hasta >= fecha_desde)`, `CHECK (cantidad_dias > 0)`, `CHECK (monto >= 0)` |
| P4-8 | `02:749-768` | `reposos_medicos` rompe el set común de columnas (sin `id_receptor`/`fecha_devuelto`; los otros 8 sí) | Alinear si el flujo de negocio también recibe/devuelve documentos |
| P4-9 | `03_ruta_procesos.sql` | Las 3 tablas ruta_procesos no tienen historial/trigger — `cronograma` es la única tabla mutada por el usuario sin audit trail | Considerar hist table si se requiere trazabilidad |
| P4-10 | `Tablas8.sql` (legacy) | Contradice el schema actual: trigger sin transition guard (fuerza FIRMADO en todo update), seeds viejos (10 gerencias vs 13 actuales). Reconstruir desde Tablas8 = comportamiento viejo buggado | Renombrar a `Tablas8.sql.legacy.txt` o eliminar |

---

## Matriz de completitud de módulos (verificada)

Los 9 módulos tienen: tabla + hist + AFTER INSERT + AFTER UPDATE + vista + índices. ✅

| Módulo | Tabla | Hist | Trg INSERT | Trg UPDATE | Vista | Índices |
|---|---|---|---|---|---|---|
| expedientes | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| req_materiales | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| memorandums | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| recobros | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| valuaciones | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| aprobacion_jd | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| certificacion_bdu | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| vacaciones | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| reposos_medicos | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## Orden sugerido de ejecución

1. **P0-1** (Gantt vacío) — una línea de SQL; recupera 95 registros invisibles
2. **P0-3** (doble historial) — revertir/rediseñar el UPDATE interno del trigger INSERT
3. **P0-2 + P1-2** (conversión stale) — mismo fix en dos funciones JS
4. **P0-4 + P1-6 + P1-7** (semántica de filtros) — sale gratis con **P2-1** (extracción del pipeline de export; evaluar eliminar `/api/csv` que no tiene caller)
5. **P0-5 + P0-6** (constraints cronograma) — dos líneas de DDL
6. **P1-1 + P3-2** (estatus por nombre) — decidir IDs literales vs guard
7. **P2-2 + P3-1 + P3-5** (helper de módulo + constantes) — fundación para el resto
8. **P2-3** (templates `{{define}}`) — refactor grande, hacer al final con todo lo demás estable
9. Resto P1/P3/P4 en cualquier orden
