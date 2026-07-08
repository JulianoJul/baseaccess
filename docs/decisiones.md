# Architecture Decision Records (ADR)

Registro cronológico de decisiones técnicas tomadas en el proyecto.

---

## DEC-001: Migración de Rust Desktop a Web (HTML + sql.js + Tailwind CSS)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** La app original era desktop Rust (GTK/Relm4). Se migró a web cliente-side para eliminar dependencias de compilación cruzada y permitir ejecución en cualquier SO sin binarios nativos.
- **Alternativas evaluadas:**
  - Tauri (Rust backend + web frontend) — descartado por complejidad de build en Termux ARM64.
  - Electron puro — elegido como capa de empaquetado, pero la app corre igual en navegador.
  - IndexedDB — descartado porque los datos ya existen en archivos `.db` SQLite.
- **Impacto:** Reescritura completa de `index.html`, creación de `vendor/` con sql.js WASM, Tailwind CSS, Font Awesome. Flujo offline-first.

---

## DEC-002: Límite de 100MB en Drag & Drop

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Archivos SQLite grandes (>100MB) saturan el heap de WASM y congelan el hilo principal del navegador. Se definió este límite como guarda en el evento `drop` y `change` del file input.
- **Alternativas evaluadas:**
  - Carga asíncrona con streaming — no es posible con sql.js (requiere `Uint8Array` completo).
  - Sin límite — descartado por riesgo de crash silencioso.
- **Impacto:** Validación en `index.html` con alerta al usuario si supera el límite.

---

## DEC-003: Dependencias Locales (vendor/) en lugar de CDN

- **Origen:** `[Suposición/Iniciativa de la IA]`
- **Contexto y Causa:** Las redes corporativas bloquean CDNs y la app debe funcionar sin internet. Se descargaron e incluyeron localmente Tailwind CSS, sql.js WASM y Font Awesome.
- **Alternativas evaluadas:**
  - CDN con fallback local — más complejo, beneficio marginal.
  - Bundler (webpack/vite) — no justificado para un solo HTML.
- **Impacto:** Creación de `vendor/` (~700KB), todos los `<link>` y `<script>` apuntan a rutas relativas.

---

## DEC-004: Electron win-unpacked sobre portable .exe

- **Origen:** `[Suposición/Iniciativa de la IA]`
- **Contexto y Causa:** El build portable single-file (.exe auto-contenido) usa NSIS + 7zip, que falla en Termux ARM64 (emulación x86 inestable). `win-unpacked` es una carpeta sin empaquetar que se copia directamente.
- **Alternativas evaluadas:**
  - `--win portable` — descartado por fallos de build.
  - `--win nsis` — requiere instalador, no es portable.
- **Impacto:** `package.json` config produce `dist/win-unpacked/`. Usuario copia carpeta a Windows y ejecuta `GestionExpedientes.exe`.

---

## DEC-005: File Input Nativo sobre IPC para Abrir BD

- **Origen:** `[Suposición/Iniciativa de la IA]`
- **Contexto y Causa:** En la primera versión se usó IPC (`dialog.showOpenDialog`), pero fallaba en ciertos entornos Windows (sin focus, `getWindow()` nulo). El `<input type="file">` es un estándar web que funciona siempre, con `file.path` como propiedad nativa de Chromium.
- **Alternativas evaluadas:**
  - `dialog.showOpenDialog` vía IPC — descartado por inestabilidad.
  - Drag & drop only — no cubre el caso "Abrir BD" desde el menú.
- **Impacto:** El botón "Abrir BD" dispara un `<input type="file" class="hidden">`. La ruta se sincroniza con `electronAPI.setDbPath()`.

---

## DEC-006: Snapshot Completo en Historial (vs Diff)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Inicialmente el historial almacenaba solo diferencias (columnas cambiadas), lo que hacía imposible reconstruir el estado exacto de un expediente en un momento dado. Se cambió a snapshot completo (34 columnas) en cada UPDATE vía trigger.
- **Alternativas evaluadas:**
  - Diff-based (solo columnas modificadas) — descartado: no permite reconstrucción fiel.
  - Subformulario de edición con historial inline — descartado por complejidad y bugs (fix #41).
- **Impacto:** Trigger `trg_exp_auditoria` sin WHEN condicional, tabla `historial_movimientos` con todas las columnas de `expedientes`.

---

## DEC-007: schema-config.js — Cero Hardcodeo del Schema

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El usuario detectó que `index.html` tenía strings literales del schema (nombres de columnas, catálogos, formato de observaciones). Se creó `schema-config.js` como fuente única de configuración específica del schema.
- **Alternativas evaluadas:**
  - Mantener constantes en `index.html` — descartado por violación DRY y difícil mantenimiento.
  - Tabla `app_config` en SQLite — la config incluye funciones JS (ej. `generarObservacion()`), no solo datos.
- **Impacto:** `index.html` refactorizado: `CATALOGO_POR_SELECT`, `CAMPOS_EDICION_FRECUENTE`, `COLS`, `generarObservacion()`, `getEstatusClass()` → todo referencias a `SCHEMA_CONFIG`.

---

## DEC-008: Observaciones de Una Línea (Sin Acumulación)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El usuario reportó que las observaciones acumulaban líneas infinitamente en cada edición. Se cambió a reemplazo completo: una sola línea auto-generada con estatus, documento y fechas. Si el usuario escribe texto libre, se extrae con `extractFreeText()` y se recoloca a la derecha al regenerar.
- **Alternativas evaluadas:**
  - Append-only con separador — descartado: el usuario quería limpieza, no acumulación.
  - Mantener `_obsPrevia` — descartado por acumulación excesiva (#49).
- **Impacto:** `observaciones` columna TEXT en BD, `previewObservacion()` reescrita, `extractFreeText()` creada.

---

## DEC-009: Notas como Columna Separada de Observaciones

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Las notas libres del usuario se mezclaban con las observaciones auto-generadas. Se añadió `notas TEXT` como columna separada en `expedientes` e `historial_movimientos`, y un textarea dedicado en el formulario y detalle.
- **Alternativas evaluadas:**
  - Un solo campo observaciones con texto libre al final — descartado: difícil de separar y parsear.
  - Tabla separada `notas` con FK — sobreingeniería para este caso de uso.
- **Impacto:** Schema v8 (`Tablas8.sql`): columnas `observaciones` y `notas`. Frontend: `f-notas` textarea, tarjeta NOTAS en desplegable.

---

## DEC-010: Switch a SQLite WASM (sql.js) sobre Rust + rusqlite

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** La versión original usaba Rust con rusqlite para acceso a BD. La migración a web requirió sql.js (sqlite compilado a WASM) que carga el mismo archivo `.db` sin modificaciones.
- **Alternativas evaluadas:**
  - sql.js (SQLite WASM) — elegido: mismo formato de archivo, misma SQL, sin migración de datos.
  - IndexedDB — descartado: requería migración desde .db.
  - SQLite por HTTP (backend) — descartado: la app debe ser 100% offline.
- **Impacto:** `vendor/sql-wasm.js` + `vendor/sql-wasm.wasm`. Toda la lógica de BD usa `db.exec()`, `db.run()` con `sanitizeNull` y `toInt()`.

---

## DEC-011: Sidebar de Frecuentes con localStorage

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El usuario pidió acceso rápido a expedientes frecuentes sin recargar. Se implementó sidebar colapsable con persistencia en localStorage (estrella en tabla para marcar/desmarcar).
- **Alternativas evaluadas:**
  - Tabla `app_config` en BD — localStorage es más simple y no requiere schema.
  - SessionStorage — no persiste entre sesiones.
- **Impacto:** `index.html`: sidebar HTML, lógica de toggle y persistencia, búsqueda sticky.

---

## DEC-012: Toggle de Orden en Edición (Secciones vs Excel)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El formulario de edición agrupa campos por secciones lógicas, pero el usuario quería poder verlos en el mismo orden que aparecen en el Excel original. Se añadió un botón toggle que clona los wrappers en una grilla plana siguiendo `SCHEMA_CONFIG.ordenExcel`.
- **Alternativas evaluadas:**
  - Reordenar los campos del DOM directamente — más frágil.
  - Dos formularios separados — duplicación de HTML.
- **Impacto:** `schema-config.js`: nuevo campo `ordenExcel`. `index.html`: función toggle en cabecera del modal de edición.

---

## DEC-013: Ruta Procesos y Documentos Pendientes como Modales

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El usuario quería dos vistas auxiliares: historial de ruteo de procesos (con emisor, receptor, estatus, fechas) y listado de expedientes pendientes de firma. Se implementaron como modales reutilizando el mismo patrón de tabla que la vista principal.
- **Alternativas evaluadas:**
  - Páginas separadas (SPA routing) — sobreingeniería para dos vistas simples.
  - Secciones expandibles en la página principal — menos visibles.
- **Impacto:** `index.html`: botones en header + modales con consultas SQL dedicadas.

---

## DEC-014: Error Boundary Global + Backup Rotativo + VACUUM + PRAGMA user_version

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El usuario identificó cuatro riesgos críticos en la sesión del 06/07/2026: corrupción de BD al escribir, schema desactualizado, errores JS congelando la UI, y crecimiento del archivo .db sin compactación. Se documentaron como normas de desarrollo en doc.md.
- **Alternativas evaluadas:**
  - N/A — son normas nuevas a implementar, no decisiones tomadas.
- **Impacto:** `doc.md`: nueva sección "Normas de Desarrollo / Buenas Prácticas". Próximos cambios en `main.js`, `index.html`, `Tablas8.sql`.

---

## DEC-015: Auditoría de Código Limpio — Centralización de Constantes

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El plan de auditoría (plan_modificaciones.md) identificó 12 violaciones en el código: números mágicos, console.log sueltos, strings literales en alertas, localStorage keys hardcodeadas, selectores DOM repetidos, SQL mezclado con UI, y `generarObservacion()` acoplada al DOM. Se resolvieron creando constantes globales en `schema-config.js`.
- **Alternativas evaluadas:**
  - Mantener las constantes en `index.html` — descartado por violación a SPOT y schema-config.js como fuente única.
  - Archivo separado `constants.js` — descartado: generar otro archivo para 15 constantes es over-engineering.
- **Impacto:**
  - `schema-config.js`: nuevas secciones `CONFIG`, `DEBUG`, `MSG`, `STORAGE_KEYS`, `SELECTORS`.
  - `index.html`: `$` helper reemplaza `document.getElementById`. Todas las alertas, console.log, localStorage keys y números mágicos referencian constantes.
  - `main.js`: console.log envueltos en `DEBUG.isEnabled`.
  - `generarObservacion()` ahora recibe parámetros en lugar de leer el DOM.
  - Nuevas funciones data layer: `obtenerRutaProcesos()`, `obtenerDocumentosPendientes()`, `validarArchivoBD()`.
   - `captureAndRestoreFormState()` hecho async.

---

## DEC-016: Implementación de las Cuatro Normas Críticas (VACUUM, Backup, Error Boundary, PRAGMA)

- **Origen:** `[Instrucción Explícita del Usuario desde DEC-014]`
- **Contexto y Causa:** Las normas críticas documentadas en DEC-014 requerían implementación concreta. Se añadieron los mecanismos de protección faltantes.
- **Alternativas evaluadas:**
  - N/A — implementación directa de lo acordado.
- **Impacto:**
  - `main.js`: nueva función `crearBackupRotativo()` con rotación de hasta 5 copias (`archivo.db.bak.1`..`.bak.5`), llamada antes de cada `save-db`.
  - `index.html`: botón "Compactar" en header que ejecuta `VACUUM`, error boundary modal (`#modal-error-critico`) activado por `window.onerror` + `window.onunhandledrejection`, y funciones `optimizarBD()` y `descargarBDError()`.
  - `schema-config.js`: nuevos selectores (`BTN_VACUUM`, `MODAL_ERROR`, `ERROR_CONTENIDO`, `BTN_DESCARGAR_BD`), mensajes `MSG_EXTRA` y constante `BACKUP`.
  - `Tablas8.sql`: añadido `PRAGMA user_version = 8`.

---

## DEC-017: MSG_EXTRA — Mensajes Fuera del Flujo Principal

- **Origen:** `[Iniciativa de la IA]`
- **Contexto y Causa:** `MSG` contenía solo mensajes del flujo principal (alertas de UI). Los mensajes de VACUUM, error boundary y otras utilidades de mantenimiento no pertenecían ahí. Se creó `MSG_EXTRA` como espacio separado.
- **Alternativas evaluadas:**
  - Fusionar todo en `MSG` — mezcla responsabilidades, viola SoC.
  - Strings literales — viola SPOT.
- **Impacto:** `schema-config.js`: nueva sección `MSG_EXTRA` con 6 entradas (VACUUM inicio/completo/error, error crítico, promesa rechazada, BD descargada).

---

## DEC-018: Gestión de Versionado del Schema con PRAGMA user_version

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Sin un versionado, cambios en `Tablas8.sql` podían dejar la BD del usuario en un schema antiguo sin detección. Se implementó validación al cargar la BD comparando `PRAGMA user_version` contra `SCHEMA_CONFIG.VERSION`.
- **Alternativas evaluadas:**
  - Sistema de migraciones con tabla `_schema_version` y scripts UP/DOWN — sobreingeniería para una app cliente-side con un solo schema.
  - Sin control — riesgo alto de errores silenciosos.
- **Impacto:** `Tablas8.sql`: `PRAGMA user_version = 8` al final. `schema-config.js`: `VERSION: 8` en `SCHEMA_CONFIG`. `_cargarBaseDatosComun()` valida al cargar y muestra alerta si no coincide.

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

---

## DEC-020: Implementación Completa del Plan de Modificaciones + Header Rework

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Se implementaron los 12 items del `plan_modificaciones.md` (3 Media, 4 Baja, 5 Propuestas) y se rediseñó el header de la UI para mejor usabilidad.
- **Alternativas evaluadas:**
  - N/A — implementación directa de lo planificado y solicitado por el usuario.
- **Impacto:**
  - `src/schema-config.js`: añadidos `BYTES_PER_MB`, `AUTOSAVE_ENABLED`, `STORAGE_KEYS.BACKUP_MAX_COPIES`, `SELECTORS.ESTADO_BD`, queries `expedientesSelect`/`expedientePorId`.
  - `src/index.html`: header rework (hamburguesa ☰, selector de orden en header, sidebar oculta por defecto), `renderBadgeEstatus()` SPOT, smoke test SELECTORS, error badge en `updateUIOnError()`, `MSG_EXTRA.BD_DESCARGADA` usado, `MSG_EXTRA.VACUUM_ERROR` en catch, `obtenerMaxBackups()` para backup configurable.
  - `main.js`: backup rotativo ahora usa `backupMaxCopies` variable (configurable vía IPC `set-backup-copies`/`get-backup-copies`).
  - `src/preload.js`: expone `setBackupCopies`/`getBackupCopies` en `electronAPI`.
  - `docs/funciones.md`: actualizado con nuevas funciones y constantes.
  - `docs/doc.md`: changelog items 66-67.

---

## DEC-021: Corrección VACUUM — Solo Botón Visual Eliminado, Código Preservado

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** En DEC-020 se eliminó por completo el botón VACUUM y la función `optimizarBD()`. El usuario aclaró que solo debía eliminarse el botón visual, conservando el código de la función para uso programático futuro.
- **Alternativas evaluadas:**
  - Eliminar todo — descartado: el usuario quería conservar el código.
  - Dejar ambos — descartado: el botón visual no debía estar.
- **Impacto:**
  - `src/index.html`: `optimizarBD()` restaurada (línea 983), botón `<button id="btn-vacuum">` no se reintroduce.
  - `src/index.html`: borde CSS eliminado del botón `btn-modo-orden` (Orden Excel/Secciones) por solicitud del usuario.
  - `src/schema-config.js`: `BTN_VACUUM`, `MSG_EXTRA.VACUUM_*` se mantienen como referencias válidas para la función.
  - `docs/doc.md`: item 68 corregido, item 69 añadido.
  - `docs/ai-context.md`: estado actual actualizado.

