.PHONY: clean commit push github combine serve electron-install electron-build

combine:
	{ \
	  echo "=== index.html ===" && cat index.html && \
	  echo "" && echo "=== Tablas7.sql ===" && cat Tablas7.sql && \
	  echo "" && echo "=== main.js ===" && cat main.js && \
	  echo "" && echo "=== package.json ===" && cat package.json && \
	  echo "" && echo "=== doc.md ===" && cat doc.md; \
	} > combined.txt
	@echo "combined.txt generado"

clean:
	rm -f combined.txt

serve:
	python3 -m http.server 8000

electron-install:
	npm install --save-dev --no-bin-links electron@latest electron-builder@latest

electron-build:
	node node_modules/electron-builder/cli.js --win portable --x64

commit:
	git add -A
	git commit -m "$(msg)"

push:
	git push

github: commit push
