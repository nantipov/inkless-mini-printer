package web

import (
	"fmt"
	"log"
	"nantipov/printerservice/internal"
	"nantipov/printerservice/internal/ipp"
	"net/http"
)

func StartWebServer() {
	settings := internal.GetSettings()

	log.Printf("start web server on port %d", settings.WebPort)

	http.HandleFunc("/ipp/print", ipp.HandleIpp)
	http.ListenAndServe(fmt.Sprintf(":%d", settings.WebPort), nil)
}
