# Architecture Decision Records (ADR)

Registro cronológico de decisiones técnicas tomadas en el proyecto.

---

## DEC-001: Migración a Wails v2 (Go backend + Web frontend)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** La app original tenía 3 runtimes (Electron/sql.js, Tauri/Rust, navegador). Electron usaba sql.js (WASM) que obligaba a exportar la BD completa en cada guardado. Tauri/Rust era complejo para la IA. Se migró a Wails v2: backend Go nativo con mattn/go-sqlite3, frontend web embebido vía `go:embed`. Un solo runtime, escrituras directas al .db, bindings automáticos Go↔JS.
- **Alternativas evaluadas:**
  - Electron + better-sqlite3 — menos complejo que Wails, pero sigue requiriendo IPC, preload.js, contextBridge
  - Tauri (Rust) — el más complejo para la IA (lifetimes, genéricos, builds lentos)
  - Wails v2 (Go) — elegido: Go simple, bindings automáticos, sin IPC boilerplate
- **Impacto:**
  - Rama `wails-migration` creada (rama paralela, `master` intacto)
  - `main.go`: entry point Wails con `go:embed` para `frontend/`
  - `app.go`: App struct con 12 métodos exportados (AbrirBaseDatos, CRUD, catálogos, VACUUM)
  - `frontend/index.html`: copia de `src/index.html` adaptada — sql.js reemplazado por `window.go.main.App.*`
  - `frontend/schema-config.js`: idéntico al original
  - `go.mod` + `wails.json`: config proyecto Wails
  - `main.js`, `src/`, `src-tauri/`, `package.json`: legacy sin eliminar (master los conserva)
  - `.gitignore`: `build/bin/` añadido para outputs Wails
  - `Makefile`: targets `wails-*`
  - `.github/workflows/build.yml`: job `wails` (Linux + Windows)
  - Windows: WebView2 Fixed Version Runtime 150.0.4078.65 portable via `windows.Options.WebviewBrowserPath`

---

## DEC-002: Límite de 100MB en Drag & Drop

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Archivos SQLite grandes (>100MB) saturan recursos. Se mantuvo el límite del legacy.
- **Impacto:** Validación en `frontend/index.html`.

---

## DEC-003: WebView2 Fixed Version Runtime Portable (Windows)

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** Wails requiere WebView2 en Windows. Para 100% portabilidad (sin instalación, sin internet), se incluye el Fixed Version Runtime junto al .exe.
- **Alternativas evaluadas:**
  - `-webview2 embed` — incluye bootstrapper pero requiere internet para descargar runtime
  - `-webview2 download` — igual, requiere internet
  - Fixed Version Runtime — elegido: 100% offline, portátil en USB
- **Impacto:**
  - `main.go`: `windows.Options.WebviewBrowserPath` apunta al directorio del runtime
  - `build.yml`: download + extract del CAB (Microsoft), cacheado, copiado a `build/bin/`

---

## DEC-004: Escrituras Directas SQLite + Backup Rotativo en Go

- **Origen:** `[Derivado de DEC-001]`
- **Contexto y Causa:** En sql.js, cada guardado exportaba la BD completa (`db.export()`) y sobreescribía el archivo. En Wails, Go escribe directamente al .db vía `database/sql`. Sin embargo, hay cortes de energía frecuentes, por lo que se implementa backup rotativo en Go antes de cada escritura (`.bak.1` a `.bak.N` con N configurable).
- **Impacto:** `app.go`: función `crearBackup()` con rotación, llamada antes de GuardarExpediente, EliminarExpediente, GuardarNuevoCatalogo, OptimizarBD. No se necesita `guardarBD()` ni autosave.

---

## DEC-005: Descargar Copia de BD desde Error Boundary

- **Origen:** `[Instrucción Explícita del Usuario]`
- **Contexto y Causa:** El error boundary tenía `descargarBDError()` como stub. Se implementó funcionalidad real para permitir al usuario guardar una copia del .db actual cuando ocurre un error crítico.
- **Alternativas evaluadas:**
  - Mantener stub — descartado: el usuario quería funcionalidad real
  - Diálogo nativo Wails para guardar archivo — elegido: usa `window.runtime.SaveFileDialog()`
- **Impacto:** `descargarBDError()` ahora abre diálogo de guardado y copia el .db vía Go.

---

## DEC-006: Sin Autoguardado (escrituras inmediatas)

- **Origen:** `[Derivado de DEC-001]`
- **Contexto y Causa:** En sql.js, los cambios estaban solo en RAM hasta que se exportaba la BD. En Wails, cada INSERT/UPDATE/DELETE escribe directamente al archivo .db. El autoguardado cada 30s ya no tiene sentido.
- **Impacto:** `AUTOSAVE_ENABLED` y `AUTOSAVE_INTERVAL_MS` eliminados de `schema-config.js`.
