-- Schema Ruta Procesos (Gantt)
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS ruta_procesos_leyenda (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    status_name TEXT NOT NULL UNIQUE,
    hex_color   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ruta_procesos_hojas (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    nombre       TEXT NOT NULL,
    fecha_inicio DATE NOT NULL,
    fecha_fin    DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS ruta_procesos_procesos (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    id_hoja     INTEGER NOT NULL,
    descripcion TEXT NOT NULL,
    db_id       INTEGER,
    activo      INTEGER DEFAULT 1,
    CONSTRAINT fk_proc_exp FOREIGN KEY (db_id) REFERENCES expedientes(id_expediente),
    CONSTRAINT fk_proc_hoja FOREIGN KEY (id_hoja) REFERENCES ruta_procesos_hojas(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS ruta_procesos_cronograma (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    id_proceso    INTEGER NOT NULL,
    id_expediente INTEGER,
    fecha         DATE NOT NULL,
    id_leyenda    INTEGER,
    nota          TEXT,
    CONSTRAINT fk_cron_proc FOREIGN KEY (id_proceso) REFERENCES ruta_procesos_procesos(id) ON DELETE CASCADE,
    CONSTRAINT fk_cron_ley FOREIGN KEY (id_leyenda) REFERENCES ruta_procesos_leyenda(id),
    CONSTRAINT fk_cron_exp FOREIGN KEY (id_expediente) REFERENCES expedientes(id_expediente),
    CONSTRAINT unq_cron_day UNIQUE (id_proceso, fecha)
);

-- Leyenda base
INSERT OR IGNORE INTO ruta_procesos_leyenda (status_name, hex_color) VALUES
('ACTIVIDADES PREVIAS (UNIDAD USUARIA)', '#FFA500'),
('INICIO (CONTRATACIÓN)', '#3B82F6'),
('VENTA DE PLIEGO DE CONDICIONES (CONTRATACIÓN)', '#8B5CF6'),
('INICIO (COMISIÓN)', '#10B981'),
('APERTURA DE OFERTAS', '#F59E0B'),
('ANÁLISIS TÉCNICO', '#06B6D4'),
('ANÁLISIS ECONÓMICO', '#14B8A6'),
('RESULTADOS', '#6366F1'),
('APROBACIÓN PRESIDENCIA', '#EC4899'),
('CONTROL DE DOCUMENTOS PRESIDENCIA', '#6B7280');
