# Mapeo Excel → Base de Datos

## Reglas generales

| Excel | DB |
|-------|----|
| OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO | `observaciones` |
| OBSERVACIONES | `notas` |

## Orden de columnas por módulo

El orden listado aquí corresponde al de las hojas del archivo `CONTROL DE DOCUMENTOS JUNIO 2026.xlsx` y se usa tanto para la opción "Orden Excel" en formularios como para la exportación predeterminada.

### Expedientes (Control Docs. Presidencia)
```
N° → id_expediente
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
SOLPED → solped
FECHA PRESUPUESTO BASE → fecha_presupuesto_base
PRESUPUESTO BASE ($.) → presupuesto_base_usd
TIPO DE CAMBIO → tipo_cambio
PRESUPUESTO BASE (Bs.) → presupuesto_base_bs
PLAN DE CONTRATACIONES → id_plan
DESCRIPCIÓN DEL PROCESO → descripcion_proceso
Modalidad de Contratación → id_modalidad
ART → id_art
TIPO DE CONTRATO → id_tipo_contrato
N° DE ACTA APERTURA → nro_acta_apertura
CANTIDAD DE FRENTES → cantidad_frentes
N° RESOLUCIÓN JD → nro_resolucion_jd
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
NUMERO DE PROCESO → nro_proceso
RESULTADOS DEL PROCESO → id_resultado
NÚMERO CONTRATO SICAC → nro_contrato_sicac
NÚMERO CONTRATO SAP → nro_contrato_sap
EMPRESA ADJUDICADA → id_empresa
TIEMPO DE EJECUCIÓN → tiempo_ejecucion
MONTO ADJUDICADO (BS.) → monto_adjudicado_bs
MONTO ADJUDICADO (USD) → monto_adjudicado_usd
FECHA FIRMA DEL CONTRATO → fecha_firma_contrato
OBSERVACIONES → notas
```

### Requisición de Materiales
```
N° → id_requisicion
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
DESCRIPCIÓN DE MATERIALES → descripcion_materiales
SERIAL DEL EQUIPO → serial_equipo
PASE SICESMA → pase_sicesma
ESTATUS DETALLE → id_estatus
OBSERVACIONES DE LA ENTREGA → observaciones_entrega
HISTORIAL → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
OBSERVACIONES → notas
```

### Memorándums
```
N° → id_memorandum
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
ASUNTO MEMORÁNDUM / DECISIÓN DE GERENCIA → asunto
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
OBSERVACIONES → notas
```

### Recobros
```
N° → id_recobro
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
ASUNTO MEMORÁNDUM → asunto
FECHA DE INICIO → fecha_inicio
FECHA FINAL → fecha_final
SERVICIOS → servicios
BENEFICIOS → beneficios
(-) NOTA DE DEBITO (REVERSO) → nota_debito_reverso
COSTO DEL SERVICIO $ → costo_servicio_usd
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
OBSERVACIONES → notas
```

### Valuaciones
```
N° → id_valuacion
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
SOLPED → solped
PRESUPUESTO BASE (Bs.) → presupuesto_base_bs
PRESUPUESTO BASE ($.) → presupuesto_base_usd
DESCRIPCIÓN DEL PROCESO → descripcion_proceso
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
NUMERO DE PROCESO → nro_proceso
NÚMERO CONTRATO SICAC → nro_contrato_sicac
NÚMERO CONTRATO SAP → nro_contrato_sap
EMPRESA ADJUDICADA → id_empresa
TIEMPO DE EJECUCIÓN → tiempo_ejecucion
MONTO ADJUDICADO (BS.) → monto_adjudicado_bs
MONTO ADJUDICADO (USD) → monto_adjudicado_usd
PERÍODO VALUACIONES DESDE → periodo_valuacion_desde
PERÍODO VALUACIONES HASTA → periodo_valuacion_hasta
MONTO VALUACIÓN → monto_valuacion
NÚMERO PROFORMA → nro_proforma
OBSERVACIONES → notas
```

### Aprobación JD
```
N° → id_aprobacion_jd
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
SOLPED → solped
FECHA PRESUPUESTO BASE → fecha_presupuesto_base
PRESUPUESTO BASE (Bs.) → presupuesto_base_bs
TIPO DE CAMBIO → tipo_cambio
PRESUPUESTO BASE ($.) → presupuesto_base_usd
PLAN DE CONTRATACIONES → id_plan
DESCRIPCIÓN DEL PROCESO → descripcion_proceso
CANTIDAD DE FRENTES → cantidad_frentes
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
TIEMPO DE EJECUCIÓN → tiempo_ejecucion
OBSERVACIONES → notas
```

### Certificación BDU
```
N° → id_certificacion_bdu
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
PRESUPUESTO BASE TOTAL ($.) → presupuesto_base_total_usd
MONTO ADJUDICADO TOTAL ($.) → monto_adjudicado_total_usd
MONTO DEL CONTRATO → monto_contrato
MONTO EJECUTADO → monto_ejecutado
MONTO PAGADO → monto_pagado
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
OBSERVACIONES → notas
```

### Vacaciones
```
N° → id_vacacion
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
AÑO → anio
CANTIDAD / DIAS → cantidad_dias
DESDE → fecha_desde
HASTA → fecha_hasta
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO → observaciones
FECHA RECIBIDO → fecha_recibido
FECHA DEVUELTO → fecha_devuelto
RECEPTOR → id_receptor
OBSERVACIONES → notas
```

### Reposos Médicos
```
N° → id_reposo_medico
GERENCIA → id_gerencia
SUPERINTENDENCIA → id_superintendencia
EMISOR → id_emisor
DOCUMENTO → id_documento
DÍAS PERÍODO → dias_periodo
DESDE → fecha_desde
HASTA → fecha_hasta
ESTATUS DETALLE → id_estatus
OBSERVACIONES HISTORIAL + HISTORIAL COMPLETO → observaciones
FECHA RECIBIDO → fecha_recibido
OBSERVACIONES → notas
```
