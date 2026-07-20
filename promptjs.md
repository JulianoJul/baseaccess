# Contexto: Migración a Alpine.js

App de escritorio con **Wails v2** (Go backend nativo, SQLite, Go `html/template` + HTMX + Tailwind CSS).

**Stack destino:** HTMX para interacción servidor + Alpine.js para UI state local + JS vainilla residual mínimo (solo lógica de negocio que Alpine no puede expresar). **Cero gluecode** — el AI no debe escribir funciones JS, solo atributos HTML y definiciones `Alpine.data()` reutilizables.

**No tocar Go backend** (`app.go`, `handler.go`). No romper HTMX. Alpine se copia a `frontend/vendor/alpine.min.js`.

---

## OBJETIVO ADICIONAL: UNIFICAR TEMPLATES

Además de migrar el JS, se busca **reducir los 18 templates** (9 `form_*.html` + 9 `tabla_*.html`) a **2 archivos** (`form.html` + `tabla.html`) o incluso 1 solo.

Cada módulo tiene los mismos campos en su tabla y formulario, solo varían las columnas. La personalización por módulo se resuelve dentro del mismo archivo sin separar:

### Con Go template (if/else server-side)

```html
{{if eq $.ModuloKey "expedientes"}}
  <input name="presupuesto_base_usd" inputmode="decimal">
  <input name="tipo_cambio" oninput="convertirMoneda()">
{{else if eq $.ModuloKey "recobros"}}
  <input name="costo_servicio_usd" inputmode="decimal">
{{else if eq $.ModuloKey "vacaciones"}}
  <input name="dias_solicitados" type="number">
{{end}}
```

O más limpio: `{{template (printf "extra_%s" $.ModuloKey) .}}` con cada sección definida en el mismo archivo.

### Con Alpine (client-side, más declarativo)

```html
<div x-data="{ modulo: '{{$.ModuloKey}}' }">
  <input x-show="modulo === 'expedientes' || modulo === 'recobros'"
         name="presupuesto_base_usd" inputmode="decimal">
  <input x-show="modulo === 'vacaciones'"
         name="dias_solicitados" type="number">
</div>
```

Esto **elimina la necesidad de tener 18 archivos** y es perfecto para AI: escribe HTML con atributos, no gluecode.

Las columnas de tabla se iteran con `{{range $col := .ColsMostrar}}` usando `cfg.Columnas` que ya está en Go, eliminando los `<th>` y `<td>` hardcodeados por módulo.

---

## ARCHIVOS A LEER

- `frontend/vendor/app.js` — ~770 líneas JS vainilla actual
- `templates/index.html` — template principal (modales, botones)
- `templates/components.html` — componentes reutilizables
- `templates/ruta_procesos.html` — IIFE del Gantt (contiene su propio JS)
- `docs/doc.md` — documentación estructural
- `docs/funciones.md` — catálogo de funciones

---

## ENTREGABLE 1: `docs/legacy/js_analysis.md`

Analiza **exhaustivamente** el JS actual y produce este archivo con:

### A. Catálogo de funciones JS

Para cada función en `app.js`, documenta:
- **Nombre**
- **Línea** en app.js
- **Qué hace** (descripción breve)
- **Cómo se activa** (onclick, hx-on, evento DOM, etc.)
- **Estado que maneja** (variable global, localStorage, DOM)
- **Template(s) donde se invoca** (archivo:línea)

Agrupa por categoría: Modales, Pines/Fijados, BD Recientes, Paginación, Exportar, Sumas, Campos Numéricos, Superintendencias, Conversión USD/Bs, Helpers.

### B. Estado global y localStorage

Lista completa de:
- Variables globales (`PAGE_DATA`, `MODAL_STACK`, `currentPage`, `_superCache`, `_convLock`, etc.)
- Claves de localStorage (`sidebarFrecuentes`, `baseaccess_recientes`)
- Variables de closure en la IIFE del Gantt

### C. Mapa de interacción HTML → JS

Para cada template, lista qué eventos HTML disparan qué funciones JS. Ej:

```
index.html
├── onclick="abrirBaseDatos()"       → app.js:49
├── onclick="abrirRecientes()"        → app.js:231
├── onclick="toggleSortDir()"         → app.js:461
├── hx-on::after-request="pushModal"  → app.js:136
...
```

---

## ENTREGABLE 2: `plan_migracion_alpine.md`

Genera un plan de migración priorizado. Para cada fase:

### Formato de cada entrada

```
## Fase N: [Nombre]

### Funciones a migrar
- [lista de funciones de app.js]

### Patrón Alpine
[Pseudocódigo HTML + Alpine concreto]

### Estado reemplazado
[Variables globales o DOM que Alpine elimina]

### JS residual
[Lo que queda como JS vainilla y por qué]

### Archivos a modificar
[rutas de archivos]
```

### Priorización (de mayor a menor ROI)

1. **Modales** (stack: `pushModal`/`cerrarModal`/backdrop) — el mayor boilerplate JS
2. **Pines/Fijados** (`toggleFrecuente`, `abrirFrecuentes`) — Alpine.data + localStorage
3. **BD Recientes** (`registrarReciente`, `abrirRecientes`) — mismo patrón
4. **Sumas** (`anyadirFilaSuma`, `calcularSumas`) — x-model + x-for
5. **Paginación cliente** (`irPagina`, `renderPaginacionControles`) — evaluar migrar a servidor vs Alpine
6. **Exportar** (`cargarColumnasExportar`, modal de selección) — solo UI, fetch queda igual
7. **Campos numéricos / Superintendencias / Conversión USD/Bs / Gantt** — baja prioridad o no migrar

Cada fase debe incluir pseudocódigo HTML+Alpine concreto y viable, no genérico.

---

## REGLAS

- **Cero gluecode:** no escribir funciones JS nuevas. Si hay lógica que no se puede expresar en atributos HTML + `Alpine.data()`, marcarla como JS residual y explicar por qué.
- **Alpine convive con HTMX:** Alpine maneja estado UI local, HTMX maneja toda comunicación servidor.
- **No modificar Go.**
- Al terminar, presentar ambos archivos sin ejecutar nada.
