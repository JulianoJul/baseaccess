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
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 8, 11},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 8},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 8},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 7, 8},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{7},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12},
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
		FechaColumna:   "fecha_recibido",
		GerenciasIDs:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13},
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

func NewApp() *App { return &App{} }

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
		`CREATE TABLE IF NOT EXISTS ruta_procesos_leyenda (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			status_name TEXT NOT NULL UNIQUE,
			hex_color   TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_hojas (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			nombre       TEXT NOT NULL,
			fecha_inicio DATE NOT NULL,
			fecha_fin    DATE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_procesos (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			id_hoja     INTEGER NOT NULL,
			modulo      TEXT NOT NULL DEFAULT 'expedientes',
			descripcion TEXT NOT NULL,
			db_id       INTEGER,
			activo      INTEGER DEFAULT 1,
			CONSTRAINT fk_proc_hoja FOREIGN KEY (id_hoja) REFERENCES ruta_procesos_hojas(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS ruta_procesos_cronograma (
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
		)`,
		`INSERT OR IGNORE INTO ruta_procesos_leyenda (status_name, hex_color) VALUES
			('ACTIVIDADES PREVIAS (UNIDAD USUARIA)', '#FF4757'),
			('ANÁLISIS ECONÓMICO', '#1E90FF'),
			('ANÁLISIS TÉCNICO', '#2ED573'),
			('APERTURA DE OFERTAS', '#FFA502'),
			('APROBACIÓN PRESIDENCIA', '#A855F7'),
			('CONTROL DE DOCUMENTOS PRESIDENCIA', '#00D2D3'),
			('INICIO (COMISIÓN)', '#FF6B81'),
			('INICIO (CONTRATACIÓN)', '#2BCBBA'),
			('RESULTADOS', '#FDCB6E'),
			('VENTA DE PLIEGO DE CONDICIONES (CONTRATACIÓN)', '#6C5CE7')`,
		`INSERT OR IGNORE INTO ruta_procesos_hojas (id, nombre, fecha_inicio, fecha_fin) VALUES (1, 'Default', '2024-01-01', '2024-12-31')`,
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
	StatusName string `json:"status_name"`
	HexColor   string `json:"hex_color"`
}

type RutaProcesosProceso struct {
	ID          int                    `json:"id"`
	Modulo      string                 `json:"modulo"`
	Descripcion string                 `json:"descripcion"`
	DbID        int                    `json:"db_id"`
	Activo      bool                   `json:"activo"`
	Solped      string                 `json:"solped"`
	Estatus     string                 `json:"estatus_detalle"`
	Receptor    string                 `json:"receptor"`
	Timeline    map[string]interface{} `json:"timeline"`
}

type RutaProcesosHoja struct {
	ID          int    `json:"id"`
	Nombre      string `json:"nombre"`
	FechaInicio string `json:"fecha_inicio"`
	FechaFin    string `json:"fecha_fin"`
}

type RutaProcesosGanttData struct {
	Legend      []RutaProcesosLegend    `json:"legend"`
	Columns     []map[string]string     `json:"columns"`
	Processes   []RutaProcesosProceso   `json:"processes"`
	Hojas       []RutaProcesosHoja      `json:"hojas"`
	CurrentHoja *RutaProcesosHoja       `json:"current_hoja"`
	OffsetWeeks int                     `json:"offset_weeks"`
}

func (a *App) ObtenerRutaProcesosData(idHoja int, offsetWeeks int) (*RutaProcesosGanttData, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return nil, fmt.Errorf("no hay BD abierta")
	}

	legendRows, err := a.db.Query("SELECT id, status_name, hex_color FROM ruta_procesos_leyenda ORDER BY status_name COLLATE NOCASE")
	if err != nil {
		return nil, err
	}
	defer legendRows.Close()
	var legend []RutaProcesosLegend
	legendMap := make(map[string]string) // Map to quickly find colors
	for legendRows.Next() {
		var l RutaProcesosLegend
		if err := legendRows.Scan(&l.ID, &l.StatusName, &l.HexColor); err != nil {
			log.Printf("ObtenerRutaProcesosData: scan leyenda: %v", err)
			continue
		}
		legend = append(legend, l)
		legendMap[l.StatusName] = l.HexColor
	}
	if legend == nil {
		legend = []RutaProcesosLegend{}
	}

	// Obtener hojas (strftime ensures YYYY-MM-DD format, avoiding RFC3339 from DATE columns)
	hojasRows, err := a.db.Query("SELECT id, nombre, strftime('%Y-%m-%d', fecha_inicio), strftime('%Y-%m-%d', fecha_fin) FROM ruta_procesos_hojas ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer hojasRows.Close()
	var hojas []RutaProcesosHoja
	for hojasRows.Next() {
		var h RutaProcesosHoja
		if err := hojasRows.Scan(&h.ID, &h.Nombre, &h.FechaInicio, &h.FechaFin); err != nil {
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

	var columns []map[string]string
	var processes []RutaProcesosProceso
	if currentHoja != nil {
		columns = buildGanttColumns(currentHoja.FechaInicio, currentHoja.FechaFin, offsetWeeks)

		procRows, err := a.db.Query(`
			SELECT p.id, p.modulo, p.descripcion, p.db_id, p.activo,
				COALESCE(e.solped, 'NO APLICA'), COALESCE(e.estatus_detalle, 'NO APLICA'), COALESCE(e.receptor, 'NO APLICA'),
				COALESCE(e.fecha_recibido, ''), COALESCE(e.fecha_devuelto, ''), COALESCE(e.fecha_firma_contrato, ''),
				COALESCE(e.documento, ''), COALESCE(e.resultados_proceso, ''),
				COALESCE(exp.id_documento, 0), COALESCE(exp.id_resultado, 0)
			FROM ruta_procesos_procesos p
			LEFT JOIN vw_reporte_excel_contrataciones e ON p.modulo = 'expedientes' AND p.db_id = e.id_expediente
			LEFT JOIN expedientes exp ON p.modulo = 'expedientes' AND p.db_id = exp.id_expediente
			WHERE p.id_hoja = ?
			ORDER BY p.id`, currentHoja.ID)
		if err != nil {
			return nil, err
		}
		defer procRows.Close()

		for procRows.Next() {
			var p RutaProcesosProceso
			var activo int
			var fRecibido, fDevuelto, fFirma string
			var doc, resProc string
			var idDoc, idRes int
			if err := procRows.Scan(&p.ID, &p.Modulo, &p.Descripcion, &p.DbID, &activo, &p.Solped, &p.Estatus, &p.Receptor, &fRecibido, &fDevuelto, &fFirma, &doc, &resProc, &idDoc, &idRes); err != nil {
				continue
			}
		p.Activo = activo == 1
		p.Timeline = map[string]interface{}{}
		
		if p.Modulo == "expedientes" {
		
		// Auto-populate timeline based on database dates and catalogs
		if fRecibido != "" && len(fRecibido) >= 10 {
			dateKey := fRecibido[:10]
			stName := "PENDIENTE"
			if doc != "" {
				stName = doc
			}
			hexColor := "#FFA500"
			if c, ok := legendMap[stName]; ok {
				hexColor = c
			}
			p.Timeline[dateKey] = map[string]string{
				"status_name": stName,
				"hex_color":   hexColor,
				"note":        "Recibido: " + doc,
			}
		}
		if fDevuelto != "" && len(fDevuelto) >= 10 {
			dateKey := fDevuelto[:10]
			stName := "DEVUELTO"
			if resProc != "" && idRes != 0 {
				stName = resProc // Explicit link to what the actual result is!
			}
			hexColor := "#EF4444"
			if c, ok := legendMap[stName]; ok {
				hexColor = c
			}
			p.Timeline[dateKey] = map[string]string{
				"status_name": stName,
				"hex_color":   hexColor,
				"note":        "Devuelto: " + stName,
			}
		}
		if fFirma != "" && len(fFirma) >= 10 {
			dateKey := fFirma[:10]
			stName := "FIRMADO"
			if resProc != "" && idRes != 0 {
				stName = resProc + " (FIRMADO)" // If there's a result, we show it
			}
			hexColor := "#10B981"
			if c, ok := legendMap[stName]; ok {
				hexColor = c
			}
			p.Timeline[dateKey] = map[string]string{
				"status_name": stName,
				"hex_color":   hexColor,
				"note":        "Fecha de firma de contrato",
			}
		}
		}
		processes = append(processes, p)
	}
	} // Cierra el if currentHoja != nil

	if len(processes) == 0 {
		return &RutaProcesosGanttData{Legend: legend, Columns: columns, Processes: []RutaProcesosProceso{}, Hojas: hojas, CurrentHoja: currentHoja, OffsetWeeks: offsetWeeks}, nil
	}

	procMap := make(map[int]*RutaProcesosProceso, len(processes))
	for i := range processes {
		procMap[processes[i].ID] = &processes[i]
	}

	args := make([]interface{}, len(processes))
	placeholders := make([]string, len(processes))
	for i, p := range processes {
		args[i] = p.ID
		placeholders[i] = "?"
	}
	cronoRows, err := a.db.Query(
		"SELECT c.id_proceso, strftime('%Y-%m-%d', c.fecha) AS fecha, c.nota, l.status_name, l.hex_color FROM ruta_procesos_cronograma c LEFT JOIN ruta_procesos_leyenda l ON c.id_leyenda = l.id WHERE c.id_proceso IN ("+strings.Join(placeholders, ",")+")",
		args...)
	if err != nil {
		return nil, err
	}
	defer cronoRows.Close()
	for cronoRows.Next() {
		var idProc int
		var fecha string
		var notaNull, statusNameNull, hexColorNull sql.NullString
		if err := cronoRows.Scan(&idProc, &fecha, &notaNull, &statusNameNull, &hexColorNull); err != nil {
			log.Printf("ObtenerRutaProcesosData: scan cronograma: %v", err)
			continue
		}
		if p, ok := procMap[idProc]; ok {
			statusName := ""
			if statusNameNull.Valid {
				statusName = statusNameNull.String
			}
			hexColor := ""
			if hexColorNull.Valid {
				hexColor = hexColorNull.String
			}
			nota := ""
			if notaNull.Valid {
				nota = notaNull.String
			}
			p.Timeline[fecha] = map[string]string{
				"status_name": statusName,
				"hex_color":   hexColor,
				"note":        nota,
			}
		}
	}

	return &RutaProcesosGanttData{
		Legend:      legend,
		Columns:     columns,
		Processes:   processes,
		Hojas:       hojas,
		CurrentHoja: currentHoja,
		OffsetWeeks: offsetWeeks,
	}, nil
}

// parseDateFlex tries YYYY-MM-DD first, then RFC3339 (from SQLite DATE columns)
func parseDateFlex(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	// Try also the T-separated without timezone
	t, err = time.Parse("2006-01-02T15:04:05Z", s)
	return t, err
}

func buildGanttColumns(inicioStr string, finStr string, offsetWeeks int) []map[string]string {
	dayNames := []string{"L", "M", "X", "J", "V"}
	inicio, err := parseDateFlex(inicioStr)
	if err != nil {
		inicio = time.Now()
	}
	fin, err := parseDateFlex(finStr)
	if err != nil || fin.Before(inicio) {
		fin = inicio.AddDate(0, 3, 0)
	}

	// Calculate the actual start considering offsetWeeks
	start := inicio
	if offsetWeeks > 0 {
		start = start.AddDate(0, 0, offsetWeeks*7)
	}

	// Adjust start to Monday
	offset := int(start.Weekday() - time.Monday)
	if offset < 0 {
		offset += 7
	}
	start = start.AddDate(0, 0, -offset)

	columns := make([]map[string]string, 0)
	maxBizDays := 365
	bizCount := 0
	weekNum := 1

	for i := 0; bizCount < maxBizDays; i++ {
		date := start.AddDate(0, 0, i)
		if date.After(fin) {
			break
		}
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			continue
		}
		bizCount++
		if bizCount > 1 && bizCount%5 == 1 {
			weekNum++
		}
		weekLabel := fmt.Sprintf("SEMANA %d", weekNum)
		dayName := dayNames[(bizCount-1)%5]
		columns = append(columns, map[string]string{
			"day_name":   dayName,
			"week_label": weekLabel,
			"date_str":   date.Format(fechaLayout),
		})
	}
	return columns
}

func (a *App) ToggleRutaProceso(id int, activo bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	v := 0
	if activo {
		v = 1
	}
	_, err := a.db.Exec("UPDATE ruta_procesos_procesos SET activo = ? WHERE id = ?", v, id)
	return err
}

func (a *App) AgregarRutaProceso(idHoja int, modulo, descripcion string, dbID int) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	var dbIDVal interface{}
	if dbID > 0 {
		dbIDVal = dbID
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_procesos (id_hoja, modulo, descripcion, db_id, activo) VALUES (?, ?, ?, ?, 1)", idHoja, modulo, descripcion, dbIDVal)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) ObtenerExpedientesDisponiblesRuta() ([]map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return nil, fmt.Errorf("no hay BD abierta")
	}
	rows, err := a.db.Query(`
		SELECT e.id_expediente, e.solped, e.descripcion_proceso
		FROM vw_reporte_excel_contrataciones e
		ORDER BY e.id_expediente DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var solped string
		var descNull sql.NullString
		if err := rows.Scan(&id, &solped, &descNull); err != nil {
			log.Printf("ObtenerExpedientesDisponiblesRuta: scan: %v", err)
			continue
		}
		desc := ""
		if descNull.Valid {
			desc = descNull.String
		}
		result = append(result, map[string]interface{}{
			"id":                 id,
			"solped":             solped,
			"descripcion_proceso": desc,
		})
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

func (a *App) ObtenerRegistrosDisponiblesRuta(modulo string) ([]map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.db == nil {
		return nil, fmt.Errorf("no hay BD abierta")
	}

	type moduleQuery struct {
		vista    string
		idCol    string
		labelCol string
	}
	queries := map[string]moduleQuery{
		"expedientes":     {"vw_reporte_excel_contrataciones", "id_expediente", "solped"},
		"requisiciones":   {"vw_reporte_req_materiales", "id_requisicion", "descripcion_materiales"},
		"memorandums":     {"vw_reporte_memorandums", "id_memorandum", "asunto"},
		"recobros":        {"vw_reporte_recobros", "id_recobro", "asunto"},
		"valuaciones":     {"vw_reporte_valuaciones", "id_valuacion", "solped"},
		"aprobacion_jd":   {"vw_reporte_aprobacion_jd", "id_aprobacion", "solped"},
		"certificacion_bdu": {"vw_reporte_certificacion_bdu", "id_certificacion", "nro_certificacion"},
		"vacaciones":      {"vw_reporte_vacaciones", "id_vacacion", "nombre_empleado"},
		"reposos_medicos": {"vw_reporte_reposos_medicos", "id_reposo", "nombre_empleado"},
	}

	q, ok := queries[modulo]
	if !ok {
		return nil, fmt.Errorf("módulo inválido: %s", modulo)
	}

	rows, err := a.db.Query(fmt.Sprintf("SELECT %s, %s FROM %s ORDER BY %s DESC", q.idCol, q.labelCol, q.vista, q.idCol))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var label string
		if err := rows.Scan(&id, &label); err != nil {
			log.Printf("ObtenerRegistrosDisponiblesRuta (%s): scan: %v", modulo, err)
			continue
		}
		result = append(result, map[string]interface{}{
			"id":     id,
			"label":  label,
			"modulo": modulo,
		})
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

func (a *App) EliminarRutaProceso(id int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	return a.withTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec("DELETE FROM ruta_procesos_cronograma WHERE id_proceso = ?", id); err != nil {
			return err
		}
		_, err := tx.Exec("DELETE FROM ruta_procesos_procesos WHERE id = ?", id)
		return err
	})
}

func (a *App) EliminarRutaCronogramaCelda(idProceso int, fechaStr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	return a.withTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM ruta_procesos_cronograma WHERE id_proceso = ? AND fecha = ?", idProceso, fechaStr)
		return err
	})
}

func (a *App) CrearRutaProcesosHoja(nombre, inicioStr, finStr string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_hojas (nombre, fecha_inicio, fecha_fin) VALUES (?, ?, ?)", nombre, inicioStr, finStr)
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
	_, err := a.db.Exec("DELETE FROM ruta_procesos_hojas WHERE id = ?", id)
	return err
}

func (a *App) CrearRutaProcesosLeyenda(nombre, color string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("no hay BD abierta")
	}
	res, err := a.db.Exec("INSERT INTO ruta_procesos_leyenda (status_name, hex_color) VALUES (?, ?)", nombre, color)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) ActualizarRutaProcesosLeyenda(id int, nombre, color string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}
	_, err := a.db.Exec("UPDATE ruta_procesos_leyenda SET status_name = ?, hex_color = ? WHERE id = ?", nombre, color, id)
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

func (a *App) GuardarCronogramaDia(idProceso int, fecha string, idLeyenda int, nota string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay BD abierta")
	}

	if err := a.crearBackup(); err != nil {
		log.Printf("Backup falló: %v", err)
	}

	// Check if entry exists
	var id int
	err := a.db.QueryRow("SELECT id FROM ruta_procesos_cronograma WHERE id_proceso = ? AND fecha = ?", idProceso, fecha).Scan(&id)
	if err == sql.ErrNoRows {
		if idLeyenda == 0 {
			// Nothing to save/delete
			return nil
		}
		_, err = a.db.Exec("INSERT INTO ruta_procesos_cronograma (id_proceso, fecha, id_leyenda, nota) VALUES (?, ?, ?, ?)", idProceso, fecha, idLeyenda, nota)
		return err
	} else if err != nil {
		return err
	}

	if idLeyenda == 0 {
		// Clear/delete cell
		_, err = a.db.Exec("DELETE FROM ruta_procesos_cronograma WHERE id = ?", id)
		return err
	}

	// Update existing
	_, err = a.db.Exec("UPDATE ruta_procesos_cronograma SET id_leyenda = ?, nota = ? WHERE id = ?", idLeyenda, nota, id)
	return err
}
