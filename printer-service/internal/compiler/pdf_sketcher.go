package compiler

import (
	"log"
	"nantipov/printerservice/internal/printingdata"
	"os"
	"strings"

	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/pdf"
	"seehuhn.de/go/pdf/pagetree"
	"seehuhn.de/go/pdf/reader"
)

func sketchFromPdf(document printingdata.IppDocument) ([]*printingdata.SketchedPage, error) {
	pdfFile, fileErr := os.Open(document.Filename)

	if fileErr != nil {
		// log.Fatalf("could not open PDF file: %s", fileErr.Error())
		//todo: return error
		return nil, fileErr
	}

	defer pdfFile.Close()

	opt := &pdf.ReaderOptions{}
	pdfReader, pdfErr := pdf.NewReader(pdfFile, opt)
	if pdfErr != nil {
		return nil, pdfErr
	}

	//todo: deal with multiple pages
	firstPage := 0
	_, page, pageErr := pagetree.GetPage(pdfReader, firstPage)
	if pageErr != nil {
		return nil, pageErr
	}

	firstSketchedPage := printingdata.NewSketch()

	contentsReader := reader.New(pdfReader, nil)
	contentsReader.EveryOp = func(op string, args []pdf.Object) error {
		adoptPdfOp(firstSketchedPage, op, args)
		return nil
	}

	parseErr := contentsReader.ParsePage(page, matrix.Identity)
	if parseErr != nil {
		return nil, parseErr
	}

	// https://github.com/seehuhn/go-pdf/blob/main/examples/pdf-extract-text/main.go

	return []*printingdata.SketchedPage{firstSketchedPage}, nil
}

func adoptPdfOp(sketchedPage *printingdata.SketchedPage, op string, args []pdf.Object) {
	//todo: draw PDF structures in Canvas/Context
	s := ""
	if len(args) > 0 {
		writer := new(strings.Builder)
		pdf.Format(writer, pdf.OptPretty, args...)
		s = writer.String()
	}
	log.Printf("op = %s, args = %s", op, s)
}
