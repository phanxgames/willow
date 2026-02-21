# Nodes

The `Node` is the fundamental building block of Willow's scene graph. Every visible element — sprites, text, particles, meshes — is a `Node` added to the scene tree.

## Node Types

| Type | Constructor | Description |
|------|-------------|-------------|
| `NodeTypeContainer` | `NewContainer(name)` | Invisible grouping node — organizes children, applies transforms |
| `NodeTypeSprite` | `NewSprite(name, region)` | Renders a texture region (or solid color) |
| `NodeTypeMesh` | `NewMesh(name, img, verts, indices)` | Custom vertex geometry |
| `NodeTypeParticleEmitter` | `NewParticleEmitter(name, cfg)` | CPU-simulated particle system |
| `NodeTypeText` | `NewText(name, content, font)` | Text rendered with bitmap or TTF font |

## Creating Nodes

```go
// Container (invisible grouping)
group := willow.NewContainer("enemies")

// Sprite from atlas region
player := willow.NewSprite("player", atlas.Region("player_idle"))

// Solid-color sprite (uses WhitePixel)
rect := willow.NewSprite("bg", willow.TextureRegion{})
rect.Color = willow.Color{R: 0, G: 0.3, B: 0.8, A: 1}
rect.ScaleX = 200
rect.ScaleY = 100

// Text
label := willow.NewText("score", "Score: 0", myFont)

// Particle emitter
emitter := willow.NewParticleEmitter("sparks", emitterConfig)

// Mesh
mesh := willow.NewMesh("grid", img, vertices, indices)
```

## Key Properties

### Visual

| Field | Type | Description |
|-------|------|-------------|
| `Visible` | `bool` | Whether the node (and children) are rendered |
| `Renderable` | `bool` | Whether this specific node renders (children still can) |
| `Alpha` | `float64` | Opacity `[0,1]`, multiplied with parent's alpha |
| `Color` | `Color` | Multiplicative tint; `{1,1,1,1}` = no tint |
| `BlendMode` | `BlendMode` | Blend operation (see below) |
| `RenderLayer` | `uint8` | Primary sort key; lower values draw first |
| `GlobalOrder` | `int` | Secondary sort key within same RenderLayer |

### Blend Modes

| Mode | Description |
|------|-------------|
| `BlendNormal` | Standard alpha blending (source-over) |
| `BlendAdd` | Additive / lighter |
| `BlendMultiply` | Multiply (only darkens) |
| `BlendScreen` | Screen (only brightens) |
| `BlendErase` | Destination-out (punch transparent holes) |
| `BlendMask` | Clip destination to source alpha |
| `BlendBelow` | Destination-over (draw behind) |
| `BlendNone` | Opaque copy (skip blending) |

### Transform

| Field | Type | Description |
|-------|------|-------------|
| `X`, `Y` | `float64` | Local position (Y-down coordinate system) |
| `ScaleX`, `ScaleY` | `float64` | Local scale (1.0 = original size) |
| `Rotation` | `float64` | Radians, clockwise |
| `SkewX`, `SkewY` | `float64` | Shear angles in radians |
| `PivotX`, `PivotY` | `float64` | Transform origin in local pixels |

### Identity

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Human-readable label |
| `ID` | `uint32` | Auto-assigned unique identifier |
| `ZIndex` | `int` | Draw order among siblings; higher = on top |
| `EntityID` | `uint32` | ECS entity bridge |
| `UserData` | `any` | Arbitrary user data |

## Tree Manipulation

```go
// Add children
parent.AddChild(child)
parent.AddChildAt(child, 0)  // insert at index

// Remove children
parent.RemoveChild(child)
removed := parent.RemoveChildAt(2)
child.RemoveFromParent()
parent.RemoveChildren()  // remove all

// Query
children := parent.Children()   // read-only slice
count := parent.NumChildren()
first := parent.ChildAt(0)
```

## ZIndex and Draw Order

Siblings are drawn in order of their `ZIndex` (lower first). Nodes with equal `ZIndex` draw in the order they were added:

```go
background.SetZIndex(-1)  // draws behind
foreground.SetZIndex(10)  // draws on top
```

Use `SetChildIndex(child, index)` to reorder children directly.

## Property Setters

For properties that affect rendering, use the setter methods. These automatically mark transforms dirty and invalidate ancestor caches:

```go
node.SetColor(willow.Color{R: 1, G: 0, B: 0, A: 1})
node.SetBlendMode(willow.BlendAdd)
node.SetVisible(false)
node.SetRenderable(false)
node.SetTextureRegion(newRegion)
node.SetRenderLayer(2)
node.SetGlobalOrder(5)
```

> **Note:** Directly assigning fields like `node.X = 10` is valid but does not call `Invalidate()` automatically. Either use `SetPosition()` or call `node.Invalidate()` after bulk field updates.

## Disposal

```go
node.Dispose()           // removes from parent, releases resources
node.IsDisposed()        // check if disposed
```

Disposing a node also disposes all its children. Disposed nodes cannot be reused.

## Callbacks

Nodes support per-node callbacks for update and interaction:

```go
node.OnUpdate = func(dt float64) {
    // called each frame during scene.Update()
}

node.OnClick = func(ctx willow.ClickContext) {
    fmt.Println("Clicked at", ctx.GlobalX, ctx.GlobalY)
}
```

See [Input, Hit Testing & Gestures](?page=input-hit-testing-and-gestures) for the full callback API.

## Custom Images

For nodes that need a manually-managed image instead of an atlas region:

```go
node.SetCustomImage(myImage)
img := node.CustomImage()
```
