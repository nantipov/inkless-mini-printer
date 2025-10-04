package compiler

import (
	"image"
	"nantipov/printerservice/internal/printingdata"

	"github.com/tdewolff/canvas"
)

type Translator struct {
	page *printingdata.SketchedPage
	currentX, currentY float64
}

func translateToGcode(pages []*printingdata.SketchedPage) ([]*printingdata.SketchedPage, error) {
	outoutPages := make([]*printingdata.SketchedPage, len(pages))
	for i, page := range pages {
		outputPage, _ := translatePage(page) //todo: handle/propagate error
		outoutPages[i] = outputPage
	}

	return outoutPages, nil
}

func translatePage(page *printingdata.SketchedPage) (*printingdata.SketchedPage, error)  {

	// page.Canvas.RenderImage()
	//todo: define renderer, start redendering with resolution - here it is a point of converting canvas coordinates into device coordinates
	//todo: use canvas traverse into gcode

	t := Translator{
		page: page,
	}
	page.Canvas.RenderTo(t)

	return page, nil
}


func (t Translator) Size() (float64, float64) {
	return t.page.Canvas.Size() //todo traslate into device coordinates?
}

func (t Translator) RenderPath(path *canvas.Path, style canvas.Style, m canvas.Matrix) {
	flattenPath := path.Flatten(0.05) // convert evertything into lines
	scanner := flattenPath.Scanner()
	for {
		if !scanner.Scan() {
			break
		}

		scanner.Cmd()
	}
}

func (t Translator) RenderText(text *canvas.Text, m canvas.Matrix) {
	// ignore
}

func (t Translator) RenderImage(img image.Image, m canvas.Matrix) {
	// ignore
}
