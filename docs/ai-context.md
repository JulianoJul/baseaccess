# AI Context — Gestión de Expedientes

## Stack (no negociable)
- SPA 100% cliente-side: HTML + Tailwind CSS + sql.js (SQLite WASM)
- Empaquetado: Tauri v2 (Rust) — AppImage (Linux) / NSIS (Windows)
- Sin backend, sin CDN, sin frameworks JS

## Líneas Rojas
- **Cero hardcodeo**: todo valor variable → constantes con nombre (`CONFIG.*`)
- **SPOT**: schema-config.js es la única fuente de verdad del schema
- **SoC**: separar SQL de UI. Las funciones de renderizado no construyen queries
- **KISS + YAGNI**: resolver solo lo pedido, sin features "por si acaso"
- **Sin efectos secundarios ocultos**: las funciones deben ser predecibles (Least Astonishment)
- **Makefile**: única fuente de automatización local

## Estado Actual (Julio 2026)
App web para gestionar expedientes de contrataciones con historial de movimientos. CRUD completo, observaciones auto-generadas, notas libres, sidebar de frecuentes oculta por defecto (hamburguesa ☰), orden por fecha movido a la barra de búsqueda (select → lupa → input), ruta de procesos, documentos pendientes, schema-config.js centralizado. BD SQLite en archivo .db, persistencia vía Tauri IPC (Rust) + autoguardado. VACUUM: función `optimizarBD()` preservada (sin botón visual) para uso programático. Botón toggle Orden Excel/Secciones sin borde visual. Layout sidebar+contenido con CSS Grid (`grid-cols-[auto_1fr]`), tabla `min-w-full` con `table-layout: auto` (se expande al desplegar fila). X button de modales unificado a `btn-icon`+`fas fa-times`. Sticky solo en barra de búsqueda principal (modales sin sticky). Recientes con `flex flex-col gap-1` sin divisores. Todos los modales se cierran al clickear fuera. Exportación CSV. Integridad de BD al cargar. Orden persistido en localStorage.

## Archivos Clave
| Archivo | Para qué |
|---------|----------|
| `src/index.html` | App completa (HTML + CSS + JS) |
| `src/schema-config.js` | Config del schema (catálogos, columnas, formato obs, estatus) |
| `src/index.html` | App completa (HTML + CSS + JS) |
| `src/schema-config.js` | Config del schema (catálogos, columnas, formato obs, estatus) |
| `src/tauri-preload.js` | Puente JS: window.electronAPI vía Tauri invoke |
| `src-tauri/src/lib.rs` | Backend Rust (comandos IPC: save_db, open_db_dialog, etc.) |
| `data/sql/Tablas8.sql` | Schema SQLite v8 |
| `docs/doc.md` | Documentación + pendientes + changelog |
| `docs/decisiones.md` | ADR: historial de decisiones técnicas |
| `docs/funciones.md` | Catálogo SPOT de todas las funciones (leer antes de crear) |
| `.clinerules` | Skill de Opencode (protocolo de modificación) |
| `Makefile` | combine / clean / commit / push / serve / tauri-build |
| `combined.txt` | Consolidado (make combine) para sesiones |

## Regla de Oro
Antes de tocar código: leer `doc.md` (pendientes) + `decisiones.md` (ADR) + `funciones.md` (catálogo) + `ai-context.md` (esto) + `.clinerules` (skill).
