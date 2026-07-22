//go:build !wails

package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	app := NewApp()
	handler, err := NewTemplateHandler(app)
	if err != nil {
		log.Fatalf("error creando template handler: %v", err)
	}

	log.Printf("BaseAccess HTTP server on http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
