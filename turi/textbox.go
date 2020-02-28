package turi

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"image"
	"image/color"
	"strings"
)

const VScrollBarWidth = 16
const lineHeight = 16

type VScrollBar struct {
	X      int
	Y      int
	Height int

	thumbRate           float64
	thumbOffset         int
	dragging            bool
	draggingStartOffset int
	draggingStartY      int
	contentOffset       int
}

func (v *VScrollBar) thumbSize() int {
	const minThumbSize = VScrollBarWidth

	r := v.thumbRate
	if r > 1 {
		r = 1
	}
	s := int(float64(v.Height) * r)
	if s < minThumbSize {
		return minThumbSize
	}
	return s
}

func (v *VScrollBar) thumbRect() image.Rectangle {
	if v.thumbRate >= 1 {
		return image.Rectangle{}
	}

	s := v.thumbSize()
	return image.Rect(v.X, v.Y+v.thumbOffset, v.X+VScrollBarWidth, v.Y+v.thumbOffset+s)
}

func (v *VScrollBar) maxThumbOffset() int {
	return v.Height - v.thumbSize()
}

func (v *VScrollBar) ContentOffset() int {
	return v.contentOffset
}

func (v *VScrollBar) Update(input *Input, contentHeight int) {
	v.thumbRate = float64(v.Height) / float64(contentHeight)

	if !v.dragging && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		tr := v.thumbRect()
		if tr.Min.X <= x && x < tr.Max.X && tr.Min.Y <= y && y < tr.Max.Y {
			v.dragging = true
			v.draggingStartOffset = v.thumbOffset
			v.draggingStartY = y
		}
	}
	if v.dragging {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			_, y := ebiten.CursorPosition()
			v.thumbOffset = v.draggingStartOffset + (y - v.draggingStartY)
			if v.thumbOffset < 0 {
				v.thumbOffset = 0
			}
			if v.thumbOffset > v.maxThumbOffset() {
				v.thumbOffset = v.maxThumbOffset()
			}
		} else {
			v.dragging = false
		}
	}

	if input.RepeatingKeyPressed(ebiten.KeyDown) {
		v.thumbOffset += 1
		if v.thumbOffset > v.maxThumbOffset() {
			v.thumbOffset = v.maxThumbOffset()
		}
	}
	if input.RepeatingKeyPressed(ebiten.KeyUp) {
		v.thumbOffset -= 1
		if v.thumbOffset < 0 {
			v.thumbOffset = 0
		}
	}

	if input.RepeatingKeyPressed(ebiten.KeyPageDown) {
		v.thumbOffset += 20
		if v.thumbOffset > v.maxThumbOffset() {
			v.thumbOffset = v.maxThumbOffset()
		}
	}

	if input.RepeatingKeyPressed(ebiten.KeyPageUp) {
		v.thumbOffset -= 20
		if v.thumbOffset < 0 {
			v.thumbOffset = 0
		}
	}

	v.contentOffset = 0
	if v.thumbRate < 1 {
		v.contentOffset = int(float64(contentHeight) * float64(v.thumbOffset) / float64(v.Height))
	}
}

func (v *VScrollBar) Draw(dst *ebiten.Image) {
	sd := image.Rect(v.X, v.Y, v.X+VScrollBarWidth, v.Y+v.Height)
	drawNinePatches(dst, sd, imageSrcRects[imageTypeVScollBarBack])

	if v.thumbRate < 1 {
		drawNinePatches(dst, v.thumbRect(), imageSrcRects[imageTypeVScollBarFront])
	}
}

const (
	textBoxPaddingLeft = 8
)

type TextBox struct {
	TypeWriter
	Rect          image.Rectangle
	ReadOnly      bool
	Mirror        *TextBox
	HideScrollBar bool

	contentBuf *ebiten.Image
	vScrollBar *VScrollBar
	offsetX    int
	offsetY    int

	focused bool
}

func (t *TextBox) Update(input *Input) {
	if t.vScrollBar == nil {
		t.vScrollBar = &VScrollBar{}
	}
	t.vScrollBar.X = t.Rect.Max.X - VScrollBarWidth
	t.vScrollBar.Y = t.Rect.Min.Y
	t.vScrollBar.Height = t.Rect.Dy()

	_, h := t.contentSize()
	t.vScrollBar.Update(input, h)

	t.offsetX = 0
	t.offsetY = t.vScrollBar.ContentOffset()
	if t.Mirror != nil {
		t.Mirror.offsetX = 0
		t.Mirror.offsetY = t.vScrollBar.ContentOffset()
	}

	c := input.IsRectClicked(t.Rect, ebiten.MouseButtonLeft)
	if !t.ReadOnly && c == InputRectValidClicked {
		t.focused = true
	} else if c == InputRectInvalidClicked {
		t.focused = false
	}

	if t.focused {
		t.TypeWriter.Update(input)
	}
}

func (t *TextBox) contentSize() (int, int) {
	h := len(strings.Split(t.Text(false), "\n")) * lineHeight
	return t.Rect.Dx(), h
}

func (t *TextBox) viewSize() (int, int) {
	return t.Rect.Dx() - VScrollBarWidth - textBoxPaddingLeft, t.Rect.Dy()
}

func (t *TextBox) contentOffset() (int, int) {
	return t.offsetX, t.offsetY
}

func (t *TextBox) Draw(dst *ebiten.Image) {
	if t.vScrollBar == nil {
		t.vScrollBar = &VScrollBar{}
	}
	drawNinePatches(dst, t.Rect, imageSrcRects[imageTypeTextLine])

	if t.contentBuf != nil {
		vw, vh := t.viewSize()
		w, h := t.contentBuf.Size()
		if vw > w || vh > h {
			t.contentBuf.Dispose()
			t.contentBuf = nil
		}
	}
	if t.contentBuf == nil {
		w, h := t.viewSize()
		t.contentBuf, _ = ebiten.NewImage(w, h, ebiten.FilterDefault)
	}

	t.contentBuf.Clear()
	for i, line := range strings.Split(t.Text(false), "\n") {
		x := -t.offsetX + textBoxPaddingLeft
		y := -t.offsetY + i*lineHeight + lineHeight - (lineHeight-uiFontMHeight)/2
		if y < -lineHeight {
			continue
		}
		if _, h := t.viewSize(); y >= h+lineHeight {
			continue
		}
		if strings.Index(line, "***") >= 0 {
			text.Draw(t.contentBuf, line, uiFont, x, y, color.RGBA{0xff, 0x00, 0x00, 0xff})
		} else {
			text.Draw(t.contentBuf, line, uiFont, x, y, color.Black)
		}
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(t.Rect.Min.X), float64(t.Rect.Min.Y))
	dst.DrawImage(t.contentBuf, op)

	if !t.HideScrollBar {
		t.vScrollBar.Draw(dst)
	}
}
