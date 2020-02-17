package turi

import (
	"github.com/atotto/clipboard"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
)

type imageType int

const (
	imageTypeButton imageType = iota
	imageTypeButtonPressed
	imageTypeTextBox
	imageTypeVScollBarBack
	imageTypeVScollBarFront
	imageTypeCheckBox
	imageTypeCheckBoxPressed
	imageTypeCheckBoxMark
)

var imageSrcRects = map[imageType]image.Rectangle{
	imageTypeButton:          image.Rect(0, 0, 16, 16),
	imageTypeButtonPressed:   image.Rect(16, 0, 32, 16),
	imageTypeTextBox:         image.Rect(0, 16, 16, 32),
	imageTypeVScollBarBack:   image.Rect(16, 16, 24, 32),
	imageTypeVScollBarFront:  image.Rect(24, 16, 32, 32),
	imageTypeCheckBox:        image.Rect(0, 32, 16, 48),
	imageTypeCheckBoxPressed: image.Rect(16, 32, 32, 48),
	imageTypeCheckBoxMark:    image.Rect(32, 32, 48, 48),
}

var (
	uiFont        font.Face
	uiFontMHeight int
	uiImage       *ebiten.Image
)

func init() {
	var err error
	uiImage, _, err = ebitenutil.NewImageFromFile("resource/ui.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	fp, err := os.Open("resource/malgun.ttf")
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(fp)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := truetype.Parse(b)
	if err != nil {
		log.Fatal(err)
	}
	uiFont = truetype.NewFace(tt, &truetype.Options{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	bound, _, _ := uiFont.GlyphBounds('M')
	uiFontMHeight = (bound.Max.Y - bound.Min.Y).Ceil()
}

func drawNinePatches(dst *ebiten.Image, dstRect image.Rectangle, srcRect image.Rectangle) {
	srcX := srcRect.Min.X
	srcY := srcRect.Min.Y
	srcW := srcRect.Dx()
	srcH := srcRect.Dy()

	dstX := dstRect.Min.X
	dstY := dstRect.Min.Y
	dstW := dstRect.Dx()
	dstH := dstRect.Dy()

	op := &ebiten.DrawImageOptions{}
	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			op.GeoM.Reset()

			sx := srcX
			sy := srcY
			sw := srcW / 4
			sh := srcH / 4
			dx := 0
			dy := 0
			dw := sw
			dh := sh
			switch i {
			case 1:
				sx = srcX + srcW/4
				sw = srcW / 2
				dx = srcW / 4
				dw = dstW - 2*srcW/4
			case 2:
				sx = srcX + 3*srcW/4
				dx = dstW - srcW/4
			}
			switch j {
			case 1:
				sy = srcY + srcH/4
				sh = srcH / 2
				dy = srcH / 4
				dh = dstH - 2*srcH/4
			case 2:
				sy = srcY + 3*srcH/4
				dy = dstH - srcH/4
			}

			op.GeoM.Scale(float64(dw)/float64(sw), float64(dh)/float64(sh))
			op.GeoM.Translate(float64(dx), float64(dy))
			op.GeoM.Translate(float64(dstX), float64(dstY))
			dst.DrawImage(uiImage.SubImage(image.Rect(sx, sy, sx+sw, sy+sh)).(*ebiten.Image), op)
		}
	}
}

type Button struct {
	Rect image.Rectangle
	Text string

	mouseDown bool

	onPressed func(b *Button)
}

func (b *Button) Update(input Input) {
	c := input.IsRectClicked(b.Rect, ebiten.MouseButtonLeft)
	if c == InputRectValidClicked {
		b.mouseDown = true
	} else if c == InputRectInvalidClicked {
		b.mouseDown = false
	} else {
		if b.mouseDown {
			if b.onPressed != nil {
				b.onPressed(b)
			}
		}
		b.mouseDown = false
	}
}

func (b *Button) Draw(dst *ebiten.Image) {
	t := imageTypeButton
	if b.mouseDown {
		t = imageTypeButtonPressed
	}
	drawNinePatches(dst, b.Rect, imageSrcRects[t])

	bounds, _ := font.BoundString(uiFont, b.Text)
	w := (bounds.Max.X - bounds.Min.X).Ceil()
	x := b.Rect.Min.X + (b.Rect.Dx()-w)/2
	y := b.Rect.Max.Y - (b.Rect.Dy()-uiFontMHeight)/2
	text.Draw(dst, b.Text, uiFont, x, y, color.Black)
}

func (b *Button) Press() {
	if b.onPressed != nil {
		b.onPressed(b)
	}
}

func (b *Button) SetOnPressed(f func(b *Button)) {
	b.onPressed = f
}

const (
	textBoxPaddingLeft = 8
	LineHeight         = 16
)

type TextBox struct {
	Rect     image.Rectangle
	Text     string
	ReadOnly bool

	contentBuf *ebiten.Image
	offsetX    int
	offsetY    int
	counter    int
	focused    bool
	ctrlVDown  bool

	onEnterPressed func(t *TextBox)
}

func (t *TextBox) AppendLine(line string) {
	if t.Text == "" {
		t.Text = line
	} else {
		t.Text += "\n" + line
	}
}

func (t *TextBox) Update(input Input) {
	t.offsetX = 0

	t.counter++
	c := input.IsRectClicked(t.Rect, ebiten.MouseButtonLeft)
	if !t.ReadOnly && c == InputRectValidClicked {
		t.focused = true
	} else if c == InputRectInvalidClicked {
		t.focused = false
	}

	if t.focused {
		t.Text += string(ebiten.InputChars())
		if input.RepeatingKeyPressed(ebiten.KeyBackspace) {
			if len(t.Text) >= 1 {
				t.Text = t.Text[:len(t.Text)-1]
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if t.onEnterPressed != nil && len(t.Text) >= 1 {
				t.onEnterPressed(t)
			}
		}

		if input.PairKeyPressed(ebiten.KeyV, ebiten.KeyControl) {
			if t.ctrlVDown == false {
				clip, err := clipboard.ReadAll()
				if err != nil {
					log.Print(err)
				} else {
					t.Text += clip
				}
			}
			t.ctrlVDown = true
		} else {
			t.ctrlVDown = false
		}

		if input.PairKeyPressed(ebiten.KeyA, ebiten.KeyControl) {
			t.Text = ""
		}
	}
}

func (t *TextBox) viewSize() (int, int) {
	return t.Rect.Dx() - textBoxPaddingLeft, t.Rect.Dy()
}

func (t *TextBox) Draw(dst *ebiten.Image) {
	drawNinePatches(dst, t.Rect, imageSrcRects[imageTypeTextBox])

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
	x := -t.offsetX + textBoxPaddingLeft
	y := (t.Rect.Max.Y - t.Rect.Min.Y + LineHeight - uiFontMHeight) / 2
	txt := t.Text
	if t.focused && t.counter%60 < 30 {
		txt += "_"
	}
	text.Draw(t.contentBuf, txt, uiFont, x, y, color.Black)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(t.Rect.Min.X), float64(t.Rect.Min.Y))
	dst.DrawImage(t.contentBuf, op)
}

func (t *TextBox) SetOnEnterPressed(f func(t *TextBox)) {
	t.onEnterPressed = f
}
