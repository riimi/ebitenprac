package main

import (
	"ebitenprac/ngword"
	"ebitenprac/turi"
	"fmt"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"image"
	"log"
	"strings"
)

type BatchScene struct {
	input     *turi.TextBox
	output    *turi.TextBox
	execBtn   *turi.Button
	uniqueBtn *turi.Button
	filter    *ngword.LocalAlignment
}

func NewBatchScene() *BatchScene {
	la := ngword.NewLocalAlignment(ngword.ReadDataframeFromCSV("resource/ngwords.new.plain.csv"))
	tb1 := &turi.TextBox{
		Rect: image.Rect(16, 16, screenWidth/2-16, screenHeight-128),
	}
	tb2 := &turi.TextBox{
		Rect:          image.Rect(screenWidth/2+16, 16, screenWidth-16, screenHeight-128),
		ReadOnly:      true,
		HideScrollBar: true,
	}
	tb1.Mirror = tb2
	btn := &turi.Button{
		Rect: image.Rect(16, screenHeight-112, 96, screenHeight-88),
		Text: "Execute",
	}
	btn.SetOnPressed(func(b *turi.Button) {
		stcs := strings.Split(tb1.Text, "\n")
		df := dataframe.New(series.Strings(stcs))
		rep, err := la.Do(df)
		if err != nil {
			log.Print(err)
		}
		rec := rep.Col("filtered").Records()
		tb2.Text = strings.Join(rec, "\n")
	})
	uni := &turi.Button{
		Rect: image.Rect(128, screenHeight-112, 208, screenHeight-88),
		Text: "Unique",
	}
	uni.SetOnPressed(func(b *turi.Button) {
		stcs := strings.Split(tb1.Text, "\n")
		stcs = Unique(stcs)
		tb1.Text = strings.Join(stcs, "\n")
	})

	return &BatchScene{
		input:     tb1,
		output:    tb2,
		execBtn:   btn,
		uniqueBtn: uni,
		filter:    la,
	}
}

func (s *BatchScene) Update(g *turi.GameState) error {
	s.input.Update(g.Input)
	//s.output.Update(g.Input)
	s.execBtn.Update(g.Input)
	s.uniqueBtn.Update(g.Input)
	return nil
}

func (s *BatchScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%3.1f", ebiten.CurrentTPS()))
	s.input.Draw(screen)
	s.output.Draw(screen)
	s.execBtn.Draw(screen)
	s.uniqueBtn.Draw(screen)
}

func Unique(slice []string) []string {
	ret := make([]string, 0, len(slice))
	m := make(map[string]struct{})

	for _, val := range slice {
		if _, ok := m[val]; !ok {
			m[val] = struct{}{}
			ret = append(ret, val)
		}
	}
	return ret
}
