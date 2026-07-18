# Falsos Positivos Conocidos — No reportar como bugs

Estos hallazgos aparecen en todas las auditorías pero NO son bugs reales. Ignorar.

## SQL Injection / Parametrización
- `sanitizarOrden()`, `ObtenerColumnasVista()` y `GuardarNuevoCatalogo()` concatenan nombres de tablas/vistas/columnas **vs whitelist** (`Modulos`, `catalogosValidos`, `columnasOrdenValidas`). Ningún input del usuario pasa sin validación. Seguro.
- Mutex redundante con `sql.DB`: `sql.DB` es thread-safe, pero el mutex protege `a.db` (pointer swap) y `a.dbPath`. Es correcto.

## Trigger Recursivo SQLite
- `trg_exp_auditoria` hace `UPDATE expedientes SET id_estatus = ...` dentro de un `AFTER UPDATE` trigger. Con `recursive_triggers=OFF` (default) la UPDATE **no se dispara recursivamente**, pero **sí modifica la fila**. La lógica funciona correctamente.
- Archivo SQL **NO está truncado** — termina correctamente en línea 395 con `-- PRAGMA user_version = 8;`. El trigger `trg_exp_auditoria` está completo (líneas 187-231).

## XSS en jsonEncode
- `json.Marshal` escapa strings + reemplazos `</script>`/`<!--` + `template.JS` es el patrón estándar de Go para inyectar JSON en `<script>`. No hay vector real.

## formatNumGo precisión
- Ya usa `math.Round(f*100)/100`. No hay error de precisión.

## parseSpanishNumber
- Solo procesa si `strings.Contains(s, ",")`. Entradas sin coma (formato inglés `1.5`) pasan sin modificar.

## buildGanttColumns — feriados / año bisiesto
- 60 días hábiles saltando sáb/dom es el estándar. Feriados son locales y no aplican. Año bisiesto lo maneja `time.Time`.

## Sin inicialización de esquema
- La app abre bases de datos existentes. La creación del schema es responsabilidad del usuario vía los scripts SQL.

## go.mod version
- Ya corregido a `go 1.23.0`.

## Paginación 100% cliente
- Intencional para datasets pequeños (<5000 filas). Se migrará a SQL cuando sea necesario.

## alert() sobrescrito globalmente
- Intencional: redirige `alert()` a toast de error.

## ObtenerFilas usa vista / GuardarFila usa tabla
- Por diseño. `ObtenerFilas` lee de la vista (joins, alias), `GuardarFila` escribe en la tabla base. Las columnas de INSERT coinciden con las columnas reales de la tabla, no de la vista.

## handleCSV no escapa comas / saltos de línea
- Usa `encoding/csv` (`csv.NewWriter`), que escapa automáticamente comas, saltos de línea y comillas.

## COALESCE redundante en subquery
- `COALESCE(db_id, 0)` donde `db_id IS NOT NULL` — es redundante pero no causa ningún bug ni error.

## handleRutaProcesos sin HasDB / Modulos en contexto
- El template `ruta_procesos.html` solo usa `Legend`, `Columns`, `Processes`. No necesita las otras variables.

## CSS/JS vendors hardcoded
- Están embebidos vía `//go:embed frontend/*`. No pueden "faltar" — si el embed falla, el binario no compila.

## handleExportarExcel archivos temporales
- `excelize.NewFile()` es 100% en memoria. No crea archivos temporales en disco.

## Rate limiting
- Es una app de escritorio Wails con un solo usuario. Rate limiting no aplica.

## beforeunload para formularios abiertos
- Mejora de UX, no un bug. No hay pérdida de datos porque el backend ya guarda.

## Inconsistencia PostFormValue vs FormValue
- `handleEliminarExpediente` usa `PostFormValue` (solo POST body) porque es un POST. `handleGuardarExpediente` usa `FormValue` (body + query) por conveniencia. Ambos funcionan correctamente.

## Validación de id negativo en handleCargarExpediente
- Si el id es inválido, devuelve formulario vacío (nuevo registro). Comportamiento esperado.

## sql.NullString para fecha/nota en ObtenerRutaProcesosData
- `fecha` y `nota` se escanean como `string`. Si la columna es NULL, `rows.Scan` devuelve error, que se captura con `continue` y `log.Printf`. No causa crash.

## Inconsistencia columnas entre tabla y vista
- Las columnas de INSERT (`cfg.Columnas`) corresponden a la tabla real. La vista tiene columnas adicionales (joins) que son solo lectura. Es correcto.

## localStorage sin límite en toggleFrecuente
- El límite de localStorage (~5MB) es amplio para listas de IDs. No hay riesgo real.
