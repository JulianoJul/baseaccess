.PHONY: clean wails-install wails-build-linux wails-build-linux-prod wails-build-win wails-build wails-dev

SQL_FILES := $(filter-out data/sql/Tablas8.sql%,$(wildcard data/sql/*.sql))

combine:
	{ \
	  echo "=== go.mod ===" && cat go.mod && \
	  echo "" && echo "=== go.sum ===" && cat go.sum && \
	  echo "" && echo "=== wails.json ===" && cat wails.json && \
	  echo "" && echo "=== main.go ===" && cat main.go && \
	  echo "" && echo "=== app.go ===" && cat app.go && \
	  echo "" && echo "=== handler.go ===" && cat handler.go && \
	  echo "" && echo "=== templates/index.html ===" && cat templates/index.html && \
	  echo "" && echo "=== templates/historial.html ===" && cat templates/historial.html && \
	  echo "" && echo "=== templates/pendientes.html ===" && cat templates/pendientes.html && \
	  echo "" && echo "=== templates/ruta_procesos.html ===" && cat templates/ruta_procesos.html && \
	  echo "" && echo "=== templates/form_expedientes.html ===" && cat templates/form_expedientes.html && \
	  echo "" && echo "=== templates/tabla_expedientes.html ===" && cat templates/tabla_expedientes.html && \
	  echo "" && echo "=== templates/form_requisiciones.html ===" && cat templates/form_requisiciones.html && \
	  echo "" && echo "=== templates/tabla_requisiciones.html ===" && cat templates/tabla_requisiciones.html && \
	  echo "" && echo "=== templates/form_memorandums.html ===" && cat templates/form_memorandums.html && \
	  echo "" && echo "=== templates/tabla_memorandums.html ===" && cat templates/tabla_memorandums.html && \
	  echo "" && echo "=== templates/form_recobros.html ===" && cat templates/form_recobros.html && \
	  echo "" && echo "=== templates/tabla_recobros.html ===" && cat templates/tabla_recobros.html && \
	  echo "" && echo "=== templates/form_valuaciones.html ===" && cat templates/form_valuaciones.html && \
	  echo "" && echo "=== templates/tabla_valuaciones.html ===" && cat templates/tabla_valuaciones.html && \
	  echo "" && echo "=== templates/form_aprobacion_jd.html ===" && cat templates/form_aprobacion_jd.html && \
	  echo "" && echo "=== templates/tabla_aprobacion_jd.html ===" && cat templates/tabla_aprobacion_jd.html && \
	  echo "" && echo "=== templates/form_certificacion_bdu.html ===" && cat templates/form_certificacion_bdu.html && \
	  echo "" && echo "=== templates/tabla_certificacion_bdu.html ===" && cat templates/tabla_certificacion_bdu.html && \
	  echo "" && echo "=== templates/form_vacaciones.html ===" && cat templates/form_vacaciones.html && \
	  echo "" && echo "=== templates/tabla_vacaciones.html ===" && cat templates/tabla_vacaciones.html && \
	  echo "" && echo "=== templates/form_reposos_medicos.html ===" && cat templates/form_reposos_medicos.html && \
	  echo "" && echo "=== templates/tabla_reposos_medicos.html ===" && cat templates/tabla_reposos_medicos.html && \
	  echo "" && echo "=== docs/falsos_positivos.md ===" && cat docs/falsos_positivos.md; \
	  $(foreach f,$(SQL_FILES), echo "" && echo "=== $(f) ===" && cat $(f) &&) :; \
	} > combined.txt
	@echo "combined.txt generado (codigo + templates + SQL + falsos_positivos)"

clean:
	rm -f combined.txt

serve:
	@echo "Abriendo http://localhost:8000/frontend/index.html"
	@lsof -ti:8000 | xargs kill -9 2>/dev/null; sleep 0.5
	python3 -m http.server 8000 --directory .

# --- Wails ---
wails-install:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest

wails-build-linux:
	wails build -platform linux/amd64 -tags webkit2_41 -debug

wails-build-linux-prod:
	wails build -platform linux/amd64 -tags webkit2_41

wails-build-win:
	wails build -platform windows/amd64 -webview2 embed

wails-build: wails-build-linux

wails-dev:
	wails dev

# --- Git ---
commit:
	git add -A
	git commit -m "$(msg)"

push:
	git push

github: commit push