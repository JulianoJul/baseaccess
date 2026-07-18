PRAGMA foreign_keys = ON;

-- ==========================================
-- 📦 MÓDULOS INDEPENDIENTES (Hojas del Excel)
-- ==========================================

-- ==========================================
-- 🔹 MÓDULO 1: REQUISICIÓN DE MATERIALES
-- ==========================================
CREATE TABLE IF NOT EXISTS req_materiales (
    id_requisicion        INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    id_documento          INTEGER,
    descripcion_materiales TEXT,
    serial_equipo         TEXT,
    pase_sicesma          TEXT,
    id_estatus            INTEGER DEFAULT 1,
    observaciones_entrega TEXT,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_req_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_req_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_req_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_req_re  FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_req_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    CONSTRAINT fk_req_doc FOREIGN KEY (id_documento) REFERENCES cat_documento(id)
);

CREATE TABLE IF NOT EXISTS hist_req_materiales (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_requisicion        INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    id_documento          INTEGER,
    descripcion_materiales TEXT,
    serial_equipo         TEXT,
    pase_sicesma          TEXT,
    id_estatus            INTEGER,
    observaciones_entrega TEXT,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_requisicion) REFERENCES req_materiales(id_requisicion),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    FOREIGN KEY (id_documento) REFERENCES cat_documento(id)
);

CREATE TRIGGER IF NOT EXISTS trg_req_mat_inicial AFTER INSERT ON req_materiales
FOR EACH ROW BEGIN
    INSERT INTO hist_req_materiales (id_requisicion, id_gerencia, id_superintendencia, id_emisor, id_documento, descripcion_materiales, serial_equipo, pase_sicesma, id_estatus, observaciones_entrega, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_requisicion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.descripcion_materiales, NEW.serial_equipo, NEW.pase_sicesma, NEW.id_estatus, NEW.observaciones_entrega, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_req_mat_auditoria AFTER UPDATE ON req_materiales
FOR EACH ROW BEGIN
    INSERT INTO hist_req_materiales (id_requisicion, id_gerencia, id_superintendencia, id_emisor, id_documento, descripcion_materiales, serial_equipo, pase_sicesma, id_estatus, observaciones_entrega, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_requisicion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.descripcion_materiales, NEW.serial_equipo, NEW.pase_sicesma, NEW.id_estatus, NEW.observaciones_entrega, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
    UPDATE req_materiales SET fecha_actualizacion = CURRENT_DATE WHERE id_requisicion = NEW.id_requisicion;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_req_materiales AS
SELECT
    r.id_requisicion,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    d.nombre AS documento,
    r.descripcion_materiales,
    r.serial_equipo,
    r.pase_sicesma,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    r.observaciones_entrega,
    r.fecha_recibido,
    r.fecha_devuelto,
    COALESCE(re.nombre, 'NO APLICA') AS receptor,
    r.observaciones,
    r.notas,
    r.fecha_creacion,
    r.fecha_actualizacion
FROM req_materiales r
LEFT JOIN cat_gerencia g ON r.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON r.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON r.id_emisor = em.id
LEFT JOIN cat_responsables re ON r.id_receptor = re.id
LEFT JOIN cat_estatus_detalle ed ON r.id_estatus = ed.id
LEFT JOIN cat_documento d ON r.id_documento = d.id;

CREATE INDEX IF NOT EXISTS idx_req_mat_estatus ON req_materiales(id_estatus);
CREATE INDEX IF NOT EXISTS idx_hist_req_mat_id ON hist_req_materiales(id_requisicion);


-- ==========================================
-- 🔹 MÓDULO 2: MEMORÁNDUM / DECISIÓN DE GERENCIA
-- ==========================================
CREATE TABLE IF NOT EXISTS memorandums (
    id_memorandum         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    asunto                TEXT,
    id_estatus            INTEGER DEFAULT 1,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_mem_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_mem_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_mem_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_mem_re  FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_mem_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TABLE IF NOT EXISTS hist_memorandums (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_memorandum         INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    asunto                TEXT,
    id_estatus            INTEGER,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_memorandum) REFERENCES memorandums(id_memorandum),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TRIGGER IF NOT EXISTS trg_mem_inicial AFTER INSERT ON memorandums
FOR EACH ROW BEGIN
    INSERT INTO hist_memorandums (id_memorandum, id_gerencia, id_superintendencia, id_emisor, documento, asunto, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_memorandum, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.asunto, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_mem_auditoria AFTER UPDATE ON memorandums
FOR EACH ROW BEGIN
    INSERT INTO hist_memorandums (id_memorandum, id_gerencia, id_superintendencia, id_emisor, documento, asunto, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_memorandum, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.asunto, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
    UPDATE memorandums SET fecha_actualizacion = CURRENT_DATE WHERE id_memorandum = NEW.id_memorandum;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_memorandums AS
SELECT
    m.id_memorandum,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    m.documento,
    m.asunto,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    m.fecha_recibido,
    m.fecha_devuelto,
    COALESCE(re.nombre, 'NO APLICA') AS receptor,
    m.observaciones,
    m.notas,
    m.fecha_creacion,
    m.fecha_actualizacion
FROM memorandums m
LEFT JOIN cat_gerencia g ON m.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON m.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON m.id_emisor = em.id
LEFT JOIN cat_responsables re ON m.id_receptor = re.id
LEFT JOIN cat_estatus_detalle ed ON m.id_estatus = ed.id;

CREATE INDEX IF NOT EXISTS idx_mem_estatus ON memorandums(id_estatus);
CREATE INDEX IF NOT EXISTS idx_hist_mem_id ON hist_memorandums(id_memorandum);


-- ==========================================
-- 🔹 MÓDULO 3: RECOBROS
-- ==========================================
CREATE TABLE IF NOT EXISTS recobros (
    id_recobro            INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    asunto                TEXT,
    fecha_inicio          DATE,
    fecha_final           DATE,
    servicios             TEXT,
    beneficios            TEXT,
    nota_debito_reverso   REAL,
    costo_servicio_usd    REAL,
    id_estatus            INTEGER DEFAULT 1,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_rec_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_rec_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_rec_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_rec_re  FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_rec_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TABLE IF NOT EXISTS hist_recobros (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_recobro            INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    asunto                TEXT,
    fecha_inicio          DATE,
    fecha_final           DATE,
    servicios             TEXT,
    beneficios            TEXT,
    nota_debito_reverso   REAL,
    costo_servicio_usd    REAL,
    id_estatus            INTEGER,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_recobro) REFERENCES recobros(id_recobro),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TRIGGER IF NOT EXISTS trg_rec_inicial AFTER INSERT ON recobros
FOR EACH ROW BEGIN
    INSERT INTO hist_recobros (id_recobro, id_gerencia, id_superintendencia, id_emisor, documento, asunto, fecha_inicio, fecha_final, servicios, beneficios, nota_debito_reverso, costo_servicio_usd, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_recobro, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.asunto, NEW.fecha_inicio, NEW.fecha_final, NEW.servicios, NEW.beneficios, NEW.nota_debito_reverso, NEW.costo_servicio_usd, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_rec_auditoria AFTER UPDATE ON recobros
FOR EACH ROW BEGIN
    INSERT INTO hist_recobros (id_recobro, id_gerencia, id_superintendencia, id_emisor, documento, asunto, fecha_inicio, fecha_final, servicios, beneficios, nota_debito_reverso, costo_servicio_usd, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_recobro, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.asunto, NEW.fecha_inicio, NEW.fecha_final, NEW.servicios, NEW.beneficios, NEW.nota_debito_reverso, NEW.costo_servicio_usd, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
    UPDATE recobros SET fecha_actualizacion = CURRENT_DATE WHERE id_recobro = NEW.id_recobro;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_recobros AS
SELECT
    r.id_recobro,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    r.documento,
    r.asunto,
    r.fecha_inicio,
    r.fecha_final,
    r.servicios,
    r.beneficios,
    r.nota_debito_reverso,
    r.costo_servicio_usd,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    r.fecha_recibido,
    r.fecha_devuelto,
    COALESCE(re.nombre, 'NO APLICA') AS receptor,
    r.observaciones,
    r.notas,
    r.fecha_creacion,
    r.fecha_actualizacion
FROM recobros r
LEFT JOIN cat_gerencia g ON r.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON r.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON r.id_emisor = em.id
LEFT JOIN cat_responsables re ON r.id_receptor = re.id
LEFT JOIN cat_estatus_detalle ed ON r.id_estatus = ed.id;

CREATE INDEX IF NOT EXISTS idx_rec_estatus ON recobros(id_estatus);
CREATE INDEX IF NOT EXISTS idx_hist_rec_id ON hist_recobros(id_recobro);


-- ==========================================
-- 🔹 MÓDULO 4: VALUACIONES
-- ==========================================
CREATE TABLE IF NOT EXISTS valuaciones (
    id_valuacion          INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    solped                TEXT,
    presupuesto_base_bs   REAL,
    presupuesto_base_usd  REAL,
    descripcion_proceso   TEXT,
    id_estatus            INTEGER DEFAULT 1,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    nro_proceso           TEXT,
    nro_contrato_sicac    TEXT,
    nro_contrato_sap      TEXT,
    id_empresa            INTEGER,
    tiempo_ejecucion      TEXT,
    monto_adjudicado_bs   REAL,
    monto_adjudicado_usd  REAL,
    periodo_valuacion_desde DATE,
    periodo_valuacion_hasta DATE,
    monto_valuacion       REAL,
    nro_proforma          TEXT,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_val_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_val_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_val_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_val_re  FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_val_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    CONSTRAINT fk_val_emp FOREIGN KEY (id_empresa) REFERENCES cat_empresas(id)
);

CREATE TABLE IF NOT EXISTS hist_valuaciones (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_valuacion          INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    solped                TEXT,
    presupuesto_base_bs   REAL,
    presupuesto_base_usd  REAL,
    descripcion_proceso   TEXT,
    id_estatus            INTEGER,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    nro_proceso           TEXT,
    nro_contrato_sicac    TEXT,
    nro_contrato_sap      TEXT,
    id_empresa            INTEGER,
    tiempo_ejecucion      TEXT,
    monto_adjudicado_bs   REAL,
    monto_adjudicado_usd  REAL,
    periodo_valuacion_desde DATE,
    periodo_valuacion_hasta DATE,
    monto_valuacion       REAL,
    nro_proforma          TEXT,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_valuacion) REFERENCES valuaciones(id_valuacion),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    FOREIGN KEY (id_empresa) REFERENCES cat_empresas(id)
);

CREATE TRIGGER IF NOT EXISTS trg_val_inicial AFTER INSERT ON valuaciones
FOR EACH ROW BEGIN
    INSERT INTO hist_valuaciones (id_valuacion, id_gerencia, id_superintendencia, id_emisor, documento, solped, presupuesto_base_bs, presupuesto_base_usd, descripcion_proceso, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, nro_proceso, nro_contrato_sicac, nro_contrato_sap, id_empresa, tiempo_ejecucion, monto_adjudicado_bs, monto_adjudicado_usd, periodo_valuacion_desde, periodo_valuacion_hasta, monto_valuacion, nro_proforma, observaciones, notas)
    VALUES (NEW.id_valuacion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.solped, NEW.presupuesto_base_bs, NEW.presupuesto_base_usd, NEW.descripcion_proceso, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.nro_proceso, NEW.nro_contrato_sicac, NEW.nro_contrato_sap, NEW.id_empresa, NEW.tiempo_ejecucion, NEW.monto_adjudicado_bs, NEW.monto_adjudicado_usd, NEW.periodo_valuacion_desde, NEW.periodo_valuacion_hasta, NEW.monto_valuacion, NEW.nro_proforma, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_val_auditoria AFTER UPDATE ON valuaciones
FOR EACH ROW BEGIN
    INSERT INTO hist_valuaciones (id_valuacion, id_gerencia, id_superintendencia, id_emisor, documento, solped, presupuesto_base_bs, presupuesto_base_usd, descripcion_proceso, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, nro_proceso, nro_contrato_sicac, nro_contrato_sap, id_empresa, tiempo_ejecucion, monto_adjudicado_bs, monto_adjudicado_usd, periodo_valuacion_desde, periodo_valuacion_hasta, monto_valuacion, nro_proforma, observaciones, notas)
    VALUES (NEW.id_valuacion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.solped, NEW.presupuesto_base_bs, NEW.presupuesto_base_usd, NEW.descripcion_proceso, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.nro_proceso, NEW.nro_contrato_sicac, NEW.nro_contrato_sap, NEW.id_empresa, NEW.tiempo_ejecucion, NEW.monto_adjudicado_bs, NEW.monto_adjudicado_usd, NEW.periodo_valuacion_desde, NEW.periodo_valuacion_hasta, NEW.monto_valuacion, NEW.nro_proforma, NEW.observaciones, NEW.notas);
    UPDATE valuaciones SET fecha_actualizacion = CURRENT_DATE WHERE id_valuacion = NEW.id_valuacion;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_valuaciones AS
SELECT
    v.id_valuacion,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    v.documento,
    v.solped,
    v.presupuesto_base_bs,
    v.presupuesto_base_usd,
    v.descripcion_proceso,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    v.fecha_recibido,
    v.fecha_devuelto,
    COALESCE(re.nombre, 'NO APLICA') AS receptor,
    v.nro_proceso,
    v.nro_contrato_sicac,
    v.nro_contrato_sap,
    COALESCE(emp.nombre, 'NO APLICA') AS empresa_adjudicada,
    v.tiempo_ejecucion,
    v.monto_adjudicado_bs,
    v.monto_adjudicado_usd,
    v.periodo_valuacion_desde,
    v.periodo_valuacion_hasta,
    v.monto_valuacion,
    v.nro_proforma,
    v.observaciones,
    v.notas,
    v.fecha_creacion,
    v.fecha_actualizacion
FROM valuaciones v
LEFT JOIN cat_gerencia g ON v.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON v.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON v.id_emisor = em.id
LEFT JOIN cat_responsables re ON v.id_receptor = re.id
LEFT JOIN cat_estatus_detalle ed ON v.id_estatus = ed.id
LEFT JOIN cat_empresas emp ON v.id_empresa = emp.id;

CREATE INDEX IF NOT EXISTS idx_val_estatus ON valuaciones(id_estatus);
CREATE INDEX IF NOT EXISTS idx_val_empresa ON valuaciones(id_empresa);
CREATE INDEX IF NOT EXISTS idx_hist_val_id ON hist_valuaciones(id_valuacion);


-- ==========================================
-- 🔹 MÓDULO 5: PARA APROBACIÓN JD
-- ==========================================
CREATE TABLE IF NOT EXISTS aprobacion_jd (
    id_aprobacion_jd      INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    id_documento          INTEGER,
    solped                TEXT,
    fecha_presupuesto_base DATE,
    presupuesto_base_bs   REAL,
    tipo_cambio           REAL,
    presupuesto_base_usd  REAL,
    id_plan               INTEGER,
    descripcion_proceso   TEXT,
    cantidad_frentes      INTEGER,
    id_estatus            INTEGER DEFAULT 1,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    tiempo_ejecucion      TEXT,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_jd_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_jd_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_jd_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_jd_re  FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_jd_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    CONSTRAINT fk_jd_plan FOREIGN KEY (id_plan) REFERENCES cat_plan_contratacion(id),
    CONSTRAINT fk_jd_doc FOREIGN KEY (id_documento) REFERENCES cat_documento(id)
);

CREATE TABLE IF NOT EXISTS hist_aprobacion_jd (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_aprobacion_jd      INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    id_documento          INTEGER,
    solped                TEXT,
    fecha_presupuesto_base DATE,
    presupuesto_base_bs   REAL,
    tipo_cambio           REAL,
    presupuesto_base_usd  REAL,
    id_plan               INTEGER,
    descripcion_proceso   TEXT,
    cantidad_frentes      INTEGER,
    id_estatus            INTEGER,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    tiempo_ejecucion      TEXT,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_aprobacion_jd) REFERENCES aprobacion_jd(id_aprobacion_jd),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    FOREIGN KEY (id_plan) REFERENCES cat_plan_contratacion(id),
    FOREIGN KEY (id_documento) REFERENCES cat_documento(id)
);

CREATE TRIGGER IF NOT EXISTS trg_jd_inicial AFTER INSERT ON aprobacion_jd
FOR EACH ROW BEGIN
    INSERT INTO hist_aprobacion_jd (id_aprobacion_jd, id_gerencia, id_superintendencia, id_emisor, id_documento, solped, fecha_presupuesto_base, presupuesto_base_bs, tipo_cambio, presupuesto_base_usd, id_plan, descripcion_proceso, cantidad_frentes, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, tiempo_ejecucion, observaciones, notas)
    VALUES (NEW.id_aprobacion_jd, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.solped, NEW.fecha_presupuesto_base, NEW.presupuesto_base_bs, NEW.tipo_cambio, NEW.presupuesto_base_usd, NEW.id_plan, NEW.descripcion_proceso, NEW.cantidad_frentes, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.tiempo_ejecucion, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_jd_auditoria AFTER UPDATE ON aprobacion_jd
FOR EACH ROW BEGIN
    INSERT INTO hist_aprobacion_jd (id_aprobacion_jd, id_gerencia, id_superintendencia, id_emisor, id_documento, solped, fecha_presupuesto_base, presupuesto_base_bs, tipo_cambio, presupuesto_base_usd, id_plan, descripcion_proceso, cantidad_frentes, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, tiempo_ejecucion, observaciones, notas)
    VALUES (NEW.id_aprobacion_jd, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.solped, NEW.fecha_presupuesto_base, NEW.presupuesto_base_bs, NEW.tipo_cambio, NEW.presupuesto_base_usd, NEW.id_plan, NEW.descripcion_proceso, NEW.cantidad_frentes, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.tiempo_ejecucion, NEW.observaciones, NEW.notas);
    UPDATE aprobacion_jd SET fecha_actualizacion = CURRENT_DATE WHERE id_aprobacion_jd = NEW.id_aprobacion_jd;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_aprobacion_jd AS
SELECT
    j.id_aprobacion_jd,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    d.nombre AS documento,
    j.solped,
    j.fecha_presupuesto_base,
    j.presupuesto_base_bs,
    j.tipo_cambio,
    j.presupuesto_base_usd,
    p.nombre AS plan_contrataciones,
    j.descripcion_proceso,
    j.cantidad_frentes,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    j.fecha_recibido,
    j.fecha_devuelto,
    COALESCE(re.nombre, 'NO APLICA') AS receptor,
    j.tiempo_ejecucion,
    j.observaciones,
    j.notas,
    j.fecha_creacion,
    j.fecha_actualizacion
FROM aprobacion_jd j
LEFT JOIN cat_gerencia g ON j.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON j.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON j.id_emisor = em.id
LEFT JOIN cat_responsables re ON j.id_receptor = re.id
LEFT JOIN cat_estatus_detalle ed ON j.id_estatus = ed.id
LEFT JOIN cat_plan_contratacion p ON j.id_plan = p.id
LEFT JOIN cat_documento d ON j.id_documento = d.id;

CREATE INDEX IF NOT EXISTS idx_jd_estatus ON aprobacion_jd(id_estatus);
CREATE INDEX IF NOT EXISTS idx_hist_jd_id ON hist_aprobacion_jd(id_aprobacion_jd);


-- ==========================================
-- 🔹 MÓDULO 6: CERTIFICACIÓN BDU
-- ==========================================
CREATE TABLE IF NOT EXISTS certificacion_bdu (
    id_certificacion_bdu  INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    id_documento          INTEGER,
    presupuesto_base_total_usd REAL,
    monto_adjudicado_total_usd REAL,
    monto_contrato        REAL,
    monto_ejecutado       REAL,
    monto_pagado          REAL,
    id_estatus            INTEGER DEFAULT 1,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_bdu_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_bdu_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_bdu_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_bdu_re  FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_bdu_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    CONSTRAINT fk_bdu_doc FOREIGN KEY (id_documento) REFERENCES cat_documento(id)
);

CREATE TABLE IF NOT EXISTS hist_certificacion_bdu (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_certificacion_bdu  INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    id_documento          INTEGER,
    presupuesto_base_total_usd REAL,
    monto_adjudicado_total_usd REAL,
    monto_contrato        REAL,
    monto_ejecutado       REAL,
    monto_pagado          REAL,
    id_estatus            INTEGER,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_certificacion_bdu) REFERENCES certificacion_bdu(id_certificacion_bdu),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id),
    FOREIGN KEY (id_documento) REFERENCES cat_documento(id)
);

CREATE TRIGGER IF NOT EXISTS trg_bdu_inicial AFTER INSERT ON certificacion_bdu
FOR EACH ROW BEGIN
    INSERT INTO hist_certificacion_bdu (id_certificacion_bdu, id_gerencia, id_superintendencia, id_emisor, id_documento, presupuesto_base_total_usd, monto_adjudicado_total_usd, monto_contrato, monto_ejecutado, monto_pagado, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_certificacion_bdu, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.presupuesto_base_total_usd, NEW.monto_adjudicado_total_usd, NEW.monto_contrato, NEW.monto_ejecutado, NEW.monto_pagado, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_bdu_auditoria AFTER UPDATE ON certificacion_bdu
FOR EACH ROW BEGIN
    INSERT INTO hist_certificacion_bdu (id_certificacion_bdu, id_gerencia, id_superintendencia, id_emisor, id_documento, presupuesto_base_total_usd, monto_adjudicado_total_usd, monto_contrato, monto_ejecutado, monto_pagado, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_certificacion_bdu, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.id_documento, NEW.presupuesto_base_total_usd, NEW.monto_adjudicado_total_usd, NEW.monto_contrato, NEW.monto_ejecutado, NEW.monto_pagado, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
    UPDATE certificacion_bdu SET fecha_actualizacion = CURRENT_DATE WHERE id_certificacion_bdu = NEW.id_certificacion_bdu;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_certificacion_bdu AS
SELECT
    b.id_certificacion_bdu,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    d.nombre AS documento,
    b.presupuesto_base_total_usd,
    b.monto_adjudicado_total_usd,
    b.monto_contrato,
    b.monto_ejecutado,
    b.monto_pagado,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    b.fecha_recibido,
    b.fecha_devuelto,
    COALESCE(re.nombre, 'NO APLICA') AS receptor,
    b.observaciones,
    b.notas,
    b.fecha_creacion,
    b.fecha_actualizacion
FROM certificacion_bdu b
LEFT JOIN cat_gerencia g ON b.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON b.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON b.id_emisor = em.id
LEFT JOIN cat_responsables re ON b.id_receptor = re.id
LEFT JOIN cat_estatus_detalle ed ON b.id_estatus = ed.id
LEFT JOIN cat_documento d ON b.id_documento = d.id;

CREATE INDEX IF NOT EXISTS idx_bdu_estatus ON certificacion_bdu(id_estatus);
CREATE INDEX IF NOT EXISTS idx_hist_bdu_id ON hist_certificacion_bdu(id_certificacion_bdu);


-- ==========================================
-- 🔹 MÓDULO 7: VACACIONES
-- ==========================================
CREATE TABLE IF NOT EXISTS vacaciones (
    id_vacacion           INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    anio                  INTEGER,
    cantidad_dias         INTEGER,
    fecha_desde           DATE,
    fecha_hasta           DATE,
    id_estatus            INTEGER DEFAULT 1,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_vac_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_vac_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_vac_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_vac_re  FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_vac_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TABLE IF NOT EXISTS hist_vacaciones (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_vacacion           INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    anio                  INTEGER,
    cantidad_dias         INTEGER,
    fecha_desde           DATE,
    fecha_hasta           DATE,
    id_estatus            INTEGER,
    fecha_recibido        DATE,
    fecha_devuelto        DATE,
    id_receptor           INTEGER,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_vacacion) REFERENCES vacaciones(id_vacacion),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TRIGGER IF NOT EXISTS trg_vac_inicial AFTER INSERT ON vacaciones
FOR EACH ROW BEGIN
    INSERT INTO hist_vacaciones (id_vacacion, id_gerencia, id_superintendencia, id_emisor, documento, anio, cantidad_dias, fecha_desde, fecha_hasta, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_vacacion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.anio, NEW.cantidad_dias, NEW.fecha_desde, NEW.fecha_hasta, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_vac_auditoria AFTER UPDATE ON vacaciones
FOR EACH ROW BEGIN
    INSERT INTO hist_vacaciones (id_vacacion, id_gerencia, id_superintendencia, id_emisor, documento, anio, cantidad_dias, fecha_desde, fecha_hasta, id_estatus, fecha_recibido, fecha_devuelto, id_receptor, observaciones, notas)
    VALUES (NEW.id_vacacion, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.anio, NEW.cantidad_dias, NEW.fecha_desde, NEW.fecha_hasta, NEW.id_estatus, NEW.fecha_recibido, NEW.fecha_devuelto, NEW.id_receptor, NEW.observaciones, NEW.notas);
    UPDATE vacaciones SET fecha_actualizacion = CURRENT_DATE WHERE id_vacacion = NEW.id_vacacion;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_vacaciones AS
SELECT
    v.id_vacacion,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    v.documento,
    v.anio,
    v.cantidad_dias,
    v.fecha_desde,
    v.fecha_hasta,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    v.fecha_recibido,
    v.fecha_devuelto,
    COALESCE(re.nombre, 'NO APLICA') AS receptor,
    v.observaciones,
    v.notas,
    v.fecha_creacion,
    v.fecha_actualizacion
FROM vacaciones v
LEFT JOIN cat_gerencia g ON v.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON v.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON v.id_emisor = em.id
LEFT JOIN cat_responsables re ON v.id_receptor = re.id
LEFT JOIN cat_estatus_detalle ed ON v.id_estatus = ed.id;

CREATE INDEX IF NOT EXISTS idx_vac_estatus ON vacaciones(id_estatus);
CREATE INDEX IF NOT EXISTS idx_hist_vac_id ON hist_vacaciones(id_vacacion);


-- ==========================================
-- 🔹 MÓDULO 8: REPOSO MÉDICO
-- ==========================================
CREATE TABLE IF NOT EXISTS reposos_medicos (
    id_reposo_medico      INTEGER PRIMARY KEY AUTOINCREMENT,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    dias_periodo          INTEGER,
    fecha_desde           DATE,
    fecha_hasta           DATE,
    id_estatus            INTEGER DEFAULT 1,
    fecha_recibido        DATE,
    observaciones         TEXT,
    notas                 TEXT,
    fecha_creacion        DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion   DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_rep_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    CONSTRAINT fk_rep_sup FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_rep_em  FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    CONSTRAINT fk_rep_est FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TABLE IF NOT EXISTS hist_reposos_medicos (
    id_movimiento         INTEGER PRIMARY KEY AUTOINCREMENT,
    id_reposo_medico      INTEGER NOT NULL,
    id_gerencia           INTEGER,
    id_superintendencia   INTEGER,
    id_emisor             INTEGER,
    documento             TEXT,
    dias_periodo          INTEGER,
    fecha_desde           DATE,
    fecha_hasta           DATE,
    id_estatus            INTEGER,
    fecha_recibido        DATE,
    observaciones         TEXT,
    notas                 TEXT,
    FOREIGN KEY (id_reposo_medico) REFERENCES reposos_medicos(id_reposo_medico),
    FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor) REFERENCES cat_responsables(id),
    FOREIGN KEY (id_estatus) REFERENCES cat_estatus_detalle(id)
);

CREATE TRIGGER IF NOT EXISTS trg_rep_inicial AFTER INSERT ON reposos_medicos
FOR EACH ROW BEGIN
    INSERT INTO hist_reposos_medicos (id_reposo_medico, id_gerencia, id_superintendencia, id_emisor, documento, dias_periodo, fecha_desde, fecha_hasta, id_estatus, fecha_recibido, observaciones, notas)
    VALUES (NEW.id_reposo_medico, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.dias_periodo, NEW.fecha_desde, NEW.fecha_hasta, NEW.id_estatus, NEW.fecha_recibido, NEW.observaciones, NEW.notas);
END;

CREATE TRIGGER IF NOT EXISTS trg_rep_auditoria AFTER UPDATE ON reposos_medicos
FOR EACH ROW BEGIN
    INSERT INTO hist_reposos_medicos (id_reposo_medico, id_gerencia, id_superintendencia, id_emisor, documento, dias_periodo, fecha_desde, fecha_hasta, id_estatus, fecha_recibido, observaciones, notas)
    VALUES (NEW.id_reposo_medico, NEW.id_gerencia, NEW.id_superintendencia, NEW.id_emisor, NEW.documento, NEW.dias_periodo, NEW.fecha_desde, NEW.fecha_hasta, NEW.id_estatus, NEW.fecha_recibido, NEW.observaciones, NEW.notas);
    UPDATE reposos_medicos SET fecha_actualizacion = CURRENT_DATE WHERE id_reposo_medico = NEW.id_reposo_medico;
END;

CREATE VIEW IF NOT EXISTS vw_reporte_reposos_medicos AS
SELECT
    r.id_reposo_medico,
    g.nombre AS gerencia,
    s.nombre AS superintendencia,
    em.nombre AS emisor,
    r.documento,
    r.dias_periodo,
    r.fecha_desde,
    r.fecha_hasta,
    COALESCE(ed.nombre, 'NO APLICA') AS estatus_detalle,
    r.fecha_recibido,
    r.observaciones,
    r.notas,
    r.fecha_creacion,
    r.fecha_actualizacion
FROM reposos_medicos r
LEFT JOIN cat_gerencia g ON r.id_gerencia = g.id
LEFT JOIN cat_superintendencia s ON r.id_superintendencia = s.id
LEFT JOIN cat_responsables em ON r.id_emisor = em.id
LEFT JOIN cat_estatus_detalle ed ON r.id_estatus = ed.id;

CREATE INDEX IF NOT EXISTS idx_rep_estatus ON reposos_medicos(id_estatus);
CREATE INDEX IF NOT EXISTS idx_hist_rep_id ON hist_reposos_medicos(id_reposo_medico);
