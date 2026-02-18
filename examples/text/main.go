// Text demonstrates TTF text rendering with willow.NewText and willow.LoadTTFFont.
// Shows different font sizes, colors, alignment modes, word wrapping, text
// outlines, and live content updates. No external asset files required —
// the Go Regular font is sourced from golang.org/x/image.
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
	screenW     = 640
	screenH     = 480
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
	addSectionLabel(root, fontSmall, "Alignment", 24, 148)

	// ── Aligned text (all share WrapWidth = screenW-48, positioned at X=24) ──
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

	// ── Divider label ─────────────────────────────────────────────────────────
	addSectionLabel(root, fontSmall, "Word Wrap  (280 px)", 24, 252)

	// ── Word-wrapped paragraph ────────────────────────────────────────────────
	para := willow.NewText("wrap", "The quick brown fox jumps over the lazy dog. Word wrapping breaks long lines automatically at the configured wrap width.", fontSmall)
	para.TextBlock.Color = willow.Color{R: 0.75, G: 0.85, B: 0.9, A: 1}
	para.TextBlock.WrapWidth = 280
	para.X = 24
	para.Y = 272
	root.AddChild(para)

	// ── Divider label ─────────────────────────────────────────────────────────
	addSectionLabel(root, fontSmall, "Outline", 340, 252)

	// ── Text with outline ─────────────────────────────────────────────────────
	outlined := willow.NewText("outlined", "Outlined text\nwith a drop shadow", fontMedium)
	outlined.TextBlock.Color = willow.Color{R: 1, G: 0.95, B: 0.7, A: 1}
	outlined.TextBlock.Outline = &willow.Outline{
		Color:     willow.Color{R: 0, G: 0, B: 0, A: 1},
		Thickness: 2,
	}
	outlined.X = 340
	outlined.Y = 272
	root.AddChild(outlined)

	// ── Divider label ─────────────────────────────────────────────────────────
	addSectionLabel(root, fontSmall, "Live update", 24, 388)

	// ── Animated frame counter ────────────────────────────────────────────────
	counter := willow.NewText("counter", "Frame: 0", fontMedium)
	counter.TextBlock.Color = willow.Color{R: 0.5, G: 1, B: 0.5, A: 1}
	counter.X = 24
	counter.Y = 408
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
func addSectionLabel(root *willow.Node, font willow.Font, text string, x, y float64) {
	n := willow.NewText("section-"+text, text, font)
	n.TextBlock.Color = willow.Color{R: 0.4, G: 0.5, B: 0.6, A: 1}
	n.X = x
	n.Y = y
	root.AddChild(n)
}
