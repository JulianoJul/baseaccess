# Plan: Modo Dual — Wails Desktop + HTTP Server

## Objetivo

Que la app Go funcione en dos modos sin duplicar lógica:
1. **Wails desktop** (como ahora) — `wails build`
2. **HTTP server standalone** — accesible desde cualquier navegador (`go build -o server .`)

## Hallazgo clave

El proyecto ya está ~95% listo. `handler.go` es un `http.Handler` estándar que no depende de Wails. La app usa HTMX + fetch para comunicarse con el backend vía `/api/*`. Solo **2 métodos** (`AbrirDialogoBD`, `GuardarDialogoBD`) dependen del runtime de Wails.

---

## Paso 1: Inicializar `ctx` en `NewApp()` (`app.go`)

- Cambiar `NewApp()` para que inicialice `ctx` con `context.Background()`
- En modo Wails, `OnStartup` lo sobrescribirá. En modo HTTP, `context.Background()` es suficiente.
- **No hace falta** modificar `Startup()`.

Verificar que `app.go` ya importa `context`. Si no, agregarlo.

---

## Paso 2: Mover métodos Wails a archivo con build tag

### 2a. Crear `app_wails.go` (nuevo)

Mover `AbrirDialogoBD()` y `GuardarDialogoBD()` desde `app.go` a este archivo.
Agregar build tag `//go:build wails`.
Importar `github.com/wailsapp/wails/v2/pkg/runtime`.

### 2b. Eliminar de `app.go`

Borrar los dos métodos y la importación del runtime de Wails de `app.go`.
Si el import de `wailsRuntime` ya no se usa en `app.go`, eliminarlo.

---

## Paso 3: Renombrar `main.go` → `main_wails.go`

- Agregar build tag `//go:build wails` al inicio
- El contenido del archivo no cambia
- Sigue usando `wails.Run()` como hasta ahora

---

## Paso 4: Crear `main_http.go` (nuevo)

Archivo sin build tag (o `//go:build !wails`). Este es el default cuando no se especifica `-tags wails`.

Contenido:
```go
package main

import (
    "flag"
    "log"
    "net/http"
)

func main() {
    addr := flag.String("addr", ":8080", "HTTP listen address")
    flag.Parse()

    app := NewApp()
    handler, err := NewTemplateHandler(app)
    if err != nil {
        log.Fatalf("error creando template handler: %v", err)
    }

    log.Printf("BaseAccess HTTP server on http://localhost%s", *addr)
    log.Fatal(http.ListenAndServe(*addr, handler))
}
```

---

## Paso 5: Verificar fallback del diálogo de archivos

El frontend (`index.html`) ya tiene fallback para cuando Wails no está disponible:
- Intenta `window.go.main.App.AbrirDialogoBD()` (Wails)
- Si falla, usa `<input type="file">` (HTML nativo)
- Luego hace `fetch('/api/abrir-bd')` para abrir la BD

Esto funciona sin cambios. Verificar que en modo HTTP el input file aparece correctamente.

---

## Paso 6: Build y prueba

### Modo HTTP
```bash
go build -o server .
./server -addr :8080
# Abrir http://localhost:8080 en el navegador
```

### Modo Wails
```bash
wails build
# Build normal de escritorio
```

---

## Resumen de archivos

| Archivo | Cambio |
|---------|--------|
| `app.go` | Init `ctx` con `context.Background()`, eliminar 2 métodos Wails |
| `app_wails.go` | **NUEVO** — `AbrirDialogoBD` + `GuardarDialogoBD` con build tag `wails` |
| `main.go` | **RENOMBRAR** a `main_wails.go` — agregar build tag `wails` |
| `main_http.go` | **NUEVO** — entry point HTTP server |
| `handler.go` | **Sin cambios** |
| `frontend/` | **Sin cambios** |
