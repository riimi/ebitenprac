package turi

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/hajimehoshi/ebiten/vector"
	"image/color"
)

const (
	plotLeftPadding = 32
	plotHeight      = 240
	plotWidth       = 480
)

func Plot(screen *ebiten.Image, minx, miny float32, x []int, y []float32) {
	if len(y) <= 0 {
		return
	}

	if x == nil {
		x = make([]int, len(y))
		for i := range x {
			x[i] = i
		}
	}
	min, max := yNormalize(y)

	var op vector.DrawPathOptions
	var base vector.Path
	base.MoveTo(plotLeftPadding+minx, miny)
	base.LineTo(plotLeftPadding+minx, miny+plotHeight)
	base.LineTo(plotLeftPadding+plotWidth, miny+plotHeight)
	op.LineWidth = 4
	op.StrokeColor = color.White
	base.Draw(screen, &op)

	// y axis
	text.Draw(screen, fmt.Sprintf("%3.1f", max), uiFont, int(minx), int(miny+12), color.White)
	text.Draw(screen, fmt.Sprintf("%3.1f", min), uiFont, int(minx), int(miny+plotHeight), color.White)

	// x axis
	for i := 0; i < len(x); i += 10 {
		text.Draw(
			screen,
			fmt.Sprintf("%d", x[i]),
			uiFont,
			int(minx+plotLeftPadding+float32(i*plotWidth)/float32(len(x))),
			int(miny+plotHeight+16),
			color.White,
		)
	}

	var data vector.Path
	data.MoveTo(plotLeftPadding+minx, miny+plotHeight-y[0])
	for i := range y {
		data.LineTo(plotLeftPadding+minx+float32(i*plotWidth)/float32(len(y)), miny+plotHeight-y[i])
	}
	op.LineWidth = 2
	op.StrokeColor = color.RGBA{0x44, 0x9a, 0xae, 0xff}
	data.Draw(screen, &op)

}

func yNormalize(y []float32) (float32, float32) {
	var min float32 = 9876554321.0
	var max float32 = -987654321.0
	for i := range y {
		if y[i] < min {
			min = y[i]
		}
		if y[i] > max {
			max = y[i]
		}
	}
	padding := (max - min) * 0.1
	min -= padding
	max += padding

	for i := range y {
		y[i] = plotHeight * (y[i] - min) / (max - min)
	}

	return min, max
}
