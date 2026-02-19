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

	// Resolve the particle texture. A directImage (e.g. WhitePixel) takes
	// priority; otherwise fall back to an atlas page.
	var subImg *ebiten.Image
	if cmd.directImage != nil {
		subImg = cmd.directImage
	} else {
		var page *ebiten.Image
		if r.Page == magentaPlaceholderPage {
			page = ensureMagentaImage()
		} else if int(r.Page) < len(s.pages) {
			page = s.pages[r.Page]
		}
		if page == nil {
			return
		}
		var subRect image.Rectangle
		if r.Rotated {
			subRect = image.Rect(int(r.X), int(r.Y), int(r.X)+int(r.Height), int(r.Y)+int(r.Width))
		} else {
			subRect = image.Rect(int(r.X), int(r.Y), int(r.X)+int(r.Width), int(r.Y)+int(r.Height))
		}
		subImg = page.SubImage(subRect).(*ebiten.Image)
	}

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

// --- Coalesced batching (BatchModeCoalesced) ---

// submitBatchesCoalesced iterates sorted commands, coalescing consecutive
// same-key atlas sprites into a single DrawTriangles32 call.
func (s *Scene) submitBatchesCoalesced(target *ebiten.Image) {
	if len(s.commands) == 0 {
		return
	}

	s.batchVerts = s.batchVerts[:0]
	s.batchInds = s.batchInds[:0]

	var currentKey batchKey
	inRun := false
	var op ebiten.DrawImageOptions

	for i := range s.commands {
		cmd := &s.commands[i]

		switch cmd.Type {
		case CommandSprite:
			if cmd.directImage != nil {
				// Direct-image sprites cannot be coalesced (different source images).
				s.flushSpriteBatch(target, currentKey)
				inRun = false
				s.submitSprite(target, cmd, &op)
				continue
			}

			key := commandBatchKey(cmd)
			if inRun && key != currentKey {
				s.flushSpriteBatch(target, currentKey)
			}
			currentKey = key
			inRun = true
			s.appendSpriteQuad(cmd)

		case CommandParticle:
			s.flushSpriteBatch(target, currentKey)
			inRun = false
			s.submitParticlesBatched(target, cmd)

		case CommandMesh:
			s.flushSpriteBatch(target, currentKey)
			inRun = false
			s.submitMesh(target, cmd)
		}
	}

	s.flushSpriteBatch(target, currentKey)
}

// appendSpriteQuad appends 4 vertices and 6 indices for a single atlas sprite.
func (s *Scene) appendSpriteQuad(cmd *RenderCommand) {
	r := &cmd.TextureRegion
	t := &cmd.Transform // [a, b, c, d, tx, ty]

	// Local quad corners before world transform.
	// Visual dimensions are always (r.Width, r.Height) for non-rotated,
	// and (r.Width, r.Height) for rotated (the original authored size).
	// Trim offset shifts the local origin.
	ox := float64(r.OffsetX)
	oy := float64(r.OffsetY)
	var w, h float64
	if r.Rotated {
		// Rotated region: visual dimensions are the original Width × Height.
		// The atlas stores them as Height × Width (swapped).
		w = float64(r.Width)
		h = float64(r.Height)
	} else {
		w = float64(r.Width)
		h = float64(r.Height)
	}

	// 4 local positions: TL, TR, BL, BR
	lx := [4]float64{ox, ox + w, ox, ox + w}
	ly := [4]float64{oy, oy, oy + h, oy + h}

	// Apply affine transform: dx = a*lx + c*ly + tx, dy = b*lx + d*ly + ty
	a, b, c, d, tx, ty := t[0], t[1], t[2], t[3], t[4], t[5]

	// Source UVs (pixel coordinates on the atlas page).
	var sx, sy [4]float32
	if r.Rotated {
		// Rotated region stored at (r.X, r.Y) with stored rect width=r.Height, height=r.Width.
		// The rotation undoes the 90° CW storage rotation.
		// Mapping visual corners → atlas coords:
		//   Visual TL (ox, oy)       → atlas (r.X + r.Height, r.Y)
		//   Visual TR (ox + w, oy)   → atlas (r.X + r.Height, r.Y + r.Width)
		//   Visual BL (ox, oy + h)   → atlas (r.X, r.Y)
		//   Visual BR (ox + w, oy+h) → atlas (r.X, r.Y + r.Width)
		rx := float32(r.X)
		ry := float32(r.Y)
		rh := float32(r.Height) // stored width in atlas
		rw := float32(r.Width)  // stored height in atlas
		sx = [4]float32{rx + rh, rx + rh, rx, rx}
		sy = [4]float32{ry, ry + rw, ry, ry + rw}
	} else {
		rx := float32(r.X)
		ry := float32(r.Y)
		rw := float32(r.Width)
		rh := float32(r.Height)
		sx = [4]float32{rx, rx + rw, rx, rx + rw}
		sy = [4]float32{ry, ry, ry + rh, ry + rh}
	}

	// Premultiplied RGBA. Zero-color sentinel → opaque white.
	var cr, cg, cb, ca float32
	ca = float32(cmd.Color.A)
	if ca == 0 && cmd.Color.R == 0 && cmd.Color.G == 0 && cmd.Color.B == 0 {
		cr, cg, cb, ca = 1, 1, 1, 1
	} else {
		cr = float32(cmd.Color.R) * ca
		cg = float32(cmd.Color.G) * ca
		cb = float32(cmd.Color.B) * ca
	}

	base := uint32(len(s.batchVerts))

	for i := 0; i < 4; i++ {
		dx := float32(a*lx[i] + c*ly[i] + tx)
		dy := float32(b*lx[i] + d*ly[i] + ty)
		s.batchVerts = append(s.batchVerts, ebiten.Vertex{
			DstX:   dx,
			DstY:   dy,
			SrcX:   sx[i],
			SrcY:   sy[i],
			ColorR: cr,
			ColorG: cg,
			ColorB: cb,
			ColorA: ca,
		})
	}

	// Two triangles: TL-TR-BL, TR-BR-BL
	s.batchInds = append(s.batchInds,
		base+0, base+1, base+2,
		base+1, base+3, base+2,
	)
}

// flushSpriteBatch submits accumulated vertices as a single DrawTriangles32 call.
func (s *Scene) flushSpriteBatch(target *ebiten.Image, key batchKey) {
	if len(s.batchVerts) == 0 {
		return
	}

	var page *ebiten.Image
	if key.page == magentaPlaceholderPage {
		page = ensureMagentaImage()
	} else if int(key.page) < len(s.pages) {
		page = s.pages[key.page]
	}
	if page == nil {
		s.batchVerts = s.batchVerts[:0]
		s.batchInds = s.batchInds[:0]
		return
	}

	var triOp ebiten.DrawTrianglesOptions
	triOp.Blend = key.blend.EbitenBlend()
	triOp.ColorScaleMode = ebiten.ColorScaleModePremultipliedAlpha

	target.DrawTriangles32(s.batchVerts, s.batchInds, page, &triOp)

	s.batchVerts = s.batchVerts[:0]
	s.batchInds = s.batchInds[:0]
}

// submitParticlesBatched draws all alive particles using a single DrawTriangles32 call.
func (s *Scene) submitParticlesBatched(target *ebiten.Image, cmd *RenderCommand) {
	e := cmd.emitter
	if e == nil || e.alive == 0 {
		return
	}

	r := &cmd.TextureRegion

	// Resolve the source image and UV bounds.
	var srcImg *ebiten.Image
	var su0, sv0, su1, sv1 float32

	if cmd.directImage != nil {
		srcImg = cmd.directImage
		b := srcImg.Bounds()
		su0, sv0 = float32(b.Min.X), float32(b.Min.Y)
		su1, sv1 = float32(b.Max.X), float32(b.Max.Y)
	} else {
		if r.Page == magentaPlaceholderPage {
			srcImg = ensureMagentaImage()
		} else if int(r.Page) < len(s.pages) {
			srcImg = s.pages[r.Page]
		}
		if srcImg == nil {
			return
		}
		if r.Rotated {
			su0, sv0 = float32(r.X), float32(r.Y)
			su1, sv1 = float32(r.X)+float32(r.Height), float32(r.Y)+float32(r.Width)
		} else {
			su0, sv0 = float32(r.X), float32(r.Y)
			su1, sv1 = float32(r.X)+float32(r.Width), float32(r.Y)+float32(r.Height)
		}
	}

	// Base transform components.
	bt := &cmd.Transform
	ba, bb, bc, bd, btx, bty := bt[0], bt[1], bt[2], bt[3], bt[4], bt[5]

	ow := float64(r.OriginalW)
	oh := float64(r.OriginalH)
	halfW := ow / 2
	halfH := oh / 2
	offX := float64(r.OffsetX)
	offY := float64(r.OffsetY)

	s.batchVerts = s.batchVerts[:0]
	s.batchInds = s.batchInds[:0]

	for i := 0; i < e.alive; i++ {
		p := &e.particles[i]

		// Build per-particle local→world transform.
		// Steps (matching the immediate-mode path):
		// 1. Trim offset
		// 2. Scale around sprite center: translate(-halfW, -halfH), scale(s,s), translate(halfW, halfH)
		// 3. Translate by particle position (p.x, p.y)
		// 4. Concat base transform

		// Combined local transform for steps 1-3:
		// After offset: (x + offX, y + offY)
		// After center-scale: ((x + offX - halfW)*s + halfW, (y + offY - halfH)*s + halfH)
		//                   = (x*s + (offX - halfW)*s + halfW, y*s + (offY - halfH)*s + halfH)
		// After particle translate: above + (p.x, p.y)
		// Local transform is: scale=s, tx=(offX - halfW)*s + halfW + p.x, ty=(offY - halfH)*s + halfH + p.y

		ps := p.scale
		localTx := (offX-halfW)*ps + halfW + p.x
		localTy := (offY-halfH)*ps + halfH + p.y

		// Concat with base: M_base * M_local
		// M_local = [ps, 0, 0, ps, localTx, localTy]
		// Result: [ba*ps, bb*ps, bc*ps, bd*ps, ba*localTx + bc*localTy + btx, bb*localTx + bd*localTy + bty]
		fa := ba * ps
		fb := bb * ps
		fc := bc * ps
		fd := bd * ps
		ftx := ba*localTx + bc*localTy + btx
		fty := bb*localTx + bd*localTy + bty

		// Handle rotated regions in UVs.
		var psx, psy [4]float32
		if cmd.directImage != nil {
			// Direct image: simple UV mapping.
			psx = [4]float32{su0, su1, su0, su1}
			psy = [4]float32{sv0, sv0, sv1, sv1}
		} else if r.Rotated {
			psx = [4]float32{su1, su1, su0, su0}
			psy = [4]float32{sv0, sv1, sv0, sv1}
		} else {
			psx = [4]float32{su0, su1, su0, su1}
			psy = [4]float32{sv0, sv0, sv1, sv1}
		}

		// Local quad: use the visual sprite dimensions (Width × Height for non-rotated).
		var qw, qh float64
		if cmd.directImage != nil {
			qw = float64(su1 - su0)
			qh = float64(sv1 - sv0)
		} else {
			qw = float64(r.Width)
			qh = float64(r.Height)
		}

		// 4 local positions: (0,0), (qw,0), (0,qh), (qw,qh)
		// After full transform: dx = fa*lx + fc*ly + ftx, dy = fb*lx + fd*ly + fty
		qlx := [4]float64{0, qw, 0, qw}
		qly := [4]float64{0, 0, qh, qh}

		// Per-particle color
		cr := float32(p.colorR*float64(cmd.Color.R)) * float32(p.alpha*float64(cmd.Color.A))
		cg := float32(p.colorG*float64(cmd.Color.G)) * float32(p.alpha*float64(cmd.Color.A))
		cb := float32(p.colorB*float64(cmd.Color.B)) * float32(p.alpha*float64(cmd.Color.A))
		ca := float32(p.alpha * float64(cmd.Color.A))

		base := uint32(len(s.batchVerts))

		for j := 0; j < 4; j++ {
			dx := float32(fa*qlx[j] + fc*qly[j] + ftx)
			dy := float32(fb*qlx[j] + fd*qly[j] + fty)
			s.batchVerts = append(s.batchVerts, ebiten.Vertex{
				DstX:   dx,
				DstY:   dy,
				SrcX:   psx[j],
				SrcY:   psy[j],
				ColorR: cr,
				ColorG: cg,
				ColorB: cb,
				ColorA: ca,
			})
		}

		s.batchInds = append(s.batchInds,
			base+0, base+1, base+2,
			base+1, base+3, base+2,
		)
	}

	if len(s.batchVerts) == 0 {
		return
	}

	var triOp ebiten.DrawTrianglesOptions
	triOp.Blend = cmd.BlendMode.EbitenBlend()
	triOp.ColorScaleMode = ebiten.ColorScaleModePremultipliedAlpha

	target.DrawTriangles32(s.batchVerts, s.batchInds, srcImg, &triOp)

	s.batchVerts = s.batchVerts[:0]
	s.batchInds = s.batchInds[:0]
}
