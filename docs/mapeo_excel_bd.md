# Mapeo Excel → Vista (exportación)

## Reglas generales

| Excel | Vista |
|-------|-------|
| OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO | `observaciones` |
| OBSERVACIONES (generales) | `notas` |

> **Nota:** Las vistas SQL renombran columnas mediante alias. Por ejemplo, `g.nombre AS gerencia`. El `OrdenExcel` en `app.go` usa estos nombres de vista, no los de tabla.

## Orden de columnas por módulo

El orden listado corresponde al de las hojas del archivo `CONTROL DE DOCUMENTOS JUNIO 2026.xlsx`.

### Expedientes (Control Docs. Presidencia)
```
N°                          → id_expediente
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
SOLPED                      → solped
FECHA PRESUPUESTO BASE      → fecha_presupuesto_base
PRESUPUESTO BASE ($.)       → presupuesto_base_usd
TIPO DE CAMBIO              → tipo_cambio
PRESUPUESTO BASE (Bs.)      → presupuesto_base_bs
PLAN DE CONTRATACIONES      → plan_contrataciones
DESCRIPCIÓN DEL PROCESO     → descripcion_proceso
Modalidad de Contratación   → modalidad_contratacion
ART                         → art
TIPO DE CONTRATO            → tipo_contrato
N° DE ACTA APERTURA         → nro_acta_apertura
CANTIDAD DE FRENTES         → cantidad_frentes
N° RESOLUCIÓN JD            → nro_resolucion_jd
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL COMPLETO      → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
NUMERO DE PROCESO           → nro_proceso
RESULTADOS DEL PROCESO      → resultados_proceso
NÚMERO CONTRATO SICAC       → nro_contrato_sicac
NÚMERO CONTRATO SAP         → nro_contrato_sap
EMPRESA ADJUDICADA          → empresa_adjudicada
TIEMPO DE EJECUCIÓN         → tiempo_ejecucion
MONTO ADJUDICADO (BS.)      → monto_adjudicado_bs
MONTO ADJUDICADO (USD)      → monto_adjudicado_usd
FECHA FIRMA DEL CONTRATO    → fecha_firma_contrato
OBSERVACIONES               → notas
```

### Requisición de Materiales
```
N°                          → id_requisicion
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
DESCRIPCIÓN DE MATERIALES   → descripcion_materiales
SERIAL DEL EQUIPO           → serial_equipo
PASE SICESMA                → pase_sicesma
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES DE LA ENTREGA → observaciones_entrega
HISTORIAL                   → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
OBSERVACIONES               → notas
```

### Memorándums
```
N°                          → id_memorandum
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
ASUNTO MEMORÁNDUM / DECISIÓN → asunto
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL               → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
OBSERVACIONES               → notas
```

### Recobros
```
N°                          → id_recobro
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
ASUNTO MEMORÁNDUM           → asunto
FECHA DE INICIO             → fecha_inicio
FECHA FINAL                 → fecha_final
SERVICIOS                   → servicios
BENEFICIOS                  → beneficios
(-) NOTA DE DEBITO (REVERSO) → nota_debito_reverso
COSTO DEL SERVICIO $        → costo_servicio_usd
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL               → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
OBSERVACIONES               → notas
```

### Valuaciones
```
N°                          → id_valuacion
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
SOLPED                      → solped
PRESUPUESTO BASE (Bs.)      → presupuesto_base_bs
PRESUPUESTO BASE ($.)       → presupuesto_base_usd
DESCRIPCIÓN DEL PROCESO     → descripcion_proceso
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL COMPLETO      → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
NUMERO DE PROCESO           → nro_proceso
NÚMERO CONTRATO SICAC       → nro_contrato_sicac
NÚMERO CONTRATO SAP         → nro_contrato_sap
EMPRESA ADJUDICADA          → empresa_adjudicada
TIEMPO DE EJECUCIÓN         → tiempo_ejecucion
MONTO ADJUDICADO (BS.)      → monto_adjudicado_bs
MONTO ADJUDICADO (USD)      → monto_adjudicado_usd
PERÍODO VALUACIONES DESDE   → periodo_valuacion_desde
PERÍODO VALUACIONES HASTA   → periodo_valuacion_hasta
MONTO VALUACIÓN             → monto_valuacion
NÚMERO PROFORMA             → nro_proforma
OBSERVACIONES               → notas
```

### Aprobación JD
```
N°                          → id_aprobacion_jd
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
SOLPED                      → solped
FECHA PRESUPUESTO BASE      → fecha_presupuesto_base
PRESUPUESTO BASE (Bs.)      → presupuesto_base_bs
TIPO DE CAMBIO              → tipo_cambio
PRESUPUESTO BASE ($.)       → presupuesto_base_usd
PLAN DE CONTRATACIONES      → plan_contrataciones
DESCRIPCIÓN DEL PROCESO     → descripcion_proceso
CANTIDAD DE FRENTES         → cantidad_frentes
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL COMPLETO      → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
TIEMPO DE EJECUCIÓN         → tiempo_ejecucion
OBSERVACIONES               → notas
```

### Certificación BDU
```
N°                          → id_certificacion_bdu
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
PRESUPUESTO BASE TOTAL ($.) → presupuesto_base_total_usd
MONTO ADJUDICADO TOTAL ($.) → monto_adjudicado_total_usd
MONTO DEL CONTRATO          → monto_contrato
MONTO EJECUTADO             → monto_ejecutado
MONTO PAGADO                → monto_pagado
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL COMPLETO      → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
OBSERVACIONES               → notas
```

### Vacaciones
```
N°                          → id_vacacion
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
AÑO                         → anio
CANTIDAD / DIAS             → cantidad_dias
DESDE                       → fecha_desde
HASTA                       → fecha_hasta
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL COMPLETO      → observaciones
FECHA RECIBIDO              → fecha_recibido
FECHA DEVUELTO              → fecha_devuelto
RECEPTOR                    → receptor
OBSERVACIONES               → notas
```

### Reposos Médicos
```
N°                          → id_reposo_medico
GERENCIA                    → gerencia
SUPERINTENDENCIA            → superintendencia
EMISOR                      → emisor
DOCUMENTO                   → documento
DÍAS PERÍODO                → dias_periodo
DESDE                       → fecha_desde
HASTA                       → fecha_hasta
ESTATUS DETALLE             → estatus_detalle
OBSERVACIONES HISTORIAL
  + HISTORIAL COMPLETO      → observaciones
FECHA RECIBIDO              → fecha_recibido
OBSERVACIONES               → notas
```
