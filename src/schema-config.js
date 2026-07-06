// ================================================================
// schema-config.js — Configuración específica del schema
// ================================================================
// Centraliza todo lo que depende del schema actual (Tablas8.sql):
// catálogos, columnas, formato de observaciones, colores de estatus.
// También define constantes de toda la app (CONFIG, MSG, etc.)
// ================================================================

// ---- CONSTANTES GLOBALES -----------------------------------------

const CONFIG = {
    MAX_FILE_SIZE_BYTES: 100 * 1024 * 1024,
    MAX_FILE_SIZE_MB: 100,
    BYTES_PER_MB: 1048576,
    AUTOSAVE_INTERVAL_MS: 30000,
    AUTOSAVE_ENABLED: true,
};

const DEBUG = {
    isEnabled: false,
    log: function(...args) { if (DEBUG.isEnabled) console.log(...args); },
    error: function(...args) { if (DEBUG.isEnabled) console.error(...args); },
};

const MSG = {
    ERROR_NO_DB: 'Primero carga un archivo .db',
    ERROR_TIPO_ARCHIVO: 'Solo se aceptan archivos .db o .sqlite',
    ERROR_TAMANO: (sizeMB) =>
        `El archivo es demasiado grande (${sizeMB.toFixed(2)} MB). Máximo: ${CONFIG.MAX_FILE_SIZE_MB} MB`,
    ERROR_LECTURA: err => 'Error al abrir archivo: ' + (err?.message || err),
    ERROR_CONSULTA: err => 'Error en consulta: ' + (err?.message || err),
    ERROR_GUARDAR: err => 'Error al guardar: ' + (err?.message || err),
    ERROR_ELIMINAR: err => 'Error al eliminar: ' + (err?.message || err),
    ERROR_NO_EXPEDIENTE: 'No se encontró el expediente.',
    ERROR_ID_INVALIDO: 'ID de expediente inválido',
    ERROR_NO_BD_VALIDA: 'El archivo no parece ser una base de datos SQLite válida.',
    ERROR_NO_REABRIR: err => 'No se pudo abrir la base de datos: ' + (err?.message || err),
    ERROR_ABRIR_BD: err => 'Error al abrir BD: ' + (err?.message || err),
    ERROR_SCHEMA_VERSION: (actual, esperada) =>
        `Schema desactualizado: versión ${actual}, esperada ${esperada}. Resincroniza la base de datos.`,
    NOMBRE_OBLIGATORIO: 'El nombre es obligatorio.',
    EXITO_ACTUALIZADO: 'Expediente actualizado correctamente. El trigger de auditoría ha registrado los cambios en el historial.',
    EXITO_CREADO: 'Expediente creado correctamente.',
    EXITO_ELIMINADO: 'Expediente eliminado correctamente.',
    FECHA_DEVUELTO_INVALIDA: 'Fecha Devuelto no puede ser anterior a Fecha Recibido.',
};

const STORAGE_KEYS = {
    FRECUENTES: 'sidebarFrecuentes',
    RECIENTES: 'recientes',
    SIDEBAR_VISIBLE: 'sidebarVisible',
    BACKUP_MAX_COPIES: 'BACKUP_MAX_COPIES',
};

const SELECTORS = {
    TABLA_CUERPO: 'tabla-cuerpo',
    FORM_MODAL: 'form-modal',
    SEARCH: 'search',
    SORT_ORDER: 'sort-order',
    SIDEBAR: 'sidebar',
    SIDEBAR_TOGGLE: 'sidebar-toggle',
    BODY: 'body',
    FILE_INPUT: 'dbfile',
    MENU_RECIENTES: 'menu-recientes',
    MODAL_RUTA: 'modal-ruta',
    RUTA_CONTENIDO: 'ruta-contenido',
    MODAL_PENDIENTES: 'modal-pendientes',
    PENDIENTES_CONTENIDO: 'pendientes-contenido',
    MODAL_HISTORIAL: 'modal-historial',
    HISTORIAL_CONTENIDO: 'historial-contenido',
    MODAL_CATALOGO: 'modal-agregar-catalogo',
    AC_NOMBRE: 'ac-nombre',
    AC_SELECT_ID: 'ac-select-id',
    AC_CATALOG_KEY: 'ac-catalog-key',
    AC_TABLA: 'ac-tabla',
    AC_CAMPO_LABEL: 'ac-campo-label',
    AC_EXTRA_FIELDS: 'ac-extra-fields',
    AC_EXTRA_LABEL: 'ac-extra-label',
    AC_EXTRA_VAL: 'ac-extra-val',
    F_OBSERVACIONES: 'f-observaciones',
    GUARDAR_BD_BTN: 'btn-guardar-bd',
    BTN_VACUUM: 'btn-vacuum',
    MODAL_ERROR: 'modal-error-critico',
    ERROR_CONTENIDO: 'error-critico-contenido',
    BTN_DESCARGAR_BD: 'btn-descargar-bd-error',
    ESTADO_BD: 'estado-bd',
};

const MSG_EXTRA = {
    VACUUM_INICIADO: 'Optimizando base de datos...',
    VACUUM_COMPLETADO: (antes, despues) =>
        `Base de datos optimizada: ${antes} → ${despues} MB (${((1 - despues/antes) * 100).toFixed(1)}% reducción)`,
    VACUUM_ERROR: err => 'Error al optimizar BD: ' + (err?.message || err),
    ERROR_CRITICO: 'Ocurrió un error inesperado. Puedes descargar la BD actual para rescatar tus datos antes de recargar.',
    PROMESA_RECHAZADA: 'Una operación asíncrona falló inesperadamente.',
    BD_DESCARGADA: 'BD descargada. Recarga la aplicación y continúa.',
};

const BACKUP = {
    MAX_COPIES: 5,
    SUFFIX: '.bak.',
};

// ---- SCHEMA CONFIG ------------------------------------------------

const SCHEMA_CONFIG = {

    // Vista principal usada como fuente de datos de la tabla
    viewName: 'vw_reporte_excel_contrataciones',

    // Columnas de la tabla expedientes (orden = orden en INSERT/UPDATE)
    columnas: [
        'solped', 'id_gerencia', 'id_superintendencia', 'id_emisor',
        'id_documento', 'fecha_presupuesto_base', 'presupuesto_base_usd',
        'tipo_cambio', 'presupuesto_base_bs', 'id_plan', 'descripcion_proceso',
        'id_modalidad', 'id_art', 'id_tipo_contrato', 'nro_acta_apertura',
        'cantidad_frentes', 'nro_resolucion_jd', 'id_estatus',
        'fecha_recibido', 'fecha_devuelto', 'id_receptor', 'nro_proceso',
        'id_resultado', 'nro_contrato_sicac', 'nro_contrato_sap', 'id_empresa',
        'tiempo_ejecucion', 'monto_adjudicado_bs', 'monto_adjudicado_usd',
        'fecha_firma_contrato', 'observaciones', 'notas'
    ],

    // Campos que se editan con frecuencia (reciben indicador visual)
    camposEdicionFrecuente: [
        'id_tipo_contrato', 'id_emisor', 'id_receptor', 'id_gerencia',
        'id_superintendencia', 'id_documento', 'id_estatus',
        'fecha_recibido', 'fecha_devuelto', 'observaciones'
    ],

    // Mapeo de select → catálogo en BD
    catalogoPorSelect: {
        'f-id_gerencia':         { key: 'gerencia',           tabla: 'cat_gerencia',         cols: 'id, nombre' },
        'f-id_superintendencia': { key: 'superintendencia',   tabla: 'cat_superintendencia', cols: 'id, nombre, id_gerencia', extra: { col: 'id_gerencia', label: 'ID Gerencia', selectId: 'f-id_gerencia' } },
        'f-id_documento':        { key: 'documento',          tabla: 'cat_documento',        cols: 'id, nombre' },
        'f-id_plan':             { key: 'plan_contratacion',  tabla: 'cat_plan_contratacion', cols: 'id, nombre' },
        'f-id_modalidad':        { key: 'modalidad',          tabla: 'cat_modalidad',        cols: 'id, nombre' },
        'f-id_art':              { key: 'art',                tabla: 'cat_art',              cols: 'id, nombre' },
        'f-id_tipo_contrato':    { key: 'tipo_contrato',      tabla: 'cat_tipo_contrato',    cols: 'id, nombre' },
        'f-id_estatus':          { key: 'estatus_detalle',    tabla: 'cat_estatus_detalle',  cols: 'id, nombre' },
        'f-id_resultado':        { key: 'resultado_proceso',  tabla: 'cat_resultado_proceso', cols: 'id, nombre' },
        'f-id_empresa':          { key: 'empresas',           tabla: 'cat_empresas',         cols: 'id, nombre' },
        'f-id_emisor':           { key: 'responsables',       tabla: 'cat_responsables',     cols: 'id, nombre' },
        'f-id_receptor':         { key: 'responsables',       tabla: 'cat_responsables',     cols: 'id, nombre' }
    },

    // Genera la línea de observación auto-generada (estatus - documento - fechas)
    generarObservacion: function(estatus, documento, fechaRecibido, fechaDevuelto) {
        const partes = [estatus || 'PENDIENTE', documento || 'N/A'];
        if (fechaRecibido) partes.push('*Fecha recibido* ' + fechaRecibido);
        if (fechaDevuelto) partes.push('*Fecha devuelto* ' + fechaDevuelto);
        return partes.join(' - ');
    },

    // Extrae texto libre del usuario restando las partes auto-generadas
    extraerTextoLibre: function(currentValue, autoLine) {
        if (!currentValue) return '';
        let remaining = currentValue;
        const autoParts = autoLine.split(' - ');
        for (const part of autoParts) {
            const idx = remaining.indexOf(part);
            if (idx !== -1) {
                remaining = remaining.slice(0, idx) + remaining.slice(idx + part.length);
            }
        }
        remaining = remaining.replace(/\*Fecha (recibido|devuelto)\*(\s+\d{4}-\d{2}-\d{2})?/g, '');
        remaining = remaining.replace(/ - /g, ' ').replace(/\s+/g, ' ').trim();
        remaining = remaining.replace(/^-+/, '').replace(/-+$/, '').trim();
        return remaining;
    },

    // Retorna clases Tailwind según el estatus del expediente
    estatusClass: function(estatus) {
        if (!estatus) return 'bg-yellow-500/20 text-yellow-400';
        switch (estatus.toUpperCase()) {
            case 'FIRMADO': return 'bg-emerald-500/20 text-emerald-400';
            case 'PENDIENTE': return 'bg-yellow-500/20 text-yellow-400';
            case 'DEVUELTO PARA CORRECCIÓN': return 'bg-orange-500/20 text-orange-400';
            case 'DEVUELTO SIN FIRMA': return 'bg-red-500/20 text-red-400';
            default: return 'bg-gray-500/20 text-gray-400';
        }
    },

    // Determina si un estatus se considera "FIRMADO"
    esEstatusFirmado: function(estatus) {
        return estatus && estatus.toUpperCase() === 'FIRMADO';
    },

    // Orden de campos en el formulario según aparecen en el Excel
    ordenExcel: [
        'f-solped', 'f-id_gerencia', 'f-id_superintendencia', 'f-id_emisor',
        'f-id_documento', 'f-fecha_presupuesto_base', 'f-presupuesto_base_usd',
        'f-tipo_cambio', 'f-presupuesto_base_bs', 'f-id_plan',
        'f-descripcion_proceso', 'f-id_modalidad', 'f-id_art', 'f-id_tipo_contrato',
        'f-nro_acta_apertura', 'f-cantidad_frentes', 'f-nro_resolucion_jd',
        'f-id_estatus', 'f-fecha_recibido', 'f-fecha_devuelto', 'f-id_receptor',
        'f-nro_proceso', 'f-id_resultado', 'f-nro_contrato_sicac',
        'f-nro_contrato_sap', 'f-id_empresa', 'f-tiempo_ejecucion',
        'f-monto_adjudicado_bs', 'f-monto_adjudicado_usd', 'f-fecha_firma_contrato',
        'f-observaciones', 'f-notas'
    ],

    // Versión del schema para validación PRAGMA user_version
    VERSION: 8,

    // Queries SQL centralizadas (SPOT: cambiar aquí si cambia el schema)
    queries: {
        reporteExcel: `SELECT * FROM vw_reporte_excel_contrataciones`,
        rutaProcesos: `SELECT e.id_expediente, e.solped, e.descripcion_proceso,
            e.emisor, e.receptor, e.documento, e.estatus_detalle,
            e.fecha_recibido, e.fecha_devuelto, e.nro_proceso
            FROM vw_reporte_excel_contrataciones e
            ORDER BY e.estatus_detalle, e.id_expediente DESC`,
        documentosPendientes: `SELECT e.id_expediente, e.solped, e.descripcion_proceso,
            e.emisor, e.receptor, e.documento, e.estatus_detalle,
            e.fecha_recibido, e.fecha_devuelto, e.nro_proceso, e.empresa_adjudicada
            FROM vw_reporte_excel_contrataciones e
            WHERE e.estatus_detalle IS NOT NULL AND UPPER(e.estatus_detalle) != 'FIRMADO'
            ORDER BY e.estatus_detalle, e.id_expediente DESC`,
        expedientesSelect: `SELECT * FROM vw_reporte_excel_contrataciones`,
        expedientePorId: `SELECT * FROM vw_reporte_excel_contrataciones WHERE id_expediente = ?`,
    }
};
