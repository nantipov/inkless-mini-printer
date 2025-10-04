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

type pdfGraphicsState struct {
	ctm       canvas.Matrix
	prevState *pdfGraphicsState
}

type pdfContext struct {
	graphicsState      *pdfGraphicsState
	linePoint          canvas.Point
	tm                 canvas.Matrix
	textSize           float64
	textFont           string
	textCharacterSpace float64
	contentsReader     *reader.Reader
}

const (
	resolutionDpiPDF = 72.0 // default user-space 72?
	//todo: read user-space from the document
)

var (
	defaultFontFamily   *canvas.FontFamily = nil
	resolutionDpiOutput                    = 72.0 // canvas.DefaultResolution.DPI() // default user-space
	resolutionOutput                       = canvas.DPI(resolutionDpiOutput)
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
		graphicsState: &pdfGraphicsState{
			ctm:       canvas.Identity,
			prevState: nil,
		},
		tm:                 canvas.Identity,
		textSize:           1.0,
		textCharacterSpace: 1.0,
		contentsReader:     contentsReader,
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

	renderers.Write("image.png", firstSketchedPage.Canvas, resolutionOutput)

	return []*printingdata.SketchedPage{firstSketchedPage}, nil
}

func adoptPdfOp(sketchedPage *printingdata.SketchedPage, pdfc *pdfContext, op string, args []pdf.Object) {
	printArgs(op, args)
	switch op {
	case "q":
		pdfc.pushGS()
	case "Q":
		pdfc.popGS()
	case "cm":
		pdfc.graphicsState.ctm = pdfc.graphicsState.ctm.Mul(toCanvasMatrix(args))
	case "Tm":
		pdfc.tm = toCanvasMatrix(args)
	case "Tf":
		pdfc.textFont = string(args[0].(pdf.Name))
		pdfc.textSize = toFloatNumbers(args)[1]
	case "Tc":
		//todo is it already read by the PDF library
		pdfc.textCharacterSpace = toFloatNumbers(args)[0]
	case "Tj":
		// x := (pdfc.ctm[4] * pdfc.tm[4]) / resolutionDpi // todo: read 72 from user-space
		// y := (pdfc.ctm[5] * pdfc.tm[5]) / resolutionDpi // todo: read 72 from user-space

		// resultTm := pdfc.graphicsState.ctm.Mul(pdfc.tm)
		// x, y := resultTm.Pos()
		// size := resultTm[0][0] * pdfc.textSize / resolutionDpiPDF * resolutionDpiOutput

		// x = x / resolutionDpiPDF * resolutionDpiOutput
		// y = y / resolutionDpiPDF * resolutionDpiOutput

		// x = x * resolutionDpiOutput
		// y = y * resolutionDpiOutput

		s := pdfc.contentsReader
		M := matrix.Matrix{s.TextFontSize * s.TextHorizontalScaling, 0, 0, s.TextFontSize, 0, s.TextRise}
		M = M.Mul(s.TextMatrix)
		M = M.Mul(s.CTM)

		x, y := pdfc.contentsReader.GetTextPositionDevice()
		size := pdfc.contentsReader.TextFontSize * M[0]

		face := getFontFace(pdfc.contentsReader.TextFont, size, color.Black)

		textValue := ""

		if textPdfString, sok := args[0].(pdf.String); sok {
			textValue = string(textPdfString)
		} else {
			textWriter := new(strings.Builder)
			pdf.Format(textWriter, pdf.OptTextStringUtf8, args[0]) //todo: handle error
			textValue = textWriter.String()
		}

		textLine := canvas.NewTextLine(face, textValue, canvas.Top)
		// sketchedPage.DrawContext.DrawText(x, y, textLine)
		textLine.RenderAsPath(sketchedPage.Canvas, pdfMatrixToCanvasMatrix(M), resolutionOutput) //todo: resoultion
		log.Printf("text '%s' at (%f, %f), size %f", textValue, x, y, size)
	case "TJ":
		//todo duplicated code
		// x := (pdfc.ctm[4] * pdfc.tm[4]) / resolutionDpi // todo: read 72 from user-space
		// y := (pdfc.ctm[5] * pdfc.tm[5]) / resolutionDpi // todo: read 72 from user-space

		// resultTm := pdfc.graphicsState.ctm.Mul(pdfc.tm)
		// x, y := resultTm.Pos()
		// size := resultTm[0][0] * pdfc.textSize / resolutionDpiPDF * resolutionDpiOutput

		// y = y / resolutionDpiPDF * resolutionDpiOutput
		// y = y * resolutionDpiOutput

		s := pdfc.contentsReader
		M := matrix.Matrix{s.TextFontSize * s.TextHorizontalScaling, 0, 0, s.TextFontSize, 0, s.TextRise}
		M = M.Mul(s.TextMatrix)
		M = M.Mul(s.CTM)

		x, y := pdfc.contentsReader.GetTextPositionDevice()
		size := pdfc.contentsReader.TextFontSize * M[0]

		textDataArray := args[0].(pdf.Array)
		for _, arg := range []pdf.Object(textDataArray) {
			textValue := ""
			if textPdfString, sok := arg.(pdf.String); sok {
				textValue = string(textPdfString)
			} else if textPdfInt, iok := arg.(pdf.Integer); iok {
				delta := int(textPdfInt)
				// x = x + float64(delta)*pdfc.textCharacterSpace //todo: depends on text oritentation
				x = x + float64(delta)*pdfc.contentsReader.TextCharacterSpacing*10.0 //todo: depends on text oritentation
				continue
			} else {
				textWriter := new(strings.Builder)
				pdf.Format(textWriter, pdf.OptTextStringUtf8, args[0]) //todo: handle error
				textValue = textWriter.String()
			}

			// x0 := x / resolutionDpiPDF * resolutionDpiOutput
			x0 := x
			M[4] = x0
			// x0 := x * resolutionDpiOutput

			fontColor := color.RGBA{139, 44, 201, 255}
			if textValue == "PDF" {
				fontColor = color.RGBA{44, 78, 201, 255}
			}

			// size = 50.0
			// textValue = "a"

			face := getFontFace(pdfc.contentsReader.TextFont, size, fontColor)

			textLine := canvas.NewTextLine(face, textValue, canvas.Left)
			// sketchedPage.DrawContext.DrawText(x0, y, textLine)
			textLine.RenderAsPath(sketchedPage.Canvas, pdfMatrixToCanvasMatrix(M), resolutionOutput) //todo: resoultion
			r := canvas.Rectangle(size, size)
			sketchedPage.DrawContext.SetStrokeColor(fontColor)
			sketchedPage.DrawContext.DrawPath(x0, y, r)
			log.Printf("text '%s' at (%f, %f), size %f", textValue, x0, y, size)
		}
	case "m":
		pdfc.setLinePoint(args)
		sketchedPage.DrawContext.MoveTo(pdfc.linePoint.X/resolutionDpiPDF, pdfc.linePoint.Y/resolutionDpiPDF)
	case "l":
		pdfc.setLinePoint(args)
		sketchedPage.DrawContext.LineTo(pdfc.linePoint.X/resolutionDpiPDF, pdfc.linePoint.Y/resolutionDpiPDF)
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

func toFloatNumbers(args []pdf.Object) []float64 {
	m := make([]float64, len(args))
	for i, arg := range args {
		if tr, trok := arg.(pdf.Real); trok {
			m[i] = float64(tr)
		} else if ti, tiok := arg.(pdf.Integer); tiok {
			m[i] = float64(ti)
		}
	}
	return m
}

func toCanvasMatrix(args []pdf.Object) canvas.Matrix {
	m := toFloatNumbers(args)
	m1 := canvas.Matrix{
		{m[0], m[1], m[4]},
		{m[2], m[3], m[5]},
	}
	return m1
}

func pdfMatrixToCanvasMatrix(pdfMatrix matrix.Matrix) canvas.Matrix {
	m := pdfMatrix
	m1 := canvas.Matrix{
		{m[0], m[1], m[4]},
		{m[2], m[3], m[5]},
	}
	return m1
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
func getFontFace(pdfFont font.Embedded, size float64, fontColor color.Color) *canvas.FontFace {
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

	return getDefaultFamily().Face(size, fontColor, canvas.FontRegular)
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

func (pdfc *pdfContext) pushGS() {
	newgs := &pdfGraphicsState{
		ctm:       canvas.Identity, // pdfc.graphicsState.ctm,
		prevState: pdfc.graphicsState,
	}
	pdfc.graphicsState = newgs
}

func (pdfc *pdfContext) popGS() {
	pdfc.graphicsState = pdfc.graphicsState.prevState
}
