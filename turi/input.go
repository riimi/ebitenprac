package turi

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"image"
)

const (
	InputRectValidClicked = iota
	InputRectInvalidClicked
	InputRectNotClicked
)

type Input struct {
}

func (input *Input) IsRectClicked(rect image.Rectangle, button ebiten.MouseButton) int {
	if ebiten.IsMouseButtonPressed(button) {
		x, y := ebiten.CursorPosition()
		if rect.Min.X <= x && x < rect.Max.X && rect.Min.Y <= y && y < rect.Max.Y {
			return InputRectValidClicked
		} else {
			return InputRectInvalidClicked
		}
	}
	return InputRectNotClicked
}

func (input *Input) RepeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func (input *Input) PairKeyPressed(key1, key2 ebiten.Key) bool {
	return ebiten.IsKeyPressed(key1) && ebiten.IsKeyPressed(key2)
}
