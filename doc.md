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
node node_modules/electron-builder/cli.js --win portable --x64
```
El build se genera en `dist/win-unpacked/`. Copiar esa carpeta a USB y ejecutar `GestionExpedientes.exe`.

> **Nota:** En Linux de escritorio (Arch, Ubuntu, etc.) usar `make electron-build-linux` o `npm run build:linux` para generar el AppImage.

**Importante:** `node_modules/` y `dist/` no se suben a git (`.gitignore`). Hay que reinstalar dependencias cada sesión nueva.

## ⚠️ Limitación: `file://` + WASM

Al abrir `index.html` con doble click (`file://` protocol), los navegadores **bloquean la carga del binario WASM** por seguridad. Síntomas:
- El botón "+ Nuevo Expediente" queda deshabilitado
- Los registros de la BDD no se muestran en la tabla

**Usar siempre el modo Electron portable** (`dist/win-unpacked/GestionExpedientes.exe`) para evitar este problema.

## Arquitectura

App web 100% cliente-side. **HTML + Tailwind CSS = UI** | **sql.js (SQLite WASM) = Data Layer**.
Sin backend, sin servidor, sin runtime externo. Un solo archivo HTML.

Dos modos de ejecución:

1. **Navegador** — abrir `index.html` directo (dependencias locales en `vendor/`)
2. **Electron portable** — `GestionExpedientes.exe` con Chromium embebido (sin depender de Firefox/Chrome)

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
│  Modo Electron (portable, sin instalación)        │
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
| `expedientes` | **Tabla principal**: ~30 columnas con fechas, montos, FK |
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

## Electron Portable

Para no depender de ningún navegador, se puede construir un `.exe` portable:

### Source files
- `main.js` — Electron main process (ventana 1400x900, sin menú)
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

El `.exe` portable se genera en `dist/` (~80MB con Chromium embebido). El AppImage en `dist/GestionExpedientes-*.AppImage`. Se ejecutan sin instalación, sin admin, sin depender del navegador del sistema.

> **Nota (Termux/Android):** el paso de empaquetado NSIS falla por falta de 7zip para ARM. Usar `--win portable --x64` para que el build se despliegue en `dist/win-unpacked/` — esa carpeta (~360MB) es funcional: copiar a USB, ejecutar `GestionExpedientes.exe` directo.

## Makefile

```bash
make combine          # Concatena index.html + Tablas8.sql + main.js + package.json + doc.md → combined.txt
make clean            # rm -f combined.txt
make commit msg="x"   # git add -A + git commit
make push             # git push
make github msg="x"   # commit + push (shortcut)
make serve            # python3 -m http.server 8000 (sirve index.html por HTTP para evitar file://)
make electron-build-win    # Build .exe portable para Windows
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

---

## Pendientes / Por Hacer

### Estado de la BDD (schema v8 actual)

El schema actual (`bdd/Tablas8.sql`) tiene 10 catálogos + expedientes + historial con snapshot completo. Cambios respecto a v7:

| Tabla | Cambio respecto a v7 |
|-------|---------------------|
| ~~`cat_estado_accion`~~ | **Eliminada** — sus valores se fusionaron en `cat_estatus_detalle` |
| `cat_estatus_detalle` | Agregados valores: "SE ENTREGA CON LA FIRMA", "SE ENTREGA CON LA MODIFICACIÓN", "SE RECIBE PARA LA FIRMA", "SE DEVUELVE CON LA FIRMA", "SE RECIBE CON LA FIRMA", "SE ENTREGA PARA LA FIRMA" |
| `cat_documento` | Agregada columna `nro_ejemplares INTEGER DEFAULT 1` |
| `expedientes.solped` | Eliminada UNIQUE constraint → `TEXT` libre |
| `expedientes.id_estado_accion` | Eliminada (fusionado con `id_estatus`) |
| `expedientes.nro_contrato_sap` | Cambiado de `INTEGER` a `TEXT` |
| `historial_movimientos` | Agregadas columnas: `nro_proceso`, `presupuesto_base_usd`, `tipo_cambio`, `monto_adjudicado_usd`, `id_resultado`, `id_empresa`, `tiempo_ejecucion`, `fecha_firma_contrato` |
| `historial_movimientos.id_estado_accion` | Eliminada (fusionado con `id_estatus`) |

### 🔴 Prioridad Alta

| # | Descripción | Archivos | Detalle |
|---|-------------|----------|---------|
| 1 | **Eliminar `cat_estado_accion` y fusionar con `cat_estatus_detalle`** | `Tablas7.sql`, `index.html` | Unificar ambos catálogos. Los valores actuales de estado_accion pasan a estatus_detalle con nombres como "Se entrega para la firma", "Se devuelve con la firma", "Se recibe para la firma", etc. Ajustar trigger y vistas. |
| 2 | **Historial normalizado que guarde todo** | `Tablas7.sql`, `index.html` | Modificar `historial_movimientos` para que almacene snapshot completo de cada cambio (todas las columnas relevantes del expediente) de forma normalizada. La UI debe seguir mostrando los mismos campos. |
| 3 | **Bug: agregar expediente no guarda** | `bdd/Tablas7.sql`, `index.html` | **RESUELTO:** SOLPED tenía UNIQUE constraint. Se elimina la constraint UNIQUE, el campo pasa a texto libre (uno o varios SOLPED separados por " / "). También se actualiza la validación en JS. |

### 🟡 Prioridad Media

| # | Descripción | Archivos | Detalle |
|---|-------------|----------|---------|
| 4 | **Autogenerar observación** | `index.html` | Al guardar un movimiento, generar texto automático: "Recibido: [fecha] / Devuelto: [fecha] — [estado_accion] — [documento]". Permitir texto extra adicional. |
| 5 | **Validación: fecha recibido ≤ fecha devuelto** | `index.html` | No permitir guardar si `fecha_recibido > fecha_devuelto`. Validar en frontend antes de enviar. |
| 6 | **Validación: solo 2 decimales** | `index.html` | Restringir input a máximo 2 decimales en campos numéricos (presupuesto, montos, tipo de cambio). `oninput` o `step="0.01"`. |
| 7 | **Bug: tipo de cambio no muestra decimales** | `index.html` | `formatNum()` muestra 2 decimales siempre, pero si el usuario escribe "1,5" debería mostrarse como "1,50". Verificar que `calcularBs()` y el formato funcionen correctamente con decimales. |
| 8 | **Botón "+" en observaciones** | `index.html` | Agregar botón para añadir múltiples entradas de observaciones (no solo un textarea). |
| 9 | **Tiempo ejecución con "DÍAS" automático** | `index.html` | El campo `tiempo_ejecucion` debe autocompletar o forzar el formato en días (ej: "30 DÍAS"). |
| 10 | **"Se han detectado cambios, ¿guardar?"** | `index.html` | Detectar cambios no guardados al cerrar modal o cambiar de expediente, preguntar si desea guardar. |
| 11 | **Número de ejemplares en DOCUMENTO** | `Tablas7.sql`, `index.html` | Agregar campo `nro_ejemplares` o similar en `cat_documento` o en el formulario al seleccionar un documento. |

### 🟢 Prioridad Baja

| # | Descripción | Archivos | Detalle |
|---|-------------|----------|---------|
| 12 | **Archivo de config específico para BDD** | Nuevo archivo | Crear archivo de configuración (ej: `bdd_config.json`) con ajustes propios de la base de datos (mappings, reglas de validación, columnas sensibles) que se cargue dinámicamente. |
| 13 | **Botón "más" en cada campo para validaciones** | `index.html` | Agregar botón "+" junto a cada campo del formulario para añadir validaciones personalizadas desde la UI. Posteriormente un menú para editarlas. |
| 14 | **Marcar celdas que suelen cambiar** | `index.html` | Resaltar visualmente las columnas que se registran en historial (id_tipo_contrato, id_emisor, id_receptor, id_gerencia, id_superintendencia, id_documento, id_estatus, id_estado_accion, fecha_recibido, fecha_devuelto, observaciones_generales) sin modificar la tabla historial. |

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
