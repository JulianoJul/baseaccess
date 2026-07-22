 # Plan de Rediseño Frontend: Modern SaaS Dashboard Bimodal

## 1. Visión del Rediseño (SaaS Dashboard Bimodal)

### 1.1 Principios Estéticos: Minimalismo Funcional

El rediseño abandona la estética de "aplicación de escritorio tradicional" para adoptar un lenguaje visual de **SaaS enterprise moderno**. Cada elemento debe justificar su existencia; no hay decoración sin función.

- **Jerarquía por espaciado y peso, no por bordes.** Se eliminan los bordes dobles, los separadores gruesos y las líneas de fieldset visuales. La separación entre regiones se logra mediante diferencias de fondo (superficies en capas), sombras de elevación sutiles y padding estratégico.
- **Radios de borde consistentes.** Sistema escalonado: 4px para badges y controles pequeños, 6px para inputs y botones, 8px para tarjetas y dropdowns, 12px para paneles y drawers, y 9999px para pills y estados. Nunca mezclar radios arbitrarios en el mismo viewport.
- **Profundidad realista.** En tema claro, la elevación se expresa mediante sombras suaves basadas en luz (más intensas cuanto más alto el elemento). En tema oscuro, la elevación se expresa mediante capas de opacidad y bordes sutilmente más claros, evitando sombras negras que ensucian la interfaz.
- **Opacidad como herramienta de estado.** Los estados deshabilitados, los placeholders y los textos de ayuda no usan grises arbitrarios; usan el color de texto primario con opacidad reducida (40% o 60%). Esto mantiene la armononía cromática en ambos temas.
- **Densidad de información alta pero respirable.** Optimizado para 720px de altura. No hay desperdicio de píxeles verticales. Los inputs son compactos (36px de altura), los márgenes verticales entre secciones se reducen a 12px, y las tarjetas usan padding de 12px en lugar de 24px.

### 1.2 Estrategia DRY General

La base de código actual contiene redundancias significativas que deben eliminarse mediante una **arquitectura de componentes Go template** estricta.

1. **Sistema de Componentes Universal.** Cada patrón visual repetido debe existir en un único template parcial parametrizado. No puede haber más de una definición de "botón primario", "input de texto" o "tarjeta de sección" en toda la aplicación. El Agente Programador debe crear una librería interna de templates: botón, input, badge, card, modal-shell, tabla de datos, sección de formulario, estado vacío y fila de acciones.
2. **Unificación de Módulos de Datos.** Los nueve módulos (expedientes, requisiciones, memorandums, etc.) actualmente generan markup de tabla divergente. Deben converger en un **único template de tabla parametrizado** que reciba la configuración de columnas directamente desde el struct `ModuloConfig` del backend. Las diferencias entre módulos se resuelven mediante arrays de metadatos (nombre, clave, ancho, prioridad, tipo de dato), no mediante bloques condicionales de HTML repetido.
3. **Layout por Slots, no por Copia.** El shell de la aplicación, los modales y los fieldsets del formulario deben convertirse en templates de layout que definan la estructura visual y reciban el contenido específico mediante bloques de contenido (slots). Esto garantiza que un cambio estructural (ej. el comportamiento responsive de un modal) se propague automáticamente a todos los usos.

### 1.3 Estrategia de Responsividad Extrema: 1280px ↔ 640px

La aplicación tiene dos modos de uso reales: **ventana completa** (1280×720) y **panel estrecho** (640×720, compartiendo pantalla con Excel). El diseño no debe "adaptarse a móviles"; debe **optimizarse para una ventana de escritorio estrecha y densa**.

- **Breakpoint crítico único: 960px.** Por encima, la interfaz usa layouts multi-columna y modales centrados. Por debajo, entra en modo Panel Estrecho.
- **Header.** En Panel Estrecho, la altura se reduce a 44px. Los botones de acción global (Nuevo Registro, Exportar, Fijados, Sumas) muestran **solo su icono Font Awesome** (sin texto), utilizando el atributo nativo `title` para tooltip. El campo de búsqueda se colapsa a un icono de lupa (`fas fa-search`) que expande un input flotante al recibir focus, liberando ancho horizontal.
- **Tabla de Datos.** En Panel Estrecho, la tabla **no** se convierte en lista de cards (eso desperdiciaría altura valiosa). Se mantiene como tabla nativa con **scroll horizontal obligatorio**. La primera columna (acciones: expandir, editar, fijar) y la segunda columna (documento/descripción) se vuelven **sticky-left** con fondo opaco para garantizar contexto durante el desplazamiento horizontal. Las columnas secundarias (SOLPED, Gerencia, fechas) permanecen en el área scrolleable. El resizer de columnas se desactiva completamente por debajo de 960px.
- **Subfilas Expandibles.** En ancho completo, se mantiene el patrón actual de subfila inline con colspan. En Panel Estrecho, el click en una fila abre un **drawer lateral de "Vista Rápida"** en lugar de expandir la subfila dentro de la tabla. Esto evita que el layout de tabla se rompa con colspans en espacio reducido y aprovecha los 640px de ancho para mostrar trazabilidad y observaciones sin comprimir.
- **Formulario de Registro.** En ancho completo, puede presentarse como modal centrado o panel inline. En Panel Estrecho, **debe convertirse obligatoriamente en un drawer derecho que ocupe el 100% del ancho disponible**. Un modal centrado en 640px dejaría márgenes inútiles y reduciría el espacio para data entry. El drawer permite maximizar el área de formulario y cerrarse rápidamente.
- **Bottom Bar de Módulos.** Se mantiene fija inferior, pero en Panel Estrecho los botones muestran **solo iconos Font Awesome** sin texto. Si el ancho es insuficiente para los 10 botones, los módulos sobrantes se agrupan bajo un botón "Más" (`fas fa-ellipsis-h`) que despliega un menú flotante compacto.
- **Gantt.** En Panel Estrecho, la columna de nombres de proceso se reduce a 140px y las celdas de día a 28px. La cabecera de semanas y la columna de procesos se vuelven **sticky** (top y left respectivamente). El selector de hojas y los controles de juntas se apilan verticalmente para no consumir ancho horizontal.

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
| bg-base | slate-950 (#020617) | Fondo de aplicación (más profundo que el actual gray-900) |
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
- **Capa 1 (Surface):** Tarjetas, celdas de tabla, bottom bar. Sombra sm en claro, sin sombra en oscuro (se diferencia por el ligero cambio de tono del fondo).
- **Capa 2 (Elevated):** Dropdowns, menús de módulos, tarjetas flotantes. Sombra md.
- **Capa 3 (Overlay):** Modales, drawers, toasts. Sombra lg o xl. En oscuro, el overlay debe incluir un fondo semitransparente muy oscuro (rgba negro al 60%) con opcional backdrop-filter blur de 4px.

### 2.3 Espaciado y Densidad (Optimizado 720p)

Sistema de 4px base. Densidad compacta para maximizar datos visibles en altura limitada.

- **space-1:** 4px (0.25rem) — Separación mínima entre icono y texto.
- **space-2:** 8px (0.5rem) — Gap entre elementos internos de una card, padding vertical de badges.
- **space-3:** 12px (0.75rem) — Padding de tarjetas y modales, padding horizontal de inputs, gap entre fieldsets.
- **space-4:** 16px (1rem) — Padding de drawers, separación entre secciones.
- **space-6:** 24px (1.5rem) — Margen entre regiones principales (usado con moderación).

**Reglas de densidad críticas:**
- Altura mínima de inputs: 36px (2.25rem). Nunca 44px.
- Altura de botones con texto: 36px. Botones de solo icono: 32px.
- Padding de celdas de tabla: 8px vertical, 12px horizontal.
- Padding de tarjetas de formulario: 12px (space-3).

### 2.4 Tipografía

- **Familia:** Inter, system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif.
- **Tamaños:**
  - 10px (0.625rem): Encabezados de Gantt, badges pequeños, labels de densidad extrema.
  - 12px (0.75rem): Labels de formulario, captions, timestamps, texto de ayuda.
  - 13px (0.8125rem): Cuerpo de tabla (ligeramente más pequeño que base para densidad).
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

- **Fila 1 (Header):** Altura fija de 48px (44px en Panel Estrecho). Flexbox con justify-content space-between. Sin borde inferior grueso; usa una sombra sm o un borde sutil de 1px con border-subtle para separación.
- **Fila 2 (Main):** flex-grow 1, overflow-y auto, padding space-3. Esta es la región de contenido donde se renderiza la tabla o el Gantt.
- **Fila 3 (Bottom Bar):** Altura fija de 52px (48px en Panel Estrecho). Fondo bg-surface, borde superior sutil, sombra md para crear separación del contenido. Contiene los botones de módulo.

El drag-and-drop de archivos .db (función de appShell) se mantiene en el nivel del body, pero su indicador visual debe modernizarse a un borde dashed de 2px con color primary y fondo semitransparente, en lugar del outline punteado actual.

### 3.2 Navegación y Bottom Bar

- **Botones de Módulo:** Crear un template parcial único `ui_module_button` que reciba: nombre del módulo, clase de icono Font Awesome (ej. fas fa-folder-open), endpoint hx-get, estado activo (booleano), y badge opcional. Este template se itera para generar todos los botones del bottom bar. En Panel Estrecho, el template omite el texto y renderiza solo el icono centrado en un botón cuadrado compacto.
- **Botón "Más":** Cuando el ancho es insuficiente para todos los módulos, los últimos elementos del array de módulos se agrupan bajo un botón con icono `fas fa-ellipsis-h`. Al hacer click, despliega un menú flotante (Capa 2) con los módulos restantes, cada uno usando el mismo `ui_module_button` pero en formato de lista vertical.
- **Botón de Tema:** Agregar al header un botón circular (32px) con iconos `fas fa-sun` (claro) y `fas fa-moon` (oscuro) que alterne. La preferencia se persiste en localStorage y se aplica al atributo `data-theme` del HTML antes del primer render para evitar flash de tema incorrecto.

### 3.3 Auditoría de Redundancias y Unificación

El Agente Programador debe eliminar las siguientes redundancias identificadas en la documentación actual:

1. **Templates de Input.** Converger `form_input_text_alpine`, `form_input_date_alpine`, `form_input_number_alpine`, `form_input_monto_alpine` y `form_input_spin_alpine` en un único `ui_input_field`. Este template recibe parámetros: tipo de input, nombre, etiqueta, modelo Alpine, directivas adicionales (ej. x-currency), clases de ancho, y un slot para elementos adjuntos (como los botones +/- del spin). El wrapper visual (label, borde, foco, espaciado) es idéntico para todos.
2. **Botones de Acción.** Converger todos los botones de guardar, eliminar, cancelar, exportar, nuevo registro, editar fila y fijar en `ui_action_button`. Parametrizado por: variante visual (primary, secondary, danger, ghost, icon-only), tamaño (sm, md), icono Font Awesome, etiqueta, y un mapa de atributos HTMX que se inyecten directamente en el elemento raíz.
3. **Tablas de Módulos.** Reemplazar las 9 variantes de tabla por un único `ui_data_table`. Las columnas se generan iterando el array `Columnas` (o un nuevo metadato de columnas) del `ModuloConfig`. Cada celda de estatus usa `ui_badge`. Las filas vacías usan `ui_empty_state`.
4. **Modales.** Reemplazar los 8 modales actuales (form, historial, ruta, pendientes, recientes, frecuentes, export, sumas) y los 5 modales del Gantt por un único template de layout `ui_modal_shell`. Este shell define el backdrop, el contenedor (modal centrado o drawer), el header con título y botón de cierre (`fas fa-times`), el cuerpo con scroll interno, y el footer opcional. Cada instancia concreta solo inyecta contenido en los slots.
5. **Fieldsets del Formulario.** Reemplazar por `ui_form_section`, una tarjeta contenedora con título, icono Font Awesome opcional, y capacidad de colapsar en Panel Estrecho. En ancho completo, las secciones pueden permanecer expandidas. En Panel Estrecho, todas excepto la primera deben iniciar colapsadas para reducir el scroll inicial.
6. **Subfilas y Detalles.** Unificar `tabla_subrow_trazabilidad` y `tabla_subrow_observaciones_notas` en `ui_table_detail_panel`. En ancho completo, se inyecta como subfila inline. En Panel Estrecho, su contenido se renderiza dentro del drawer de Vista Rápida.

---

## 4. Guía de Refactorización Crítica

### 4.1 Tablas de Datos

- **Contenedor y Scroll.** Envolver la tabla en un contenedor con overflow-x-auto y un min-width explícito basado en la suma de los anchos de columna definidos en la configuración. Esto fuerza el scroll horizontal en Panel Estrecho en lugar de comprimir columnas hasta la ilegibilidad.
- **Cabecera Sticky.** El thead debe usar position sticky con top 0, z-index 10, y fondo bg-surface sólido (sin transparencia) para evitar efectos de superposición feos al hacer scroll vertical.
- **Columna de Acciones Sticky.** La primera columna (que contiene los botones de expandir, editar y fijar) debe ser position sticky con left 0, z-index 20, y fondo bg-surface sólido. Esto garantiza que el usuario siempre vea las acciones disponibles mientras navega horizontalmente en 640px.
- **Resizer Condicional.** Encapsular la lógica de resize de columnas en una función que verifique `window.innerWidth >= 960` antes de adjuntar event listeners. Por debajo de ese ancho, las columnas usan ancho intrínseco o el ancho definido en el metadato del módulo, y el cursor de resize no aparece.
- **Columnas Responsivas.** Definir una clase de utilidad que oculte columnas secundarias en Panel Estrecho. Las columnas siempre visibles son: Acciones, Documento/Descripción (truncada con ellipsis), Estatus. Los datos restantes (SOLPED, Gerencia, fechas, montos) se trasladan al drawer de Vista Rápida.
- **Paginación.** Extraer la paginación actual a un template parcial `ui_pagination`. Los botones de página mantienen sus atributos hx-get hacia `/api/filtrar-expedientes`, hx-vals con el número de página, y hx-target="#vista-tabla".

### 4.2 Formularios

- **Grid del Formulario.** El contenedor del formulario debe usar CSS Grid con una sola columna en Panel Estrecho y un máximo de tres columnas en ancho completo (usando repeat con minmax de 240px). Esto permite que los campos fluyan automáticamente sin media queries complejas por campo.
- **Secciones Colapsables.** Cada fieldset actual se convierte en una instancia de `ui_form_section`. En Panel Estrecho, todas las secciones excepto "Información General" deben iniciar en estado colapsado. El usuario expande la que necesita, ahorrando scroll vertical crítico en 720px.
- **Agrupación de Montos.** Los pares de inputs de monto (USD/Bs) deben agruparse visualmente en un contenedor flex de dos columnas con un separador central que muestre el tipo de cambio, en lugar de estar dispersos en el grid general. Esto reduce la confusión en data entry rápido.
- **Observación Auto-generada.** Distinguir claramente el textarea de observación auto-generada (solo lectura) del textarea de observaciones manuales. El auto-generado debe tener un fondo distintivo (bg-active con opacidad reducida) y un borde dashed, indicando que es computado, no editable.
- **Drawers en Panel Estrecho.** El formulario de registro debe renderizarse dentro de `ui_modal_shell` configurado como drawer derecho en <960px. Debe ocupar el 100% del ancho para aprovechar los 640px al máximo, con padding generoso interno (space-4) y scroll vertical independiente.

**Atributos obligatorios a preservar en Tablas y Formularios:**
- Todos los `hx-get`, `hx-post`, `hx-target`, `hx-include`, `hx-vals`, `hx-trigger` y `hx-swap` existentes. Específicamente: el trigger de búsqueda (`input changed delay:200ms`), los targets `#vista-tabla`, `#form-expediente`, `#tabla-cuerpo`, y todos los endpoints `/api/filtrar-expedientes`, `/api/cargar-expediente`, `/api/guardar-expediente`, `/api/eliminar-expediente`.
- El `x-data="formularioModulo('{{.ActiveModule}}', {{jsonEncode .Registro}})"` exacto, incluyendo sus parámetros.
- Todos los `x-model` en inputs, los `@change` en fechas que invocan `validarFechas()`, los `@input` en montos que invocan `onMontoInput()`, y la directiva `x-currency`.
- Los métodos `prepararObservaciones()`, `validarAntesGuardar()`, `toggleOrden()`, `appendDias()`, `spinFrente()` y el tracking de `lastSource` para conversión de moneda.
- El `hx-confirm` en el botón de eliminar.

### 4.3 Diagrama de Gantt (Ruta de Procesos)

- **Abandono de Tabla Nativa.** Reestructurar el markup del Gantt para usar un contenedor CSS Grid en lugar de una tabla HTML. Las columnas del grid se definen como: una columna fija de 200px (140px en Panel Estrecho) para los nombres de proceso, seguida de N columnas de 1fr para las semanas. Cada celda de semana internamente usa flex o sub-grid para los 5 días.
- **Sticky Estratégico.** La columna de nombres de proceso es sticky-left. La cabecera que contiene las fechas "desde", "al" y los días de la semana es sticky-top. Esto permite navegar cronogramas largos tanto en ancho como en alto sin perder contexto.
- **Celdas de Entrada.** Las entradas del cronograma se renderizan como pequeños badges o barras de color dentro de la celda correspondiente, usando el color de la leyenda como fondo. Deben tener un mínimo de altura para ser clickeables (28px en ancho completo, 24px en Panel Estrecho).
- **Densidad en Panel Estrecho.** Reducir la fuente de los headers de semana a 10px. Reducir el padding de celdas. Apilar verticalmente los controles de hoja y juntas.

**Atributos obligatorios a preservar en el Gantt:**
- El bloque IIFE de JavaScript vanilla y su parsing inicial de la variable `data` del servidor.
- Todas las funciones internas del IIFE: `toggleModal`, `esc`, `jsonPost`, `reload`, `cambiarHoja`, `eliminarHojaActual`, `crearHoja`, `guardarJunta`, `eliminarJunta`, `crearJunta`, `agregarSemana`, `guardarSemana`, `abrirEliminarSemanas`, `eliminarSemanasConfirmar`, `agregarProceso`, `eliminarProceso`, y todo el sistema de leyendas y cronograma.
- Las llamadas fetch POST a todos los endpoints `/api/ruta-procesos-*`.
- La función `colorearCeldas()` y su lógica de búsqueda de celdas por atributos `data-proc` y `data-fecha`. Si las clases CSS de las celdas cambian (por ejemplo, de `gantt-cell-empty` a una nueva clase base), el programador debe actualizar los selectores dentro de `colorearCeldas` consistentemente, manteniendo la lógica de reemplazo de clases y renderizado de entradas.
- Los atributos HTMX: `hx-get="/api/ruta-procesos"` y `hx-target="#ruta-contenido"`.

### 4.4 Modales, Drawers y Toasts

- **Comportamiento Híbrido.** El `ui_modal_shell` debe implementar una lógica de presentación dual:
  - **Ancho completo (≥960px):** Modal centrado con overlay. Tamaños: 480px para confirmaciones, 720px para formularios complejos, 960px para el Gantt. Overlay con fondo semitransparente oscuro (60% opacidad) y opcional backdrop-filter blur de 4px.
  - **Panel Estrecho (<960px):** Drawer fijo derecho que ocupa el 100% del ancho del viewport. Animación de entrada slide-in desde la derecha (transform translateX a 0). El overlay debe ser más oscuro (70-80% opacidad) para enfocar la atención en el drawer, dado que el usuario está en modo de alta concentración junto a Excel.
- **Botón de Cierre.** Siempre visible en la esquina superior derecha del header del modal/drawer, usando el icono `fas fa-times`. En drawers, debe tener un área de hit extendida (44×44px) para facilitar el cierre táctil/ rápido.
- **Modales Secundarios.** Historial, pendientes, recientes, frecuentes, exportar y sumas deben migrar a `ui_modal_shell`. En Panel Estrecho, los modales pequeños (como sumas o exportar) pueden optar por un bottom sheet (slide desde abajo) en lugar de drawer derecho, si su contenido es vertical y corto.
- **Toasts Responsive.** En ancho completo, posición top-right con ancho máximo de 360px y margen de space-4. En Panel Estrecho, posición bottom-center con ancho 100% menos space-6 de margen horizontal, para no interferir con el drawer ni el bottom bar. Cada toast debe incluir un icono Font Awesome según su tipo: `fas fa-check-circle` para éxito, `fas fa-exclamation-circle` para error, `fas fa-info-circle` para info, `fas fa-exclamation-triangle` para advertencia.

**Atributos obligatorios a preservar:**
- `Alpine.store('modals')` con su array `stack`, y los métodos `abrir`, `cerrar`, `toggle`, `tiene`, `cerrarClickFuera`. El cierre con tecla Escape debe seguir funcionando.
- `Alpine.store('toast')` y sus métodos `mostrar`, `error`, `success`, `info`.
- Todos los `hx-get` que cargan contenido en modales (`/api/historial`, `/api/pendientes`, `/api/cargar-expediente` para editar, etc.) deben mantener sus `hx-target` originales (`#historial-cuerpo`, `#pendientes-contenido`, `#form-expediente`, etc.).
- Los eventos Alpine que invocan `$store.modals.cerrar('id-modal')` deben permanecer intactos en todos los botones de cancelar/cerrar.

### 4.5 Atributos HTMX y Alpine.js Obligatoriamente Preservados (Recopilación Final)

El Agente Programador debe tratar los siguientes elementos como **inmutables en lógica y atributos**, migrándolos exactamente al nuevo markup:

**HTMX:** Todos los atributos `hx-get`, `hx-post`, `hx-target`, `hx-include`, `hx-vals`, `hx-trigger`, `hx-swap`, `hx-confirm` y `hx-indicator` existentes en el documento actual deben migrar sin modificación de endpoint ni de valor de target. Específicamente proteger: los triggers de búsqueda (`input changed delay:200ms`), los targets `#vista-tabla`, `#form-expediente`, `#tabla-cuerpo`, `#historial-cuerpo`, `#pendientes-contenido`, `#ruta-contenido`, `#excel-order-container`, y todos los endpoints bajo `/api/`.

**Alpine.js:** Todos los `x-data` raíz (`appShell`, `bdRecientes`, `fijados`, `calculadoraSumas`, `exportarExcel`, `filtroSuperintendencias`, `formularioModulo`) deben permanecer en sus elementos contenedores con los mismos parámetros de inicialización. Los stores (`modals`, `toast`, `fijados`) deben seguir registrándose en el archivo de inicialización de Alpine. La directiva personalizada `x-currency` debe seguir funcionando en `alpine-directives.js`. El puente `alpine-htmx-bridge.js` debe seguir escuchando el evento `htmx:afterSettle` para invocar `Alpine.initTree(evt.detail.target)` y el evento `htmx:afterSwap` para re-inicializar el store de fijados y actualizar colores de botones pin. Todos los eventos `@click`, `@change`, `@input` y los métodos que invocan (`toggle`, `onMontoInput`, `validarFechas`, `prepararObservaciones`, `spinFrente`, etc.) no deben perderse ni cambiar de nombre.

**Go Templates:** Las funciones de template como `rowGetStr`, `truncate`, `estatusClass`, `default`, `jsonEncode` deben seguir siendo invocables y su comportamiento debe preservarse.

---

## 5. Observaciones y Deshacer/Rehacer

### 5.1 Observaciones: Tabla Compacta + Tabla Flotante de Historial

**Problema actual:** En la subfila expandible de la tabla, el campo `observaciones` muestra todo el texto concatenado (observaciones auto-generadas + manuales, separadas por `\n---\n`). Al haber muchas entradas históricas, esto ocupa mucho espacio vertical y dificulta leer el estado actual.

**Solución propuesta:**

La tabla principal es convencional. Cada fila muestra el registro con la observación más reciente truncada (ellipsis). Al expandir la subfila, se muestra la observación más reciente completa. Al hacer click en "Ver observaciones anteriores", aparece un **menú flotante** (popover/dropdown posicionado junto al botón) que contiene una **tabla pequeña** donde cada fila representa una edición anterior del registro.

**Comportamiento en la tabla principal:**

- **Fila principal:** Mostrar la última línea de observación directamente en la columna de observaciones, truncada a una línea con ellipsis. El texto se obtiene parseando el campo `observaciones` del servidor (extraer la última entrada después del último `\n---\n`).
- **Subfila expandible (click en la fila):** Mostrar la observación más reciente completa (sin truncar) en una sección destacada. Debajo, un botón "Ver observaciones anteriores" (`fas fa-history`).

**Menú flotante de historial de observaciones:**

- Al hacer click en "Ver observaciones anteriores", se despliega un **menú flotante** (popover con posición absolute, ancho máximo ~500px, fondo `bg-surface-elevated`, borde sutil, sombra `shadow-lg`, scroll interno si excede la altura).
- El menú contiene una **tabla compacta** con las siguientes columnas:
  - **#** — Número de edición (1 = más antigua, N = más reciente)
  - **Fecha** — Timestamp de la edición (del campo `fecha_creacion` o `fecha_modificacion` del historial)
  - **Documento** — Valor del campo `id_documento` en esa edición
  - **Estatus** — Valor del campo `id_estatus` en esa edición
  - **Observación** — Texto de la observación en esa edición (truncado con ellipsis, expandible al hover o click)
  - **Emisor** — Valor del campo `id_emisor` en esa edición
  - **Receptor** — Valor del campo `id_receptor` en esa edición
- Cada fila de la tabla de historial representa **una edición completa del registro**, no solo un cambio de observación. Si se editó el registro 7 veces, hay 7 filas.
- La tabla de historial se ordena de más reciente a más antigua (la fila 1 es la más antigua, la última es la anterior a la actual).
- Cerrar el menú al hacer click fuera de él o al presionar Escape.

**Pseudocódigo: Parseo de observación más reciente**

```
parseObservacionReciente(textoCompleto):
  IF textoCompleto contiene "\n---\n":
    partes = split(textoCompleto, "\n---\n")
    ultimaParte = partes[len(partes)-1].trim()
    historialAnterior = partes[0:len(partes)-1]
    RETURN {reciente: ultimaParte, anteriores: historialAnterior}
  ELSE:
    RETURN {reciente: textoCompleto, anteriores: []}
```

**Pseudocódigo: Menú flotante de historial**

```
Al hacer click en "Ver observaciones anteriores":
  IF menú ya está abierto → cerrar menú, return
  Cargar datos del historial:
    GET /api/historial?modulo={{.ActiveModule}}&id={{.ID}}
    Parsear respuesta HTML o solicitar JSON con los campos relevantes
  Renderizar tabla en el menú flotante:
    Para cada fila del historial (excluyendo la última = la actual):
      <tr>
        <td>{{index}}</td>
        <td>{{fecha_modificacion}}</td>
        <td>{{documento}}</td>
        <td>{{estatus}}</td>
        <td>{{observaciones | truncate 60}}</td>
        <td>{{emisor}}</td>
        <td>{{receptor}}</td>
      </tr>
  Posicionar el menú junto al botón (absolute, top/right calculados)
  Mostrar menú con animación fade-in

Al hacer click fuera del menú → cerrar menú
Al presionar Escape → cerrar menú
```

**Nota sobre el endpoint de historial:** El endpoint actual `/api/historial` devuelve HTML. Para el menú flotante, se puede:
- Opción A: Usar HTMX para inyectar el HTML del historial directamente en el menú flotante (`hx-get="/api/historial"`, `hx-target="#menu-historial-{{id}}"`).
- Opción B: Crear un nuevo endpoint `/api/historial-json` que devuelva JSON y renderizar la tabla con Alpine.js.
- **Recomendación:** Opción A (reutilizar el endpoint existente) para mantener DRY. El HTML devuelto por `/api/historial` ya incluye una tabla con las columnas necesarias. Se puede reutilizar ese HTML dentro del menú flotante, ajustando solo los estilos para que sea compacto.

**Atributos a preservar:**
- El campo `observaciones` en la base de datos sigue siendo un solo campo concatenado con separador `\n---\n`. El parseo se hace en el frontend al momento de renderizar.
- El método `prepararObservaciones()` en `formularioModulo` sigue concatenando la observación auto-generada con la manual antes de guardar. No se modifica la lógica de escritura.
- El endpoint `/api/historial` y su target `#historial-cuerpo` no se modifican. El menú flotante es un uso adicional del mismo endpoint.

### 5.2 Botones de Deshacer/Rehacer (Undo/Redo)

**Objetivo:** Permitir al usuario deshacer y rehacer cambios en el formulario de edición sin cerrar la modal ni perder el contexto. Esto es especialmente útil para operaciones como:
- Deshacer un cambio accidental en un campo de monto
- Rehacer una observación que se borró por error
- Revertir una selección de catálogo equivocada

**Alcance:** El sistema de deshacer/rehacer opera **solo sobre el estado local del formulario** (el objeto `registro` en Alpine.js). No afecta al servidor. El usuario puede deshacer cambios en memoria y luego decidir si guarda o no.

**Implementación propuesta:**

- **Stack de estados:** En el componente `formularioModulo`, mantener un array `history: []` que almacene snapshots del objeto `registro` en cada cambio significativo. Un puntero `historyIndex` indica la posición actual en el historial.

- **Captura de cambios:** Usar un watcher profundo (`$watch('registro', handler, {deep: true})`) que, ante cualquier cambio en un campo, pushee un clon profundo del `registro` al `history` (truncando todo lo que esté adelante del `historyIndex` si el usuario estaba en un punto intermedio del historial). Implementar un debounce de 500ms para no crear un snapshot por cada tecla en un input de texto.

- **Botones en el formulario:** Agregar dos botones en el header o footer del formulario:
  - **Deshacer** (`fas fa-undo`): Decrementa `historyIndex`, restaura el snapshot anterior al objeto `registro` y actualiza todos los inputs del DOM. Se desactiva cuando `historyIndex <= 0`.
  - **Rehacer** (`fas fa-redo`): Incrementa `historyIndex`, restaura el snapshot siguiente. Se desactiva cuando `historyIndex >= history.length - 1`.

- **Límite de historial:** Máximo 50 entradas. Al exceder, se elimina la entrada más antigua (FIFO).

- **Reset al guardar/cerrar:** Al guardar exitosamente o cerrar el formulario sin guardar, se limpia el `history` y el `historyIndex`.

**Pseudocódigo: Componente formularioModulo extendido**

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
        // Truncar historial adelante del índice actual
        history = history.slice(0, historyIndex + 1)
        // Agregar nuevo snapshot
        history.push(cloneDeep(newVal))
        // Limitar a 50 entradas
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

  // Reset al guardar o cerrar
  limpiarHistorial():
    history = [cloneDeep(registro)]
    historyIndex = 0
```

**Botones en el template:**

```
En el header del formulario (junto al toggle "Orden Excel"):
  Botón Deshacer:
    :disabled="!puedeDeshacer"
    @click="deshacer()"
    title="Deshacer"
    Icono: fas fa-undo
  Botón Rehacer:
    :disabled="!puedeRehacer"
    @click="rehacer()"
    title="Rehacer"
    Icono: fas fa-redo

Estilos:
  - Habilitados: bg-gray-700, text-gray-300, hover:bg-gray-600
  - Deshabilitados: opacity-0.4, cursor-not-allowed, pointer-events-none
  - Tamaño: 32x32px, icono 14px
```

**Integración con el guardado:**

- El botón "Guardar" (`hx-post="/api/guardar-expediente"`) no se modifica. Guarda el estado actual del `registro` en el servidor.
- Al recibir respuesta exitosa del servidor, se ejecuta `limpiarHistorial()`.
- Al cerrar la modal sin guardar (botón Cancelar o Escape), se ejecuta `limpiarHistorial()` sin guardar en servidor.

**Atributos a preservar:**
- El `x-data="formularioModulo('{{.ActiveModule}}', {{jsonEncode .Registro}})"` no cambia.
- El `prepararObservaciones()` se ejecuta antes del submit HTMX, igual que ahora.
- El `hx-post="/api/guardar-expediente"` y `hx-post="/api/eliminar-expediente"` no se modifican.
- Los botones de deshacer/rehacer son adicionales al markup existente, no reemplazan nada.

### 5.3 Documentos: Selección Múltiple (Backend + Frontend)

**Problema actual:** La tabla `cat_documento` contiene entradas compuestas como "CONTRATO Y ACTA", "OFICIO Y CONTRATO", etc. Esto genera duplicidad semántica y dificulta búsquedas/filtros por tipo de documento. El campo `id_documento` en cada módulo es un solo FK, lo que limita a un solo documento por registro.

**Solución propuesta:** Permitir que un registro tenga **múltiples documentos** asociados. Cada tipo de documento en `cat_documento` debe ser único (CONTRATO, ACTA, OFICIO, SOLPED, etc.). En el formulario, el usuario puede seleccionar cuantos documentos desee.

#### Cambios en Backend (Base de Datos + Go)

**Paso 1: Limpiar `cat_documento`**

Eliminar las entradas compuestas de `cat_documento` y dejar solo tipos únicos:

```
ANTES (ejemplo):
  id=1  CONTRATO
  id=2  ACTA
  id=3  CONTRATO Y ACTA        ← eliminar
  id=4  OFICIO
  id=5  OFICIO Y CONTRATO      ← eliminar

DESPUÉS:
  id=1  CONTRATO
  id=2  ACTA
  id=4  OFICIO
```

Los registros existentes en `expedientes` (y otros módulos) que usen `id_documento=3` (CONTRATO Y ACTA) deben migrarse: crear dos entradas en la tabla de unión, una con `id=1` (CONTRATO) y otra con `id=2` (ACTA).

**Paso 2: Crear tabla de unión `expediente_documentos` (y equivalentes por módulo)**

```
Tabla: expediente_documentos
  id_expediente_documento  INTEGER PRIMARY KEY AUTOINCREMENT
  id_expediente            INTEGER NOT NULL → FK expedientes(id_expediente)
  id_documento             INTEGER NOT NULL → FK cat_documento(id)
  UNIQUE(id_expediente, id_documento)
```

Equivalente para cada módulo que use documentos (requisiciones, memorandums, recobros, valuaciones, aprobacion_jd, certificacion_bdu, vacaciones, reposos_medicos). Cada módulo tendría su propia tabla de unión o una tabla genérica:

```
Tabla genérica (alternativa):
  modulo_documento
    id_modulo_documento  INTEGER PRIMARY KEY AUTOINCREMENT
    modulo               TEXT NOT NULL       ← clave del módulo (ej: "expedientes")
    id_registro          INTEGER NOT NULL    ← ID del registro en su tabla
    id_documento         INTEGER NOT NULL → FK cat_documento(id)
    UNIQUE(modulo, id_registro, id_documento)
```

**Recomendación:** Usar la tabla genérica `modulo_documento` para evitar crear 9 tablas de unión. Es más DRY y más fácil de mantener.

**Paso 3: Modificar `GuardarFila` en `app.go`**

- El campo `id_documento` deja de ser un solo valor en la tabla del módulo. Se elimina la columna `id_documento` de las tablas de módulos (o se ignora si se quiere mantener compatibilidad temporal).
- Al guardar, el backend recibe un array de IDs de documentos (ej: `id_documento=1&id_documento=2`) y los inserta/actualiza en `modulo_documento`.
- Al leer, el backend consulta `modulo_documento` para obtener los documentos asociados y los incluye en la respuesta.

**Paso 4: Modificar endpoints de lectura**

- `ObtenerFilas`, `ObtenerFilasPaginado`, `ObtenerFilaPorId`: Incluir los documentos como un array en la respuesta (ej: `registro["documentos"] = [{id: 1, nombre: "CONTRATO"}, {id: 2, nombre: "ACTA"}]`).
- `preparePageData`: Incluir el catálogo de documentos en `Catalogs` (ya lo hace actualmente).
- Las vistas SQL (`vw_reporte_*`) necesitan un JOIN adicional o un GROUP_CONCAT para mostrar los documentos en la vista de tabla.

**Paso 5: Modificar `HistorialTabla`**

El historial de movimientos (`historial_movimientos`) actualmente almacena el estado completo del registro. Al migrar a documentos múltiples, el historial debe almacenar los documentos asociados en cada snapshot (o consultar `modulo_documento` con el timestamp del historial).

#### Cambios en Frontend (Formulario + Tabla)

**En el formulario (`form.html` / `formularioModulo`):**

- Reemplazar el `<select>` de documento por un **multi-select** o **lista de checkboxes** con chips/tags.
- Cada documento seleccionado se muestra como un chip (badge) con botón de eliminar (fas fa-times).
- El usuario puede agregar documentos desde un dropdown select. Al seleccionar un documento, se agrega como chip y se marca como seleccionado.
- El array de documentos seleccionados se envía como múltiples valores `id_documento=1&id_documento=2` en el POST.

**Pseudocódigo: Multi-select de documentos**

```
En formularioModulo:
  State:
    documentosSeleccionados: [{id: 1, nombre: "CONTRATO"}, {id: 2, nombre: "ACTA"}]

  INIT:
    IF registro.documentos:
      documentosSeleccionados = registro.documentos
    ELSE IF registro.id_documento:
      // Migración: convertir id_documento único a array
      documentosSeleccionados = [{id: registro.id_documento, nombre: getNombreDoc(registro.id_documento)}]

  agregarDocumento(id, nombre):
    IF !documentosSeleccionados.find(d => d.id === id):
      documentosSeleccionados.push({id, nombre})

  quitarDocumento(id):
    documentosSeleccionados = documentosSeleccionados.filter(d => d.id !== id)

  documentosDisponibles: getter
    catalogo completo de cat_documento menos los ya seleccionados
```

**En la tabla (`tabla.html`):**

- La columna "Documento" muestra los documentos como badges separados (ej: `CONTRATO` `ACTA` en lugar de `CONTRATO Y ACTA`).
- En la subfila expandible, se muestran todos los documentos asociados.
- En el menú flotante de historial, la columna "Documento" muestra los documentos de cada snapshot.

**Atributos a preservar:**
- El endpoint `/api/guardar-expediente` recibe los documentos como múltiples valores `id_documento`. El backend parsea el array.
- El endpoint `/api/cargar-expediente` devuelve el registro con el array de documentos en `registro.documentos`.
- El catálogo `cat_documento` sigue existiendo, pero limpio de entradas compuestas.
- Los filtros por documento en la tabla (si existen) deben manejar múltiples documentos por registro.

**Impacto en otros módulos:**

El cambio afecta a los 9 módulos que usan `id_documento`. La tabla genérica `modulo_documento` centraliza esta relación. El formulario de cada módulo debe usar el mismo componente multi-select de documentos. La migración de datos existentes (entradas compuestas → múltiples entradas) debe ejecutarse una sola vez.

---

## 6. Hoja de Ruta de Ejecución (Step-by-Step)

**Paso 1: Preparación y Backup.** Crear una copia de seguridad completa de la carpeta `templates/new/` y del archivo `styles.css`. Verificar que Alpine.js v3.14.8, HTMX y Font Awesome Free 7.3 permanezcan cargados en `index.html`. No actualizar versiones de librerías a menos que sea estrictamente necesario para compatibilidad.

**Paso 2: Fundación de Tokens CSS.** Crear un nuevo archivo de tema (por ejemplo, `theme.css`) que contenga las variables `:root`, `[data-theme="dark"]`, y las clases de utilidad del design system (surface, elevated, input, text-primary, etc.). Este archivo se cargará después de Tailwind y antes del CSS legacy. No eliminar `styles.css` aún; superponer y deprecar progresivamente.

**Paso 3: Templates Base DRY.** Crear los templates parciales fundamentales en una subcarpeta `templates/new/components/`: `ui_button.html`, `ui_input.html`, `ui_badge.html`, `ui_card.html`, `ui_modal_shell.html`, `ui_empty_state.html`, `ui_form_section.html`. Cada uno debe ser parametrizado mediante el sistema `dict` de Go templates y bloques de contenido.

**Paso 4: Refactorización del appShell (`index.html`).** Reestructurar el body con el grid de tres filas. Implementar el header compacto con comportamiento de iconos-solo en <960px. Implementar el bottom bar responsive con scroll interno o botón "Más". Integrar un toggle de tema claro/oscuro (iconos `fas fa-sun` / `fas fa-moon`) que alterne `data-theme` en el elemento HTML y persista la preferencia en `localStorage`.

**Paso 5: Refactorización de la Tabla de Datos + Observaciones.** Reemplazar `tabla.html` por `ui_data_table.html` unificado. Implementar sticky headers, sticky columnas de acción, scroll horizontal controlado, y resizer condicional (solo ≥960px). Crear `ui_table_detail_panel` para el contenido de subfilas. Configurar el modo Panel Estrecho para que el click en fila abra un drawer de Vista Rápida en lugar de expandir subfila inline. Asegurar que los 9 módulos usen este único template, alimentado por metadatos de `ModuloConfig`. Implementar el parseo de observaciones: en la fila principal mostrar solo la última observación truncada (ellipsis); en la subfila expandible mostrar la última completa + botón "Ver observaciones anteriores" que despliegue un menú flotante con una tabla compacta del historial de ediciones (cada fila = una edición anterior del registro, con columnas: #, Fecha, Documento, Estatus, Observación, Emisor, Receptor). Reutilizar el endpoint existente `/api/historial` vía HTMX para poblar el menú.

**Paso 6: Refactorización del Sistema de Formularios + Deshacer/Rehacer.** Reemplazar `form.html` y `components.html` por `ui_form.html` y `ui_form_section.html`. Unificar todos los tipos de input en `ui_input_field`. Implementar el grid responsive en las secciones. Configurar el formulario para que use `ui_modal_shell` con comportamiento drawer en <960px. Implementar los botones de deshacer/rehacer (`fas fa-undo` / `fas fa-redo`) en el header del formulario, con stack de snapshots del objeto `registro` (máximo 50 entradas, debounce 500ms). Verificar que todos los `x-model`, `@change` y `@input` se trasladen correctamente al nuevo markup. Verificar que `prepararObservaciones()` sigue concatenando correctamente antes del submit HTMX.

**Paso 7: Documentos Múltiples (Backend + Frontend).** Crear tabla genérica `modulo_documento` (modulo, id_registro, id_documento). Limpiar `cat_documento` eliminando entradas compuestas. Migrar datos existentes (entradas compuestas → múltiples entradas en la tabla de unión). Modificar `GuardarFila` para insertar/actualizar documentos en `modulo_documento`. Modificar `ObtenerFilas`/`ObtenerFilaPorId` para incluir array de documentos. Crear componente multi-select de documentos en el formulario (chips con botón de eliminar + dropdown para agregar). Actualizar la columna "Documento" en la tabla para mostrar badges separados. Actualizar el historial y el menú flotante de observaciones para reflejar documentos múltiples.

**Paso 8: Refactorización de Modales Secundarios.** Migrar `historial.html`, `pendientes.html`, y los modales de recientes, frecuentes, exportar y sumas a `ui_modal_shell`. Asegurar que sus `hx-get` y `hx-target` originales se mantengan. Ajustar sus contenidos internos para usar `ui_card`, `ui_badge` y `ui_empty_state` donde aplique.

**Paso 9: Refactorización del Gantt.** Modernizar `ruta_procesos.html` para usar CSS Grid en lugar de tabla nativa. Implementar sticky left en la columna de procesos y sticky top en la cabecera de semanas. Reducir densidad visual en <960px (celdas más angostas, fuentes más pequeñas, controles apilados). Preservar el IIFE de JavaScript vanilla y todas sus funciones. Actualizar únicamente los selectores de `colorearCeldas()` si las clases CSS base de las celdas cambian de nombre.

**Paso 10: Sistema de Toasts y Micro-interacciones.** Reemplazar el contenedor de toasts actual por el nuevo posicionamiento responsive. Añadir iconos Font Awesome a cada tipo de toast. Implementar transiciones de entrada y salida suaves (combinación de opacity y translate). Verificar que `Alpine.store('toast')` siga siendo el único punto de invocación desde el resto de la aplicación.

**Paso 11: Pruebas de Responsividad y Depuración.** Probar la interfaz explícitamente en resoluciones 1280×720 y 640×720. Verificar que todos los flujos HTMX funcionan sin errores (cambio de módulo, búsqueda con debounce, guardado, eliminación con confirmación). Verificar que Alpine inicializa correctamente tras cada swap de HTMX (especialmente en tabla y formulario). Revisar contraste de colores en ambos temas para accesibilidad. Una vez validada la estabilidad, eliminar las clases obsoletas del `styles.css` antiguo o remover su carga por completo.