package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/vector"
	"image"
	"image/color"
	"log"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type UI struct {
	button1  *Button
	TextBox1 *TextBox
}

func NewUI() *UI {
	b1 := &Button{
		Rect: image.Rect(16, 16, 144, 48),
		Text: "Button 1",
	}
	b1.SetOnPressed(func(b *Button) {
		log.Printf("%s clicked", b.Text)
	})
	t1 := &TextBox{
		Rect: image.Rect(160, 16, 440, 48),
		Text: "Sample",
	}
	t1.SetOnEnterPressed(func(t *TextBox) {
		log.Printf("%s entered", t.Text)
		t.Text = ""
		b1.Press()
	})
	return &UI{button1: b1, TextBox1: t1}
}

func (ui *UI) Update(input Input) error {
	ui.button1.Update(input)
	ui.TextBox1.Update(input)

	return nil
}

func (ui *UI) Draw(screen *ebiten.Image) {
	ui.button1.Draw(screen)
	ui.TextBox1.Draw(screen)
	DrawLine(screen, []int{2, 6, 10, 10, 9, 8, 7, 6, 6, 5, 4, 3, 2, 1, 0})
}

func DrawLine(screen *ebiten.Image, ints []int) {
	var path vector.Path

	path.MoveTo(16, 100+float32(ints[0]))

	for i, v := range ints {
		path.LineTo(16+float32(i*20), 150-float32(v*5))
	}
	op := &vector.DrawPathOptions{}
	op.LineWidth = 4
	op.StrokeColor = color.RGBA{0xdb, 0x56, 0x20, 0xff}
	path.Draw(screen, op)
}

type Game struct {
	ui    *UI
	input Input
}

func (g *Game) update(screen *ebiten.Image) error {
	g.ui.Update(g.input)

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	screen.Fill(color.RGBA{0xeb, 0xeb, 0xeb, 0xff})
	g.ui.Draw(screen)

	return nil
}

func main() {
	g := &Game{
		ui: NewUI(),
	}
	if err := ebiten.Run(g.update, screenWidth, screenHeight, 1, "verbose"); err != nil {
		log.Fatal(err)
	}
}
