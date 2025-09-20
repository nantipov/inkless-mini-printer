package printingdata

import (
	"github.com/tdewolff/canvas"
)

//todo: move to Job?

type SketchedPage struct {
	Canvas      *canvas.Canvas
	DrawContext *canvas.Context
}

func NewSketch() *SketchedPage {
	// todo: paper size in mm, e.g. A4; create Sketch from Job
	c := canvas.New(10.0, 10.0)
	return &SketchedPage{
		Canvas:      c,
		DrawContext: canvas.NewContext(c),
	}
}
