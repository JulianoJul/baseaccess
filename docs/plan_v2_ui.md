# Plan v2 UI — Rediseño Frontend y Experiencia de Usuario

> Documento exclusivo para cambios visuales, HTML, CSS, Alpine.js y HTMX.
> Cada tarea está diseñada para ser implementada de forma autónoma con DeepSeek Flash.
> **Regla:** Preservar OBLIGATORIAMENTE todos los atributos HTMX (`hx-get`, `hx-post`, `hx-target`, etc.) y Alpine.js (`x-data`, `x-model`, `@click`, etc.) existentes. Solo migrar el markup visual.

---

## 1. Visión del Rediseño (SaaS Dashboard Bimodal)

### 1.1 Principios Estéticos: Minimalismo Funcional

El rediseño abandona la estética de "aplicación de escritorio tradicional" para adoptar un lenguaje visual de **SaaS enterprise moderno**. Cada elemento debe justificar su existencia; no hay decoración sin función.

- **Jerarquía por espaciado y peso, no por bordes.** Se eliminan los bordes dobles, los separadores gruesos y las líneas de fieldset visuales. La separación entre regiones se logra mediante diferencias de fondo (superficies en capas), sombras de elevación sutiles y padding estratégico.
- **Radios de borde consistentes.** Sistema escalonado: 4px para badges y controles pequeños, 6px para inputs y botones, 8px para tarjetas y dropdowns, 12px para paneles y drawers, y 9999px para pills y estados. Nunca mezclar radios arbitrarios en el mismo viewport.
- **Profundidad realista.** En tema claro, la elevación se expresa mediante sombras suaves basadas en luz (más intensas cuanto más alto el elemento). En tema oscuro, la elevación se expresa mediante capas de opacidad y bordes sutilmente más claros, evitando sombras negras que ensucian la interfaz.
- **Opacidad como herramienta de estado.** Los estados deshabilitados, los placeholders y los textos de ayuda no usan grises arbitrarios; usan el color de texto primario con opacidad reducida (40% o 60%). Esto mantiene la armonía cromática en ambos temas.
- **Densidad de información alta pero respirable.** Optimizado para 720px de altura. Inputs compactos (36px de altura), márgenes verticales entre secciones reducidos a 12px, y tarjetas con padding de 12px en lugar de 24px.

### 1.2 Estrategia DRY General

1. **Sistema de Componentes Universal.** Cada patrón visual repetido debe existir en un único template parcial parametrizado. No puede haber más de una definición de "botón primario", "input de texto" o "tarjeta de sección" en toda la aplicación. Crear librería interna de templates: botón, input, badge, card, modal-shell, tabla de datos, sección de formulario, estado vacío y fila de acciones.
2. **Unificación de Módulos de Datos.** Los 9 módulos (expedientes, requisiciones, etc.) actualmente generan markup de tabla divergente. Deben converger en un **único template de tabla parametrizado** que reciba la configuración de columnas directamente desde el struct `ModuloConfig` del backend. Las diferencias entre módulos se resuelven mediante arrays de metadatos (nombre, clave, ancho, prioridad, tipo de dato).
3. **Layout por Slots, no por Copia.** El shell de la aplicación, los modales y los fieldsets del formulario deben convertirse en templates de layout que definan la estructura visual y reciban el contenido específico mediante bloques de contenido (slots). Esto garantiza que un cambio estructural se propague automáticamente a todos los usos.

### 1.3 Estrategia de Responsividad Extrema: 1280px ↔ 640px

La aplicación tiene dos modos de uso reales: **ventana completa** (1280×720) y **panel estrecho** (640×720, compartiendo pantalla con Excel).

- **Breakpoint crítico único: 960px.** Por encima, layouts multi-columna y modales centrados. Por debajo, modo Panel Estrecho.
- **Header.** En Panel Estrecho, altura 44px. Los botones de acción global muestran **solo su icono Font Awesome** (sin texto), usando `title` para tooltip. El campo de búsqueda se colapsa a un icono `fas fa-search` que expande un input flotante al recibir focus.
- **Tabla de Datos.** En Panel Estrecho, se mantiene como tabla nativa con **scroll horizontal obligatorio**. La primera columna (acciones) y la segunda columna (documento/descripción) se vuelven **sticky-left** con fondo opaco. Las columnas secundarias permanecen en el área scrolleable. El resizer de columnas se desactiva completamente por debajo de 960px.
- **Subfilas Expandibles.** En ancho completo, se mantiene el patrón actual de subfila inline con colspan. En Panel Estrecho, el click en una fila abre un **drawer lateral de "Vista Rápida"** en lugar de expandir la subfila dentro de la tabla.
- **Formulario de Registro.** En ancho completo, modal centrado o panel inline. En Panel Estrecho, **drawer derecho que ocupe el 100% del ancho disponible**.
- **Bottom Bar de Módulos.** Se mantiene fija inferior. En Panel Estrecho los botones muestran **solo iconos Font Awesome** sin texto. Si el ancho es insuficiente para los 10 botones, los módulos sobrantes se agrupan bajo un botón "Más" (`fas fa-ellipsis-h`) que despliega un menú flotante compacto.
- **Gantt.** En Panel Estrecho, la columna de nombres de proceso se reduce a 140px y las celdas de día a 28px. La cabecera de semanas y la columna de procesos se vuelven **sticky** (top y left respectivamente). Los controles de hojas y juntas se apilan verticalmente.

---

## 2. Sistema de Diseño y Tokens (Variables CSS)

El sistema debe implementarse en un archivo de tema dedicado, utilizando variables CSS nativas. El tema por defecto es claro; el oscuro se activa mediante el atributo `data-theme="dark"` en el elemento HTML, con fallback a `prefers-color-scheme: dark`.

### 2.1 Paleta de Colores Exacta

**Tema Claro**

| Token | Valor | Uso |
|-------|-------|-----|
| bg-base | slate-50 (#f8fafc) | Fondo de la aplicación |
| bg-surface | white (#ffffff) | Tarjetas, paneles, celdas de tabla |
| bg-surface-elevated | white | Dropdowns, menús flotantes (con sombra md) |
| bg-input | slate-100 (#f1f5f9) | Campos de formulario |
| bg-hover | slate-100 (#f1f5f9) | Estado hover de filas y botones |
| bg-active | slate-200 (#e2e8f0) | Estado active/selected |
| border-subtle | slate-200 (#e2e8f0) | Separadores de bajo contraste |
| border-default | slate-300 (#cbd5e1) | Bordes de inputs y tarjetas |
| border-focus | teal-500 (#14b8a6) | Anillo de foco y estados activos |
| text-primary | slate-900 (#0f172a) | Texto principal y títulos |
| text-secondary | slate-600 (#475569) | Labels, subtítulos, datos secundarios |
| text-tertiary | slate-400 (#94a3b8) | Placeholders, hints, timestamps |
| text-inverse | white | Texto sobre fondos de acento |
| primary | teal-600 (#0d9488) | Botones primarios, enlaces activos |
| primary-hover | teal-500 (#14b8a6) | Hover de elementos primarios |
| primary-text | teal-500 (#14b8a6) | Texto de acento, íconos activos |
| success | emerald-500 (#10b981) | Estados positivos, badges de éxito |
| warning | amber-500 (#f59e0b) | Advertencias, presupuestos |
| danger | red-500 (#ef4444) | Eliminación, errores críticos |
| info | blue-500 (#3b82f6) | Información neutral |

**Tema Oscuro**

| Token | Valor | Uso |
|-------|-------|-----|
| bg-base | slate-950 (#020617) | Fondo de aplicación |
| bg-surface | slate-900 (#0f172a) | Tarjetas y paneles |
| bg-surface-elevated | slate-800 (#1e293b) | Dropdowns, modales, drawers |
| bg-input | slate-900 (#0f172a) | Campos de formulario con borde slate-700 |
| bg-hover | slate-800 (#1e293b) | Hover de filas |
| bg-active | slate-700 (#334155) | Estado selected |
| border-subtle | slate-800 (#1e293b) | Separadores sutiles |
| border-default | slate-700 (#334155) | Bordes de inputs y tarjetas |
| border-focus | teal-400 (#2dd4bf) | Foco en modo oscuro |
| text-primary | slate-50 (#f8fafc) | Texto principal |
| text-secondary | slate-400 (#94a3b8) | Labels y datos secundarios |
| text-tertiary | slate-500 (#64748b) | Placeholders y hints |
| text-inverse | slate-900 | Texto sobre fondos claros de acento |
| primary | teal-500 (#14b8a6) | Botones y acentos principales |
| primary-hover | teal-400 (#2dd4bf) | Hover primario |
| primary-text | teal-400 (#2dd4bf) | Texto de acento |
| success | emerald-400 (#34d399) | Éxito |
| warning | amber-400 (#fbbf24) | Advertencia |
| danger | red-400 (#f87171) | Peligro |
| info | blue-400 (#60a5fa) | Información |

### 2.2 Superficies y Elevación

- **Capa 0 (Base):** Fondo de la aplicación. Sin sombra.
- **Capa 1 (Surface):** Tarjetas, celdas de tabla, bottom bar. Sombra sm en claro, sin sombra en oscuro.
- **Capa 2 (Elevated):** Dropdowns, menús de módulos, tarjetas flotantes. Sombra md.
- **Capa 3 (Overlay):** Modales, drawers, toasts. Sombra lg o xl. En oscuro, overlay semitransparente muy oscuro (rgba negro al 60%) con opcional backdrop-filter blur de 4px.

### 2.3 Espaciado y Densidad (Optimizado 720p)

Sistema de 4px base. Densidad compacta para maximizar datos visibles en altura limitada.

- **space-1:** 4px — Separación mínima entre icono y texto.
- **space-2:** 8px — Gap entre elementos internos de una card.
- **space-3:** 12px — Padding de tarjetas y modales, padding horizontal de inputs, gap entre fieldsets.
- **space-4:** 16px — Padding de drawers, separación entre secciones.
- **space-6:** 24px — Margen entre regiones principales (usado con moderación).

**Reglas de densidad críticas:**
- Altura mínima de inputs: 36px (2.25rem). Nunca 44px.
- Altura de botones con texto: 36px. Botones de solo icono: 32px.
- Padding de celdas de tabla: 8px vertical, 12px horizontal.
- Padding de tarjetas de formulario: 12px (space-3).

### 2.4 Tipografía

- **Familia:** Inter, system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif.
- **Tamaños:**
  - 10px (0.625rem): Encabezados de Gantt, badges pequeños.
  - 12px (0.75rem): Labels de formulario, captions, timestamps.
  - 13px (0.8125rem): Cuerpo de tabla (más pequeño que base para densidad).
  - 14px (0.875rem): Inputs, botones, cuerpo general.
  - 16px (1rem): Títulos de sección (legend/card header).
  - 18px (1.125rem): Título de modal/drawer.
  - 20px (1.25rem): Brand/header principal.
- **Pesos:** 400 (cuerpo), 500 (labels, botones), 600 (headers, títulos de sección), 700 (títulos de página, números destacados).
- **Line-height:** 1.25 para tablas y datos densos; 1.5 para párrafos de observaciones y textos largos.

### 2.5 Radios y Sombras

| Token | Valor | Uso |
|-------|-------|-----|
| radius-sm | 4px | Badges, tags, checkboxes, celdas de Gantt |
| radius-md | 6px | Inputs, botones pequeños, items de lista |
| radius-lg | 8px | Tarjetas, modales, dropdowns, bottom-bar |
| radius-xl | 12px | Drawers, paneles grandes de formulario |
| radius-full | 9999px | Pills, toggles, badges de estado, botones circulares |

| Token | Valor (Tema Claro) | Uso |
|-------|-------------------|-----|
| shadow-sm | 0 1px 2px rgba(0,0,0,0.05) | Inputs enfocados, badges |
| shadow-md | 0 4px 6px rgba(0,0,0,0.07) | Tarjetas, dropdowns, bottom-bar |
| shadow-lg | 0 10px 15px rgba(0,0,0,0.1) | Modales, drawers |
| shadow-xl | 0 20px 25px rgba(0,0,0,0.12) | Toasts, notificaciones críticas |

En tema oscuro, las sombras deben incrementar su opacidad en un 30% relativo o sustituirse por un sutil brillo de borde en elementos elevados.

---

## 3. Estructura de Layout y Componentes DRY

### 3.1 Grid Base y appShell

El body debe reestructurarse como un **grid CSS de tres filas** que ocupe el 100% del viewport y elimine el scroll global de la página. El scroll debe ocurrir únicamente dentro de las regiones designadas.

- **Fila 1 (Header):** Altura fija de 48px (44px en Panel Estrecho). Flexbox con justify-content space-between. Sin borde inferior grueso; usa sombra sm o borde sutil de 1px con border-subtle.
- **Fila 2 (Main):** flex-grow 1, overflow-y auto, padding space-3. Región de contenido donde se renderiza la tabla o el Gantt.
- **Fila 3 (Bottom Bar):** Altura fija de 52px (48px en Panel Estrecho). Fondo bg-surface, borde superior sutil, sombra md.

El drag-and-drop de archivos .db se mantiene en el nivel del body, pero su indicador visual debe modernizarse a un borde dashed de 2px con color primary y fondo semitransparente.

### 3.2 Navegación y Bottom Bar

- **Botones de Módulo:** Crear template parcial único `ui_module_button` que reciba: nombre del módulo, clase de icono Font Awesome (`fas fa-*`), endpoint hx-get, estado activo (booleano), y badge opcional. En Panel Estrecho, el template omite el texto y renderiza solo el icono centrado en un botón cuadrado compacto.
- **Botón "Más":** Cuando el ancho es insuficiente para todos los módulos, los últimos elementos se agrupan bajo un botón con icono `fas fa-ellipsis-h`. Al hacer click, despliega un menú flotante (Capa 2) con los módulos restantes.
- **Botón de Tema:** Agregar al header un botón circular (32px) con iconos `fas fa-sun` (claro) y `fas fa-moon` (oscuro) que alterne. La preferencia se persiste en localStorage y se aplica al atributo `data-theme` del HTML antes del primer render para evitar flash de tema incorrecto.

### 3.3 Auditoría de Redundancias y Unificación

El Agente Programador debe eliminar las siguientes redundancias:

1. **Templates de Input.** Converger `form_input_text_alpine`, `form_input_date_alpine`, `form_input_number_alpine`, `form_input_monto_alpine` y `form_input_spin_alpine` en un único `ui_input_field`. Este template recibe parámetros: tipo de input, nombre, etiqueta, modelo Alpine, directivas adicionales (ej. x-currency), clases de ancho, y un slot para elementos adjuntos (como los botones +/- del spin). El wrapper visual (label, borde, foco, espaciado) es idéntico para todos.
2. **Botones de Acción.** Converger todos los botones de guardar, eliminar, cancelar, exportar, nuevo registro, editar fila y fijar en `ui_action_button`. Parametrizado por: variante visual (primary, secondary, danger, ghost, icon-only), tamaño (sm, md), icono Font Awesome (`fas fa-*`), etiqueta, y un mapa de atributos HTMX que se inyecten directamente en el elemento raíz.
3. **Tablas de Módulos.** Reemplazar las 9 variantes de tabla por un único `ui_data_table`. Las columnas se generan iterando el array `Columnas` (o un nuevo metadato de columnas) del `ModuloConfig`. Cada celda de estatus usa `ui_badge`. Las filas vacías usan `ui_empty_state`.
4. **Modales.** Reemplazar los 8 modales actuales (form, historial, ruta, pendientes, recientes, frecuentes, export, sumas) y los 5 modales del Gantt por un único template de layout `ui_modal_shell`. Este shell define el backdrop, el contenedor (modal centrado o drawer), el header con título y botón de cierre (`fas fa-times`), el cuerpo con scroll interno, y el footer opcional. Cada instancia concreta solo inyecta contenido en los slots.
5. **Fieldsets del Formulario.** Reemplazar por `ui_form_section`, una tarjeta contenedora con título, icono Font Awesome opcional, y capacidad de colapsar en Panel Estrecho. En ancho completo, las secciones pueden permanecer expandidas. En Panel Estrecho, todas excepto la primera deben iniciar colapsadas.
6. **Subfilas y Detalles.** Unificar `tabla_subrow_trazabilidad` y `tabla_subrow_observaciones_notas` en `ui_table_detail_panel`. En ancho completo, se inyecta como subfila inline. En Panel Estrecho, su contenido se renderiza dentro del drawer de Vista Rápida.

---

## 4. Guía de Refactorización Crítica

### 4.1 Tablas de Datos

- **Contenedor y Scroll.** Envolver la tabla en un contenedor con `overflow-x-auto` y un `min-width` explícito basado en la suma de los anchos de columna definidos en la configuración. Fuerza scroll horizontal en Panel Estrecho en lugar de comprimir columnas hasta la ilegibilidad.
- **Cabecera Sticky.** El thead debe usar `position: sticky` con `top: 0`, `z-index: 10`, y fondo `bg-surface` sólido (sin transparencia).
- **Columna de Acciones Sticky.** La primera columna (botones de expandir, editar y fijar) debe ser `position: sticky` con `left: 0`, `z-index: 20`, y fondo `bg-surface` sólido.
- **Resizer Condicional.** Encapsular la lógica de resize de columnas en una función que verifique `window.innerWidth >= 960` antes de adjuntar event listeners. Por debajo de ese ancho, las columnas usan ancho intrínseco o el ancho definido en el metadato del módulo.
- **Columnas Responsivas.** Definir una clase de utilidad que oculte columnas secundarias en Panel Estrecho. Las columnas siempre visibles son: Acciones, Documento/Descripción (truncada con ellipsis), Estatus. Los datos restantes (SOLPED, Gerencia, fechas, montos) se trasladan al drawer de Vista Rápida.
- **Paginación.** Extraer la paginación actual a un template parcial `ui_pagination`. Los botones de página mantienen sus atributos `hx-get` hacia `/api/filtrar-expedientes`, `hx-vals` con el número de página, y `hx-target="#vista-tabla"`.

### 4.2 Formularios

- **Grid del Formulario.** El contenedor del formulario debe usar CSS Grid con una sola columna en Panel Estrecho y un máximo de tres columnas en ancho completo (usando `repeat` con `minmax` de 240px).
- **Secciones Colapsables.** Cada fieldset actual se convierte en una instancia de `ui_form_section`. En Panel Estrecho, todas las secciones excepto "Información General" deben iniciar en estado colapsado.
- **Agrupación de Montos.** Los pares de inputs de monto (USD/Bs) deben agruparse visualmente en un contenedor flex de dos columnas con un separador central que muestre el tipo de cambio.
- **Observación Auto-generada.** Distinguir claramente el textarea de observación auto-generada (solo lectura) del textarea de observaciones manuales. El auto-generado debe tener un fondo distintivo (bg-active con opacidad reducida) y un borde dashed, indicando que es computado, no editable.
- **Drawers en Panel Estrecho.** El formulario de registro debe renderizarse dentro de `ui_modal_shell` configurado como drawer derecho en <960px. Debe ocupar el 100% del ancho para aprovechar los 640px al máximo, con padding generoso interno (space-4) y scroll vertical independiente.

### 4.3 Diagrama de Gantt (Ruta de Procesos) — UI Only

#### 4.3.1 Visión General

La Ruta de Procesos es un módulo independiente de la BD principal. La UI se organiza en bloques verticales (scroll) donde cada junta es un bloque con: tabla de datos → Gantt → leyendas.

```
┌──────────────────────────────────────────────────────┐
│  [Select Hoja ▼]  [+ Nueva Hoja] [🗑]               │  ← Barra superior
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌─ JUNTA #1 ────────────────────────────────────┐   │
│  │  TABLA (1 fila, editable)                      │   │
│  │  JUNTA DIRECTIVA │ Nº │ CONSEC │ FECHA │ Guardar│  │
│  │                                                │   │
│  │  GANTT                                         │   │
│  │  ┌────┬─────────┬──────────────┬──────────┬──┐ │   │
│  │  │    │         │desde 01/06   │desde 08/06│  │ │   │  ← fila 1: desde
│  │  │    │         │al 05/06      │al 12/06   │  │ │   │  ← fila 2: al
│  │  │    │         │  SEMANA 1    │ SEMANA 2 │[+][-]│  │  ← fila 3: semanas
│  │  │    │         ├──────┬───────┼──────┬────┤  │ │   │
│  │  │ N° │ Proceso │L M X J V│...│      │    │  │ │   │  ← fila 4: días
│  │  ├────┼─────────┼──────┼───────┼──────┼────┤  │ │   │
│  │  │ 1  │ Algo    │🔵🔴 │  🟢🟡  │      │    │  │ │   │  ← fila 5+: procesos
│  │  │ 2  │ Otro    │🟢   │  🔵🔴🟡 │      │    │  │ │   │
│  │  ├────┼─────────┴──────┴───────┴──────┴────┤  │ │   │
│  │  │    │ [+] Añadir proceso                 │  │ │   │  ← fila extra
│  │  └────┴────────────────────────────────────┴──┘ │   │
│  │                                                │   │
│  │  LEYENDAS                                      │   │
│  │  🟢 Aprobado ▲▼  🔒  🟡 Espera ▲▼  🔒        │   │
│  │  [+ Añadir leyenda] (ámbito: junta/hoja/global)│   │
│  └────────────────────────────────────────────────┘   │
│                                                      │
│  ┌─ JUNTA #2 ────────────────────────────────────┐   │
│  │  (misma estructura: tabla + Gantt + leyendas)  │   │
│  └────────────────────────────────────────────────┘   │
│                                                      │
│              [+ NUEVA JUNTA]                          │  ← Botón al fondo
└──────────────────────────────────────────────────────┘
```

#### 4.3.2 Gantt — Estructura de Filas

- **fila 1:** fechas "desde" por semana
- **fila 2:** fechas "al" por semana
- **fila 3:** "SEMANA N" + botones `[+]` agregar / `[🗑]` eliminar
- **fila 4:** encabezados de días (L M X J V) por semana
- **fila 5+:** datos de procesos (N°, nombre, celdas de día)
- **fila N:** botón "Añadir proceso" al fondo

**Botones:**
- `[+]` semanas: abre modal con fechas lunes→viernes precalculadas.
- `[🗑]` semanas: abre modal con checkboxes de semanas.
- `[+]` procesos: fila al fondo con input de texto para nombre.
- Celdas vacías: click abre modal para asignar leyenda + nota.
- Celdas con datos: muestran badges de color apilados.

#### 4.3.3 Tabla de Datos de la Junta (1 fila editable)

| Campo | Tipo | Ejemplo |
|-------|------|---------|
| JUNTA DIRECTIVA | label fijo | "JUNTA DIRECTIVA" |
| Nº REUNIÓN | input number | 1 |
| CONSECUTIVA | input number | 100 |
| FECHA | date picker | 01/06/2026 |

Botón "Guardar" al lado. Botón de eliminar junta si no está bloqueada.

#### 4.3.4 Leyendas por Junta

Debajo del Gantt, las leyendas específicas de esa junta. Cada una muestra:
- Círculo de color
- Nombre
- Botón ▲▼ para reordenar
- Botón 🔒 si está bloqueada
- Botón ✏️ para editar
- Botón ✖ para eliminar (si no bloqueada)
- Botón "Añadir leyenda" con selector de ámbito: Esta junta / Esta hoja / Global

#### 4.3.5 CSS Relevante (en styles.css)

| Clase | Propósito |
|-------|-----------|
| `.gantt-table` | table-layout fixed, border-collapse, width max-content |
| `.gantt-col-num` | Columna N° sticky 44px, bg-surface sólido |
| `.gantt-col-day` | Columna de día 32px, altura 40px, padding 0 |
| `.gantt-week-header` | 9px, texto teal, bg semi-transparente |
| `.gantt-week-subheader` | 9px, texto más claro |
| `.gantt-day-header` | 10px bold |
| `.gantt-cell-empty` | bg transparente, cursor pointer |
| `.gantt-cell-active` | Cursor pointer, posición relative |
| `.gantt-cell-entries` | Flex column, gap 3px |
| `.gantt-cell-entry` | min-height 12px, radius 3px |
| `.gantt-legend-grid` | Grid 2 columnas (5 en desktop) |
| `.gantt-legend-item` | Fila flex, card bg, efecto hover |
| `.gantt-legend-circle` | Círculo de color 24px |
| `.leyenda-color-wrap input[type="color"]` | Selector de color circular |

#### 4.3.6 Densidad en Panel Estrecho

- Columna de proceso: reducida a 140px
- Celdas de día: 28px
- Cabecera de semanas: sticky-top
- Columna de procesos: sticky-left
- Controles de hoja y juntas: apilados verticalmente
- Fuente de headers de semana: 10px

#### 4.3.7 Atributos Obligatorios a Preservar

- El bloque IIFE de JavaScript vanilla y su parsing inicial de la variable `data` del servidor.
- Todas las funciones internas del IIFE: `toggleModal`, `cambiarHoja`, `eliminarHojaActual`, `crearHoja`, `guardarJunta`, `eliminarJunta`, `crearJunta`, `agregarSemana`, `guardarSemana`, `abrirEliminarSemanas`, `eliminarSemanasConfirmar`, `agregarProceso`, `eliminarProceso`, `abrirCrearLeyenda`, `editarLeyenda`, `guardarLeyenda`, `eliminarLeyenda`, `moverLeyenda`, `toggleBloquearLeyenda`, `abrirEditarCronograma`, `guardarCronoDia`, `eliminarCronoEntry`, `cerrarCronoModal`, `esc`, `jsonPost`, `reload`.
- Las llamadas fetch POST a todos los endpoints `/api/ruta-procesos-*`.
- Los atributos HTMX: `hx-get="/api/ruta-procesos"` y `hx-target="#ruta-contenido"`.
- Nota: si se migra a CSS Grid, se necesitará una función equivalente a `colorearCeldas()` para aplicar colores basándose en `procesos[].timeline`.

### 4.4 Modales, Drawers y Toasts

- **Comportamiento Híbrido.** El `ui_modal_shell` debe implementar lógica de presentación dual:
  - **Ancho completo (≥960px):** Modal centrado con overlay. Tamaños: 480px para confirmaciones, 720px para formularios complejos, 960px para el Gantt.
  - **Panel Estrecho (<960px):** Drawer fijo derecho que ocupa el 100% del ancho del viewport. Animación de entrada slide-in desde la derecha. El overlay debe ser más oscuro (70-80% opacidad).
- **Botón de Cierre.** Siempre visible en la esquina superior derecha del header, usando el icono `fas fa-times`. En drawers, área de hit extendida (44×44px).
- **Modales Secundarios.** Historial, pendientes, recientes, frecuentes, exportar y sumas deben migrar a `ui_modal_shell`. En Panel Estrecho, los modales pequeños (sumas o exportar) pueden optar por un bottom sheet (slide desde abajo).
- **Toasts Responsive.** En ancho completo, posición top-right con ancho máximo de 360px y margen de space-4. En Panel Estrecho, posición bottom-center con ancho 100% menos space-6 de margen horizontal. Cada toast debe incluir un icono Font Awesome según su tipo: `fas fa-check-circle` (éxito), `fas fa-exclamation-circle` (error), `fas fa-info-circle` (info), `fas fa-exclamation-triangle` (advertencia).

**Atributos obligatorios a preservar:**
- `Alpine.store('modals')` con su array `stack`, y los métodos `abrir`, `cerrar`, `toggle`, `tiene`, `cerrarClickFuera`. Cierre con Escape debe seguir funcionando.
- `Alpine.store('toast')` y sus métodos `mostrar`, `error`, `success`, `info`.
- Todos los `hx-get` que cargan contenido en modales deben mantener sus `hx-target` originales (`#historial-cuerpo`, `#pendientes-contenido`, `#form-expediente`, etc.).
- Los eventos Alpine que invocan `$store.modals.cerrar('id-modal')` deben permanecer intactos.

### 4.5 Atributos HTMX y Alpine.js Obligatoriamente Preservados (Recopilación Final)

**HTMX:** Todos los atributos `hx-get`, `hx-post`, `hx-target`, `hx-include`, `hx-vals`, `hx-trigger`, `hx-swap`, `hx-confirm` y `hx-indicator` existentes deben migrar sin modificación de endpoint ni de valor de target. Proteger específicamente: triggers de búsqueda (`input changed delay:200ms`), targets `#vista-tabla`, `#form-expediente`, `#tabla-cuerpo`, `#historial-cuerpo`, `#pendientes-contenido`, `#ruta-contenido`, `#excel-order-container`, y todos los endpoints bajo `/api/`.

**Alpine.js:** Todos los `x-data` raíz (`appShell`, `bdRecientes`, `fijados`, `calculadoraSumas`, `exportarExcel`, `filtroSuperintendencias`, `formularioModulo`) deben permanecer en sus elementos contenedores con los mismos parámetros de inicialización. Los stores (`modals`, `toast`, `fijados`) deben seguir registrándose. La directiva personalizada `x-currency` debe seguir funcionando. El puente `alpine-htmx-bridge.js` debe seguir escuchando `htmx:afterSettle` para invocar `Alpine.initTree(evt.detail.target)` y `htmx:afterSwap` para re-inicializar el store de fijados. Todos los eventos `@click`, `@change`, `@input` y los métodos que invocan (`toggle`, `onMontoInput`, `validarFechas`, `prepararObservaciones`, `spinFrente`, etc.) no deben perderse ni cambiar de nombre.

**Go Templates:** Las funciones de template como `rowGetStr`, `truncate`, `estatusClass`, `default`, `jsonEncode` deben seguir siendo invocables.

---

## 5. Features Adicionales de UI

### 5.1 Observaciones: Tabla Compacta + Menú Flotante de Historial

**Problema actual:** En la subfila expandible, el campo `observaciones` muestra todo el texto concatenado (auto-generadas + manuales, separadas por `\n---\n`). Al haber muchas entradas históricas, ocupa mucho espacio vertical.

**Solución propuesta:**

- **Fila principal:** Mostrar la última línea de observación truncada a una línea con ellipsis.
- **Subfila expandible (click en la fila):** Mostrar la observación más reciente completa en una sección destacada. Debajo, un botón "Ver observaciones anteriores" (`fas fa-history`).
- **Menú flotante de historial:** Al hacer click en "Ver observaciones anteriores", se despliega un **popover** (posición absolute, ancho máximo ~500px, fondo `bg-surface-elevated`, borde sutil, sombra `shadow-lg`, scroll interno).
- El menú contiene una **tabla compacta** con columnas: **#**, **Fecha**, **Documento**, **Estatus**, **Observación** (truncado), **Emisor**, **Receptor**.
- Cada fila representa **una edición completa del registro**. Se ordena de más reciente a más antigua.
- Cerrar el menú al hacer click fuera o al presionar Escape.

**Pseudocódigo: Parseo de observación más reciente (frontend)**

```
parseObservacionReciente(textoCompleto):
  IF textoCompleto contiene "\n---\n":
    partes = split(textoCompleto, "\n---\n")
    ultimaParte = partes[len(partes)-1].trim()
    RETURN {reciente: ultimaParte, anteriores: partes[0:len(partes)-1]}
  ELSE:
    RETURN {reciente: textoCompleto, anteriores: []}
```

**Implementación del menú flotante:**

- Opción A (recomendada): Usar HTMX para inyectar el HTML del historial directamente en el menú flotante (`hx-get="/api/historial"`, `hx-target="#menu-historial-{{id}}"`). Reutilizar el endpoint existente.
- El HTML devuelto por `/api/historial` ya incluye una tabla. Se reutiliza dentro del menú flotante, ajustando estilos para que sea compacto.

**Atributos a preservar:**
- El campo `observaciones` en la BD sigue siendo un solo campo concatenado con separador `\n---\n`.
- El método `prepararObservaciones()` en `formularioModulo` sigue concatenando la observación auto-generada con la manual antes de guardar.
- El endpoint `/api/historial` y su target `#historial-cuerpo` no se modifican.

### 5.2 Botones de Deshacer/Rehacer (Undo/Redo)

**Objetivo:** Permitir al usuario deshacer y rehacer cambios en el formulario de edición sin cerrar la modal ni perder el contexto.

**Alcance:** Opera **solo sobre el estado local del formulario** (el objeto `registro` en Alpine.js). No afecta al servidor.

**Implementación:**

- **Stack de estados:** En `formularioModulo`, mantener un array `history: []` que almacene snapshots del objeto `registro`. Un puntero `historyIndex` indica la posición actual.
- **Captura de cambios:** Usar un watcher profundo (`$watch('registro', handler, {deep: true})`) con debounce de 500ms. Ante cada cambio, truncar historial adelante del índice actual y agregar nuevo snapshot. Límite máximo: 50 entradas (FIFO).
- **Botones en el template:**
  - **Deshacer** (`fas fa-undo`): `:disabled="!puedeDeshacer"`, `@click="deshacer()"`.
  - **Rehacer** (`fas fa-redo`): `:disabled="!puedeRehacer"`, `@click="rehacer()"`.
- **Estilos:** Habilitados: bg-gray-700, text-gray-300. Deshabilitados: opacity-0.4. Tamaño 32×32px, icono 14px.
- **Reset:** Al guardar exitosamente o cerrar el formulario sin guardar, se limpia el `history` y el `historyIndex`.

**Pseudocódigo:**

```
formularioModulo(modulo, registroInicial):
  INIT:
    registro = cloneDeep(registroInicial) || {}
    history = [cloneDeep(registro)]
    historyIndex = 0
    _historyDebounceTimer = null

    $watch('registro', (newVal) =>
      clearTimeout(_historyDebounceTimer)
      _historyDebounceTimer = setTimeout(() =>
        history = history.slice(0, historyIndex + 1)
        history.push(cloneDeep(newVal))
        if history.length > 50 → history.shift()
        historyIndex = history.length - 1
      , 500)
    , {deep: true})

  deshacer():
    IF historyIndex > 0:
      historyIndex--
      Object.assign(registro, cloneDeep(history[historyIndex]))

  rehacer():
    IF historyIndex < history.length - 1:
      historyIndex++
      Object.assign(registro, cloneDeep(history[historyIndex]))

  puedeDeshacer: getter → historyIndex > 0
  puedeRehacer: getter → historyIndex < history.length - 1
  limpiarHistorial():
    history = [cloneDeep(registro)]
    historyIndex = 0
```

**Atributos a preservar:**
- `x-data="formularioModulo('{{.ActiveModule}}', {{jsonEncode .Registro}})"` no cambia.
- `prepararObservaciones()` se ejecuta antes del submit HTMX, igual que ahora.
- `hx-post="/api/guardar-expediente"` y `hx-post="/api/eliminar-expediente"` no se modifican.
- Los botones de deshacer/rehacer son adicionales al markup existente.

### 5.3 Documentos Múltiples — Frontend Only

**Cambios en el formulario:**

- Reemplazar el `<select>` de documento por un **multi-select** visual: chips/tags con botón de eliminar (`fas fa-times`) + dropdown para agregar.
- Cada documento seleccionado se muestra como un chip/badge.
- El array de documentos seleccionados se envía como múltiples valores `id_documento=1&id_documento=2` en el POST.

**Pseudocódigo: Multi-select de documentos (Alpine.js)**

```
En formularioModulo:
  State:
    documentosSeleccionados: [{id: 1, nombre: "CONTRATO"}, ...]

  INIT:
    IF registro.documentos:
      documentosSeleccionados = registro.documentos
    ELSE IF registro.id_documento:
      // Fallback migración
      documentosSeleccionados = [{id: registro.id_documento, nombre: getNombreDoc(...)}]

  agregarDocumento(id, nombre):
    IF !yaExiste(id):
      documentosSeleccionados.push({id, nombre})

  quitarDocumento(id):
    documentosSeleccionados = filter(d => d.id !== id)

  documentosDisponibles: getter
    catalogo completo de cat_documento menos los ya seleccionados
```

**Cambios en la tabla:**
- La columna "Documento" muestra los documentos como badges separados (ej. `CONTRATO` `ACTA`).
- En la subfila expandible, se muestran todos los documentos asociados.
- En el menú flotante de historial, la columna "Documento" muestra los documentos de cada snapshot.

**Atributos a preservar:**
- El endpoint `/api/guardar-expediente` recibe los documentos como múltiples valores `id_documento`.
- El endpoint `/api/cargar-expediente` devuelve el registro con el array de documentos en `registro.documentos`.
- El catálogo `cat_documento` sigue existiendo.

---

## 6. Hoja de Ruta Frontend (Step-by-Step para DeepSeek Flash)

Cada paso es una tarea **autónoma, pequeña y verificable**. No combinar pasos.

### Fase A: Fundación Visual (Riesgo bajo)

- **UI-01:** Crear archivo `theme.css` con todas las variables `:root` (claro) y `[data-theme="dark"]` definidas en la sección 2. Cargarlo después de Tailwind y antes de `styles.css`. No eliminar `styles.css` aún.
- **UI-02:** Implementar toggle de tema claro/oscuro en el header: botón circular 32px con iconos `fas fa-sun` / `fas fa-moon`. Alternar atributo `data-theme` en `<html>` y persistir en `localStorage`. Aplicar antes del primer render para evitar flash.
- **UI-03:** Reestructurar `index.html`: grid CSS de 3 filas (header, main, bottom-bar). Header 48px (44px en <960px), bottom-bar 52px (48px en <960px). Scroll solo dentro de Main.
- **UI-04:** Modernizar bottom bar: crear template `ui_module_button`. En <960px mostrar solo iconos Font Awesome sin texto. Si no caben los 10 botones, agrupar sobrantes bajo botón "Más" (`fas fa-ellipsis-h`) con menú flotante.
- **UI-05:** Modernizar header: en <960px, botones de acción global muestran solo iconos (`fas fa-plus`, `fas fa-download`, `fas fa-thumbtack`, `fas fa-calculator`, etc.) usando atributo `title` para tooltip. Campo de búsqueda colapsa a icono `fas fa-search` que expande input flotante al focus.

### Fase B: Componentes Base DRY (Riesgo medio)

- **UI-06:** Crear carpeta `templates/new/components/` y templates parciales:
  - `ui_button.html` (variantes: primary, secondary, danger, ghost, icon-only; tamaños sm/md; icono Font Awesome; atributos HTMX inyectables)
  - `ui_input.html` (tipos: text, date, number, monto, spin; label, x-model, directivas, slot para adjuntos)
  - `ui_badge.html` (color, texto, icono opcional)
  - `ui_card.html` (título, icono opcional, slot contenido, comportamiento colapsable)
  - `ui_empty_state.html` (mensaje, icono opcional)
- **UI-07:** Crear `ui_modal_shell.html`: layout dual (modal centrado en ≥960px / drawer derecho 100% en <960px). Header con título y botón cerrar (`fas fa-times`). Body con scroll interno. Footer opcional. Backdrop semitransparente. Integrar con `Alpine.store('modals')`.
- **UI-08:** Crear `ui_form_section.html`: reemplaza fieldsets. Tarjeta contenedora con título, icono Font Awesome, capacidad de colapsar. En <960px, iniciar colapsada excepto la primera.
- **UI-09:** Crear `ui_table_detail_panel.html`: unifica `tabla_subrow_trazabilidad` y `tabla_subrow_observaciones_notas`. En ancho completo se inyecta como subfila inline. En <960px, su contenido se renderiza dentro del drawer de Vista Rápida.

### Fase C: Tabla de Datos (Riesgo medio)

- **UI-10:** Crear `ui_data_table.html`: template único parametrizado por metadatos de `ModuloConfig`. Generar columnas iterando array de configuración. Primera columna sticky-left (acciones). Header sticky-top. Contenedor con `overflow-x-auto` y `min-width`. Resizer de columnas condicional (solo ≥960px).
- **UI-11:** Implementar parseo de observaciones en fila principal: mostrar solo la última observación truncada con ellipsis. En subfila expandible: mostrar observación completa + botón "Ver observaciones anteriores" (`fas fa-history`).
- **UI-12:** Implementar menú flotante de historial de observaciones: popover absolute con tabla compacta (columnas: #, Fecha, Documento, Estatus, Observación, Emisor, Receptor). Usar HTMX (`hx-get="/api/historial"`) para inyectar contenido. Cerrar al click fuera o Escape.
- **UI-13:** Implementar drawer de "Vista Rápida" para <960px: al hacer click en una fila, abrir drawer derecho con el contenido de `ui_table_detail_panel` en lugar de expandir subfila inline. Preservar botón de editar y fijar dentro del drawer.
- **UI-14:** Extraer paginación a `ui_pagination.html`. Mantener atributos `hx-get="/api/filtrar-expedientes"`, `hx-vals`, `hx-target="#vista-tabla"`.

### Fase D: Formularios (Riesgo medio)

- **UI-15:** Reescribir `form.html` usando `ui_form_section.html` y `ui_input.html`. Grid CSS: 1 columna en <960px, máximo 3 columnas en ancho completo (`repeat(auto-fill, minmax(240px, 1fr))`).
- **UI-16:** Agrupar pares de montos USD/Bs visualmente: contenedor flex de 2 columnas con separador central mostrando tipo de cambio.
- **UI-17:** Distinguir textarea de observación auto-generada (readonly): fondo `bg-active` con opacidad reducida, borde dashed. Textarea de observaciones manuales: estilo normal.
- **UI-18:** Implementar botones Deshacer/Rehacer (`fas fa-undo`, `fas fa-redo`) en header del formulario. Watcher profundo con debounce 500ms. Stack máximo 50 snapshots. Deshabilitar cuando no aplicable. Resetear al guardar/cerrar.
- **UI-19:** Implementar multi-select de documentos en el formulario: chips con botón eliminar (`fas fa-times`) + dropdown para agregar. Enviar múltiples `id_documento` en POST. Mostrar en tabla como badges separados.
- **UI-20:** Verificar que el formulario en <960px se renderice dentro de `ui_modal_shell` como drawer derecho al 100% de ancho, con padding interno generoso y scroll independiente.

### Fase E: Modales Secundarios y Toasts (Riesgo bajo)

- **UI-21:** Migrar modales secundarios a `ui_modal_shell`: historial, pendientes, recientes, frecuentes, exportar, sumas. Mantener sus `hx-get` y `hx-target` originales.
- **UI-22:** Ajustar `historial.html`, `pendientes.html` para usar `ui_card`, `ui_badge` y `ui_empty_state`.
- **UI-23:** Implementar bottom sheet para modales pequeños (sumas, exportar) en <960px: slide desde abajo en lugar de drawer derecho.
- **UI-24:** Reemplazar sistema de toasts: posición top-right en ≥960px, bottom-center en <960px. Añadir iconos Font Awesome por tipo (`fas fa-check-circle`, `fas fa-exclamation-circle`, `fas fa-info-circle`, `fas fa-exclamation-triangle`). Transiciones de opacity + translate.

### Fase F: Gantt (Riesgo medio)

- **UI-25:** Modernizar `ruta_procesos.html`: aplicar tokens CSS del `theme.css`. Mantener estructura de tabla nativa o migrar a CSS Grid (opcional, puede aplazarse si la tabla actual funciona bien).
- **UI-26:** Implementar sticky left en columna de procesos y sticky top en cabecera de semanas. Reducir densidad en <960px (celdas más angostas, fuentes más pequeñas).
- **UI-27:** Apilar controles de hoja y juntas verticalmente en <960px para no consumir ancho horizontal.
- **UI-28:** Preservar IIFE de JavaScript vanilla y todas sus funciones. Si se migra a CSS Grid, crear función `colorearCeldas()` equivalente basada en `procesos[].timeline`.

### Fase G: Pruebas y Limpieza Final

- **UI-29:** Probar en resoluciones 1280×720 y 640×720. Verificar flujos HTMX (cambio de módulo, búsqueda con debounce, guardado, eliminación con confirmación).
- **UI-30:** Verificar que Alpine inicializa correctamente tras cada swap de HTMX (especialmente en tabla y formulario). Revisar contraste de colores en ambos temas.
- **UI-31:** Una vez validada la estabilidad, eliminar clases obsoletas de `styles.css` o remover su carga por completo, dejando solo `theme.css` + Tailwind.

---

## 7. Glosario de Iconos Font Awesome Obligatorios

| Contexto | Icono | Clase |
|----------|-------|-------|
| Guardar | Floppy disk | `fas fa-save` |
| Eliminar | Trash | `fas fa-trash-alt` |
| Cancelar/Cerrar | Times | `fas fa-times` |
| Editar | Pencil | `fas fa-pen` |
| Fijar | Thumbtack | `fas fa-thumbtack` |
| Nuevo registro | Plus | `fas fa-plus` |
| Exportar | Download | `fas fa-download` |
| Pendientes | Clipboard list | `fas fa-clipboard-list` |
| Recientes | Clock | `fas fa-clock` |
| Frecuentes | Star | `fas fa-star` |
| Sumas | Calculator | `fas fa-calculator` |
| Buscar | Search | `fas fa-search` |
| Ordenar | Sort | `fas fa-sort` |
| Deshacer | Undo | `fas fa-undo` |
| Rehacer | Redo | `fas fa-redo` |
| Historial | History | `fas fa-history` |
| Tema claro | Sun | `fas fa-sun` |
| Tema oscuro | Moon | `fas fa-moon` |
| Más opciones | Ellipsis horizontal | `fas fa-ellipsis-h` |
| Éxito (toast) | Check circle | `fas fa-check-circle` |
| Error (toast) | Exclamation circle | `fas fa-exclamation-circle` |
| Info (toast) | Info circle | `fas fa-info-circle` |
| Advertencia (toast) | Exclamation triangle | `fas fa-exclamation-triangle` |
| Semana agregar | Plus | `fas fa-plus` |
| Semana eliminar | Trash | `fas fa-trash-alt` |
| Proceso agregar | Plus | `fas fa-plus` |
| Proceso eliminar | Trash | `fas fa-trash-alt` |
| Leyenda bloquear | Lock | `fas fa-lock` |
| Leyenda editar | Pencil | `fas fa-pen` |
| Leyenda eliminar | Times | `fas fa-times` |
| Leyenda reordenar | Arrows up/down | `fas fa-arrow-up` / `fas fa-arrow-down` |
| Ruta Procesos | Route / Project diagram | `fas fa-project-diagram` |

---

*Versión UI v2 — Documento para implementación gradual con DeepSeek Flash.*
