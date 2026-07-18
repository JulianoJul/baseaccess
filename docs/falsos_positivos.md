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
- Ya corregido a `go 1.23.0`.

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
