# Getting Started

<p align="center">
  <img src="gif/shapes.gif" alt="Shapes demo" width="300">
  <img src="gif/masks.gif" alt="Masks demo" width="300">
  <img src="gif/shaders.gif" alt="Shaders demo" width="300">
  <img src="gif/watermesh.gif" alt="Watermesh demo" width="300">
</p>

**[GitHub](https://github.com/phanxgames/willow)** | **[API Reference](https://pkg.go.dev/github.com/phanxgames/willow)**

## Installation

```bash
go get github.com/phanxgames/willow@latest
```

Willow requires **Go 1.24+** and [Ebitengine](https://ebitengine.org/) v2.

## How It Works

Willow uses a retained-mode design pattern. You create a tree of nodes representing your game objects, and Willow handles passing the draw commands down to Ebitengine. There is no need to issue draw commands yourself — just add nodes to the tree, set their properties, and Willow takes care of the rest.

The core loop is:

1. **`scene.Update()`** — processes input, runs callbacks, updates transforms
2. **`scene.Draw(screen)`** — traverses the scene tree, sorts render commands, submits to Ebitengine

There are two ways to run this loop — both are first-class.

## Two Paths: Run Wrapper vs Custom Game Loop

### Path 1: `willow.Run()` — Quick Setup

Call `willow.Run(scene, config)` and Willow handles the window, game loop, and calling `Update`/`Draw` for you. Best for getting started quickly or projects that don't need custom Ebitengine lifecycle control.

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

Use `SetUpdateFunc` to attach game logic without a custom struct:

```go
scene.SetUpdateFunc(func() error {
    sprite.X += 1 // move right every frame
    return nil
})
```

### Path 2: Custom Game Loop — Full Control

Implement `ebiten.Game` yourself and call `scene.Update()` and `scene.Draw(screen)` directly. This gives you full control over the Ebitengine lifecycle — useful when you need custom `Layout` logic, multiple scenes, or integration with other Ebitengine features.

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

## Solid-Color Sprites

Pass an empty `TextureRegion{}` to create shapes without loading textures:

```go
rect := willow.NewSprite("rect", willow.TextureRegion{})
rect.Color = willow.Color{R: 0, G: 0.5, B: 1, A: 1}  // blue
rect.ScaleX = 200  // width in pixels
rect.ScaleY = 50   // height in pixels
```

This uses `willow.WhitePixel` (a shared 1x1 white image) internally. The `Color` field tints it, and `ScaleX`/`ScaleY` size it. Never create unique `*ebiten.Image` instances for solid colors.

## Next Steps

- [Examples](?page=examples) — browse runnable demos with GIFs and source code
- [Architecture](?page=architecture) — understand how Willow works under the hood
- [Nodes](?page=nodes) — learn about node types and tree manipulation
- [Sprites & Atlas](?page=sprites-and-atlas) — load texture atlases
- [Camera & Viewport](?page=camera-and-viewport) — set up viewport and scrolling
