# LLM Quick-Reference

Common pitfalls and patterns that trip up AI code generators. Read this before writing Willow code.

## Camera Setup

When creating a camera, you must position it at the center of the viewport — not at (0, 0). The camera's X/Y is where it *looks*, so centering it on screen means the world origin aligns with the top-left corner of the window.

```go
cam := scene.NewCamera(willow.Rect{X: 0, Y: 0, Width: screenW, Height: screenH})
cam.X = screenW / 2
cam.Y = screenH / 2
cam.Invalidate()
```

Without this, everything renders offset by half the screen.

## Interactable Must Be Enabled

Nodes have `Interactable = false` by default. If you set `OnClick`, `OnDrag`, or any pointer callback without enabling it, nothing will happen — the node is excluded from hit testing entirely.

```go
node.Interactable = true   // required for OnClick, OnDrag, etc.
node.OnClick = func(ctx willow.ClickContext) {
    // ...
}
```

## Hit Testing and Draggable Sprites

For WhitePixel sprites (created with `willow.TextureRegion{}`), **do not set HitShape** unless you have a specific reason. The default hit test derives an AABB from the node's dimensions automatically.

```go
// Correct — let the default AABB handle it:
sp := willow.NewSprite("handle", willow.TextureRegion{})
sp.ScaleX = 32
sp.ScaleY = 32
sp.PivotX = 0.5
sp.PivotY = 0.5
sp.Interactable = true
sp.OnDrag = func(ctx willow.DragContext) {
    sp.X += ctx.DeltaX
    sp.Y += ctx.DeltaY
    sp.Invalidate()
}
```

Setting `HitShape = willow.HitRect{X: -0.5, Y: -0.5, Width: 1, Height: 1}` will **break** hit testing because `WorldToLocal` returns coordinates in the node's un-pivoted local space (0 to 1 for a 1×1 white pixel), not centered coordinates.

If you do need an explicit HitShape on a WhitePixel sprite, use:
```go
sp.HitShape = willow.HitRect{Width: 1, Height: 1}  // origin at (0,0), not centered
```

## WhitePixel Sprites (Solid Color Shapes)

Create solid-color shapes with an empty TextureRegion. Size them with ScaleX/Y and tint with Color:

```go
sprite := willow.NewSprite("box", willow.TextureRegion{})
sprite.ScaleX = 40   // width in pixels
sprite.ScaleY = 40   // height in pixels
sprite.Color = willow.Color{R: 1, G: 0, B: 0, A: 1}  // red
```

Do **not** create unique `ebiten.Image` instances for solid-color rectangles.

## Invalidate After Direct Field Mutation

When setting node fields directly (X, Y, Rotation, Alpha, ScaleX, etc.) in a loop, call `Invalidate()` once after all changes:

```go
node.X += dx
node.Y += dy
node.Rotation += 0.01
node.Invalidate()  // one call covers all field changes
```

The setter methods (`SetRotation`, `SetAlpha`, etc.) call `Invalidate()` internally, but direct field access is preferred in hot loops for performance.

## Rope Endpoint Binding (Pointer Stability)

Ropes read their Start/End/Controls positions through pointers. You must ensure the pointers remain valid and that you mutate the **same** Vec2 the Rope references.

**Common mistake:** creating a local Vec2, passing its address to `NewRope`, then copying it into a struct. The Rope still points at the original local — your struct copy is a different variable.

```go
// WRONG — pointer aliases a loop-local that won't be updated:
for i := range n {
    startPos := willow.Vec2{X: 100, Y: 200}
    r, rn := willow.NewRope("rope", img, nil, willow.RopeConfig{
        Start: &startPos,  // points at loop-local
    })
    ropes[i] = myRope{start: startPos, r: r}  // copies value, Rope still reads loop-local
}
ropes[0].start.X = 999  // Rope never sees this!

// CORRECT — allocate the struct first, then point at its fields:
rb := &ropeBinding{
    start: willow.Vec2{X: 100, Y: 200},
    end:   willow.Vec2{X: 400, Y: 200},
}
r, rn := willow.NewRope("rope", img, nil, willow.RopeConfig{
    Start: &rb.start,  // points directly at the struct field
    End:   &rb.end,
})
rb.r = r

// Each frame — mutate the same Vec2s the Rope reads:
rb.start.X = newX
rb.r.Update()
```

## Scene Graph Z-Order

Nodes render in tree order (depth-first). Later children render on top of earlier ones. There is no `MoveToFront` method — control z-order by adding nodes in the right sequence:

```go
// Ropes first (behind), then handles (on top)
scene.Root().AddChild(ropeNode)
scene.Root().AddChild(handleNode)
```

Or use `SetZIndex(z)` for explicit ordering within a parent.

## Loading Standalone Images as Sprites

Use `SetCustomImage` for one-off images not in an atlas, or register them as atlas pages:

```go
// Option A: RegisterPage + TextureRegion
scene.RegisterPage(0, img)
region := willow.TextureRegion{Page: 0, Width: 128, Height: 128, OriginalW: 128, OriginalH: 128}
sprite := willow.NewSprite("name", region)

// Option B: SetCustomImage (simpler for previews/one-offs)
sprite := willow.NewSprite("preview", willow.TextureRegion{})
sprite.SetCustomImage(img)
```

## Build and Test

```bash
go build ./...
go vet ./...
go test ./...
cd ecs && go test ./...
```
