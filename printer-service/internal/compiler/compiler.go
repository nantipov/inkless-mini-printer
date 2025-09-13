package compiler

import (
	"log"
	"nantipov/printerservice/internal/printingdata"

	"seehuhn.de/go/pdf"
)

//todo: naming

// todo: consider move structs to "domain" package
// todo: combine Job and MachineTask?
func Compile(job *printingdata.Job) {

	//todo: interate over documents in the job
	//todo: depending on mime type, consider different readers

	opt := &pdf.ReaderOptions{}
	reader, err := pdf.Open("/tmp/f.pdf", opt) //todo: read from Job/Task

	if err != nil {
		log.Fatalf("could not open PDF file: %s", err.Error())
		//todo: store error to Job/Task and switch processing state
	}

	defer reader.Close()

	// reader.GetMeta().Catalog.Pages
	// https://github.com/seehuhn/go-pdf/blob/main/examples/pdf-extract-text/main.go

	job.SetState(printingdata.JobStateCompiled)
}
