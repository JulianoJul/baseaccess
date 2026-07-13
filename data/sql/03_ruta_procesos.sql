-- Schema Ruta Procesos (Gantt)
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS ruta_procesos_leyenda (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    status_name TEXT NOT NULL UNIQUE,
    hex_color   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ruta_procesos_cronograma (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    id_proceso    INTEGER NOT NULL,
    id_expediente INTEGER,
    fecha         DATE NOT NULL,
    id_leyenda    INTEGER,
    nota          TEXT,
    CONSTRAINT fk_cron_ley FOREIGN KEY (id_leyenda) REFERENCES ruta_procesos_leyenda(id),
    CONSTRAINT fk_cron_exp FOREIGN KEY (id_expediente) REFERENCES expedientes(id_expediente)
);

CREATE TABLE IF NOT EXISTS ruta_procesos_procesos (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    descripcion TEXT NOT NULL,
    db_id       INTEGER,
    activo      INTEGER DEFAULT 1,
    CONSTRAINT fk_proc_exp FOREIGN KEY (db_id) REFERENCES expedientes(id_expediente)
);

-- Leyenda base
INSERT OR IGNORE INTO ruta_procesos_leyenda (status_name, hex_color) VALUES
('PENDIENTE', '#FFA500'),
('EN REVISION', '#3B82F6'),
('FIRMADO', '#10B981'),
('DEVUELTO', '#EF4444'),
('SIN NOVEDAD', '#6B7280'),
('OTRO', '#000000');
