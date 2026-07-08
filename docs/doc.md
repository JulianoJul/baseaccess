# Gestión de Expedientes con Historial — Documentación

> **Ver también:** [`decisiones.md`](decisiones.md) — ADR con historial de decisiones técnicas.
> **Anchor IA:** [`ai-context.md`](ai-context.md) — stack, líneas rojas, estado actual (lee esto primero).
> **Catálogo:** [`funciones.md`](funciones.md) — SPOT de funciones (DRY: verificar antes de crear).

## Contexto Termux (Android)

Este proyecto se edita y construye desde **Termux** en Android. Si inicias una sesión nueva:

| Ítem | Valor |
|------|-------|
| Directorio | `/storage/emulated/0/baseaccess` |
| Repositorio | `git@github.com:JulianoJul/baseaccess.git` |
| Node.js | `pkg install nodejs` (si no está) |
| Descargas | `curl` viene preinstalado |

**Comandos clave para reconstruir el `.exe` (solo en Termux/Android):**
```bash
npm install --save-dev --no-bin-links electron@latest electron-builder@latest
node node_modules/electron-builder/cli.js --win dir --x64
```
El build se genera en `dist/win-unpacked/`. Copiar esa carpeta a USB y ejecutar `GestionExpedientes.exe`.

> **Nota:** En Linux de escritorio (Arch, Ubuntu, etc.) usar `make electron-build-linux` o `npm run build:linux` para generar el AppImage.

**Importante:** `node_modules/` y `dist/` no se suben a git (`.gitignore`). Hay que reinstalar dependencias cada sesión nueva.

## ⚠️ Limitación: `file://` + WASM

Al abrir `src/index.html` con doble click (`file://` protocol), los navegadores **bloquean la carga del binario WASM** por seguridad. Síntomas:
- El botón "+ Nuevo Expediente" queda deshabilitado
- Los registros de la BDD no se muestran en la tabla

**Usar siempre Electron WinUnpacked** (`dist/win-unpacked/GestionExpedientes.exe`) para evitar este problema.

## Arquitectura

App web 100% cliente-side. **HTML + Tailwind CSS = UI** | **sql.js (SQLite WASM) = Data Layer**.
Sin backend, sin servidor, sin runtime externo. Un solo archivo HTML.

Dos modos de ejecución:

1. **Navegador** — abrir `src/index.html` directo (dependencias locales en `src/vendor/`)
2. **Electron WinUnpacked** — `GestionExpedientes.exe` con Chromium embebido (sin depender de Firefox/Chrome)

```
┌─────────────────────────────────────────────────┐
│  Modo Navegador (Firefox/Chrome/Edge)             │
│  ├── index.html                                  │
│  │   ├── vendor/tailwind.min.css — UI            │
│  │   ├── vendor/sql-wasm.js — SQLite WASM loader │
│  │   ├── vendor/sql-wasm.wasm — Motor SQLite     │
│  │   └── JavaScript — lógica CRUD                │
│  └── Archivo .db / .sqlite (cargado por usuario) │
├─────────────────────────────────────────────────┤
│  Modo Electron (win-unpacked, sin instalación)    │
│  ├── GestionExpedientes.exe (Chromium + app)     │
│  └── resources/vendor/ (CSS, WASM, etc.)         │
└─────────────────────────────────────────────────┘
```

## Principio Fundamental

**Cero assumptions del schema.** Todo se genera dinámicamente analizando la BD al cargarla:
- Catálogos → selectores poblados con `cargarCatalogos()`
- Vistas → tabla basada en `vw_reporte_excel_contrataciones`
- Historial → consulta JOIN bajo demanda al expandir fila

## Flujo de Datos

```
Usuario → [Selecciona .db] → FileReader → Uint8Array → SQL.Database
                                                              │
                    ┌─────────────────────────────────────────┤
                    ▼                                         ▼
           cargarCatalogos()                          cargarDatos()
                    │                                         │
                    ▼                                         ▼
           poblarSelectores()                    vw_reporte_excel_contrataciones
           (12 catálogos)                        → renderizarTabla()
```

## Esquema de Colores

Tailwind CSS (dark mode personalizado):
- Fondo: `bg-gray-900` | Superficie: `bg-gray-800` | Bordes: `border-gray-700`
- Texto: `text-gray-100` | Secundario: `text-gray-400`
- Acento: `teal-400` (botones, encabezados) | `teal-600` (botón primario)
- Estados: `emerald-400` (adjudicado) | `amber-400` (presupuesto) | `red-700` (eliminar)

## Estructura del Proyecto

```
baseaccess/
├── src/                  # Código fuente
│   ├── index.html        # App completa (HTML + CSS + JS)
│   ├── schema-config.js  # Config específica del schema (catálogos, columnas, formato observaciones, estatus)
│   ├── preload.js        # contextBridge para IPC
│   └── vendor/           # Dependencias locales (sin CDN)
│       ├── tailwind.min.css # Tailwind CSS build estático (16KB, tree-shaken)
│       ├── sql-wasm.js      # sql.js loader
│       ├── sql-wasm.wasm    # Motor SQLite WASM (~600KB)
│       ├── styles.css       # Estilos adicionales
│       ├── fontawesome.min.css # Font Awesome Free
│       └── webfonts/         # Fuentes de iconos
├── data/                 # Archivos de datos
│   ├── sql/
│   │   └── Tablas8.sql   # Schema SQLite v8 (actual)
│   ├── importar_datos.py # Script de importación desde Excel (openpyxl)
│   └── *.db              # Bases de datos (gitignored)
├── docs/                 # Documentación
│   ├── doc.md            # Documentación + pendientes + changelog
│   ├── decisiones.md     # ADR: Architecture Decision Records
│   ├── ai-context.md     # Anchor file para IAs (stack, líneas rojas, estado actual)
│   └── funciones.md      # Catálogo SPOT de funciones (DRY)
├── main.js              # Electron main process (ventana 1400x900)
├── package.json         # Electron + electron-builder config
├── prompt               # Prompt para Qwen Coder (planificador)
├── .clinerules           # Skill de Opencode (protocolo de modificación)
├── combined.txt         # Consolidado para auditorías (make combine)
├── Makefile             # combine / clean / commit / push / github / serve
├── .gitignore           # node_modules/, dist/, *.db
├── dist/                # Builds de Electron (AppImage, .deb, win-unpacked)
└── node_modules/        # Dependencias (gitignored)
```

## Tablas del Schema (Tablas8.sql)

| Tabla | Propósito |
|-------|-----------|
| `cat_gerencia` | Catálogo de gerencias |
| `cat_superintendencia` | Catálogo de superintendencias (FK → gerencia) |
| `cat_documento` | Tipos de documento (28 registros) |
| `cat_plan_contratacion` | Planes de contratación |
| `cat_modalidad` | Modalidades de contratación |
| `cat_art` | Artículos de normativa interna |
| `cat_tipo_contrato` | Tipos de contrato (PU, SG, MIXTO) |
| `cat_estatus_detalle` | Estatus (Pendiente, Firmado, Devuelto...) |
| `cat_resultado_proceso` | Resultados (Adjudicado, Desierto...) |
| `cat_empresas` | Empresas adjudicadas |
| `cat_responsables` | Emisores/Receptores |
| `cat_estado_accion` | Estado acción (Firma, Modificación, Recibo) |
| `expedientes` | **Tabla principal**: ~30 columnas con fechas, montos, FK, observaciones, notas |
| `historial_movimientos` | Traza de cambios: INSERT automático vía trigger |
| `vw_reporte_excel_contrataciones` | Vista JOIN completo para reportes |

## Dependencias Locales (vendor/)

Para evitar CDNs y funcionar sin internet, todo está en `src/vendor/`:

| Archivo | Fuente | Tamaño |
|---------|--------|--------|
| `tailwind.min.css` | Tailwind CSS v3.4.19 (JIT build, solo clases usadas) | ~16KB |
| `sql-wasm.js` | sql.js v1.8.0 | ~51KB |
| `sql-wasm.wasm` | sql.js WASM binary | ~600KB |

Regenerar `tailwind.min.css` si se agregan nuevas clases:
```bash
npm install --save-dev --no-bin-links tailwindcss@3.4.19
# crear tailwind.config.js apuntando a index.html
npx tailwindcss -i input.css -o src/vendor/tailwind.min.css --minify
```

## Electron WinUnpacked

Para no depender de ningún navegador, se construye `dist/win-unpacked/` con Chromium embebido.

### Source files
- `main.js` — Electron main process (ventana 1400x900, sin menú)
- `src/preload.js` — contextBridge para IPC seguro
- `src/index.html` — UI de la aplicación
- `package.json` — `electron` + `electron-builder` como devDeps

### Build (requiere Node.js + npm)

**Windows (desde Termux/Android):**
```bash
make electron-build-win
# o directamente:
npm run build
```

**Linux (AppImage):**
```bash
make electron-build-linux
# o directamente:
npm run build:linux
```

Carpeta `dist/win-unpacked/` (~360MB): copiar a Windows, ejecutar `GestionExpedientes.exe`. Sin instalación, sin admin.

> **Nota:** `--win portable` (single-file `.exe`) no se usa porque `win-unpacked` es más estable, permite reemplazar recursos sin re-empaquetar, y evita problemas con NSIS/7zip en Termux ARM64.

## Makefile

```bash
make combine          # Concatena src/index.html + src/schema-config.js + data/sql/Tablas8.sql + main.js + src/preload.js + package.json + docs/doc.md + docs/decisiones.md + docs/ai-context.md + docs/funciones.md + .clinerules → combined.txt
make clean            # rm -f combined.txt
make commit msg="x"   # git add -A + git commit
make push             # git push
make github msg="x"   # commit + push (shortcut)
make serve            # python3 -m http.server 8000 (kills old server, sirve src/index.html por HTTP para evitar file://)
make electron-build-win    # Build win-unpacked para Windows
make electron-build-linux  # Build AppImage para Linux
```

El schema usado en `make combine` se configura con `SCHEMA=data/sql/Tablas8.sql make combine` (por defecto usa `data/sql/Tablas8.sql`). También concatena `src/schema-config.js`.

## Reglas del Proceso

1. **ai-context.md + doc.md + decisiones.md primero**: `ai-context.md` orienta en 10 segundos (stack, líneas rojas). `doc.md` tiene pendientes y changelog. `decisiones.md` contiene el ADR con el porqué de cada decisión técnica.
2. **Makefile siempre**: después de cambios, ejecutar `make combine`.
3. **Sin hardcodeo**: cero assumptions de naming conventions. Toda heurística debe ser configurable.
4. **Historial de cambios**: cada cambio debe agregarse a la cronología en `doc.md` con fecha, archivo, y razón.
5. **DRY + Reutilización**: toda pieza de lógica debe tener una representación única. No repetir código ni copiar-pegar bloques. Si un patrón aparece en más de un lugar, extraer a función reutilizable. La modularidad no se mide en líneas por archivo ni por función, sino en ausencia de redundancia y en que cada función tenga una única responsabilidad (SRP). Una función de 200 líneas sin duplicación interna es mejor que 4 funciones de 50 líneas con lógica repetida.
6. **Commits estructurados**: toda confirmación debe incluir `RAZÓN TÉCNICA` y `SUPOSICIÓN` para trazabilidad de decisiones de la IA.

---

## Normas de Desarrollo / Buenas Prácticas

### 1. Backup Rotativo antes de cada `saveDb()`

**Riesgo:** Si ocurre un corte de energía, fallo del sistema o cierre forzado mientras Electron escribe el buffer en disco, el archivo `.db` se corrompe irremediablemente. La app escribe directamente sobre el .db en cada guardado, eliminación y autoguardado cada 30s.

**Norma:** Antes de invocar `saveDb()`, copiar el `.db` actual a una carpeta oculta con rotación de 3 backups (`.bak0`, `.bak1`, `.bak2`). Así, si el principal se daña, lo peor que se pierden son ~30s de trabajo.

**Archivos afectados:** `main.js` (IPC handler `save-db`), `src/preload.js` (exponer función de backup), o lógica en `src/index.html`.

### 2. Control de Versión del Schema via `PRAGMA user_version`

**Riesgo:** El schema se versiona externamente (Tablas6.sql → data/sql/Tablas8.sql), pero el frontend no valida la versión de la BD al cargarla. Un usuario podría cargar por error un archivo `.db` de una versión anterior, causando fallas silenciosas en vistas o triggers.

**Norma:** Asignar `PRAGMA user_version = 8;` al crear la BD en `Tablas8.sql` (o el script que la genere). Al cargar un archivo, el frontend ejecuta `SELECT pragma_user_version` y si no coincide con la esperada, muestra un cartel: *"Schema desactualizado: versión X, esperada Y. Resincroniza la BD."*

**Archivos afectados:** `data/sql/Tablas8.sql` (agregar PRAGMA), `src/index.html` (validación al cargar).

### 3. Error Boundary Global (window.onerror + unhandledrejection)

**Riesgo:** Como es SPA (index.html gestiona toda la UI y lógica), un error JS no controlado rompe el hilo de ejecución y congela la interfaz sin feedback al usuario.

**Norma:** Registrar `window.onerror` y `window.onunhandledrejection` al inicio del script. Ante un error crítico: mostrar modal elegante *"Algo salió mal"* con opción de **"Descargar BD actual"** (exportar buffer en memoria) para rescatar datos antes de recargar.

**Archivos afectados:** `src/index.html` (bloque de inicialización).

### 4. Mantenimiento de la BD (VACUUM)

**Riesgo:** SQLite no reduce el tamaño del archivo en disco al eliminar/actualizar registros; solo marca bloques como reutilizables. Con ediciones constantes, el archivo crece innecesariamente, alargando lecturas y transferencias IPC.

**Norma:** Añadir botón "Compactar BD" en la UI que ejecute `VACUUM;`. Opcional: ejecutar VACUUM automático al cerrar la app en Electron (evento `before-quit`).

**Archivos afectados:** `src/index.html` (botón + lógica), `main.js` (opcional, VACUUM en cierre).

---

### 5. SPOT — Single Point of Truth

Un dato o lógica debe existir en un solo lugar. Si cambia, se actualiza en ese único punto y el resto del sistema lo refleja automáticamente.

**Ejemplos en el proyecto:**
- `src/schema-config.js` es el SPOT para todo lo específico del schema (columnas, catálogos, formato de observaciones, estatus). `src/index.html` solo referencia `SCHEMA_CONFIG.*`.
- `CATALOGO_POR_SELECT` es el SPOT para los mapeos select→catálogo. `cargarCatalogos()` y `poblarSelectores()` iteran sobre él.
- `CONFIG.MAX_FILE_SIZE_BYTES` sería el SPOT para el límite de drag & drop, en vez del literal `104857600`.

**Violación detectada:** `if (file.size > 104857600)` en `src/index.html` — número mágico sin constante.

### 6. KISS — Keep It Simple, Stupid

El código debe ser lo más sencillo posible. Simple no es trivial: es la solución más directa que cumple el requerimiento sin over-engineering.

**Ejemplos en el proyecto:**
- Modales en vez de SPA routing para Ruta Procesos y Documentos Pendientes (DEC-013).
- localStorage en vez de tabla `app_config` en BD para la sidebar de frecuentes (DEC-011).
- `observaciones` de una sola línea en vez de append con separadores (DEC-008).

**Contraejemplo a evitar:** Un sistema de migraciones con versionado complejo cuando `PRAGMA user_version` + un `if` alcanza.

### 7. Evitar Números/Textos Mágicos

Los valores literales sin nombre (hardcodeados) se llaman "mágicos" porque su significado no es evidente. Se solucionan asignándolos a constantes con nombre descriptivo.

| Mal | Bien |
|-----|------|
| `if (file.size > 104857600)` | `if (file.size > CONFIG.MAX_FILE_SIZE_BYTES)` |
| `if (edad > 18)` | `if (edad > EDAD_MINIMA_VOTAR)` |
| `toast("Guardado exitoso")` | `toast(MSG_GUARDADO_EXITOSO)` |

**Regla:** En el proyecto, cero literales numéricos o strings de mensaje dentro de funciones de lógica. Todo debe estar definido como constante en un objeto `CONFIG` o al inicio del script.

### 8. YAGNI — You Aren't Gonna Need It

No programar funcionalidades hasta que sean estrictamente necesarias. Las IA tienden a agregar "por si acaso" (filtros avanzados, exportaciones múltiples, ordenamientos extra). Eso es código que se puede romper sin valor agregado.

**Ejemplo:** Si se pide un botón "Abrir BD", no programar también sistema de recientes, drag & drop con validación y selector de último archivo. Eso se agrega cuando se pide explícitamente.

**Regla:** Resolver única y exclusivamente lo solicitado en el prompt actual. No asumir necesidades futuras.

### 9. SoC — Separation of Concerns

Separar estrictamente la lógica de acceso a datos (SQL/SQLite) de la lógica de renderizado (DOM/HTML). Ninguna función de UI debe construir queries SQL directamente.

**Ejemplo en el proyecto:**
- `dbToObjects()` aísla la ejecución SQL → datos planos.
- `renderizarTabla()` solo recibe datos y pinta filas.
- Las funciones que construyen INSERT/UPDATE están separadas de las que manipulan el DOM del modal.

**Regla:** Una función que maneja eventos del DOM no debe tener strings SQL en su cuerpo.

### 10. Principio de Menor Sorpresa (Least Astonishment)

Las funciones deben ser predecibles y hacer una sola tarea asociada a su nombre. `obtenerExpediente()` solo obtiene y devuelve; no debe limpiar formularios ni modificar variables globales.

**Regla:** Cohesión alta: el nombre de la función debe describir exactamente lo que hace, sin efectos secundarios ocultos.

### 11. Cohesión Alta, Acoplamiento Bajo (High Cohesion, Low Coupling)

- **Alta cohesión:** Lo que está dentro de una función coopera para el mismo fin.
- **Bajo acoplamiento:** Si cambia la BD, el módulo que dibuja tablas no debe romperse.

**Regla:** La interfaz gráfica no debe importar ni conocer la estructura interna de las tablas SQL (eso está en `src/schema-config.js`).

---

## Cambios Realizados

### Migración a Web HTML/JS (Julio 2026)

| # | Archivo | Cambio | Razón |
|---|---------|--------|-------|
| 1 | `index.html` | **Creado**: app web completa con Tailwind CSS + sql.js | Migración de Rust desktop a web cliente-side |
| 2 | `Tablas6.sql` | **Creado**: schema v6 con historial_movimientos, trigger auditoría, 2 vistas, datos iniciales | Nueva versión del schema con trazabilidad |
| 3 | `prompt` | Reescrito: contexto web (index.html + Tablas6.sql), reglas HTML/JS | Reflejar el nuevo proyecto en las auditorías |
| 4 | `doc.md` | Reescrita: arquitectura web, dependencias, estructura, Tablas6.sql | Documentar el nuevo stack |
| 5 | `Makefile` | Simplificado: eliminados targets Rust, combine ahora concatena index.html + SQL + doc | Adaptado a proyecto web |
| 6 | `vendor/` | **Creado**: tailwind.min.css, sql-wasm.js, sql-wasm.wasm | Dependencias locales para funcionar sin CDN ni internet |
| 7 | `index.html` | CDNs reemplazadas por rutas locales `vendor/` | Offline-first: sin depender de CDNs corporativas bloqueadas |
| 8 | `main.js` + `package.json` | **Creado**: Electron main process + electron-builder config | App de escritorio portable sin depender del navegador |
| 9 | `.gitignore` | **Creado**: node_modules/, dist/ | Prevenir commits de dependencias y builds |
| 10 | `doc.md` | Agregada sección Contexto Termux + advertencia `file://` WASM | Documentar entorno de desarrollo y limitación conocida |
| 11 | `Makefile` | Agregado target `serve` (python3 http.server) | Alternativa HTTP para evitar bloqueo WASM en file:// |
| 12 | `Tablas7.sql` | **Creado** a partir de Tablas6.sql + columna `observaciones_generales` en `historial_movimientos` | Capturar snapshot de observaciones en cada movimiento |
| 13 | `Tablas7.sql` | Trigger `trg_exp_auditoria` actualizado para detectar cambios en `observaciones_generales` | Sincronizar con el nuevo campo |
| 14 | `index.html` | `toggleDesplegable` refactorizado: carga solo último movimiento, botón "Ver historial completo" para expandir | Click-to-expand historial en grilla |
| 15 | `index.html` | `cargarHistorialFormulario` refactorizado con mismo patrón click-to-expand | Consistencia entre grilla y modal |
| 16 | `index.html` | Agregadas `expandirHistorialCompleto`, `cargarUltimoMovimiento`, `expandirHistorialFormulario` | Lógica reutilizable para carga progresiva |
| 17 | `index.html` | Agregada `toggleDetalleMov` con detalle expandible por movimiento (incluye observaciones) | Ver detalle completo sin recargar |
| 18 | `Makefile` | combine target apunta a Tablas7.sql en lugar de Tablas6.sql | Reflejar schema actual |
| 19 | `doc.md` | Documentación actualizada: Tablas6.sql→Tablas7.sql | Sincronizar documentación con schema v7 |
| 20 | `index.html` | Eliminada columna "Monto Adjudicado" de la tabla principal + colspan 8→7 | Simplificar vista principal, monto visible solo en detalle expandible |
| 21 | `package.json` | Agregado script `build:linux`, sección `linux` con targets AppImage/deb, campo `author` | Build para Linux (AppImage generado) |
| 22 | `bdd/Tablas7.sql`, `index.html` | Eliminada UNIQUE constraint de `solped`, ahora permite texto libre (múltiples SOLPED) | Los expedientes pueden tener uno o varios números SOLPED |
| 23 | `.gitignore`, `Makefile`, `prompt`, `doc.md`, `data/sql/Tablas8.sql` | Reorganización del proyecto: SQL movidos a `data/sql/`, Makefile con `SCHEMA` variable y targets win/linux, prompt actualizado a Tablas8.sql, gitignore mejorado | Reflejar estructura actual y dar soporte multiplataforma |
| 24 | `index.html` | Agregado botón "📋 Historial" en detalle de expediente + modal con historial completo (todas las columnas del snapshot) | Acceder al historial completo sin perder el foco en observaciones |
| 25 | `index.html`, `prompt`, `doc.md` | Fix HIGH: eliminadas refs a `cat_estado_accion` y `id_estado_accion`, unificados toggles a `toggleDetalle(prefix, id)`, sanitizados IDs SQL con `toInt()`, renombrado `escapeSql`→`sanitizeNull`, eliminado `console.error` | Auditoría de código: cerrar hallazgos prioritarios |
| 26 | `index.html` | Agregada función `execSafe()` con manejo de errores y `toInt()` para validación de IDs | Prevención de SQL injection y errores silenciosos |
| 27 | `index.html`, `main.js` | Fix: `const electronAPI` → `var electronAPI` para evitar error de redeclaración en Electron + F12 abre DevTools | Debugging en producción |
| 28 | `index.html` | Fix: `formatNum()` detecta `typeof v === 'number'` y va directo a `toLocaleString`; `calcularBs()` pasa número sin `.toFixed(2)` | Bug de tipo de cambio: 32,5 USD × 123 = 3.997,50 (antes 39.975,00) |
| 29 | `vendor/styles.css`, `vendor/fontawesome.min.css`, `vendor/webfonts/`, `index.html` | Modularización del diseño: CSS extraído a `vendor/styles.css` con variables y clases reutilizables (.btn, .input, .card, .label, .legend, etc.); Font Awesome Free reemplaza todos los emojis por iconos vectoriales; fuentes incluidas en vendor/ para portabilidad | Diseño mantenible y portable sin emojis |
| 30 | `index.html` | P1: Resaltar campos de edición frecuente (Tablas7) con punto ámbar + `CAMPOS_EDICION_FRECUENTE` constante | Indicador visual en 10 campos de cambio frecuente |
| 31 | `index.html` | P2: Observaciones automáticas append-only con snapshot de edición, bloque colapsable de observaciones anteriores | No se pierde el historial al editar |
| 32 | `index.html` | P3: Botones "+" por campo para editor de validaciones (scaffold) + modal genérico | Preparación para reglas de validación |
| 33 | `index.html` | P5: FormatTiempoEjecucion — sufijo "DÍAS" automático al perder el foco si el valor es numérico | Consistencia en campo Tiempo Ejecución |
| 34 | `index.html` | P6: Nro. ejemplares del documento visible en formulario (junto al select) y en el detalle de la tabla | Dato faltante de cat_documento ahora visible en frontend |
| 35 | `index.html` | P7: Botón "Recientes" con menú desplegable y localStorage; en Electron reabre por path, en navegador abre picker | Acceso rápido a BD abiertas recientemente |
| 36 | `index.html`, `data/sql/Tablas8.sql` | P6: `nro_ejemplares` movido de `cat_documento` a `expedientes` como campo editable por registro. Eliminados `actualizarNroEjemplares()`, display en catálogo, y columna de schema en cat_documento | El nro. de ejemplares varía por expediente, no por tipo de documento |
| 37 | `index.html` | Fix: `calcularBs()` ahora calcula monto_adjudicado en ambos sentidos (USD→BS y BS→USD) cuando cambia tipo de cambio, independientemente del orden en que se llenen los campos | Bidireccionalidad completa en cálculo automático de montos |
| 38 | `index.html` | Fix: texto libre en observaciones ahora se guarda correctamente (save lee el textarea en lugar de solo generar línea automática). `actualizarObservacion()` append en vez de reemplazar para no perder escritura del usuario | El texto libre del usuario se perdía al guardar |
| 37 | `index.html` | Unificación DRY: `CATALOGO_POR_SELECT` como fuente única de verdad para catálogos. Expandido con campo `cols`. `cargarCatalogos()` y `poblarSelectores()` ahora iteran sobre `CATALOGO_POR_SELECT` eliminando los mapeos paralelos duplicados | Eliminar duplicación de 3 estructuras de datos que mapeaban select→catálogo (regla DRY del doc.md) |
| 38 | `index.html` | Fix: `captureAndRestoreFormState()` para preservar valores de selects al repoblar catálogos | Evitar que campos del formulario se vacíen al añadir nuevos registros a catálogos |
| 39 | `index.html` | Fix: Eliminado `e.stopPropagation()` del botón '+' de catálogo | El botón '+' no despliega menú al tocar el ícono exacto |
| 40 | `index.html` | Feature: tipo_cambio aplica automáticamente a monto_adjudicado_bs | Calcular monto adjudicado en BS al cambiar monto USD o tipo de cambio |
| 41 | `index.html`, `data/sql/Tablas8.sql` | Historial overhaul: subformulario eliminado en edición, trigger INSERT para snapshot inicial, observaciones con formato sin prefijos (solo valores), ficha muestra solo observación más nueva con expand, "ver historial completo" como tabla de snapshots | Reemplazar modelo de diferencias por snapshot completo desde creación |
| 42 | `index.html` | Bug ENOENT: `mostrarMenuRecientes()` usa data-attributes con `encodeURIComponent` + listener delegativo en lugar de inline onclick. Agregadas `escapeHtml()` y `eliminarRecienteIndex()` | Caracteres especiales en rutas de BD recientes causaban error al abrir |
| 43 | `index.html` | Botón "+" de catálogos: `pointer-events-none` en `<i>` + `preventDefault`/`stopPropagation` en onclick | Click en ícono no propagaba al botón |
| 44 | `index.html` | `captureAndRestoreFormState()` captura TODOS los elementos del formulario (inputs, textareas, selects) con restauración asíncrona + `guardarNuevoCatalogo()` repuebla solo el select afectado | Campos se vaciaban al añadir nuevo registro a catálogo |
| 45 | `data/sql/Tablas8.sql` | `historial_movimientos` ampliado a 34 columnas con snapshot completo. `trg_exp_auditoria` sin WHEN condicional (registra en todo UPDATE). Triggers incluyen solped, plan, modalidad, art, presupuesto_bs, monto_bs, descripción, nro_ejemplares, etc. | Snapshot incompleto no capturaba todos los campos del expediente |
| 46 | `index.html` | Botón "Abrir BD" se contrae a solo ícono al cargar base de datos | Liberar espacio horizontal cuando ya hay BD abierta |
| 47 | `data/sql/Tablas8.sql`, `index.html`, `data/importar_datos.py` | `observaciones_generales` → `observaciones`, añadida columna `notas TEXT`, eliminada columna `nro_ejemplares` de ambas tablas, triggers y vista | Separar observaciones auto-generadas de notas libres del usuario |
| 48 | `index.html` | Removido encabezado izquierdo ("Carga tu base de datos..."), añadida tarjeta NOTAS condicional en desplegable, ícono lupa verde en buscador | UI cleanup solicitado por usuario |
| 49 | `index.html` | `observaciones`: reemplazo de una sola línea (sin acumulación). Nueva `extractFreeText()` que resta partes auto-generadas del textarea para preservar solo el texto libre del usuario. `previewObservacion()` y `guardarExpediente()` ya no concatenan con `_obsPrevia`. | Evitar acumulación de líneas; texto libre se mantiene al regenerar la parte auto-generada |
| 50 | `index.html` | Añadida columna "Descripción" visible en tabla principal (8 columnas). Añadido selector de orden (Reciente/Fecha creación/Fecha modificación) con función `cambiarOrden()`. | Pendientes #7 y #8 |
| 51 | `index.html` | Añadidos botones "Ruta Procesos" (#4) y "Documentos Pendientes" (#5) en header. Modales independientes con tabla de ruteo y listado de pendientes de firma. | Pendientes #4 y #5 |
| 52 | `schema-config.js`, `index.html` | Creado `schema-config.js` con toda la configuración específica del schema (catálogos, columnas, formato de observaciones, colores de estatus). `index.html` refactorizado para usar `SCHEMA_CONFIG` en lugar de constantes/funciones hardcodeadas. | DRY + modularización; eliminar hardcodeo del schema en index.html (pendiente #1) |
| 53 | `index.html`, `schema-config.js` | Sidebar de documentos frecuentes colapsable + búsqueda sticky (#9). Toggle de orden de campos en edición (secciones / orden Excel) (#2) con `ordenExcel` en `schema-config.js`. | Pendientes #2 y #9 |
| 54 | `decisiones.md`, `prompt`, `doc.md`, `Makefile` | **Creado** `decisiones.md` con 14 ADR entries. `prompt` actualizado con ADR y normas de código limpio. `doc.md` referencias a `decisiones.md`. `Makefile` combine incluye `decisiones.md`. | Bitácora de decisiones técnicas para trazabilidad de arquitectura |
| 55 | `doc.md`, `prompt` | Agregados principios SPOT, KISS y anti-magic-numbers en `doc.md` (sección Normas de Desarrollo) y `prompt` (NORMAS DE CÓDIGO LIMPIO) | Formalizar principios de diseño que aplican al proyecto offline-first |
| 56 | `doc.md`, `prompt` | Agregados principios YAGNI, SoC, Least Astonishment, High Cohesion/Low Coupling en `doc.md` y `prompt` | Completar catálogo de principios de ingeniería de software para guiar a las IA |
| 57 | `ai-context.md`, `prompt`, `doc.md`, `Makefile` | **Creado** `ai-context.md` (anchor file IA: stack, líneas rojas, estado actual). `prompt` actualizado con formato de commits estructurado (RAZÓN TÉCNICA + SUPOSICIÓN). `Makefile` combine incluye `ai-context.md` | Pipeline nativo IA: anchor + commit logs + PAR integrados |
| 58 | `funciones.md`, `.clinerules`, `doc.md`, `prompt`, `Makefile`, `ai-context.md` | **Creados** `funciones.md` (catálogo SPOT con 58 funciones) y `.clinerules` (skill de Opencode con protocolo de modificación). Todos los archivos de contexto actualizados para referenciarlos | Cerrar el círculo DRY: la IA debe verificar funciones.md antes de escribir código nuevo |
| 59 | `schema-config.js`, `index.html`, `main.js`, `funciones.md`, `doc.md` | **Auditoría de código limpio**: CONFIG, DEBUG, MSG, STORAGE_KEYS, SELECTORS creados en schema-config.js. `$` helper añadido. Todas las alertas, console.*, localStorage keys y números mágicos reemplazados por referencias a constantes. `generarObservacion()` desacoplada del DOM. SQL extraído a data layer (`obtenerRutaProcesos`, `obtenerDocumentosPendientes`). `captureAndRestoreFormState` hecho async. Drag-drop validation extraída a `validarArchivoBD()`. DEBUG condicional en main.js + index.html | Fix de 12 hallazgos del plan_modificaciones.md (números mágicos, console.log, strings literales, localStorage keys, selectores, SoC SQL/UI, acoplamiento DOM) |
| 60 | `main.js` | **Backup rotativo**: nueva función `crearBackupRotativo()` con rotación de 5 copias (`archivo.db.bak.1`..`.bak.5`), llamada antes de cada `save-db`. Config `BACKUP` en schema-config.js | Riesgo crítico: proteger contra corrupción por corte de energía durante escritura |
| 61 | `index.html` | **Botón VACUUM (Compactar)** en header, ejecuta `db.run('VACUUM')` con reporte de tamaño antes/después. Deshabilitado hasta cargar BD | Mantenimiento de BD: SQLite no libera espacio en disco al eliminar/actualizar |
| 62 | `index.html` | **Error boundary global**: `window.onerror` + `window.onunhandledrejection` con modal `#modal-error-critico`, botón "Descargar BD actual" (`descargarBDError()`), y deshabilitación de botones de edición (`updateUIOnError()`) | Evitar UI congelada sin feedback; permitir rescatar datos en memoria |
| 63 | `schema-config.js` | Nuevos selectores (`BTN_VACUUM`, `MODAL_ERROR`, `ERROR_CONTENIDO`, `BTN_DESCARGAR_BD`), mensajes `MSG_EXTRA` (6 entradas para VACUUM y error boundary), y constante `BACKUP` | SPOT: centralizar todo en schema-config.js |
| 64 | `data/sql/Tablas8.sql` | Añadido `PRAGMA user_version = 8;` al final del archivo | Versionado de schema para validación al cargar BD |
| 65 | `decisiones.md` | Añadidos DEC-016 (VACUUM+Backup+Error+PRAGMA), DEC-017 (MSG_EXTRA), DEC-018 (PRAGMA user_version) | Trazabilidad de implementación de normas críticas |
| 66 | `src/schema-config.js`, `src/index.html`, `main.js`, `src/preload.js` | **Auditoría plan_modificaciones.md**: 12 items implementados — BYTES_PER_MB, VACUUM catch, error badge, BD_DESCARGADA, backup configurable vía localStorage+IPC, validación VERSION, log SQL errors, queries centralizadas, AUTOSAVE_ENABLED, renderBadgeEstatus(), smoke test SELECTORS | Cierre completo del plan |
| 67 | `src/index.html` | **Header rework**: botones alineados a la izquierda, hamburguesa (☰) togglea sidebar, selector de orden movido al header junto a hamburguesa, sidebar oculta por defecto | UX: sidebar no ocupa espacio si no se usa |
| 68 | `src/index.html` | Botón "Compactar" (VACUUM) eliminado visualmente del header; función `optimizarBD()` preservada | Se quitó solo el botón visual; el código de VACUUM se conserva para uso programático futuro |
| 69 | `src/index.html` | Borde eliminado del botón `btn-modo-orden` (Orden Excel/Secciones) en modal de edición | Limpieza visual: el botón toggle no necesita borde distintivo |

---

## Pendientes / Por Hacer

| # | Prioridad | Descripción | Archivos | Estado |
|---|-----------|-------------|----------|--------|
| — | 🟡 Media | ~~Archivo separado para ajustes de BD (opción A: tabla `app_config` en SQLite vs opción B: `db-settings.js`)~~ Reemplazado por `schema-config.js` | `schema-config.js` | **reemplazado** |
| — | 🟢 Baja | ~~Archivo de config específico para BDD (`bdd_config.json`)~~ Reemplazado por `schema-config.js` | — | **reemplazado** |
| 1 | 🟡 Media | **`schema-config.js`**: archivo JS aparte con constantes del schema (columnas, etiquetas, campos de edición frecuente, etc.) para no tenerlo hardcodeado en `index.html` | `schema-config.js`, `index.html` | **completado** |
| 2 | 🟡 Media | **Dos modos de orden en edición**: mantener el actual (campos agrupados por secciones) + agregar modo con el mismo orden que aparece en el Excel | `index.html`, `schema-config.js` | **completado** |
| 3 | 🟢 Baja | **Colores por frecuencia de edición**: color distinto para campos según qué tan frecuente se editan (1ra, 2da, 3ra vez, etc.) | `src/index.html`, `src/vendor/styles.css` | pendiente |
| 4 | 🟡 Media | **Menú Ruta Procesos**: botón que lleve a una pantalla distinta imitando el comportamiento del Excel | `src/index.html` | **completado** |
| 5 | 🟡 Media | **Botón Documentos Pendientes**: listado/modal con todos los expedientes cuyo estatus no sea FIRMADO | `src/index.html` | **completado** |
| 6 | 🔴 Alta | **Schemas separados para demás hojas del Excel**: cada hoja del Excel es un módulo independiente con su propio schema (ej. `Tablas8_hoja2.sql`), sin contaminar el schema principal | `data/sql/*.sql` | pendiente |
| 7 | 🟢 Baja | **Orden por fecha en pantalla principal**: ordenar tabla por `fecha_creacion` y `fecha_actualizacion` (independiente de los modos de orden del formulario de edición) | `src/index.html` | **completado** |
| 8 | 🟢 Baja | **Columna "descripción de proceso" visible** en la tabla principal (actualmente solo en el desplegable) | `src/index.html` | **completado** |
| 9 | 🟡 Media | **Sidebar de documentos frecuentes** (colapsable, arrastrar expedientes del usuario) + **barra de búsqueda sticky** (position: sticky al hacer scroll) | `src/index.html`, `src/vendor/styles.css` | **completado** |
| — | 🔴 Alta | **Backup rotativo automático**: copia el .db actual antes de cada escritura con rotación de 5 backups | `main.js`, `src/schema-config.js` | **completado** |
| — | 🔴 Alta | **PRAGMA user_version**: validación al cargar BD contra `SCHEMA_CONFIG.VERSION` | `data/sql/Tablas8.sql`, `src/schema-config.js`, `src/index.html` | **completado** |
| — | 🟡 Media | **Error boundary global**: `window.onerror` + `window.onunhandledrejection` con modal de rescate | `src/index.html` | **completado** |
| — | 🟡 Media | **Función VACUUM** (`optimizarBD()`) disponible, sin botón visual en header (eliminado por solicitud) | `src/index.html` | **completado** |

---
### Bug de persistencia resuelto (Electron)

Antes: sql.js modificaba la BD en RAM, nunca escribía al disco.
Ahora: se agregó `preload.js` + IPC handlers en `main.js` para leer/escribir archivos `.db`. Después de cada `guardarExpediente()` y `eliminarExpediente()`, se exporta el buffer de sql.js (`db.export()`) y se escribe al archivo `.db` vía `electronAPI.saveDb()`. Además hay autoguardado cada 30s, al cerrar la ventana, y atajo Ctrl+S.

### Apertura de Base de Datos (Electron)

El flujo de apertura usa **`<input type="file">` nativo del navegador** (no IPC), por confiabilidad:
1. El botón "Abrir Base de Datos" dispara un `<input type="file" id="dbfile" accept=".db,.sqlite" class="hidden">`.
2. El `change` event lee el archivo con `FileReader` → `Uint8Array` → `new SQL.Database(bytes)`.
3. La ruta del archivo se obtiene de `f.path` (propiedad nativa de Electron/Chromium para drag & drop y file input).
4. Se sincroniza con el backend vía `electronAPI.setDbPath(f.path)` para que `saveDb()` sepa dónde escribir.
5. Drag & drop: mismo flujo via `FileReader` + `file.path`.

**Por qué no IPC para abrir:** El `<input type="file">` es un estándar web que funciona siempre, sin depender de preload/contextBridge. En la primera versión se intentó con IPC (`pickDbFile` → `dialog.showOpenDialog`) pero fallaba en ciertos entornos (Windows sin focus, problemas con `getWindow()`).

### Rama `tauri-migration`

Existe la rama `tauri-migration` que reemplaza Electron por Tauri v2 (Rust). `master` queda intacto con Electron. Ver esa rama para los detalles de la migración.


### Análisis: Bug "Agregué un expediente y no se guardó"

Tras revisar el código de `guardarExpediente()` y el schema:

| Aspecto | Estado |
|---------|--------|
| **Cantidad columnas vs params** | ✅ Correcto (32 columnas, 32 placeholders, 32 params) |
| **Validación SOLPED vacío** | ✅ Alerta y detiene el guardado |
| **SOLPED UNIQUE** | ❌ El schema tiene `solped TEXT UNIQUE`. Si se intenta insertar un SOLPED ya existente, SQLite lanza `UNIQUE constraint failed`. El error se captura en el `catch` y se muestra en alert. |
| **`escapeSql` nullifica vacíos** | ✅ Correcto, SQLite acepta NULL en columnas sin NOT NULL |
| **Trigger `trg_exp_auditoria`** | ✅ Solo se ejecuta en UPDATE, no afecta INSERT |
| **Posible causa** | **SOLPED duplicado** es la causa más probable. También verificar que la BDD cargada no tenga `PRAGMA foreign_keys = ON` conflictivo con FKs sin valores. |
