package turi

import (
	"github.com/atotto/clipboard"
	"github.com/hajimehoshi/ebiten"
	"log"
	"unicode/utf8"
)

type TypeWriter struct {
	text        string
	ctrlVDown   bool
	cursor      int
	IgnoreEnter bool
}

func (w *TypeWriter) Update(input *Input) {
	if input.PairKeyPressed(ebiten.KeyV, ebiten.KeyControl) {
		if w.ctrlVDown == false {
			clip, err := clipboard.ReadAll()
			if err != nil {
				log.Print(err)
			} else {
				w.text += clip
			}
		}
		w.ctrlVDown = true
	} else {
		w.ctrlVDown = false
	}

	if input.PairKeyPressed(ebiten.KeyA, ebiten.KeyControl) {
		w.SetText("")
	}

	if input.RepeatingKeyPressed(ebiten.KeyLeft) {
		_, size := utf8.DecodeLastRuneInString(w.text[:w.cursor])
		w.cursor -= size
		if w.cursor < 0 {
			w.cursor = 0
		}
	}
	if input.RepeatingKeyPressed(ebiten.KeyRight) {
		_, size := utf8.DecodeRuneInString(w.text[w.cursor:])
		w.cursor += size
		if w.cursor >= len(w.text) {
			w.cursor = len(w.text)
		}
	}

	if w.cursor >= 0 && w.cursor <= len(w.text) {
		ch := string(ebiten.InputChars())
		if w.cursor == len(w.text) {
			w.text += ch
		} else {
			w.text = w.text[:w.cursor] + ch + w.text[w.cursor:]
		}
		w.cursor += len(ch)
	}

	if input.RepeatingKeyPressed(ebiten.KeyBackspace) {
		if len(w.text) >= 1 {
			_, size := utf8.DecodeLastRuneInString(w.text[:w.cursor])
			w.text = w.text[:w.cursor-size] + w.text[w.cursor:]
			w.cursor -= size
		}
	}

	if !w.IgnoreEnter && input.RepeatingKeyPressed(ebiten.KeyEnter) {
		if w.cursor == len(w.text) {
			w.text += "\n"
		} else {
			w.text = w.text[:w.cursor] + "\n" + w.text[w.cursor:]
		}
		w.cursor += len("\n")
	}

}

func (w *TypeWriter) Text(cursor bool) string {
	if cursor {
		return w.text[:w.cursor] + "|" + w.text[w.cursor:]
	}
	return w.text
}

func (w *TypeWriter) SetText(txt string) {
	w.text = txt
	w.cursor = len(w.text)
}
