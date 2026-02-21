# Architecture

## What is Willow?

Willow is a **retained-mode** 2D game framework built on [Ebitengine](https://ebitengine.org/). You create a tree of nodes representing your game objects, and Willow handles passing the draw commands down to Ebitengine. This means you describe *what* to render by building a tree of nodes rather than *how* to render by issuing draw commands each frame.

You build a tree of `Node` objects — containers, sprites, text, particles, meshes — and Willow traverses that tree each frame to produce an optimized list of render commands for Ebitengine.

It sits between Ebitengine and your game:

```
Your Game             - gameplay, content, logic
Willow                - scene graph, rendering, interaction
Ebitengine            - GPU backend, window, audio, platform
```

## How You Use It

You own the game loop. You create a `Scene`, call `scene.Update()` and `scene.Draw(screen)` from your Ebitengine game, and Willow handles everything else. The `willow.Run()` helper is a convenience wrapper, not a requirement.

## Update / Draw Separation

```
┌─────────────────────────────────────────────┐
│  Your Game Loop                             │
│                                             │
│  Update():                                  │
│    scene.Update()                           │
│      ├─ Process input (mouse, touch, pinch) │
│      ├─ Fire callbacks (click, drag, etc.)  │
│      ├─ Run OnUpdate functions              │
│      ├─ Update particle simulations         │
│      ├─ Update camera follow / scroll       │
│      └─ Recompute dirty world transforms    │
│                                             │
│  Draw(screen):                              │
│    scene.Draw(screen)                       │
│      ├─ Traverse scene tree                 │
│      ├─ Cull invisible / off-screen nodes   │
│      ├─ Emit render commands                │
│      ├─ Sort by (RenderLayer, GlobalOrder)  │
│      ├─ Coalesce batches                    │
│      └─ Submit to Ebitengine                │
└─────────────────────────────────────────────┘
```

All simulation — input, particles, tweens, callbacks — happens in `Update()`. The `Draw()` phase is purely for rendering and never mutates game state.

## Scene Tree Structure

```
Scene
 └─ Root (container node)
     ├─ Background (sprite)
     ├─ GameWorld (container)
     │   ├─ TileMap (tilemap viewport)
     │   ├─ Player (sprite)
     │   ├─ Enemies (container)
     │   │   ├─ Enemy1 (sprite)
     │   │   └─ Enemy2 (sprite)
     │   └─ Particles (particle emitter)
     ├─ LightLayer (light layer node)
     └─ UI (container, RenderLayer=1)
         ├─ ScoreText (text)
         └─ Button (sprite, interactable)
```

Every visible element is a `Node` in this tree. Nodes inherit their parent's transform — moving a container moves all its children. The `RenderLayer` field provides coarse draw-order control (lower values draw first), while `ZIndex` controls order among siblings.

## Render Pipeline

When `scene.Draw()` is called:

1. **Tree traversal** — walk the scene graph depth-first, skipping invisible nodes
2. **Culling** — if a camera has `CullEnabled`, skip nodes whose world-space bounds don't intersect the viewport
3. **Command emission** — each visible node emits one or more `RenderCommand` structs (sprites, meshes, particles, tilemaps)
4. **Sorting** — commands are sorted by `(RenderLayer, GlobalOrder, tree order)` using a pre-allocated merge sort (zero heap allocations)
5. **Batching** — in `BatchModeCoalesced` (default), sprites sharing the same atlas page and blend mode accumulate vertices into a single `DrawTriangles32` submission
6. **Filter pass** — nodes with filters are rendered to offscreen buffers, filters applied in sequence, then composited back
7. **Submission** — final commands are submitted to Ebitengine

## Performance Design

Willow targets **10,000+ sprites at 120+ FPS** on desktop and **60+ FPS** on mobile. Key design choices:

- **Zero heap allocations per frame** on the hot path — sort buffers, vertex buffers, hit-test buffers, and particle pools are pre-allocated
- **Dirty flag transforms** — world matrices are only recomputed when a node or ancestor changes
- **CacheAsTree** — skips re-traversal of static subtrees by caching the render command list
- **CacheAsTexture** — renders a subtree to an offscreen image, redraws only on invalidation
- **Coalesced batching** — minimizes Ebitengine `DrawTriangles32` submissions by grouping compatible sprites

## Package Structure

Willow is a single flat Go package (`github.com/phanxgames/willow`). There are no `internal/` sub-packages. Go's unexported (lowercase) visibility serves as the internal boundary.

The ECS adapter lives in a separate submodule at `github.com/phanxgames/willow/ecs`.
