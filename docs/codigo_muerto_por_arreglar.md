# Código Muerto Detectado

> Código definido pero nunca usado. Se puede eliminar sin afectar la funcionalidad.

## 1. `tabla_action_buttons_alpine` — `components.html:151-168`

Template Go define botones de acción (editar + fijar) para filas de tabla, pero **nunca se invoca**. `tabla.html` tiene sus propios botones inline (líneas 94-106).

```go
{{define "tabla_action_buttons_alpine"}}
...botones editar + fijar...
{{end}}
```

**Acción:** Eliminar el template.

## 2. `form_emisor_receptor_alpine` — `components.html:136-139`

Template Go que agrupa selects de emisor + receptor, pero **nunca se invoca**. `form.html` llama `form_select_alpine` directamente para cada uno.

```go
{{define "form_emisor_receptor_alpine"}}
{{template "form_select_alpine" dict ... "Emisor"...}}
{{template "form_select_alpine" dict ... "Receptor"...}}
{{end}}
```

**Acción:** Eliminar el template.

## 3. `moverProceso()` — `ruta_procesos.html:~651`

Función JS definida para reordenar procesos (arriba/abajo), pero **nunca se llama**. No hay botones ▲▼ en la UI para procesos.

```javascript
window.moverProceso = function(id, direction) { ... }
```

**Acción:** Eliminar la función, o agregar botones de reordenamiento en la UI si se necesita.

## 4. `form_hidden_id_alpine` — conflicto `x-model` + `value` — `components.html:142-144`

Template que usa `x-model` y `value` en el mismo `<input type="hidden">`. Alpine puede sobreescribir el `value` server-rendered.

```html
<input type="hidden" id="f-{{.IDColumna}}" name="{{.IDColumna}}"
       x-model="registro.{{.IDColumna}}"
       value="{{if .Registro}}{{rowGetStr .Registro .IDColumna}}{{end}}">
```

**Acción:** Usar `:value` en lugar de `value`, o eliminar `value` y confiar en `x-model`.

---

*Detectado el 2026-07-22 durante análisis de consistencia Plan_ui.md vs código actual.*
