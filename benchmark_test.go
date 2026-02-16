package willow

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// setupBenchScene creates a Scene with n sprite nodes for benchmark use.
// Each sprite gets a magenta placeholder TextureRegion (no atlas page needed).
func setupBenchScene(n int) *Scene {
	s := NewScene()
	root := s.Root()
	region := TextureRegion{
		Page:      magentaPlaceholderPage,
		Width:     32,
		Height:    32,
		OriginalW: 32,
		OriginalH: 32,
	}
	for i := 0; i < n; i++ {
		sp := NewSprite("sp", region)
		sp.X = float64(i%100) * 40
		sp.Y = float64(i/100) * 40
		root.AddChild(sp)
	}
	return s
}

// --- Sprite Rendering Benchmarks ---

func BenchmarkDraw_10000Sprites_Static(b *testing.B) {
	s := setupBenchScene(10000)
	screen := ebiten.NewImage(1280, 720)

	// Warm up: first draw populates sortBuf etc.
	s.Draw(screen)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Draw(screen)
	}
}

func BenchmarkDraw_10000Sprites_Rotating(b *testing.B) {
	s := setupBenchScene(10000)
	screen := ebiten.NewImage(1280, 720)
	children := s.Root().Children()

	s.Draw(screen) // warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Dirty all transforms by rotating every sprite.
		for _, child := range children {
			child.Rotation += 0.01
			child.transformDirty = true
		}
		s.Draw(screen)
	}
}

func BenchmarkDraw_10000Sprites_AlphaVarying(b *testing.B) {
	s := setupBenchScene(10000)
	screen := ebiten.NewImage(1280, 720)
	children := s.Root().Children()

	s.Draw(screen) // warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j, child := range children {
			child.Alpha = 0.5 + 0.5*math.Sin(float64(i+j)*0.001)
			child.transformDirty = true
		}
		s.Draw(screen)
	}
}

// --- Transform Benchmarks ---

func BenchmarkTransform_10000Dirty(b *testing.B) {
	s := setupBenchScene(10000)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Mark all transforms dirty.
		markSubtreeDirty(s.Root())
		updateWorldTransform(s.Root(), identityTransform, 1.0, false)
	}
}

func BenchmarkTransform_10000Clean(b *testing.B) {
	s := setupBenchScene(10000)
	// Pre-compute so nothing is dirty.
	updateWorldTransform(s.Root(), identityTransform, 1.0, true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		updateWorldTransform(s.Root(), identityTransform, 1.0, false)
	}
}

// --- Command Sort Benchmark ---

func BenchmarkCommandSort_10000(b *testing.B) {
	s := setupBenchScene(10000)
	screen := ebiten.NewImage(1280, 720)

	// Generate commands via a real traversal.
	s.Draw(screen)

	// Save commands for reset.
	saved := make([]RenderCommand, len(s.commands))
	copy(saved, s.commands)

	// Warm up sortBuf to high-water mark.
	s.mergeSort()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Restore unsorted commands.
		s.commands = s.commands[:len(saved)]
		copy(s.commands, saved)
		s.mergeSort()
	}
}

// --- Culling Benchmarks ---

func BenchmarkCulling_On_10000(b *testing.B) {
	s := setupBenchScene(10000)
	cam := s.NewCamera(Rect{X: 0, Y: 0, Width: 640, Height: 480})
	cam.CullEnabled = true
	screen := ebiten.NewImage(640, 480)

	s.Draw(screen) // warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Draw(screen)
	}
}

func BenchmarkCulling_Off_10000(b *testing.B) {
	s := setupBenchScene(10000)
	cam := s.NewCamera(Rect{X: 0, Y: 0, Width: 640, Height: 480})
	cam.CullEnabled = false
	screen := ebiten.NewImage(640, 480)

	s.Draw(screen) // warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Draw(screen)
	}
}

// --- Hit Testing Benchmark ---

func BenchmarkHitTest_1000Interactable(b *testing.B) {
	s := NewScene()
	for i := 0; i < 1000; i++ {
		n := NewSprite("n", TextureRegion{OriginalW: 10, OriginalH: 10})
		n.Interactable = true
		n.X = float64(i%100) * 12
		n.Y = float64(i/100) * 12
		s.Root().AddChild(n)
	}
	updateWorldTransform(s.root, identityTransform, 1.0, true)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.hitTest(500, 50)
	}
}

// --- Particle Benchmarks ---

func BenchmarkParticle_10000Particles(b *testing.B) {
	cfg := EmitterConfig{
		MaxParticles: 10000,
		EmitRate:     100000, // emit fast to fill pool
		Lifetime:     Range{Min: 10, Max: 10},
		Speed:        Range{Min: 10, Max: 50},
		Angle:        Range{Min: 0, Max: 2 * math.Pi},
		StartScale:   Range{Min: 1, Max: 1},
		EndScale:     Range{Min: 0.1, Max: 0.1},
		StartAlpha:   Range{Min: 1, Max: 1},
		EndAlpha:     Range{Min: 0, Max: 0},
		Gravity:      Vec2{X: 0, Y: 100},
		StartColor:   Color{1, 1, 1, 1},
		EndColor:     Color{1, 0, 0, 1},
		Region: TextureRegion{
			Page:      magentaPlaceholderPage,
			Width:     8,
			Height:    8,
			OriginalW: 8,
			OriginalH: 8,
		},
	}
	emitter := newParticleEmitter(cfg)
	emitter.Start()

	// Fill to capacity.
	for emitter.alive < cfg.MaxParticles {
		emitter.update(1.0 / 60.0)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		emitter.update(1.0 / 60.0)
	}
}

func BenchmarkParticle_MaxEmitters(b *testing.B) {
	const numEmitters = 10
	emitters := make([]*ParticleEmitter, numEmitters)
	cfg := EmitterConfig{
		MaxParticles: 1000,
		EmitRate:     100000,
		Lifetime:     Range{Min: 10, Max: 10},
		Speed:        Range{Min: 10, Max: 50},
		Angle:        Range{Min: 0, Max: 2 * math.Pi},
		StartScale:   Range{Min: 1, Max: 1},
		EndScale:     Range{Min: 0.1, Max: 0.1},
		StartAlpha:   Range{Min: 1, Max: 1},
		EndAlpha:     Range{Min: 0, Max: 0},
		Gravity:      Vec2{X: 0, Y: 100},
		StartColor:   Color{1, 1, 1, 1},
		EndColor:     Color{1, 0, 0, 1},
		Region: TextureRegion{
			Page:      magentaPlaceholderPage,
			Width:     8,
			Height:    8,
			OriginalW: 8,
			OriginalH: 8,
		},
	}
	for i := range emitters {
		emitters[i] = newParticleEmitter(cfg)
		emitters[i].Start()
		for emitters[i].alive < cfg.MaxParticles {
			emitters[i].update(1.0 / 60.0)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range emitters {
			e.update(1.0 / 60.0)
		}
	}
}

// --- Filter Benchmarks ---

func BenchmarkFilter_BlurOutline(b *testing.B) {
	s := NewScene()
	root := s.Root()
	region := TextureRegion{
		Page:      magentaPlaceholderPage,
		Width:     32,
		Height:    32,
		OriginalW: 32,
		OriginalH: 32,
	}
	for i := 0; i < 100; i++ {
		sp := NewSprite("sp", region)
		sp.X = float64(i%10) * 40
		sp.Y = float64(i/10) * 40
		sp.Filters = []Filter{NewBlurFilter(4), NewOutlineFilter(2, ColorWhite)}
		root.AddChild(sp)
	}
	screen := ebiten.NewImage(640, 480)
	s.Draw(screen) // warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Draw(screen)
	}
}

// --- Lighting Benchmark ---

func BenchmarkLighting_MultipleLights(b *testing.B) {
	ll := NewLightLayer(512, 512, 0.7)
	defer ll.Dispose()

	for i := 0; i < 50; i++ {
		ll.AddLight(&Light{
			X:         float64(i%10) * 50,
			Y:         float64(i/10) * 50,
			Radius:    30,
			Intensity: 0.8,
			Enabled:   true,
		})
	}
	ll.Redraw() // warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ll.Redraw()
	}
}
