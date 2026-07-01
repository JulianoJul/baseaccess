.PHONY: clean commit push github combine

combine:
	{ \
	  echo "=== index.html ===" && cat index.html && \
	  echo "" && echo "=== Tablas6.sql ===" && cat Tablas6.sql && \
	  echo "" && echo "=== main.js ===" && cat main.js && \
	  echo "" && echo "=== package.json ===" && cat package.json && \
	  echo "" && echo "=== doc.md ===" && cat doc.md; \
	} > combined.txt
	@echo "combined.txt generado"

clean:
	rm -f combined.txt

commit:
	git add -A
	git commit -m "$(msg)"

push:
	git push

github: commit push
