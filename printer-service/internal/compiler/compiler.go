package compiler

import (
	"errors"
	"nantipov/printerservice/internal/printingdata"
)

var (
	ErrorNoSuitableSketcher = errors.New("unknown document type")
)

//todo: naming

// todo: consider move structs to "domain" package
// todo: combine Job and MachineTask?
func Compile(job *printingdata.Job) {

	//todo: interate over documents in the job
	//todo: depending on mime type, consider different readers

	for _, document := range job.Ipp.Documents {
		pages, _ := sketch(document) // todo: handle error
		pages, _ = translateToGcode(pages)
		job.Pages = append(job.Pages, pages...)
	}

	job.SetState(printingdata.JobStateCompiled)
}

func sketch(document printingdata.IppDocument) ([]*printingdata.SketchedPage, error) {
	switch document.DocumentType {
	case printingdata.DocumentTypePDF:
		return sketchFromPdf(document)
	case printingdata.DocumentTypeRaster:
		return sketchFromRaster(document)
	default:
		return nil, ErrorNoSuitableSketcher
	}
	//todo: check if a sketcher should be an interface?
}
