# Plan de Modificaciones

**Fecha de auditoría:** Julio 2026
**Archivos revisados:** `combined.txt` (3800 líneas), `doc.md`, `decisiones.md`, `ai-context.md`, `funciones.md`

**Estado general:** El proyecto está en un estado sólido. Todos los hallazgos de auditoría y propuestas de mejora han sido implementados.

---

## Pendientes desde doc.md

| # | Prioridad | Descripción | Archivos | Estado |
|---|-----------|-------------|----------|--------|
| 3 | 🟢 Baja | **Colores por frecuencia de edición**: color distinto para campos según qué tan frecuente se editan (1ra, 2da, 3ra vez, etc.) | `src/index.html`, `src/vendor/styles.css` | `pendiente` |
| 6 | 🔴 Alta | **Schemas separados para demás hojas del Excel**: cada hoja del Excel es un módulo independiente con su propio schema (ej. `Tablas8_hoja2.sql`), sin contaminar el schema principal | `data/sql/*.sql` | `pendiente` |

---

## Hallazgos de auditoría

### Media Prioridad

| ID | Archivo | Línea | Descripción | Esfuerzo | Estado |
|----|---------|-------|-------------|----------|--------|
| **AUD-001** | `src/index.html` | ~469 | `console.error` directo en smoke test en lugar de `DEBUG.error` | Bajo | ✅ |
| **AUD-002** | `src/index.html` | ~1423 | SQL hardcodeada en `cargarExpediente()` en lugar de usar `SCHEMA_CONFIG.queries.expedientePorId` | Bajo | ✅ |
| **AUD-003** | `src/index.html` | ~1342 | `validarFechas()` acoplada al DOM: crea/remueve elementos directamente | Medio | ✅ |

### Baja Prioridad

| ID | Archivo | Línea | Descripción | Esfuerzo | Estado |
|----|---------|-------|-------------|----------|--------|
| **AUD-004** | `src/index.html` | ~795 | Límite de 8 recientes hardcodeado (`recientes.length > 8`) | Bajo | ✅ |
| **AUD-005** | `src/index.html` | ~917, ~974 | `chunkSize = 8192` repetido 3 veces | Bajo | ✅ |
| **AUD-006** | `src/index.html` | ~33-381 | Clases CSS de modales repetidas inline en 5 modales | Bajo | ✅ |
| **AUD-007** | `src/index.html` | ~443 | Error boundary no bloqueaba scroll del body | Bajo | ✅ |

---

## Propuestas de mejora

| ID | Descripción | Esfuerzo | Estado |
|----|-------------|----------|--------|
| **PROP-001** | PRAGMA integrity_check al cargar BD | Bajo | ✅ |
| **PROP-002** | Mostrar `fecha_actualizacion` en fila desplegable | Bajo | ✅ |
| **PROP-003** | Confirmación VACUUM para BDs >50MB | Bajo | ✅ |
| **PROP-004** | Exportar CSV desde vista `vw_reporte_excel_contrataciones` | Medio | ✅ |
| **PROP-005** | Persistir orden preferido en localStorage | Bajo | ✅ |

---

## Implementación completada

- **AUD-001**: `console.error` → `DEBUG.error` en smoke test de SELECTORS
- **AUD-002**: `cargarExpediente()` ahora usa `SCHEMA_CONFIG.queries.expedientePorId`
- **AUD-003**: `validarFechas()` dividida en `validarFechasEntre()` (pura) + UI en `validarFechas()`
- **AUD-004**: `CONFIG.MAX_RECIENTES = 8` en schema-config.js
- **AUD-005**: `CONFIG.EXPORT_CHUNK_SIZE = 8192` en schema-config.js, usado en `guardarBD()` y `descargarBDError()`
- **AUD-006**: Clase `.modal-content` en `styles.css` aplicada a los 5 modales
- **AUD-007**: `document.body.style.overflow = 'hidden'` en `window.onerror` y `window.onunhandledrejection`
- **PROP-001**: `PRAGMA integrity_check` después de crear `SQL.Database` en `_cargarBaseDatosComun()`
- **PROP-002**: Campo "Última Modificación" añadido en "Valores y Trazabilidad" del desplegable
- **PROP-003**: Confirmación en `optimizarBD()` si la BD > `CONFIG.VACUUM_CONFIRM_THRESHOLD_MB` (50MB)
- **PROP-004**: Botón "CSV" en header + función `exportarCSV()` que descarga datos de `vw_reporte_excel_contrataciones`
- **PROP-005**: `cambiarOrden()` persiste en `localStorage`; `_cargarBaseDatosComun()` restaura al cargar BD
