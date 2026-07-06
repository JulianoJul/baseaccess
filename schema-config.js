// ================================================================
// schema-config.js — Configuración específica del schema
// ================================================================
// Centraliza todo lo que depende del schema actual (Tablas8.sql):
// catálogos, columnas, formato de observaciones, colores de estatus.
// ================================================================

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
    generarObservacion: function() {
        const obtenerTexto = function(id) {
            const el = document.getElementById(id);
            if (!el || !el.options) return '';
            return el.options[el.selectedIndex] ? el.options[el.selectedIndex].text : '';
        };
        const estatus = obtenerTexto('f-id_estatus') || 'PENDIENTE';
        const doc = obtenerTexto('f-id_documento') || 'N/A';
        const recibido = document.getElementById('f-fecha_recibido')?.value || '';
        const devuelto = document.getElementById('f-fecha_devuelto')?.value || '';
        const partes = [estatus, doc];
        if (recibido) partes.push('*Fecha recibido* ' + recibido);
        if (devuelto) partes.push('*Fecha devuelto* ' + devuelto);
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
    }
};
