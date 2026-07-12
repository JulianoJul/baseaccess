package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const backupMaxCopies = 2

type App struct {
	ctx    context.Context
	db     *sql.DB
	dbPath string
	mu     sync.Mutex
}

type Row map[string]interface{}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
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
	oldest := dir + "/" + base + ".bak." + strconv.Itoa(backupMaxCopies)
	os.Remove(oldest)
	for i := backupMaxCopies - 1; i >= 1; i-- {
		src := dir + "/" + base + ".bak." + strconv.Itoa(i)
		dst := dir + "/" + base + ".bak." + strconv.Itoa(i+1)
		if _, err := os.Stat(src); err == nil {
			os.Rename(src, dst)
		}
	}
	src := a.dbPath
	dst := dir + "/" + base + ".bak.1"
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("error leyendo BD para backup: %w", err)
	}
	return os.WriteFile(dst, input, 0644)
}

func (a *App) DescargarBD(destPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.dbPath == "" {
		return fmt.Errorf("no hay base de datos abierta")
	}
	input, err := os.ReadFile(a.dbPath)
	if err != nil {
		return fmt.Errorf("error leyendo BD: %w", err)
	}
	return os.WriteFile(destPath, input, 0644)
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
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
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

var columnasOrdenValidas = map[string]bool{
	"id_expediente":       true,
	"fecha_creacion":      true,
	"fecha_actualizacion": true,
	"solped":              true,
	"gerencia":            true,
	"estatus_detalle":     true,
}

func sanitizarOrden(orden string) string {
	partes := strings.Fields(orden)
	if len(partes) == 0 {
		return "id_expediente DESC"
	}
	col := partes[0]
	dir := "DESC"
	if len(partes) > 1 {
		d := strings.ToUpper(partes[1])
		if d == "ASC" || d == "DESC" {
			dir = d
		}
	}
	if !columnasOrdenValidas[strings.ToLower(col)] {
		return "id_expediente DESC"
	}
	return col + " " + dir
}

func (a *App) ObtenerExpedientes(orden string) ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	orden = sanitizarOrden(orden)
	return a.queryRows(`SELECT * FROM vw_reporte_excel_contrataciones ORDER BY ` + orden)
}

func (a *App) ObtenerExpedientePorId(id int) (Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	rows, err := a.queryRows(`SELECT * FROM vw_reporte_excel_contrataciones WHERE id_expediente = ?`, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("expediente %d no encontrado", id)
	}
	return rows[0], nil
}

func (a *App) ObtenerRutaProcesos() ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.queryRows(`SELECT e.id_expediente, e.solped, e.descripcion_proceso,
		e.emisor, e.receptor, e.documento, e.estatus_detalle,
		e.fecha_recibido, e.fecha_devuelto, e.nro_proceso
		FROM vw_reporte_excel_contrataciones e
		ORDER BY e.estatus_detalle, e.id_expediente DESC`)
}

func (a *App) ObtenerDocumentosPendientes() ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.queryRows(`SELECT e.id_expediente, e.solped, e.descripcion_proceso,
		e.emisor, e.receptor, e.documento, e.estatus_detalle,
		e.fecha_recibido, e.fecha_devuelto, e.nro_proceso, e.empresa_adjudicada
		FROM vw_reporte_excel_contrataciones e
		WHERE e.estatus_detalle IS NOT NULL AND UPPER(e.estatus_detalle) != 'FIRMADO'
		ORDER BY e.estatus_detalle, e.id_expediente DESC`)
}

func (a *App) ObtenerHistorialCompleto(id int) ([]Row, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.queryRows(`
		SELECT
			h.id_movimiento,
			COALESCE(tc.nombre, '-')           AS tipo_contrato,
			COALESCE(g.nombre, '-')             AS gerencia,
			COALESCE(s.nombre, '-')             AS superintendencia,
			COALESCE(d.nombre, '-')             AS documento,
			COALESCE(em.nombre, '-')            AS emisor,
			COALESCE(rec.nombre, '-')           AS receptor,
			COALESCE(ed.nombre, '-')            AS estatus,
			COALESCE(h.fecha_recibido, '-')     AS fecha_recibido,
			COALESCE(h.fecha_devuelto, '-')     AS fecha_devuelto,
			COALESCE(h.nro_proceso, '-')        AS nro_proceso,
			h.presupuesto_base_usd,
			h.tipo_cambio,
			h.monto_adjudicado_usd,
			COALESCE(rp.nombre, '-')            AS resultado,
			COALESCE(emp.nombre, '-')           AS empresa,
			COALESCE(h.tiempo_ejecucion, '-')   AS tiempo_ejecucion,
			COALESCE(h.fecha_firma_contrato, '-') AS fecha_firma_contrato,
			COALESCE(h.observaciones, '') AS observaciones,
			COALESCE(h.notas, '') AS notas
		FROM historial_movimientos h
		LEFT JOIN cat_tipo_contrato tc      ON h.id_tipo_contrato    = tc.id
		LEFT JOIN cat_gerencia g            ON h.id_gerencia         = g.id
		LEFT JOIN cat_superintendencia s    ON h.id_superintendencia = s.id
		LEFT JOIN cat_documento d           ON h.id_documento        = d.id
		LEFT JOIN cat_responsables em       ON h.id_emisor           = em.id
		LEFT JOIN cat_responsables rec      ON h.id_receptor         = rec.id
		LEFT JOIN cat_estatus_detalle ed    ON h.id_estatus          = ed.id
		LEFT JOIN cat_resultado_proceso rp  ON h.id_resultado        = rp.id
		LEFT JOIN cat_empresas emp          ON h.id_empresa          = emp.id
		WHERE h.id_expediente = ?
		ORDER BY h.id_movimiento DESC`, id)
}

var columnasExpedientes = []string{
	"solped", "id_gerencia", "id_superintendencia", "id_emisor",
	"id_documento", "fecha_presupuesto_base", "presupuesto_base_usd",
	"tipo_cambio", "presupuesto_base_bs", "id_plan", "descripcion_proceso",
	"id_modalidad", "id_art", "id_tipo_contrato", "nro_acta_apertura",
	"cantidad_frentes", "nro_resolucion_jd", "id_estatus",
	"fecha_recibido", "fecha_devuelto", "id_receptor", "nro_proceso",
	"id_resultado", "nro_contrato_sicac", "nro_contrato_sap", "id_empresa",
	"tiempo_ejecucion", "monto_adjudicado_bs", "monto_adjudicado_usd",
	"fecha_firma_contrato", "observaciones", "notas",
}

func (a *App) GuardarExpediente(data map[string]interface{}) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db == nil {
		return 0, fmt.Errorf("no hay base de datos abierta")
	}

	if err := a.crearBackup(); err != nil {
		fmt.Printf("Backup falló: %v\n", err)
	}

	id, _ := data["id_expediente"].(float64)
	delete(data, "id_expediente")

	vals := make([]interface{}, len(columnasExpedientes))
	for i, col := range columnasExpedientes {
		v, ok := data[col]
		if !ok || v == nil {
			vals[i] = nil
		} else {
			vals[i] = v
		}
	}

	if id > 0 {
		sets := make([]string, len(columnasExpedientes))
		for i, col := range columnasExpedientes {
			sets[i] = col + " = ?"
		}
		q := `UPDATE expedientes SET ` + strings.Join(sets, ", ") +
			`, fecha_actualizacion = CURRENT_DATE WHERE id_expediente = ?`
		_, err := a.db.Exec(q, append(vals, id)...)
		if err != nil {
			return 0, fmt.Errorf("error al actualizar: %w", err)
		}
		return int64(id), nil
	}

	placeholders := make([]string, len(columnasExpedientes))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	q := `INSERT INTO expedientes (` + strings.Join(columnasExpedientes, ", ") +
		`) VALUES (` + strings.Join(placeholders, ", ") + `)`
	res, err := a.db.Exec(q, vals...)
	if err != nil {
		return 0, fmt.Errorf("error al insertar: %w", err)
	}
	return res.LastInsertId()
}

func (a *App) EliminarExpediente(id int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return fmt.Errorf("no hay base de datos abierta")
	}
	if err := a.crearBackup(); err != nil {
		fmt.Printf("Backup falló: %v\n", err)
	}
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM historial_movimientos WHERE id_expediente = ?", id); err != nil {
		return err
	}
	if _, err = tx.Exec("DELETE FROM expedientes WHERE id_expediente = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

type CatalogoItem struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

func (a *App) ObtenerCatalogos() (map[string][]CatalogoItem, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	tables := map[string]string{
		"gerencia":         "cat_gerencia",
		"superintendencia": "cat_superintendencia",
		"documento":        "cat_documento",
		"plan_contratacion": "cat_plan_contratacion",
		"modalidad":        "cat_modalidad",
		"art":              "cat_art",
		"tipo_contrato":    "cat_tipo_contrato",
		"estatus_detalle":  "cat_estatus_detalle",
		"resultado_proceso": "cat_resultado_proceso",
		"empresas":         "cat_empresas",
		"responsables":     "cat_responsables",
	}

	result := make(map[string][]CatalogoItem)
	for key, table := range tables {
		cols := "id, nombre"
		if key == "superintendencia" {
			cols = "id, nombre, id_gerencia"
		}
		rows, err := a.db.Query("SELECT "+cols+" FROM "+table+" ORDER BY nombre")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var items []CatalogoItem
		for rows.Next() {
			var item CatalogoItem
			if err := rows.Scan(&item.ID, &item.Nombre); err != nil {
				return nil, err
			}
			items = append(items, item)
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
		fmt.Printf("Backup falló: %v\n", err)
	}
	_, err := a.db.Exec("VACUUM")
	return err
}

func (a *App) GuardarNuevoCatalogo(tabla, nombre string, extra map[string]interface{}) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db != nil {
		if err := a.crearBackup(); err != nil {
			fmt.Printf("Backup falló: %v\n", err)
		}
	}

	cols := "nombre"
	vals := nombre
	args := []interface{}{nombre}

	if v, ok := extra["col"]; ok && v != "" {
		if val, ok2 := extra["val"]; ok2 && val != nil {
			cols += ", " + v.(string)
			vals += ", ?"
			args = append(args, val)
		}
	}

	res, err := a.db.Exec("INSERT INTO "+tabla+" ("+cols+") VALUES ("+vals+")", args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
