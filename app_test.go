package main

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSanitizarOrden(t *testing.T) {
	columnasValidas := []string{"col_a", "col_b", "fecha_creacion"}
	def := "fecha_creacion"

	tests := []struct {
		name     string
		orden    string
		expected string
	}{
		{"default empty", "", "fecha_creacion DESC"},
		{"valid column asc", "col_a ASC", "col_a ASC"},
		{"valid column desc", "col_b DESC", "col_b DESC"},
		{"valid column default dir", "col_a", "col_a DESC"},
		{"invalid column", "hacked; DROP TABLE", "fecha_creacion DESC"},
		{"case insensitive dir", "col_a asc", "col_a ASC"},
		{"extra words ignored", "col_b ASC EXTRA", "col_b ASC"},
		{"default column explicit", "fecha_creacion ASC", "fecha_creacion ASC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizarOrden(tt.orden, def, columnasValidas)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSanitizarOrden_ColumnasOrdenValidas(t *testing.T) {
	columnasOrdenValidas["extra_col"] = true
	defer func() { delete(columnasOrdenValidas, "extra_col") }()

	got := sanitizarOrden("extra_col", "col_a", []string{"col_a"})
	if got != "extra_col DESC" {
		t.Errorf("expected extra_col DESC, got %q", got)
	}
}

// --- Integration tests (SQLite) ---

func testApp(t *testing.T) *App {
	t.Helper()
	path := t.TempDir() + "/test.db"
	db, err := sql.Open("sqlite3", path+dsnParams)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	for _, f := range []string{
		"data/sql/01_master_control_docs_presidencia.sql",
		"data/sql/02_modulos_adicionales.sql",
		"data/sql/03_ruta_procesos.sql",
	} {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("ReadFile %s: %v", f, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("exec %s: %v", f, err)
		}
	}

	return &App{db: db}
}

func TestGuardarFila_InsertAndUpdate(t *testing.T) {
	app := testApp(t)
	m := app

	data := map[string]interface{}{
		"solped":               "TEST-SOLPED-01",
		"id_gerencia":          1,
		"id_superintendencia":  1,
		"id_emisor":            1,
		"id_documento":         1,
		"id_plan":              1,
		"id_modalidad":         1,
		"id_art":               1,
		"id_tipo_contrato":     1,
		"id_estatus":           1,
		"observaciones":        "test obs",
	}

	id, err := m.GuardarFila("expedientes", data)
	if err != nil {
		t.Fatalf("GuardarFila insert: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	fila, err := m.ObtenerFilaPorId("expedientes", int(id))
	if err != nil {
		t.Fatalf("ObtenerFilaPorId: %v", err)
	}
	if got := rowGetStr(fila, "solped"); got != "TEST-SOLPED-01" {
		t.Errorf("solped = %q", got)
	}

	data["solped"] = "UPDATED-SOLPED"
	data["id_expediente"] = id
	_, err = m.GuardarFila("expedientes", data)
	if err != nil {
		t.Fatalf("GuardarFila update: %v", err)
	}

	fila, err = m.ObtenerFilaPorId("expedientes", int(id))
	if err != nil {
		t.Fatalf("ObtenerFilaPorId after update: %v", err)
	}
	if got := rowGetStr(fila, "solped"); got != "UPDATED-SOLPED" {
		t.Errorf("solped after update = %q, want UPDATED-SOLPED", got)
	}
}

func TestGuardarFila_ClearFields(t *testing.T) {
	app := testApp(t)
	m := app

	id, err := m.GuardarFila("expedientes", map[string]interface{}{
		"solped":        "CLEAR-TEST",
		"id_gerencia":   1,
		"id_estatus":    1,
		"observaciones": "should be cleared",
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	_, err = m.GuardarFila("expedientes", map[string]interface{}{
		"id_expediente": id,
		"solped":        "CLEAR-TEST",
		"id_gerencia":   1,
		"id_estatus":    1,
		"observaciones": nil,
	})
	if err != nil {
		t.Fatalf("update with nil observaciones: %v", err)
	}

	fila, err := m.ObtenerFilaPorId("expedientes", int(id))
	if err != nil {
		t.Fatalf("ObtenerFilaPorId: %v", err)
	}
	if got := rowGetStr(fila, "observaciones"); got != "" {
		t.Errorf("observaciones should be cleared, got %q", got)
	}
}

func TestEliminarFila(t *testing.T) {
	app := testApp(t)
	m := app

	id, err := m.GuardarFila("expedientes", map[string]interface{}{
		"solped":      "DELETE-TEST",
		"id_gerencia": 1,
		"id_estatus":  1,
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	err = m.EliminarFila("expedientes", id)
	if err != nil {
		t.Fatalf("EliminarFila: %v", err)
	}

	_, err = m.ObtenerFilaPorId("expedientes", int(id))
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestObtenerFilas(t *testing.T) {
	app := testApp(t)
	m := app

	for i := 0; i < 3; i++ {
		_, err := m.GuardarFila("expedientes", map[string]interface{}{
			"solped":      "FILAS-TEST",
			"id_gerencia": 1,
			"id_estatus":  1,
		})
		if err != nil {
			t.Fatalf("insert %d: %v", i, err)
		}
	}

	filas, err := m.ObtenerFilas("expedientes", "id_expediente DESC")
	if err != nil {
		t.Fatalf("ObtenerFilas: %v", err)
	}
	if len(filas) < 3 {
		t.Errorf("expected >= 3 rows, got %d", len(filas))
	}
}

func TestObtenerCatalogos(t *testing.T) {
	app := testApp(t)
	m := app

	cats, err := m.ObtenerCatalogos()
	if err != nil {
		t.Fatalf("ObtenerCatalogos: %v", err)
	}
	if cats["gerencia"] == nil {
		t.Error("gerencia catalog should exist")
	}
	if cats["superintendencia"] == nil {
		t.Error("superintendencia catalog should exist")
	}
	if len(cats["gerencia"]) == 0 {
		t.Error("gerencia catalog should not be empty")
	}
}

func TestGuardarNuevoCatalogo(t *testing.T) {
	app := testApp(t)
	m := app

	id, err := m.GuardarNuevoCatalogo("cat_gerencia", "TEST-GERENCIA", nil)
	if err != nil {
		t.Fatalf("GuardarNuevoCatalogo: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected new id > 0, got %d", id)
	}

	_, err = m.GuardarNuevoCatalogo("cat_gerencia", "TEST-GERENCIA", nil)
	if err == nil {
		t.Error("duplicate insert should fail")
	}
}

func TestObtenerRutaProcesosData(t *testing.T) {
	app := testApp(t)
	m := app

	data, err := m.ObtenerRutaProcesosData()
	if err != nil {
		t.Fatalf("ObtenerRutaProcesosData: %v", err)
	}
	if data.Legend == nil || len(data.Legend) == 0 {
		t.Error("legend should not be empty")
	}
	if data.Columns == nil {
		t.Error("columns should not be nil")
	}
	if data.Processes == nil {
		t.Error("processes should not be nil")
	}
	if len(data.Legend) != 6 {
		t.Errorf("expected 6 legend entries, got %d", len(data.Legend))
	}
}

func TestObtenerExpedientesDisponiblesRuta(t *testing.T) {
	app := testApp(t)
	m := app

	_, err := m.GuardarFila("expedientes", map[string]interface{}{
		"solped":               "RUTA-TEST",
		"id_gerencia":          1,
		"id_estatus":           1,
		"descripcion_proceso":  "Test ruta process",
	})
	if err != nil {
		t.Fatalf("GuardarFila: %v", err)
	}

	disp, err := m.ObtenerExpedientesDisponiblesRuta()
	if err != nil {
		t.Fatalf("ObtenerExpedientesDisponiblesRuta: %v", err)
	}
	if len(disp) == 0 {
		t.Error("should return at least 1 expediente")
	}
	if len(disp) > 0 {
		r := disp[0]
		if _, ok := r["id"]; !ok {
			t.Error("result row should have 'id' key")
		}
	}
}

func TestCrearBackup(t *testing.T) {
	app := testApp(t)
	m := app

	tmpDir := t.TempDir()
	m.dbPath = tmpDir + "/test.db"

	db2, err := sql.Open("sqlite3", m.dbPath+"?_journal_mode=WAL")
	if err != nil {
		t.Fatalf("create disk db: %v", err)
	}
	db2.Exec("CREATE TABLE IF NOT EXISTS t(x)")
	db2.Close()

	err = m.crearBackup()
	if err != nil {
		t.Fatalf("crearBackup: %v", err)
	}

	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "test.db.bak") {
			return
		}
	}
	t.Error("no backup file found in temp dir")
}
