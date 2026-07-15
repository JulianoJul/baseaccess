.PHONY: clean wails-install wails-build-linux wails-build-linux-prod wails-build-win wails-build wails-dev

combine:
	{ \
	  echo "=== index.html ===" && cat templates/index.html && \
	  echo "" && echo "=== app.go ===" && cat app.go && \
	  echo "" && echo "=== handler.go ===" && cat handler.go && \
	  echo "" && echo "=== main.go ===" && cat main.go && \
	  echo "" && echo "=== go.mod ===" && cat go.mod && \
	  echo "" && echo "=== wails.json ===" && cat wails.json && \
	  echo "" && echo "=== doc.md ===" && cat docs/doc.md && \
	  echo "" && echo "=== decisiones.md ===" && cat docs/decisiones.md && \
	  echo "" && echo "=== ai-context.md ===" && cat docs/ai-context.md && \
	  echo "" && echo "=== funciones.md ===" && cat docs/funciones.md; \
	} > combined.txt
	@echo "combined.txt generado"

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