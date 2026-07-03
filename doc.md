# Gestión de Expedientes con Historial — Documentación

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

Al abrir `index.html` con doble click (`file://` protocol), los navegadores **bloquean la carga del binario WASM** por seguridad. Síntomas:
- El botón "+ Nuevo Expediente" queda deshabilitado
- Los registros de la BDD no se muestran en la tabla

**Usar siempre Electron WinUnpacked** (`dist/win-unpacked/GestionExpedientes.exe`) para evitar este problema.

## Arquitectura

App web 100% cliente-side. **HTML + Tailwind CSS = UI** | **sql.js (SQLite WASM) = Data Layer**.
Sin backend, sin servidor, sin runtime externo. Un solo archivo HTML.

Dos modos de ejecución:

1. **Navegador** — abrir `index.html` directo (dependencias locales en `vendor/`)
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
├── index.html           # App completa (HTML + CSS + JS)
├── main.js              # Electron main process (ventana 1400x900)
├── package.json         # Electron + electron-builder config
├── vendor/              # Dependencias locales (sin CDN)
│   ├── tailwind.min.css # Tailwind CSS build estático (16KB, tree-shaken)
│   ├── sql-wasm.js      # sql.js loader
│   └── sql-wasm.wasm    # Motor SQLite WASM (~600KB)
├── bdd/                 # Schemas y bases de datos
│   ├── Tablas6.sql      # Schema SQLite v6 (legacy)
│   ├── Tablas7.sql      # Schema SQLite v7
│   ├── Tablas8.sql      # Schema SQLite v8 (actual)
│   └── si.db            # Base de datos de prueba
├── doc.md               # Esta documentación
├── prompt               # Prompt para auditorías (opencode)
├── combined.txt         # Consolidado para auditorías (make combine)
├── Makefile             # combine / clean / commit / push / github / serve
├── .gitignore           # node_modules/, dist/, *.db
└── dist/                # Builds de Electron (AppImage, .deb, win-unpacked)
```

## Tablas del Schema (Tablas7.sql)

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
| `expedientes` | **Tabla principal**: ~31 columnas con fechas, montos, FK, nro_ejemplares |
| `historial_movimientos` | Traza de cambios: INSERT automático vía trigger |
| `vw_reporte_excel_contrataciones` | Vista JOIN completo para reportes |
| `vw_historial_celdas_multilinea` | Vista con GROUP_CONCAT para LibreOffice |

## Dependencias Locales (vendor/)

Para evitar CDNs y funcionar sin internet, todo está en `vendor/`:

| Archivo | Fuente | Tamaño |
|---------|--------|--------|
| `tailwind.min.css` | Tailwind CSS v3.4.19 (JIT build, solo clases usadas) | ~16KB |
| `sql-wasm.js` | sql.js v1.8.0 | ~51KB |
| `sql-wasm.wasm` | sql.js WASM binary | ~600KB |

Regenerar `tailwind.min.css` si se agregan nuevas clases:
```bash
npm install --save-dev --no-bin-links tailwindcss@3.4.19
# crear tailwind.config.js apuntando a index.html
npx tailwindcss -i input.css -o vendor/tailwind.min.css --minify
```

## Electron WinUnpacked

Para no depender de ningún navegador, se construye `dist/win-unpacked/` con Chromium embebido.

### Source files
- `main.js` — Electron main process (ventana 1400x900, sin menú)
- `preload.js` — contextBridge para IPC seguro
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
make combine          # Concatena index.html + Tablas8.sql + main.js + package.json + doc.md → combined.txt
make clean            # rm -f combined.txt
make commit msg="x"   # git add -A + git commit
make push             # git push
make github msg="x"   # commit + push (shortcut)
make serve            # python3 -m http.server 8000 (sirve index.html por HTTP para evitar file://)
make electron-build-win    # Build win-unpacked para Windows
make electron-build-linux  # Build AppImage para Linux
```

El schema usado en `make combine` se configura con `SCHEMA=bdd/Tablas7.sql make combine` (por defecto usa `bdd/Tablas8.sql`).

## Reglas del Proceso

1. **doc.md primero**: antes de cualquier implementación o cambio de código, actualizar esta documentación con lo que se planea hacer.
2. **Makefile siempre**: después de cambios, ejecutar `make combine`.
3. **Sin hardcodeo**: cero assumptions de naming conventions. Toda heurística debe ser configurable.
4. **Historial de cambios**: cada cambio debe agregarse a la cronología en `doc.md` con fecha, archivo, y razón.
5. **DRY + Reutilización**: toda pieza de lógica debe tener una representación única. No repetir código ni copiar-pegar bloques. Si un patrón aparece en más de un lugar, extraer a función reutilizable. La modularidad no se mide en líneas por archivo ni por función, sino en ausencia de redundancia y en que cada función tenga una única responsabilidad (SRP). Una función de 200 líneas sin duplicación interna es mejor que 4 funciones de 50 líneas con lógica repetida.

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
| 23 | `.gitignore`, `Makefile`, `prompt`, `doc.md`, `bdd/Tablas8.sql` | Reorganización del proyecto: SQL movidos a `bdd/`, Makefile con `SCHEMA` variable y targets win/linux, prompt actualizado a Tablas8.sql, gitignore mejorado | Reflejar estructura actual y dar soporte multiplataforma |
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
| 36 | `index.html`, `bdd/Tablas8.sql` | P6: `nro_ejemplares` movido de `cat_documento` a `expedientes` como campo editable por registro. Eliminados `actualizarNroEjemplares()`, display en catálogo, y columna de schema en cat_documento | El nro. de ejemplares varía por expediente, no por tipo de documento |
| 37 | `index.html` | Unificación DRY: `CATALOGO_POR_SELECT` como fuente única de verdad para catálogos. Expandido con campo `cols`. `cargarCatalogos()` y `poblarSelectores()` ahora iteran sobre `CATALOGO_POR_SELECT` eliminando los mapeos paralelos duplicados | Eliminar duplicación de 3 estructuras de datos que mapeaban select→catálogo (regla DRY del doc.md) |
| 38 | `index.html` | Fix: `captureAndRestoreFormState()` para preservar valores de selects al repoblar catálogos | Evitar que campos del formulario se vacíen al añadir nuevos registros a catálogos |
| 39 | `index.html` | Fix: Eliminado `e.stopPropagation()` del botón '+' de catálogo | El botón '+' no despliega menú al tocar el ícono exacto |
| 40 | `index.html` | Feature: tipo_cambio aplica automáticamente a monto_adjudicado_bs | Calcular monto adjudicado en BS al cambiar monto USD o tipo de cambio |
| 41 | `index.html`, `bdd/Tablas8.sql` | Historial overhaul: subformulario eliminado en edición, trigger INSERT para snapshot inicial, observaciones con formato sin prefijos (solo valores), ficha muestra solo observación más nueva con expand, "ver historial completo" como tabla de snapshots | Reemplazar modelo de diferencias por snapshot completo desde creación |

---

## Pendientes / Por Hacer

### Estado de la BDD (schema v8 actual)

El schema actual (`bdd/Tablas8.sql`) tiene 10 catálogos + expedientes + historial con snapshot completo. Todos los puntos del plan de funcionalidades han sido implementados excepto los siguientes:

| # | Prioridad | Descripción | Archivos | Estado |
|---|-----------|-------------|----------|--------|
| 4 | 🟡 Media | Archivo separado para ajustes de BD (opción A: tabla `app_config` en SQLite vs opción B: `db-settings.js`) | `db-settings.js` o schema | pendiente — requiere decisión A vs B |
| — | 🟢 Baja | Archivo de config específico para BDD (`bdd_config.json`) | Nuevo archivo | pendiente |

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
