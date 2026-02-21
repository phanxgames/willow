# Scene

The `Scene` is the top-level manager in Willow. It owns the root node, cameras, input state, and the render pipeline. Every Willow application creates at least one `Scene`.

## Creating a Scene

```go
scene := willow.NewScene()
```

This creates a scene with a root container node, default settings, and no cameras (a default full-screen camera is used automatically if none are added).

## Scene Fields

| Field | Type | Description |
|-------|------|-------------|
| `ClearColor` | `Color` | Background fill color when using `Run()` |
| `ScreenshotDir` | `string` | Output folder for screenshots (default: `"screenshots"`) |

## The Root Node

Every scene has a root container node:

```go
root := scene.Root()
root.AddChild(mySprite)
```

All visible content must be added to the root (or a descendant of the root) to be rendered.

## Update and Draw

```go
func (g *Game) Update() error {
    g.scene.Update()
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.scene.Draw(screen)
}
```

`Update()` processes input, fires callbacks, updates transforms, and simulates particles. `Draw()` traverses the tree, sorts commands, and renders to the screen.

## SetUpdateFunc

For simple projects, attach an update callback instead of implementing a full game struct:

```go
scene.SetUpdateFunc(func() error {
    // game logic here
    return nil
})
```

This function is called once per frame during `scene.Update()`.

## Run Helper

The `Run()` function wraps Ebitengine's `RunGame` with sensible defaults:

```go
willow.Run(scene, willow.RunConfig{
    Title:   "My Game",
    Width:   800,
    Height:  600,
    ShowFPS: true,
})
```

### RunConfig

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Title` | `string` | `""` | Window title |
| `Width` | `int` | `640` | Window width |
| `Height` | `int` | `480` | Window height |
| `ShowFPS` | `bool` | `false` | Show FPS/TPS counter overlay |

When `ShowFPS` is true, an FPS widget is added at `RenderLayer` 255 (always on top).

## Camera Management

```go
cam := scene.NewCamera(willow.Rect{X: 0, Y: 0, Width: 800, Height: 600})
scene.RemoveCamera(cam)
cams := scene.Cameras()
```

See the [Camera & Viewport](?page=camera-and-viewport) page for details.

## Atlas Loading

```go
atlas, err := scene.LoadAtlas(jsonData, []*ebiten.Image{pageImg})
region := atlas.Region("player_idle")
```

You can also register atlas pages independently:

```go
scene.RegisterPage(0, pageImage)
```

See [Sprites & Atlas](?page=sprites-and-atlas) for details.

## Debug and Batch Mode

```go
scene.SetDebugMode(true)   // enable debug overlays
scene.SetBatchMode(willow.BatchModeImmediate)  // switch to per-sprite rendering
```

### BatchMode

| Mode | Description |
|------|-------------|
| `BatchModeCoalesced` | Default. Accumulates vertices and submits one `DrawTriangles32` per batch run. Best performance. |
| `BatchModeImmediate` | One `DrawImage` per sprite. Useful for debugging. |

## ECS Integration

```go
scene.SetEntityStore(store)
```

See [ECS Integration](?page=ecs-integration) for details.
