# Text & Fonts

Willow supports two font backends: **BitmapFont** (BMFont format) and **TTFFont** (TrueType via Ebitengine's text/v2). Both implement the `Font` interface.

## Font Interface

```go
type Font interface {
    MeasureString(text string) (width, height float64)
    LineHeight() float64
}
```

## BitmapFont

Load a BMFont `.fnt` file (text format):

```go
fntData, _ := os.ReadFile("myfont.fnt")
font, err := willow.LoadBitmapFont(fntData)
```

The font's atlas page defaults to page 0. Register the page image with the scene:

```go
scene.RegisterPage(0, fontPageImage)
```

For multi-page fonts or when sharing page indices with other atlases:

```go
font, err := willow.LoadBitmapFontPage(fntData, 3)  // use page index 3
scene.RegisterPage(3, fontPageImage)
```

BitmapFont glyphs are rendered as individual sprites, fully batched with other sprites on the same atlas page. Supports ASCII fast-path (fixed array lookup) and Unicode extension (map lookup).

## TTFFont

Load a TrueType font:

```go
ttfData, _ := os.ReadFile("Roboto-Regular.ttf")
font, err := willow.LoadTTFFont(ttfData, 24)  // 24px size
```

TTFFont renders to a cached image (re-rendered only when the text or layout changes). For direct access to the Ebitengine text face:

```go
face := font.Face()  // *text.GoTextFace
```

## Creating Text Nodes

```go
label := willow.NewText("label", "Hello, Willow!", font)
label.X = 100
label.Y = 50
scene.Root().AddChild(label)
```

## TextBlock

The `TextBlock` struct controls text layout and appearance:

```go
type TextBlock struct {
    Content    string
    Font       Font
    Align      TextAlign    // Left, Center, Right
    WrapWidth  float64      // 0 = no wrapping
    Color      Color
    Outline    *Outline     // nil = no outline
    LineHeight float64      // 0 = use Font.LineHeight()
}
```

Access the text block via the node:

```go
node := willow.NewText("msg", "Initial text", font)
node.TextBlock.Align = willow.TextAlignCenter
node.TextBlock.WrapWidth = 300
node.TextBlock.Color = willow.Color{R: 1, G: 1, B: 0, A: 1}
```

### Updating Text

```go
node.TextBlock.Content = "New text content"
node.TextBlock.Invalidate()
```

Always call `Invalidate()` after changing `TextBlock` properties to trigger re-layout.

## Text Alignment

```go
willow.TextAlignLeft    // default
willow.TextAlignCenter
willow.TextAlignRight
```

Alignment is relative to the node's X position (left-aligned) or within the `WrapWidth` (center/right-aligned).

## Word Wrapping

Set `WrapWidth` to enable automatic line breaking:

```go
node.TextBlock.WrapWidth = 400  // wrap at 400 pixels
node.TextBlock.Invalidate()
```

## Text Outline

Add an outline effect to text:

```go
node.TextBlock.Outline = &willow.Outline{
    Color:     willow.Color{R: 0, G: 0, B: 0, A: 1},
    Thickness: 2,
}
node.TextBlock.Invalidate()
```

## Measuring Text

```go
width, height := font.MeasureString("Hello!")
lineH := font.LineHeight()
```

## Next Steps

- [Tilemap Viewport](?page=tilemap-viewport) — tile-based map rendering
- [Polygons](?page=polygons) — point-list polygon shapes

## Related

- [Sprites & Atlas](?page=sprites-and-atlas) — atlas loading and page registration for bitmap fonts
- [Nodes](?page=nodes) — node types and visual properties
