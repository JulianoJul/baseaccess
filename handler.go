package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//go:embed all:frontend
var frontendFS embed.FS

//go:embed templates/*
var templateFS embed.FS

type TemplateHandler struct {
	app    *App
	tmpl   *template.Template
	static http.Handler
}

func NewTemplateHandler(app *App) (*TemplateHandler, error) {
	funcMap := template.FuncMap{
		"safeHTML":   func(s string) template.HTML { return template.HTML(s) },
		"safeURL":    func(s string) template.URL { return template.URL(s) },
		"safeJS":     func(s string) template.JS { return template.JS(s) },
		"add":        func(a, b int) int { return a + b },
		"sub":        func(a, b int) int { return a - b },
		"seq":        seq,
		"seqFromTo":  seqFromTo,
		"dict":       dict,
		"rowGet":     rowGet,
		"rowGetStr":  rowGetStr,
		"rowGetNum":  rowGetNum,
		"estatusClass": estatusClass,
		"formatNum":  formatNumGo,
		"jsonEncode": jsonEncode,
		"hasDB":      func() bool { return app.db != nil },
		"truncate":   truncate,
		"isSelected": isSelected,
		"default":    defaultVal,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	subFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		return nil, err
	}

	return &TemplateHandler{
		app:    app,
		tmpl:   tmpl,
		static: http.FileServer(http.FS(subFS)),
	}, nil
}

// --- Funciones helper para templates ---

func seq(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i + 1
	}
	return s
}

func seqFromTo(from, to int) []int {
	if to < from {
		return nil
	}
	s := make([]int, to-from+1)
	for i := range s {
		s[i] = from + i
	}
	return s
}

func dict(values ...interface{}) map[string]interface{} {
	d := make(map[string]interface{})
	for i := 0; i+1 < len(values); i += 2 {
		if key, ok := values[i].(string); ok {
			d[key] = values[i+1]
		}
	}
	return d
}

func rowGet(r Row, key string) interface{} {
	if r == nil {
		return nil
	}
	return r[key]
}

func rowGetStr(r Row, key string) string {
	if r == nil {
		return ""
	}
	v, ok := r[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func rowGetNum(r Row, key string) float64 {
	if r == nil {
		return 0
	}
	v, ok := r[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	}
	return 0
}

func estatusClass(estatus string) string {
	if estatus == "" {
		return "bg-yellow-500/20 text-yellow-400"
	}
	switch strings.ToUpper(estatus) {
	case "FIRMADO":
		return "bg-emerald-500/20 text-emerald-400"
	case "PENDIENTE":
		return "bg-yellow-500/20 text-yellow-400"
	case "DEVUELTO PARA CORRECCIÓN":
		return "bg-orange-500/20 text-orange-400"
	case "DEVUELTO SIN FIRMA":
		return "bg-red-500/20 text-red-400"
	default:
		return "bg-gray-500/20 text-gray-400"
	}
}

func formatNumGo(v interface{}) string {
	if v == nil {
		return ""
	}
	switch n := v.(type) {
	case float64:
		if n == 0 {
			return ""
		}
		return fmt.Sprintf("%.2f", n)
	case int64:
		return strconv.FormatInt(n, 10)
	case int:
		return strconv.Itoa(n)
	case string:
		if n == "" {
			return ""
		}
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return n
		}
		return fmt.Sprintf("%.2f", f)
	}
	return fmt.Sprintf("%v", v)
}

func jsonEncode(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return template.JS(b)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func isSelected(val interface{}, target string) bool {
	return fmt.Sprintf("%v", val) == target
}

func defaultVal(val, def interface{}) interface{} {
	s := fmt.Sprintf("%v", val)
	if s == "" || s == "<nil>" {
		return def
	}
	return val
}

// --- PageData ---

type PageData struct {
	Title        string
	HasDB        bool
	Catalogs     map[string][]CatalogoItem
	ActiveModule string
	Modulos      map[string]ModuloConfig
	Filas        []Row
	PageSize     int
	TotalPages   int
	CurrentPage  int
	SortColumn   string
	SortDir      string
	DBPath       string
	Registro     Row
}

func (h *TemplateHandler) preparePageData(r *http.Request) *PageData {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}

	data := &PageData{
		Title:        "Control de Documentos",
		ActiveModule: modulo,
		Modulos:      Modulos,
		PageSize:     10,
		CurrentPage:  1,
		SortColumn:   "fecha_creacion",
		SortDir:      "DESC",
	}

	data.HasDB = h.app.db != nil
	data.DBPath = h.app.dbPath

	if !data.HasDB {
		return data
	}

	catalogs, err := h.app.ObtenerCatalogos()
	if err != nil {
		log.Printf("preparePageData: error catalogs: %v", err)
		catalogs = make(map[string][]CatalogoItem)
	}
	data.Catalogs = catalogs

	cfg, ok := Modulos[modulo]
	if !ok {
		modulo = "expedientes"
		cfg = Modulos[modulo]
	}
	data.ActiveModule = modulo
	data.SortColumn = cfg.IDColumna

	filas, err := h.app.ObtenerFilas(modulo, cfg.IDColumna+" DESC")
	if err != nil {
		log.Printf("preparePageData: error filas: %v", err)
	}
	data.Filas = filas

	if len(filas) > 0 {
		data.TotalPages = (len(filas) + data.PageSize - 1) / data.PageSize
		if data.TotalPages < 1 {
			data.TotalPages = 1
		}
	}

	return data
}

// --- ServeHTTP ---

func (h *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path

	// --- API routes ---
	switch {
	case p == "/api/guardar-expediente" && r.Method == http.MethodPost:
		h.handleGuardarExpediente(w, r)
		return
	case p == "/api/eliminar-expediente" && r.Method == http.MethodPost:
		h.handleEliminarExpediente(w, r)
		return
	case p == "/api/cargar-expediente" && r.Method == http.MethodGet:
		h.handleCargarExpediente(w, r)
		return
	case p == "/api/filtrar-expedientes" && r.Method == http.MethodGet:
		h.handleFiltrarExpedientes(w, r)
		return
	case p == "/api/cambiar-modulo" && r.Method == http.MethodGet:
		h.handleCambiarModulo(w, r)
		return
	case p == "/api/historial" && r.Method == http.MethodGet:
		h.handleHistorial(w, r)
		return
	case p == "/api/abrir-bd" && r.Method == http.MethodPost:
		h.handleAbrirBD(w, r)
		return
	case p == "/api/ruta-procesos" && r.Method == http.MethodGet:
		h.handleRutaProcesos(w, r)
		return
	case p == "/api/pendientes" && r.Method == http.MethodGet:
		h.handlePendientes(w, r)
		return
	case p == "/api/guardar-catalogo" && r.Method == http.MethodPost:
		h.handleGuardarCatalogo(w, r)
		return
	case p == "/api/optimizar-bd" && r.Method == http.MethodPost:
		h.handleOptimizarBD(w, r)
		return
	case p == "/api/csv" && r.Method == http.MethodGet:
		h.handleCSV(w, r)
		return
	}

	// --- Page routes ---
	if p == "/" || p == "/index.html" {
		data := h.preparePageData(r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := h.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			log.Printf("template error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// --- Static files ---
	r2 := new(http.Request)
	*r2 = *r
	r2.URL.Path = strings.TrimPrefix(p, "/")
	r2.RequestURI = r2.URL.RequestURI()

	h.static.ServeHTTP(w, r2)
}

// --- API handlers ---

func (h *TemplateHandler) handleGuardarExpediente(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSONError(w, "error parseando formulario", http.StatusBadRequest)
		return
	}

	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}
	cfg, ok := Modulos[modulo]
	if !ok {
		writeJSONError(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	data := make(map[string]interface{})
	for _, col := range cfg.Columnas {
		val := r.FormValue(col)
		if val == "" {
			data[col] = nil
		} else {
			data[col] = val
		}
	}

	idStr := r.FormValue(cfg.IDColumna)
	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil && id > 0 {
			data[cfg.IDColumna] = id
		}
	}

	newID, err := h.app.GuardarFila(modulo, data)
	if err != nil {
		writeJSONError(w, "error al guardar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"success": true,
		"id":      newID,
		"message": "Registro guardado correctamente",
	})
}

func (h *TemplateHandler) handleEliminarExpediente(w http.ResponseWriter, r *http.Request) {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}
	cfg, ok := Modulos[modulo]
	if !ok {
		writeJSONError(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue(cfg.IDColumna)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, "ID inválido", http.StatusBadRequest)
		return
	}

	err = h.app.EliminarFila(modulo, id)
	if err != nil {
		writeJSONError(w, "error al eliminar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"success": true,
		"message": "Registro eliminado correctamente",
	})
}

func (h *TemplateHandler) handleCargarExpediente(w http.ResponseWriter, r *http.Request) {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}
	if _, ok := Modulos[modulo]; !ok {
		http.Error(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	idStr := r.URL.Query().Get("id")
	var registro Row
	if idStr != "" && idStr != "null" {
		id, err := strconv.Atoi(idStr)
		if err == nil && id > 0 {
			registro, _ = h.app.ObtenerFilaPorId(modulo, id)
		}
	}

	catalogs, err2 := h.app.ObtenerCatalogos()
	if err2 != nil {
		log.Printf("handleCargarExpediente: error catalogs: %v", err2)
		catalogs = make(map[string][]CatalogoItem)
	}

	data := map[string]interface{}{
		"Catalogs":     catalogs,
		"Registro":     registro,
		"Expediente":   registro,
		"ActiveModule": modulo,
	}

	tmplName := "form_" + modulo + ".html"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		log.Printf("render error for %s: %v", tmplName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) handleFiltrarExpedientes(w http.ResponseWriter, r *http.Request) {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}
	cfg, ok := Modulos[modulo]
	if !ok {
		http.Error(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
	sortCol := r.URL.Query().Get("sort")
	dir := r.URL.Query().Get("dir")

	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	if sortCol == "" {
		sortCol = cfg.IDColumna
	}

	filas, err := h.app.ObtenerFilas(modulo, sortCol+" "+dir)
	if err != nil {
		log.Printf("handleFiltrarExpedientes: error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var filtered []Row
	if q == "" {
		filtered = filas
	} else {
		for _, row := range filas {
			matches := false
			for _, val := range row {
				if val != nil {
					sVal := strings.ToLower(fmt.Sprintf("%v", val))
					if strings.Contains(sVal, q) {
						matches = true
						break
					}
				}
			}
			if matches {
				filtered = append(filtered, row)
			}
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"ActiveModule": modulo,
		"Filas":        filtered,
		"Modulos":      Modulos,
		"HasDB":        h.app.db != nil,
	}

	tmplName := "tabla_" + modulo + ".html"
	if err := h.tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		log.Printf("render error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) handleCambiarModulo(w http.ResponseWriter, r *http.Request) {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}
	cfg, ok := Modulos[modulo]
	if !ok {
		http.Error(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	filas, err := h.app.ObtenerFilas(modulo, cfg.IDColumna+" DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ActiveModule": modulo,
		"Filas":        filas,
		"Modulos":      Modulos,
		"HasDB":        h.app.db != nil,
	}

	tmplName := "tabla_" + modulo + ".html"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		log.Printf("handleCambiarModulo: error rendering template %s: %v", tmplName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) handleHistorial(w http.ResponseWriter, r *http.Request) {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = "expedientes"
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	rows, err := h.app.ObtenerHistorialFila(modulo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Rows         []Row
		ActiveModule string
	}{
		Rows:         rows,
		ActiveModule: modulo,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmplName := "historial.html"
	if err := h.tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		log.Printf("render error for %s: %v", tmplName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) handleAbrirBD(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	if path == "" {
		writeJSONError(w, "ruta vacía", http.StatusBadRequest)
		return
	}

	if err := h.app.AbrirBaseDatos(path); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"success": true,
		"path":    path,
		"message": "Base de datos abierta correctamente",
	})
}

func (h *TemplateHandler) handleRutaProcesos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.app.ObtenerRutaProcesos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "ruta_procesos.html", rows); err != nil {
		log.Printf("render error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) handlePendientes(w http.ResponseWriter, r *http.Request) {
	rows, err := h.app.ObtenerDocumentosPendientes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "pendientes.html", rows); err != nil {
		log.Printf("render error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) handleGuardarCatalogo(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSONError(w, "error parseando formulario", http.StatusBadRequest)
		return
	}

	tabla := r.FormValue("tabla")
	nombre := r.FormValue("nombre")
	extra := make(map[string]interface{})
	if col := r.FormValue("extra_col"); col != "" {
		extra["col"] = col
		extra["val"] = r.FormValue("extra_val")
	}

	id, err := h.app.GuardarNuevoCatalogo(tabla, nombre, extra)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"success": true,
		"id":      id,
		"message": "Registro agregado al catálogo",
	})
}

func (h *TemplateHandler) handleOptimizarBD(w http.ResponseWriter, r *http.Request) {
	if err := h.app.OptimizarBD(); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{
		"success": true,
		"message": "Base de datos optimizada",
	})
}

func (h *TemplateHandler) handleCSV(w http.ResponseWriter, r *http.Request) {
	data, err := h.app.ObtenerFilas("expedientes", "id_expediente DESC")
	if err != nil || len(data) == 0 {
		writeJSONError(w, "no hay datos", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="reporte_expedientes.csv"`)

	if len(data) == 0 {
		return
	}

	headers := make([]string, 0, len(data[0]))
	for k := range data[0] {
		headers = append(headers, k)
	}

	csv := strings.Join(headers, ",") + "\n"
	for _, row := range data {
		vals := make([]string, len(headers))
		for i, h := range headers {
			v := row[h]
			if v == nil {
				vals[i] = ""
			} else {
				s := fmt.Sprintf("%v", v)
				if strings.ContainsAny(s, ",\"\n") {
					s = "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
				}
				vals[i] = s
			}
		}
		csv += strings.Join(vals, ",") + "\n"
	}

	w.Write([]byte(csv))
}

// --- JSON helpers ---

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   msg,
	})
}
