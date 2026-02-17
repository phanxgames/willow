// Package willow is a retained-mode 2D scene graph, interaction layer, and
// render compiler for [Ebitengine].
//
// Willow provides the scene graph, transform hierarchy, sprite batching, input
// handling, camera viewports, text rendering, and particle systems that every
// non-trivial 2D project needs — packaged as a single, focused library.
//
// # Quick start
//
//	scene := willow.NewScene()
//
//	sprite := willow.NewSprite("hero", atlas.Region("hero_idle"))
//	sprite.X, sprite.Y = 400, 300
//	scene.Root().AddChild(sprite)
//
//	// In your ebiten.Game:
//	func (g *Game) Update() error { g.scene.Update(); return nil }
//	func (g *Game) Draw(screen *ebiten.Image) { g.scene.Draw(screen) }
//
// # Key features
//
//   - Scene graph with parent/child transform inheritance (position, rotation, scale, skew, pivot)
//   - TexturePacker JSON atlas loading with sprite batching
//   - Camera system with follow, scroll-to, frustum culling, and world/screen conversion
//   - Hierarchical hit testing with pointer capture, drag, and two-finger pinch
//   - Bitmap font and TTF text rendering with alignment and word wrap
//   - CPU-simulated particle system with preallocated pools
//   - Mesh support via DrawTriangles with rope, polygon, and grid helpers
//   - Composable Kage shader filter chains with render-target masking
//   - Animation and tweening via gween with 45+ easing functions
//   - Optional ECS integration via the willow/ecs submodule
//
// # Performance
//
// Willow targets zero heap allocations per frame on the hot path. Render
// commands, sort buffers, and vertex arrays are preallocated and reused.
// Dirty flags skip transform recomputation for static subtrees.
//
// See the project README for full documentation: https://github.com/phanxgames/willow
//
// [Ebitengine]: https://ebitengine.org
package willow
