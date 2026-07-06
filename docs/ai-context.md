# AI Context — Gestión de Expedientes

## Stack (no negociable)
- SPA 100% cliente-side: HTML + Tailwind CSS + sql.js (SQLite WASM)
- Empaquetado: Electron win-unpacked (sin instalación)
- Sin backend, sin CDN, sin frameworks JS

## Líneas Rojas
- **Cero hardcodeo**: todo valor variable → constantes con nombre (`CONFIG.*`)
- **SPOT**: schema-config.js es la única fuente de verdad del schema
- **SoC**: separar SQL de UI. Las funciones de renderizado no construyen queries
- **KISS + YAGNI**: resolver solo lo pedido, sin features "por si acaso"
- **Sin efectos secundarios ocultos**: las funciones deben ser predecibles (Least Astonishment)
- **Makefile**: única fuente de automatización local

## Estado Actual (Julio 2026)
App web para gestionar expedientes de contrataciones con historial de movimientos. CRUD completo, observaciones auto-generadas, notas libres, sidebar de frecuentes, orden por fecha, ruta de procesos, documentos pendientes, schema-config.js centralizado. BD SQLite en archivo .db, persistencia vía Electron IPC + autoguardado.

## Archivos Clave
| Archivo | Para qué |
|---------|----------|
| `index.html` | App completa (HTML + CSS + JS) |
| `schema-config.js` | Config del schema (catálogos, columnas, formato obs, estatus) |
| `main.js` | Electron main process |
| `data/sql/Tablas8.sql` | Schema SQLite v8 |
| `doc.md` | Documentación + pendientes + changelog |
| `decisiones.md` | ADR: historial de decisiones técnicas |
| `funciones.md` | Catálogo SPOT de todas las funciones (leer antes de crear) |
| `.clinerules` | Skill de Opencode (protocolo de modificación) |
| `Makefile` | combine / clean / commit / push / serve / electron-build |
| `combined.txt` | Consolidado (make combine) para sesiones |

## Regla de Oro
Antes de tocar código: leer `doc.md` (pendientes) + `decisiones.md` (ADR) + `funciones.md` (catálogo) + `ai-context.md` (esto) + `.clinerules` (skill).
