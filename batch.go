package willow

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// batchKey groups render commands that can be submitted in a single draw call.
type batchKey struct {
	targetID uint16
	shaderID uint16
	blend    BlendMode
	page     uint16
}

func commandBatchKey(cmd *RenderCommand) batchKey {
	return batchKey{
		targetID: cmd.TargetID,
		shaderID: cmd.ShaderID,
		blend:    cmd.BlendMode,
		page:     cmd.TextureRegion.Page,
	}
}

// submitBatches iterates sorted commands, groups them by batch key, and submits
// draw calls to the target image.
func (s *Scene) submitBatches(target *ebiten.Image) {
	if len(s.commands) == 0 {
		return
	}

	var op ebiten.DrawImageOptions

	for i := range s.commands {
		cmd := &s.commands[i]

		switch cmd.Type {
		case CommandSprite:
			s.submitSprite(target, cmd, &op)
		case CommandParticle:
			s.submitParticles(target, cmd, &op)
		case CommandMesh:
			s.submitMesh(target, cmd)
		}
	}
}

// submitSprite draws a single sprite command using DrawImage.
func (s *Scene) submitSprite(target *ebiten.Image, cmd *RenderCommand, op *ebiten.DrawImageOptions) {
	// Direct image path: draw a pre-rendered offscreen texture directly.
	if cmd.directImage != nil {
		op.GeoM.Reset()
		op.GeoM.Concat(commandGeoM(cmd))
		op.ColorScale.Reset()
		a := float32(cmd.Color.A)
		if a == 0 && cmd.Color.R == 0 && cmd.Color.G == 0 && cmd.Color.B == 0 {
			a = 1
			op.ColorScale.Scale(1, 1, 1, 1)
		} else {
			op.ColorScale.Scale(float32(cmd.Color.R)*a, float32(cmd.Color.G)*a, float32(cmd.Color.B)*a, a)
		}
		op.Blend = cmd.BlendMode.EbitenBlend()
		target.DrawImage(cmd.directImage, op)
		return
	}

	r := &cmd.TextureRegion

	// Resolve the atlas page image.
	var page *ebiten.Image
	if r.Page == magentaPlaceholderPage {
		page = ensureMagentaImage()
	} else if int(r.Page) < len(s.pages) {
		page = s.pages[r.Page]
	}
	if page == nil {
		return
	}

	// Compute SubImage rect
	var subRect image.Rectangle
	if r.Rotated {
		subRect = image.Rect(int(r.X), int(r.Y), int(r.X)+int(r.Height), int(r.Y)+int(r.Width))
	} else {
		subRect = image.Rect(int(r.X), int(r.Y), int(r.X)+int(r.Width), int(r.Y)+int(r.Height))
	}
	subImg := page.SubImage(subRect).(*ebiten.Image)

	op.GeoM.Reset()

	// Handle rotated regions: rotate -90° and shift
	if r.Rotated {
		// Rotated regions in atlas are stored rotated 90° CW.
		// To draw correctly: rotate -90° (CCW) then shift right by height.
		op.GeoM.Rotate(-1.5707963267948966) // -π/2
		op.GeoM.Translate(0, float64(r.Width))
	}

	// Apply trim offset
	if r.OffsetX != 0 || r.OffsetY != 0 {
		op.GeoM.Translate(float64(r.OffsetX), float64(r.OffsetY))
	}

	// Apply world transform
	op.GeoM.Concat(commandGeoM(cmd))

	// Apply premultiplied color scale
	op.ColorScale.Reset()
	a := float32(cmd.Color.A)
	if a == 0 && cmd.Color.R == 0 && cmd.Color.G == 0 && cmd.Color.B == 0 {
		a = 1
		op.ColorScale.Scale(1, 1, 1, 1)
	} else {
		op.ColorScale.Scale(float32(cmd.Color.R)*a, float32(cmd.Color.G)*a, float32(cmd.Color.B)*a, a)
	}

	op.Blend = cmd.BlendMode.EbitenBlend()

	target.DrawImage(subImg, op)
}

// submitParticles draws all alive particles for a CommandParticle command.
func (s *Scene) submitParticles(target *ebiten.Image, cmd *RenderCommand, op *ebiten.DrawImageOptions) {
	e := cmd.emitter
	if e == nil || e.alive == 0 {
		return
	}

	r := &cmd.TextureRegion

	// Resolve the atlas page image.
	var page *ebiten.Image
	if r.Page == magentaPlaceholderPage {
		page = ensureMagentaImage()
	} else if int(r.Page) < len(s.pages) {
		page = s.pages[r.Page]
	}
	if page == nil {
		return
	}

	// Compute SubImage rect.
	var subRect image.Rectangle
	if r.Rotated {
		subRect = image.Rect(int(r.X), int(r.Y), int(r.X)+int(r.Height), int(r.Y)+int(r.Width))
	} else {
		subRect = image.Rect(int(r.X), int(r.Y), int(r.X)+int(r.Width), int(r.Y)+int(r.Height))
	}
	subImg := page.SubImage(subRect).(*ebiten.Image)

	// Transform for positioning: world transform for attached, view-only for world-space.
	baseGeoM := commandGeoM(cmd)

	for i := 0; i < e.alive; i++ {
		p := &e.particles[i]

		op.GeoM.Reset()

		// Handle rotated regions.
		if r.Rotated {
			op.GeoM.Rotate(-1.5707963267948966) // -π/2
			op.GeoM.Translate(0, float64(r.Width))
		}

		// Apply trim offset.
		if r.OffsetX != 0 || r.OffsetY != 0 {
			op.GeoM.Translate(float64(r.OffsetX), float64(r.OffsetY))
		}

		// Per-particle scale (around sprite center).
		op.GeoM.Translate(-float64(r.OriginalW)/2, -float64(r.OriginalH)/2)
		op.GeoM.Scale(p.scale, p.scale)
		op.GeoM.Translate(float64(r.OriginalW)/2, float64(r.OriginalH)/2)

		// Per-particle translation.
		// For world-space particles, (p.x, p.y) are absolute world coords.
		// For attached particles, (p.x, p.y) are relative to the emitter.
		op.GeoM.Translate(p.x, p.y)

		// Apply base transform (emitter world transform or view-only transform).
		op.GeoM.Concat(baseGeoM)

		// Per-particle color: particle color * emitter node color, particle alpha * emitter worldAlpha.
		cr := float32(p.colorR * float64(cmd.Color.R))
		cg := float32(p.colorG * float64(cmd.Color.G))
		cb := float32(p.colorB * float64(cmd.Color.B))
		ca := float32(p.alpha * float64(cmd.Color.A))
		op.ColorScale.Reset()
		op.ColorScale.Scale(cr*ca, cg*ca, cb*ca, ca)

		op.Blend = cmd.BlendMode.EbitenBlend()

		target.DrawImage(subImg, op)
	}
}

// submitMesh draws a mesh command using DrawTriangles.
func (s *Scene) submitMesh(target *ebiten.Image, cmd *RenderCommand) {
	if cmd.meshImage == nil || len(cmd.meshVerts) == 0 || len(cmd.meshInds) == 0 {
		return
	}

	var triOp ebiten.DrawTrianglesOptions
	triOp.Blend = cmd.BlendMode.EbitenBlend()

	target.DrawTriangles(cmd.meshVerts, cmd.meshInds, cmd.meshImage, &triOp)
}

// commandGeoM converts a command's [6]float64 transform into an ebiten.GeoM.
func commandGeoM(cmd *RenderCommand) ebiten.GeoM {
	var m ebiten.GeoM
	m.SetElement(0, 0, cmd.Transform[0])
	m.SetElement(1, 0, cmd.Transform[1])
	m.SetElement(0, 1, cmd.Transform[2])
	m.SetElement(1, 1, cmd.Transform[3])
	m.SetElement(0, 2, cmd.Transform[4])
	m.SetElement(1, 2, cmd.Transform[5])
	return m
}
