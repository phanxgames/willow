package willow

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- BMFont test fixture ---

// Minimal BMFont .fnt text data with ASCII glyphs for "ABCDEFGHIJ" + space.
const testFntData = `info face="TestFont" size=32 bold=0 italic=0 charset="" unicode=1 stretchH=100 smooth=1 aa=1 padding=0,0,0,0 spacing=0,0
common lineHeight=40 base=30 scaleW=256 scaleH=256 pages=1 packed=0
page id=0 file="test.png"
chars count=11
char id=32  x=0   y=0   width=0   height=0   xoffset=0   yoffset=0   xadvance=10  page=0
char id=65  x=0   y=0   width=20  height=30  xoffset=1   yoffset=2   xadvance=22  page=0
char id=66  x=20  y=0   width=18  height=30  xoffset=1   yoffset=2   xadvance=20  page=0
char id=67  x=38  y=0   width=19  height=30  xoffset=1   yoffset=2   xadvance=21  page=0
char id=68  x=57  y=0   width=20  height=30  xoffset=1   yoffset=2   xadvance=22  page=0
char id=69  x=77  y=0   width=16  height=30  xoffset=1   yoffset=2   xadvance=18  page=0
char id=70  x=93  y=0   width=15  height=30  xoffset=1   yoffset=2   xadvance=17  page=0
char id=71  x=108 y=0   width=20  height=30  xoffset=1   yoffset=2   xadvance=22  page=0
char id=72  x=128 y=0   width=20  height=30  xoffset=1   yoffset=2   xadvance=22  page=0
char id=73  x=148 y=0   width=8   height=30  xoffset=1   yoffset=2   xadvance=10  page=0
char id=74  x=156 y=0   width=12  height=30  xoffset=0   yoffset=2   xadvance=14  page=0
kernings count=2
kerning first=65 second=66 amount=-2
kerning first=65 second=67 amount=-1
`

// testFntDataNoLineHeight is malformed .fnt data missing lineHeight.
const testFntDataNoLineHeight = `info face="Bad" size=32
page id=0 file="test.png"
chars count=1
char id=65 x=0 y=0 width=10 height=10 xoffset=0 yoffset=0 xadvance=12 page=0
`

// testFntDataNoChars is .fnt data with no char definitions.
const testFntDataNoChars = `info face="Bad" size=32
common lineHeight=40 base=30 scaleW=256 scaleH=256 pages=1 packed=0
page id=0 file="test.png"
`

func loadTestFont(t *testing.T) *BitmapFont {
	t.Helper()
	f, err := LoadBitmapFont([]byte(testFntData))
	if err != nil {
		t.Fatalf("LoadBitmapFont: %v", err)
	}
	return f
}

// --- LoadBitmapFont tests ---

func TestLoadBitmapFont_GlyphCount(t *testing.T) {
	f := loadTestFont(t)

	// Count populated ASCII glyphs
	count := 0
	for i := range f.asciiSet {
		if f.asciiSet[i] {
			count++
		}
	}
	if count != 11 {
		t.Errorf("glyph count = %d, want 11", count)
	}
}

func TestLoadBitmapFont_LineHeight(t *testing.T) {
	f := loadTestFont(t)
	if f.lineHeight != 40 {
		t.Errorf("lineHeight = %f, want 40", f.lineHeight)
	}
}

func TestLoadBitmapFont_InvalidData(t *testing.T) {
	_, err := LoadBitmapFont([]byte("not valid fnt data at all"))
	if err == nil {
		t.Error("expected error for invalid data, got nil")
	}
}

func TestLoadBitmapFont_MissingLineHeight(t *testing.T) {
	_, err := LoadBitmapFont([]byte(testFntDataNoLineHeight))
	if err == nil {
		t.Error("expected error for missing lineHeight, got nil")
	}
}

func TestLoadBitmapFont_NoChars(t *testing.T) {
	_, err := LoadBitmapFont([]byte(testFntDataNoChars))
	if err == nil {
		t.Error("expected error for no char definitions, got nil")
	}
}

// --- LoadTTFFont ---

func TestLoadTTFFont_InvalidData(t *testing.T) {
	_, err := LoadTTFFont([]byte("not a TTF file"), 16)
	if err == nil {
		t.Error("expected error for invalid TTF data, got nil")
	}
}

// --- Font interface: MeasureString ---

func TestBitmapFont_MeasureString_SingleLine(t *testing.T) {
	f := loadTestFont(t)
	// "AB" = A(xadvance=22) + kern(A,B)=-2 + B(xadvance=20) = 40
	w, h := f.MeasureString("AB")
	if w != 40 {
		t.Errorf("MeasureString(\"AB\") width = %f, want 40", w)
	}
	if h != 40 {
		t.Errorf("MeasureString(\"AB\") height = %f, want 40", h)
	}
}

func TestBitmapFont_MeasureString_MultiLine(t *testing.T) {
	f := loadTestFont(t)
	// "A\nB" = two lines
	w, h := f.MeasureString("A\nB")
	if w != 22 { // max of A(22) and B(20)
		t.Errorf("width = %f, want 22", w)
	}
	if h != 80 { // 2 lines * 40 lineHeight
		t.Errorf("height = %f, want 80", h)
	}
}

func TestBitmapFont_MeasureString_Empty(t *testing.T) {
	f := loadTestFont(t)
	w, h := f.MeasureString("")
	if w != 0 || h != 40 { // 1 line * lineHeight even for empty
		t.Errorf("MeasureString(\"\") = (%f, %f), want (0, 40)", w, h)
	}
}

// --- Font interface: LineHeight ---

func TestBitmapFont_LineHeight(t *testing.T) {
	f := loadTestFont(t)
	var font Font = f
	if font.LineHeight() != 40 {
		t.Errorf("LineHeight() = %f, want 40", font.LineHeight())
	}
}

// --- Kerning ---

func TestBitmapFont_Kerning(t *testing.T) {
	f := loadTestFont(t)

	// "AB" with kerning -2 vs "AC" with kerning -1
	wAB, _ := f.MeasureString("AB")
	wAC, _ := f.MeasureString("AC")

	// AB = 22 + (-2) + 20 = 40
	// AC = 22 + (-1) + 21 = 42
	if wAB != 40 {
		t.Errorf("AB width = %f, want 40", wAB)
	}
	if wAC != 42 {
		t.Errorf("AC width = %f, want 42", wAC)
	}
}

func TestBitmapFont_NoKerning(t *testing.T) {
	f := loadTestFont(t)

	// "CD" has no kerning entry
	wCD, _ := f.MeasureString("CD")
	// C(21) + D(22) = 43
	if wCD != 43 {
		t.Errorf("CD width = %f, want 43", wCD)
	}
}

// --- Word wrapping ---

func TestTextBlock_WordWrap(t *testing.T) {
	f := loadTestFont(t)

	tb := &TextBlock{
		Content:     "AB CD",
		Font:        f,
		WrapWidth:   45, // AB(40) fits, "AB CD" doesn't
		Color:       Color{1, 1, 1, 1},
		layoutDirty: true,
	}

	lines := tb.layout()
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(lines))
	}
}

func TestTextBlock_NoWrap_WhenZeroWrapWidth(t *testing.T) {
	f := loadTestFont(t)

	tb := &TextBlock{
		Content:     "ABCDEFGHIJ",
		Font:        f,
		WrapWidth:   0, // no wrapping
		Color:       Color{1, 1, 1, 1},
		layoutDirty: true,
	}

	lines := tb.layout()
	if len(lines) != 1 {
		t.Errorf("line count = %d, want 1 (no wrapping)", len(lines))
	}
}

// --- Text alignment ---

func TestTextBlock_AlignCenter(t *testing.T) {
	f := loadTestFont(t)

	tb := &TextBlock{
		Content:     "A\nAB",
		Font:        f,
		Align:       TextAlignCenter,
		Color:       Color{1, 1, 1, 1},
		layoutDirty: true,
	}

	lines := tb.layout()
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(lines))
	}

	// Line 1 "A" (width 22) centered in measured width (40 from "AB")
	// Offset = (40 - 22) / 2 = 9
	if len(lines[0].glyphs) > 0 {
		gx := lines[0].glyphs[0].x
		expected := float64(1) + 9.0 // xOffset(1) + center offset(9)
		if math.Abs(gx-expected) > 0.01 {
			t.Errorf("center-aligned glyph x = %f, want %f", gx, expected)
		}
	}
}

func TestTextBlock_AlignRight(t *testing.T) {
	f := loadTestFont(t)

	tb := &TextBlock{
		Content:     "A\nAB",
		Font:        f,
		Align:       TextAlignRight,
		Color:       Color{1, 1, 1, 1},
		layoutDirty: true,
	}

	lines := tb.layout()
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(lines))
	}

	// Line 1 "A" (width 22) right-aligned in measured width (40 from "AB")
	// Offset = 40 - 22 = 18
	if len(lines[0].glyphs) > 0 {
		gx := lines[0].glyphs[0].x
		expected := float64(1) + 18.0 // xOffset(1) + right offset(18)
		if math.Abs(gx-expected) > 0.01 {
			t.Errorf("right-aligned glyph x = %f, want %f", gx, expected)
		}
	}
}

// --- TextBlock color tint ---

func TestTextBlock_ColorTint(t *testing.T) {
	f := loadTestFont(t)

	tb := &TextBlock{
		Content:     "A",
		Font:        f,
		Color:       Color{0.5, 0.8, 1.0, 0.9},
		layoutDirty: true,
	}

	// Verify the color is preserved on the TextBlock
	if tb.Color.R != 0.5 || tb.Color.G != 0.8 || tb.Color.B != 1.0 || tb.Color.A != 0.9 {
		t.Errorf("TextBlock color = %+v, want {0.5 0.8 1.0 0.9}", tb.Color)
	}
}

// --- Layout caching ---

func TestTextBlock_LayoutCaching(t *testing.T) {
	f := loadTestFont(t)

	tb := &TextBlock{
		Content:     "ABC",
		Font:        f,
		Color:       Color{1, 1, 1, 1},
		layoutDirty: true,
	}

	// First layout
	lines1 := tb.layout()
	w1, h1 := tb.measuredW, tb.measuredH

	// Second layout without changes — should not recompute
	lines2 := tb.layout()
	w2, h2 := tb.measuredW, tb.measuredH

	if len(lines1) != len(lines2) {
		t.Error("cached layout returned different line count")
	}
	if w1 != w2 || h1 != h2 {
		t.Error("cached layout returned different dimensions")
	}

	// layoutDirty should be false
	if tb.layoutDirty {
		t.Error("layoutDirty should be false after layout()")
	}

	// Now change content and mark dirty
	tb.Content = "AB"
	tb.layoutDirty = true
	lines3 := tb.layout()
	if tb.measuredW == w1 {
		t.Error("layout should recompute after content change")
	}
	_ = lines3
}

// --- Text node in scene graph ---

func TestTextNode_InSceneGraph_EmitsCommands(t *testing.T) {
	f := loadTestFont(t)
	atlas := ebiten.NewImage(256, 256)

	s := NewScene()
	s.RegisterPage(0, atlas)

	textNode := NewText("greeting", "AB", f)
	s.Root().AddChild(textNode)

	traverseScene(s)

	// "AB" = 2 glyphs → 2 CommandSprite commands
	if len(s.commands) != 2 {
		t.Fatalf("commands = %d, want 2", len(s.commands))
	}
	for i, cmd := range s.commands {
		if cmd.Type != CommandSprite {
			t.Errorf("commands[%d].Type = %d, want CommandSprite", i, cmd.Type)
		}
	}
}

func TestTextNode_InheritsTransform(t *testing.T) {
	f := loadTestFont(t)
	atlas := ebiten.NewImage(256, 256)

	s := NewScene()
	s.RegisterPage(0, atlas)

	parent := NewContainer("parent")
	parent.X = 100
	parent.Y = 50

	textNode := NewText("t", "A", f)
	parent.AddChild(textNode)
	s.Root().AddChild(parent)

	traverseScene(s)

	if len(s.commands) != 1 {
		t.Fatalf("commands = %d, want 1", len(s.commands))
	}

	// The glyph transform should include the parent offset.
	// Glyph 'A' has xOffset=1, yOffset=2.
	// World = parent(100,50) + glyph(1,2) = (101, 52) in tx/ty
	cmd := s.commands[0]
	tx, ty := float64(cmd.Transform[4]), float64(cmd.Transform[5])
	if math.Abs(tx-101) > 0.01 {
		t.Errorf("glyph tx = %f, want 101", tx)
	}
	if math.Abs(ty-52) > 0.01 {
		t.Errorf("glyph ty = %f, want 52", ty)
	}
}

func TestTextNode_InheritsAlpha(t *testing.T) {
	f := loadTestFont(t)
	atlas := ebiten.NewImage(256, 256)

	s := NewScene()
	s.RegisterPage(0, atlas)

	parent := NewContainer("parent")
	parent.Alpha = 0.5

	textNode := NewText("t", "A", f)
	textNode.Alpha = 0.8
	parent.AddChild(textNode)
	s.Root().AddChild(parent)

	traverseScene(s)

	if len(s.commands) != 1 {
		t.Fatalf("commands = %d, want 1", len(s.commands))
	}

	// worldAlpha = 0.5 * 0.8 = 0.4
	// Color.A = textBlock.Color.A(1) * node.Color.A(1) * worldAlpha(0.4) = 0.4
	if got := float64(s.commands[0].Color.A); math.Abs(got-0.4) > 1e-6 {
		t.Errorf("cmd.Color.A = %v, want ~0.4", got)
	}
}

// --- Outline ---

func TestTextNode_Outline_EmitsExtraCommands(t *testing.T) {
	f := loadTestFont(t)
	atlas := ebiten.NewImage(256, 256)

	s := NewScene()
	s.RegisterPage(0, atlas)

	textNode := NewText("t", "A", f)
	textNode.TextBlock.Outline = &Outline{
		Color:     Color{0, 0, 0, 1},
		Thickness: 2,
	}
	s.Root().AddChild(textNode)

	traverseScene(s)

	// 1 glyph * 8 outline directions + 1 fill = 9 commands
	if len(s.commands) != 9 {
		t.Fatalf("commands with outline = %d, want 9", len(s.commands))
	}
}

// --- NewText constructor ---

func TestNewText_SetsTextBlock(t *testing.T) {
	f := loadTestFont(t)
	n := NewText("label", "Hello", f)

	if n.Type != NodeTypeText {
		t.Errorf("Type = %d, want NodeTypeText", n.Type)
	}
	if n.TextBlock == nil {
		t.Fatal("TextBlock is nil")
	}
	if n.TextBlock.Content != "Hello" {
		t.Errorf("Content = %q, want \"Hello\"", n.TextBlock.Content)
	}
	if n.TextBlock.Font != f {
		t.Error("Font not set correctly")
	}
	if n.TextBlock.Color != (Color{1, 1, 1, 1}) {
		t.Errorf("TextBlock.Color = %+v, want white", n.TextBlock.Color)
	}
	if !n.TextBlock.layoutDirty {
		t.Error("layoutDirty should be true initially")
	}
}

// --- composeGlyphTransform ---

func TestComposeGlyphTransform_Identity(t *testing.T) {
	world := identityTransform
	result := composeGlyphTransform(world, 10, 20)

	if result[4] != 10 || result[5] != 20 {
		t.Errorf("translate = (%f, %f), want (10, 20)", result[4], result[5])
	}
	if result[0] != 1 || result[3] != 1 {
		t.Error("scale should remain identity")
	}
}

func TestComposeGlyphTransform_Scaled(t *testing.T) {
	world := [6]float64{2, 0, 0, 2, 50, 100} // 2x scale, translate(50,100)
	result := composeGlyphTransform(world, 10, 20)

	// tx = 2*10 + 0*20 + 50 = 70
	// ty = 0*10 + 2*20 + 100 = 140
	if math.Abs(result[4]-70) > 0.01 {
		t.Errorf("tx = %f, want 70", result[4])
	}
	if math.Abs(result[5]-140) > 0.01 {
		t.Errorf("ty = %f, want 140", result[5])
	}
}

// --- Text node culling ---

func TestTextNode_CullDimensions(t *testing.T) {
	f := loadTestFont(t)
	n := NewText("t", "AB", f)

	// Ensure layout is computed
	n.TextBlock.layout()

	w, h := nodeDimensions(n)
	if w == 0 || h == 0 {
		t.Errorf("text node dimensions = (%f, %f), want non-zero", w, h)
	}
}

// --- Benchmarks ---

func BenchmarkBitmapFont_MeasureString(b *testing.B) {
	f, _ := LoadBitmapFont([]byte(testFntData))
	s := "ABCDEFGHIJ"
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		f.MeasureString(s)
	}
}

func BenchmarkTextBlock_Layout(b *testing.B) {
	f, _ := LoadBitmapFont([]byte(testFntData))

	tb := &TextBlock{
		Content:     "ABCDEFGHIJ",
		Font:        f,
		Color:       Color{1, 1, 1, 1},
		layoutDirty: true,
	}
	// Warm up
	tb.layout()

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		// Cached layout (no recompute)
		tb.layout()
	}
}

func BenchmarkTextBlock_LayoutDirty(b *testing.B) {
	f, _ := LoadBitmapFont([]byte(testFntData))

	tb := &TextBlock{
		Content:     "ABCDEFGHIJ",
		Font:        f,
		Color:       Color{1, 1, 1, 1},
		layoutDirty: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		tb.layoutDirty = true
		tb.layout()
	}
}
