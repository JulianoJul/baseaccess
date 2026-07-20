# Auditoría Pre-Swap: Migración Alpine.js

App de escritorio **Wails v2** (Go backend, SQLite, Go `html/template` + HTMX + Tailwind).

## Estado actual

La migración a Alpine.js ya se implementó completamente en subcarpetas sin alterar los originales.
Además, se preparó el backend nuevo en `backend/new/` con soporte de paginación en el servidor.

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
├── tabla.html                 # Tabla unificada (9 módulos, Go if/eq) + controles de paginación server-side
└── ruta_procesos.html         # Sin cambios (IIFE Gantt, intencionalmente NO migrado)

backend/new/
├── app.go                     # Añade ObtenerFilasPaginado (LIMIT/OFFSET en SQLite)
└── handler.go                 # Añade pagRange en FuncMap, paginación en preparePageData / handleFiltrarExpedientes / handleCambiarModulo
```

## Lo que reemplazan

| Antes | Después | Reducción |
|-------|---------|-----------| 
| `frontend/vendor/app.js` (765 líneas, JS vainilla) | 3 Alpine JS (448 líneas) | -41% |
| 9 `form_*.html` + 9 `tabla_*.html` (1015 líneas) | 2 unificados (458 líneas) | -55% |
| `templates/index.html` (298 líneas) | `templates/new/index.html` (495 líneas) | +197 (modales inline con x-show) |
| Paginación DOM en JS (app.js) | Paginación en servidor (tabla.html + handler.go + app.go) | Cero JS cliente |

## Lo que NO se migró (intencionalmente)

- **Ruta Procesos / Gantt**: mantiene su IIFE propia (~300 líneas en ruta_procesos.html). Tiene estado complejo (processes, timeline, legend, columns), renderizado tabular dinámico y lógica de negocio embebida. Alpine no aporta valor aquí.
- **Apertura BD**: JS vainilla mínimo (~25 líneas) porque usa binding Wails (`window.go.main.App.AbrirDialogoBD()`). No reemplazable por Alpine.

## Cambios ya aplicados en los archivos actuales

Los siguientes cambios ya están commiteados (diferencia respecto al baseline de la rama):

- **`handler.go`**: Se añadió `pagRange` al FuncMap (línea ~116).
- **`templates/new/index.html`**: `alpine-htmx-bridge.js` movido al `<head>` (antes estaba antes del cierre de `</body>`). Se eliminó el `<div id="paginacion">` vacío del layout.
- **`templates/new/tabla.html`**: Se añadieron controles de paginación server-side al final (usando `pagRange`, `TotalPages`, `CurrentPage`). Se añadió `IDColumna` al dict de `tabla_subrow_trazabilidad`.
- **`templates/new/form.html`**: `formularioModulo` ahora recibe 3 argumentos: `('{{.ActiveModule}}', {{jsonEncode .Registro}}, 0)`.
- **`templates/new/components.html`**: Botón editar usa `(index $.Modulos $.ActiveModule).Nombre` en lugar de `.ModuloLabel`.
- **`backend/new/app.go`**: Nuevo método `ObtenerFilasPaginado(moduloKey, orden, pagina, pageSize)` — COUNT + SELECT con LIMIT/OFFSET.
- **`backend/new/handler.go`**: `preparePageData`, `handleFiltrarExpedientes` y `handleCambiarModulo` leen `?pagina=` y `?page_size=` y pasan `CurrentPage`, `TotalPages`, `PageSize` al template.

## Swap pendiente (cambios en handler.go raíz)

El backend nuevo está en `backend/new/` listo para copiar. Los cambios concretos son:

1. `template.ParseFS(templateFS, "templates/*.html")` → agregar `"templates/new/*.html"`
2. `pagRange` ya añadido al FuncMap en handler.go raíz (línea ~116)
3. `handleCargarExpediente`: `"form_" + modulo + ".html"` → `"form.html"`
4. `handleFiltrarExpedientes` + `handleCambiarModulo`: `"tabla_" + modulo + ".html"` → `"tabla.html"` y adoptar lógica de paginación de `backend/new/handler.go`
5. Copiar o incorporar `ObtenerFilasPaginado` de `backend/new/app.go` a `app.go`
6. Eliminar `frontend/vendor/app.js`, los 18 templates viejos, y referencias a onclick

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
7. **form.html**: la variable `$idColumna` se setea con `if/eq` hardcodeado. Verificar que estén todos los módulos. `formularioModulo` ahora recibe 3 args (`modulo`, `registro`, `tipoCambioInicial`).
8. **tabla.html**: columnas y subfilas por módulo — verificar que cada módulo tenga sus columnas correctas y que los helpers `estatusClass`, `formatNum`, `rowGetStr`, `rowGetNum`, `rowGet` existan en el FuncMap. Verificar que `IDColumna` llega correctamente a `tabla_subrow_trazabilidad`.
9. **Paginación server-side en tabla.html**: verificar que `pagRange`, `TotalPages`, `CurrentPage`, `add`, `sub` están disponibles en el FuncMap y en el context del template tanto para `handleFiltrarExpedientes` como `handleCambiarModulo`.
10. **Compatibilidad con el FuncMap Go**: `jsonEncode`, `default`, `isSelected`, `estatusClass`, `formatNum`, `rowGetStr`, `rowGetNum`, `rowGet`, `dict`, `pagRange` — todos disponibles desde handler.go.
11. **Qué falta o está mal** antes de hacer el swap.

## REGLAS

- **No modificar Go backend** raíz (app.go, handler.go). Los cambios del backend están preparados en `backend/new/` para ser aplicados manualmente.
- **Señalar si falta algo** en el checklist de swap.
- **El Gantt (ruta_procesos.html) no se toca**.
- Entregar una lista de issues encontrados con archivo:línea, prioridad (critical/major/minor), y cómo arreglarlo.
- Si no hay issues críticos, dar **go/no-go** para el swap.

## Referencias útiles

- `handler.go:47-80` — FuncMap de Go (funciones disponibles en templates; incluye `pagRange` en línea ~116)
- `handler.go:335-351` — PageData struct (variables disponibles en index.html)
- `backend/new/handler.go:353-406` — `preparePageData` actualizado (lee `?pagina=`, llama `ObtenerFilasPaginado`)
- `backend/new/handler.go:640-735` — `handleFiltrarExpedientes` actualizado (paginación en memoria post-filtro)
- `backend/new/handler.go:737-776` — `handleCambiarModulo` actualizado (llama `ObtenerFilasPaginado`)
- `backend/new/app.go` — `ObtenerFilasPaginado` (COUNT + LIMIT/OFFSET)
- `app.go:31-40` — ModuloConfig struct con Columnas, IDColumna, Nombre
- `app.go:42-183` — Definición de los 9 módulos
