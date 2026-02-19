package willow

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// CommandType identifies the kind of render command.
type CommandType uint8

const (
	CommandSprite   CommandType = iota // DrawImage
	CommandMesh                        // DrawTriangles
	CommandParticle                    // particle quads (batches as sprites)
)

// color32 is a compact RGBA color using float32, for render commands only.
type color32 struct {
	R, G, B, A float32
}

// RenderCommand is a single draw instruction emitted during scene traversal.
type RenderCommand struct {
	Type          CommandType
	Transform     [6]float32
	TextureRegion TextureRegion
	Color         color32
	BlendMode     BlendMode
	ShaderID      uint16
	TargetID      uint16
	RenderLayer   uint8
	GlobalOrder   int
	treeOrder     int // assigned during traversal for stable sort

	// Mesh-only fields (slice headers, not copies of vertex data).
	meshVerts []ebiten.Vertex
	meshInds  []uint16
	meshImage *ebiten.Image

	// directImage, when non-nil, is drawn directly instead of looking up an
	// atlas page. Used for cached/filtered/masked node output (Phase 09).
	directImage *ebiten.Image

	// emitter references the particle emitter for CommandParticle commands.
	emitter            *ParticleEmitter
	worldSpaceParticle bool // particles store world positions; Transform is view-only
}

// identityTransform32 is the identity affine matrix as float32.
var identityTransform32 = [6]float32{1, 0, 0, 1, 0, 0}

// affine32 converts a [6]float64 affine matrix to [6]float32.
func affine32(m [6]float64) [6]float32 {
	return [6]float32{float32(m[0]), float32(m[1]), float32(m[2]), float32(m[3]), float32(m[4]), float32(m[5])}
}

// traverse walks the node tree depth-first, updating transforms and emitting
// render commands for visible, renderable leaf nodes.
func (s *Scene) traverse(n *Node, parentTransform [6]float64, parentAlpha float64, parentRecomputed bool, treeOrder *int) {
	if !n.Visible {
		return
	}

	// Update world transform
	recompute := n.transformDirty || parentRecomputed
	if recompute {
		local := computeLocalTransform(n)
		n.worldTransform = multiplyAffine(parentTransform, local)
		n.worldAlpha = parentAlpha * n.Alpha
		n.transformDirty = false
	}

	// Determine if this node is culled. Culling only suppresses this node's
	// command emission — children are ALWAYS traversed because any node type
	// may have children whose world positions differ from the parent's AABB.
	culled := s.cullActive && n.Renderable && shouldCull(n, s.cullBounds)

	// Special path: nodes with masks, cache, or filters render their subtree
	// to an offscreen image and emit a single directImage command.
	if !culled && (n.mask != nil || n.cacheEnabled || len(n.Filters) > 0) {
		s.renderSpecialNode(n, treeOrder)
		return
	}

	// Emit command for renderable leaf-type nodes
	if n.Renderable && !culled {
		switch n.Type {
		case NodeTypeSprite:
			*treeOrder++
			cmd := RenderCommand{
				Type:        CommandSprite,
				Transform:   affine32(n.worldTransform),
				Color:       color32{float32(n.Color.R), float32(n.Color.G), float32(n.Color.B), float32(n.Color.A * n.worldAlpha)},
				BlendMode:   n.BlendMode,
				RenderLayer: n.RenderLayer,
				GlobalOrder: n.GlobalOrder,
				treeOrder:   *treeOrder,
			}
			if n.customImage != nil {
				cmd.directImage = n.customImage
			} else {
				cmd.TextureRegion = n.TextureRegion
			}
			s.commands = append(s.commands, cmd)
		case NodeTypeMesh:
			if len(n.Vertices) == 0 || len(n.Indices) == 0 {
				break
			}
			tintColor := Color{n.Color.R, n.Color.G, n.Color.B, n.Color.A * n.worldAlpha}
			dst := ensureTransformedVerts(n)
			transformVertices(n.Vertices, dst, n.worldTransform, tintColor)
			*treeOrder++
			s.commands = append(s.commands, RenderCommand{
				Type:        CommandMesh,
				Transform:   affine32(n.worldTransform),
				BlendMode:   n.BlendMode,
				RenderLayer: n.RenderLayer,
				GlobalOrder: n.GlobalOrder,
				treeOrder:   *treeOrder,
				meshVerts:   dst,
				meshInds:    n.Indices,
				meshImage:   n.MeshImage,
			})
		case NodeTypeParticleEmitter:
			if n.Emitter != nil && n.Emitter.alive > 0 {
				*treeOrder++
				particleTransform := n.worldTransform
				ws := n.Emitter.config.WorldSpace
				if ws {
					particleTransform = s.viewTransform
				}
				s.commands = append(s.commands, RenderCommand{
					Type:               CommandParticle,
					Transform:          affine32(particleTransform),
					TextureRegion:      n.TextureRegion,
					directImage:        n.customImage,
					Color:              color32{float32(n.Color.R), float32(n.Color.G), float32(n.Color.B), float32(n.Color.A * n.worldAlpha)},
					BlendMode:          n.BlendMode,
					RenderLayer:        n.RenderLayer,
					GlobalOrder:        n.GlobalOrder,
					treeOrder:          *treeOrder,
					emitter:            n.Emitter,
					worldSpaceParticle: ws,
				})
			}
		case NodeTypeText:
			if n.TextBlock != nil && n.TextBlock.Font != nil {
				switch n.TextBlock.Font.(type) {
				case *BitmapFont:
					s.commands = emitBitmapTextCommands(n.TextBlock, n, s.commands, treeOrder)
				case *TTFFont:
					s.commands, s.pages = emitTTFTextCommand(n.TextBlock, n, s.commands, treeOrder, s.pages, &s.nextPage)
				}
			}
			// NodeTypeContainer doesn't emit commands
		}
	}

	// Traverse children (ZIndex sorted if needed)
	if len(n.children) == 0 {
		return
	}
	children := n.children
	if !n.childrenSorted {
		s.rebuildSortedChildren(n)
	}
	if n.sortedChildren != nil {
		children = n.sortedChildren
	}
	for _, child := range children {
		s.traverse(child, n.worldTransform, n.worldAlpha, recompute, treeOrder)
	}
}

// rebuildSortedChildren rebuilds the ZIndex-sorted traversal order for a node.
// Uses insertion sort: zero allocations, stable, and optimal for the typical
// case of few children that are nearly sorted (O(n) when already sorted).
func (s *Scene) rebuildSortedChildren(n *Node) {
	nc := len(n.children)
	if cap(n.sortedChildren) < nc {
		n.sortedChildren = make([]*Node, nc)
	}
	n.sortedChildren = n.sortedChildren[:nc]
	copy(n.sortedChildren, n.children)
	// Stable insertion sort by ZIndex.
	for i := 1; i < nc; i++ {
		key := n.sortedChildren[i]
		j := i - 1
		for j >= 0 && n.sortedChildren[j].ZIndex > key.ZIndex {
			n.sortedChildren[j+1] = n.sortedChildren[j]
			j--
		}
		n.sortedChildren[j+1] = key
	}
	n.childrenSorted = true
}

// --- Merge sort ---

// commandLessOrEqual returns true if a should sort before or at the same position as b.
// Using <= for treeOrder ensures stability.
func commandLessOrEqual(a, b RenderCommand) bool {
	if a.RenderLayer != b.RenderLayer {
		return a.RenderLayer < b.RenderLayer
	}
	if a.GlobalOrder != b.GlobalOrder {
		return a.GlobalOrder < b.GlobalOrder
	}
	return a.treeOrder <= b.treeOrder
}

// mergeSort sorts s.commands in-place using s.sortBuf as scratch space.
// Bottom-up merge sort: zero allocations after the sort buffer reaches high-water mark.
func (s *Scene) mergeSort() {
	n := len(s.commands)
	if n <= 1 {
		return
	}
	if cap(s.sortBuf) < n {
		s.sortBuf = make([]RenderCommand, n)
	}
	s.sortBuf = s.sortBuf[:n]

	a := s.commands
	b := s.sortBuf
	swapped := false

	for width := 1; width < n; width *= 2 {
		for i := 0; i < n; i += 2 * width {
			lo := i
			mid := lo + width
			if mid > n {
				mid = n
			}
			hi := lo + 2*width
			if hi > n {
				hi = n
			}
			mergeRun(a, b, lo, mid, hi)
		}
		a, b = b, a
		swapped = !swapped
	}

	if swapped {
		copy(s.commands, s.sortBuf)
	}
}

// mergeRun merges two sorted runs [lo, mid) and [mid, hi) from src into dst.
func mergeRun(src, dst []RenderCommand, lo, mid, hi int) {
	i, j, k := lo, mid, lo
	for i < mid && j < hi {
		if commandLessOrEqual(src[i], src[j]) {
			dst[k] = src[i]
			i++
		} else {
			dst[k] = src[j]
			j++
		}
		k++
	}
	for i < mid {
		dst[k] = src[i]
		i++
		k++
	}
	for j < hi {
		dst[k] = src[j]
		j++
		k++
	}
}

// --- Special node rendering (Phase 09) ---

// renderSpecialNode handles nodes with masks, cache, or filters.
// Processing order: bounds → adjust transform → cache check → render subtree → apply mask → apply filters → cache store → emit command.
func (s *Scene) renderSpecialNode(n *Node, treeOrder *int) {
	// Compute bounds up-front — needed by all paths to build the correct
	// screen-space transform.  RT pixel (0,0) corresponds to local
	// (bounds.X, bounds.Y), not local (0,0), so the world transform must be
	// shifted by the world-space equivalent of that offset.
	bounds := subtreeBounds(n)
	padding := filterChainPadding(n.Filters)
	bounds.X -= float64(padding)
	bounds.Y -= float64(padding)
	bounds.Width += float64(padding * 2)
	bounds.Height += float64(padding * 2)

	// adjustedTransform places RT(0,0) = local(bounds.X, bounds.Y) at the
	// correct screen position: screen(tx + a*bX + c*bY, ty + b*bX + d*bY).
	bX, bY := bounds.X, bounds.Y
	a, b, c, d := n.worldTransform[0], n.worldTransform[1], n.worldTransform[2], n.worldTransform[3]
	adjustedTransform := n.worldTransform
	adjustedTransform[4] += a*bX + c*bY
	adjustedTransform[5] += b*bX + d*bY

	// Cache hit: reuse existing cached texture.
	at32 := affine32(adjustedTransform)
	if n.cacheEnabled && n.cacheTexture != nil && !n.cacheDirty {
		*treeOrder++
		s.commands = append(s.commands, RenderCommand{
			Type:        CommandSprite,
			Transform:   at32,
			Color:       color32{1, 1, 1, float32(n.worldAlpha)},
			BlendMode:   n.BlendMode,
			RenderLayer: n.RenderLayer,
			GlobalOrder: n.GlobalOrder,
			treeOrder:   *treeOrder,
			directImage: n.cacheTexture,
		})
		return
	}

	w := int(math.Ceil(bounds.Width))
	h := int(math.Ceil(bounds.Height))
	if w <= 0 || h <= 0 {
		return
	}

	// Render subtree to offscreen.
	rt := s.rtPool.Acquire(w, h)
	renderSubtree(s, n, rt, bounds)
	result := rt

	// Apply mask if present.
	if n.mask != nil {
		maskRT := s.rtPool.Acquire(w, h)
		renderSubtree(s, n.mask, maskRT, bounds)

		// Composite: keep only the parts of result where mask has alpha.
		var op ebiten.DrawImageOptions
		op.Blend = BlendMask.EbitenBlend()
		result.DrawImage(maskRT, &op)

		s.rtPool.Release(maskRT)
	}

	// Apply filter chain.
	if len(n.Filters) > 0 {
		filtered := applyFilters(n.Filters, result, &s.rtPool)
		if filtered != result {
			// The old result RT needs to be released.
			s.rtPool.Release(result)
			result = filtered
		}
	}

	// Cache the result if caching is enabled.
	if n.cacheEnabled {
		// Dispose old cache texture if present.
		if n.cacheTexture != nil {
			n.cacheTexture.Deallocate()
		}
		// Copy result to a non-pooled texture for caching.
		cacheImg := ebiten.NewImage(w, h)
		var op ebiten.DrawImageOptions
		cacheImg.DrawImage(result, &op)
		n.cacheTexture = cacheImg
		n.cacheDirty = false

		// Release the pooled RT immediately since we copied to cache.
		s.rtPool.Release(result)

		// Emit using the cached texture.
		*treeOrder++
		s.commands = append(s.commands, RenderCommand{
			Type:        CommandSprite,
			Transform:   at32,
			Color:       color32{1, 1, 1, float32(n.worldAlpha)},
			BlendMode:   n.BlendMode,
			RenderLayer: n.RenderLayer,
			GlobalOrder: n.GlobalOrder,
			treeOrder:   *treeOrder,
			directImage: n.cacheTexture,
		})
		return
	}

	// Not cached: defer release until after submitBatches.
	s.rtDeferred = append(s.rtDeferred, result)

	*treeOrder++
	s.commands = append(s.commands, RenderCommand{
		Type:        CommandSprite,
		Transform:   at32,
		Color:       color32{1, 1, 1, float32(n.worldAlpha)},
		BlendMode:   n.BlendMode,
		RenderLayer: n.RenderLayer,
		GlobalOrder: n.GlobalOrder,
		treeOrder:   *treeOrder,
		directImage: result,
	})
}
