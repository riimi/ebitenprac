package main

import (
	"bytes"
	"ebitenprac/ngword"
	"ebitenprac/turi"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
	"golang.org/x/text/unicode/norm"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"image"
	"image/color"
	"io/ioutil"
	"log"
)

type UI struct {
	button1   *turi.Button
	TextLine1 *turi.TextLine
	TextLine2 *turi.TextLine
	TextLine3 *turi.TextLine
	debug     string
	font      font.Face
	sResult   ngword.SmithWatermanEnd
	barGraph  *ebiten.Image
	filter    *ngword.LocalAlignmentDebug
}

func NewUI() *UI {
	ui := &UI{}

	b, err := ioutil.ReadFile("resource/malgun.ttf")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := truetype.Parse(b)
	if err != nil {
		log.Fatal(err)
	}
	ui.font = truetype.NewFace(tt, &truetype.Options{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	vg.AddFont("resource/malgun.ttf", tt)

	ui.button1 = &turi.Button{
		Rect: image.Rect(416, 16, 500, 48),
		Text: "Button 1",
	}
	ui.TextLine1 = &turi.TextLine{
		Rect: image.Rect(16, 16, 400, 48),
		Text: "Hello@@@world",
	}
	ui.TextLine2 = &turi.TextLine{
		Rect: image.Rect(16, 64, 400, 96),
		Text: "@@@",
	}
	ui.TextLine3 = &turi.TextLine{
		Rect:     image.Rect(16, 112, 400, 144),
		Text:     "Replaced",
		ReadOnly: true,
	}
	ui.TextLine1.SetOnEnterPressed(func(t *turi.TextLine) {
		ui.button1.Press()
	})
	ui.button1.SetOnPressed(func(b *turi.Button) {
		ui.debug = ""

		replaced, _ := ui.filter.Replace(ui.TextLine1.Text)
		//vs := make([]float64, len(ui.filter.End))
		//ticks := make([]string, len(ui.filter.End))
		vs := make([]float64, 10)
		ticks := make([]string, 10)
		for i, e := range ui.filter.End[:10] {
			vs[i] = float64(e.MaxAgreement) / float64(e.CompleteAgreement)
			ticks[i] = e.MatchWord
		}
		var err error
		ui.barGraph, err = NewBarGraph(vs, ticks)
		if err != nil {
			log.Fatal(err)
		}

		ui.TextLine3.Text = norm.NFC.String(string(replaced))
	})

	ui.barGraph, err = NewBarGraph([]float64{0.0}, []string{"One"})
	if err != nil {
		log.Fatal(err)
	}

	df := ngword.ReadDataframeFromCSV("resource/ngwords.new.plain.csv")
	ui.filter = ngword.NewLocalAlignmentDebug(df)

	return ui
}

func NewBarGraph(vs []float64, ticks []string) (*ebiten.Image, error) {
	groupA := plotter.Values(vs)

	plot.DefaultFont = "resource/malgun.ttf"
	p, err := plot.New()
	if err != nil {
		return nil, err
	}
	p.Title.Text = "Bar chart"
	p.Title.Font, err = vg.MakeFont("resource/malgun.ttf", 12)
	if err != nil {
		return nil, err
	}
	p.Y.Label.Text = "Similarity"

	w := vg.Points(20)

	barsA, err := plotter.NewBarChart(groupA, w)
	if err != nil {
		return nil, err
	}
	barsA.LineStyle.Width = vg.Length(0)
	barsA.Color = plotutil.Color(0)

	p.Add(barsA)
	//p.Legend.Add("Group A", barsA)
	//p.Legend.Top = true
	p.NominalX(ticks...)
	p.Y.Max = 1.0

	writer, err := p.WriterTo(8*vg.Inch, 3*vg.Inch, "png")
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	_, err = writer.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(buf)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img, ebiten.FilterDefault)
}

func (ui *UI) Update(s *turi.GameState) error {
	ui.button1.Update(s.Input)
	ui.TextLine1.Update(s.Input)
	ui.TextLine2.Update(s.Input)
	ui.TextLine3.Update(s.Input)
	return nil
}

func (ui *UI) Draw(screen *ebiten.Image) {
	ui.button1.Draw(screen)
	ui.TextLine1.Draw(screen)
	ui.TextLine2.Draw(screen)
	ui.TextLine3.Draw(screen)

	text.Draw(screen, ui.debug, ui.font, 16, 160, color.Black)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(16, 240)
	screen.DrawImage(ui.barGraph, op)
}
