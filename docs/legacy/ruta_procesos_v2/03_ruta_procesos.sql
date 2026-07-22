-- ============================================================
-- Ruta de Procesos v2 — Schema desligado de la BD principal
-- ============================================================

-- 1. Hoja (mes)
CREATE TABLE IF NOT EXISTS ruta_procesos_hoja (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre TEXT    NOT NULL
);

-- 2. Junta Directiva (pertenece a una hoja)
CREATE TABLE IF NOT EXISTS ruta_procesos_junta (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    id_hoja     INTEGER NOT NULL REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,
    numero      INTEGER NOT NULL,
    consecutiva INTEGER NOT NULL,
    fecha       TEXT    NOT NULL,
    UNIQUE(id_hoja, numero)
);

-- 3. Semanas del Gantt (dinámicas, por junta)
CREATE TABLE IF NOT EXISTS ruta_procesos_junta_semana (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta     INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    numero       INTEGER NOT NULL,
    fecha_inicio TEXT    NOT NULL,
    fecha_fin    TEXT    NOT NULL,
    UNIQUE(id_junta, numero)
);

-- 4. Procesos de una junta (filas del Gantt)
CREATE TABLE IF NOT EXISTS ruta_procesos_junta_proceso (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta  INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    numero    INTEGER NOT NULL,
    proceso   TEXT    NOT NULL,
    UNIQUE(id_junta, numero)
);

-- 5. Cronograma diario de cada proceso
CREATE TABLE IF NOT EXISTS ruta_procesos_cronograma (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta_proceso INTEGER NOT NULL REFERENCES ruta_procesos_junta_proceso(id) ON DELETE CASCADE,
    fecha            TEXT    NOT NULL,
    id_leyenda       INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE RESTRICT,
    nota             TEXT    DEFAULT ''
);

-- 6. Leyendas (ámbito: junta | hoja | global)
CREATE TABLE IF NOT EXISTS ruta_procesos_leyenda (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre    TEXT    NOT NULL,
    color     TEXT    NOT NULL DEFAULT '#FFFFFF',
    ambito    TEXT    NOT NULL DEFAULT 'junta',
    id_hoja   INTEGER REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,
    bloqueado INTEGER DEFAULT 0
);

-- 7. Relación junta ↔ leyenda (con orden personalizable)
CREATE TABLE IF NOT EXISTS ruta_procesos_junta_leyenda (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_junta   INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
    id_leyenda INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE CASCADE,
    orden      INTEGER NOT NULL DEFAULT 0,
    UNIQUE(id_junta, id_leyenda)
);

-- ============================================================
-- Leyendas base globales (heredadas del schema anterior)
-- ============================================================
INSERT OR IGNORE INTO ruta_procesos_leyenda (id, nombre, color, ambito, id_hoja, bloqueado) VALUES
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
