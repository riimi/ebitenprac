package main

import (
	"ebitenprac/ngword"
	"ebitenprac/turi"
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"golang.org/x/text/unicode/norm"
	"image"
	"image/color"
	"log"
)

const (
	screenWidth  = 960
	screenHeight = 640
)

type UI struct {
	button1  *turi.Button
	TextBox1 *turi.TextBox
	TextBox2 *turi.TextBox
	TextBox3 *turi.TextBox
	debug    string
	sResult  ngword.SmithWatermanResult
}

func NewUI(wf *WordFilter) *UI {
	ui := &UI{}

	ui.button1 = &turi.Button{
		Rect: image.Rect(416, 16, 500, 48),
		Text: "Button 1",
	}
	ui.TextBox1 = &turi.TextBox{
		Rect: image.Rect(16, 16, 400, 48),
		Text: "Hello@@@world",
	}
	ui.TextBox2 = &turi.TextBox{
		Rect: image.Rect(16, 64, 400, 96),
		Text: "@@@",
	}
	ui.TextBox3 = &turi.TextBox{
		Rect:     image.Rect(16, 112, 400, 144),
		Text:     "Replaced",
		ReadOnly: true,
	}
	ui.TextBox1.SetOnEnterPressed(func(t *turi.TextBox) {
		ui.button1.Press()
	})
	ui.button1.SetOnPressed(func(b *turi.Button) {
		log.Printf("%s clicked", b.Text)
		ui.debug = ""

		origin := []rune(norm.NFKD.String(ui.TextBox1.Text))
		rng := []rune(norm.NFKD.String(ui.TextBox2.Text))
		ch := ngword.SmithWaterman(origin, rng, 0.8)
		for r := range ch {
			if r.LastNodes != nil {
				ui.debug += fmt.Sprintf("[completeAgreement] %v\n", r.CompleteAgreement)
				ui.debug += fmt.Sprintf("[threshold] %3.1f\n", float32(r.CompleteAgreement)*0.8)
				for _, n := range r.LastNodes {
					ui.debug += fmt.Sprintf("%d ", n.Score)
				}
				ui.debug += fmt.Sprintln()
				ui.sResult = r
			} else {
				ui.debug += fmt.Sprintln("[ng]", r.StartPos, r.EndPos, r.MatchWord)
				for i := r.StartPos; i <= r.EndPos; i++ {
					origin[i] = rune('*')
				}
			}
		}

		ui.TextBox3.Text = norm.NFC.String(string(origin))
		ui.TextBox1.Text = ""
	})

	return ui
}

func (ui *UI) Update(input turi.Input) error {
	ui.button1.Update(input)
	ui.TextBox1.Update(input)
	ui.TextBox2.Update(input)
	ui.TextBox3.Update(input)
	return nil
}

func (ui *UI) Draw(screen *ebiten.Image) {
	ui.button1.Draw(screen)
	ui.TextBox1.Draw(screen)
	ui.TextBox2.Draw(screen)
	ui.TextBox3.Draw(screen)
	ebitenutil.DebugPrintAt(screen, ui.debug, 16, 160)
	values := make([]float32, len(ui.sResult.LastNodes))
	for i, v := range ui.sResult.LastNodes {
		values[i] = float32(v.Score)
	}
	turi.Plot(screen, 16, 240, nil, values)
	//ui.DrawLine(screen)
}

//func (ui *UI) DrawLine(screen *ebiten.Image) {
//	if ui.wf.F.Result.LastNodes == nil {
//		return
//	}
//	var base vector.Path
//	base.MoveTo(16, 200)
//	base.LineTo(screenWidth-16*2, 200)
//	thresh := 200-float32(ui.wf.F.Result.CompleteAgreement)*0.9
//	base.MoveTo(16, thresh)
//	base.LineTo(screenWidth-16*2, thresh)
//	op := &vector.DrawPathOptions{
//		LineWidth: 2,
//		StrokeColor: color.Black,
//	}
//	base.Draw(screen, op)
//
//	var path vector.Path
//	path.MoveTo(16, 200-float32(ui.wf.F.Result.LastNodes[0].Score))
//
//	for i, v := range ui.wf.F.Result.LastNodes {
//		path.LineTo(16+float32(i*10), 200-float32(v.Score*5))
//	}
//	op = &vector.DrawPathOptions{}
//	op.LineWidth = 4
//	op.StrokeColor = color.RGBA{0xdb, 0x56, 0x20, 0xff}
//	path.Draw(screen, op)
//}

type Game struct {
	ui    *UI
	input turi.Input
}

func (g *Game) update(screen *ebiten.Image) error {
	g.ui.Update(g.input)

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	screen.Fill(color.RGBA{0x22, 0x22, 0x22, 0xff})
	g.ui.Draw(screen)

	return nil
}

type WordFilter struct{}

func NewWordFilter() *WordFilter {
	return &WordFilter{}
}

func main() {
	wf := NewWordFilter()
	g := &Game{
		ui: NewUI(wf),
	}
	if err := ebiten.Run(g.update, screenWidth, screenHeight, 1, "verbose"); err != nil {
		log.Fatal(err)
	}
}
