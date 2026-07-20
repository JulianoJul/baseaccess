.PHONY: combine clean wails-install wails-build-linux wails-build-linux-prod wails-build-win wails-build wails-dev

combine:
	{ \
	  echo "=== go.mod ===" && cat go.mod && \
	  echo "" && echo "=== wails.json ===" && cat wails.json && \
	  echo "" && echo "=== main.go ===" && cat main.go && \
	  echo "" && echo "=== app.go ===" && cat app.go && \
	  echo "" && echo "=== handler.go ===" && cat handler.go && \
	  echo "" && echo "=== templates/new/index.html ===" && cat templates/new/index.html && \
	  echo "" && echo "=== templates/new/components.html ===" && cat templates/new/components.html && \
	  echo "" && echo "=== templates/new/form.html ===" && cat templates/new/form.html && \
	  echo "" && echo "=== templates/new/tabla.html ===" && cat templates/new/tabla.html && \
	  echo "" && echo "=== templates/new/ruta_procesos.html ===" && cat templates/new/ruta_procesos.html && \
	  echo "" && echo "=== templates/historial.html ===" && cat templates/historial.html && \
	  echo "" && echo "=== templates/pendientes.html ===" && cat templates/pendientes.html && \
	  echo "" && echo "=== frontend/new/vendor/alpine-app.js ===" && cat frontend/new/vendor/alpine-app.js && \
	  echo "" && echo "=== frontend/new/vendor/alpine-directives.js ===" && cat frontend/new/vendor/alpine-directives.js && \
	  echo "" && echo "=== frontend/new/vendor/alpine-htmx-bridge.js ===" && cat frontend/new/vendor/alpine-htmx-bridge.js; \
	} > combined.txt
	@echo "combined.txt generado"

clean:
	rm -f combined.txt

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
