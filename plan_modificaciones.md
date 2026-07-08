# Plan de Modificaciones

**Generado:** Julio 2026
**Estado del proyecto:** ✅ Todos los items implementados. Ver `docs/decisiones.md` DEC-020.

Prioridad: 🔴 Alta > 🟡 Media > 🟢 Baja

---

## Pendientes desde doc.md

| # | Prioridad | Descripción | Archivo | Estado |
|---|-----------|-------------|---------|--------|
| 1 | 🔴 Alta | **Cero hardcodeo del schema**: Todo lo específico del schema debe estar en `schema-config.js`. Nombres de catálogos, columnas, formato de observaciones, colores de estatus, orden de campos. | `src/index.html` | ✅ Completado (DEC-007, DEC-015) |
| 2 | 🟡 Media | **Toggle de orden en edición**: Botón para alternar entre orden por secciones lógicas y orden Excel en el formulario de edición. | `src/index.html`, `src/schema-config.js` | ✅ Completado (DEC-012) |
| 4 | 🟢 Baja | **Botón "Ruta Procesos"**: Modal con tabla de ruteo de procesos (emisor, receptor, estatus, fechas). | `src/index.html` | ✅ Completado (DEC-013) |
| 5 | 🟢 Baja | **Botón "Documentos Pendientes"**: Listado de expedientes no firmados. | `src/index.html` | ✅ Completado (DEC-013) |
| 7 | 🟢 Baja | **Columna "Descripción" en tabla principal**: Visible junto a SOLPED, Gerencia, Documento. | `src/index.html` | ✅ Completado |
| 8 | 🟢 Baja | **Selector de orden**: Reciente / Fecha Creación / Fecha Modificación. | `src/index.html` | ✅ Completado |
| 9 | 🟢 Baja | **Sidebar de frecuentes + búsqueda sticky**: Colapsable, persistencia localStorage. | `src/index.html` | ✅ Completado (DEC-011) |

**Conclusión:** Todos los pendientes documentados en `doc.md` han sido implementados. No hay pendientes activos desde la documentación.

---

## Hallazgos de auditoría

### 🔴 Alta Prioridad (Bloqueantes / Riesgo Crítico)

No se encontraron hallazgos de alta prioridad. Las siguientes violaciones críticas ya fueron resueltas en la auditoría anterior (cambio #59):

- ✅ Números mágicos reemplazados por `CONFIG.*`
- ✅ `console.log` envueltos en `DEBUG.isEnabled`
- ✅ Strings literales en alertas → `MSG.*`
- ✅ localStorage keys → `STORAGE_KEYS.*`
- ✅ Selectores DOM repetidos → `SELECTORS.*` + helper `$()`
- ✅ SQL mezclado con UI → funciones data layer (`obtenerRutaProcesos`, `obtenerDocumentosPendientes`)
- ✅ `generarObservacion()` desacoplada del DOM (ahora recibe parámetros)

### 🟡 Media Prioridad (Mejora Significativa)

#### 1. Validación de archivo BD usa división en alerta en lugar de constante formateada

- **Archivo:** `src/index.html`
- **Línea:** ~708
- **Descripción del problema:** `alert(MSG.ERROR_TAMANO(file.size / 1024 / 1024))` — la función de mensaje recibe un cálculo en lugar de usar `CONFIG.MAX_FILE_SIZE_MB` directamente o tener el formateo dentro de `MSG`.
- **Fix sugerido:** Extraer cálculo a constante local o mover la lógica de formateo dentro de `MSG.ERROR_TAMANO`:
  ```javascript
  const sizeMB = file.size / CONFIG.BYTES_PER_MB;
  alert(MSG.ERROR_TAMANO(sizeMB));
  ```
  Y añadir `BYTES_PER_MB: 1048576` a `CONFIG`.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

#### 2. Función `optimizarBD()` podría reportar error vía `MSG_EXTRA.VACUUM_ERROR`

- **Archivo:** `src/index.html`
- **Línea:** ~956-967
- **Descripción del problema:** La función `optimizarBD()` tiene un bloque try-catch pero no usa `MSG_EXTRA.VACUUM_ERROR(err)` para mostrar el mensaje de error al usuario, solo hace `DEBUG.error()`.
- **Fix sugerido:** Añadir `alert(MSG_EXTRA.VACUUM_ERROR(err))` en el catch para consistencia con otros mensajes de mantenimiento.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

#### 3. `updateUIOnError()` deshabilita botones pero no muestra mensaje explicativo

- **Archivo:** `src/index.html`
- **Línea:** ~967-972
- **Descripción del problema:** Cuando ocurre un error crítico, los botones se deshabilitan pero el usuario no recibe feedback visual de *por qué* están deshabilitados (solo el modal de error).
- **Fix sugerido:** Añadir un badge o texto junto a `#estado-bd` que diga *"Modo solo-lectura (error crítico)"* cuando `updateUIOnError()` se ejecuta.
- **Esfuerzo:** Medio
- **Estado:** ✅ Completado (DEC-020)

### 🟢 Baja Prioridad (Nice-to-Have)

#### 4. `MSG_EXTRA.BD_DESCARGADA` no se usa en ningún lado

- **Archivo:** `src/schema-config.js`
- **Descripción del problema:** El mensaje `BD_DESCARGADA: 'Base de datos descargada correctamente'` está definido pero no se referencia en `descargarBDError()` ni en ninguna otra función.
- **Fix sugerido:** Añadir `alert(MSG_EXTRA.BD_DESCARGADA)` al final de `descargarBDError()` después de crear el download, o eliminar la entrada si no es necesaria.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

#### 5. `BACKUP.MAX_COPIAS` hardcodeado como 5, podría ser configurable por usuario

- **Archivo:** `src/schema-config.js`
- **Descripción del problema:** El límite de 5 copias de backup está fijo. En entornos con poco disco, el usuario podría preferir 3; en entornos críticos, 10.
- **Fix sugerido:** Mover a localStorage con fallback: `const maxCopies = parseInt(localStorage.getItem('BACKUP_MAX_COPIES') || BACKUP.MAX_COPIES, 10)`. Opcional: añadir UI para configurar.
- **Esfuerzo:** Medio
- **Estado:** ✅ Completado (DEC-020)

#### 6. No hay validación de que `SCHEMA_CONFIG.VERSION` sea número entero

- **Archivo:** `src/index.html`
- **Línea:** ~724
- **Descripción del problema:** Si `SCHEMA_CONFIG.VERSION` está mal definido (string, null), la comparación `version !== SCHEMA_CONFIG.VERSION` puede fallar silenciosamente.
- **Fix sugerido:** Añadir validación al inicio: `if (typeof SCHEMA_CONFIG.VERSION !== 'number') throw new Error('SCHEMA_CONFIG.VERSION debe ser número')`.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

#### 7. `execSafe()` no registra errores en bitácora interna

- **Archivo:** `src/index.html`
- **Línea:** ~1010-1025
- **Descripción del problema:** La función `execSafe()` muestra alertas pero no guarda un log interno de errores de SQL para debugging posterior (ej.: array `window.__sqlErrors`).
- **Fix sugerido:** Añadir `if (DEBUG.isEnabled) { window.__sqlErrors = window.__sqlErrors || []; window.__sqlErrors.push({sql, params, err, ts: Date.now()}); }` en el catch.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

---

## Propuestas de mejora

### 1. Centralizar queries SQL en `SCHEMA_CONFIG.queries` (ya iniciado)

- **Justificación:** Actualmente `SCHEMA_CONFIG.queries` tiene `rutaProcesos`, `documentosPendientes`, y `reporteExcel`. Pero hay otras consultas repetidas o embebidas en funciones (ej.: SELECT de expedientes en `cargarDatos()`).
- **Beneficio:** SPOT total de SQL, más fácil auditar y optimizar queries.
- **Esfuerzo:** Medio
- **Estado:** ✅ Completado (DEC-020)

### 2. Añadir `CONFIG.AUTOSAVE_ENABLED` para controlar autoguardado

- **Justificación:** El autoguardado cada 30s está siempre activo. Algunos usuarios podrían preferir desactivarlo y guardar manualmente.
- **Beneficio:** Flexibilidad para usuarios avanzados, menor desgaste de disco en equipos antiguos.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

### 3. Crear función `renderBadgeEstatus(estatus)` como SPOT único

- **Justificación:** Actualmente `getEstatusClass()` delega en `SCHEMA_CONFIG.estatusClass()`, pero el HTML del badge (`<span class="px-2 py-0.5...">`) está duplicado en `renderizarTabla()` y posiblemente en otros lugares.
- **Beneficio:** Si cambia el diseño de badges, se actualiza en un solo punto.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

### 4. Documentar en `funciones.md` las funciones nuevas de la auditoría (#59-#65)

- **Justificación:** `funciones.md` debe ser el SPOT de todas las funciones. Tras la auditoría de código limpio, se añadieron funciones como `validarArchivoBD()`, `obtenerRutaProcesos()`, `obtenerDocumentosPendientes()`, `descargarBDError()`, `updateUIOnError()`, `optimizarBD()`.
- **Beneficio:** DRY: la IA verificará `funciones.md` antes de crear funciones duplicadas.
- **Esfuerzo:** Bajo
- **Estado:** ✅ Completado (DEC-020)

### 5. Añadir test de humo en `index.html` al cargar

- **Justificación:** Verificar que todos los elementos del DOM referenciados en `SELECTORS` existen. Si falta uno, fallar temprano con mensaje claro.
- **Beneficio:** Detectar errores de tipeo en IDs inmediatamente, no en tiempo de ejecución.
- **Esfuerzo:** Medio
- **Estado:** ✅ Completado (DEC-020)

---

## Resumen de estado

| Categoría | Cantidad | Estado |
|-----------|----------|--------|
| Pendientes de `doc.md` | 7 | ✅ 100% completados |
| Hallazgos Alta Prioridad | 0 | ✅ Sin bloqueantes |
| Hallazgos Media Prioridad | 3 | ✅ Completados |
| Hallazgos Baja Prioridad | 4 | ✅ Completados |
| Propuestas de mejora | 5 | ✅ Implementadas |

**Próximo sprint recomendado:** No hay pendientes activos. Evaluar nuevas features o refactors según necesidad.

---

## DEC-019: Auditoría de Código Julio 2026 — Plan de Modificaciones Actualizado

- **Origen:** `[Iniciativa de la IA]`
- **Contexto y Causa:** Se realizó una auditoría completa del código consolidado en `combined.txt` (3640 líneas) para verificar el cumplimiento de las normas de código limpio y el estado de los pendientes documentados en `doc.md`. La auditoría confirmó que todos los pendientes históricos están completados y las violaciones críticas fueron resueltas en la iteración anterior (#59).
- **Alternativas evaluadas:**
  - No realizar auditoría — descartado: sin revisión periódica, el código tiende a degradarse con nuevas features.
  - Auditoría automatizada con linter — descartado: el proyecto es 100% cliente-side sin build step, las reglas son específicas del proyecto (SPOT, SoC, anti-hardcodeo) que un linter genérico no detecta.
- **Impacto:**
  - `plan_modificaciones.md`: creado con 3 hallazgos de Media Prioridad, 4 de Baja Prioridad, y 5 propuestas de mejora.
  - `decisiones.md`: este registro ADR para trazabilidad.
  - Ningún cambio en código fuente (solo documentación).

**Hallazgos principales:**
1. ✅ Todos los pendientes de `doc.md` están completados (7/7).
2. ✅ Violaciones críticas de la auditoría anterior fueron fixeadas (CONFIG, MSG, SELECTORS, DEBUG, SoC SQL/UI).
3. ✅ Tres mejoras de Media Prioridad implementadas (formateo de mensajes, error reporting en VACUUM, feedback visual en error crítico).
4. ✅ Cuatro mejoras de Baja Prioridad implementadas (mensaje no usado, backup configurable, validación de VERSION, logging de errores SQL).
5. ✅ Cinco propuestas de mejora implementadas (centralizar queries, controlar autoguardado, SPOT de badges, actualizar funciones.md, test de humo).
