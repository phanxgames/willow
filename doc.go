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
//		dt := float32(1.0 / float64(ebiten.TPS()))
//		myTween.Update(dt)
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
// # Lighting
//
// A [LightLayer] creates a dark overlay with soft circles of light cut
// through it. Add the layer as the topmost scene node so it composites
// over all content:
//
//	ll := willow.NewLightLayer(640, 480, 0.9) // ambientAlpha: 1=black, 0=fully lit
//	scene.Root().AddChild(ll.Node())
//
// Add point lights with [LightLayer.AddLight]:
//
//	torch := &willow.Light{
//		X: 200, Y: 300,
//		Radius: 120, Intensity: 1.0,
//		Color:   willow.Color{R: 1, G: 0.7, B: 0.3, A: 1},
//		Enabled: true,
//	}
//	ll.AddLight(torch)
//
// To attach a light to a moving node, set [Light.Target] instead of
// fixed X/Y coordinates. Call [LightLayer.Redraw] each frame to
// rebuild the light texture before [Scene.Draw].
//
// # ECS integration
//
// Set a node's [Node.EntityID] and call [Scene.SetEntityStore] with an
// [EntityStore] implementation. Interaction events are forwarded as
// [InteractionEvent] values, bridging the scene graph into your
// entity-component system. A [Donburi] adapter ships in the willow/ecs
// submodule.
//
// # Subtree command caching
//
// [Node.SetCacheAsTree] caches all render commands emitted by a container's
// subtree. On subsequent frames the cached commands are replayed with a
// delta transform remap instead of walking the tree — the CPU cost drops
// from O(N) traversal to O(N) memcpy with a matrix multiply per command.
//
// Camera panning, parent movement, and parent alpha changes are handled
// automatically via delta remapping and never invalidate the cache.
//
// Two modes are available:
//
//   - [CacheTreeAuto] (default) — setters on descendant nodes
//     auto-invalidate the cache. Always correct, small per-setter overhead.
//   - [CacheTreeManual] — only [Node.InvalidateCacheTree] triggers a
//     rebuild. Zero setter overhead. Best for large tilemaps where the
//     developer knows exactly when tiles change.
//
// Basic usage:
//
//	tilemap := willow.NewContainer("tilemap")
//	for _, tile := range tiles {
//	    tilemap.AddChild(tile)
//	}
//	tilemap.SetCacheAsTree(true) // auto mode (default)
//
// For maximum performance with large tilemaps:
//
//	tilemap.SetCacheAsTree(true, willow.CacheTreeManual)
//	// ... later, when tiles change:
//	tilemap.InvalidateCacheTree()
//
// # Animated tiles (texture swaps)
//
// Animated tiles (water, lava, torches) that swap UV coordinates on the
// same atlas page work without cache invalidation. The first
// [Node.SetTextureRegion] call with a same-page region automatically
// registers the node as "animated" — subsequent replays read the live
// TextureRegion from the node instead of the cached snapshot:
//
//	// waterTile is a child of a cached tilemap container.
//	// This does NOT invalidate the cache:
//	waterTile.SetTextureRegion(waterFrames[frame])
//
// Changing to a different atlas page does invalidate the cache, since
// page changes affect batch keys. All animation frames for a tile should
// be on the same atlas page (TexturePacker typically does this
// automatically for related frames).
//
// # When to use SetCacheAsTree
//
// The cache stores commands per-container. Any mutation to a child
// invalidates the entire container's cache (in auto mode) or produces
// stale output (in manual mode until you call InvalidateCacheTree).
//
// This makes it ideal for large, mostly-static subtrees:
//
//   - Tilemaps (thousands of tiles, camera scrolling every frame)
//   - Static UI panels (dozens of widgets, occasionally updated)
//   - Background decoration layers
//
// It is NOT useful for containers where children move individually every
// frame — one child's [Node.SetPosition] dirties the whole container,
// so you pay rebuild cost plus cache management overhead.
//
// The recommended pattern is to separate static and dynamic content:
//
//	// Cached: 10K tiles, camera scrolls for free
//	tilemap := willow.NewContainer("tilemap")
//	tilemap.SetCacheAsTree(true, willow.CacheTreeManual)
//
//	// NOT cached: entities move every frame
//	entities := willow.NewContainer("entities")
//
//	scene.Root().AddChild(tilemap)
//	scene.Root().AddChild(entities)
//
// Mesh nodes and particle emitters inside a cached subtree block
// caching — the cache stays dirty and falls back to normal traversal.
// Move particles and meshes to a separate uncached container.
//
// # SetCacheAsTree vs SetCacheAsTexture
//
// Willow has two user-facing caching mechanisms:
//
//   - [Node.SetCacheAsTree] skips CPU traversal. Output is N render
//     commands replayed from cache. Pixel-perfect at any zoom. Animated
//     tiles work for free.
//   - [Node.SetCacheAsTexture] skips GPU draw calls. Output is a single
//     texture. Blurs if zoomed past cached resolution. Best for
//     filtered/masked nodes and complex visual effects.
//
// # Performance
//
// Willow targets zero heap allocations per frame on the hot path. Render
// commands, sort buffers, and vertex arrays are preallocated and reused.
// Dirty flags skip transform recomputation for static subtrees.
// Typed callback slices avoid interface boxing in event dispatch.
//
// # Caching strategy
//
// Willow provides two caching mechanisms that solve different bottlenecks.
// Choosing the right one (or combining both) is the single biggest lever
// for frame-time optimization.
//
// # SetCacheAsTree — skip CPU traversal
//
// [Node.SetCacheAsTree] stores the render commands emitted by a container's
// subtree and replays them on subsequent frames with a matrix multiply per
// command instead of re-walking the tree. This targets CPU time: the
// traversal, transform computation, and command emission for N children
// collapse to a memcpy + delta remap.
//
// When to use it:
//
//   - Large, mostly-static containers (tilemaps, background layers, UI panels)
//   - Scenes where the camera scrolls every frame but the content does not move
//   - Containers with animated tiles that swap UV coordinates on the same atlas page
//
// When NOT to use it:
//
//   - Containers where many children move individually every frame — one child's
//     [Node.SetPosition] invalidates the entire container's cache (auto mode)
//     or produces stale output (manual mode)
//   - Containers that hold mesh nodes or particle emitters — these block caching
//     entirely and fall back to normal traversal
//
// Two modes are available:
//
//   - [CacheTreeAuto] (default) — descendant setters auto-invalidate the cache.
//     Always correct, small per-setter overhead. Good for UI panels that
//     update occasionally.
//   - [CacheTreeManual] — only [Node.InvalidateCacheTree] triggers a rebuild.
//     Zero setter overhead. Best for large tilemaps where you control exactly
//     when tiles change.
//
// Setup example (tilemap with animated water):
//
//	tilemap := willow.NewContainer("tilemap")
//	tilemap.SetCacheAsTree(true, willow.CacheTreeManual)
//	for _, tile := range tiles {
//	    tilemap.AddChild(tile)
//	}
//	scene.Root().AddChild(tilemap)
//
//	// Animate water tiles every N frames — does NOT invalidate the cache
//	// because the UV swap stays on the same atlas page:
//	waterTile.SetTextureRegion(waterFrames[frame])
//
//	// If you add/remove tiles at runtime, invalidate manually:
//	tilemap.InvalidateCacheTree()
//
// Caveats:
//
//   - The cache is per-container. Separate static content (tilemaps) from
//     dynamic content (players, projectiles) into different containers.
//   - Mesh and particle emitter children block caching — move them to an
//     uncached sibling container.
//   - Changing a tile's texture to a different atlas page invalidates the
//     cache because page changes affect batch keys. Keep all animation
//     frames on the same atlas page.
//   - In manual mode, forgetting to call [Node.InvalidateCacheTree] after
//     structural changes (AddChild, RemoveChild, SetVisible) produces stale
//     visual output. Prefer auto mode unless profiling shows the setter
//     overhead matters.
//
// Benchmark results (10K sprites):
//
//	Manual cache, camera scrolling       ~39 µs   (~125x faster than uncached)
//	Manual cache, 100 animated UV swaps  ~1.97 ms (~2.5x faster)
//	Auto cache, 1% children moving       ~4.0 ms  (~1.2x faster)
//	No cache (baseline)                  ~4.9 ms
//
// # SetCacheAsTexture — skip GPU draw calls
//
// [Node.SetCacheAsTexture] renders a node's entire subtree to an offscreen
// image once, then draws that single texture on subsequent frames. This
// targets GPU draw-call count: N children become one textured quad.
//
// When to use it:
//
//   - Nodes with expensive filter chains (blur, outline, palette swap) — the
//     filter is applied once and the result is reused
//   - Complex visual effects that combine masking and filters
//   - Small, visually-rich subtrees where draw-call reduction matters more
//     than pixel-perfect scaling
//
// When NOT to use it:
//
//   - Large scrolling tilemaps — the texture is rasterized at a fixed
//     resolution and will blur if the camera zooms in past that resolution
//   - Subtrees that change frequently — every change requires a full
//     re-render to the offscreen texture
//   - Nodes that need pixel-perfect rendering at varying zoom levels
//
// Setup example (filtered badge):
//
//	badge := willow.NewContainer("badge")
//	badge.AddChild(icon)
//	badge.AddChild(label)
//	badge.Filters = []willow.Filter{
//	    willow.NewOutlineFilter(2, willow.Color{R: 1, G: 1, B: 0, A: 1}),
//	}
//	badge.SetCacheAsTexture(true)
//	scene.Root().AddChild(badge)
//
//	// When the badge content changes, invalidate:
//	badge.InvalidateCache()
//
// Caveats:
//
//   - The offscreen texture is allocated at the subtree's bounding-box size
//     (rounded up to the next power of two). Very large subtrees consume
//     significant GPU memory.
//   - Zooming the camera past the cached resolution produces blurry output.
//     If the camera zooms dynamically, prefer [Node.SetCacheAsTree] instead.
//   - The texture is re-rendered from scratch on invalidation — there is no
//     partial update. Frequent invalidation can be worse than no caching.
//
// # Choosing between the two
//
// Use [Node.SetCacheAsTree] when the bottleneck is CPU traversal of large
// static subtrees (tilemaps, backgrounds). The output is N render commands
// replayed from cache — still pixel-perfect at any zoom, animated tiles
// work for free, and camera movement never invalidates.
//
// Use [Node.SetCacheAsTexture] when the bottleneck is GPU draw calls or
// filter cost. The output is a single textured quad — minimal GPU work,
// but locked to a fixed resolution.
//
// They can be combined: a tilemap container uses SetCacheAsTree for fast
// traversal, while a filtered HUD badge inside it uses SetCacheAsTexture
// to avoid re-applying shaders every frame.
//
// # Recommended scene layout for performance
//
// Organize your scene tree to maximize cache effectiveness:
//
//	scene.Root()
//	├── tilemap      (SetCacheAsTree, manual mode — thousands of tiles)
//	├── entities     (NO cache — players, enemies, projectiles move every frame)
//	├── particles    (NO cache — emitters block tree caching)
//	└── ui           (SetCacheAsTree, auto mode — panels that update occasionally)
//	    └── badge    (SetCacheAsTexture — filtered, rarely changes)
//
// [Ebitengine]: https://ebitengine.org
// [TexturePacker]: https://www.codeandweb.com/texturepacker
// [gween]: https://github.com/tanema/gween
// [Donburi]: https://github.com/yohamta/donburi
package willow
