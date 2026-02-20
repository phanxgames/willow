package willow

import (
	"math"
	"sort"
	"testing"
)

// helper to build a scene and run traverse without Draw (no ebiten.Image needed)
func traverseScene(s *Scene) {
	s.commands = s.commands[:0]
	// Compute world transforms first (mirrors Scene.Update), then traverse
	// read-only with identity view.
	updateWorldTransform(s.root, identityTransform, 1.0, false, false)
	s.viewTransform = identityTransform
	treeOrder := 0
	s.traverse(s.root, &treeOrder)
}

// --- Command emission ---

func TestSingleSpriteEmitsOneCommand(t *testing.T) {
	s := NewScene()
	sprite := NewSprite("s", TextureRegion{Width: 32, Height: 32})
	s.Root().AddChild(sprite)

	traverseScene(s)

	if len(s.commands) != 1 {
		t.Fatalf("commands = %d, want 1", len(s.commands))
	}
	if s.commands[0].Type != CommandSprite {
		t.Errorf("Type = %d, want CommandSprite", s.commands[0].Type)
	}
}

func TestInvisibleNodeNoCommands(t *testing.T) {
	s := NewScene()
	sprite := NewSprite("s", TextureRegion{Width: 32, Height: 32})
	sprite.Visible = false
	s.Root().AddChild(sprite)

	traverseScene(s)

	if len(s.commands) != 0 {
		t.Errorf("commands = %d, want 0 for invisible node", len(s.commands))
	}
}

func TestInvisibleSubtreeSkipped(t *testing.T) {
	s := NewScene()
	parent := NewContainer("parent")
	parent.Visible = false
	child := NewSprite("child", TextureRegion{Width: 32, Height: 32})
	parent.AddChild(child)
	s.Root().AddChild(parent)

	traverseScene(s)

	if len(s.commands) != 0 {
		t.Errorf("commands = %d, want 0 for invisible subtree", len(s.commands))
	}
}

func TestNonRenderableNodeSkipped(t *testing.T) {
	s := NewScene()
	parent := NewSprite("parent", TextureRegion{Width: 32, Height: 32})
	parent.Renderable = false
	child := NewSprite("child", TextureRegion{Width: 16, Height: 16})
	parent.AddChild(child)
	s.Root().AddChild(parent)

	traverseScene(s)

	// Parent not renderable → 0 commands from parent
	// But child is renderable → 1 command
	if len(s.commands) != 1 {
		t.Fatalf("commands = %d, want 1", len(s.commands))
	}
	if s.commands[0].TextureRegion.Width != 16 {
		t.Error("command should be from child, not parent")
	}
}

func TestContainerNoCommand(t *testing.T) {
	s := NewScene()
	s.Root().AddChild(NewContainer("empty"))

	traverseScene(s)

	if len(s.commands) != 0 {
		t.Errorf("containers should not emit commands, got %d", len(s.commands))
	}
}

func TestTreeOrderAssignment(t *testing.T) {
	s := NewScene()
	a := NewSprite("a", TextureRegion{Width: 1})
	b := NewSprite("b", TextureRegion{Width: 2})
	c := NewSprite("c", TextureRegion{Width: 3})
	s.Root().AddChild(a)
	s.Root().AddChild(b)
	s.Root().AddChild(c)

	traverseScene(s)

	if len(s.commands) != 3 {
		t.Fatalf("commands = %d, want 3", len(s.commands))
	}
	for i := 1; i < len(s.commands); i++ {
		if s.commands[i].treeOrder <= s.commands[i-1].treeOrder {
			t.Errorf("treeOrder not strictly increasing: [%d]=%d, [%d]=%d",
				i-1, s.commands[i-1].treeOrder, i, s.commands[i].treeOrder)
		}
	}
}

func TestWorldAlphaInCommand(t *testing.T) {
	s := NewScene()
	parent := NewContainer("parent")
	parent.Alpha = 0.5
	child := NewSprite("child", TextureRegion{Width: 32, Height: 32})
	child.Alpha = 0.8
	parent.AddChild(child)
	s.Root().AddChild(parent)

	traverseScene(s)

	if len(s.commands) != 1 {
		t.Fatalf("commands = %d, want 1", len(s.commands))
	}
	// worldAlpha = 0.5 * 0.8 = 0.4, Color.A = 1.0 * 0.4 = 0.4
	if got := float64(s.commands[0].Color.A); math.Abs(got-0.4) > 1e-6 {
		t.Errorf("cmd.Color.A = %v, want ~0.4", got)
	}
}

// --- Sorting ---

func TestRenderLayerSorting(t *testing.T) {
	s := NewScene()
	a := NewSprite("a", TextureRegion{Width: 1})
	a.RenderLayer = 1
	b := NewSprite("b", TextureRegion{Width: 2})
	b.RenderLayer = 0
	s.Root().AddChild(a)
	s.Root().AddChild(b)

	traverseScene(s)
	s.mergeSort()

	if s.commands[0].RenderLayer != 0 {
		t.Errorf("first command should be layer 0, got %d", s.commands[0].RenderLayer)
	}
	if s.commands[1].RenderLayer != 1 {
		t.Errorf("second command should be layer 1, got %d", s.commands[1].RenderLayer)
	}
}

func TestGlobalOrderSorting(t *testing.T) {
	s := NewScene()
	a := NewSprite("a", TextureRegion{Width: 1})
	a.GlobalOrder = 10
	b := NewSprite("b", TextureRegion{Width: 2})
	b.GlobalOrder = 5
	s.Root().AddChild(a)
	s.Root().AddChild(b)

	traverseScene(s)
	s.mergeSort()

	if s.commands[0].GlobalOrder != 5 {
		t.Errorf("first command GlobalOrder = %d, want 5", s.commands[0].GlobalOrder)
	}
}

func TestTreeOrderPreservedWithinLayer(t *testing.T) {
	s := NewScene()
	// All same layer, GlobalOrder = 0 → tree order should be preserved
	for i := 0; i < 5; i++ {
		sp := NewSprite("", TextureRegion{Width: uint16(i + 1)})
		s.Root().AddChild(sp)
	}

	traverseScene(s)
	s.mergeSort()

	for i := 0; i < 5; i++ {
		if s.commands[i].TextureRegion.Width != uint16(i+1) {
			t.Errorf("commands[%d].Width = %d, want %d", i, s.commands[i].TextureRegion.Width, i+1)
		}
	}
}

// --- ZIndex ---

func TestZIndexSorting(t *testing.T) {
	s := NewScene()
	a := NewSprite("a", TextureRegion{Width: 1})
	b := NewSprite("b", TextureRegion{Width: 2})
	c := NewSprite("c", TextureRegion{Width: 3})
	a.SetZIndex(2)
	b.SetZIndex(0)
	c.SetZIndex(1)
	s.Root().AddChild(a)
	s.Root().AddChild(b)
	s.Root().AddChild(c)

	traverseScene(s)

	// b(z=0) should be first, then c(z=1), then a(z=2)
	if len(s.commands) != 3 {
		t.Fatalf("commands = %d, want 3", len(s.commands))
	}
	if s.commands[0].TextureRegion.Width != 2 {
		t.Errorf("first should be b (width=2), got width=%d", s.commands[0].TextureRegion.Width)
	}
	if s.commands[1].TextureRegion.Width != 3 {
		t.Errorf("second should be c (width=3), got width=%d", s.commands[1].TextureRegion.Width)
	}
	if s.commands[2].TextureRegion.Width != 1 {
		t.Errorf("third should be a (width=1), got width=%d", s.commands[2].TextureRegion.Width)
	}
}

// --- Merge sort ---

func TestMergeSortMatchesStdlib(t *testing.T) {
	s := NewScene()
	// Create commands with varied layers and orders
	cmds := []RenderCommand{
		{RenderLayer: 2, GlobalOrder: 0, treeOrder: 1},
		{RenderLayer: 0, GlobalOrder: 3, treeOrder: 2},
		{RenderLayer: 0, GlobalOrder: 1, treeOrder: 3},
		{RenderLayer: 1, GlobalOrder: 0, treeOrder: 4},
		{RenderLayer: 0, GlobalOrder: 1, treeOrder: 5},
		{RenderLayer: 2, GlobalOrder: 0, treeOrder: 6},
		{RenderLayer: 0, GlobalOrder: 0, treeOrder: 7},
	}

	// Reference: stdlib stable sort
	ref := make([]RenderCommand, len(cmds))
	copy(ref, cmds)
	sort.SliceStable(ref, func(i, j int) bool {
		a, b := ref[i], ref[j]
		if a.RenderLayer != b.RenderLayer {
			return a.RenderLayer < b.RenderLayer
		}
		if a.GlobalOrder != b.GlobalOrder {
			return a.GlobalOrder < b.GlobalOrder
		}
		return a.treeOrder < b.treeOrder
	})

	// Our merge sort
	s.commands = make([]RenderCommand, len(cmds))
	copy(s.commands, cmds)
	s.mergeSort()

	for i := range s.commands {
		a, b := s.commands[i], ref[i]
		if a.RenderLayer != b.RenderLayer || a.GlobalOrder != b.GlobalOrder || a.treeOrder != b.treeOrder {
			t.Errorf("index %d: mergeSort=(%d,%d,%d), stdlib=(%d,%d,%d)",
				i, a.RenderLayer, a.GlobalOrder, a.treeOrder,
				b.RenderLayer, b.GlobalOrder, b.treeOrder)
		}
	}
}

func TestMergeSortStable(t *testing.T) {
	s := NewScene()
	// All same layer and GlobalOrder — treeOrder should be preserved
	s.commands = make([]RenderCommand, 100)
	for i := range s.commands {
		s.commands[i] = RenderCommand{treeOrder: i}
	}

	s.mergeSort()

	for i := range s.commands {
		if s.commands[i].treeOrder != i {
			t.Fatalf("stability broken at index %d: treeOrder=%d", i, s.commands[i].treeOrder)
		}
	}
}

func TestMergeSortBufferReuse(t *testing.T) {
	s := NewScene()

	// First sort: allocates buffer
	s.commands = make([]RenderCommand, 50)
	for i := range s.commands {
		s.commands[i] = RenderCommand{treeOrder: 50 - i}
	}
	s.mergeSort()
	bufCap := cap(s.sortBuf)

	// Second sort with smaller input: should not reallocate
	s.commands = make([]RenderCommand, 30)
	for i := range s.commands {
		s.commands[i] = RenderCommand{treeOrder: 30 - i}
	}
	s.mergeSort()

	if cap(s.sortBuf) != bufCap {
		t.Errorf("sortBuf reallocated: was %d, now %d", bufCap, cap(s.sortBuf))
	}
}

func TestMergeSortEmpty(t *testing.T) {
	s := NewScene()
	s.commands = nil
	s.mergeSort() // should not panic
}

func TestMergeSortSingleElement(t *testing.T) {
	s := NewScene()
	s.commands = []RenderCommand{{treeOrder: 1}}
	s.mergeSort() // should not panic
	if s.commands[0].treeOrder != 1 {
		t.Error("single element should remain unchanged")
	}
}

// --- Benchmarks ---

func buildSpriteScene(count int) *Scene {
	s := NewScene()
	for i := 0; i < count; i++ {
		sp := NewSprite("", TextureRegion{Width: 32, Height: 32})
		s.Root().AddChild(sp)
	}
	return s
}

func BenchmarkTraverse1000(b *testing.B) {
	s := buildSpriteScene(1000)
	// Warm up transforms
	traverseScene(s)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		s.commands = s.commands[:0]
		treeOrder := 0
		s.traverse(s.root, &treeOrder)
	}
}

func BenchmarkTraverse10000(b *testing.B) {
	s := buildSpriteScene(10000)
	traverseScene(s)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		s.commands = s.commands[:0]
		treeOrder := 0
		s.traverse(s.root, &treeOrder)
	}
}

func BenchmarkCommandSort10000(b *testing.B) {
	s := NewScene()
	s.commands = make([]RenderCommand, 10000)
	for i := range s.commands {
		s.commands[i] = RenderCommand{
			RenderLayer: uint8(i % 4),
			GlobalOrder: i % 10,
			treeOrder:   i,
		}
	}
	// Warmup to allocate sortBuf
	s.mergeSort()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Reverse to force work
		for i, j := 0, len(s.commands)-1; i < j; i, j = i+1, j-1 {
			s.commands[i], s.commands[j] = s.commands[j], s.commands[i]
		}
		s.mergeSort()
	}
}
