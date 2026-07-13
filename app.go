package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const DefaultBackupMaxCopies = 2

var backupMaxCopies = DefaultBackupMaxCopies

type ModuloConfig struct {
	Nombre         string
	Tabla          string
	Vista          string
	IDColumna      string
	HistorialTabla string
	Columnas       []string
	QueryHistorial string
}

var Modulos = map[string]ModuloConfig{
	"expedientes": {
		Nombre:         "Control Docs. Presidencia",
		Tabla:          "expedientes",
		Vista:          "vw_reporte_excel_contrataciones",
		IDColumna:      "id_expediente",
		HistorialTabla: "historial_movimientos",
		Columnas: []string{
			"solped", "id_gerencia", "id_superintendencia", "id_emisor",
			"id_documento", "fecha_presupuesto_base", "presupuesto_base_usd",
			"tipo_cambio", "presupuesto_base_bs", "id_plan", "descripcion_proceso",
			"id_modalidad", "id_art", "id_tipo_contrato", "nro_acta_apertura",
			"cantidad_frentes", "nro_resolucion_jd", "id_estatus",
			"fecha_recibido", "fecha_devuelto", "id_receptor", "nro_proceso",
			"id_resultado", "nro_contrato_sicac", "nro_contrato_sap", "id_empresa",
			"tiempo_ejecucion", "monto_adjudicado_bs", "monto_adjudicado_usd",
			"fecha_firma_contrato", "observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(tc.nombre, '-') AS tipo_contrato, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(d.nombre, '-') AS documento, COALESCE(em.nombre, '-') AS emisor, COALESCE(rec.nombre, '-') AS receptor, COALESCE(ed.nombre, '-') AS estatus, COALESCE(h.fecha_recibido, '-') AS fecha_recibido, COALESCE(h.fecha_devuelto, '-') AS fecha_devuelto, COALESCE(h.nro_proceso, '-') AS nro_proceso, h.presupuesto_base_usd, h.tipo_cambio, h.monto_adjudicado_usd, COALESCE(rp.nombre, '-') AS resultado, COALESCE(emp.nombre, '-') AS empresa, COALESCE(h.tiempo_ejecucion, '-') AS tiempo_ejecucion, COALESCE(h.fecha_firma_contrato, '-') AS fecha_firma_contrato, COALESCE(h.observaciones, '') AS observaciones, COALESCE(h.notas, '') AS notas FROM historial_movimientos h LEFT JOIN cat_tipo_contrato tc ON h.id_tipo_contrato = tc.id LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_documento d ON h.id_documento = d.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_resultado_proceso rp ON h.id_resultado = rp.id LEFT JOIN cat_empresas emp ON h.id_empresa = emp.id WHERE h.id_expediente = ? ORDER BY h.id_movimiento DESC`,
	},
	"requisiciones": {
		Nombre:         "Requisición de Materiales",
		Tabla:          "req_materiales",
		Vista:          "vw_reporte_req_materiales",
		IDColumna:      "id_requisicion",
		HistorialTabla: "hist_req_materiales",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"descripcion_materiales", "serial_equipo", "pase_sicesma", "id_estatus",
			"observaciones_entrega", "fecha_recibido", "fecha_devuelto", "id_receptor",
			"observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, COALESCE(d.nombre, '-') AS documento, h.descripcion_materiales, h.serial_equipo, h.pase_sicesma, COALESCE(ed.nombre, '-') AS estatus, h.observaciones_entrega, h.fecha_recibido, h.fecha_devuelto, COALESCE(rec.nombre, '-') AS receptor, h.observaciones, h.notas FROM hist_req_materiales h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_documento d ON h.id_documento = d.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id WHERE h.id_requisicion = ? ORDER BY h.id_movimiento DESC`,
	},
	"memorandums": {
		Nombre:         "Memorándums",
		Tabla:          "memorandums",
		Vista:          "vw_reporte_memorandums",
		IDColumna:      "id_memorandum",
		HistorialTabla: "hist_memorandums",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "documento",
			"asunto", "id_estatus", "fecha_recibido", "fecha_devuelto",
			"id_receptor", "observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, h.documento, h.asunto, COALESCE(ed.nombre, '-') AS estatus, h.fecha_recibido, h.fecha_devuelto, COALESCE(rec.nombre, '-') AS receptor, h.observaciones, h.notas FROM hist_memorandums h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id WHERE h.id_memorandum = ? ORDER BY h.id_movimiento DESC`,
	},
	"recobros": {
		Nombre:         "Recobros",
		Tabla:          "recobros",
		Vista:          "vw_reporte_recobros",
		IDColumna:      "id_recobro",
		HistorialTabla: "hist_recobros",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "documento",
			"asunto", "fecha_inicio", "fecha_final", "servicios", "beneficios",
			"nota_debito_reverso", "costo_servicio_usd", "id_estatus",
			"fecha_recibido", "fecha_devuelto", "id_receptor", "observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, h.documento, h.asunto, h.fecha_inicio, h.fecha_final, h.servicios, h.beneficios, h.nota_debito_reverso, h.costo_servicio_usd, COALESCE(ed.nombre, '-') AS estatus, h.fecha_recibido, h.fecha_devuelto, COALESCE(rec.nombre, '-') AS receptor, h.observaciones, h.notas FROM hist_recobros h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id WHERE h.id_recobro = ? ORDER BY h.id_movimiento DESC`,
	},
	"valuaciones": {
		Nombre:         "Valuaciones",
		Tabla:          "valuaciones",
		Vista:          "vw_reporte_valuaciones",
		IDColumna:      "id_valuacion",
		HistorialTabla: "hist_valuaciones",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "documento",
			"solped", "presupuesto_base_bs", "presupuesto_base_usd", "descripcion_proceso",
			"id_estatus", "fecha_recibido", "fecha_devuelto", "id_receptor", "nro_proceso",
			"nro_contrato_sicac", "nro_contrato_sap", "id_empresa", "tiempo_ejecucion",
			"monto_adjudicado_bs", "monto_adjudicado_usd", "periodo_valuacion_desde",
			"periodo_valuacion_hasta", "monto_valuacion", "nro_proforma", "observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, h.documento, h.solped, h.presupuesto_base_bs, h.presupuesto_base_usd, h.descripcion_proceso, COALESCE(ed.nombre, '-') AS estatus, h.fecha_recibido, h.fecha_devuelto, COALESCE(rec.nombre, '-') AS receptor, h.nro_proceso, h.nro_contrato_sicac, h.nro_contrato_sap, COALESCE(emp.nombre, '-') AS empresa, h.tiempo_ejecucion, h.monto_adjudicado_bs, h.monto_adjudicado_usd, h.periodo_valuacion_desde, h.periodo_valuacion_hasta, h.monto_valuacion, h.nro_proforma, h.observaciones, h.notas FROM hist_valuaciones h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id LEFT JOIN cat_empresas emp ON h.id_empresa = emp.id WHERE h.id_valuacion = ? ORDER BY h.id_movimiento DESC`,
	},
	"aprobacion_jd": {
		Nombre:         "Aprobación JD",
		Tabla:          "aprobacion_jd",
		Vista:          "vw_reporte_aprobacion_jd",
		IDColumna:      "id_aprobacion_jd",
		HistorialTabla: "hist_aprobacion_jd",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"solped", "fecha_presupuesto_base", "presupuesto_base_bs", "tipo_cambio",
			"presupuesto_base_usd", "id_plan", "descripcion_proceso", "cantidad_frentes",
			"id_estatus", "fecha_recibido", "fecha_devuelto", "id_receptor", "tiempo_ejecucion",
			"observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, COALESCE(d.nombre, '-') AS documento, h.solped, h.fecha_presupuesto_base, h.presupuesto_base_bs, h.tipo_cambio, h.presupuesto_base_usd, COALESCE(p.nombre, '-') AS plan_contrataciones, h.descripcion_proceso, h.cantidad_frentes, COALESCE(ed.nombre, '-') AS estatus, h.fecha_recibido, h.fecha_devuelto, COALESCE(rec.nombre, '-') AS receptor, h.tiempo_ejecucion, h.observaciones, h.notas FROM hist_aprobacion_jd h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_documento d ON h.id_documento = d.id LEFT JOIN cat_plan_contratacion p ON h.id_plan = p.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id WHERE h.id_aprobacion_jd = ? ORDER BY h.id_movimiento DESC`,
	},
	"certificacion_bdu": {
		Nombre:         "Certificación BDU",
		Tabla:          "certificacion_bdu",
		Vista:          "vw_reporte_certificacion_bdu",
		IDColumna:      "id_certificacion_bdu",
		HistorialTabla: "hist_certificacion_bdu",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"presupuesto_base_total_usd", "monto_adjudicado_total_usd", "monto_contrato",
			"monto_ejecutado", "monto_pagado", "id_estatus", "fecha_recibido",
			"fecha_devuelto", "id_receptor", "observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, COALESCE(d.nombre, '-') AS documento, h.presupuesto_base_total_usd, h.monto_adjudicado_total_usd, h.monto_contrato, h.monto_ejecutado, h.monto_pagado, COALESCE(ed.nombre, '-') AS estatus, h.fecha_recibido, h.fecha_devuelto, COALESCE(rec.nombre, '-') AS receptor, h.observaciones, h.notas FROM hist_certificacion_bdu h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_documento d ON h.id_documento = d.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id WHERE h.id_certificacion_bdu = ? ORDER BY h.id_movimiento DESC`,
	},
	"vacaciones": {
		Nombre:         "Vacaciones",
		Tabla:          "vacaciones",
		Vista:          "vw_reporte_vacaciones",
		IDColumna:      "id_vacacion",
		HistorialTabla: "hist_vacaciones",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "documento",
			"anio", "cantidad_dias", "fecha_desde", "fecha_hasta", "id_estatus",
			"fecha_recibido", "fecha_devuelto", "id_receptor", "observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, h.documento, h.anio, h.cantidad_dias, h.fecha_desde, h.fecha_hasta, COALESCE(ed.nombre, '-') AS estatus, h.fecha_recibido, h.fecha_devuelto, COALESCE(rec.nombre, '-') AS receptor, h.observaciones, h.notas FROM hist_vacaciones h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id WHERE h.id_vacacion = ? ORDER BY h.id_movimiento DESC`,
	},
	"reposos_medicos": {
		Nombre:         "Reposos Médicos",
		Tabla:          "reposos_medicos",
		Vista:          "vw_reporte_reposos_medicos",
		IDColumna:      "id_reposo_medico",
		HistorialTabla: "hist_reposos_medicos",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "documento",
			"dias_periodo", "fecha_desde", "fecha_hasta", "id_estatus",
			"fecha_recibido", "observaciones", "notas",
		},
		QueryHistorial: `SELECT h.id_movimiento, COALESCE(g.nombre, '-') AS gerencia, COALESCE(s.nombre, '-') AS superintendencia, COALESCE(em.nombre, '-') AS emisor, h.documento, h.dias_periodo, h.fecha_desde, h.fecha_hasta, COALESCE(ed.nombre, '-') AS estatus, h.fecha_recibido, h.observaciones, h.notas FROM hist_reposos_medicos h LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id LEFT JOIN cat_responsables em ON h.id_emisor = em.id LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id WHERE h.id_reposo_medico = ? ORDER BY h.id_movimiento DESC`,
	},
}

const (
	queryObtenerRutaProcesos   = `SELECT e.id_expediente, e.solped, e.descripcion_proceso, e.emisor, e.receptor, e.documento, e.estatus_detalle, e.fecha_recibido, e.fecha_devuelto, e.nro_proceso FROM vw_reporte_excel_contrataciones e ORDER BY e.estatus_detalle, e.id_expediente DESC`
	queryObtenerDocsPendientes = `SELECT e.id_expediente, e.solped, e.descripcion_proceso, e.emisor, e.receptor, e.documento, e.estatus_detalle, e.fecha_recibido, e.fecha_devuelto, e.nro_proceso, e.empresa_adjudicada FROM vw_reporte_excel_contrataciones e WHERE e.estatus_detalle IS NOT NULL AND UPPER(e.estatus_detalle) != 'FIRMADO' ORDER BY e.estatus_detalle, e.id_expediente DESC`
	queryVacuum                 = `VACUUM`
)

var catalogosValidos = map[string]string{
	"gerencia":          "cat_gerencia",
	"superintendencia":  "cat_superintendencia",
	"documento":         "cat_documento",
	"plan_contratacion": "cat_plan_contratacion",
	"modalidad":         "cat_modalidad",
	"art":               "cat_art",
	"tipo_contrato":     "cat_tipo_contrato",
	"estatus_detalle":   "cat_estatus_detalle",
	"resultado_proceso": "cat_resultado_proceso",
	"empresas":          "cat_empresas",
	"responsables":      "cat_responsables",
}

var columnasExtraValidas = map[string]bool{
	"id_gerencia": true,
}

var columnasOrdenValidas = map[string]bool{
	"id_expediente":       true,
	"fecha_creacion":      true,
	"fecha_actualizacion": true,
	"solped":              true,
	"gerencia":            true,
	"estatus_detalle":     true,
}


type App struct {
	ctx    context.Context
	db     *sql.DB
	dbPath string
	mu     sync.Mutex
}

type Row map[string]interface{}

func NewApp() *App { return &App{} }

func (a *App) Startup(ctx context.Context) { a.ctx = ctx }

func (a *App) SetBackupMaxCopies(n int) {
	if n < 1 {
		n = 1
	}
	if n > 20 {
		n = 20
	}
	backupMaxCopies = n
}

func (a *App) GetBackupMaxCopies() int { return backupMaxCopies }

func (a *App) AbrirDialogoBD() (string, error) {
	return wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Seleccionar base de datos",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "SQLite DB", Pattern: "*.db;*.sqlite"},
		},
	})
}

func (a *App) GuardarDialogoBD(nombreDefault string) (string, error) {
	return wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Guardar copia de base de datos",
		DefaultFilename: nombreDefault,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "SQLite DB", Pattern: "*.db;*.sqlite"},
		},
	})
}

func (a *App) AbrirBaseDatos(filePath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db != nil {
		a.db.Close()
	}

	db, err := sql.Open("sqlite3", filePath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("no se pudo abrir BD: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("no se pudo conectar a BD: %w", err)
	}

	a.db = db
	a.dbPath = filePath
	return nil
}

func (a *App) CerrarBaseDatos() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}
	a.dbPath = ""
}

func (a *App) crearBackup() error {
	if a.dbPath == "" {
		return nil
	}
	dir := filepath.Dir(a.dbPath)
	base := filepath.Base(a.dbPath)
	sep := string(filepath.Separator)

	oldest := dir + sep + base + ".bak." + strconv.Itoa(backupMaxCopies)
	os.Remove(oldest)

	for i := backupMaxCopies - 1; i >= 1; i-- {
		src := dir + sep + base + ".bak." + strconv.Itoa(i)
		dst := dir + sep + base + ".bak." + strconv.Itoa(i+1)
		if _, err := os.Stat(src); err == nil {
			os.Rename(src, dst)
		}
	}

	srcFile, err := os.Open(a.dbPath)
	if err != nil {
		return fmt.Errorf("error abriendo BD para backup: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dir + sep + base + ".bak.1")
	if err != nil {
		return fmt.Errorf("error creando backup: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copiando backup: %w", err)
	}
	return nil
}

func (a *App) DescargarBD(destPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.dbPath == "" {
		return fmt.Errorf("no hay base de datos abierta")
	}
	srcFile, err := os.Open(a.dbPath)
	if err != nil {
		return fmt.Errorf("error abriendo BD: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creando archivo destino: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copiando BD: %w", err)
	}
	return nil
}

func (a *App) queryRows(query string, args ...interface{}) ([]Row, error) {
	if a.db == nil {
		return nil, fmt.Errorf("no hay base de datos abierta")
	}
	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []Row
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(Row)
		for i, col := range cols {
			val := values[i]
			switch v := val.(type) {
			case []byte:
				row[col] = string(v)
			case time.Time:
				if v.IsZero() {
					row[col] = ""
				} else {
					row[col] = v.Format("2006-01-02")
				}
			default:
				row[col] = val
			}
		}
		results = append(results, row)
	}
	return results, nil
}

func (a *App) exec(query string, args ...interface{}) (sql.Result, error) {
	if a.db == nil {
		return nil, fmt.Errorf("no hay base de datos abierta")
	}
	return a.db.Exec(query, args...)
}

func sanitizarOrden(orden string, def string) string {
	partes := strings.Fields(orden)
	if len(partes) == 0 {
		return def + " DESC"
	}
	col := partes[0]
	dir := "DESC"
	if len(partes) > 1 {
		d := strings.ToUpper(partes[1])
		if d == "ASC" || d == "DESC" {
			dir = d
		}
	}
	colClean := ""
	for _, r := range col {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' {
			colClean += string(r)
		}
	}
	if colClean == "" {
		return def + " DESC"
	}
	return colClean + " " + dir
}

func (a *App) ObtenerFilas(moduloKey string, orden string) ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cfg, ok := Modulos[moduloKey]
	if !ok {
		return nil, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	orden = sanitizarOrden(orden, cfg.IDColumna)
	q := `SELECT * FROM ` + cfg.Vista + ` ORDER BY ` + orden
	return a.queryRows(q)
}

func (a *App) ObtenerFilaPorId(moduloKey string, id int) (Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cfg, ok := Modulos[moduloKey]
	if !ok {
		return nil, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	q := `SELECT * FROM ` + cfg.Vista + ` WHERE ` + cfg.IDColumna + ` = ?`
	rows, err := a.queryRows(q, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("registro %d no encontrado en %s", id, moduloKey)
	}
	return rows[0], nil
}

func (a *App) GuardarFila(moduloKey string, data map[string]interface{}) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db == nil {
		return 0, fmt.Errorf("no hay base de datos abierta")
	}

	cfg, ok := Modulos[moduloKey]
	if !ok {
		return 0, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}

	idVal, ok := data[cfg.IDColumna]
	delete(data, cfg.IDColumna)

	var id float64
	if ok && idVal != nil {
		switch v := idVal.(type) {
		case float64:
			id = v
		case int:
			id = float64(v)
		case int64:
			id = float64(v)
		case string:
			if v != "" {
				parsed, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return 0, fmt.Errorf("id inválido: %v", idVal)
				}
				id = parsed
			}
		default:
			return 0, fmt.Errorf("id tipo no soportado: %T", idVal)
		}
	}

	vals := make([]interface{}, len(cfg.Columnas))
	for i, col := range cfg.Columnas {
		v, ok := data[col]
		if !ok || v == nil {
			vals[i] = nil
		} else {
			vals[i] = v
		}
	}

	if id > 0 {
		sets := make([]string, len(cfg.Columnas))
		for i, col := range cfg.Columnas {
			sets[i] = col + " = ?"
		}
		q := `UPDATE ` + cfg.Tabla + ` SET ` + strings.Join(sets, ", ") +
			`, fecha_actualizacion = CURRENT_DATE WHERE ` + cfg.IDColumna + ` = ?`
		res, err := a.exec(q, append(vals, id)...)
		if err != nil {
			return 0, fmt.Errorf("error al actualizar: %w", err)
		}
		return res.LastInsertId()
	}

	placeholders := make([]string, len(cfg.Columnas))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	q := `INSERT INTO ` + cfg.Tabla + ` (` + strings.Join(cfg.Columnas, ", ") +
		`) VALUES (` + strings.Join(placeholders, ", ") + `)`
	res, err := a.exec(q, vals...)
	if err != nil {
		return 0, fmt.Errorf("error al insertar: %w", err)
	}
	return res.LastInsertId()
}

func (a *App) EliminarFila(moduloKey string, id int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db == nil {
		return fmt.Errorf("no hay base de datos abierta")
	}

	cfg, ok := Modulos[moduloKey]
	if !ok {
		return fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}

	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if cfg.HistorialTabla != "" {
		qHist := `DELETE FROM ` + cfg.HistorialTabla + ` WHERE ` + cfg.IDColumna + ` = ?`
		if _, err = tx.Exec(qHist, id); err != nil {
			return err
		}
	}

	qDel := `DELETE FROM ` + cfg.Tabla + ` WHERE ` + cfg.IDColumna + ` = ?`
	if _, err = tx.Exec(qDel, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (a *App) ObtenerHistorialFila(moduloKey string, id int) ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	cfg, ok := Modulos[moduloKey]
	if !ok {
		return nil, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	if cfg.QueryHistorial == "" {
		return nil, fmt.Errorf("modulo %s no tiene soporte para historial", moduloKey)
	}

	return a.queryRows(cfg.QueryHistorial, id)
}

func (a *App) ObtenerRutaProcesos() ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.queryRows(queryObtenerRutaProcesos)
}

func (a *App) ObtenerDocumentosPendientes() ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.queryRows(queryObtenerDocsPendientes)
}


type CatalogoItem struct {
	ID         int    `json:"id"`
	Nombre     string `json:"nombre"`
	IDGerencia int    `json:"id_gerencia,omitempty"`
}

func (a *App) ObtenerCatalogos() (map[string][]CatalogoItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db == nil {
		return nil, fmt.Errorf("no hay base de datos abierta")
	}

	result := make(map[string][]CatalogoItem)
	for key, table := range catalogosValidos {
		cols := "id, nombre"
		var extraCol string
		if key == "superintendencia" {
			cols = "id, nombre, id_gerencia"
			extraCol = "id_gerencia"
		}
		rows, err := a.db.Query("SELECT " + cols + " FROM " + table + " ORDER BY nombre")
		if err != nil {
			return nil, err
		}

		var items []CatalogoItem
		for rows.Next() {
			var item CatalogoItem
			var extraVal sql.NullInt64
			scanArgs := []interface{}{&item.ID, &item.Nombre}
			if extraCol != "" {
				scanArgs = append(scanArgs, &extraVal)
			}
			if err := rows.Scan(scanArgs...); err != nil {
				rows.Close()
				return nil, err
			}
			if extraCol != "" && extraVal.Valid {
				item.IDGerencia = int(extraVal.Int64)
			}
			items = append(items, item)
		}
		rows.Close()
		result[key] = items
	}
	return result, nil
}

func (a *App) OptimizarBD() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay base de datos abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.exec(queryVacuum)
	return err
}

func (a *App) GuardarNuevoCatalogo(tabla, nombre string, extra map[string]interface{}) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db == nil {
		return 0, fmt.Errorf("no hay base de datos abierta")
	}

	tablaValida := false
	for _, t := range catalogosValidos {
		if t == tabla {
			tablaValida = true
			break
		}
	}
	if !tablaValida {
		return 0, fmt.Errorf("tabla de catálogo no válida: %s", tabla)
	}

	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}

	cols := []string{"nombre"}
	args := []interface{}{nombre}
	placeholders := []string{"?"}

	if v, ok := extra["col"]; ok && v != "" {
		colName, isStr := v.(string)
		if !isStr || !columnasExtraValidas[colName] {
			return 0, fmt.Errorf("columna extra no válida: %v", v)
		}
		if val, ok2 := extra["val"]; ok2 && val != nil {
			cols = append(cols, colName)
			args = append(args, val)
			placeholders = append(placeholders, "?")
		}
	}

	q := `INSERT INTO ` + tabla + ` (` + strings.Join(cols, ", ") +
		`) VALUES (` + strings.Join(placeholders, ", ") + `)`
	res, err := a.exec(q, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
