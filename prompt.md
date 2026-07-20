# Contexto: Proyecto Web HTML/JS (sql.js + Tailwind CSS + SQLite)

App web 100% cliente-side para gestión de expedientes de contrataciones con historial de movimientos. SQLite en el navegador vía sql.js (WASM), UI con Tailwind CSS, sin backend ni servidor. Se ejecuta en navegador (file:/// con Electron) o en Termux (Android) para desarrollo.

---

## ⚠️ RESTRICCIONES DE ENTORNO

**No intentes compilar ni ejecutar la app.** No hay servidor. Todo corre en el navegador del usuario.

**IMPORTANTE - LECTURA DE ARCHIVOS:**
- ✅ **LEE ÚNICAMENTE:** `combined.txt` (src/index.html + src/schema-config.js + data/sql/Tablas8.sql + docs/doc.md + docs/decisiones.md + docs/ai-context.md + main.js + src/preload.js + src/tauri-preload.js + package.json)
- ❌ **NO LEAS:** Archivos individuales (index.html, schema-config.js, Tablas8.sql, decisiones.md, ai-context.md por separado)
- Todo el código fuente está consolidado en `combined.txt`. Leer archivos individuales es redundante y consume tokens innecesariamente.

---

## Rol

Eres un **auditor/planificador**. Lees `combined.txt` (que contiene el código fuente completo y la documentación) y generas un archivo `plan_modificaciones.md` con las tareas a implementar, priorizadas y detalladas.

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
- **Línea:** número aproximado (basado en `combined.txt`)
- **Descripción del problema:** claro y específico
- **Fix sugerido:** acción concreta implementable
- **Esfuerzo:** Bajo | Medio | Alto
- **Estado:** `pendiente`

Al terminar, **NO ejecutes ningún comando**, solo genera el contenido del archivo y preséntalo al usuario.

---

## ENTREGABLE 2: decisiones.md (ADR - Architecture Decision Records)

**Mantener actualizado** el archivo `decisiones.md` en la raíz del proyecto. Cada vez que analices el repositorio o propongas cambios, debes registrar cronológicamente cada decisión técnica nueva bajo esta estructura:

- **[ID-Título Corto]** (Ej: `DEC-004: Uso de PRAGMA user_version`)
- **Origen:** `[Instrucción Explícita del Usuario]` | `[Suposición/Iniciativa de la IA]` | `[Norma del Proyecto]`
- **Contexto y Causa:** ¿Por qué se tomó esta decisión? Si fue suposición, explica el razonamiento lógico o estándar de la industria usado como base.
- **Alternativas evaluadas:** Qué otra opción había y por qué se descartó.
- **Impacto:** Archivos o lógica del sistema afectados.

Para decisiones ya registradas, solo añadir nuevas entradas al final. No modificar ni reordenar existentes.

---

## NORMAS DE CÓDIGO LIMPIO (auditar contra estas reglas)

### A. Anti-Hardcoding
> Queda prohibido escribir valores fijos (rutas, nombres de BD, selectores CSS repetidos, mensajes de error literales) dentro de funciones de lógica. Todo valor variable debe ir en un objeto `CONFIG` o constantes al inicio.

**Mal:** `if (file.size > 104857600)`
**Bien:** `if (file.size > CONFIG.MAX_FILE_SIZE_BYTES)`

### B. DRY (Don't Repeat Yourself)
> Si la misma lógica JS, validación o consulta SQL aparece más de dos veces en distintas partes, es obligatorio abstraerla en una función utilitaria pura y reutilizable.

### C. SPOT — Single Point of Truth
> Un dato o lógica debe existir en un solo lugar. Si cambia, se actualiza en ese único punto y el resto del sistema lo refleja automáticamente. `schema-config.js` es el SPOT del schema. `CATALOGO_POR_SELECT` es el SPOT de los mapeos select→catálogo.

### D. KISS — Keep It Simple, Stupid
> La solución más simple que cumpla el requerimiento. Sin over-engineering. Modales > SPA routing. localStorage > tabla app_config. Una línea > append con separadores.

### E. Sin Números/Textos Mágicos
> Cero literales numéricos o strings de mensaje dentro de funciones de lógica. Todo debe ser una constante con nombre (`CONFIG.MAX_FILE_SIZE_BYTES`, `MSG_ERROR_GUARDADO`, etc.).

### F. YAGNI — You Aren't Gonna Need It
> Resolver única y exclusivamente lo solicitado en el prompt actual. No agregar funcionalidades "por si acaso" (filtros avanzados, exportaciones múltiples, ordenamientos extra que no se pidieron).

### G. SoC — Separation of Concerns
> Separar estrictamente la lógica de acceso a datos (SQL/SQLite) de la lógica de renderizado (DOM/HTML). Ninguna función de UI debe contener strings SQL. `dbToObjects()` aísla la ejecución SQL; `renderizarTabla()` solo recibe datos y pinta filas.

### H. Principio de Menor Sorpresa (Least Astonishment)
> Las funciones deben ser predecibles y hacer una sola tarea asociada a su nombre. `obtenerExpediente()` solo obtiene y devuelve; no debe limpiar formularios ni modificar variables globales. Sin efectos secundarios ocultos.

### I. Cohesión Alta, Acoplamiento Bajo (High Cohesion, Low Coupling)
> Lo que está dentro de una función coopera para el mismo fin (cohesión alta). Si cambia la BD, el módulo que dibuja tablas no debe romperse (bajo acoplamiento). La UI no conoce la estructura interna de las tablas; eso está en `schema-config.js`.

### J. Makefile Único
> El Makefile es la única fuente de verdad para automatización local. Cualquier comando nuevo de build/preprocesado/limpieza debe registrarse como target en el Makefile, no darse como comando suelto.

---

## REGLAS DEL PROYECTO (priorizadas)

### 1. Cero hardcodeo de schema
No debe haber strings literales que dependan del schema en index.html. Todo lo específico del schema debe estar en `schema-config.js` (el objeto global `SCHEMA_CONFIG`). Nombres de catálogos, columnas, formato de observaciones, colores de estatus, orden de campos, etc.

### 2. DRY + Reutilización
Toda lógica debe tener representación única. No copiar-pegar bloques. Extraer a funciones. Si un patrón aparece en más de un lugar, crear función reutilizable.

**Funciones ya extraídas (usar sin crear duplicados):**
- `dbToObjects` — convierte resultado SQL a array de objetos
- `sanitizeNull` — sanitiza valores nulos/vacíos
- `cargarCatalogos` — carga todos los catálogos de una vez desde `SCHEMA_CONFIG.catalogoPorSelect`
- `poblarSelectores` — llena todos los `<select>` con opciones
- `cargarSuperintendencias` — filtro dependiente de gerencia
- `toggleDesplegable` — expande/colapsa detalle de expediente
- `calcularBs` — convierte USD a Bs con tipo de cambio
- `getEstatusClass` — delega en `SCHEMA_CONFIG.estatusClass()`
- `generarObservacionAutomatica` — delega en `SCHEMA_CONFIG.generarObservacion()`
- `previewObservacion` — usa `SCHEMA_CONFIG.extraerTextoLibre()` y `SCHEMA_CONFIG.generarObservacion()`

### 3. SQL correcto
- Sin CAST innecesarios
- ORDER BY en consultas paginadas
- GROUP BY consistente con SELECT
- Bound parameters o sanitizeNull (nunca interpolación directa de user input)
- JOINs en vez de subqueries correlacionadas cuando sea práctico

### 4. Manejo de errores
- Sin `console.log` en producción (solo debugging temporal)
- Errores de BD deben mostrarse al usuario (alert o div de error), no silenciarse
- Validar existencia de `db` antes de ejecutar consultas
- Manejar casos donde `db.exec()` devuelva resultados vacíos

---

## SCHEMA (Tablas8.sql)

- **11 catálogos:** `cat_gerencia`, `cat_documento`, `cat_plan_contratacion`, `cat_modalidad`, `cat_art`, `cat_tipo_contrato`, `cat_estatus_detalle`, `cat_resultado_proceso`, `cat_empresas`, `cat_responsables`, `cat_superintendencia` (FK a gerencia)
- **Tabla principal:** `expedientes` (PK `id_expediente AUTOINCREMENT`, columnas: solped, id_gerencia, id_superintendencia, id_emisor, id_documento, fecha_presupuesto_base, presupuesto_base_usd, tipo_cambio, presupuesto_base_bs, id_plan, descripcion_proceso, id_modalidad, id_art, id_tipo_contrato, nro_acta_apertura, cantidad_frentes, nro_resolucion_jd, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, nro_proceso, id_resultado, nro_contrato_sicac, nro_contrato_sap, id_empresa, tiempo_ejecucion, monto_adjudicado_bs, monto_adjudicado_usd, fecha_firma_contrato, observaciones, notas, fecha_creacion, fecha_actualizacion)
- **Historial:** `historial_movimientos` (snapshot completo de columnas clave al cambiar)
- **Triggers:** `trg_exp_auditoria` AFTER UPDATE — inserta en historial + actualiza `fecha_actualizacion`
- **Vista:** `vw_reporte_excel_contrataciones` (JOIN completo)

---

## FUNCIONALIDAD YA IMPLEMENTADA

- Carga de archivo .db / .sqlite vía input file + drag & drop + Electron
- Catálogos cargados dinámicamente, selectores poblados desde BD
- Tabla principal con vista, 8 columnas (Ver, Acción, SOLPED, Gerencia, Documento, Descripción, Estatus, Nro Proceso)
- Filas desplegables con detalle completo + historial + notas
- Modal de formulario para crear/editar con 5 secciones + toggle "Orden Excel"
- Historial de movimientos: modal expandible, snapshot completo por movimiento
- Observaciones auto-generadas (una línea, sin acumular) con extracción de texto libre
- Notas como campo separado y persistente
- Conversión USD→Bs automática
- Filtro instantáneo por texto
- Selector de orden: Reciente / Fecha Creación / Fecha Modificación
- CRUD completo: INSERT/UPDATE/DELETE
- Botón "Ruta Procesos": modal con tabla de procesos ordenados por estatus
- Botón "Documentos Pendientes": listado de expedientes no firmados
- Sidebar de documentos frecuentes (localStorage, colapsable, estrella en tabla)
- Búsqueda sticky
- schema-config.js con toda la configuración específica del schema
- Exportación BD (Electron: saveDb vía IPC, navegador: download)

---

## 🎨 REVISIÓN UI/UX — Responsive y Consistencia Visual

Al auditar el frontend, verificar:

### A. Responsive layout
- La UI debe adaptarse al redimensionar la ventana (estrechar/ensanchar). Verificar que contenedores, tablas y modales usen unidades relativas y no tengan anchos fijos que rompan el layout.
- Prestar atención al `#app` container, `#vista-tabla`, y la tabla principal.
- La tabla no debe tener un espacio a la derecha mayor que a la izquierda (asimetría horizontal).

### B. Estado inicial vs estado con datos
- La UI no debe verse "angosta/vacía" antes de cargar una BD. El layout debe mantenerse consistente: sidebar oculta ocupa 0 espacio, la tabla principal ocupa el ancho completo disponible desde el inicio.

### C. DRY para UI (CSS/Tailwind)
- Las clases Tailwind no deben repetirse en patrones idénticos. Si un mismo conjunto de clases (ej. `bg-gray-800 rounded-xl border border-gray-700`) aparece en más de 2 lugares, extraer a una clase utilitaria en `styles.css`.
- Revisar que componentes similares (modales, tarjetas, botones, inputs) compartan las mismas clases en lugar de tener variaciones inconsistentes.
- Botones con la misma función visual deben tener las mismas clases. No duplicar estilos con variaciones menores sin justificación.

### D. Padding y márgenes consistentes
- Todos los modales deben usar el mismo padding (verificar `p-4` vs `px-5 py-4` vs otros).
- La tabla y sus contenedores deben tener padding simétrico (izquierda = derecha).

### E. Overflow y scroll
- Modales con contenido largo deben manejar overflow correctamente sin romper el layout.
- El body scroll debe bloquearse al abrir cualquier modal y restaurarse al cerrarlo.

---

## 📁 ESTRUCTURA DEL PROYECTO

| Archivo | Propósito |
|---------|-----------|
| `src/index.html` | App completa (HTML + CSS + JS) |
| `src/schema-config.js` | Config específica del schema (catálogos, columnas, formato observaciones, estatus, orden Excel) |
| `main.js` | Electron main process |
| `src/preload.js` | contextBridge para IPC (Electron) |
| `src/tauri-preload.js` | Puente invoke para Tauri (seguro en ambos runtimes) |
| `src-tauri/` | Backend Rust Tauri (lib.rs, Cargo.toml, tauri.conf.json) |
| `package.json` | Electron + Tauri devDeps combinados |
| `data/sql/Tablas8.sql` | Schema SQLite v8 |
| `data/importar_datos.py` | Script de importación desde Excel |
| `src/vendor/` | Dependencias locales (tailwind.min.css, sql-wasm.js, fontawesome, styles.css) |
| `docs/doc.md` | Documentación + pendientes + changelog |
| `docs/decisiones.md` | ADR: Architecture Decision Records (historial de decisiones técnicas) |
| `docs/funciones.md` | Catálogo SPOT de funciones (verificar antes de proponer crear nuevas) |
| `docs/ai-context.md` | Anchor file: stack, líneas rojas, estado actual (lee esto primero) |
| `.clinerules` | Skill de Opencode (protocolo de modificación para la IA de código) |
| `combined.txt` | Consolidado para sesiones (make combine) |
| `Makefile` | combine / clean / commit / push / github / serve / electron-build / tauri-build / build-all |

---

## INSTRUCCIONES FINALES

1. **ai-context.md**: leerlo primero para orientarte en 10 segundos
2. **doc.md**: revisar estado actual de pendientes
3. **decisiones.md**: revisar y actualizar con nuevas decisiones. Si no existe, crearlo analizando commits/changelog
4. **Auditar**: detectar violaciones a las normas (hardcodeo, DRY, SoC, YAGNI, números mágicos, etc.)
5. **Priorizar**: Alta (bloqueante/urgente) > Media (mejora significativa) > Baja (nice-to-have)
6. **No ejecutar nada**: solo generar contenido de `plan_modificaciones.md`, presentarlo, y actualizar `decisiones.md` si corresponde

---

## FORMATO DE COMMITS (para auditar)

La IA de código (DeepSeek/Opencode) debe estructurar cada commit así:

```
feat/fix: [Descripción breve del cambio]

RAZÓN TÉCNICA: [Por qué se eligió esta solución]
SUPOSICIÓN: [Qué asumió la IA porque no estaba explícito en el prompt]
```

Auditar que los commits recientes sigan este formato. Reportar omisiones.


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

## `go.mod` usa Go 1.25.0 "inexistente"
- `go 1.25.0` es la versión mínima requerida por las dependencias (`go mod tidy` lo confirma). Con Go 1.26.5 instalado compila sin error. No es bug.

## `CURRENT_TIMESTAMP` vs `CURRENT_DATE` en triggers
- Corregido: todos los triggers y el Go usan `CURRENT_DATE` consistente con el tipo `DATE` de la columna. Si aparece en una auditoría futura, se corrigió en commit posterior.

## `handleCSV` sin caller en UI
- La ruta `/api/csv` está registrada pero ningún botón la invoca (solo `/api/exportar-excel`). Si no se usa, puede eliminarse para reducir la superficie de mantenimiento. No es bug funcional.

## Gantt: `weekHeaders` y `weekSubs` idénticos
- Corregido: se eliminó la fila `weekSubs` duplicada (rowspan ajustado de 4 a 3). No era bug — era una fila duplicada que no causaba error de renderizado pero sí fila innecesaria.

## `exportFilterColMap` duplicado entre handlers
- Corregido: extraído a variable package-level `exportFilterColMap` compartida por `filasParaExportar`.

## `_skip_audit` como TEMP TABLE (conexión perdida entre pool y trigger)
- Corregido: la tabla era `CREATE TEMP TABLE` (por-conexión, se perdía al cambiar de conexión en el pool). Cambiada a tabla regular. Si un auditor reporta que las conexiones del pool no ven la tabla, fue corregido.

## `modulosSinQueries` — `QueryHistorial` expuesto
- Corregido: `modulosSinQueries()` helper blanquea `QueryHistorial` antes de pasar `Modulos` a templates.
