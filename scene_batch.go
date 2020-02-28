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
	input      *turi.TextBox
	output     *turi.TextBox
	execBtn    *turi.Button
	uniqueBtn  *turi.Button
	summaryBtn *turi.Button
	filter     *ngword.LocalAlignmentTrie
}

func NewBatchScene() *BatchScene {
	//la := ngword.NewLocalAlignment(ngword.ReadDataframeFromCSV("resource/ngwords.new.plain.csv"))
	la := ngword.NewLocalAlignmentTrie(ngword.ReadDataframeFromCSV("resource/ngwords.origin.csv"))
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
		stcs := strings.Split(tb1.Text(false), "\n")
		df := dataframe.New(series.Strings(stcs))
		rep, err := la.Do(df)
		if err != nil {
			log.Print(err)
		}
		rec := rep.Col("filtered").Records()
		tb2.SetText(strings.Join(rec, "\n"))
	})
	uni := &turi.Button{
		Rect: image.Rect(128, screenHeight-112, 208, screenHeight-88),
		Text: "Unique",
	}
	uni.SetOnPressed(func(b *turi.Button) {
		stcs := strings.Split(tb1.Text(false), "\n")
		stcs = Unique(stcs)
		tb1.SetText(strings.Join(stcs, "\n"))
	})
	summary := &turi.Button{
		Rect: image.Rect(224, screenHeight-112, 304, screenHeight-88),
		Text: "summary",
	}
	summary.SetOnPressed(func(b *turi.Button) {
		t2 := strings.Split(tb2.Text(false), "\n")
		t1 := strings.Split(tb1.Text(false), "\n")
		a2 := make([]string, 0, len(t2))
		a1 := make([]string, 0, len(t1))
		for i := range t2 {
			if strings.Index(t2[i], "***") < 0 {
				continue
			}
			a1 = append(a1, t1[i])
			a2 = append(a2, t2[i])
		}
		tb1.SetText(strings.Join(a1, "\n"))
		tb2.SetText(strings.Join(a2, "\n"))
	})

	return &BatchScene{
		input:      tb1,
		output:     tb2,
		execBtn:    btn,
		uniqueBtn:  uni,
		summaryBtn: summary,
		filter:     la,
	}
}

func (s *BatchScene) Update(g *turi.GameState) error {
	s.input.Update(g.Input)
	//s.output.Update(g.Input)
	s.execBtn.Update(g.Input)
	s.uniqueBtn.Update(g.Input)
	s.summaryBtn.Update(g.Input)
	return nil
}

func (s *BatchScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%3.1f", ebiten.CurrentTPS()))
	s.input.Draw(screen)
	s.output.Draw(screen)
	s.execBtn.Draw(screen)
	s.uniqueBtn.Draw(screen)
	s.summaryBtn.Draw(screen)
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
