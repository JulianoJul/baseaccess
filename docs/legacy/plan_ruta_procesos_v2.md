# Plan Ruta Procesos v2

> La BD de Ruta Procesos se desliga de la BD principal. Tiene sus propios procesos, sin referencias a módulos.

---

## 1. Flujo de UI

Todas las juntas de la misma hoja se muestran **una debajo de otra** (scroll vertical). Cada junta repite el mismo bloque: tabla → Gantt → leyendas.

```
┌──────────────────────────────────────────────────────┐
│  [Select Hoja ▼]  [+ Nueva Hoja] [🗑]               │  ← Barra superior
├──────────────────────────────────────────────────────┤
│                                                      │
│  ┌─ JUNTA #1 ────────────────────────────────────┐   │
│  │  TABLA (1 fila, editable)                    │   │
│  │  ┌──────────────┬──────┬────────┬──────────┐  │   │
│  │  │ JUNTA DIRECTIVA │ Nº │ CONSEC │ FECHA    │  │   │
│  │  └──────────────┴──────┴────────┴──────────┘  │   │
│  │                                                │   │
│  │  GANTT                                         │   │
│  │  ┌────┬─────────┬──────────────┬──────────┬──┐ │   │
│  │  │    │         │desde 01/06   │desde 08/06│  │ │   │  ← fila 1: desde
│  │  │    │         │al 05/06      │al 12/06   │  │ │   │  ← fila 2: al
│  │  │    │         │  SEMANA 1    │ SEMANA 2 │  │ │   │  ← fila 3: semanas + botones
│  │  │    │         ├──────┬───────┼──────┬────┤  │ │   │
│  │  │ N° │ Proceso │L M X J V│...│    │  │ │   │  ← fila 4: días
│  │  ├────┼─────────┼──────┼───────┼──────┼────┤  │ │   │
│  │  │ 1  │ Algo    │🔵🔴 │  🟢🟡  │      │    │  │ │   │  ← fila 5+: procesos
│  │  │ 2  │ Otro    │🟢   │  🔵🔴🟡 │      │    │  │ │   │
│  │  ├────┼─────────┴──────┴───────┴──────┴────┤  │ │   │
│  │  │    │ [+] Añadir proceso                 │  │ │   │  ← fila extra al fondo
│  │  ├────┼─────────┼──────┼───────┼──────┼────┤  │ │   │
│  │  │ 1  │ Algo    │🔵🔴 │  🟢🟡  │      │    │  │ │   │
│  │  │ 2  │ Otro    │🟢   │  🔵🔴🟡 │      │    │  │ │   │
│  │  └────┴─────────┴──────┴───────┴──────┴────┘  │ │   │
│  │                                                │   │
│  │  LEYENDAS de Junta #1                          │   │
│  │  🟢 Aprobado ▲▼  🔒  🟡 Espera ▲▼  🔒        │   │
│  │  [+ Añadir leyenda]  (ámbito: junta/hoja/global)│  │
│  └────────────────────────────────────────────────┘   │
│                                                      │
│  ┌─ JUNTA #2 ────────────────────────────────────┐   │
│  │  (misma estructura: tabla + Gantt + leyendas)  │   │
│  └────────────────────────────────────────────────┘   │
│                                                      │
│  ┌─ JUNTA #3 ────────────────────────────────────┐   │
│  │  ...                                           │   │
│  └────────────────────────────────────────────────┘   │
│                                                      │
│              [+ NUEVA JUNTA]                          │  ← Botón grande al fondo
└──────────────────────────────────────────────────────┘
```

- Cada junta es un bloque **colapsable** o **siempre visible**.
- Las leyendas se crean por junta con selector de ámbito: **Esta junta** / **Esta hoja** / **Global**.

---

## 2. Tablas SQL

### 2.1 `ruta_procesos_hoja`
```sql
CREATE TABLE ruta_procesos_hoja (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT    NOT NULL       -- ej: "Junio 2026"
);
```

### 2.2 `ruta_procesos_junta`
```sql
CREATE TABLE ruta_procesos_junta (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    id_hoja     INTEGER NOT NULL REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,
    numero      INTEGER NOT NULL,      -- nro reunión (NO autoincrement)
    consecutiva INTEGER NOT NULL,      -- consecutivo
    fecha       TEXT    NOT NULL,      -- fecha única YYYY-MM-DD
    UNIQUE(id_hoja, numero)
);
```

### 2.3 `ruta_procesos_junta_semana`
**Nueva.** Semanas dinámicas dentro del Gantt de cada junta.
```sql
CREATE TABLE ruta_procesos_junta_semana (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta     INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    numero       INTEGER NOT NULL,      -- 1, 2, 3... (se reenumera al borrar)
    fecha_inicio TEXT    NOT NULL,      -- YYYY-MM-DD (lunes)
    fecha_fin    TEXT    NOT NULL,      -- YYYY-MM-DD (viernes)
    UNIQUE(id_junta, numero)
);
```

### 2.4 `ruta_procesos_junta_proceso`
```sql
CREATE TABLE ruta_procesos_junta_proceso (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta  INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    numero    INTEGER NOT NULL,         -- nro secuencial (NO autoincrement)
    proceso   TEXT    NOT NULL,         -- nombre del proceso
    UNIQUE(id_junta, numero)
);
```

### 2.5 `ruta_procesos_cronograma`
```sql
CREATE TABLE ruta_procesos_cronograma (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta_proceso INTEGER NOT NULL REFERENCES ruta_procesos_junta_proceso(id) ON DELETE CASCADE,
    fecha            TEXT    NOT NULL,  -- YYYY-MM-DD
    id_leyenda       INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE RESTRICT,
    nota             TEXT    DEFAULT ''
);
```

### 2.6 `ruta_procesos_leyenda`
```sql
CREATE TABLE ruta_procesos_leyenda (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre    TEXT    NOT NULL,
    color     TEXT    NOT NULL DEFAULT '#FFFFFF',
    ambito    TEXT    NOT NULL DEFAULT 'junta',  -- 'junta' | 'hoja' | 'global'
    id_hoja   INTEGER REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,  -- solo si ambito='hoja'
    bloqueado INTEGER DEFAULT 0   -- 1 = bloqueado, no se puede eliminar
);
```

### 2.7 `ruta_procesos_junta_leyenda`
```sql
CREATE TABLE ruta_procesos_junta_leyenda (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta   INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    id_leyenda INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE CASCADE,
    orden      INTEGER NOT NULL DEFAULT 0,
    UNIQUE(id_junta, id_leyenda)
);
```

**Lógica de ámbito:**
- **`junta`** → solo para la junta actual. Se inserta solo en esa junta.
- **`hoja`** → para todas las juntas de esta hoja. Al crearla, se inserta en todas las juntas existentes de la misma hoja. Al crear una nueva junta en esta hoja, hereda todas las leyendas con `ambito='hoja'` de esa hoja.
- **`global`** → para todas las juntas de TODAS las hojas. Al crearla, se inserta en todas las juntas existentes (de todas las hojas). Al crear cualquier nueva junta en cualquier hoja, hereda las globales.

**Eliminar leyenda:**
- Eliminar una leyenda → CASCADE borra sus filas en `ruta_procesos_junta_leyenda` de todas las juntas donde exista.

---

## 3. Gantt — Detalle por semana

Cada semana tiene su propio rango de fechas ("desde" y "al") visibles arriba de la fila L M X J V.

```
     desde 01/06/2026      desde 08/06/2026      ← fila 1
     al 05/06/2026         al 12/06/2026         ← fila 2
          SEMANA 1              SEMANA 2      [+] [🗑]  ← fila 3
     ┌──────┬───────┐      ┌──────┬───────┐
     │L M X J V│      │      │L M X J V│      │      ← fila 4
     ├──────┼───────┤      ├──────┼───────┤
     │🔵🔴   │  🟢🟡 │      │🟢    │ 🔵🔴🟡 │
     └──────┴───────┘      └──────┴───────┘
```

- **Botón `[+]`**: diálogo modal con las fechas de la nueva semana (lunes→viernes), precalculadas consecutivas a la última semana. El usuario puede ajustarlas. Al confirmar, se inserta en `ruta_procesos_junta_semana` con el siguiente `numero`.
- **Botón `[🗑]`**: diálogo con checkboxes de semanas a eliminar. Al borrar, las semanas restantes se **renumeran** (si borro semana 1, semana 2 pasa a ser 1, semana 3 pasa a ser 2...).

---

## 4. Tabla de datos de la Junta (1 fila)

Campos editables inline (reutilizando los componentes `form_input_*_alpine`):

| Campo | Tipo |
|-------|------|
| JUNTA DIRECTIVA | texto fijo (label) |
| Nº Reunión | input number |
| Consecutiva | input number |
| Fecha | date picker |

Un botón de **Guardar** al lado. Estos campos se persisten en `ruta_procesos_junta`.

---

## 5. Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/ruta-procesos` | Cargar hoja + juntas + junta activa + Gantt + leyendas |
| POST | `/api/ruta-procesos-hoja-crear` | Crear hoja |
| POST | `/api/ruta-procesos-hoja-eliminar` | Eliminar hoja |
| POST | `/api/ruta-procesos-junta-crear` | Crear junta |
| POST | `/api/ruta-procesos-junta-actualizar` | Actualizar junta |
| POST | `/api/ruta-procesos-junta-eliminar` | Eliminar junta |
| POST | `/api/ruta-procesos-leyenda-bloquear` | Toggle bloquear leyenda |
| POST | `/api/ruta-procesos-semana-agregar` | Agregar semana al Gantt |
| POST | `/api/ruta-procesos-semana-eliminar` | Eliminar semanas (con reenumeración) |
| POST | `/api/ruta-procesos-proceso-agregar` | Agregar proceso a junta |
| POST | `/api/ruta-procesos-proceso-eliminar` | Eliminar proceso |
| POST | `/api/ruta-procesos-proceso-reordenar` | Reordenar procesos |
| POST | `/api/ruta-procesos-cronograma-guardar` | Guardar entrada de cronograma |
| POST | `/api/ruta-procesos-cronograma-eliminar` | Eliminar entrada |
| POST | `/api/ruta-procesos-leyenda-crear` | Crear leyenda |
| POST | `/api/ruta-procesos-leyenda-actualizar` | Actualizar leyenda |
| POST | `/api/ruta-procesos-leyenda-eliminar` | Eliminar leyenda |
| POST | `/api/ruta-procesos-leyenda-reordenar` | Reordenar leyendas de una junta |

---

## 6. Orden de implementación

1. **SQL** — Respaldar `data/sql/03_ruta_procesos.sql` a `data/sql/03_ruta_procesos.sql.legacy`. Crear nuevo `data/sql/03_ruta_procesos.sql` con las 7 tablas.
2. **app.go** — Structs y funciones DB. Remover código viejo referente a módulos en ruta_procesos.
3. **handler.go** — Nuevos endpoints. Remover handlers viejos no utilizados.
4. **ruta_procesos.html** — Reconstruir UI completa.
5. **Documentación** — Actualizar `docs/mapeo_excel_bd.md`, `docs/doc.md`, `docs/funciones.md` y `prompt.md` con los nuevos cambios.
6. **Build** — Actualizar `.gitignore` y `Makefile` si es necesario (nuevos binarios, assets).
7. **Verificar SQL principales** — Revisar `01_master_control_docs_presidencia.sql` y `02_modulos_adicionales.sql` por si requieren cambios (referencias a ruta_procesos, claves foráneas, etc.).

---

## 7. Confirmado

- Navegación entre juntas: **scroll vertical**, cada junta es un bloque con tabla + Gantt + leyendas.
- Ámbito de leyendas: **junta** (solo esa junta) / **hoja** (todas las juntas de la hoja) / **global** (todas las juntas de todas las hojas).
