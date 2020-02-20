package main

import (
	"ebitenprac/turi"
	"github.com/hajimehoshi/ebiten"
	"image"
	"image/color"
	"log"
)

const (
	screenWidth  = 960
	screenHeight = 640
)

type Game struct {
	input *turi.Input
	navi  *Navi
}

func (g *Game) update(screen *ebiten.Image) error {

	g.navi.Update(g.input)

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	//screen.Fill(color.RGBA{0x22, 0x22, 0x22, 0xff})
	screen.Fill(color.White)
	g.navi.Draw(screen)

	return nil
}

func main() {
	g := &Game{
		navi: NewNavi(image.Rect(16, screenHeight-48, screenWidth-16, screenHeight-16)),
	}
	if err := ebiten.Run(g.update, screenWidth, screenHeight, 1, "verbose"); err != nil {
		log.Fatal(err)
	}
}
