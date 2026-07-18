package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Pure functions ---

func TestSeq(t *testing.T) {
	got := seq(3)
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("seq(3) = %v", got)
	}
	if len(seq(0)) != 0 {
		t.Error("seq(0) should be empty")
	}
}

func TestSeqFromTo(t *testing.T) {
	tests := []struct {
		from, to int
		want     []int
	}{
		{1, 3, []int{1, 2, 3}},
		{5, 7, []int{5, 6, 7}},
		{5, 3, nil},
		{1, 1, []int{1}},
	}
	for _, tt := range tests {
		got := seqFromTo(tt.from, tt.to)
		if len(got) != len(tt.want) {
			t.Errorf("seqFromTo(%d,%d) len=%d, want %d", tt.from, tt.to, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("seqFromTo(%d,%d)[%d]=%d, want %d", tt.from, tt.to, i, got[i], tt.want[i])
			}
		}
	}
}

func TestDict(t *testing.T) {
	d := dict("a", 1, "b", "two")
	if d["a"] != 1 || d["b"] != "two" {
		t.Errorf("dict(a,1,b,two) = %v", d)
	}

	d = dict("a", 1, 2, "three") // odd: non-string key skipped
	if d["a"] != 1 || len(d) != 1 {
		t.Errorf("dict with non-string key: %v", d)
	}
}

func TestRowGet(t *testing.T) {
	r := Row{"name": "test", "age": 25}
	if rowGet(r, "name") != "test" {
		t.Error("rowGet name failed")
	}
	if rowGet(r, "missing") != nil {
		t.Error("rowGet missing should be nil")
	}
	if rowGet(nil, "any") != nil {
		t.Error("rowGet nil row should be nil")
	}
}

func TestRowGetStr(t *testing.T) {
	r := Row{"name": "test", "num": 42, "nil": nil}
	if rowGetStr(r, "name") != "test" {
		t.Error("rowGetStr name failed")
	}
	if rowGetStr(r, "num") != "42" {
		t.Errorf("rowGetStr num = %q", rowGetStr(r, "num"))
	}
	if rowGetStr(r, "nil") != "" {
		t.Error("rowGetStr nil should be empty")
	}
	if rowGetStr(r, "missing") != "" {
		t.Error("rowGetStr missing should be empty")
	}
	if rowGetStr(nil, "any") != "" {
		t.Error("rowGetStr nil row should be empty")
	}
}

func TestRowGetNum(t *testing.T) {
	r := Row{"f": 3.14, "i": int64(42), "s": "2.5", "nil": nil}
	if rowGetNum(r, "f") != 3.14 {
		t.Error("rowGetNum float64 failed")
	}
	if rowGetNum(r, "i") != 42 {
		t.Error("rowGetNum int64 failed")
	}
	if rowGetNum(r, "s") != 2.5 {
		t.Errorf("rowGetNum string = %f", rowGetNum(r, "s"))
	}
	if rowGetNum(r, "nil") != 0 {
		t.Error("rowGetNum nil should be 0")
	}
	if rowGetNum(r, "missing") != 0 {
		t.Error("rowGetNum missing should be 0")
	}
	if rowGetNum(nil, "any") != 0 {
		t.Error("rowGetNum nil row should be 0")
	}
}

func TestEstatusClass(t *testing.T) {
	tests := map[string]string{
		"":                        "bg-yellow-500/20 text-yellow-400",
		"FIRMADO":                 "bg-emerald-500/20 text-emerald-400",
		"firmado":                 "bg-emerald-500/20 text-emerald-400",
		"PENDIENTE":               "bg-yellow-500/20 text-yellow-400",
		"DEVUELTO PARA CORRECCIÓN": "bg-orange-500/20 text-orange-400",
		"DESCONOCIDO":             "bg-gray-500/20 text-gray-400",
	}
	for input, want := range tests {
		got := estatusClass(input)
		if got != want {
			t.Errorf("estatusClass(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFormatNumGo(t *testing.T) {
	tests := []struct {
		input interface{}
		want  string
	}{
		{nil, ""},
		{float64(0), ""},
		{float64(1234.56), "1.234,56"},
		{int64(5000000), "5.000.000,00"},
		{"", ""},
		{"0", ""},
		{"1234.56", "1.234,56"},
		{"notanumber", "notanumber"},
		{int(42), "42,00"},
		{true, "true"},
	}
	for _, tt := range tests {
		got := formatNumGo(tt.input)
		if got != tt.want {
			t.Errorf("formatNumGo(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseSpanishNumber(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1.500,75", "1500.75"},
		{"1,500.75", "1.50075"},         // removes all commas
		{"42", "42"},
		{"3,1416", "3.1416"},
		{"", ""},
		{"sin coma", "sin coma"},
	}
	for _, tt := range tests {
		got := parseSpanishNumber(tt.input)
		if got != tt.want {
			t.Errorf("parseSpanishNumber(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestJsonEncode(t *testing.T) {
	tests := []struct {
		input interface{}
		check string
	}{
		{map[string]string{"key": "value"}, `"key":"value"`},
		{[]int{1, 2, 3}, `[1,2,3]`},
		{map[string]string{"html": "</script>"}, `\u003c/script\u003e`},
	}
	for _, tt := range tests {
		got := string(jsonEncode(tt.input))
		if !strings.Contains(got, tt.check) {
			t.Errorf("jsonEncode(%v): expected to contain %q, got %q", tt.input, tt.check, got)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("truncate = %q, want hello...", got)
	}
	if got := truncate("hi", 10); got != "hi" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("", 5); got != "" {
		t.Errorf("truncate empty = %q", got)
	}
}

func TestIsSelected(t *testing.T) {
	if !isSelected(1, "1") {
		t.Error("isSelected(1, \"1\") should be true")
	}
	if isSelected(2, "1") {
		t.Error("isSelected(2, \"1\") should be false")
	}
	if !isSelected("text", "text") {
		t.Error("isSelected(\"text\", \"text\") should be true")
	}
}

func TestDefaultVal(t *testing.T) {
	if defaultVal(nil, "fallback") != "fallback" {
		t.Error("defaultVal(nil) failed")
	}
	if defaultVal("real", "fallback") != "real" {
		t.Error("defaultVal(real) failed")
	}
}

// --- Handler helpers ---

func TestModuloDesdeRequest_Default(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/registro", nil)
	modulo, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		t.Error("moduloDesdeRequest should return ok=true for default module")
	}
	if modulo != "expedientes" {
		t.Errorf("default module = %q, want expedientes", modulo)
	}
	if cfg.Tabla == "" {
		t.Error("cfg.Tabla should not be empty for expedientes")
	}
}

func TestModuloDesdeRequest_Explicit(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/registro?modulo=memorandums", nil)
	modulo, cfg, ok := moduloDesdeRequest(r)
	if !ok {
		t.Error("moduloDesdeRequest should return ok=true for memorandums")
	}
	if modulo != "memorandums" {
		t.Errorf("module = %q, want memorandums", modulo)
	}
	if cfg.Tabla == "" {
		t.Error("cfg.Tabla should not be empty")
	}
}

func TestModuloDesdeRequest_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/registro?modulo=inexistente", nil)
	_, _, ok := moduloDesdeRequest(r)
	if ok {
		t.Error("moduloDesdeRequest should return ok=false for invalid module")
	}
}


func TestUnifiedCatalogFilters_Coverage(t *testing.T) {
	if _, ok := UnifiedCatalogFilters["id_gerencia"]; !ok {
		t.Error("UnifiedCatalogFilters missing id_gerencia")
	}
	if _, ok := UnifiedCatalogFilters["id_estatus"]; !ok {
		t.Error("UnifiedCatalogFilters missing id_estatus")
	}
	if len(UnifiedCatalogFilters) != 12 {
		t.Errorf("UnifiedCatalogFilters has %d entries, expected 12", len(UnifiedCatalogFilters))
	}
}
