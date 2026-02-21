# Demos

Live WASM demos running directly in your browser — no install required.

## 10k Sprites

10,000 whelp sprites rotating, scaling, fading, and bouncing simultaneously. A stress test for the Willow rendering pipeline.

<a href="demos/sprites10k/" target="_blank">Launch Demo</a> · <a href="https://github.com/phanxgames/willow/tree/main/demos/sprites10k" target="_blank">Source Code</a>

## Physics Shapes

~50 random shapes (circles, squares, triangles, pentagons, hexagons) with gravity, collisions, and click-to-jump. Heavier shapes fall faster and are harder to push — click any shape to give it an upward impulse.

<a href="demos/physics/" target="_blank">Launch Demo</a> · <a href="https://github.com/phanxgames/willow/tree/main/demos/physics" target="_blank">Source Code</a>

## Rope Garden

A cable-untangling puzzle. Eight color-coded cables connect sockets on the left to matching sockets on the right. The pegs start shuffled, creating a tangle — drag each peg to the socket that matches its color to straighten every cable and win.

<a href="demos/ropegarden/" target="_blank">Launch Demo</a> · <a href="https://github.com/phanxgames/willow/tree/main/demos/ropegarden" target="_blank">Source Code</a>

---

Want to run demos locally? Clone the repo and use `go run`:

```bash
git clone https://github.com/phanxgames/willow.git
cd willow
go run ./demos/sprites10k
go run ./demos/physics
go run ./demos/ropegarden
```
