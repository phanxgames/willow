// Text demonstrates TTF text rendering with willow.NewText and willow.LoadTTFFont.
// Shows different font sizes, colors, alignment modes, word wrapping, text
// outlines, multi-line alignment, and live content updates. No external asset
// files required — the Go Regular font is sourced from golang.org/x/image.
package main

import (
	"fmt"
	"log"

	"github.com/phanxgames/willow"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	windowTitle = "Willow — Text Example"
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

	fontLarge, err := willow.LoadTTFFont(goregular.TTF, 34)
	if err != nil {
		log.Fatalf("load font large: %v", err)
	}
	fontMedium, err := willow.LoadTTFFont(goregular.TTF, 18)
	if err != nil {
		log.Fatalf("load font medium: %v", err)
	}
	fontSmall, err := willow.LoadTTFFont(goregular.TTF, 14)
	if err != nil {
		log.Fatalf("load font small: %v", err)
	}

	root := scene.Root()

	// ── Title ─────────────────────────────────────────────────────────────────
	title := willow.NewText("title", "Willow — Text Rendering", fontLarge)
	title.TextBlock.Align = willow.TextAlignRight
	title.TextBlock.WrapWidth = screenW
	title.TextBlock.Color = willow.Color{R: 1, G: 1, B: 1, A: 1}
	title.X = 0
	title.Y = 18
	root.AddChild(title)

	// ── Divider label ─────────────────────────────────────────────────────────
	addSectionLabel(root, fontSmall, "Colors", 24, 88)

	// ── Colored text ──────────────────────────────────────────────────────────
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
		n := willow.NewText("color-"+c.label, c.label, fontMedium)
		n.TextBlock.Color = c.color
		n.X = x
		n.Y = 108
		root.AddChild(n)
		w, _ := fontMedium.MeasureString(c.label)
		x += w + 24
	}

	// ── Divider label ─────────────────────────────────────────────────────────
	addSectionLabel(root, fontSmall, "Single-line Alignment", 24, 148)

	// ── Single-line aligned text ──────────────────────────────────────────────
	alignLeft := willow.NewText("align-left", "Left aligned", fontMedium)
	alignLeft.TextBlock.Color = willow.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	alignLeft.TextBlock.Align = willow.TextAlignLeft
	alignLeft.TextBlock.WrapWidth = screenW - 48
	alignLeft.X = 24
	alignLeft.Y = 168
	root.AddChild(alignLeft)

	alignCenter := willow.NewText("align-center", "Center aligned", fontMedium)
	alignCenter.TextBlock.Color = willow.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	alignCenter.TextBlock.Align = willow.TextAlignCenter
	alignCenter.TextBlock.WrapWidth = screenW - 48
	alignCenter.X = 24
	alignCenter.Y = 192
	root.AddChild(alignCenter)

	alignRight := willow.NewText("align-right", "Right aligned", fontMedium)
	alignRight.TextBlock.Color = willow.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	alignRight.TextBlock.Align = willow.TextAlignRight
	alignRight.TextBlock.WrapWidth = screenW - 48
	alignRight.X = 24
	alignRight.Y = 216
	root.AddChild(alignRight)

	// ── Multi-line Wrapped Alignment ──────────────────────────────────────────
	// This section exercises per-line alignment — each line should independently
	// align within the WrapWidth, not shift as a single block.
	const wrapW = 300.0
	const colY = 264.0

	addSectionLabel(root, fontSmall, "Multi-line Alignment  (wrap 300px)", 24, 248)

	// Guide lines: left and right edges of the wrap region for each column.
	guideColor := willow.Color{R: 0.25, G: 0.3, B: 0.35, A: 1}
	colOffsets := []float64{24, 280, 536}
	for _, cx := range colOffsets {
		addGuideLine(root, cx, colY, 100, guideColor)
		addGuideLine(root, cx+wrapW, colY, 100, guideColor)
	}

	loremShort := "Short line.\nA somewhat longer second line here.\nTiny."

	wrapLeft := willow.NewText("wrap-left", loremShort, fontSmall)
	wrapLeft.TextBlock.Color = willow.Color{R: 0.7, G: 0.9, B: 0.7, A: 1}
	wrapLeft.TextBlock.Align = willow.TextAlignLeft
	wrapLeft.TextBlock.WrapWidth = wrapW
	wrapLeft.X = colOffsets[0]
	wrapLeft.Y = colY
	root.AddChild(wrapLeft)

	wrapCenter := willow.NewText("wrap-center", loremShort, fontSmall)
	wrapCenter.TextBlock.Color = willow.Color{R: 0.7, G: 0.8, B: 1, A: 1}
	wrapCenter.TextBlock.Align = willow.TextAlignCenter
	wrapCenter.TextBlock.WrapWidth = wrapW
	wrapCenter.X = colOffsets[1]
	wrapCenter.Y = colY
	root.AddChild(wrapCenter)

	wrapRight := willow.NewText("wrap-right", loremShort, fontSmall)
	wrapRight.TextBlock.Color = willow.Color{R: 1, G: 0.8, B: 0.7, A: 1}
	wrapRight.TextBlock.Align = willow.TextAlignRight
	wrapRight.TextBlock.WrapWidth = wrapW
	wrapRight.X = colOffsets[2]
	wrapRight.Y = colY
	root.AddChild(wrapRight)

	// ── Word-wrapped paragraph alignment ─────────────────────────────────────
	addSectionLabel(root, fontSmall, "Word-wrapped Paragraph Alignment  (wrap 220px)", 24, 370)

	const paraWrap = 220.0
	paraY := 390.0
	paraText := "The quick brown fox jumps over the lazy dog. Alignment should affect each wrapped line independently."
	paraColOffsets := []float64{24, 280, 536}

	for _, cx := range paraColOffsets {
		addGuideLine(root, cx, paraY, 90, guideColor)
		addGuideLine(root, cx+paraWrap, paraY, 90, guideColor)
	}

	paraLeft := willow.NewText("para-left", paraText, fontSmall)
	paraLeft.TextBlock.Color = willow.Color{R: 0.7, G: 0.9, B: 0.7, A: 1}
	paraLeft.TextBlock.Align = willow.TextAlignLeft
	paraLeft.TextBlock.WrapWidth = paraWrap
	paraLeft.X = paraColOffsets[0]
	paraLeft.Y = paraY
	root.AddChild(paraLeft)

	paraCenter := willow.NewText("para-center", paraText, fontSmall)
	paraCenter.TextBlock.Color = willow.Color{R: 0.7, G: 0.8, B: 1, A: 1}
	paraCenter.TextBlock.Align = willow.TextAlignCenter
	paraCenter.TextBlock.WrapWidth = paraWrap
	paraCenter.X = paraColOffsets[1]
	paraCenter.Y = paraY
	root.AddChild(paraCenter)

	paraRight := willow.NewText("para-right", paraText, fontSmall)
	paraRight.TextBlock.Color = willow.Color{R: 1, G: 0.8, B: 0.7, A: 1}
	paraRight.TextBlock.Align = willow.TextAlignRight
	paraRight.TextBlock.WrapWidth = paraWrap
	paraRight.X = paraColOffsets[2]
	paraRight.Y = paraY
	root.AddChild(paraRight)

	// ── Outline + Live update ────────────────────────────────────────────────
	addSectionLabel(root, fontSmall, "Outline", 24, 500)

	outlined := willow.NewText("outlined", "Outlined text\nwith a drop shadow", fontMedium)
	outlined.TextBlock.Color = willow.Color{R: 1, G: 0.95, B: 0.7, A: 1}
	outlined.TextBlock.Outline = &willow.Outline{
		Color:     willow.Color{R: 0, G: 0, B: 0, A: 1},
		Thickness: 2,
	}
	outlined.X = 24
	outlined.Y = 520
	root.AddChild(outlined)

	addSectionLabel(root, fontSmall, "Live update", 340, 500)

	counter := willow.NewText("counter", "Frame: 0", fontMedium)
	counter.TextBlock.Color = willow.Color{R: 0.5, G: 1, B: 0.5, A: 1}
	counter.X = 340
	counter.Y = 520
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
	d.counter.TextBlock.Invalidate()
	return nil
}

// addSectionLabel places a small muted label at (x, y).
func addSectionLabel(root *willow.Node, font willow.Font, text string, x, y float64) {
	n := willow.NewText("section-"+text, text, font)
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
