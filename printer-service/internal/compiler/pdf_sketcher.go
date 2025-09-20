package compiler

import (
	"log"
	"nantipov/printerservice/internal/printingdata"
	"os"
	"strings"

	"github.com/tdewolff/canvas"
	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/pdf"
	"seehuhn.de/go/pdf/font"
	"seehuhn.de/go/pdf/font/dict"
	"seehuhn.de/go/pdf/pagetree"
	"seehuhn.de/go/pdf/reader"

	"image/png"

	"github.com/tdewolff/canvas/renderers/rasterizer"
)

type pdfBlock struct {
	tm        []pdf.Object
	tf        []pdf.Object
	parentBlk *pdfBlock
}

type pdfContext struct {
	currentBlk     *pdfBlock
	contentsReader *reader.Reader
}

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
	pdfCtx := &pdfContext{
		currentBlk:     nil,
		contentsReader: contentsReader,
	}
	// pdfCtx.enterBlock()
	contentsReader.EveryOp = func(op string, args []pdf.Object) error {
		adoptPdfOp(firstSketchedPage, pdfCtx, op, args)
		return nil
	}

	parseErr := contentsReader.ParsePage(page, matrix.Identity)
	if parseErr != nil {
		return nil, parseErr
	}

	// https://github.com/seehuhn/go-pdf/blob/main/examples/pdf-extract-text/main.go

	image := rasterizer.Draw(firstSketchedPage.Canvas, canvas.DefaultResolution, canvas.DefaultColorSpace)
	f, _ := os.Create("image.png")
	png.Encode(f, image)
	defer f.Close()

	return []*printingdata.SketchedPage{firstSketchedPage}, nil
}

func adoptPdfOp(sketchedPage *printingdata.SketchedPage, pdfc *pdfContext, op string, args []pdf.Object) {
	//todo: draw PDF structures in Canvas/Context
	s := ""
	if len(args) > 0 {
		writer := new(strings.Builder)
		pdf.Format(writer, pdf.OptPretty, args...)
		s = writer.String()
	}
	log.Printf("op = %s, args = %s", op, s)

	switch op {
	// case "BT":
	// 	pdfc.enterBlock()
	// case "ET":
	// 	pdfc.leaveBlock()
	// case "Tm":
	// 	pdfc.currentBlk.tm = args
	// case "Tf":
	// 	pdfc.currentBlk.tf = args
	case "Tj":
		// sketchedPage.Context.
		// pdfc.contentsReader.CurrentX
		// a, e := pdfc.contentsReader.TextFont.(font.FromFile)
		// face := canvas.
		face := getFontFace(pdfc.contentsReader.TextFont, pdfc.contentsReader.TextFontSize)
		textWriter := new(strings.Builder)
		pdf.Format(textWriter, pdf.OptPretty, args[0]) //todo: handle error
		textOpts := &canvas.TextOptions{}
		textBox := canvas.NewTextBox(face, textWriter.String(), 0, 0, canvas.Left, canvas.Top, textOpts)
		sketchedPage.DrawContext.DrawText(pdfc.contentsReader.CurrentX, pdfc.contentsReader.CurrentY, textBox)
	}
}

// func (pdfc *pdfContext) enterBlock() {
// 	newBlk := &pdfBlock{
// 		tm:        pdfc.currentBlk.tm,
// 		tf:        pdfc.currentBlk.tf,
// 		parentBlk: pdfc.currentBlk,
// 	}
// 	pdfc.currentBlk = newBlk
// }

// func (pdfc *pdfContext) leaveBlock() {
// 	pdfc.currentBlk = pdfc.currentBlk.parentBlk
// }


//todo: pre load fallback / default font family on application startup
//todo: cache already loaded fonts
func getFontFace(pdfFont font.Embedded, size float64) *canvas.FontFace {
	family := canvas.NewFontFamily("f")

	if fontFile, ok := pdfFont.(font.FromFile); ok {
		dictionary := fontFile.GetDict()
		fontInfo := dictionary.FontInfo()
		if emmFont, ok1 := fontInfo.(*dict.FontInfoGlyfEmbedded); ok1 {
			if emmFont != nil {
				// emmFont.
				//todo: load embedded font
			}
		}
	}

	// fallback here - default font?
	err := family.LoadSystemFont("futura", canvas.FontRegular)
	if err != nil {
		log.Panic(err)
	}

	return family.Face(size)
}
