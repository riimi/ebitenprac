package main

import (
	"ebitenprac/turi"
	"github.com/hajimehoshi/ebiten"
	"image"
	"image/color"
)

type Navi struct {
	btns         []*turi.Button
	scenes       map[*turi.Button]turi.Scene
	curScene     turi.Scene
	rect         image.Rectangle
	background   *ebiten.Image
	sceneManager *turi.SceneManager
}

func NewNavi(rect image.Rectangle) *Navi {
	navi := &Navi{
		btns:       make([]*turi.Button, 0),
		scenes:     make(map[*turi.Button]turi.Scene),
		rect:       rect,
		background: nil,
	}

	padding := 10
	width := 100
	sx := rect.Min.X
	next := func(x int) int {
		return x + width + padding
	}

	b1 := &turi.Button{
		Text: "One sentence",
		Rect: image.Rect(sx, rect.Min.Y, sx+width, rect.Max.Y),
	}
	navi.scenes[b1] = NewUI()
	b1.SetOnPressed(func(b *turi.Button) {
		navi.sceneManager.GoTo(navi.scenes[b])
	})
	sx = next(sx)

	b2 := &turi.Button{
		Text: "Batch",
		Rect: image.Rect(sx, rect.Min.Y, sx+width, rect.Max.Y),
	}
	navi.scenes[b2] = NewBatchScene()
	b2.SetOnPressed(func(b *turi.Button) {
		navi.sceneManager.GoTo(navi.scenes[b])
	})
	sx = next(sx)

	navi.btns = append(navi.btns, b1, b2)

	return navi
}

func (navi *Navi) Update(input *turi.Input) {
	if navi.sceneManager == nil {
		navi.curScene = NewUI()
		navi.sceneManager = turi.NewSceneManager(screenWidth, screenHeight)
		navi.sceneManager.GoTo(navi.curScene)
	}
	if navi.background == nil {
		navi.background, _ = ebiten.NewImage(navi.rect.Dx(), navi.rect.Dy(), ebiten.FilterDefault)
		navi.background.Fill(color.RGBA{0xdd, 0xdd, 0xdd, 0xff})
	}

	navi.sceneManager.Update(input)

	for _, b := range navi.btns {
		b.Update(input)
	}
}

func (navi *Navi) Draw(screen *ebiten.Image) {
	navi.sceneManager.Draw(screen)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(navi.rect.Min.X), float64(navi.rect.Min.Y))
	screen.DrawImage(navi.background, op)
	for _, b := range navi.btns {
		b.Draw(screen)
	}
}
