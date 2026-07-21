-- Schema Ruta Procesos (Gantt)
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS ruta_procesos_leyenda (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    status_name TEXT NOT NULL UNIQUE,
    hex_color   TEXT NOT NULL,
    orden       INTEGER NOT NULL DEFAULT 0
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
    modulo      TEXT NOT NULL DEFAULT 'expedientes',
    descripcion TEXT NOT NULL,
    db_id       INTEGER,
    activo      INTEGER DEFAULT 1,
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

-- Leyenda base (orden personalizado)
INSERT OR IGNORE INTO ruta_procesos_leyenda (status_name, hex_color, orden) VALUES
('ACTIVIDADES PREVIAS (UNIDAD USUARIA)', '#FF4757', 1),
('INICIO (CONTRATACIÓN)', '#2BCBBA', 2),
('VENTA DE PLIEGO DE CONDICIONES (CONTRATACIÓN)', '#6C5CE7', 3),
('INICIO (COMISIÓN)', '#FF6B81', 4),
('APERTURA DE OFERTAS', '#FFA502', 5),
('ANÁLISIS TÉCNICO', '#2ED573', 6),
('ANÁLISIS ECONÓMICO', '#1E90FF', 7),
('RESULTADOS', '#FDCB6E', 8),
('APROBACIÓN PRESIDENCIA', '#A855F7', 9),
('CONTROL DE DOCUMENTOS PRESIDENCIA', '#00D2D3', 10);
