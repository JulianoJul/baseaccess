# Falsos Positivos Conocidos — No reportar como bugs

Estos hallazgos aparecen en todas las auditorías pero NO son bugs reales. Ignorar.

## SQL Injection
- `sanitizarOrden()`, `ObtenerColumnasVista()` y `GuardarNuevoCatalogo()` usan nombres de tablas/vistas/columnas vs whitelist (`Modulos`, `catalogosValidos`, `columnasOrdenValidas`). No hay concatenación de input del usuario. Seguro.
- Mutex redundante con `sql.DB`: `sql.DB` es thread-safe, pero el mutex protege `a.db` (pointer swap) y `a.dbPath`. Es correcto.

## Trigger Recursivo SQLite
- `trg_exp_auditoria` hace `UPDATE expedientes SET id_estatus = ...` dentro de un `AFTER UPDATE` trigger. Con `recursive_triggers=OFF` (default) la UPDATE **no se dispara recursivamente**, pero **sí modifica la fila**. La lógica funciona.

## XSS en jsonEncode
- `json.Marshal` escapa strings correctamente + reemplazos `</script>`/`<!--` + `template.JS` es el patrón estándar de Go para inyectar JSON en `<script>`. No hay vector real.

## formatNumGo precisión
- Ya usa `math.Round(f*100)/100`. No hay error de precisión.

## parseSpanishNumber
- Solo procesa si `strings.Contains(s, ",")`. Entradas sin coma (formato inglés `1.5`) pasan sin modificar. Es correcto.

## buildGanttColumns — no maneja feriados
- 60 días hábiles saltando sáb/dom es el estándar. Los feriados son locales y no aplican.

## Sin inicialización de esquema
- La app abre bases de datos existentes. La creación del schema es responsabilidad del DBA/usuario vía los scripts SQL. No es función de la app.

## go.mod version
- Ya corregido a `go 1.23.0`.

## Paginación 100% cliente
- Intencional para datasets pequeños (<5000 filas). Cuando sea necesario se migrará a SQL con LIMIT/OFFSET.

## alert() sobrescrito globalmente
- Intencional: redirige `alert()` a toast de error. No es bug de seguridad.
