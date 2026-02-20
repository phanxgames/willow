// Text demonstrates bitmap font (BMFont) text rendering with willow.NewText.
// Shows alignment modes, word wrapping, colors, and multi-line alignment.
// No external asset files — a bitmap font atlas is generated at startup from
// the Go Regular TTF font using Ebitengine's text/v2 renderer.
package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/phanxgames/willow"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	windowTitle = "Willow — Bitmap Text Example"
	showFPS     = true
	screenW     = 800
	screenH     = 600
)

type demo struct {
	scene    *willow.Scene
	counter  *willow.Node
	frameNum int
}

func main() {
	scene := willow.NewScene()
	scene.ClearColor = willow.Color{R: 0.08, G: 0.08, B: 0.1, A: 1}

	// Generate bitmap font atlases from TTF at startup.
	bmFont, atlasImg := generateBitmapFont(18, 0)
	scene.RegisterPage(0, atlasImg)

	bmFontSmall, atlasSmall := generateBitmapFont(14, 1)
	scene.RegisterPage(1, atlasSmall)

	root := scene.Root()

	// ── Title ─────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "Willow - Bitmap Text", bmFont)
	title.TextBlock.Align = willow.TextAlignRight
	title.TextBlock.WrapWidth = screenW
	title.TextBlock.Color = willow.Color{R: 1, G: 1, B: 1, A: 1}
	title.X = 0
	title.Y = 18
	root.AddChild(title)

	// ── Colors ────────────────────────────────────────────────────────────────
	addSectionLabel(root, bmFontSmall, "Colors", 24, 58)

	colors := []struct {
		label string
		color willow.Color
	}{
		{"White", willow.Color{R: 1, G: 1, B: 1, A: 1}},
		{"Cyan", willow.Color{R: 0.4, G: 0.9, B: 1, A: 1}},
		{"Orange", willow.Color{R: 1, G: 0.6, B: 0.2, A: 1}},
		{"Pink", willow.Color{R: 1, G: 0.4, B: 0.7, A: 1}},
	}
	x := 24.0
	for _, c := range colors {
		n := willow.NewText("color-"+c.label, c.label, bmFont)
		n.TextBlock.Color = c.color
		n.X = x
		n.Y = 78
		root.AddChild(n)
		w, _ := bmFont.MeasureString(c.label)
		x += w + 24
	}

	// ── Single-line Alignment ─────────────────────────────────────────────────
	addSectionLabel(root, bmFontSmall, "Single-line Alignment", 24, 118)

	alignLeft := willow.NewText("align-left", "Left aligned", bmFont)
	alignLeft.TextBlock.Color = willow.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	alignLeft.TextBlock.Align = willow.TextAlignLeft
	alignLeft.TextBlock.WrapWidth = screenW - 48
	alignLeft.X = 24
	alignLeft.Y = 138
	root.AddChild(alignLeft)

	alignCenter := willow.NewText("align-center", "Center aligned", bmFont)
	alignCenter.TextBlock.Color = willow.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	alignCenter.TextBlock.Align = willow.TextAlignCenter
	alignCenter.TextBlock.WrapWidth = screenW - 48
	alignCenter.X = 24
	alignCenter.Y = 162
	root.AddChild(alignCenter)

	alignRight := willow.NewText("align-right", "Right aligned", bmFont)
	alignRight.TextBlock.Color = willow.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	alignRight.TextBlock.Align = willow.TextAlignRight
	alignRight.TextBlock.WrapWidth = screenW - 48
	alignRight.X = 24
	alignRight.Y = 186
	root.AddChild(alignRight)

	// ── Multi-line Alignment ──────────────────────────────────────────────────
	const wrapW = 300.0
	const colY = 234.0
	addSectionLabel(root, bmFontSmall, "Multi-line Alignment  (wrap 300px)", 24, 218)

	guideColor := willow.Color{R: 0.25, G: 0.3, B: 0.35, A: 1}
	colOffsets := []float64{24, 280, 536}
	for _, cx := range colOffsets {
		addGuideLine(root, cx, colY, 80, guideColor)
		addGuideLine(root, cx+wrapW, colY, 80, guideColor)
	}

	multiLine := "Short line.\nA somewhat longer second line here.\nTiny."

	wrapLeft := willow.NewText("wrap-left", multiLine, bmFontSmall)
	wrapLeft.TextBlock.Color = willow.Color{R: 0.7, G: 0.9, B: 0.7, A: 1}
	wrapLeft.TextBlock.Align = willow.TextAlignLeft
	wrapLeft.TextBlock.WrapWidth = wrapW
	wrapLeft.X = colOffsets[0]
	wrapLeft.Y = colY
	root.AddChild(wrapLeft)

	wrapCenter := willow.NewText("wrap-center", multiLine, bmFontSmall)
	wrapCenter.TextBlock.Color = willow.Color{R: 0.7, G: 0.8, B: 1, A: 1}
	wrapCenter.TextBlock.Align = willow.TextAlignCenter
	wrapCenter.TextBlock.WrapWidth = wrapW
	wrapCenter.X = colOffsets[1]
	wrapCenter.Y = colY
	root.AddChild(wrapCenter)

	wrapRight := willow.NewText("wrap-right", multiLine, bmFontSmall)
	wrapRight.TextBlock.Color = willow.Color{R: 1, G: 0.8, B: 0.7, A: 1}
	wrapRight.TextBlock.Align = willow.TextAlignRight
	wrapRight.TextBlock.WrapWidth = wrapW
	wrapRight.X = colOffsets[2]
	wrapRight.Y = colY
	root.AddChild(wrapRight)

	// ── Word-wrapped Paragraph Alignment ──────────────────────────────────────
	const paraWrap = 220.0
	paraY := 340.0
	addSectionLabel(root, bmFontSmall, "Word-wrapped Paragraph  (wrap 220px)", 24, 324)

	paraText := "The quick brown fox jumps over the lazy dog. Alignment should affect each wrapped line independently."
	paraColOffsets := []float64{24, 280, 536}
	for _, cx := range paraColOffsets {
		addGuideLine(root, cx, paraY, 100, guideColor)
		addGuideLine(root, cx+paraWrap, paraY, 100, guideColor)
	}

	paraLeft := willow.NewText("para-left", paraText, bmFontSmall)
	paraLeft.TextBlock.Color = willow.Color{R: 0.7, G: 0.9, B: 0.7, A: 1}
	paraLeft.TextBlock.Align = willow.TextAlignLeft
	paraLeft.TextBlock.WrapWidth = paraWrap
	paraLeft.X = paraColOffsets[0]
	paraLeft.Y = paraY
	root.AddChild(paraLeft)

	paraCenter := willow.NewText("para-center", paraText, bmFontSmall)
	paraCenter.TextBlock.Color = willow.Color{R: 0.7, G: 0.8, B: 1, A: 1}
	paraCenter.TextBlock.Align = willow.TextAlignCenter
	paraCenter.TextBlock.WrapWidth = paraWrap
	paraCenter.X = paraColOffsets[1]
	paraCenter.Y = paraY
	root.AddChild(paraCenter)

	paraRight := willow.NewText("para-right", paraText, bmFontSmall)
	paraRight.TextBlock.Color = willow.Color{R: 1, G: 0.8, B: 0.7, A: 1}
	paraRight.TextBlock.Align = willow.TextAlignRight
	paraRight.TextBlock.WrapWidth = paraWrap
	paraRight.X = paraColOffsets[2]
	paraRight.Y = paraY
	root.AddChild(paraRight)

	// ── Live update ───────────────────────────────────────────────────────────
	addSectionLabel(root, bmFontSmall, "Live update", 24, 470)

	counter := willow.NewText("counter", "Frame: 0", bmFont)
	counter.TextBlock.Color = willow.Color{R: 0.5, G: 1, B: 0.5, A: 1}
	counter.X = 24
	counter.Y = 490
	root.AddChild(counter)

	d := &demo{scene: scene, counter: counter}
	scene.SetUpdateFunc(d.update)

	if err := willow.Run(scene, willow.RunConfig{
		Title:   windowTitle,
		Width:   screenW,
		Height:  screenH,
		ShowFPS: showFPS,
	}); err != nil {
		log.Fatal(err)
	}
}

func (d *demo) update() error {
	d.frameNum++
	d.counter.TextBlock.Content = fmt.Sprintf("Frame: %d", d.frameNum)
	d.counter.TextBlock.MarkDirty()
	return nil
}

// addSectionLabel places a small muted label at (x, y).
func addSectionLabel(root *willow.Node, font willow.Font, s string, x, y float64) {
	n := willow.NewText("section-"+s, s, font)
	n.TextBlock.Color = willow.Color{R: 0.4, G: 0.5, B: 0.6, A: 1}
	n.X = x
	n.Y = y
	root.AddChild(n)
}

// addGuideLine draws a thin vertical line at (x, y) with the given height.
func addGuideLine(root *willow.Node, x, y, height float64, color willow.Color) {
	line := willow.NewSprite("guide", willow.TextureRegion{})
	line.Color = color
	line.X = x
	line.Y = y
	line.ScaleX = 1
	line.ScaleY = height
	root.AddChild(line)
}

// --- Programmatic bitmap font generation ---

// generateBitmapFont renders printable ASCII glyphs (32-126) into an atlas image
// using Ebitengine's TTF renderer, then creates a BitmapFont from the layout.
// Returns the font and the atlas image (caller must RegisterPage).
func generateBitmapFont(size float64, pageIndex int) (*willow.BitmapFont, *ebiten.Image) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatalf("font source: %v", err)
	}
	face := &text.GoTextFace{Source: src, Size: size}

	lineHeight := face.Metrics().HAscent + face.Metrics().HDescent + face.Metrics().HLineGap
	base := face.Metrics().HAscent

	// Measure all printable ASCII glyphs and compute atlas layout.
	type glyphInfo struct {
		r                rune
		w, h             int
		xOffset, yOffset int
		xAdvance         int
	}
	var glyphs []glyphInfo
	for r := rune(32); r <= 126; r++ {
		s := string(r)
		w, _ := text.Measure(s, face, lineHeight)
		gw := int(math.Ceil(w))
		gh := int(math.Ceil(lineHeight))
		adv := gw
		if r == ' ' {
			gw = 0
			gh = 0
		}
		glyphs = append(glyphs, glyphInfo{
			r: r, w: gw, h: gh,
			xOffset: 0, yOffset: 0,
			xAdvance: adv,
		})
	}

	// Pack glyphs into rows (simple shelf packing).
	const atlasW = 512
	curX, curY, rowH := 0, 0, 0
	type packedGlyph struct {
		glyphInfo
		atlasX, atlasY int
	}
	var packed []packedGlyph
	for _, g := range glyphs {
		if g.w == 0 {
			packed = append(packed, packedGlyph{glyphInfo: g})
			continue
		}
		if curX+g.w > atlasW {
			curY += rowH + 1
			curX = 0
			rowH = 0
		}
		packed = append(packed, packedGlyph{glyphInfo: g, atlasX: curX, atlasY: curY})
		curX += g.w + 1
		if g.h > rowH {
			rowH = g.h
		}
	}
	atlasH := curY + rowH + 1

	// Render glyphs into atlas image.
	atlas := ebiten.NewImage(atlasW, atlasH)
	for i := range packed {
		pg := &packed[i]
		if pg.w == 0 {
			continue
		}
		// Draw glyph into a temp image, then copy to atlas.
		tmp := ebiten.NewImage(pg.w, pg.h)
		op := &text.DrawOptions{}
		op.ColorScale.Scale(1, 1, 1, 1)
		op.LineSpacing = lineHeight
		text.Draw(tmp, string(pg.r), face, op)

		var dop ebiten.DrawImageOptions
		dop.GeoM.Translate(float64(pg.atlasX), float64(pg.atlasY))
		atlas.DrawImage(tmp, &dop)
		tmp.Deallocate()
	}

	// Build BMFont .fnt text data.
	fnt := fmt.Sprintf(
		"info face=\"Generated\" size=%d bold=0 italic=0\n"+
			"common lineHeight=%d base=%d scaleW=%d scaleH=%d pages=1 packed=0\n"+
			"page id=0 file=\"generated.png\"\n"+
			"chars count=%d\n",
		int(size), int(lineHeight), int(base), atlasW, atlasH, len(packed),
	)
	for _, pg := range packed {
		fnt += fmt.Sprintf("char id=%d x=%d y=%d width=%d height=%d xoffset=%d yoffset=%d xadvance=%d page=%d\n",
			pg.r, pg.atlasX, pg.atlasY, pg.w, pg.h, pg.xOffset, pg.yOffset, pg.xAdvance, pageIndex)
	}

	// Determine page index from atlas image pointer identity.
	bmFont, err := willow.LoadBitmapFontPage([]byte(fnt), uint16(pageIndex))
	if err != nil {
		log.Fatalf("LoadBitmapFont: %v", err)
	}

	return bmFont, atlas
}

// init registers an offscreen image so Ebitengine's internal state is
// initialized before we create atlas images in main().
func init() {
	// Force Ebitengine to initialize its graphics context by creating a tiny image.
	_ = ebiten.NewImage(1, 1)
	_ = image.Rect(0, 0, 1, 1)
}
