PRAGMA foreign_keys = ON;

-- ==========================================
-- 🔹 1. CATÁLOGOS MAESTROS (Independientes)
-- ==========================================
CREATE TABLE cat_gerencia (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_documento (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_plan_contratacion (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_modalidad (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_art (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_tipo_contrato (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_estatus_detalle (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_resultado_proceso (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_empresas (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);
CREATE TABLE cat_responsables (id INTEGER PRIMARY KEY, nombre TEXT UNIQUE);

-- ==========================================
-- 🔹 2. CATÁLOGOS CON RELACIONES
-- ==========================================
CREATE TABLE cat_superintendencia (
    id INTEGER PRIMARY KEY,
    nombre TEXT UNIQUE,
    id_gerencia INTEGER,
    CONSTRAINT fk_sup_ger FOREIGN KEY (id_gerencia) REFERENCES cat_gerencia(id)
);

-- ==========================================
-- 🔹 3. TABLA PRINCIPAL: EXPEDIENTES
-- ==========================================
CREATE TABLE expedientes (
    id_expediente           INTEGER PRIMARY KEY AUTOINCREMENT,
    solped                  TEXT,
    id_gerencia             INTEGER,
    id_superintendencia     INTEGER,
    id_emisor               INTEGER,
    id_documento            INTEGER,
    fecha_presupuesto_base  DATE,
    presupuesto_base_usd    REAL,
    tipo_cambio             REAL,
    presupuesto_base_bs     REAL,
    id_plan                 INTEGER,
    descripcion_proceso     TEXT,
    id_modalidad            INTEGER,
    id_art                  INTEGER,
    id_tipo_contrato        INTEGER,
    nro_acta_apertura       TEXT,
    cantidad_frentes        INTEGER,
    nro_resolucion_jd       TEXT,
    id_estatus              INTEGER DEFAULT 1,
    fecha_recibido          DATE,
    fecha_devuelto          DATE,
    id_receptor             INTEGER,
    nro_proceso             TEXT,
    id_resultado            INTEGER,
    nro_contrato_sicac      TEXT,
    nro_contrato_sap        TEXT,
    id_empresa              INTEGER,
    tiempo_ejecucion        TEXT,
    monto_adjudicado_bs     REAL,
    monto_adjudicado_usd    REAL,
    fecha_firma_contrato    DATE,
    observaciones           TEXT,
    notas                   TEXT,
    fecha_creacion          DATE DEFAULT CURRENT_DATE,
    fecha_actualizacion     DATE DEFAULT CURRENT_DATE,
    CONSTRAINT fk_exp_ger      FOREIGN KEY (id_gerencia)         REFERENCES cat_gerencia(id),
    CONSTRAINT fk_exp_sup      FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    CONSTRAINT fk_exp_emisor   FOREIGN KEY (id_emisor)           REFERENCES cat_responsables(id),
    CONSTRAINT fk_exp_receptor FOREIGN KEY (id_receptor)         REFERENCES cat_responsables(id),
    CONSTRAINT fk_exp_doc      FOREIGN KEY (id_documento)        REFERENCES cat_documento(id),
    CONSTRAINT fk_exp_plan     FOREIGN KEY (id_plan)             REFERENCES cat_plan_contratacion(id),
    CONSTRAINT fk_exp_mod      FOREIGN KEY (id_modalidad)        REFERENCES cat_modalidad(id),
    CONSTRAINT fk_exp_art      FOREIGN KEY (id_art)              REFERENCES cat_art(id),
    CONSTRAINT fk_exp_tc       FOREIGN KEY (id_tipo_contrato)    REFERENCES cat_tipo_contrato(id),
    CONSTRAINT fk_exp_est      FOREIGN KEY (id_estatus)          REFERENCES cat_estatus_detalle(id),
    CONSTRAINT fk_exp_res      FOREIGN KEY (id_resultado)        REFERENCES cat_resultado_proceso(id),
    CONSTRAINT fk_exp_emp      FOREIGN KEY (id_empresa)          REFERENCES cat_empresas(id)
);

-- ==========================================
-- 🔹 4. HISTORIAL DE MOVIMIENTOS (Snapshot Normalizado)
-- ==========================================
CREATE TABLE historial_movimientos (
    id_movimiento           INTEGER PRIMARY KEY AUTOINCREMENT,
    id_expediente           INTEGER NOT NULL,
    solped                  TEXT,
    id_gerencia             INTEGER,
    id_superintendencia     INTEGER,
    id_emisor               INTEGER,
    id_receptor             INTEGER,
    id_documento            INTEGER,
    id_plan                 INTEGER,
    id_modalidad            INTEGER,
    id_art                  INTEGER,
    id_tipo_contrato        INTEGER,
    id_estatus              INTEGER,
    id_resultado            INTEGER,
    id_empresa              INTEGER,
    fecha_recibido          DATE,
    fecha_devuelto          DATE,
    fecha_presupuesto_base  DATE,
    fecha_firma_contrato    DATE,
    nro_proceso             TEXT,
    nro_acta_apertura       TEXT,
    nro_resolucion_jd       TEXT,
    nro_contrato_sicac      TEXT,
    nro_contrato_sap        TEXT,
    descripcion_proceso     TEXT,
    presupuesto_base_usd    REAL,
    presupuesto_base_bs     REAL,
    tipo_cambio             REAL,
    monto_adjudicado_usd    REAL,
    monto_adjudicado_bs     REAL,
    tiempo_ejecucion        TEXT,
    cantidad_frentes        INTEGER,
    observaciones           TEXT,
    notas                   TEXT,
    FOREIGN KEY (id_expediente)       REFERENCES expedientes(id_expediente),
    FOREIGN KEY (id_gerencia)         REFERENCES cat_gerencia(id),
    FOREIGN KEY (id_superintendencia) REFERENCES cat_superintendencia(id),
    FOREIGN KEY (id_emisor)           REFERENCES cat_responsables(id),
    FOREIGN KEY (id_receptor)         REFERENCES cat_responsables(id),
    FOREIGN KEY (id_documento)        REFERENCES cat_documento(id),
    FOREIGN KEY (id_plan)             REFERENCES cat_plan_contratacion(id),
    FOREIGN KEY (id_modalidad)        REFERENCES cat_modalidad(id),
    FOREIGN KEY (id_art)              REFERENCES cat_art(id),
    FOREIGN KEY (id_tipo_contrato)    REFERENCES cat_tipo_contrato(id),
    FOREIGN KEY (id_estatus)          REFERENCES cat_estatus_detalle(id),
    FOREIGN KEY (id_resultado)        REFERENCES cat_resultado_proceso(id),
    FOREIGN KEY (id_empresa)          REFERENCES cat_empresas(id)
);

-- ==========================================
-- 🔹 5. ÍNDICES PARA RENDIMIENTO
-- ==========================================
CREATE INDEX idx_exp_solped              ON expedientes(solped);
CREATE INDEX idx_exp_gerencia            ON expedientes(id_gerencia);
CREATE INDEX idx_exp_estatus             ON expedientes(id_estatus);
CREATE INDEX idx_exp_empresa             ON expedientes(id_empresa);
CREATE INDEX idx_exp_fecha_presup        ON expedientes(fecha_presupuesto_base);
CREATE INDEX idx_exp_fecha_creacion      ON expedientes(fecha_creacion);
CREATE INDEX idx_exp_fecha_actualizacion ON expedientes(fecha_actualizacion);

CREATE INDEX idx_hist_mov_expediente     ON historial_movimientos(id_expediente);
CREATE INDEX idx_hist_mov_estatus        ON historial_movimientos(id_estatus);
CREATE INDEX idx_hist_mov_emisor         ON historial_movimientos(id_emisor);
CREATE INDEX idx_hist_mov_receptor       ON historial_movimientos(id_receptor);

-- ==========================================
-- 🔹 6. TRIGGERS DE AUDITORÍA
-- ==========================================

-- Snapshot inicial al crear expediente
CREATE TRIGGER trg_exp_snapshot_inicial AFTER INSERT ON expedientes
FOR EACH ROW
BEGIN
    INSERT INTO historial_movimientos (
        id_expediente, solped, id_gerencia, id_superintendencia,
        id_emisor, id_receptor, id_documento, id_plan,
        id_modalidad, id_art, id_tipo_contrato, id_estatus,
        id_resultado, id_empresa,
        fecha_recibido, fecha_devuelto, fecha_presupuesto_base,
        fecha_firma_contrato,
        nro_proceso, nro_acta_apertura, nro_resolucion_jd,
        nro_contrato_sicac, nro_contrato_sap,
        descripcion_proceso,
        presupuesto_base_usd, presupuesto_base_bs, tipo_cambio,
        monto_adjudicado_usd, monto_adjudicado_bs,
        tiempo_ejecucion, cantidad_frentes,
        observaciones, notas
    ) VALUES (
        NEW.id_expediente, NEW.solped, NEW.id_gerencia, NEW.id_superintendencia,
        NEW.id_emisor, NEW.id_receptor, NEW.id_documento, NEW.id_plan,
        NEW.id_modalidad, NEW.id_art, NEW.id_tipo_contrato, NEW.id_estatus,
        NEW.id_resultado, NEW.id_empresa,
        NEW.fecha_recibido, NEW.fecha_devuelto, NEW.fecha_presupuesto_base,
        NEW.fecha_firma_contrato,
        NEW.nro_proceso, NEW.nro_acta_apertura, NEW.nro_resolucion_jd,
        NEW.nro_contrato_sicac, NEW.nro_contrato_sap,
        NEW.descripcion_proceso,
        NEW.presupuesto_base_usd, NEW.presupuesto_base_bs, NEW.tipo_cambio,
        NEW.monto_adjudicado_usd, NEW.monto_adjudicado_bs,
        NEW.tiempo_ejecucion, NEW.cantidad_frentes,
        NEW.observaciones, NEW.notas
    );
END;
CREATE TRIGGER trg_exp_auditoria AFTER UPDATE ON expedientes
FOR EACH ROW
BEGIN
    INSERT INTO historial_movimientos (
        id_expediente, solped, id_gerencia, id_superintendencia,
        id_emisor, id_receptor, id_documento, id_plan,
        id_modalidad, id_art, id_tipo_contrato, id_estatus,
        id_resultado, id_empresa,
        fecha_recibido, fecha_devuelto, fecha_presupuesto_base,
        fecha_firma_contrato,
        nro_proceso, nro_acta_apertura, nro_resolucion_jd,
        nro_contrato_sicac, nro_contrato_sap,
        descripcion_proceso,
        presupuesto_base_usd, presupuesto_base_bs, tipo_cambio,
        monto_adjudicado_usd, monto_adjudicado_bs,
        tiempo_ejecucion, cantidad_frentes,
        observaciones, notas
    ) VALUES (
        NEW.id_expediente, NEW.solped, NEW.id_gerencia, NEW.id_superintendencia,
        NEW.id_emisor, NEW.id_receptor, NEW.id_documento, NEW.id_plan,
        NEW.id_modalidad, NEW.id_art, NEW.id_tipo_contrato, NEW.id_estatus,
        NEW.id_resultado, NEW.id_empresa,
        NEW.fecha_recibido, NEW.fecha_devuelto, NEW.fecha_presupuesto_base,
        NEW.fecha_firma_contrato,
        NEW.nro_proceso, NEW.nro_acta_apertura, NEW.nro_resolucion_jd,
        NEW.nro_contrato_sicac, NEW.nro_contrato_sap,
        NEW.descripcion_proceso,
        NEW.presupuesto_base_usd, NEW.presupuesto_base_bs, NEW.tipo_cambio,
        NEW.monto_adjudicado_usd, NEW.monto_adjudicado_bs,
        NEW.tiempo_ejecucion, NEW.cantidad_frentes,
        NEW.observaciones, NEW.notas
    );

    UPDATE expedientes
    SET id_estatus = (SELECT id FROM cat_estatus_detalle WHERE nombre = 'PENDIENTE' LIMIT 1)
    WHERE NEW.fecha_firma_contrato IS NULL
      AND OLD.fecha_firma_contrato IS NOT NULL
      AND id_expediente = NEW.id_expediente;

    UPDATE expedientes
    SET id_estatus = (SELECT id FROM cat_estatus_detalle WHERE nombre = 'FIRMADO' LIMIT 1)
    WHERE NEW.fecha_firma_contrato IS NOT NULL
      AND id_expediente = NEW.id_expediente;

    UPDATE expedientes
    SET fecha_actualizacion = CURRENT_DATE
    WHERE id_expediente = NEW.id_expediente;
END;

-- ==========================================
-- 🔹 7. VISTA PARA EXPORTAR A EXCEL
-- ==========================================
CREATE VIEW vw_reporte_excel_contrataciones AS
SELECT 
    e.id_expediente,
    COALESCE(e.solped, 'SIN_SOLPED')            AS solped,
    g.nombre                                     AS gerencia,
    s.nombre                                     AS superintendencia,
    emisor.nombre                                AS emisor,
    d.nombre                                     AS documento,
    e.fecha_presupuesto_base,
    e.presupuesto_base_usd,
    e.tipo_cambio,
    e.presupuesto_base_bs,
    p.nombre                                     AS plan_contrataciones,
    e.descripcion_proceso,
    m.nombre                                     AS modalidad_contratacion,
    a.nombre                                     AS art,
    tc.nombre                                    AS tipo_contrato,
    COALESCE(e.nro_acta_apertura, 'NO POSEE')    AS nro_acta_apertura,
    e.cantidad_frentes,
    COALESCE(e.nro_resolucion_jd, 'NO APLICA')   AS nro_resolucion_jd,
    COALESCE(ed.nombre, 'NO APLICA')             AS estatus_detalle,
    e.fecha_recibido,
    e.fecha_devuelto,
    COALESCE(receptor.nombre, 'NO APLICA')       AS receptor,
    COALESCE(e.nro_proceso, 'NO APLICA')         AS nro_proceso,
    COALESCE(rp.nombre, 'NO APLICA')             AS resultados_proceso,
    COALESCE(e.nro_contrato_sicac, 'NO POSEE')   AS nro_contrato_sicac,
    e.nro_contrato_sap,
    COALESCE(emp.nombre, 'NO APLICA')            AS empresa_adjudicada,
    e.tiempo_ejecucion,
    e.monto_adjudicado_bs,
    e.monto_adjudicado_usd,
    COALESCE(e.fecha_firma_contrato, 'NO APLICA') AS fecha_firma_contrato,
    e.observaciones,
    e.notas,
    e.fecha_creacion,
    e.fecha_actualizacion
FROM expedientes e
LEFT JOIN cat_gerencia g          ON e.id_gerencia         = g.id
LEFT JOIN cat_superintendencia s  ON e.id_superintendencia = s.id
LEFT JOIN cat_documento d         ON e.id_documento        = d.id
LEFT JOIN cat_plan_contratacion p ON e.id_plan             = p.id
LEFT JOIN cat_modalidad m         ON e.id_modalidad        = m.id
LEFT JOIN cat_art a               ON e.id_art              = a.id
LEFT JOIN cat_tipo_contrato tc    ON e.id_tipo_contrato    = tc.id
LEFT JOIN cat_estatus_detalle ed  ON e.id_estatus          = ed.id
LEFT JOIN cat_resultado_proceso rp ON e.id_resultado       = rp.id
LEFT JOIN cat_empresas emp        ON e.id_empresa          = emp.id
LEFT JOIN cat_responsables emisor ON e.id_emisor           = emisor.id
LEFT JOIN cat_responsables receptor ON e.id_receptor       = receptor.id;

-- ==========================================
-- 🔹 8. DATOS INICIALES (CATÁLOGOS)
-- ==========================================

-- 1. GERENCIAS
INSERT INTO cat_gerencia (id, nombre) VALUES
(1, 'SIHO-A'), (2, 'TÉCNICA'), (3, 'OPERACIONES'), (4, 'SSGG'), (5, 'JURÍDICO'),
(6, 'FINANZAS'), (7, 'CONTRATACIÓN'), (8, 'RRHH'), (9, 'ASUNTOS GUBERNAMENTALES'), (10, 'COMISIÓN');

-- 2. SUPERINTENDENCIAS
INSERT INTO cat_superintendencia (id, nombre, id_gerencia) VALUES
(1, 'SIHO-A', 1),
(2, 'INFRAESTRUCTURA', 2), (3, 'PERFORACIÓN', 2), (4, 'YACIMIENTOS', 2), (5, 'OPTIMIZACIÓN', 2),
(6, 'OPERACIÓN DE PRODUCCIÓN', 3), (7, 'MANTENIMIENTO', 3),
(8, 'SSGG', 4), (9, 'JURÍDICO', 5), (10, 'FINANZAS', 6),
(11, 'CONTRATACIÓN', 7), (12, 'RRHH', 8),
(13, 'ASUNTOS GUBERNAMENTALES', 9), (14, 'COMISIÓN', 10);

-- 3. DOCUMENTOS
INSERT INTO cat_documento (id, nombre) VALUES
(1,  'ACTA DE INICIO SOLICITUD (A)'),
(2,  'ACTA DE MODIFICACIÓN DEL CONTRATO (A)'),
(3,  'ACTA DE OTRAS CONSIDERACIONES (A)'),
(4,  'ACTA DE OTRAS CONSIDERACIONES (A) / ACTA DE OTORGAMIENTO / NOTIFICACIÓN DE ADJUDICACIÓN'),
(5,  'ACTA DE OTRAS CONSIDERACIONES (A) / ACTO MOTIVADO / ACTA DE OTORGAMIENTO / NOTIFICACIÓN'),
(6,  'ACTA DE OTRAS CONSIDERACIONES / ANÁLISIS ECONÓMICO'),
(7,  'ACTA DE OTORGAMIENTO / NOTIFICACIÓN DE ADJUDICACIÓN'),
(8,  'ACTA DE RESULTADOS DE CALIFICACIÓN Y EVALUACIÓN (A)'),
(9,  'ACTO MOTIVADO'),
(10, 'ACTO MOTIVADO Y ACTA DE OTRAS CONSIDERACIONES (A)'),
(11, 'ACTUALIZACIÓN DE PRESUPUESTO BASE'),
(12, 'ADDENDUM / DECISIÓN DE GERENCIA APROBACIÓN DE LA MODIFICACIÓN'),
(13, 'ANÁLISIS ECONÓMICO / ACTA DE OTORGAMIENTO / CONTRATO'),
(14, 'ANÁLISIS ECONÓMICO / ACTA DE RESULTADOS DE CALIFICACIÓN Y EVALUACIÓN (A)'),
(15, 'ANÁLISIS ECONÓMICO / CONTRATO'),
(16, 'ANÁLISIS ECONÓMICO REV.1'),
(17, 'CONTRATO'),
(18, 'CONTRATO DE SERVICIOS'),
(19, 'DECISIÓN DE GERENCIA'),
(20, 'DECISIÓN DE GERENCIA INICIO'),
(21, 'DECISIÓN DE GERENCIA MODIFICACIÓN / ADDENDUM'),
(22, 'DESCRIPCIÓN DE PROCESO Y ESPECIFICACIONES TÉCNICAS'),
(23, 'DESCRIPCIÓN DE PROCESO, ESPECIFICACIONES TÉCNICAS Y JUSTIFICACIÓN'),
(24, 'ESPECIFICACIONES TÉCNICAS Y DESCRIPCIÓN DEL PROCESO'),
(25, 'JUSTIFICACIÓN, MODIFICACIÓN Y ACTA DE OTRAS CONSIDERACIONES (A)'),
(26, 'PRESUPUESTO BASE / ESPECIFICACIONES TÉCNICAS / ACTUALIZACIÓN DE PRESUPUESTO BASE'),
(27, 'PRESUPUESTO BASE / ESPECIFICACIONES TÉCNICAS Y DESCRIPCIÓN DEL PROCESO'),
(28, 'SOLPED / PRESUPUESTO BASE / DESCRIPCIÓN DEL PROCESO / JUSTIFICACIÓN / INFORME TÉCNICO DE PRECALIFICACIÓN / ESPECIFICACIONES TÉCNICAS');

-- 4. PLANES DE CONTRATACIÓN
INSERT INTO cat_plan_contratacion (id, nombre) VALUES
(1, 'ARRASTRE 2025'), (2, 'PLAN 2026'), (3, 'ADICIONAL (DIRECTOS) 2026'), (4, 'PLAN-ADICIONAL 2026');

-- 5. MODALIDADES
INSERT INTO cat_modalidad (id, nombre) VALUES
(1, 'CONCURSO ABIERTO'), (2, 'CONCURSO CERRADO'), (3, 'CONSULTA DE PRECIOS'), (4, 'CONTRATACIÓN DIRECTA');

-- 6. ARTÍCULOS NORMATIVA INTERNA
INSERT INTO cat_art (id, nombre) VALUES
(1, '5 N - 08'), (2, '77 N - 01'), (3, '77 N - 02'), (4, '77 N - 03'),
(5, '101 N - 01'), (6, '101 N - 02'), (7, '101 N - 03'), (8, '101 N - 04'), (9, '5 N - 06');

-- 7. TIPO DE CONTRATO
INSERT INTO cat_tipo_contrato (id, nombre) VALUES
(1, 'PU'), (2, 'SG'), (3, 'MIXTO');

-- 8. ESTATUS DETALLE (fusionado con antiguo cat_estado_accion)
INSERT INTO cat_estatus_detalle (id, nombre) VALUES
(1, 'PENDIENTE'),
(2, 'FIRMADO'),
(3, 'DEVUELTO PARA CORRECCIÓN'),
(4, 'DEVUELTO SIN FIRMA'),
(5, 'SE ENTREGA CON LA FIRMA'),
(6, 'SE ENTREGA CON LA MODIFICACIÓN'),
(7, 'SE RECIBE PARA LA FIRMA'),
(8, 'SE DEVUELVE CON LA FIRMA'),
(9, 'SE RECIBE CON LA FIRMA'),
(10, 'SE ENTREGA PARA LA FIRMA');

-- 9. RESULTADOS DEL PROCESO
INSERT INTO cat_resultado_proceso (id, nombre) VALUES
(1, 'ADJUDICADO'),
(2, 'DESIERTO 113 # 1'), (3, 'DESIERTO 113 # 2'), (4, 'DESIERTO 113 # 3'),
(5, 'DESIERTO 113 # 4'), (6, 'DESIERTO 113 # 5'), (7, 'DAR POR TERMINADO');

-- 10. EMPRESAS ADJUDICADAS
INSERT INTO cat_empresas (id, nombre) VALUES
(1, 'PRODUCTORA Y DISTRIBUIDORA VENEZOLANA DE ALIMENTOS, S.A (PDVAL)'),
(2, 'TRANSPORTE ROJAS GARCÍA,C.A.'),
(3, 'CRANE & HEAVY SERVICE DE VENEZUELA'),
(4, 'AGROPECUARIA LA ROSALIERA'),
(5, 'SERVICIOS Y SUMINISTROS KAMULY K&M C.A'),
(6, 'IMSUPETROL, C.A'),
(7, 'CORPORACIÓN SAN REMO, C.A'),
(8, 'INVERSIONES ROYPA, S.A'),
(9, 'CONCRELAND, C.A'),
(10, 'SERVICIOS Y SUMINISTROS DAVNA, C.A.'),
(11, 'METALMECANICA CONTRERAS, C.A'),
(12, 'SERVICIOS Y TRANSPORTE LOS 2 HERMANOS, C.A'),
(13, 'POWERLINE CONSTRUCCIONES, C.A');

-- 11. RESPONSABLES
INSERT INTO cat_responsables (id, nombre) VALUES
(1, 'SIN IDENTIFICAR');

-- PRAGMA user_version = 8;

