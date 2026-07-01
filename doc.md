# Gestión de Expedientes con Historial — Documentación

## Contexto Termux (Android)

Este proyecto se edita y construye desde **Termux** en Android. Si inicias una sesión nueva:

| Ítem | Valor |
|------|-------|
| Directorio | `/storage/emulated/0/baseaccess` |
| Repositorio | `git@github.com:JulianoJul/baseaccess.git` |
| Node.js | `pkg install nodejs` (si no está) |
| Descargas | `curl` viene preinstalado |

**Comandos clave para reconstruir el `.exe`:**
```bash
npm install --save-dev --no-bin-links electron@latest electron-builder@latest
node node_modules/electron-builder/cli.js --win portable --x64
```
El build se genera en `dist/win-unpacked/`. Copiar esa carpeta a USB y ejecutar `GestionExpedientes.exe`.

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
├── vendor/              # Dependencias locales (sin CDN)
│   ├── tailwind.min.css # Tailwind CSS build estático (16KB, tree-shaken)
│   ├── sql-wasm.js      # sql.js loader
│   └── sql-wasm.wasm    # Motor SQLite WASM (~600KB)
├── main.js              # Electron main process (ventana 1400x900)
├── package.json         # Electron + electron-builder config
├── Tablas6.sql           # Schema SQLite v6
├── doc.md                # Esta documentación
├── prompt                # Prompt para auditorías (opencode/Qwen)
├── Makefile              # combine / clean / commit / push / github
├── .gitignore            # node_modules/, dist/
└── intento               # (reservado)
```

## Tablas del Schema (Tablas6.sql)

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
```bash
npm install --save-dev --no-bin-links electron@latest electron-builder@latest
node node_modules/electron-builder/cli.js --win portable --x64
```

El `.exe` portable se genera en `dist/` (~80MB con Chromium embebido). Se ejecuta sin instalación, sin admin, sin depender del navegador del sistema.

> **Nota:** En Android/Termux el paso de empaquetado NSIS falla por falta de 7zip para ARM. Usar `--x64` para que el build se despliegue en `dist/win-unpacked/` — esa carpeta (~360MB) es funcional: copiar a USB, ejecutar `GestionExpedientes.exe` directo.

## Makefile

```bash
make combine          # Concatena index.html + Tablas6.sql + main.js + package.json + doc.md → combined.txt
make clean            # rm -f combined.txt
make commit msg="x"   # git add -A + git commit
make push             # git push
make github msg="x"   # commit + push (shortcut)
make serve            # python3 -m http.server 8000 (sirve index.html por HTTP para evitar file://)
```

## Reglas del Proceso

1. **doc.md primero**: antes de cualquier implementación o cambio de código, actualizar esta documentación con lo que se planea hacer.
2. **Makefile siempre**: después de cambios, ejecutar `make combine`.
3. **Sin hardcodeo**: cero assumptions de naming conventions. Toda heurística debe ser configurable.
4. **Historial de cambios**: cada cambio debe agregarse a la cronología en `doc.md` con fecha, archivo, y razón.
5. **DRY + Reutilización**: toda pieza de lógica debe tener una representación única. No repetir código ni copiar-pegar bloques. Si un patrón aparece en más de un lugar, extraer a función reutilizable. La modularidad no se mide en líneas por archivo ni por función, sino en ausencia de redundancia y en que cada función tenga una única responsabilidad (SRP). Una función de 200 líneas sin duplicación interna es mejor que 4 funciones de 50 líneas con lógica repetida.

---

## UI: Rediseño tipo LibreOffice Base (Plan)

### Objetivo
Reestructurar la interfaz para que tenga una navegación lateral (panel izquierdo) similar a LibreOffice Base, con vistas intercambiables.

### Layout nuevo
```
┌──────────────────────────────────────────────────┐
│  Header (input file, botones, búsqueda)           │
├──────────┬───────────────────────────────────────┤
│  NAV     │  VISTA PRINCIPAL (cambia según        │
│  lateral │  selección del nav)                    │
│          │                                        │
│ 📋 Exp.  │  Home: selector de formularios        │
│ ⏳ Hist. │  Expedientes: grid Excel              │
│          │    └─ fila expandible → sub-grid      │
│          │       historial                       │
│          │         └─ fila expandible → detalle  │
│          │            del movimiento             │
└──────────┴───────────────────────────────────────┘
```

### Cambios en index.html
- Layout `flex`: nav lateral (`<aside>`) + main content (`<section>`)
- Nav con 2 entradas: "📋 Expedientes", "⏳ Historial"
- Vista Home: tarjetas de formularios al cargar BD
- Grid principal mejorado estilo Excel (alternar filas, mejor scroll)
- Sub-grid de historial dentro de cada fila expandida, con filas expandibles

### Lo que NO cambia
- Esquema de colores oscuro (gray-900, teal-400)
- Modal de formulario (nuevo/editar expediente)
- Lógica CRUD (guardar, eliminar)
- Dependencias vendor/

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

### Rediseño LibreOffice Base (Julio 2026)

| # | Archivo | Cambio | Razón |
|---|---------|--------|-------|
| 1 | `index.html` | Layout flex con nav lateral + main intercambiable | Navegación tipo LibreOffice Base |
| 2 | `index.html` | Nueva vista inicio con tarjetas de formularios | Pantalla de selección al cargar BD |
| 3 | `index.html` | Grid Excel con filas alternadas (even/odd) | Mejor legibilidad de datos |
| 4 | `index.html` | Fila expandible con tabs: Movimientos (sub-grid) / Ficha (tarjetas) | Historial como grid expandible + ficha del expediente como vista alterna |
| 5 | `index.html` | Sub-grid de historial con filas expandibles a detalle | Cada movimiento del historial expandible para ver información completa |
| 6 | `index.html` | Nueva vista "Historial Global" en nav lateral | Traza completa de todos los movimientos |
| 7 | `doc.md` | Documentado el plan de rediseño antes de implementar | Regla: doc.md primero |
