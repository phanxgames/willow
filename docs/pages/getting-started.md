# Getting Started

## Installation

```bash
go get github.com/phanxgames/willow@latest
```

Willow requires **Go 1.24+** and [Ebitengine](https://ebitengine.org/) v2.

## Quick Start

For quick setup, call `willow.Run(scene, config)` and Willow handles the window and game loop. For full control, implement `ebiten.Game` yourself and call `scene.Update()` and `scene.Draw(screen)` directly — both paths are first-class.

```go
package main

import (
    "log"

    "github.com/phanxgames/willow"
)

func main() {
    scene := willow.NewScene()

    sprite := willow.NewSprite("hero", willow.TextureRegion{})
    sprite.ScaleX = 40
    sprite.ScaleY = 40
    sprite.Color = willow.Color{R: 0.3, G: 0.7, B: 1, A: 1}
    sprite.X = 300
    sprite.Y = 220
    scene.Root().AddChild(sprite)

    if err := willow.Run(scene, willow.RunConfig{
        Title:  "My Game",
        Width:  640,
        Height: 480,
    }); err != nil {
        log.Fatal(err)
    }
}
```

This creates a 640x480 window with a blue square. `NewSprite` with an empty `TextureRegion{}` creates a solid-color sprite using `willow.WhitePixel` — set `Color` for the fill and `ScaleX`/`ScaleY` for the size.

## How It Works

Willow uses a retained-mode design pattern. You create a tree of nodes representing your game objects, and Willow handles passing the draw commands down to Ebitengine. There is no need to issue draw commands yourself — just add nodes to the tree, set their properties, and Willow takes care of the rest.

The core loop is:

1. **`scene.Update()`** — processes input, runs callbacks, updates transforms
2. **`scene.Draw(screen)`** — traverses the scene tree, sorts render commands, submits to Ebitengine

When using `willow.Run()`, both are called for you.

## Custom Game Loop

If you need full control over the Ebitengine game loop:

```go
package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/phanxgames/willow"
)

type Game struct {
    scene *willow.Scene
}

func (g *Game) Update() error {
    g.scene.Update()
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.scene.Draw(screen)
}

func (g *Game) Layout(w, h int) (int, int) {
    return w, h
}

func main() {
    scene := willow.NewScene()
    // ... set up your scene ...

    ebiten.SetWindowSize(640, 480)
    ebiten.SetWindowTitle("Custom Loop")
    ebiten.RunGame(&Game{scene: scene})
}
```

## SetUpdateFunc

For simple projects, attach update logic directly to the scene without a custom game struct:

```go
scene := willow.NewScene()

sprite := willow.NewSprite("player", region)
sprite.X = 100
sprite.Y = 100
scene.Root().AddChild(sprite)

scene.SetUpdateFunc(func() error {
    sprite.X += 1 // move right every frame
    return nil
})

willow.Run(scene, willow.RunConfig{
    Title: "Moving Sprite",
    Width: 640,
    Height: 480,
})
```

## Solid-Color Sprites

Pass an empty `TextureRegion{}` to create shapes without loading textures:

```go
rect := willow.NewSprite("rect", willow.TextureRegion{})
rect.Color = willow.Color{R: 0, G: 0.5, B: 1, A: 1}  // blue
rect.ScaleX = 200  // width in pixels
rect.ScaleY = 50   // height in pixels
```

This uses `willow.WhitePixel` (a shared 1x1 white image) internally. The `Color` field tints it, and `ScaleX`/`ScaleY` size it. Never create unique `*ebiten.Image` instances for solid colors.

## Examples

Runnable examples are included in the [examples/](https://github.com/phanxgames/willow/tree/main/examples) directory:

```bash
go run ./examples/basic        # Bouncing colored sprite
go run ./examples/shapes       # Rotating polygon hierarchy
go run ./examples/interaction  # Draggable, clickable rectangles
go run ./examples/text         # Bitmap font text with alignment and wrapping
go run ./examples/texttf       # TTF text with outline
go run ./examples/tweens       # Position, scale, rotation, alpha, and color tweens
go run ./examples/particles    # Fountain, campfire, and sparkler effects
go run ./examples/shaders      # Built-in shader filters showcase
go run ./examples/outline      # Outline and inline filters
go run ./examples/masks        # Star polygon, cursor-following, and erase masking
go run ./examples/lighting     # Dark dungeon with colored torches
go run ./examples/atlas        # TexturePacker atlas loading
go run ./examples/tilemap      # Tile map rendering with camera panning
go run ./examples/rope         # Draggable endpoints connected by a textured rope
go run ./examples/watermesh    # Water surface with per-vertex wave animation
```

## Next Steps

- [Architecture](?page=architecture) — understand how Willow works under the hood
- [Nodes](?page=nodes) — learn about node types and tree manipulation
- [Sprites & Atlas](?page=sprites-and-atlas) — load texture atlases
- [Camera & Viewport](?page=camera-and-viewport) — set up viewport and scrolling
