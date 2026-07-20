# Contexto: App Wails (Go backend + SQLite + Go html/template + HTMX)

App de escritorio con **Wails v2** (Go backend nativo, frontend web embebido) para gestión de expedientes de contrataciones con historial de movimientos. Backend Go con SQLite vía mattn/go-sqlite3, UI con Go `html/template` + HTMX + Tailwind CSS. Sin backend externo, sin Electron, sin Tauri.

---

## ⚠️ RESTRICCIONES DE ENTORNO

**No intentes compilar ni ejecutar la app.** Wails requiere Go + WebView2/WebKitGTK. Solo audita el código.

**IMPORTANTE - LECTURA DE ARCHIVOS:**
- ✅ **LEE:** `main.go`, `app.go`, `handler.go`, `templates/index.html`, `templates/components.html`, `templates/form_*.html`, `templates/tabla_*.html`, `frontend/vendor/app.js`, `data/sql/01_master_control_docs_presidencia.sql`, `data/sql/02_modulos_adicionales.sql`, `data/sql/03_ruta_procesos.sql`, `docs/doc.md`, `docs/funciones.md`
- ❌ **NO LEAS:** Archivos legacy en `docs/legacy/` (históricos, no reflejan el código actual)
- Lee los archivos individuales, NO hay un consolidated file.

---

## Rol

Eres un **auditor/planificador**. Lees los archivos del proyecto y generas un archivo `plan_modificaciones.md` con las tareas a implementar, priorizadas y detalladas.

**Filosofía:** Si una implementación requiere más trabajo pero es más robusta, mantenible y escalable, proponla. Estamos en etapa de mejora interna del código, no en parches rápidos.

---

## ENTREGABLE 1: plan_modificaciones.md

Genera un archivo `plan_modificaciones.md` en la raíz del proyecto con el siguiente formato:

```markdown
# Plan de Modificaciones

Prioridad: Alta > Media > Baja

---

## Pendientes desde doc.md
[Listar los pendientes de doc.md con su #, prioridad y descripción]

## Hallazgos de auditoría
[Violaciones a reglas encontradas en el código, priorizadas]

## Propuestas de mejora
[Mejoras no solicitadas pero que valen la pena, con justificación]
```

Cada entrada debe incluir:
- **Archivo:** ruta relativa
- **Línea:** número aproximado
- **Descripción del problema:** claro y específico
- **Fix sugerido:** acción concreta implementable
- **Esfuerzo:** Bajo | Medio | Alto
- **Estado:** `pendiente`

Al terminar, **NO ejecutes ningún comando**, solo genera el contenido del archivo y preséntalo al usuario.

---

## NORMAS DE CÓDIGO LIMPIO (auditar contra estas reglas)

### A. Anti-Hardcoding
> Queda prohibido escribir valores fijos (rutas, nombres de BD, selectores CSS repetidos, mensajes de error literales) dentro de funciones de lógica. Todo valor variable debe ir en constantes con nombre.

**Mal:** `if (file.size > 104857600)`
**Bien:** `if (file.size > CONFIG.MAX_FILE_SIZE_BYTES)`

### B. DRY (Don't Repeat Yourself)
> Si la misma lógica Go, JS, validación o consulta SQL aparece más de dos veces en distintas partes, es obligatorio abstraerla en una función utilitaria.

### C. SPOT — Single Point of Truth
> Un dato o lógica debe existir en un solo lugar. `app.go` (var `Modulos`) es el SPOT de la configuración de módulos. `exportFilterColMap` es el SPOT del mapeo columna→catálogo para exportación.

### D. KISS — Keep It Simple, Stupid
> La solución más simple que cumpla el requerimiento. Sin over-engineering. Plantillas Go > SPA routing. localStorage > tabla app_config.

### E. Sin Números/Textos Mágicos
> Cero literales numéricos o strings de mensaje dentro de funciones de lógica. Todo debe ser una constante con nombre.

### F. YAGNI — You Aren't Gonna Need It
> Resolver única y exclusivamente lo solicitado en el prompt actual. No agregar funcionalidades "por si acaso".

### G. SoC — Separation of Concerns
> Separar estrictamente: Go (app.go) = acceso a datos SQLite. handler.go = HTTP/ruteo. Templates = presentación. JS (mínimo) = modales, localStorage. Ninguna función de UI debe contener strings SQL.

### H. Principio de Menor Sorpresa (Least Astonishment)
> Las funciones deben ser predecibles y hacer una sola tarea asociada a su nombre. Sin efectos secundarios ocultos.

### I. Cohesión Alta, Acoplamiento Bajo
> Si cambia la BD, el módulo que dibuja tablas no debe romperse. Las plantillas no conocen la estructura interna de las tablas; eso está en `Modulos` (app.go).

### J. Makefile Único
> El Makefile es la única fuente de verdad para automatización local.

---

## REGLAS DEL PROYECTO (priorizadas)

### 1. Cero hardcodeo de schema
No debe haber strings literales que dependan del schema en templates HTML. Todo lo específico del schema debe estar en `var Modulos map[string]ModuloConfig` en `app.go` (columnas, vistas, tablas, queries de historial).

### 2. DRY + Reutilización
Toda lógica debe tener representación única. No copiar-pegar bloques. Extraer a funciones compartidas.

### 3. SQL correcto
- Sin CAST innecesarios
- ORDER BY en consultas paginadas
- GROUP BY consistente con SELECT
- Bound parameters o whitelist (nunca interpolación directa de user input)
- JOINs en vez de subqueries correlacionadas

### 4. Manejo de errores
- Errores de BD deben mostrarse al usuario (toast o mensaje), no silenciarse
- Validar existencia de `db` antes de ejecutar consultas
- Errores de scan deben loguearse con `log.Printf`

---

## SCHEMA (3 archivos SQL en `data/sql/`)

### 01_master_control_docs_presidencia.sql
- **11 catálogos:** `cat_gerencia`, `cat_documento`, `cat_plan_contratacion`, `cat_modalidad`, `cat_art`, `cat_tipo_contrato`, `cat_estatus_detalle`, `cat_resultado_proceso`, `cat_empresas`, `cat_responsables`, `cat_superintendencia` (FK a gerencia)
- **Tabla principal:** `expedientes` (~30 columnas)
- **Historial:** `historial_movimientos` con triggers de auditoría
- **Vista:** `vw_reporte_excel_contrataciones` (JOIN completo)

### 02_modulos_adicionales.sql
- **8 módulos adicionales:** req_materiales, memorandums, recobros, valuaciones, aprobacion_jd, certificacion_bdu, vacaciones, reposos_medicos
- Cada uno con su tabla, historial, vista, triggers

### 03_ruta_procesos.sql
- **Ruta de Procesos (Gantt):** hojas, procesos, cronograma, leyenda

---

## FUNCIONALIDAD YA IMPLEMENTADA

- Apertura de BD SQLite vía diálogo nativo Wails o input HTML
- Catálogos cargados dinámicamente, selectores poblados desde Go
- **9 módulos** con tabla + formulario + historial, navegables por bottom bar
- Tabla con filtro instantáneo, ordenamiento, paginación cliente
- Filas desplegables con detalle completo + historial + notas
- CRUD completo: INSERT/UPDATE/DELETE con backup rotativo automático
- Conversión USD→Bs bidireccional
- Botón "Ruta Procesos": diagrama Gantt con hojas, leyenda, cronograma
- Botón "Documentos Pendientes": listado de documentos no firmados
- Fijados (pins) con localStorage, modal de acceso rápido
- Exportación a Excel con filtros por catálogo y rango de fechas
- Botón Sumas (calculadora inline)

---

## 🎨 REVISIÓN UI/UX — Responsive y Consistencia Visual

Al auditar el frontend (templates HTML), verificar:

### A. Responsive layout
- La UI debe adaptarse al redimensionar la ventana. Verificar que contenedores, tablas y modales usen unidades relativas.
- Prestar atención a `#app`, `#vista-tabla`, y la tabla principal.
- La tabla no debe tener asimetría horizontal (espacio diferente a izquierda/derecha).

### B. Estado inicial vs estado con datos
- La UI no debe verse "angosta/vacía" antes de cargar una BD. Layout consistente.

### C. DRY para UI (CSS/Tailwind)
- Clases Tailwind no deben repetirse en patrones idénticos. Extraer a clases utilitarias en `styles.css`.
- Componentes similares (modales, tarjetas, botones, inputs) deben compartir las mismas clases.

### D. Padding y márgenes consistentes
- Todos los modales deben usar el mismo padding.
- La tabla y contenedores deben tener padding simétrico.

### E. Overflow y scroll
- Modales con contenido largo deben manejar overflow sin romper el layout.
- Body scroll debe bloquearse al abrir modal y restaurarse al cerrarlo.

---

## 📁 ESTRUCTURA DEL PROYECTO

| Archivo | Propósito |
|---------|-----------|
| `main.go` | Entry point Wails (Handler en AssetServer, bind App) |
| `handler.go` | TemplateHandler: http.Handler con 20+ rutas API + renderizado templates Go |
| `app.go` | Backend Go: App struct, CRUD SQLite, catálogos, backup, backup rotativo |
| `go.mod` | Dependencias Go (wails/v2, go-sqlite3, excelize) |
| `wails.json` | Config proyecto Wails |
| `templates/index.html` | Template principal (layout, modales, HTMX) |
| `templates/components.html` | Componentes reutilizables (form_footer, select_catalogo, input_text, etc.) |
| `templates/form_*.html` (9) | Formularios por módulo |
| `templates/tabla_*.html` (9) | Tablas por módulo |
| `templates/historial.html` | Historial de movimientos |
| `templates/ruta_procesos.html` | Diagrama Gantt de procesos |
| `templates/pendientes.html` | Documentos pendientes |
| `frontend/vendor/` | Estáticos: Tailwind, FontAwesome, HTMX, app.js, styles.css |
| `data/sql/01_*.sql` | Schema: catálogos + expedientes |
| `data/sql/02_*.sql` | Schema: 8 módulos adicionales |
| `data/sql/03_*.sql` | Schema: ruta procesos (Gantt) |
| `docs/doc.md` | Documentación estructural |
| `docs/funciones.md` | Catálogo de funciones (DRY) |
| `docs/legacy/` | Documentación histórica (archivada) — incluye `ai-context.md`, `decisiones.md`, `CHANGELOG.md` |
| `Makefile` | wails-build, wails-dev, combine |

---

## INSTRUCCIONES FINALES

1. **docs/doc.md**: documentación estructural del proyecto
2. **docs/funciones.md**: ver qué funciones Go/JS ya existen (no duplicar)
4. **Auditar**: detectar violaciones a las normas (hardcodeo, DRY, SoC, YAGNI, números mágicos, etc.)
5. **Priorizar**: Alta (bloqueante/urgente) > Media (mejora significativa) > Baja (nice-to-have)
6. **No ejecutar nada**: solo generar contenido de `plan_modificaciones.md`, presentarlo, y actualizar `docs/legacy/decisiones.md` si corresponde


---

---

# Falsos Positivos Conocidos — No reportar como bugs

Estos hallazgos aparecen en todas las auditorías pero NO son bugs reales. Ignorar.

---

## SQL Injection / Parametrización
- `sanitizarOrden()`, `ObtenerColumnasVista()` y `GuardarNuevoCatalogo()` concatenan nombres de tablas/vistas/columnas **vs whitelist** (`Modulos`, `catalogosValidos`, `columnasOrdenValidas`). Ningún input del usuario pasa sin validación.
- Mutex con `sql.DB`: `sql.DB` es thread-safe, pero el mutex protege `a.db` (pointer swap) y `a.dbPath`. Correcto.
- Concatenación de strings para SQL: los strings concatenados vienen de whitelist/constantes, no de input del usuario. No hay vector de inyección.

## Trigger Recursivo SQLite
- `trg_exp_auditoria` hace `UPDATE expedientes SET ...` dentro de `AFTER UPDATE`. Con `recursive_triggers=OFF` (default) la UPDATE no se dispara recursivamente, pero sí modifica la fila. Lógica correcta.
- Archivo `01_master_control_docs_presidencia.sql` **NO está truncado**: termina en línea 395 con `-- PRAGMA user_version = 8;`. El trigger está completo (líneas 187-231).
- **Triggers de los 8 módulos adicionales sí existen**: están completos en `02_modulos_adicionales.sql` (líneas 990+). Si solo se revisó `01_*.sql`, parecen faltar pero no es así.

## XSS en jsonEncode
- `json.Marshal` escapa strings + reemplazos `</script>`/`<!--` + `template.JS`. Es el patrón estándar de Go para inyectar JSON en `<script>`. No hay vector real.

## formatNumGo precisión
- Ya usa `math.Round(f*100)/100`. No hay error de precisión.

## parseSpanishNumber
- Solo procesa si `strings.Contains(s, ",")`. Entradas sin coma (formato inglés `1.5`) pasan sin modificar.

## buildGanttColumns — feriados / año bisiesto
- 60 días hábiles saltando sáb/dom es el estándar. Feriados son locales y no aplican. Año bisiesto lo maneja `time.Time`.

## Sin inicialización de esquema
- La app abre bases de datos existentes. La creación del schema es responsabilidad del usuario vía scripts SQL.

## go.mod version
- `go 1.25.0` es la versión mínima requerida por las dependencias del proyecto (determinado por `go mod tidy`). Con Go 1.26.5 instalado compila sin errores. Si un auditor reporta "versión inexistente", está usando una instalación de Go desactualizada.

## Paginación 100% cliente
- Intencional para datasets pequeños (<5000 filas). Se migrará a SQL con LIMIT/OFFSET cuando sea necesario.

## alert() sobrescrito globalmente
- Intencional: redirige `alert()` a toast de error. No es bug de seguridad.

## ObtenerFilas usa vista / GuardarFila usa tabla
- Por diseño. `ObtenerFilas` lee de la vista (joins, alias). `GuardarFila` escribe en la tabla base. Las columnas de INSERT coinciden con las de la tabla real, no con la vista.

## handleCSV
- **No escapa comas**: `encoding/csv` escapa automáticamente comas, saltos de línea y comillas.
- **Falta `wr.Flush()`**: El `Flush()` SÍ está presente (línea 895 de handler.go), seguido de `wr.Error()` (línea 896). El CSV se escribe completo.
- **Sin BOM UTF-8**: Excel abre CSV UTF-8 correctamente desde hace años. BOM es opcional.

## COALESCE redundante
- `COALESCE(db_id, 0)` donde `db_id IS NOT NULL` es redundante pero no causa ningún bug.

## handleRutaProcesos sin HasDB / Modulos en contexto
- El template `ruta_procesos.html` solo usa `Legend`, `Columns`, `Processes`. No necesita otras variables.

## CSS/JS vendors hardcoded
- Embebidos vía `//go:embed frontend/*`. Si el embed falla, el binario no compila. No pueden "faltar".

## handleExportarExcel archivos temporales
- `excelize.NewFile()` es 100% en memoria. No crea archivos temporales en disco.

## Rate limiting
- App de escritorio Wails con un solo usuario. Rate limiting no aplica.

## beforeunload
- Mejora de UX, no un bug. No hay pérdida de datos porque el backend ya guarda.

## PostFormValue vs FormValue
- `handleEliminarExpediente` usa `PostFormValue` (solo body de POST). `handleGuardarExpediente` usa `FormValue` (body + query). Ambos funcionan correctamente en su contexto.

## Validación de id negativo
- Si el id es inválido, devuelve formulario vacío (nuevo registro). Comportamiento esperado.

## Inconsistencia columnas tabla vs vista
- Columnas de INSERT (`cfg.Columnas`) corresponden a la tabla real. La vista tiene columnas adicionales (joins) de solo lectura. Es correcto.

## localStorage sin límite
- El límite de ~5MB es amplio para listas de IDs. No hay riesgo real.

## WebView2 Runtime Dir hardcodeado
- Wails requiere una versión específica del runtime para el modo `embed`. Es una constante de build, no un path de usuario. Correcto.

## Path Traversal en AbrirBaseDatos
- El path del archivo .db lo selecciona el usuario mediante el diálogo nativo del SO. No es input web arbitrario. Riesgo nulo.

## Sin graceful shutdown
- `wails.Run()` bloquea hasta que se cierra la ventana. SQLite en modo WAL sobrevive a crashes. No se necesita `SIGTERM` handler en app desktop.

## Sin context.Context en queries SQL
- App monousuario con queries que retornan en <100ms. `context.Context` agrega overhead sin beneficio en este escenario.

## Variables globales mutables
- `Modulos`, `catalogosValidos`, `columnasNumericas` etc. son efectivamente inmutables en runtime. Go no permite `const` para maps/slices — es la forma idiomática.

## División por cero en conversión USD/Bs
- `if (!tc) return;` cubre `0`, `NaN`, `undefined` y `null`. No hay división por cero posible.

## int vs int64 en IDs
- En arquitectura 64-bit (único target), `int` = `int64`. SQLite ROWIDs nunca excederán 2^31 en esta app.

## Sin tests / Sin GoDoc / Sin health check / Logging no estructurado
- Son métricas de calidad de código, no bugs funcionales. Fuera del scope de una auditoría de bugs.

## `f.Close()` en excelize no existe
- excelize v2.11.0 **sí** tiene método `Close()` (`File.Close()`). El proyecto compila sin errores.

## `trg_exp_auditoria`: inconsistencia historial vs tabla
- El trigger auto-corrige `id_estatus` basado en `fecha_firma_contrato` DESPUÉS del snapshot. El historial captura el valor ingresado por el usuario; la tabla queda con el valor corregido. Es intencional: la BD asegura integridad de datos (contrato firmado = FIRMADO, sin contrato = PENDIENTE), mientras el historial preserva lo que el usuario envió.

## `ObtenerCatalogos` sin caché
- Se llama en cada carga/refresco sin caché. Para una app desktop monousuario con <100 items por catálogo, el overhead es insignificante (~11 queries sub-milisegundo en SQLite). No justifica la complejidad de una caché con invalidación.

## `formatNum` JS definido pero no usado
- Definido en `index.html` como helper de formateo legacy. Quedó sin invocar tras la migración a `formatNumGo` en templates Go. Inofensivo.

## `PageData.SortColumn` con valor inicial muerto
- `SortColumn: "fecha_creacion"` se asigna y luego se pisa con `cfg.IDColumna`. Es una inicialización inofensiva de struct.

## `DescargarBD` código muerto
- Expuesta vía Wails `Bind` para uso futuro (exportar respaldo vía menú nativo). No es bug, es feature pendiente de UI.

## `SetBackupMaxCopies` sin UI
- Expuesta vía Wails `Bind`, configurable a futuro desde settings. No es bug.

## `CURRENT_TIMESTAMP` vs `CURRENT_DATE` en triggers
- Los triggers de los 9 módulos usan `CURRENT_TIMESTAMP` en UPDATE de `fecha_actualizacion`, mientras las columnas tienen `DEFAULT CURRENT_DATE`. SQLite almacena correctamente ambos, pero es una inconsistencia menor. No causa bugs funcionales.

## `handleCSV` sin caller en UI
- La ruta `/api/csv` está registrada pero ningún botón la invoca (solo `/api/exportar-excel`). Si no se usa, puede eliminarse para reducir la superficie de mantenimiento. No es bug funcional.


