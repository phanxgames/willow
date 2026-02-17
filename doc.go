// Package willow is a retained-mode 2D scene graph, interaction layer, and
// render compiler for [Ebitengine].
//
// Willow provides the scene graph, transform hierarchy, sprite batching, input
// handling, camera viewports, text rendering, and particle systems that every
// non-trivial 2D project needs — packaged as a single, focused library.
//
// # Game loop
//
// There are two ways to integrate Willow into your application.
//
// The simplest is [Run], which creates a window and game loop for you:
//
//	scene := willow.NewScene()
//	// ... add nodes ...
//	willow.Run(scene, willow.RunConfig{
//		Title: "My Game", Width: 640, Height: 480,
//	})
//
// Register a callback with [Scene.SetUpdateFunc] to run your game logic
// each tick:
//
//	scene.SetUpdateFunc(func() error {
//		hero.X += speed
//		hero.MarkDirty()
//		return nil
//	})
//
// For full control, implement [ebiten.Game] yourself and call
// [Scene.Update] and [Scene.Draw] directly:
//
//	type Game struct{ scene *willow.Scene }
//
//	func (g *Game) Update() error         { g.scene.Update(); return nil }
//	func (g *Game) Draw(s *ebiten.Image)  { g.scene.Draw(s) }
//	func (g *Game) Layout(w, h int) (int, int) { return w, h }
//
// Both approaches are fully supported. [Run] is a thin wrapper around
// ebiten.RunGame — it does not add hidden behavior.
//
// # Nodes and the scene graph
//
// Every visual element is a [Node]. Nodes form a tree rooted at
// [Scene.Root]. Children inherit their parent's transform (position,
// rotation, scale, skew) and alpha.
//
// Create nodes with the typed constructors:
//
//   - [NewContainer] — group node, no visual output
//   - [NewSprite] — renders a [TextureRegion] (or a solid color via [WhitePixel])
//   - [NewText] — renders text using a [BitmapFont] or [TTFFont]
//   - [NewParticleEmitter] — CPU-simulated particle system
//   - [NewMesh] — low-level DrawTriangles
//   - [NewPolygon] / [NewPolygonTextured] — filled polygon helper
//
// Add nodes to the tree with [Node.AddChild]:
//
//	container := willow.NewContainer("ui")
//	scene.Root().AddChild(container)
//
//	btn := willow.NewSprite("btn", willow.TextureRegion{})
//	btn.X, btn.Y = 100, 50
//	container.AddChild(btn)
//
// # Sprites and solid colors
//
// [NewSprite] with a zero-value [TextureRegion] produces a 1x1 white pixel.
// Use [Node.ScaleX], [Node.ScaleY], and [Node.Color] to turn it into a
// solid-color rectangle of any size:
//
//	box := willow.NewSprite("box", willow.TextureRegion{})
//	box.ScaleX = 80
//	box.ScaleY = 40
//	box.Color = willow.Color{R: 0.3, G: 0.7, B: 1, A: 1}
//
// For atlas-based sprites, load a [TexturePacker] JSON with [LoadAtlas]
// and look up regions by name:
//
//	atlas, _ := willow.LoadAtlas(jsonBytes, pages)
//	hero := willow.NewSprite("hero", atlas.Region("hero_idle"))
//
// # Transforms
//
// Each node has local transform fields: X, Y, ScaleX, ScaleY, Rotation,
// SkewX, SkewY, PivotX, PivotY. World transforms are computed lazily via
// dirty-flag propagation — static subtrees skip recomputation entirely.
//
// Call [Node.MarkDirty] after changing a transform field at runtime
// (X, Y, Rotation, etc.) to ensure the change takes effect:
//
//	sprite.X += 5
//	sprite.MarkDirty()
//
// PivotX and PivotY set the transform origin as a fraction of the node's
// size (0.5, 0.5 = center). By default pivots are (0, 0) meaning the
// top-left corner.
//
// # Cameras and viewports
//
// A [Camera] defines a viewport rectangle on screen and a position in world
// space. Create cameras with [Scene.NewCamera]:
//
//	cam := scene.NewCamera(willow.Rect{X: 0, Y: 0, Width: 640, Height: 480})
//	cam.X = worldCenterX
//	cam.Y = worldCenterY
//	cam.MarkDirty()
//
// Cameras support smooth follow, animated scroll-to with 45+ easing
// functions, zoom, rotation, bounds clamping, and frustum culling.
// Convert between coordinate spaces with [Camera.WorldToScreen] and
// [Camera.ScreenToWorld].
//
// When no cameras are created, the scene renders with a default 1:1
// mapping from world to screen coordinates.
//
// # Interaction and input
//
// Set [Node.Interactable] to true and assign a [HitShape] to make a node
// respond to pointer input. Assign callback functions directly to node
// fields:
//
//	btn.Interactable = true
//	btn.HitShape = willow.HitRect{Width: 80, Height: 40}
//
//	btn.OnClick = func(ctx willow.ClickContext) {
//		fmt.Println("clicked at", ctx.WorldX, ctx.WorldY)
//	}
//	btn.OnDrag = func(ctx willow.DragContext) {
//		btn.X += ctx.DeltaX
//		btn.Y += ctx.DeltaY
//		btn.MarkDirty()
//	}
//
// Available callbacks: OnPointerDown, OnPointerUp, OnPointerMove,
// OnClick, OnDragStart, OnDrag, OnDragEnd, OnPinch, OnPointerEnter,
// OnPointerLeave.
//
// Hit testing runs in reverse painter order (front-to-back). Built-in
// hit shapes are [HitRect], [HitCircle], and [HitPolygon].
//
// # Per-node update callbacks
//
// Each node has an [Node.OnUpdate] callback that fires every frame during
// [Scene.Update]. Use it for self-contained per-node logic like spinning,
// bobbing, or AI:
//
//	sprite.OnUpdate = func(dt float64) {
//		sprite.Rotation += 0.02
//	}
//
// # Filters and visual effects
//
// Assign a slice of [Filter] implementations to [Node.Filters] to apply
// shader effects. Filters are applied in order during rendering:
//
//	sprite.Filters = []willow.Filter{
//		willow.NewBlurFilter(3),
//		willow.NewOutlineFilter(2, willow.Color{R: 1, G: 1, B: 0, A: 1}),
//	}
//
// Built-in filters: [ColorMatrixFilter] (brightness, contrast,
// saturation, and arbitrary 4x5 matrix), [BlurFilter], [OutlineFilter],
// [PixelPerfectOutlineFilter], [PixelPerfectInlineFilter],
// [PaletteFilter], and [CustomShaderFilter] for user-provided Kage
// shaders.
//
// Nodes can also be cached as textures ([Node.SetCacheAsTexture]) or
// masked by another node ([Node.SetMask]).
//
// # Animation and tweens
//
// Willow includes tween helpers built on [gween]:
//
//	tween := willow.TweenPosition(sprite, 500, 300, 1.5, ease.OutBounce)
//
// Advance tweens manually each frame:
//
//	dt := float32(1.0 / float64(ebiten.TPS()))
//	tween.Update(dt)
//	if tween.Done { /* finished */ }
//
// Convenience constructors: [TweenPosition], [TweenScale],
// [TweenRotation], [TweenAlpha], [TweenColor]. Tweens automatically
// stop if their target node is disposed.
//
// There is no global tween manager — you own the [TweenGroup] values and
// call Update yourself, giving you full control over timing and lifetime.
//
// # Text rendering
//
// Create text nodes with [NewText] and either a [BitmapFont] (pixel-perfect
// BMFont .fnt files) or a [TTFFont] (Ebitengine text/v2 wrapper):
//
//	font, _ := willow.LoadBitmapFont(fntData, pageImages)
//	label := willow.NewText("title", "Hello Willow", font)
//	label.TextBlock.Align = willow.TextAlignCenter
//	label.TextBlock.WrapWidth = 200
//
// # Particles
//
// Create emitters with [NewParticleEmitter] and an [EmitterConfig]:
//
//	emitter := willow.NewParticleEmitter("sparks", willow.EmitterConfig{
//		MaxParticles: 200,
//		EmitRate:     60,
//		Lifetime:     willow.Range{Min: 0.5, Max: 1.5},
//		Speed:        willow.Range{Min: 50, Max: 150},
//	})
//	scene.Root().AddChild(emitter)
//	emitter.Emitter.Start()
//
// Particles are CPU-simulated in Update (not Draw) with preallocated
// pools. Set WorldSpace to true in the config to emit particles in world
// coordinates instead of local to the emitter.
//
// # Meshes and polygons
//
// For custom geometry, use [NewMesh] with vertex and index buffers,
// or the higher-level helpers:
//
//   - [NewPolygon] — solid-color convex polygon
//   - [NewPolygonTextured] — textured convex polygon
//   - [NewRope] — textured ribbon along a point path
//   - [NewDistortionGrid] — deformable texture grid
//
// # Blend modes
//
// Set [Node.BlendMode] to control compositing. Available modes:
// [BlendNormal] (default source-over), [BlendAdd], [BlendMultiply],
// [BlendScreen], [BlendErase], [BlendMask], [BlendBelow], [BlendNone].
//
// # Draw ordering
//
// Siblings are drawn in child-list order by default. Use [Node.SetZIndex]
// to reorder among siblings without changing the tree. For coarser
// control, [Node.RenderLayer] provides a primary sort key across the
// entire scene.
//
// # ECS integration
//
// Set a node's [Node.EntityID] and call [Scene.SetEntityStore] with an
// [EntityStore] implementation. Interaction events are forwarded as
// [InteractionEvent] values, bridging the scene graph into your
// entity-component system. A [Donburi] adapter ships in the willow/ecs
// submodule.
//
// # Performance
//
// Willow targets zero heap allocations per frame on the hot path. Render
// commands, sort buffers, and vertex arrays are preallocated and reused.
// Dirty flags skip transform recomputation for static subtrees.
// Typed callback slices avoid interface boxing in event dispatch.
//
// [Ebitengine]: https://ebitengine.org
// [TexturePacker]: https://www.codeandweb.com/texturepacker
// [gween]: https://github.com/tanema/gween
// [Donburi]: https://github.com/yohamta/donburi
package willow
