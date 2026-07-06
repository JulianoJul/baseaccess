.PHONY: clean commit push github combine serve electron-install electron-build

SCHEMA ?= data/sql/Tablas8.sql

combine:
	{ \
	  echo "=== index.html ===" && cat src/index.html && \
	  echo "" && echo "=== schema-config.js ===" && cat src/schema-config.js && \
	  echo "" && echo "=== Tablas8.sql ===" && cat $(SCHEMA) && \
	  echo "" && echo "=== main.js ===" && cat main.js && \
	  echo "" && echo "=== preload.js ===" && cat src/preload.js && \
	  echo "" && echo "=== package.json ===" && cat package.json && \
	  echo "" && echo "=== doc.md ===" && cat docs/doc.md && \
	  echo "" && echo "=== decisiones.md ===" && cat docs/decisiones.md && \
	  echo "" && echo "=== ai-context.md ===" && cat docs/ai-context.md && \
	  echo "" && echo "=== funciones.md ===" && cat docs/funciones.md && \
	  echo "" && echo "=== .clinerules ===" && cat .clinerules; \
	} > combined.txt
	@echo "combined.txt generado (schema: $(SCHEMA))"

clean:
	rm -f combined.txt

serve:
	python3 -m http.server 8000 --directory .

electron-install:
	npm install --save-dev electron@latest electron-builder@latest

electron-build-win:
	npm run build

electron-build-linux:
	npm run build:linux

electron-build: electron-build-linux

commit:
	git add -A
	git commit -m "$(msg)"

push:
	git push

github: commit push
