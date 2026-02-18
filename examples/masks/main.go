// Masks demonstrates three node-masking techniques in Willow:
//   - Circle: a pulsing circular mask clips a rainbow tile grid.
//   - Star: a rotating star mask clips the whelp sprite.
//   - Scroll: a fixed rectangular mask clips upward-scrolling colour bars.
//
// Click any panel to remove its mask for 1.5 s and see the full content.
package main

import (
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/phanxgames/willow"
)

const (
	windowTitle = "Willow — Masks"
	showFPS     = true
	screenW     = 800
	screenH     = 520
)

// panelState holds the masked host node and its mask shape for one demo panel.
type panelState struct {
	host       *willow.Node
	maskNode   *willow.Node
	revealed   bool
	revealLeft float64
}

func (p *panelState) toggle() {
	if p.revealed {
		return
	}
	p.host.ClearMask()
	p.revealed = true
	p.revealLeft = 1.5
}

func (p *panelState) tick(dt float64) {
	if !p.revealed {
		return
	}
	p.revealLeft -= dt
	if p.revealLeft <= 0 {
		p.host.SetMask(p.maskNode)
		p.revealed = false
	}
}

type demo struct {
	panels     [3]*panelState
	scrollBars []*willow.Node
	time       float64
	scrollY    float64
}

func (d *demo) update() error {
	dt := 1.0 / float64(ebiten.TPS())
	d.time += dt

	// Panel 0: circle – gentle pulse.
	s0 := 1.0 + 0.22*math.Sin(d.time*1.7)
	d.panels[0].maskNode.ScaleX = s0
	d.panels[0].maskNode.ScaleY = s0
	d.panels[0].tick(dt)

	// Panel 1: star – rotate and pulse.
	d.panels[1].maskNode.Rotation = d.time * 0.55
	s1 := 1.0 + 0.12*math.Sin(d.time*2.3)
	d.panels[1].maskNode.ScaleX = s1
	d.panels[1].maskNode.ScaleY = s1
	d.panels[1].tick(dt)

	// Panel 2: scroll bars upward.
	barH := 30.0
	totalH := barH * float64(len(d.scrollBars))
	d.scrollY += 50 * dt
	for i, bar := range d.scrollBars {
		naturalY := -100.0 + float64(i)*barH - d.scrollY
		for naturalY < -100-barH {
			naturalY += totalH
		}
		bar.Y = naturalY
	}
	d.panels[2].tick(dt)

	return nil
}

func main() {
	f, err := os.Open("examples/_assets/whelp.png")
	if err != nil {
		log.Fatalf("open whelp.png: %v", err)
	}
	defer f.Close()

	whelpImg, _, err := ebitenutil.NewImageFromReader(f)
	if err != nil {
		log.Fatalf("decode whelp.png: %v", err)
	}

	scene := willow.NewScene()
	scene.ClearColor = willow.Color{R: 0.09, G: 0.09, B: 0.14, A: 1}

	d := &demo{}

	const panelY = 240.0
	panelXs := [3]float64{140, 400, 660}

	// ── Panel 0: pulsing circle mask over a rainbow tile grid ────────────────

	p0 := willow.NewContainer("circle-panel")
	p0.X = panelXs[0]
	p0.Y = panelY

	// 10×10 grid of 19 px tiles on a 20 px step, centred at local (0, 0).
	for row := 0; row < 10; row++ {
		for col := 0; col < 10; col++ {
			tile := willow.NewSprite("t", willow.TextureRegion{})
			tile.ScaleX = 19
			tile.ScaleY = 19
			tile.X = float64(col)*20 - 100
			tile.Y = float64(row)*20 - 100
			tile.Color = hsvToColor(float64(col+row)/18.0, 0.82, 0.92)
			p0.AddChild(tile)
		}
	}

	circleMask := willow.NewPolygon("circle-mask", circlePoints(90, 64))
	circleMask.Color = willow.Color{R: 1, G: 1, B: 1, A: 1}
	p0.SetMask(circleMask)

	p0.Interactable = true
	p0.HitShape = willow.HitCircle{CenterX: 0, CenterY: 0, Radius: 95}
	scene.Root().AddChild(p0)

	ps0 := &panelState{host: p0, maskNode: circleMask}
	d.panels[0] = ps0
	p0.OnClick = func(_ willow.ClickContext) { ps0.toggle() }

	// ── Panel 1: rotating star mask over the whelp sprite ────────────────────

	p1 := willow.NewContainer("star-panel")
	p1.X = panelXs[1]
	p1.Y = panelY

	whelp := willow.NewSprite("whelp", willow.TextureRegion{})
	whelp.SetCustomImage(whelpImg)
	whelp.X = -float64(whelpImg.Bounds().Dx()) / 2
	whelp.Y = -float64(whelpImg.Bounds().Dy()) / 2
	p1.AddChild(whelp)

	starMask := willow.NewPolygon("star-mask", starPoints(62, 26, 5))
	starMask.Color = willow.Color{R: 1, G: 1, B: 1, A: 1}
	p1.SetMask(starMask)

	hw := float64(whelpImg.Bounds().Dx()) / 2
	hh := float64(whelpImg.Bounds().Dy()) / 2
	p1.Interactable = true
	p1.HitShape = willow.HitRect{X: -hw, Y: -hh, Width: hw * 2, Height: hh * 2}
	scene.Root().AddChild(p1)

	ps1 := &panelState{host: p1, maskNode: starMask}
	d.panels[1] = ps1
	p1.OnClick = func(_ willow.ClickContext) { ps1.toggle() }

	// ── Panel 2: fixed rect mask with scrolling colour bars ──────────────────

	p2 := willow.NewContainer("scroll-panel")
	p2.X = panelXs[2]
	p2.Y = panelY

	barColors := []willow.Color{
		{R: 0.95, G: 0.25, B: 0.25, A: 1}, // red
		{R: 0.95, G: 0.55, B: 0.10, A: 1}, // orange
		{R: 0.95, G: 0.90, B: 0.10, A: 1}, // yellow
		{R: 0.25, G: 0.85, B: 0.30, A: 1}, // green
		{R: 0.15, G: 0.65, B: 0.95, A: 1}, // blue
		{R: 0.55, G: 0.25, B: 0.95, A: 1}, // purple
		{R: 0.95, G: 0.25, B: 0.75, A: 1}, // pink
		{R: 0.35, G: 0.90, B: 0.85, A: 1}, // cyan
	}
	for i, c := range barColors {
		bar := willow.NewSprite("bar", willow.TextureRegion{})
		bar.ScaleX = 136
		bar.ScaleY = 28
		bar.X = -68
		bar.Y = -100 + float64(i)*30
		bar.Color = c
		p2.AddChild(bar)
		d.scrollBars = append(d.scrollBars, bar)
	}

	rectMask := willow.NewPolygon("rect-mask", []willow.Vec2{
		{X: -70, Y: -100},
		{X: 70, Y: -100},
		{X: 70, Y: 100},
		{X: -70, Y: 100},
	})
	rectMask.Color = willow.Color{R: 1, G: 1, B: 1, A: 1}
	p2.SetMask(rectMask)

	p2.Interactable = true
	p2.HitShape = willow.HitRect{X: -70, Y: -100, Width: 140, Height: 200}
	scene.Root().AddChild(p2)

	ps2 := &panelState{host: p2, maskNode: rectMask}
	d.panels[2] = ps2
	p2.OnClick = func(_ willow.ClickContext) { ps2.toggle() }

	// ── Labels ────────────────────────────────────────────────────────────────

	for i, lbl := range []string{"Circle Mask", "Star Mask", "Scroll Mask"} {
		n := makeLabel(lbl)
		n.X = panelXs[i] - float64(len(lbl)*6)/2
		n.Y = panelY + 112
		scene.Root().AddChild(n)
	}

	const hintText = "click any panel to reveal unmasked content for 1.5 s"
	hint := makeLabel(hintText)
	hint.X = float64(screenW)/2 - float64(len(hintText)*6)/2
	hint.Y = float64(screenH) - 26
	scene.Root().AddChild(hint)

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

// circlePoints returns n evenly-spaced points on a circle of the given radius.
func circlePoints(radius float64, n int) []willow.Vec2 {
	pts := make([]willow.Vec2, n)
	for i := range pts {
		a := float64(i) * 2 * math.Pi / float64(n)
		pts[i] = willow.Vec2{X: math.Cos(a) * radius, Y: math.Sin(a) * radius}
	}
	return pts
}

// starPoints returns vertices for a numPoints-pointed star.
// outerR is the tip radius; innerR is the valley radius.
func starPoints(outerR, innerR float64, numPoints int) []willow.Vec2 {
	pts := make([]willow.Vec2, numPoints*2)
	for i := range pts {
		a := -math.Pi/2 + float64(i)*math.Pi/float64(numPoints)
		r := outerR
		if i%2 == 1 {
			r = innerR
		}
		pts[i] = willow.Vec2{X: math.Cos(a) * r, Y: math.Sin(a) * r}
	}
	return pts
}

// hsvToColor converts HSV (all in [0, 1]) to willow.Color.
func hsvToColor(h, s, v float64) willow.Color {
	h6 := h * 6
	i := int(h6)
	f := h6 - float64(i)
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))
	switch i % 6 {
	case 0:
		return willow.Color{R: v, G: t, B: p, A: 1}
	case 1:
		return willow.Color{R: q, G: v, B: p, A: 1}
	case 2:
		return willow.Color{R: p, G: v, B: t, A: 1}
	case 3:
		return willow.Color{R: p, G: q, B: v, A: 1}
	case 4:
		return willow.Color{R: t, G: p, B: v, A: 1}
	default:
		return willow.Color{R: v, G: p, B: q, A: 1}
	}
}

// makeLabel renders s into a small sprite using the Ebitengine debug font.
func makeLabel(s string) *willow.Node {
	img := ebiten.NewImage(len(s)*6+4, 16)
	ebitenutil.DebugPrint(img, s)
	n := willow.NewSprite("label", willow.TextureRegion{})
	n.SetCustomImage(img)
	return n
}
