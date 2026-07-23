package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

//go:embed all:frontend
var frontendFS embed.FS

//go:embed all:templates
var templateFS embed.FS

type TemplateHandler struct {
	app    *App
	tmpl   *template.Template
	static http.Handler
}

func (h *TemplateHandler) renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, tmplName, data); err != nil {
		log.Printf("render error for %s: %v", tmplName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}


func NewTemplateHandler(app *App) (*TemplateHandler, error) {
	funcMap := template.FuncMap{
		"add":        func(a, b int) int { return a + b },
		"sub":        func(a, b int) int { return a - b },
		"seq":        seq,
		"seqFromTo":  seqFromTo,
		"pagRange":   pagRange,
		"dict":       dict,
		"list":       list,
		"rowGet":     rowGet,
		"rowGetStr":  rowGetStr,
		"rowGetNum":  rowGetNum,
		"estatusClass": estatusClass,
		"formatNum":  formatNumGo,
		"jsonEncode": jsonEncode,
		"hasDB":      func() bool { app.mu.RLock(); defer app.mu.RUnlock(); return app.db != nil },
		"truncate":   truncate,
		"isSelected": isSelected,
		"default":    defaultVal,
		"excelOrder": func(modulo, col string) int {
			cfg, ok := Modulos[modulo]
			if !ok {
				return 999
			}
			viewAlias := map[string]string{
				"id_gerencia": "gerencia", "id_superintendencia": "superintendencia",
				"id_emisor": "emisor", "id_documento": "documento",
				"id_receptor": "receptor", "id_estatus": "estatus_detalle",
				"id_plan": "plan_contrataciones", "id_modalidad": "modalidad_contratacion",
				"id_art": "art", "id_tipo_contrato": "tipo_contrato",
				"id_resultado": "resultados_proceso", "id_empresa": "empresa_adjudicada",
			}
			if alias, ok := viewAlias[col]; ok {
				col = alias
			}
			for i, c := range cfg.OrdenExcel {
				if c == col {
					return i + 1
				}
			}
			return 999
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html", "templates/new/*.html", "templates/new/components/*.html")
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

const moduloDefault = "expedientes"

func modulosSinQueries() map[string]ModuloConfig {
	return Modulos
}

func moduloDesdeRequest(r *http.Request) (string, ModuloConfig, bool) {
	modulo := r.URL.Query().Get("modulo")
	if modulo == "" {
		modulo = moduloDefault
	}
	cfg, ok := Modulos[modulo]
	return modulo, cfg, ok
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

// pagRange devuelve una ventana de números de página alrededor de `cur`.
// window es el número de páginas a mostrar a cada lado de la página actual.
func pagRange(cur, total, window int) []int {
	if total <= 1 {
		return nil
	}
	start := cur - window
	if start < 1 {
		start = 1
	}
	end := cur + window
	if end > total {
		end = total
	}
	result := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		result = append(result, i)
	}
	return result
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

func list(values ...interface{}) []interface{} {
	return values
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
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			log.Printf("toFloat64: error parsing %q: %v", n, err)
		}
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
	var f float64
	switch n := v.(type) {
	case float64:
		if n == 0 {
			return ""
		}
		f = n
	case int64:
		f = float64(n)
	case int:
		f = float64(n)
	case string:
		if n == "" {
			return ""
		}
		parsed, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return n
		}
		if parsed == 0 {
			return ""
		}
		f = parsed
	default:
		return fmt.Sprintf("%v", v)
	}
	rounded := math.Round(f*100) / 100
	isNegative := rounded < 0
	if isNegative {
		rounded = -rounded
	}
	intPart := int64(rounded)
	decPart := int64((rounded-float64(intPart))*100 + 0.5)
	intStr := strconv.FormatInt(intPart, 10)
	var withDots strings.Builder
	for i, c := range intStr {
		if i > 0 && (len(intStr)-i)%3 == 0 {
			withDots.WriteByte('.')
		}
		withDots.WriteByte(byte(c))
	}
	result := withDots.String()
	if isNegative {
		result = "-" + result
	}
	return result + "," + fmt.Sprintf("%02d", decPart)
}

var columnasNumericas = map[string]bool{
	"presupuesto_base_usd": true, "presupuesto_base_bs": true, "tipo_cambio": true,
	"monto_adjudicado_usd": true, "monto_adjudicado_bs": true, "monto_valuacion": true,
	"presupuesto_base_total_usd": true, "monto_adjudicado_total_usd": true,
	"monto_contrato": true, "monto_ejecutado": true, "monto_pagado": true,
	"costo_servicio_usd": true, "nota_debito_reverso": true,
	"cantidad_frentes": true, "cantidad_dias": true, "dias_periodo": true,
	"tiempo_ejecucion": true, "anio": true,
}

func parseSpanishNumber(s string) string {
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.Replace(s, ",", ".", 1)
	}
	return s
}

func jsonEncode(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	s := string(b)
	s = strings.ReplaceAll(s, "</script>", "<\\/script>")
	s = strings.ReplaceAll(s, "<!--", "<\\!--")
	return template.JS(s)
}

func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n]) + "..."
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

type CatalogFilter struct {
	Label  string `json:"label"`
	Key    string `json:"key"`
	RowKey string `json:"-"`
}

var UnifiedCatalogFilters = map[string]CatalogFilter{
	"id_gerencia":         {"Gerencia", "gerencia", "gerencia"},
	"id_superintendencia": {"Superintendencia", "superintendencia", "superintendencia"},
	"id_documento":        {"Documento", "documento", "documento"},
	"id_plan":             {"Plan Contratación", "plan_contratacion", "plan_contrataciones"},
	"id_modalidad":        {"Modalidad", "modalidad", "modalidad_contratacion"},
	"id_art":              {"Art. Normativa", "art", "art"},
	"id_tipo_contrato":    {"Tipo Contrato", "tipo_contrato", "tipo_contrato"},
	"id_estatus":          {"Estatus Detalle", "estatus_detalle", "estatus_detalle"},
	"id_resultado":        {"Resultado Proceso", "resultado_proceso", "resultados_proceso"},
	"id_empresa":          {"Empresa", "empresas", "empresa_adjudicada"},
	"id_emisor":           {"Emisor / Remitente", "responsables", "emisor"},
	"id_receptor":         {"Receptor", "responsables", "receptor"},
}

// --- PageData ---

type PageData struct {
	Title          string
	HasDB          bool
	Catalogs       map[string][]CatalogoItem
	ActiveModule   string
	Modulos        map[string]ModuloConfig
	Filas          []Row
	Error          string
	PageSize       int
	TotalPages     int
	CurrentPage    int
	SortColumn     string
	SortDir        string
	DBPath         string
	Registro       Row
	CatalogFilters map[string]CatalogFilter
}

func (h *TemplateHandler) preparePageData(r *http.Request) *PageData {
	modulo, cfg, _ := moduloDesdeRequest(r)

	pagina := 1
	if p := r.URL.Query().Get("pagina"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			pagina = n
		}
	}

	data := &PageData{
		Title:          "App Control Documentos Presidencia",
		HasDB:          h.app.db != nil,
		DBPath:         h.app.dbPath,
		ActiveModule:   modulo,
		Modulos:        Modulos,
		PageSize:       10,
		CurrentPage:    pagina,
		TotalPages:     1,
		SortColumn:     "fecha_creacion",
		SortDir:        "DESC",
		CatalogFilters: UnifiedCatalogFilters,
	}

	h.app.mu.RLock()
	data.HasDB = h.app.db != nil
	data.DBPath = h.app.dbPath
	h.app.mu.RUnlock()

	if !data.HasDB {
		return data
	}

	catalogs, err := h.app.ObtenerCatalogos()
	if err != nil {
		log.Printf("preparePageData: error catalogs: %v", err)
		catalogs = make(map[string][]CatalogoItem)
	}
	data.Catalogs = catalogs

	data.ActiveModule = modulo
	data.SortColumn = cfg.IDColumna

	filas, totalPages, err := h.app.ObtenerFilasPaginado(modulo, cfg.IDColumna+" DESC", pagina, data.PageSize)
	if err != nil {
		log.Printf("preparePageData: error filas: %v", err)
		data.Error = "Error al cargar datos: " + err.Error()
	}
	data.Filas = filas
	data.CurrentPage = pagina
	data.TotalPages = totalPages

	return data
}

// --- ServeHTTP ---

func (h *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path

	// Prevenir cache para rutas dinámicas (API y páginas HTML)
	if p == "/" || p == "/index.html" || strings.HasPrefix(p, "/api/") {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}

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
	case p == "/api/exportar-excel" && r.Method == http.MethodGet:
		h.handleExportarExcel(w, r)
		return
	case p == "/api/columnas-modulo" && r.Method == http.MethodGet:
		h.handleColumnasModulo(w, r)
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
	case p == "/api/ruta-procesos-cronograma-guardar" && r.Method == http.MethodPost:
		h.handleGuardarCronogramaDia(w, r)
		return
	case p == "/api/ruta-procesos-cronograma-eliminar" && r.Method == http.MethodPost:
		h.handleEliminarCronogramaDia(w, r)
		return
	case p == "/api/ruta-procesos-hoja-crear" && r.Method == http.MethodPost:
		h.handleCrearRutaProcesoHoja(w, r)
		return
	case p == "/api/ruta-procesos-hoja-eliminar" && r.Method == http.MethodPost:
		h.handleEliminarRutaProcesoHoja(w, r)
		return
	case p == "/api/ruta-procesos-junta-crear" && r.Method == http.MethodPost:
		h.handleCrearJunta(w, r)
		return
	case p == "/api/ruta-procesos-junta-actualizar" && r.Method == http.MethodPost:
		h.handleActualizarJunta(w, r)
		return
	case p == "/api/ruta-procesos-junta-eliminar" && r.Method == http.MethodPost:
		h.handleEliminarJunta(w, r)
		return
	case p == "/api/ruta-procesos-semana-agregar" && r.Method == http.MethodPost:
		h.handleAgregarSemana(w, r)
		return
	case p == "/api/ruta-procesos-semana-eliminar" && r.Method == http.MethodPost:
		h.handleEliminarSemanas(w, r)
		return
	case p == "/api/ruta-procesos-proceso-agregar" && r.Method == http.MethodPost:
		h.handleAgregarProceso(w, r)
		return
	case p == "/api/ruta-procesos-proceso-eliminar" && r.Method == http.MethodPost:
		h.handleEliminarProceso(w, r)
		return
	case p == "/api/ruta-procesos-proceso-reordenar" && r.Method == http.MethodPost:
		h.handleReordenarProceso(w, r)
		return
	case p == "/api/ruta-procesos-leyenda-crear" && r.Method == http.MethodPost:
		h.handleCrearLeyenda(w, r)
		return
	case p == "/api/ruta-procesos-leyenda-actualizar" && r.Method == http.MethodPost:
		h.handleActualizarLeyenda(w, r)
		return
	case p == "/api/ruta-procesos-leyenda-eliminar" && r.Method == http.MethodPost:
		h.handleEliminarLeyenda(w, r)
		return
	case p == "/api/ruta-procesos-leyenda-reordenar" && r.Method == http.MethodPost:
		h.handleReordenarLeyenda(w, r)
		return
	case p == "/api/ruta-procesos-leyenda-bloquear" && r.Method == http.MethodPost:
		h.handleToggleBloquearLeyenda(w, r)
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
	}

	// --- Page routes ---
	if p == "/" || p == "/index.html" {
		data := h.preparePageData(r)
	h.renderTemplate(w, "index.html", data)
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

	modulo, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		writeJSONError(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	data := make(map[string]interface{})
	for _, col := range cfg.Columnas {
		if col == "id_documento" {
			docs := r.Form["id_documento"]
			if len(docs) > 0 {
				data["id_documento"] = docs
			} else {
				data["id_documento"] = nil
			}
			continue
		}
		v := r.FormValue(col)
		if v == "" {
			data[col] = nil
		} else if columnasNumericas[col] && strings.Contains(v, ",") {
			data[col] = parseSpanishNumber(v)
		} else {
			data[col] = v
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
	modulo, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		writeJSONError(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	idStr := r.PostFormValue(cfg.IDColumna)
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
	modulo, _, ok := moduloDesdeRequest(r)
	if !ok {
		http.Error(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	idStr := r.URL.Query().Get("id")
	var registro Row
	if idStr != "" && idStr != "null" {
		id, err := strconv.Atoi(idStr)
		if err == nil && id > 0 {
			registro, err = h.app.ObtenerFilaPorId(modulo, id)
			if err != nil {
				log.Printf("handleCargarExpediente: error obteniendo registro %d en %s: %v", id, modulo, err)
			}
		}
	}

	catalogs, err2 := h.app.ObtenerCatalogos()
	if err2 != nil {
		log.Printf("handleCargarExpediente: error catalogs: %v", err2)
		catalogs = make(map[string][]CatalogoItem)
	}

	cfg, _ := Modulos[modulo]
	if cfg.GerenciasIDs != nil {
		permitidas := map[int]bool{}
		for _, id := range cfg.GerenciasIDs {
			permitidas[id] = true
		}
		filtradasG := make([]CatalogoItem, 0, len(cfg.GerenciasIDs))
		for _, g := range catalogs["gerencia"] {
			if permitidas[g.ID] {
				filtradasG = append(filtradasG, g)
			}
		}
		catalogs["gerencia"] = filtradasG
		filtradasS := make([]CatalogoItem, 0)
		for _, s := range catalogs["superintendencia"] {
			if permitidas[s.IDGerencia] {
				filtradasS = append(filtradasS, s)
			}
		}
		catalogs["superintendencia"] = filtradasS
	}

	data := map[string]interface{}{
		"Catalogs":     catalogs,
		"Registro":     registro,
		"Expediente":   registro,
		"ActiveModule": modulo,
	}

	tmplName := "form.html"
	h.renderTemplate(w, tmplName, data)
}

func (h *TemplateHandler) handleFiltrarExpedientes(w http.ResponseWriter, r *http.Request) {
	modulo, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		http.Error(w, "modulo invalido", http.StatusBadRequest)
		return
	}
	q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
	sortCol := r.URL.Query().Get("sort")
	dir := strings.ToUpper(r.URL.Query().Get("dir"))
	pagina := 1
	pageSize := 10

	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	if sortCol == "" {
		sortCol = cfg.IDColumna
	}
	if p := r.URL.Query().Get("pagina"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			pagina = n
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 && n <= 100 {
			pageSize = n
		}
	}

	filas, err := h.app.ObtenerFilas(modulo, sortCol+" "+dir)
	if err != nil {
		log.Printf("handleFiltrarExpedientes: error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filas = h.filtrarPorGerencias(modulo, filas)

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

	total := len(filtered)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if pagina > totalPages {
		pagina = totalPages
	}

	start := (pagina - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pageRows := filtered[start:end]

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.app.mu.RLock()
	hasDB := h.app.db != nil
	h.app.mu.RUnlock()

	data := map[string]interface{}{
		"ActiveModule": modulo,
		"Filas":        pageRows,
		"Modulos":      modulosSinQueries(),
		"HasDB":        hasDB,
		"Q":            q,
		"SortColumn":   sortCol,
		"SortDir":      dir,
		"CurrentPage":  pagina,
		"TotalPages":   totalPages,
		"PageSize":     pageSize,
	}

	tmplName := "tabla.html"
	h.renderTemplate(w, tmplName, data)
}

func (h *TemplateHandler) handleCambiarModulo(w http.ResponseWriter, r *http.Request) {
	modulo, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		http.Error(w, "modulo invalido", http.StatusBadRequest)
		return
	}

	pagina := 1
	pageSize := 10
	if p := r.URL.Query().Get("pagina"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			pagina = n
		}
	}

	filas, totalPages, err := h.app.ObtenerFilasPaginado(modulo, cfg.IDColumna+" DESC", pagina, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filas = h.filtrarPorGerencias(modulo, filas)

	h.app.mu.RLock()
	hasDB := h.app.db != nil
	h.app.mu.RUnlock()
	data := map[string]interface{}{
		"ActiveModule": modulo,
		"Filas":        filas,
		"Modulos":      modulosSinQueries(),
		"HasDB":        hasDB,
		"Q":            "",
		"SortColumn":   cfg.IDColumna,
		"SortDir":      "DESC",
		"CurrentPage":  pagina,
		"TotalPages":   totalPages,
		"PageSize":     pageSize,
	}

	tmplName := "tabla.html"
	h.renderTemplate(w, tmplName, data)
}

func (h *TemplateHandler) handleHistorial(w http.ResponseWriter, r *http.Request) {
	modulo, _, ok := moduloDesdeRequest(r)
	if !ok {
		http.Error(w, "modulo invalido", http.StatusBadRequest)
		return
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
	h.renderTemplate(w, tmplName, data)
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
	idHoja, _ := strconv.Atoi(r.URL.Query().Get("hoja"))
	idJunta, _ := strconv.Atoi(r.URL.Query().Get("junta"))

	data, err := h.app.ObtenerRutaProcesosData(idHoja, idJunta)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, "ruta_procesos.html", data)
}

// --- Junta handlers ---

func (h *TemplateHandler) handleCrearJunta(w http.ResponseWriter, r *http.Request) {
	idHoja, _ := strconv.Atoi(r.FormValue("id_hoja"))
	numero, _ := strconv.Atoi(r.FormValue("numero"))
	consecutiva, _ := strconv.Atoi(r.FormValue("consecutiva"))
	fecha := r.FormValue("fecha")
	if idHoja == 0 || numero == 0 || fecha == "" {
		writeJSONError(w, "Faltan campos obligatorios", http.StatusBadRequest)
		return
	}
	id, err := h.app.CrearRutaProcesosJunta(idHoja, numero, consecutiva, fecha)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true, "id": id})
}

func (h *TemplateHandler) handleActualizarJunta(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	numero, _ := strconv.Atoi(r.FormValue("numero"))
	consecutiva, _ := strconv.Atoi(r.FormValue("consecutiva"))
	fecha := r.FormValue("fecha")
	if id == 0 || numero == 0 || fecha == "" {
		writeJSONError(w, "Faltan campos obligatorios", http.StatusBadRequest)
		return
	}
	if err := h.app.ActualizarRutaProcesosJunta(id, numero, consecutiva, fecha); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

func (h *TemplateHandler) handleEliminarJunta(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		writeJSONError(w, "id inválido", http.StatusBadRequest)
		return
	}
	if err := h.app.EliminarRutaProcesosJunta(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

// --- Hoja handlers ---

func (h *TemplateHandler) handleCrearRutaProcesoHoja(w http.ResponseWriter, r *http.Request) {
	nombre := strings.TrimSpace(r.FormValue("nombre"))
	if nombre == "" {
		writeJSONError(w, "Nombre requerido", http.StatusBadRequest)
		return
	}
	id, err := h.app.CrearRutaProcesosHoja(nombre)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true, "id": id})
}

func (h *TemplateHandler) handleEliminarRutaProcesoHoja(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		writeJSONError(w, "id_hoja inválido", http.StatusBadRequest)
		return
	}
	if err := h.app.EliminarRutaProcesosHoja(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

// --- Leyenda handlers ---

func (h *TemplateHandler) handleCrearLeyenda(w http.ResponseWriter, r *http.Request) {
	nombre := strings.TrimSpace(r.FormValue("nombre"))
	color := strings.TrimSpace(r.FormValue("color"))
	ambito := r.FormValue("ambito")
	if ambito == "" {
		ambito = "junta"
	}
	if nombre == "" || color == "" {
		writeJSONError(w, "Faltan campos (nombre, color)", http.StatusBadRequest)
		return
	}
	var idHoja *int
	if hVal := r.FormValue("id_hoja"); hVal != "" {
		if v, err := strconv.Atoi(hVal); err == nil {
			idHoja = &v
		}
	}
	var idJunta *int
	if jVal := r.FormValue("id_junta"); jVal != "" {
		if v, err := strconv.Atoi(jVal); err == nil {
			idJunta = &v
		}
	}
	id, err := h.app.CrearRutaProcesosLeyenda(nombre, color, ambito, idHoja, idJunta)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true, "id": id})
}

func (h *TemplateHandler) handleActualizarLeyenda(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	nombre := strings.TrimSpace(r.FormValue("nombre"))
	color := strings.TrimSpace(r.FormValue("color"))
	if id == 0 || nombre == "" || color == "" {
		writeJSONError(w, "Faltan campos", http.StatusBadRequest)
		return
	}
	if err := h.app.ActualizarRutaProcesosLeyenda(id, nombre, color); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

func (h *TemplateHandler) handleEliminarLeyenda(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		writeJSONError(w, "id inválido", http.StatusBadRequest)
		return
	}
	if err := h.app.EliminarRutaProcesosLeyenda(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

func (h *TemplateHandler) handleReordenarLeyenda(w http.ResponseWriter, r *http.Request) {
	idJunta, _ := strconv.Atoi(r.FormValue("id_junta"))
	idLeyenda, _ := strconv.Atoi(r.FormValue("id_leyenda"))
	direction, _ := strconv.Atoi(r.FormValue("direction"))
	if idJunta == 0 || idLeyenda == 0 || (direction != -1 && direction != 1) {
		writeJSONError(w, "Parámetros inválidos", http.StatusBadRequest)
		return
	}
	if err := h.app.ReordenarRutaProcesosLeyenda(idJunta, idLeyenda, direction); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

func (h *TemplateHandler) handleToggleBloquearLeyenda(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		writeJSONError(w, "id inválido", http.StatusBadRequest)
		return
	}
	if err := h.app.ToggleBloquearRutaProcesosLeyenda(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

// --- Semana handlers ---

func (h *TemplateHandler) handleAgregarSemana(w http.ResponseWriter, r *http.Request) {
	idJunta, _ := strconv.Atoi(r.FormValue("id_junta"))
	numero, _ := strconv.Atoi(r.FormValue("numero"))
	fechaInicio := r.FormValue("fecha_inicio")
	fechaFin := r.FormValue("fecha_fin")
	if idJunta == 0 || numero == 0 || fechaInicio == "" || fechaFin == "" {
		writeJSONError(w, "Faltan campos", http.StatusBadRequest)
		return
	}
	id, err := h.app.AgregarRutaProcesosSemana(idJunta, numero, fechaInicio, fechaFin)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true, "id": id})
}

func (h *TemplateHandler) handleEliminarSemanas(w http.ResponseWriter, r *http.Request) {
	idJunta, _ := strconv.Atoi(r.FormValue("id_junta"))
	numerosRaw := r.FormValue("numeros")

	var numeros []int
	if err := json.Unmarshal([]byte(numerosRaw), &numeros); err != nil {
		writeJSONError(w, "numeros debe ser un array JSON", http.StatusBadRequest)
		return
	}
	if idJunta == 0 || len(numeros) == 0 {
		writeJSONError(w, "Faltan datos", http.StatusBadRequest)
		return
	}
	if err := h.app.EliminarRutaProcesosSemanas(idJunta, numeros); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

// --- Proceso de Junta handlers ---

func (h *TemplateHandler) handleAgregarProceso(w http.ResponseWriter, r *http.Request) {
	idJunta, _ := strconv.Atoi(r.FormValue("id_junta"))
	numero, _ := strconv.Atoi(r.FormValue("numero"))
	proceso := strings.TrimSpace(r.FormValue("proceso"))
	if idJunta == 0 || proceso == "" {
		writeJSONError(w, "Faltan campos", http.StatusBadRequest)
		return
	}
	id, err := h.app.AgregarRutaProcesosProceso(idJunta, numero, proceso)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true, "id": id})
}

func (h *TemplateHandler) handleEliminarProceso(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		writeJSONError(w, "id inválido", http.StatusBadRequest)
		return
	}
	if err := h.app.EliminarRutaProcesosProceso(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

func (h *TemplateHandler) handleReordenarProceso(w http.ResponseWriter, r *http.Request) {
	idJunta, _ := strconv.Atoi(r.FormValue("id_junta"))
	idProceso, _ := strconv.Atoi(r.FormValue("id_proceso"))
	direction, _ := strconv.Atoi(r.FormValue("direction"))
	if idJunta == 0 || idProceso == 0 || (direction != -1 && direction != 1) {
		writeJSONError(w, "Parámetros inválidos", http.StatusBadRequest)
		return
	}
	if err := h.app.ReordenarRutaProcesosProceso(idJunta, idProceso, direction); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

// --- Cronograma (mismos endpoints, cambia id_proceso por id_junta_proceso) ---

func (h *TemplateHandler) handleGuardarCronogramaDia(w http.ResponseWriter, r *http.Request) {
	idProceso, _ := strconv.Atoi(r.FormValue("id_proceso"))
	fecha := r.FormValue("fecha")
	idLeyenda, _ := strconv.Atoi(r.FormValue("id_leyenda"))
	nota := r.FormValue("nota")
	if idProceso == 0 || fecha == "" || idLeyenda == 0 {
		writeJSONError(w, "Faltan datos", http.StatusBadRequest)
		return
	}
	if _, err := h.app.GuardarCronogramaDia(idProceso, fecha, idLeyenda, nota); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

func (h *TemplateHandler) handleEliminarCronogramaDia(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	if id == 0 {
		writeJSONError(w, "id inválido", http.StatusBadRequest)
		return
	}
	if err := h.app.EliminarCronogramaDia(id); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}

func (h *TemplateHandler) handlePendientes(w http.ResponseWriter, r *http.Request) {
	rows, err := h.app.ObtenerDocumentosPendientes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, "pendientes.html", rows)
}

func (h *TemplateHandler) handleGuardarCatalogo(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSONError(w, "error parseando formulario", http.StatusBadRequest)
		return
	}

	tabla := r.FormValue("tabla")
	nombre := strings.TrimSpace(r.FormValue("nombre"))
	if nombre == "" {
		writeJSONError(w, "nombre requerido", http.StatusBadRequest)
		return
	}
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

func (h *TemplateHandler) filtrarPorGerencias(modulo string, filas []Row) []Row {
	cfg, ok := Modulos[modulo]
	if !ok || cfg.GerenciasIDs == nil || len(cfg.GerenciasIDs) == 0 {
		return filas
	}

	catalogs, err := h.app.ObtenerCatalogos()
	if err != nil {
		return filas
	}

	catGer := catalogs["gerencia"]
	if catGer == nil {
		return filas
	}

	permitidasNames := map[string]bool{}
	for _, item := range catGer {
		for _, gid := range cfg.GerenciasIDs {
			if item.ID == gid {
				permitidasNames[item.Nombre] = true
				break
			}
		}
	}

	if len(permitidasNames) == 0 {
		return filas
	}

	filtered := make([]Row, 0, len(filas))
	for _, row := range filas {
		gerName, _ := row["gerencia"].(string)
		if gerName == "" || permitidasNames[gerName] {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func (h *TemplateHandler) filasParaExportar(r *http.Request) (cfg ModuloConfig, filas []Row, err error) {
	modulo, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		return cfg, nil, fmt.Errorf("modulo invalido")
	}

	filas, err = h.app.ObtenerFilas(modulo, cfg.IDColumna+" DESC")
	if err != nil {
		return cfg, nil, err
	}

	catalogs, cerr := h.app.ObtenerCatalogos()
	if cerr != nil {
		return cfg, nil, cerr
	}
	catMaps := make(map[string]map[string]string)
	for catKey, items := range catalogs {
		catMaps[catKey] = make(map[string]string)
		for _, item := range items {
			catMaps[catKey][strconv.Itoa(item.ID)] = item.Nombre
		}
	}

	if cfg.GerenciasIDs != nil && len(cfg.GerenciasIDs) > 0 {
		catGer := catMaps["gerencia"]
		permitidasNames := map[string]bool{}
		for _, id := range cfg.GerenciasIDs {
			if name, ok := catGer[strconv.Itoa(id)]; ok {
				permitidasNames[name] = true
			}
		}
		filteredByGer := make([]Row, 0, len(filas))
		for _, row := range filas {
			gerName, _ := row["gerencia"].(string)
			if gerName == "" || permitidasNames[gerName] {
				filteredByGer = append(filteredByGer, row)
			}
		}
		filas = filteredByGer
	}

	fechaDesde := r.URL.Query().Get("fecha_desde")
	fechaHasta := r.URL.Query().Get("fecha_hasta")

	filters := make(map[string]string)
	for k, v := range r.URL.Query() {
		if strings.HasPrefix(k, "id_") && len(v) > 0 && v[0] != "" {
			filters[k] = v[0]
		}
	}

	var filtered []Row
	for _, row := range filas {
		fr, _ := row[cfg.FechaColumna].(string)
		if (fechaDesde != "" || fechaHasta != "") && fr == "" {
			continue
		}
		if fechaDesde != "" && fr != "" && fr < fechaDesde {
			continue
		}
		if fechaHasta != "" && fr != "" && fr > fechaHasta {
			continue
		}

		match := true
		for paramKey, paramVal := range filters {
			mapping, ok := UnifiedCatalogFilters[paramKey]
			if !ok {
				match = false
				break
			}
			rowKey, catKey := mapping.RowKey, mapping.Key
			expectedName := catMaps[catKey][paramVal]
			if expectedName == "" {
				match = false
				break
			}
			rowVal, exists := row[rowKey]
			if !exists {
				match = false
				break
			}
			rowValStr, _ := rowVal.(string)
			if strings.ToLower(rowValStr) != strings.ToLower(expectedName) {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		filtered = append(filtered, row)
	}

	return cfg, filtered, nil
}

func (h *TemplateHandler) handleExportarExcel(w http.ResponseWriter, r *http.Request) {
	cfg, filas, err := h.filasParaExportar(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	columnasParam := r.URL.Query().Get("columnas")
	var columnasSel []string
	if columnasParam != "" {
		columnasSel = strings.Split(columnasParam, ",")
	}

	if len(filas) == 0 {
		http.Error(w, "No hay datos para exportar con los filtros aplicados", http.StatusBadRequest)
		return
	}

	keysOrdered := make([]string, 0, len(filas[0]))
	// Build column order: OrdenExcel first, then any extra columns sorted alphabetically
	excelKeys := map[string]bool{}
	for _, k := range cfg.OrdenExcel {
		if _, ok := filas[0][k]; ok {
			keysOrdered = append(keysOrdered, k)
			excelKeys[k] = true
		}
	}
	var extra []string
	for k := range filas[0] {
		if !excelKeys[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	keysOrdered = append(keysOrdered, extra...)

	// Filter out backend-only columns (used for ordering, not for export)
	filteredKeys := keysOrdered[:0]
	for _, k := range keysOrdered {
		if k != "fecha_creacion" && k != "fecha_actualizacion" {
			filteredKeys = append(filteredKeys, k)
		}
	}
	keysOrdered = filteredKeys

	if len(columnasSel) > 0 {
		sel := map[string]bool{}
		for _, c := range columnasSel {
			sel[c] = true
		}
		filteredKeys := make([]string, 0, len(columnasSel))
		for _, k := range keysOrdered {
			if sel[k] {
				filteredKeys = append(filteredKeys, k)
			}
		}
		if len(filteredKeys) > 0 {
			keysOrdered = filteredKeys
		}
	}

	labelOf := func(k string) string {
		words := strings.Split(strings.ReplaceAll(k, "_", " "), " ")
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		return strings.Join(words, " ")
	}

	f := excelize.NewFile()
	defer f.Close()
	sheetName := cfg.Nombre
	if len(sheetName) > 31 {
		sheetName = sheetName[:31]
	}
	f.SetSheetName(f.GetSheetName(0), sheetName)

	for i, k := range keysOrdered {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, labelOf(k))
	}
	styleID, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#0F766E"}},
	})
	lastCell, _ := excelize.CoordinatesToCellName(len(keysOrdered), 1)
	f.SetCellStyle(sheetName, "A1", lastCell, styleID)

	for ri, row := range filas {
		for ci, k := range keysOrdered {
			cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
			v := row[k]
			if v == nil {
				continue
			}
			f.SetCellValue(sheetName, cell, v)
		}
	}

	for i, k := range keysOrdered {
		width := float64(utf8.RuneCountInString(labelOf(k))) + 4
		for _, row := range filas {
			if v := row[k]; v != nil {
				s := fmt.Sprintf("%v", v)
				if utf8.RuneCountInString(s) > int(width) {
					width = float64(utf8.RuneCountInString(s)) + 2
				}
			}
		}
		colName, err := excelize.ColumnNumberToName(i + 1)
		if err == nil {
			f.SetColWidth(sheetName, colName, colName, width)
		}
	}

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze: true, YSplit: 1,
		TopLeftCell: "A2", ActivePane: "bottomLeft",
	})
	endCell, _ := excelize.CoordinatesToCellName(len(keysOrdered), len(filas)+1)
	f.AutoFilter(sheetName, "A1:"+endCell, []excelize.AutoFilterOptions{})

	filename := cfg.Nombre + "_" + time.Now().Format("2006-01-02") + ".xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	if err := f.Write(w); err != nil {
		log.Printf("exportar-excel: error escribiendo: %v", err)
	}
}

func (h *TemplateHandler) handleColumnasModulo(w http.ResponseWriter, r *http.Request) {
	_, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		writeJSONError(w, "modulo invalido", http.StatusBadRequest)
		return
	}
	viewCols, err := h.app.ObtenerColumnasVista(cfg.Vista)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{
		"view_cols":  viewCols,
		"table_cols": cfg.Columnas,
	})
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
