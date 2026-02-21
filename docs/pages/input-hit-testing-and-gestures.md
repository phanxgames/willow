# Input, Hit Testing & Gestures

Willow provides a complete input system with hit testing, pointer events, drag, and multi-touch pinch — all integrated into the scene graph.

## Making Nodes Interactive

Set `Interactable` to true and assign a `HitShape`:

```go
button := willow.NewSprite("button", region)
button.Interactable = true
button.HitShape = willow.HitRect{X: 0, Y: 0, Width: 64, Height: 32}

button.OnClick = func(ctx willow.ClickContext) {
    fmt.Println("Button clicked!")
}
```

## Hit Shapes

Three built-in hit shape types:

```go
// Rectangle
willow.HitRect{X: 0, Y: 0, Width: 100, Height: 50}

// Circle
willow.HitCircle{CenterX: 32, CenterY: 32, Radius: 32}

// Convex polygon
willow.HitPolygon{Points: []willow.Vec2{
    {X: 0, Y: 0}, {X: 50, Y: 0}, {X: 25, Y: 50},
}}
```

Custom shapes can implement the `HitShape` interface:

```go
type HitShape interface {
    Contains(x, y float64) bool
}
```

## Node-Level Callbacks

Assign callback functions directly to nodes:

```go
node.OnPointerDown = func(ctx willow.PointerContext) { /* ... */ }
node.OnPointerUp   = func(ctx willow.PointerContext) { /* ... */ }
node.OnPointerMove = func(ctx willow.PointerContext) { /* ... */ }
node.OnClick       = func(ctx willow.ClickContext)   { /* ... */ }
node.OnDragStart   = func(ctx willow.DragContext)    { /* ... */ }
node.OnDrag        = func(ctx willow.DragContext)     { /* ... */ }
node.OnDragEnd     = func(ctx willow.DragContext)     { /* ... */ }
node.OnPinch       = func(ctx willow.PinchContext)    { /* ... */ }
node.OnPointerEnter = func(ctx willow.PointerContext) { /* ... */ }
node.OnPointerLeave = func(ctx willow.PointerContext) { /* ... */ }
```

## Context Types

### PointerContext

```go
type PointerContext struct {
    Node      *Node
    EntityID  uint32
    UserData  any
    GlobalX, GlobalY float64  // world coordinates
    LocalX, LocalY   float64  // node-local coordinates
    Button    MouseButton
    PointerID int
    Modifiers KeyModifiers
}
```

### ClickContext

Same fields as `PointerContext`.

### DragContext

```go
type DragContext struct {
    Node              *Node
    EntityID          uint32
    UserData          any
    GlobalX, GlobalY  float64  // current world position
    LocalX, LocalY    float64  // current node-local position
    StartX, StartY    float64  // drag start position
    DeltaX, DeltaY    float64  // world-space delta since last frame
    ScreenDeltaX, ScreenDeltaY float64  // screen-space delta
    Button            MouseButton
    PointerID         int
    Modifiers         KeyModifiers
}
```

### PinchContext

```go
type PinchContext struct {
    CenterX, CenterY  float64  // pinch center
    Scale, ScaleDelta  float64  // cumulative and per-frame scale
    Rotation, RotDelta float64  // cumulative and per-frame rotation
}
```

## Scene-Level Handlers

Register handlers that fire for *any* node interaction. Scene-level handlers fire before per-node callbacks:

```go
handle := scene.OnClick(func(ctx willow.ClickContext) {
    fmt.Printf("Click on node: %s\n", ctx.Node.Name)
})

// Remove when no longer needed
handle.Remove()
```

All scene-level registration methods return a `CallbackHandle`:

```go
scene.OnPointerDown(fn)   // CallbackHandle
scene.OnPointerUp(fn)
scene.OnPointerMove(fn)
scene.OnPointerEnter(fn)
scene.OnPointerLeave(fn)
scene.OnClick(fn)
scene.OnDragStart(fn)
scene.OnDrag(fn)
scene.OnDragEnd(fn)
scene.OnPinch(fn)
```

## Pointer Capture

Force all pointer events to go to a specific node, regardless of hit testing:

```go
scene.CapturePointer(0, myNode)   // pointerID 0 = primary mouse
scene.ReleasePointer(0)
```

## Drag Dead Zone

Configure how many pixels the pointer must move before a drag starts (default: 4):

```go
scene.SetDragDeadZone(8.0)
```

## Mouse Buttons

```go
willow.MouseButtonLeft
willow.MouseButtonRight
willow.MouseButtonMiddle
```

## Key Modifiers

```go
willow.ModShift
willow.ModCtrl
willow.ModAlt
willow.ModMeta
```

Check modifiers in callbacks:

```go
node.OnClick = func(ctx willow.ClickContext) {
    if ctx.Modifiers&willow.ModShift != 0 {
        fmt.Println("Shift+click!")
    }
}
```

## Event Types

```go
willow.EventPointerDown
willow.EventPointerUp
willow.EventPointerMove
willow.EventClick
willow.EventDragStart
willow.EventDrag
willow.EventDragEnd
willow.EventPinch
willow.EventPointerEnter
willow.EventPointerLeave
```

These are used primarily with the ECS integration for `InteractionEvent.Type`.
