# Auditoría Pre-Swap: Migración Alpine.js

App de escritorio **Wails v2** (Go backend, SQLite, Go `html/template` + HTMX + Tailwind).

## Estado actual

La migración a Alpine.js ya se implementó completamente en subcarpetas sin alterar los originales:

```
frontend/new/vendor/
├── alpine.min.js              # Alpine.js v3.14.8
├── alpine-app.js              # Stores + Alpine.data() — modales, fijados, recientes, sumas, exportar, formulario
├── alpine-directives.js       # Directiva x-currency (formato numérico ES)
└── alpine-htmx-bridge.js      # Puente HTMX→Alpine (initTree, pines post-swap)

templates/new/
├── index.html                 # Shell con x-data + $store.modals + @click
├── components.html            # Sub-templates Alpine (form_*_alpine, tabla_*_alpine, filtro superintendencias)
├── form.html                  # Formulario unificado (9 módulos, Go if/eq)
├── tabla.html                 # Tabla unificada (9 módulos, Go if/eq)
└── ruta_procesos.html         # Sin cambios (IIFE Gantt, intencionalmente NO migrado)
```

## Lo que reemplazan

| Antes | Después | Reducción |
|-------|---------|-----------|
| `frontend/vendor/app.js` (765 líneas, JS vainilla) | 3 Alpine JS (448 líneas) | -41% |
| 9 `form_*.html` + 9 `tabla_*.html` (1015 líneas) | 2 unificados (458 líneas) | -55% |
| `templates/index.html` (298 líneas) | `templates/new/index.html` (495 líneas) | +197 (modales inline con x-show) |

## Lo que NO se migró (intencionalmente)

- **Ruta Procesos / Gantt**: mantiene su IIFE propia (~300 líneas en ruta_procesos.html). Tiene estado complejo (processes, timeline, legend, columns), renderizado tabular dinámico y lógica de negocio embebida. Alpine no aporta valor aquí.
- **Paginación cliente**: se deja como JS residual en el index hasta decidir migrar a servidor.
- **Apertura BD**: JS vainilla mínimo (~25 líneas) porque usa binding Wails (`window.go.main.App.AbrirDialogoBD()`). No reemplazable por Alpine.

## Swap pendiente (cambios en handler.go)

1. `template.ParseFS(templateFS, "templates/*.html")` → agregar `"templates/new/*.html"`
2. `handleCargarExpediente`: `"form_" + modulo + ".html"` → `"form.html"`
3. `handleFiltrarExpedientes` + `handleCambiarModulo`: `"tabla_" + modulo + ".html"` → `"tabla.html"`
4. Eliminar `frontend/vendor/app.js`, los 18 templates viejos, y referencias a onclick

## ARCHIVOS A AUDITAR

- `frontend/new/vendor/alpine-app.js`
- `frontend/new/vendor/alpine-directives.js`
- `frontend/new/vendor/alpine-htmx-bridge.js`
- `templates/new/index.html`
- `templates/new/components.html`
- `templates/new/form.html`
- `templates/new/tabla.html`
- `docs/doc.md` (sección "Migración a Alpine.js" al final)

## TAREA

Auditar todo el código nuevo. Buscar:

1. **Errores de sintaxis** en Go templates (llaves, pipes, funciones que no existen)
2. **Referencias rotas**: variables Go que no están en el context del template, helpers que faltan
3. **Alpine mal usado**: `x-data`, `x-model`, `x-show`, `@click`, `$store` bien referenciados
4. **Eventos HTMX → Alpine**: los `hx-on::after-request` acceden a `window.Alpine.store()` correctamente
5. **x-currency vs x-model**: no debe haber conflicto entre la directiva custom y el binding bidireccional
6. **Inicialización de componentes Alpine tras HTMX swap** — `Alpine.initTree()` está en el bridge
7. **form.html**: la variable `$idColumna` se setea con `if/eq` hardcodeado (porque el handler Go no pasa `.Modulos` en el context del form). Verificar que estén todos los módulos.
8. **tabla.html**: columnas y subfilas por módulo — verificar que cada módulo tenga sus columnas correctas y que los helpers `estatusClass`, `formatNum`, `rowGetStr`, `rowGetNum`, `rowGet` existan en el FuncMap de Go.
9. **Compatibilidad con el FuncMap Go**: `jsonEncode`, `default`, `isSelected`, `estatusClass`, `formatNum`, `rowGetStr`, `rowGetNum`, `rowGet`, `dict` — todos disponibles desde handler.go.
10. **Qué falta o está mal** antes de hacer el swap.

## REGLAS

- **No modificar Go backend** (app.go, handler.go). El swap de handler.go es manual y está documentado.
- **Señalar si falta algo** en el checklist de swap.
- **El Gantt (ruta_procesos.html) no se toca**.
- Entregar una lista de issues encontrados con archivo:línea, prioridad (critical/major/minor), y cómo arreglarlo.
- Si no hay issues críticos, dar **go/no-go** para el swap.

## Referencias útiles

- `handler.go:47-80` — FuncMap de Go (funciones disponibles en templates)
- `handler.go:313-329` — PageData struct (variables disponibles en index.html)
- `handler.go:554-610` — handleCargarExpediente (datos pasados a form)
- `handler.go:612-700` — handleFiltrarExpedientes + handleCambiarModulo (datos pasados a tabla)
- `app.go:31-40` — ModuloConfig struct con Columnas, IDColumna, Nombre
- `app.go:42-183` — Definición de los 9 módulos
