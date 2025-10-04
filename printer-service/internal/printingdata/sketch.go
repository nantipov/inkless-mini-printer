package printingdata

import (
	"fmt"
	"strings"

	"github.com/tdewolff/canvas"
)

type InstructionType int

const (
	InstGCODE InstructionType = iota
)

type InstructionArg interface {
	AsString() string
	AsFloat() float64
}

type Instruction struct {
	Type InstructionType
	Args []InstructionArg
}

type SketchedPage struct {
	Canvas      *canvas.Canvas
	DrawContext *canvas.Context
	Instructions []Instruction
}

func NewSketch() *SketchedPage {
	// todo: paper size in mm, e.g. A4; create Sketch from Job
	// c := canvas.New(canvas.A4.W, canvas.A4.H)
	c := canvas.New(1000.0, 1000.0)
	return &SketchedPage{
		Canvas:      c,
		DrawContext: canvas.NewContext(c),
		Instructions: []Instruction{},
	}
}

type ArgString string

func (x ArgString) AsString() string {
	return string(x)
}

func (x ArgString) AsFloat() float64 {
	return 0.0
}

type ArgFloat float64

func (x ArgFloat) AsString() string {
	return fmt.Sprintf("%f", float64(x)) //todo: formatting
}

func (x ArgFloat) AsFloat() float64 {
	return float64(x)
}

func (i Instruction) AsGcodeString() string {
	if i.Type != InstGCODE {
		panic("not a GCODE instruction")
	}

	s := ""
	for _, t := range i.Args {
		s = s + t.AsString() + " "
	}
	return strings.TrimSpace(s)
}

func (p *SketchedPage) AddGCodeInstruction(command string, args []float64) {
	instArgs := make([]InstructionArg, len(args) + 1)
	instArgs[0] = ArgString(command)
	for i := range args {
		instArgs[i + 1] = ArgFloat(args[i])
	}
	inst := Instruction{
		Type: InstGCODE,
		Args: instArgs,
	}
	p.Instructions = append(p.Instructions, inst)
}

