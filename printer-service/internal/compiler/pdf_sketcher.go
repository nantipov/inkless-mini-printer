package compiler

import (
	"image/color"
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

	"github.com/tdewolff/canvas/renderers"
)

type pdfContext struct {
	tm []float64
	ctm []float64 //todo: use canvas types?
	linePoint canvas.Point
	contentsReader *reader.Reader
}

var (
	defaultFontFamily *canvas.FontFamily = nil
)

const (
	resolutionDpi = 300.0
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
	pdfCtx := &pdfContext{
		contentsReader: contentsReader,
	}
	contentsReader.EveryOp = func(op string, args []pdf.Object) error {
		adoptPdfOp(firstSketchedPage, pdfCtx, op, args)
		return nil
	}

	parseErr := contentsReader.ParsePage(page, matrix.Identity)
	if parseErr != nil {
		return nil, parseErr
	}

	// https://github.com/seehuhn/go-pdf/blob/main/examples/pdf-extract-text/main.go

	// image := rasterizer.Draw(firstSketchedPage.Canvas, canvas.DefaultResolution, canvas.DefaultColorSpace)

	// f, _ := os.Create("image.png")
	// png.Encode(f, image)
	// defer f.Close()

	// firstSketchedPage.DrawContext.DrawPath(50.0, 50.0, canvas.Rectangle(50.0, 50.0))

	renderers.Write("image.png", firstSketchedPage.Canvas, canvas.DPI(resolutionDpi))

	return []*printingdata.SketchedPage{firstSketchedPage}, nil
}

func adoptPdfOp(sketchedPage *printingdata.SketchedPage, pdfc *pdfContext, op string, args []pdf.Object) {
	//todo: draw PDF structures in Canvas/Context
	printArgs(op, args)
	switch op {
	case "cm":
		pdfc.addCtm(args)
	case "Tm":
		pdfc.setTm(args)
	// case "Tf":
	//todo: calculate matrix
	case "Tj":
		x := (pdfc.ctm[4] * pdfc.tm[4]) / resolutionDpi // todo: read 72 from user-space
		y := (pdfc.ctm[5] * pdfc.tm[5]) / resolutionDpi // todo: read 72 from user-space
		size := pdfc.tm[0]

		face := getFontFace(pdfc.contentsReader.TextFont, size)

		textValue := ""

		if textPdfString, sok := args[0].(pdf.String); sok {
			textValue = string(textPdfString)
		} else {
			textWriter := new(strings.Builder)
			pdf.Format(textWriter, pdf.OptTextStringUtf8, args[0]) //todo: handle error
			textValue = textWriter.String()
		}

		textLine := canvas.NewTextLine(face, textValue, canvas.Top)
		
		sketchedPage.DrawContext.DrawText(x, y, textLine)
		// textLine.RenderAsPath(sketchedPage.Canvas, canvas.Identity, canvas.DPI(resolutionDpi)) //todo: resoultion
		log.Printf("text '%s' at (%f, %f), size %f", textValue, x, y, size)
	case "TJ":
		//todo duplicated code
		x := (pdfc.ctm[4] * pdfc.tm[4]) / resolutionDpi // todo: read 72 from user-space
		y := (pdfc.ctm[5] * pdfc.tm[5]) / resolutionDpi // todo: read 72 from user-space
		size := pdfc.tm[0]

		face := getFontFace(pdfc.contentsReader.TextFont, size)

	    textDataArray := args[0].(pdf.Array)
		for _, arg := range []pdf.Object(textDataArray) {
			textValue := ""
			if textPdfString, sok := arg.(pdf.String); sok {
				textValue = string(textPdfString)
			} else if textPdfInt, iok := arg.(pdf.Integer); iok {
				delta := int(textPdfInt)
				x = x + float64(delta) //todo: depends on text oritentation
				continue
			} else {
				textWriter := new(strings.Builder)
				pdf.Format(textWriter, pdf.OptTextStringUtf8, args[0]) //todo: handle error
				textValue = textWriter.String()
			}

			textLine := canvas.NewTextLine(face, textValue, canvas.Top)
			sketchedPage.DrawContext.DrawText(x, y, textLine)
			//textLine.RenderAsPath(sketchedPage.Canvas, canvas.Identity, canvas.DPI(resolutionDpi)) //todo: resoultion
			log.Printf("text '%s' at (%f, %f), size %f", textValue, x, y, size)
		}
	case "m":
		pdfc.setLinePoint(args)
		sketchedPage.DrawContext.MoveTo(pdfc.linePoint.X / resolutionDpi, pdfc.linePoint.Y / resolutionDpi)
	case "l":
		pdfc.setLinePoint(args)
		sketchedPage.DrawContext.LineTo(pdfc.linePoint.X / resolutionDpi, pdfc.linePoint.Y / resolutionDpi)
	}
}

func printArgs(op string, args []pdf.Object) {
	s := ""
	if len(args) > 0 {
		writer := new(strings.Builder)
		pdf.Format(writer, pdf.OptPretty, args...)
		s = writer.String()
	}
	log.Printf("op = %s, args = %s", op, s)
}

func (pdfc *pdfContext) addCtm(args []pdf.Object) {
	if len(pdfc.ctm) == 0 {
		pdfc.ctm = make([]float64, len(args))
	}
	for i, arg := range args {
		if tr, trok := arg.(pdf.Real); trok {
			pdfc.ctm[i] = pdfc.ctm[i] + float64(tr)
		} else if ti, tiok := arg.(pdf.Integer); tiok {
			pdfc.ctm[i] = pdfc.ctm[i] + float64(ti)
		}
	}
}

func (pdfc *pdfContext) setTm(args []pdf.Object) {
//	printArgs(">>> tm", args)
	pdfc.tm = make([]float64, len(args))
	for i, arg := range args {
		if tr, trok := arg.(pdf.Real); trok {
			pdfc.tm[i] = float64(tr)
		} else if ti, tiok := arg.(pdf.Integer); tiok {
			pdfc.tm[i] = float64(ti)
		} else {
			pdfc.tm[i] = 0.0
		}
	}
}

func (pdfc *pdfContext) setLinePoint(args []pdf.Object) {
	values := make([]float64, len(args))
	for i, arg := range args {
		if tr, trok := arg.(pdf.Real); trok {
			values[i] = float64(tr)
		} else if ti, tiok := arg.(pdf.Integer); tiok {
			values[i] = float64(ti)
		} else {
			values[i] = 0.0
		}
	}
	pdfc.linePoint.X = values[0]
	pdfc.linePoint.Y = values[1]
}


// todo: pre load fallback / default font family on application startup
// todo: cache already loaded fonts
func getFontFace(pdfFont font.Embedded, size float64) *canvas.FontFace {
	// family := canvas.NewFontFamily("f")

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
	// https://fonts.google.com/specimen/Spline+Sans+Mono/license?categoryFilters=Appearance:%2FMonospace%2FMonospace

	// err := family.LoadSystemFont("futura", canvas.FontRegular)
	// if err != nil {
	// 	log.Panic(err)
	// }

	return getDefaultFamily().Face(size, color.Black)
}

func getDefaultFamily() *canvas.FontFamily {
	//todo: concurrent?
	if defaultFontFamily != nil {
		return defaultFontFamily
	}

	family := canvas.NewFontFamily("default")
	//todo: read path from settings
	family.LoadFontFile("resources/SplineSansMono-Regular.ttf", canvas.FontRegular)

	defaultFontFamily = family
	return family
}
