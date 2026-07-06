.PHONY: clean commit push github combine serve electron-install electron-build

SCHEMA ?= bdd/Tablas8.sql

combine:
	{ \
	  echo "=== index.html ===" && cat index.html && \
	  echo "" && echo "=== schema-config.js ===" && cat schema-config.js && \
	  echo "" && echo "=== Tablas8.sql ===" && cat $(SCHEMA) && \
	  echo "" && echo "=== main.js ===" && cat main.js && \
	  echo "" && echo "=== preload.js ===" && cat preload.js && \
	  echo "" && echo "=== package.json ===" && cat package.json && \
	  echo "" && echo "=== doc.md ===" && cat doc.md; \
	} > combined.txt
	@echo "combined.txt generado (schema: $(SCHEMA))"

clean:
	rm -f combined.txt

serve:
	python3 -m http.server 8000

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
