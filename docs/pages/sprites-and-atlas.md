# Sprites & Atlas

## TextureRegion

A `TextureRegion` describes a rectangular sub-image within an atlas page:

```go
type TextureRegion struct {
    Page      uint16  // atlas page index
    X, Y      uint16  // top-left corner in atlas page
    Width     uint16  // sub-image width (may differ from OriginalW if trimmed)
    Height    uint16  // sub-image height
    OriginalW uint16  // untrimmed width as authored
    OriginalH uint16  // untrimmed height as authored
    OffsetX   int16   // horizontal trim offset
    OffsetY   int16   // vertical trim offset
    Rotated   bool    // true if stored 90 degrees CW in atlas
}
```

An empty `TextureRegion{}` signals "use the 1x1 WhitePixel", which is the standard way to create solid-color shapes.

## Atlas

An `Atlas` holds one or more page images and a name-to-region map:

```go
type Atlas struct {
    Pages []*ebiten.Image
}
```

### Loading an Atlas

Willow supports the [TexturePacker](https://www.codeandweb.com/texturepacker) JSON format (both hash and array variants):

```go
jsonData, _ := os.ReadFile("sprites.json")
pageImg, _, _ := ebitenutil.NewImageFromFile("sprites.png")

atlas, err := willow.LoadAtlas(jsonData, []*ebiten.Image{pageImg})
if err != nil {
    log.Fatal(err)
}
```

Or through a scene (which also registers pages for bitmap font rendering):

```go
atlas, err := scene.LoadAtlas(jsonData, []*ebiten.Image{pageImg})
```

### Getting Regions

```go
region := atlas.Region("player_idle")
sprite := willow.NewSprite("player", region)
```

If a name isn't found, `Region()` returns a magenta placeholder and logs a warning (in debug mode), so missing sprites are immediately visible.

## Solid-Color Sprites

The standard pattern for colored rectangles, backgrounds, and UI elements:

```go
box := willow.NewSprite("box", willow.TextureRegion{})
box.Color = willow.Color{R: 0.2, G: 0.6, B: 1, A: 1}
box.ScaleX = 150  // width
box.ScaleY = 80   // height
```

This uses `willow.WhitePixel` (a shared 1x1 white `*ebiten.Image`). The `Color` field tints it to any color, and `ScaleX`/`ScaleY` determine the size. This approach avoids creating unique images for each color and enables efficient batching.

## TexturePacker Workflow

1. Add your source sprites to TexturePacker
2. Export as **JSON (Hash)** or **JSON (Array)** format
3. Load the JSON and page image(s) at startup
4. Use `atlas.Region("name")` to get regions for sprites

TexturePacker features supported:
- **Trimming** — `OffsetX`/`OffsetY` and `OriginalW`/`OriginalH` handle trimmed whitespace
- **Rotation** — 90-degree CW rotation is handled automatically
- **Multi-page** — pass multiple page images to `LoadAtlas`

## Multi-Page Atlases

```go
atlas, err := willow.LoadAtlas(jsonData, []*ebiten.Image{
    page0Img,
    page1Img,
    page2Img,
})
```

Each `TextureRegion` stores its `Page` index, so sprites from different pages render correctly.

## Registering Pages

For bitmap fonts and tilemaps that reference atlas pages by index:

```go
scene.RegisterPage(0, pageImage)
```
