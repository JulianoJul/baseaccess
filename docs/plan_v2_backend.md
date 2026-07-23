# Plan v2 Backend — Limpieza, Mejoras y Arquitectura de Datos

> Documento exclusivo para cambios en Go, SQL y API. Sin código HTML/CSS/JS de UI.
> Cada tarea está diseñada para ser implementada de forma autónoma con DeepSeek Flash.

---

## 1. Limpieza de Código Muerto

Código definido pero nunca usado. Eliminar sin afectar funcionalidad.

### 1.1 `tabla_action_buttons_alpine` — `components.html:151-168`

Template Go que define botones de acción (editar + fijar) para filas de tabla, pero **nunca se invoca**. `tabla.html` tiene sus propios botones inline (líneas 94-106).

```go
{{define "tabla_action_buttons_alpine"}}
...botones editar + fijar...
{{end}}
```

**Acción:** Eliminar el template.

### 1.2 `form_emisor_receptor_alpine` — `components.html:136-139`

Template Go que agrupa selects de emisor + receptor, pero **nunca se invoca**. `form.html` llama `form_select_alpine` directamente para cada uno.

**Acción:** Eliminar el template.

### 1.3 `moverProceso()` — `ruta_procesos.html:~651`

Función JS definida para reordenar procesos (arriba/abajo), pero **nunca se llama**. No hay botones en la UI para procesos.

**Acción:** Eliminar la función.

### 1.4 `form_hidden_id_alpine` — conflicto `x-model` + `value`

Template que usa `x-model` y `value` en el mismo `<input type="hidden">`. Alpine puede sobreescribir el `value` server-rendered.

**Acción:** Usar `:value` en lugar de `value`, o eliminar `value` y confiar en `x-model`.

---

## 2. Ruta de Procesos v2 — Referencia de Implementación

> **Estado:** Ya implementado. Esta sección documenta el modelo para mantenimiento futuro.

### 2.1 Esquema SQL (7 tablas)

```sql
-- 1. Hoja (mes)
CREATE TABLE ruta_procesos_hoja (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT NOT NULL
);

-- 2. Junta Directiva
CREATE TABLE ruta_procesos_junta (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    id_hoja     INTEGER NOT NULL REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,
    numero      INTEGER NOT NULL,
    consecutiva INTEGER NOT NULL,
    fecha       TEXT NOT NULL,
    UNIQUE(id_hoja, numero)
);

-- 3. Semanas del Gantt
CREATE TABLE ruta_procesos_junta_semana (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta     INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    numero       INTEGER NOT NULL,
    fecha_inicio TEXT NOT NULL,
    fecha_fin    TEXT NOT NULL,
    UNIQUE(id_junta, numero)
);

-- 4. Procesos de una junta
CREATE TABLE ruta_procesos_junta_proceso (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta  INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    numero    INTEGER NOT NULL,
    proceso   TEXT NOT NULL,
    UNIQUE(id_junta, numero)
);

-- 5. Cronograma diario
CREATE TABLE ruta_procesos_cronograma (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta_proceso INTEGER NOT NULL REFERENCES ruta_procesos_junta_proceso(id) ON DELETE CASCADE,
    fecha            TEXT NOT NULL,
    id_leyenda       INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE RESTRICT,
    nota             TEXT DEFAULT ''
);

-- 6. Leyendas
CREATE TABLE ruta_procesos_leyenda (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre    TEXT NOT NULL,
    color     TEXT NOT NULL DEFAULT '#FFFFFF',
    ambito    TEXT NOT NULL DEFAULT 'junta',
    id_hoja   INTEGER REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,
    bloqueado INTEGER DEFAULT 0
);

-- 7. Relación junta ↔ leyenda
CREATE TABLE ruta_procesos_junta_leyenda (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta   INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    id_leyenda INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE CASCADE,
    orden      INTEGER NOT NULL DEFAULT 0,
    UNIQUE(id_junta, id_leyenda)
);
```

**Leyendas base globales (10 predefinidas):**
```sql
INSERT INTO ruta_procesos_leyenda (id, nombre, color, ambito, id_hoja, bloqueado) VALUES
(1, 'ACTIVIDADES PREVIAS (UNIDAD USUARIA)', '#FF4757', 'global', NULL, 0),
(2, 'INICIO (CONTRATACIÓN)', '#2BCBBA', 'global', NULL, 0),
(3, 'VENTA DE PLIEGO DE CONDICIONES (CONTRATACIÓN)', '#6C5CE7', 'global', NULL, 0),
(4, 'INICIO (COMISIÓN)', '#FF6B81', 'global', NULL, 0),
(5, 'APERTURA DE OFERTAS', '#FFA502', 'global', NULL, 0),
(6, 'ANÁLISIS TÉCNICO', '#2ED573', 'global', NULL, 0),
(7, 'ANÁLISIS ECONÓMICO', '#1E90FF', 'global', NULL, 0),
(8, 'RESULTADOS', '#FDCB6E', 'global', NULL, 0),
(9, 'APROBACIÓN PRESIDENCIA', '#A855F7', 'global', NULL, 0),
(10, 'CONTROL DE DOCUMENTOS PRESIDENCIA', '#00D2D3', 'global', NULL, 0);
```

### 2.2 Lógica de Ámbito de Leyendas

- **`global`** → aparece en TODAS las juntas de TODAS las hojas.
- **`hoja`** → aparece en todas las juntas de ESA hoja.
- **`junta`** → aparece solo en ESA junta.
- Al crear una leyenda, se inserta automáticamente en `ruta_procesos_junta_leyenda` para las juntas correspondientes.
- Al crear una nueva junta, hereda las leyendas `global` y `hoja`.
- `bloqueado=1` impide eliminación.

### 2.3 Datos del Servidor (Structs Go)

```go
type RutaProcesosGanttData struct {
    Hojas        []RutaProcesosHoja        `json:"hojas"`
    CurrentHoja  *RutaProcesosHoja         `json:"current_hoja"`
    Juntas       []RutaProcesosJunta       `json:"juntas"`
    CurrentJunta *RutaProcesosJunta        `json:"current_junta"`
    Semanas      []RutaProcesosJuntaSemana `json:"semanas"`
    Procesos     []RutaProcesosJuntaProceso `json:"procesos"`
    Legend       []RutaProcesosLegend      `json:"legend"`
    JuntaLegend  []RutaProcesosJuntaLeyenda `json:"junta_legend"`
}
```

### 2.4 Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/ruta-procesos` | Cargar Gantt completa |
| POST | `/api/ruta-procesos-hoja-crear` | Crear hoja |
| POST | `/api/ruta-procesos-hoja-eliminar` | Eliminar hoja |
| POST | `/api/ruta-procesos-junta-crear` | Crear junta |
| POST | `/api/ruta-procesos-junta-actualizar` | Actualizar junta |
| POST | `/api/ruta-procesos-junta-eliminar` | Eliminar junta |
| POST | `/api/ruta-procesos-semana-agregar` | Agregar semana |
| POST | `/api/ruta-procesos-semana-eliminar` | Eliminar semanas con reenumeración |
| POST | `/api/ruta-procesos-proceso-agregar` | Agregar proceso |
| POST | `/api/ruta-procesos-proceso-eliminar` | Eliminar proceso |
| POST | `/api/ruta-procesos-cronograma-guardar` | Guardar entrada diaria |
| POST | `/api/ruta-procesos-cronograma-eliminar` | Eliminar entrada |
| POST | `/api/ruta-procesos-leyenda-crear` | Crear leyenda |
| POST | `/api/ruta-procesos-leyenda-actualizar` | Actualizar leyenda |
| POST | `/api/ruta-procesos-leyenda-eliminar` | Eliminar leyenda |
| POST | `/api/ruta-procesos-leyenda-reordenar` | Reordenar leyenda |
| POST | `/api/ruta-procesos-leyenda-bloquear` | Toggle bloquear leyenda |

---

## 3. Feature: Documentos Múltiples

### 3.1 Problema

La tabla `cat_documento` tiene entradas compuestas como "CONTRATO Y ACTA". El campo `id_documento` en cada módulo es un solo FK, limitando a un documento por registro.

### 3.2 Solución Backend

**Paso A: Limpiar `cat_documento`**

Eliminar entradas compuestas, dejar solo tipos únicos:
- CONTRATO, ACTA, OFICIO, SOLPED, etc.

**Paso B: Crear tabla genérica `modulo_documento`**

```sql
CREATE TABLE modulo_documento (
    id_modulo_documento INTEGER PRIMARY KEY AUTOINCREMENT,
    modulo              TEXT NOT NULL,       -- clave del módulo
    id_registro         INTEGER NOT NULL,    -- ID del registro
    id_documento        INTEGER NOT NULL,    -- FK cat_documento
    UNIQUE(modulo, id_registro, id_documento)
);
```

**Paso C: Migrar datos existentes**

Registros que usen entradas compuestas (ej. "CONTRATO Y ACTA") deben convertirse en múltiples filas en `modulo_documento`.

**Paso D: Modificar `GuardarFila` (app.go)**

- Recibir múltiples valores `id_documento=1&id_documento=2`.
- Eliminar/insertar en `modulo_documento` en lugar de escribir `id_documento` en la tabla del módulo.
- Opcional: mantener columna `id_documento` temporalmente para compatibilidad, pero deprecarla.

**Paso E: Modificar lectura**

- `ObtenerFilaPorId`: JOIN a `modulo_documento` + `cat_documento` para devolver array `documentos`.
- `ObtenerFilas`/`ObtenerFilasPaginado`: Incluir documentos (GROUP_CONCAT o array).
- `preparePageData`: El catálogo de documentos ya existe en `Catalogs`.
- Vistas SQL (`vw_reporte_*`): Ajustar JOIN o usar GROUP_CONCAT para columna Documento.

**Paso F: Modificar historial**

El historial de movimientos debe reflejar los documentos asociados en cada snapshot. Dado que `historial_movimientos` almacena JSON o copia completa del registro, asegurar que el array de documentos se incluya.

---

## 4. Otras Mejoras Backend Identificadas

### 4.1 Función `calcDiasSemana()` en Go

Ya implementada. Calcula 5 fechas L-V desde `fecha_inicio`. Usada en el Gantt v2.

### 4.2 `ObtenerRutaProcesosData`

Ya reescrita para v2. Recibe `(idHoja, idJunta)` en lugar de `(idHoja, offsetWeeks)`.

---

## 5. Hoja de Ruta Backend (Step-by-Step para DeepSeek Flash)

Cada paso es una tarea **autónoma, pequeña y verificable**. No combinar pasos.

### Fase A: Limpieza Inmediata (Riesgo bajo)

- **Backend-01:** Eliminar template `tabla_action_buttons_alpine` de `components.html`. Verificar que `tabla.html` sigue renderizando botones editar/fijar correctamente.
- **Backend-02:** Eliminar template `form_emisor_receptor_alpine` de `components.html`. Verificar que `form.html` sigue funcionando.
- **Backend-03:** Eliminar función `moverProceso()` de `ruta_procesos.html`. Verificar que no hay referencias en el resto del template.
- **Backend-04:** Corregir `form_hidden_id_alpine`: cambiar `value="..."` por `:value="..."` o remover `value` y dejar solo `x-model`.

### Fase B: Documentos Múltiples (Riesgo medio)

- **Backend-05:** Crear tabla `modulo_documento` en schema SQL. Ejecutar en BD existente.
- **Backend-06:** Script de migración: leer registros con `id_documento` que apunte a entradas compuestas de `cat_documento`, crear múltiples filas en `modulo_documento`. Verificar conteo antes/después.
- **Backend-07:** Limpiar `cat_documento`: eliminar filas compuestas (ej. "CONTRATO Y ACTA", "OFICIO Y CONTRATO"). Verificar que no queden FK huérfanas.
- **Backend-08:** Modificar `GuardarFila` en `app.go`: leer múltiples `id_documento` del form POST, insertar en `modulo_documento`. Ignorar (o deprecar) columna `id_documento` de la tabla del módulo.
- **Backend-09:** Modificar `ObtenerFilaPorId`: JOIN con `modulo_documento` + `cat_documento` para poblar `registro["documentos"]` como array de `{id, nombre}`. Verificar con `go test`.
- **Backend-10:** Modificar `ObtenerFilasPaginado`: incluir documentos. Decidir si GROUP_CONCAT en SQL o join en Go.
- **Backend-11:** Modificar `HistorialTabla` o las funciones de snapshot para incluir documentos del registro en cada entrada de historial.
- **Backend-12:** Crear endpoint `/api/columnas-modulo` si no existe, para alimentar el multi-select de documentos en el frontend (o reutilizar `Catalogs` que ya trae `cat_documento`).

### Fase C: Testing y Verificación

- **Backend-13:** Ejecutar `go test ./...` después de cada paso de la Fase B.
- **Backend-14:** Regenerar base de datos de prueba con `importar_datos.py` y verificar que módulos cargan sin error.
- **Backend-15:** Revisar que `wails build` compila sin errores tras todos los cambios.

### Nota sobre Ruta Procesos v2

Esta funcionalidad ya está implementada y en producción. Los pasos Backend-05 a Backend-15 son las próximas tareas pendientes.
