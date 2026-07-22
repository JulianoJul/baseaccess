package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed data/sql/*.sql
var sqlFS embed.FS

const DefaultBackupMaxCopies = 2

var backupMaxCopies atomic.Int64

func init() { backupMaxCopies.Store(DefaultBackupMaxCopies) }

type ModuloConfig struct {
	Nombre         string
	Tabla          string
	Vista          string
	IDColumna      string
	HistorialTabla string
	Columnas       []string
	FechaColumna   string
	GerenciasIDs   []int
	OrdenExcel     []string
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
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		OrdenExcel: []string{
			"id_expediente", "gerencia", "superintendencia", "emisor",
			"documento", "solped", "fecha_presupuesto_base", "presupuesto_base_usd",
			"tipo_cambio", "presupuesto_base_bs", "plan_contrataciones", "descripcion_proceso",
			"modalidad_contratacion", "art", "tipo_contrato", "nro_acta_apertura",
			"cantidad_frentes", "nro_resolucion_jd", "estatus_detalle",
			"observaciones", "fecha_recibido", "fecha_devuelto", "receptor",
			"nro_proceso", "resultados_proceso", "nro_contrato_sicac", "nro_contrato_sap",
			"empresa_adjudicada", "tiempo_ejecucion", "monto_adjudicado_bs", "monto_adjudicado_usd",
			"fecha_firma_contrato", "notas",
		},
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
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 8, 11},
		OrdenExcel: []string{
			"id_requisicion", "gerencia", "superintendencia", "emisor",
			"documento", "descripcion_materiales", "serial_equipo", "pase_sicesma",
			"estatus_detalle", "observaciones_entrega", "observaciones",
			"fecha_recibido", "fecha_devuelto", "receptor", "notas",
		},
	},
	"memorandums": {
		Nombre:         "Memorándums",
		Tabla:          "memorandums",
		Vista:          "vw_reporte_memorandums",
		IDColumna:      "id_memorandum",
		HistorialTabla: "hist_memorandums",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"asunto", "id_estatus", "fecha_recibido", "fecha_devuelto",
			"id_receptor", "observaciones", "notas",
		},
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		OrdenExcel: []string{
			"id_memorandum", "gerencia", "superintendencia", "emisor",
			"documento", "asunto", "estatus_detalle", "observaciones",
			"fecha_recibido", "fecha_devuelto", "receptor", "notas",
		},
	},
	"recobros": {
		Nombre:         "Recobros",
		Tabla:          "recobros",
		Vista:          "vw_reporte_recobros",
		IDColumna:      "id_recobro",
		HistorialTabla: "hist_recobros",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"asunto", "fecha_inicio", "fecha_final", "servicios", "beneficios",
			"nota_debito_reverso", "costo_servicio_usd", "id_estatus",
			"fecha_recibido", "fecha_devuelto", "id_receptor", "observaciones", "notas",
		},
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 8},
		OrdenExcel: []string{
			"id_recobro", "gerencia", "superintendencia", "emisor",
			"documento", "asunto", "fecha_inicio", "fecha_final",
			"servicios", "beneficios", "nota_debito_reverso", "costo_servicio_usd",
			"estatus_detalle", "observaciones", "fecha_recibido", "fecha_devuelto",
			"receptor", "notas",
		},
	},
	"valuaciones": {
		Nombre:         "Valuaciones",
		Tabla:          "valuaciones",
		Vista:          "vw_reporte_valuaciones",
		IDColumna:      "id_valuacion",
		HistorialTabla: "hist_valuaciones",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"solped", "presupuesto_base_bs", "presupuesto_base_usd", "descripcion_proceso",
			"id_estatus", "fecha_recibido", "fecha_devuelto", "id_receptor", "nro_proceso",
			"nro_contrato_sicac", "nro_contrato_sap", "id_empresa", "tiempo_ejecucion",
			"monto_adjudicado_bs", "monto_adjudicado_usd", "periodo_valuacion_desde",
			"periodo_valuacion_hasta", "monto_valuacion", "nro_proforma", "observaciones", "notas",
		},
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 8},
		OrdenExcel: []string{
			"id_valuacion", "gerencia", "superintendencia", "emisor",
			"documento", "solped", "presupuesto_base_bs", "presupuesto_base_usd",
			"descripcion_proceso", "estatus_detalle", "observaciones",
			"fecha_recibido", "fecha_devuelto", "receptor", "nro_proceso",
			"nro_contrato_sicac", "nro_contrato_sap", "empresa_adjudicada",
			"tiempo_ejecucion", "monto_adjudicado_bs", "monto_adjudicado_usd",
			"periodo_valuacion_desde", "periodo_valuacion_hasta",
			"monto_valuacion", "nro_proforma", "notas",
		},
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
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 7, 8},
		OrdenExcel: []string{
			"id_aprobacion_jd", "gerencia", "superintendencia", "emisor",
			"documento", "solped", "fecha_presupuesto_base", "presupuesto_base_bs",
			"tipo_cambio", "presupuesto_base_usd", "plan_contrataciones", "descripcion_proceso",
			"cantidad_frentes", "estatus_detalle", "observaciones",
			"fecha_recibido", "fecha_devuelto", "receptor",
			"tiempo_ejecucion", "notas",
		},
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
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{7},
		OrdenExcel: []string{
			"id_certificacion_bdu", "gerencia", "superintendencia", "emisor",
			"documento", "presupuesto_base_total_usd", "monto_adjudicado_total_usd",
			"monto_contrato", "monto_ejecutado", "monto_pagado",
			"estatus_detalle", "observaciones", "fecha_recibido", "fecha_devuelto",
			"receptor", "notas",
		},
	},
	"vacaciones": {
		Nombre:         "Vacaciones",
		Tabla:          "vacaciones",
		Vista:          "vw_reporte_vacaciones",
		IDColumna:      "id_vacacion",
		HistorialTabla: "hist_vacaciones",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"anio", "cantidad_dias", "fecha_desde", "fecha_hasta", "id_estatus",
			"fecha_recibido", "fecha_devuelto", "id_receptor", "observaciones", "notas",
		},
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12},
		OrdenExcel: []string{
			"id_vacacion", "gerencia", "superintendencia", "emisor",
			"documento", "anio", "cantidad_dias", "fecha_desde", "fecha_hasta",
			"estatus_detalle", "observaciones", "fecha_recibido", "fecha_devuelto",
			"receptor", "notas",
		},
	},
	"reposos_medicos": {
		Nombre:         "Reposos Médicos",
		Tabla:          "reposos_medicos",
		Vista:          "vw_reporte_reposos_medicos",
		IDColumna:      "id_reposo_medico",
		HistorialTabla: "hist_reposos_medicos",
		Columnas: []string{
			"id_gerencia", "id_superintendencia", "id_emisor", "id_documento",
			"dias_periodo", "fecha_desde", "fecha_hasta", "id_estatus",
			"fecha_recibido", "observaciones", "notas",
		},
		FechaColumna:   "fecha_creacion",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13},
		OrdenExcel: []string{
			"id_reposo_medico", "gerencia", "superintendencia", "emisor",
			"documento", "dias_periodo", "fecha_desde", "fecha_hasta",
			"estatus_detalle", "observaciones", "fecha_recibido", "notas",
		},
	},
}

func (cfg ModuloConfig) buildQueryHistorial() string {
	var selectCols []string
	selectCols = append(selectCols, "h.id_movimiento")
	var joins []string
	joinMap := make(map[string]bool)

	addJoin := func(join string) {
		if !joinMap[join] {
			joinMap[join] = true
			joins = append(joins, join)
		}
	}

	for _, col := range cfg.Columnas {
		switch col {
		case "id_gerencia":
			selectCols = append(selectCols, "COALESCE(g.nombre, '-') AS gerencia")
			addJoin("LEFT JOIN cat_gerencia g ON h.id_gerencia = g.id")
		case "id_superintendencia":
			selectCols = append(selectCols, "COALESCE(s.nombre, '-') AS superintendencia")
			addJoin("LEFT JOIN cat_superintendencia s ON h.id_superintendencia = s.id")
		case "id_emisor":
			selectCols = append(selectCols, "COALESCE(em.nombre, '-') AS emisor")
			addJoin("LEFT JOIN cat_responsables em ON h.id_emisor = em.id")
		case "id_receptor":
			selectCols = append(selectCols, "COALESCE(rec.nombre, '-') AS receptor")
			addJoin("LEFT JOIN cat_responsables rec ON h.id_receptor = rec.id")
		case "id_documento":
			selectCols = append(selectCols, "COALESCE(d.nombre, '-') AS documento")
			addJoin("LEFT JOIN cat_documento d ON h.id_documento = d.id")
		case "id_estatus":
			selectCols = append(selectCols, "COALESCE(ed.nombre, '-') AS estatus")
			addJoin("LEFT JOIN cat_estatus_detalle ed ON h.id_estatus = ed.id")
		case "id_plan":
			selectCols = append(selectCols, "COALESCE(p.nombre, '-') AS plan_contrataciones")
			addJoin("LEFT JOIN cat_plan_contratacion p ON h.id_plan = p.id")
		case "id_resultado":
			selectCols = append(selectCols, "COALESCE(rp.nombre, '-') AS resultado")
			addJoin("LEFT JOIN cat_resultado_proceso rp ON h.id_resultado = rp.id")
		case "id_empresa":
			selectCols = append(selectCols, "COALESCE(emp.nombre, '-') AS empresa")
			addJoin("LEFT JOIN cat_empresas emp ON h.id_empresa = emp.id")
		case "id_tipo_contrato":
			selectCols = append(selectCols, "COALESCE(tc.nombre, '-') AS tipo_contrato")
			addJoin("LEFT JOIN cat_tipo_contrato tc ON h.id_tipo_contrato = tc.id")
		case "id_modalidad":
			selectCols = append(selectCols, "COALESCE(mo.nombre, '-') AS modalidad")
			addJoin("LEFT JOIN cat_modalidad mo ON h.id_modalidad = mo.id")
		case "id_art":
			selectCols = append(selectCols, "COALESCE(a.nombre, '-') AS art")
			addJoin("LEFT JOIN cat_art a ON h.id_art = a.id")
		default:
			selectCols = append(selectCols, fmt.Sprintf("h.%s", col))
		}
	}

	query := fmt.Sprintf("SELECT %s FROM %s h %s WHERE h.%s = ? ORDER BY h.id_movimiento DESC",
		strings.Join(selectCols, ", "),
		cfg.HistorialTabla,
		strings.Join(joins, " "),
		cfg.IDColumna,
	)
	return query
}

const (
	queryObtenerRutaProcesos   = `SELECT e.id_expediente, e.solped, e.descripcion_proceso, e.emisor, e.receptor, e.documento, e.estatus_detalle, e.fecha_recibido, e.fecha_devuelto, e.nro_proceso FROM vw_reporte_excel_contrataciones e ORDER BY e.estatus_detalle, e.id_expediente DESC`
	queryObtenerDocsPendientes = `SELECT e.id_expediente, e.solped, e.descripcion_proceso, e.emisor, e.receptor, e.documento, e.estatus_detalle, e.fecha_recibido, e.fecha_devuelto, e.nro_proceso, e.empresa_adjudicada FROM vw_reporte_excel_contrataciones e WHERE e.estatus_detalle IS NOT NULL AND UPPER(e.estatus_detalle) != 'FIRMADO' ORDER BY e.estatus_detalle, e.id_expediente DESC`
	queryVacuum                 = `VACUUM`
	fechaLayout                 = "2006-01-02"         // usado en queryRows y buildGanttColumns
	estatusFirmado              = "FIRMADO"            // seed cat_estatus_detalle id=2
	dsnParams                   = "?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000"
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
	"fecha_creacion":      true,
	"fecha_actualizacion": true,
}


type App struct {
	ctx    context.Context
	db     *sql.DB
	dbPath string
	mu     sync.RWMutex
}

type Row map[string]interface{}

func NewApp() *App { return &App{ctx: context.Background()} }

func (a *App) Startup(ctx context.Context) { a.ctx = ctx }

func (a *App) SetBackupMaxCopies(n int) {
	if n < 1 {
		n = 1
	}
	if n > 20 {
		n = 20
	}
	backupMaxCopies.Store(int64(n))
}

func (a *App) GetBackupMaxCopies() int { return int(backupMaxCopies.Load()) }

func (a *App) AbrirBaseDatos(filePath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db != nil {
		a.db.Close()
	}

	if strings.Contains(filePath, "?") {
		return fmt.Errorf("el nombre del archivo no puede contener el carácter '?'")
	}
	db, err := sql.Open("sqlite3", filePath+dsnParams)
	if err != nil {
		return fmt.Errorf("no se pudo abrir BD: %w", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("no se pudo conectar a BD: %w", err)
	}

	a.db = db
	a.dbPath = filePath

	if err := a.initRutaProcesosSchema(); err != nil {
		return fmt.Errorf("no se pudo inicializar schema ruta procesos: %w", err)
	}

	for _, f := range []string{
		"data/sql/01_master_control_docs_presidencia.sql",
		"data/sql/02_modulos_adicionales.sql",
		"data/sql/03_ruta_procesos.sql",
	} {
		content, err := sqlFS.ReadFile(f)
		if err != nil {
			return fmt.Errorf("error leyendo %s: %w", f, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("error ejecutando %s: %w", f, err)
		}
	}

	return nil
}

func (a *App) initRutaProcesosSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS ruta_procesos_hoja (
			id     INTEGER PRIMARY KEY AUTOINCREMENT,
			nombre TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_junta (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			id_hoja     INTEGER NOT NULL REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,
			numero      INTEGER NOT NULL,
			consecutiva INTEGER NOT NULL,
			fecha       TEXT NOT NULL,
			UNIQUE(id_hoja, numero)
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_junta_semana (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			id_junta     INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
			numero       INTEGER NOT NULL,
			fecha_inicio TEXT NOT NULL,
			fecha_fin    TEXT NOT NULL,
			UNIQUE(id_junta, numero)
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_junta_proceso (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			id_junta  INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
			numero    INTEGER NOT NULL,
			proceso   TEXT NOT NULL,
			UNIQUE(id_junta, numero)
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_cronograma (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			id_junta_proceso INTEGER NOT NULL REFERENCES ruta_procesos_junta_proceso(id) ON DELETE CASCADE,
			fecha            TEXT NOT NULL,
			id_leyenda       INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE RESTRICT,
			nota             TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_leyenda (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			nombre    TEXT NOT NULL,
			color     TEXT NOT NULL DEFAULT '#FFFFFF',
			ambito    TEXT NOT NULL DEFAULT 'junta',
			id_hoja   INTEGER REFERENCES ruta_procesos_hoja(id) ON DELETE CASCADE,
			bloqueado INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_junta_leyenda (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			id_junta   INTEGER NOT NULL REFERENCES ruta_procesos_junta(id) ON DELETE CASCADE,
			id_leyenda INTEGER NOT NULL REFERENCES ruta_procesos_leyenda(id) ON DELETE CASCADE,
			orden      INTEGER NOT NULL DEFAULT 0,
			UNIQUE(id_junta, id_leyenda)
		)`,
		`INSERT OR IGNORE INTO ruta_procesos_leyenda (id, nombre, color, ambito, id_hoja, bloqueado) VALUES
			(1, 'ACTIVIDADES PREVIAS (UNIDAD USUARIA)', '#FF4757', 'global', NULL, 0),
			(2, 'INICIO (CONTRATACIÓN)', '#2BCBBA', 'global', NULL, 0),
			(3, 'VENTA DE PLIEGO DE CONDICIONES (CONTRATACIÓN)', '#6C5CE7', 'global', NULL, 0),
			(4, 'INICIO (COMISIÓN)', '#FF6B81', 'global', NULL, 0),
			(5, 'APERTURA DE OFERTAS', '#FFA502', 'global', NULL, 0),
			(6, 'ANÁLISIS TÉCNICO', '#2ED573', 'global', NULL, 0),
			(7, 'ANÁLISIS ECONÓMICO', '#1E90FF', 'global', NULL, 0),
			(8, 'RESULTADOS', '#FDCB6E', 'global', NULL, 0),
			(9, 'APROBACIÓN PRESIDENCIA', '#A855F7', 'global', NULL, 0),
			(10, 'CONTROL DE DOCUMENTOS PRESIDENCIA', '#00D2D3', 'global', NULL, 0)`,
	}
	for _, s := range statements {
		if _, err := a.db.Exec(s); err != nil {
			return fmt.Errorf("error ejecutando: %s: %w", s[:60], err)
		}
	}
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
	if a.dbPath == "" || a.db == nil {
		return nil
	}

	dir := filepath.Dir(a.dbPath)
	base := filepath.Base(a.dbPath)
	sep := string(filepath.Separator)
	maxCopies := int(backupMaxCopies.Load())
	tmpPath := dir + sep + base + ".bak.tmp"

	if err := a.copyDBCheckpointed(a.dbPath, tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error creando backup temporal: %w", err)
	}

	oldest := dir + sep + base + ".bak." + strconv.Itoa(maxCopies)
	if err := os.Remove(oldest); err != nil && !os.IsNotExist(err) {
		log.Printf("error removiendo backup antiguo %s: %v", oldest, err)
	}
	for i := maxCopies - 1; i >= 1; i-- {
		src := dir + sep + base + ".bak." + strconv.Itoa(i)
		dst := dir + sep + base + ".bak." + strconv.Itoa(i+1)
		if _, err := os.Stat(src); err == nil {
			if err := os.Rename(src, dst); err != nil {
				log.Printf("error renombrando backup %s → %s: %v", src, dst, err)
			}
		}
	}
	if err := os.Rename(tmpPath, dir+sep+base+".bak.1"); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error moviendo backup temporal: %w", err)
	}
	return nil
}

func (a *App) copyDBCheckpointed(srcPath, destPath string) error {
	// Forzar checkpoint WAL para que todos los cambios estén en el archivo principal
	if _, err := a.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		log.Printf("WAL checkpoint falló (no crítico): %v", err)
	}

	srcFile, err := os.Open(srcPath)
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

func (a *App) DescargarBD(destPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.dbPath == "" {
		return fmt.Errorf("no hay base de datos abierta")
	}

	return a.copyDBCheckpointed(a.dbPath, destPath)
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
					row[col] = v.Format(fechaLayout)
				}
			default:
				if val == nil {
					row[col] = ""
				} else {
					row[col] = val
				}
			}
		}
			results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return results, err
	}
	return results, nil
}

func (a *App) exec(query string, args ...interface{}) (sql.Result, error) {
	if a.db == nil {
		return nil, fmt.Errorf("no hay base de datos abierta")
	}
	return a.db.Exec(query, args...)
}

func sanitizarOrden(orden string, def string, columnasVista []string) string {
	validMap := make(map[string]bool, len(columnasVista)+len(columnasOrdenValidas)+1)
	for _, c := range columnasVista {
		validMap[c] = true
	}
	for c := range columnasOrdenValidas {
		validMap[c] = true
	}
	validMap[def] = true

	partes := strings.Fields(orden)
	if len(partes) == 0 {
		return def + " DESC"
	}
	col := partes[0]
	if !validMap[col] {
		return def + " DESC"
	}
	dir := "DESC"
	if len(partes) > 1 {
		d := strings.ToUpper(partes[1])
		if d == "ASC" || d == "DESC" {
			dir = d
		}
	}
	return col + " " + dir
}

func (a *App) ObtenerFilas(moduloKey string, orden string) ([]Row, error) {
	// Need to get columns before locking the main mutex if we use ObtenerColumnasVista
	// But actually ObtenerColumnasVista only takes an RLock, which is safe.
	// However, to avoid any lock issues, we just do the query first.
	cfg, ok := Modulos[moduloKey]
	if !ok {
		return nil, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	cols, err := a.ObtenerColumnasVista(cfg.Vista)
	if err != nil {
		cols = cfg.Columnas // Fallback
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	orden = sanitizarOrden(orden, cfg.IDColumna, cols)
	q := `SELECT * FROM ` + cfg.Vista + ` ORDER BY ` + orden
	return a.queryRows(q)
}

func (a *App) ObtenerFilasPaginado(moduloKey, orden string, pagina, pageSize int) ([]Row, int, error) {
	cfg, ok := Modulos[moduloKey]
	if !ok {
		return nil, 0, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	cols, err := a.ObtenerColumnasVista(cfg.Vista)
	if err != nil {
		cols = cfg.Columnas
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	if pagina < 1 {
		pagina = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	orden = sanitizarOrden(orden, cfg.IDColumna, cols)

	var total int
	countQ := `SELECT COUNT(*) FROM ` + cfg.Vista
	if err := a.db.QueryRow(countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("contar filas: %w", err)
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if pagina > totalPages {
		pagina = totalPages
	}

	offset := (pagina - 1) * pageSize
	q := `SELECT * FROM ` + cfg.Vista + ` ORDER BY ` + orden + ` LIMIT ? OFFSET ?`
	filas, err := a.queryRows(q, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return filas, totalPages, nil
}

func (a *App) ObtenerFilaPorId(moduloKey string, id int) (Row, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cfg, ok := Modulos[moduloKey]
	if !ok {
		return nil, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	q := `SELECT * FROM ` + cfg.Tabla + ` WHERE ` + cfg.IDColumna + ` = ?`
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

	var id int64
	if ok && idVal != nil {
		switch v := idVal.(type) {
		case float64:
			id = int64(v)
		case int:
			id = int64(v)
		case int64:
			id = v
		case string:
			if v != "" {
				parsed, err := strconv.ParseInt(v, 10, 64)
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
		var sets []string
		var setVals []interface{}
		for i, col := range cfg.Columnas {
			if col == cfg.IDColumna {
				continue
			}
			if vals[i] != nil {
				sets = append(sets, col+" = ?")
				setVals = append(setVals, vals[i])
			} else {
				sets = append(sets, col+" = NULL")
			}
		}
		if len(sets) == 0 {
			return id, nil
		}
		setVals = append(setVals, id)
		q := `UPDATE ` + cfg.Tabla + ` SET ` + strings.Join(sets, ", ") +
			`, fecha_actualizacion = CURRENT_DATE WHERE ` + cfg.IDColumna + ` = ?`
		_, err := a.exec(q, setVals...)
		if err != nil {
			return 0, fmt.Errorf("error al actualizar: %w", err)
		}
		return id, nil
	}

	var insCols []string
	var insVals []interface{}
	for i, col := range cfg.Columnas {
		if col == cfg.IDColumna {
			continue
		}
		if vals[i] != nil {
			insCols = append(insCols, col)
			insVals = append(insVals, vals[i])
		}
	}
	placeholders := make([]string, len(insCols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	q := `INSERT INTO ` + cfg.Tabla + ` (` + strings.Join(insCols, ", ") +
		`) VALUES (` + strings.Join(placeholders, ", ") + `)`
	res, err := a.exec(q, insVals...)
	if err != nil {
		return 0, fmt.Errorf("error al insertar: %w", err)
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo id insertado: %w", err)
	}
	return id, nil
}

func (a *App) withTx(fn func(tx *sql.Tx) error) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
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

	return a.withTx(func(tx *sql.Tx) error {
		if cfg.HistorialTabla != "" {
			qHist := `DELETE FROM ` + cfg.HistorialTabla + ` WHERE ` + cfg.IDColumna + ` = ?`
			if _, err := tx.Exec(qHist, id); err != nil {
				return err
			}
		}

		if moduloKey == "expedientes" {
			if _, err := tx.Exec("DELETE FROM ruta_procesos_cronograma WHERE id_expediente = ? OR id_proceso IN (SELECT id FROM ruta_procesos_procesos WHERE db_id = ?)", id, id); err != nil {
				return err
			}
			if _, err := tx.Exec("DELETE FROM ruta_procesos_procesos WHERE db_id = ?", id); err != nil {
				return err
			}
		}

		qDel := `DELETE FROM ` + cfg.Tabla + ` WHERE ` + cfg.IDColumna + ` = ?`
		_, err := tx.Exec(qDel, id)
		return err
	})
}

func (a *App) ObtenerHistorialFila(moduloKey string, id int) ([]Row, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cfg, ok := Modulos[moduloKey]
	if !ok {
		return nil, fmt.Errorf("modulo no soportado: %s", moduloKey)
	}

	if cfg.HistorialTabla == "" {
		return nil, fmt.Errorf("modulo %s no tiene soporte para historial", moduloKey)
	}

	query := cfg.buildQueryHistorial()
	return a.queryRows(query, id)
}

func (a *App) ObtenerColumnasVista(vista string) ([]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return nil, fmt.Errorf("no hay BD abierta")
	}

	// Validar que la vista esté en el whitelist de módulos conocidos
	vistaValida := false
	for _, cfg := range Modulos {
		if cfg.Vista == vista {
			vistaValida = true
			break
		}
	}
	if !vistaValida {
		return nil, fmt.Errorf("vista no válida: %s", vista)
	}

	rows, err := a.db.Query("SELECT * FROM " + vista + " LIMIT 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rows.Columns()
}

func (a *App) ObtenerRutaProcesos() ([]Row, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.queryRows(queryObtenerRutaProcesos)
}

func (a *App) ObtenerDocumentosPendientes() ([]Row, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.queryRows(queryObtenerDocsPendientes)
}

type RutaProcesosLegend struct {
	ID         int    `json:"id"`
	Nombre     string `json:"nombre"`
	Color      string `json:"color"`
	Ambito     string `json:"ambito"`
	IDHoja     *int   `json:"id_hoja,omitempty"`
	Bloqueado  bool   `json:"bloqueado"`
}

type RutaProcesosJunta struct {
	ID          int    `json:"id"`
	IDHoja      int    `json:"id_hoja"`
	Numero      int    `json:"numero"`
	Consecutiva int    `json:"consecutiva"`
	Fecha       string `json:"fecha"`
}

type RutaProcesosJuntaSemana struct {
	ID          int      `json:"id"`
	IDJunta     int      `json:"id_junta"`
	Numero      int      `json:"numero"`
	FechaInicio string   `json:"fecha_inicio"`
	FechaFin    string   `json:"fecha_fin"`
	Dias        []string `json:"dias"` // 5 fechas: L M X J V
}

type RutaProcesosJuntaProceso struct {
	ID       int    `json:"id"`
	IDJunta  int    `json:"id_junta"`
	Numero   int    `json:"numero"`
	Proceso  string `json:"proceso"`
	Timeline map[string][]RutaProcesosCronogramaEntry `json:"timeline"`
}

type RutaProcesosCronogramaEntry struct {
	ID        int    `json:"id"`
	IDProceso int    `json:"id_junta_proceso"`
	Fecha     string `json:"fecha"`
	IDLeyenda int    `json:"id_leyenda"`
	Nota      string `json:"nota,omitempty"`
	NombreLeyenda string `json:"status_name,omitempty"`
	HexColor      string `json:"hex_color,omitempty"`
}

type RutaProcesosJuntaLeyenda struct {
	ID        int    `json:"id"`
	IDJunta   int    `json:"id_junta"`
	IDLeyenda int    `json:"id_leyenda"`
	Orden     int    `json:"orden"`
}

type RutaProcesosHoja struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type RutaProcesosGanttData struct {
	Hojas        []RutaProcesosHoja       `json:"hojas"`
	CurrentHoja  *RutaProcesosHoja        `json:"current_hoja"`
	Juntas       []RutaProcesosJunta      `json:"juntas"`
	CurrentJunta *RutaProcesosJunta       `json:"current_junta"`
	Semanas      []RutaProcesosJuntaSemana `json:"semanas"`
	Procesos     []RutaProcesosJuntaProceso `json:"procesos"`
	Legend       []RutaProcesosLegend     `json:"legend"`
	JuntaLegend  []RutaProcesosJuntaLeyenda `json:"junta_legend"`
}

func (a *App) ObtenerRutaProcesosData(idHoja int, idJunta int) (*RutaProcesosGanttData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return nil, fmt.Errorf("no hay BD abierta")
	}

	// 1. Leyendas
	legendRows, err := a.db.Query(`SELECT id, nombre, color, ambito, id_hoja, bloqueado FROM ruta_procesos_leyenda`)
	if err != nil {
		return nil, err
	}
	defer legendRows.Close()
	var legend []RutaProcesosLegend
	for legendRows.Next() {
		var l RutaProcesosLegend
		var bloqueado int
		if err := legendRows.Scan(&l.ID, &l.Nombre, &l.Color, &l.Ambito, &l.IDHoja, &bloqueado); err != nil {
			log.Printf("ObtenerRutaProcesosData: scan leyenda: %v", err)
			continue
		}
		l.Bloqueado = bloqueado == 1
		legend = append(legend, l)
	}
	if legend == nil {
		legend = []RutaProcesosLegend{}
	}

	// 2. Hojas
	hojaRows, err := a.db.Query("SELECT id, nombre FROM ruta_procesos_hoja ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer hojaRows.Close()
	var hojas []RutaProcesosHoja
	for hojaRows.Next() {
		var h RutaProcesosHoja
		if err := hojaRows.Scan(&h.ID, &h.Nombre); err != nil {
			log.Printf("ObtenerRutaProcesosData: scan hoja: %v", err)
			continue
		}
		hojas = append(hojas, h)
	}
	if hojas == nil {
		hojas = []RutaProcesosHoja{}
	}

	var currentHoja *RutaProcesosHoja
	if len(hojas) > 0 {
		if idHoja == 0 {
			currentHoja = &hojas[0]
			idHoja = currentHoja.ID
		} else {
			for i := range hojas {
				if hojas[i].ID == idHoja {
					currentHoja = &hojas[i]
					break
				}
			}
		}
	}

	// 3. Juntas
	var juntas []RutaProcesosJunta
	if currentHoja != nil {
		juntaRows, err := a.db.Query("SELECT id, id_hoja, numero, consecutiva, fecha FROM ruta_procesos_junta WHERE id_hoja = ? ORDER BY numero", currentHoja.ID)
		if err != nil {
			return nil, err
		}
		defer juntaRows.Close()
		for juntaRows.Next() {
			var j RutaProcesosJunta
			if err := juntaRows.Scan(&j.ID, &j.IDHoja, &j.Numero, &j.Consecutiva, &j.Fecha); err != nil {
				log.Printf("ObtenerRutaProcesosData: scan junta: %v", err)
				continue
			}
			juntas = append(juntas, j)
		}
	}
	if juntas == nil {
		juntas = []RutaProcesosJunta{}
	}

	var currentJunta *RutaProcesosJunta
	if len(juntas) > 0 {
		if idJunta == 0 {
			currentJunta = &juntas[0]
		} else {
			for i := range juntas {
				if juntas[i].ID == idJunta {
					currentJunta = &juntas[i]
					break
				}
			}
		}
	}

	// 4. Semanas, Procesos, JuntaLegend (para TODAS las juntas de la hoja)
	var semanas []RutaProcesosJuntaSemana
	var procesos []RutaProcesosJuntaProceso
	var juntaLegend []RutaProcesosJuntaLeyenda

	if len(juntas) > 0 {
		juntaIDs := make([]interface{}, len(juntas))
		phs := make([]string, len(juntas))
		for i, j := range juntas {
			juntaIDs[i] = j.ID
			phs[i] = "?"
		}
		inClause := strings.Join(phs, ",")

		// 4a. Semanas
		semRows, err := a.db.Query("SELECT id, id_junta, numero, fecha_inicio, fecha_fin FROM ruta_procesos_junta_semana WHERE id_junta IN ("+inClause+") ORDER BY numero", juntaIDs...)
		if err == nil {
			for semRows.Next() {
				var s RutaProcesosJuntaSemana
				if err := semRows.Scan(&s.ID, &s.IDJunta, &s.Numero, &s.FechaInicio, &s.FechaFin); err == nil {
					s.Dias = calcDiasSemana(s.FechaInicio)
					semanas = append(semanas, s)
				}
			}
			semRows.Close()
		}

		// 4b. Procesos
		procRows, err := a.db.Query("SELECT id, id_junta, numero, proceso FROM ruta_procesos_junta_proceso WHERE id_junta IN ("+inClause+") ORDER BY numero", juntaIDs...)
		if err == nil {
			for procRows.Next() {
				var p RutaProcesosJuntaProceso
				if err := procRows.Scan(&p.ID, &p.IDJunta, &p.Numero, &p.Proceso); err == nil {
					p.Timeline = make(map[string][]RutaProcesosCronogramaEntry)
					procesos = append(procesos, p)
				}
			}
			procRows.Close()
		}

		// 4c. Cronograma (cargar entradas diarias para todos los procesos)
		if len(procesos) > 0 {
			args := make([]interface{}, len(procesos))
			cPhs := make([]string, len(procesos))
			procMap := make(map[int]*RutaProcesosJuntaProceso, len(procesos))
			for i := range procesos {
				procMap[procesos[i].ID] = &procesos[i]
				args[i] = procesos[i].ID
				cPhs[i] = "?"
			}

			cronoRows, err := a.db.Query(
				`SELECT c.id, c.id_junta_proceso, c.fecha, c.nota, l.nombre, l.color
				 FROM ruta_procesos_cronograma c
				 LEFT JOIN ruta_procesos_leyenda l ON c.id_leyenda = l.id
				 WHERE c.id_junta_proceso IN (`+strings.Join(cPhs, ",")+`)`,
				args...)
			if err == nil {
				for cronoRows.Next() {
					var e RutaProcesosCronogramaEntry
					var notaNull, nomNull, colorNull sql.NullString
					if err := cronoRows.Scan(&e.ID, &e.IDProceso, &e.Fecha, &notaNull, &nomNull, &colorNull); err == nil {
						if notaNull.Valid { e.Nota = notaNull.String }
						if nomNull.Valid { e.NombreLeyenda = nomNull.String }
						if colorNull.Valid { e.HexColor = colorNull.String }

						if p, ok := procMap[e.IDProceso]; ok {
							p.Timeline[e.Fecha] = append(p.Timeline[e.Fecha], e)
						}
					}
				}
				cronoRows.Close()
			}
		}

		// 4d. Junta — Leyendas
		jlRows, err := a.db.Query("SELECT id, id_junta, id_leyenda, orden FROM ruta_procesos_junta_leyenda WHERE id_junta IN ("+inClause+") ORDER BY orden", juntaIDs...)
		if err == nil {
			for jlRows.Next() {
				var jl RutaProcesosJuntaLeyenda
				if err := jlRows.Scan(&jl.ID, &jl.IDJunta, &jl.IDLeyenda, &jl.Orden); err == nil {
					juntaLegend = append(juntaLegend, jl)
				}
			}
			jlRows.Close()
		}
	}

	if semanas == nil { semanas = []RutaProcesosJuntaSemana{} }
	if procesos == nil { procesos = []RutaProcesosJuntaProceso{} }
	if juntaLegend == nil { juntaLegend = []RutaProcesosJuntaLeyenda{} }

	return &RutaProcesosGanttData{
		Hojas:        hojas,
		CurrentHoja:  currentHoja,
		Juntas:       juntas,
		CurrentJunta: currentJunta,
		Semanas:      semanas,
		Procesos:     procesos,
		Legend:       legend,
		JuntaLegend:  juntaLegend,
	}, nil
}

func calcDiasSemana(inicio string) []string {
	t, err := time.Parse("2006-01-02", inicio)
	if err != nil {
		return []string{inicio, inicio, inicio, inicio, inicio}
	}
	dias := make([]string, 5)
	for i := 0; i < 5; i++ {
		d := t.AddDate(0, 0, i)
		dias[i] = d.Format("2006-01-02")
	}
	return dias
}

// ============================================================
// Funciones CRUD v2 — Ruta Procesos
// ============================================================

// --- Hojas ---

func (a *App) CrearRutaProcesosHoja(nombre string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_hoja (nombre) VALUES (?)", nombre)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) EliminarRutaProcesosHoja(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("DELETE FROM ruta_procesos_hoja WHERE id = ?", id)
	return err
}

// --- Juntas ---

func (a *App) CrearRutaProcesosJunta(idHoja, numero, consecutiva int, fecha string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_junta (id_hoja, numero, consecutiva, fecha) VALUES (?, ?, ?, ?)", idHoja, numero, consecutiva, fecha)
	if err != nil {
		return 0, err
	}
	idJunta, _ := res.LastInsertId()

	// Heredar leyendas globales y de esta hoja
	rows, err := a.db.Query("SELECT id FROM ruta_procesos_leyenda WHERE ambito = 'global' OR (ambito = 'hoja' AND id_hoja = ?)", idHoja)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var idLey int
			if rows.Scan(&idLey) == nil {
				if _, err := a.db.Exec("INSERT INTO ruta_procesos_junta_leyenda (id_junta, id_leyenda, orden) VALUES (?, ?, 0)", idJunta, idLey); err != nil {
					log.Printf("Error heredando leyenda %d a junta %d: %v", idLey, idJunta, err)
				}
			}
		}
	} else {
		log.Printf("Error consultando leyendas para heredar: %v", err)
	}

	return idJunta, nil
}

func (a *App) ActualizarRutaProcesosJunta(id, numero, consecutiva int, fecha string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("UPDATE ruta_procesos_junta SET numero = ?, consecutiva = ?, fecha = ? WHERE id = ?", numero, consecutiva, fecha, id)
	return err
}

func (a *App) EliminarRutaProcesosJunta(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("DELETE FROM ruta_procesos_junta WHERE id = ?", id)
	return err
}

// --- Semanas ---

func (a *App) AgregarRutaProcesosSemana(idJunta, numero int, fechaInicio, fechaFin string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_junta_semana (id_junta, numero, fecha_inicio, fecha_fin) VALUES (?, ?, ?, ?)", idJunta, numero, fechaInicio, fechaFin)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) EliminarRutaProcesosSemanas(idJunta int, numeros []int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	for _, n := range numeros {
		if _, err := a.db.Exec("DELETE FROM ruta_procesos_junta_semana WHERE id_junta = ? AND numero = ?", idJunta, n); err != nil {
			return err
		}
	}
	// Renumerar las semanas restantes
	restantes, err := a.db.Query("SELECT id FROM ruta_procesos_junta_semana WHERE id_junta = ? ORDER BY numero", idJunta)
	if err != nil {
		return err
	}
	defer restantes.Close()
	var ids []int
	for restantes.Next() {
		var id int
		if restantes.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	for i, id := range ids {
		if _, err := a.db.Exec("UPDATE ruta_procesos_junta_semana SET numero = ? WHERE id = ?", i+1, id); err != nil {
			return err
		}
	}
	return nil
}

// --- Procesos de Junta ---

func (a *App) AgregarRutaProcesosProceso(idJunta, unusedNumero int, proceso string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	var numero int
	err := a.db.QueryRow("SELECT COALESCE(MAX(numero), 0) + 1 FROM ruta_procesos_junta_proceso WHERE id_junta = ?", idJunta).Scan(&numero)
	if err != nil {
		return 0, err
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_junta_proceso (id_junta, numero, proceso) VALUES (?, ?, ?)", idJunta, numero, proceso)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) EliminarRutaProcesosProceso(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("DELETE FROM ruta_procesos_junta_proceso WHERE id = ?", id)
	return err
}

func (a *App) ReordenarRutaProcesosProceso(idJunta, idProceso, direction int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	var curNum int
	err := a.db.QueryRow("SELECT numero FROM ruta_procesos_junta_proceso WHERE id = ?", idProceso).Scan(&curNum)
	if err != nil {
		return err
	}
	newNum := curNum + direction
	var swapID int
	err = a.db.QueryRow("SELECT id FROM ruta_procesos_junta_proceso WHERE id_junta = ? AND numero = ?", idJunta, newNum).Scan(&swapID)
	if err != nil {
		return nil // no hay con quien intercambiar
	}
	a.db.Exec("UPDATE ruta_procesos_junta_proceso SET numero = ? WHERE id = ?", newNum, idProceso)
	a.db.Exec("UPDATE ruta_procesos_junta_proceso SET numero = ? WHERE id = ?", curNum, swapID)
	return nil
}

// --- Leyendas ---

func (a *App) CrearRutaProcesosLeyenda(nombre, color, ambito string, idHoja *int, idJunta *int) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}

	var idHojaVal interface{}
	if ambito == "hoja" && idHoja != nil {
		idHojaVal = *idHoja
	}

	res, err := a.db.Exec("INSERT INTO ruta_procesos_leyenda (nombre, color, ambito, id_hoja) VALUES (?, ?, ?, ?)", nombre, color, ambito, idHojaVal)
	if err != nil {
		return 0, err
	}
	idLey, _ := res.LastInsertId()

	// Insertar en ruta_procesos_junta_leyenda según ámbito
	switch ambito {
	case "global":
		// Insertar en todas las juntas existentes
		rows, _ := a.db.Query("SELECT id FROM ruta_procesos_junta")
		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var idJ int
				if rows.Scan(&idJ) == nil {
					a.db.Exec("INSERT OR IGNORE INTO ruta_procesos_junta_leyenda (id_junta, id_leyenda, orden) VALUES (?, ?, 0)", idJ, idLey)
				}
			}
		}
	case "hoja":
		if idHoja != nil {
			rows, _ := a.db.Query("SELECT id FROM ruta_procesos_junta WHERE id_hoja = ?", *idHoja)
			if rows != nil {
				defer rows.Close()
				for rows.Next() {
					var idJ int
					if rows.Scan(&idJ) == nil {
						a.db.Exec("INSERT OR IGNORE INTO ruta_procesos_junta_leyenda (id_junta, id_leyenda, orden) VALUES (?, ?, 0)", idJ, idLey)
					}
				}
			}
		}
	case "junta":
		if idJunta != nil {
			a.db.Exec("INSERT OR IGNORE INTO ruta_procesos_junta_leyenda (id_junta, id_leyenda, orden) VALUES (?, ?, 0)", *idJunta, idLey)
		}
	}
	return idLey, nil
}

func (a *App) ActualizarRutaProcesosLeyenda(id int, nombre, color string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("UPDATE ruta_procesos_leyenda SET nombre = ?, color = ? WHERE id = ?", nombre, color, id)
	return err
}

func (a *App) EliminarRutaProcesosLeyenda(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("DELETE FROM ruta_procesos_leyenda WHERE id = ?", id)
	return err
}

func (a *App) ReordenarRutaProcesosLeyenda(idJunta, idLeyenda, direction int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}

	rows, err := a.db.Query("SELECT id, id_leyenda, orden FROM ruta_procesos_junta_leyenda WHERE id_junta = ? ORDER BY orden ASC, id ASC", idJunta)
	if err != nil {
		return err
	}
	defer rows.Close()

	type item struct {
		id        int
		idLeyenda int
		orden     int
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.idLeyenda, &it.orden); err == nil {
			items = append(items, it)
		}
	}

	idx := -1
	for i, it := range items {
		if it.idLeyenda == idLeyenda {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("leyenda no encontrada en la junta")
	}

	swapIdx := idx + direction
	if swapIdx >= 0 && swapIdx < len(items) {
		items[idx], items[swapIdx] = items[swapIdx], items[idx]
	}

	// Update all to ensure strict 0, 1, 2... sequence (fixes any gaps or ties)
	for i, it := range items {
		a.db.Exec("UPDATE ruta_procesos_junta_leyenda SET orden = ? WHERE id = ?", i, it.id)
	}

	return nil
}

func (a *App) ToggleBloquearRutaProcesosLeyenda(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("UPDATE ruta_procesos_leyenda SET bloqueado = CASE WHEN bloqueado THEN 0 ELSE 1 END WHERE id = ?", id)
	return err
}

// --- Cronograma ---

func (a *App) GuardarCronogramaDia(idProceso int, fecha string, idLeyenda int, nota string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_cronograma (id_junta_proceso, fecha, id_leyenda, nota) VALUES (?, ?, ?, ?)", idProceso, fecha, idLeyenda, nota)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) EliminarCronogramaDia(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}
	_, err := a.db.Exec("DELETE FROM ruta_procesos_cronograma WHERE id = ?", id)
	return err
}

type CatalogoItem struct {
	ID         int    `json:"id"`
	Nombre     string `json:"nombre"`
	IDGerencia int    `json:"id_gerencia,omitempty"`
}

func (a *App) ObtenerCatalogos() (map[string][]CatalogoItem, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

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
		var loopErr error
		for rows.Next() {
			var item CatalogoItem
			var extraVal sql.NullInt64
			scanArgs := []interface{}{&item.ID, &item.Nombre}
			if extraCol != "" {
				scanArgs = append(scanArgs, &extraVal)
			}
			if err := rows.Scan(scanArgs...); err != nil {
				loopErr = err
				break
			}
			if extraCol != "" && extraVal.Valid {
				item.IDGerencia = int(extraVal.Int64)
			}
			items = append(items, item)
		}
		rows.Close()
		if loopErr != nil {
			return nil, loopErr
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("ObtenerCatalogos: iteración %s: %w", key, err)
		}
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

