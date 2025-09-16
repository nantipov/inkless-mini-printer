package compiler

import "nantipov/printerservice/internal/printingdata"

func sketchFromRaster(document printingdata.IppDocument) ([]*printingdata.SketchedPage, error) {
	sketch := printingdata.NewSketch()
	return []*printingdata.SketchedPage{sketch}, nil
}
