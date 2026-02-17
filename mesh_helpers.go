package willow

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Rope ---

// RopeJoinMode controls how segments join in a Rope mesh.
type RopeJoinMode uint8

const (
	// RopeJoinMiter extends segment corners to a sharp point.
	RopeJoinMiter RopeJoinMode = iota
	// RopeJoinBevel flattens corners by inserting extra vertices, avoiding spikes.
	RopeJoinBevel
)

// RopeConfig configures a Rope mesh.
type RopeConfig struct {
	Width    float64
	JoinMode RopeJoinMode
}

// Rope generates a ribbon/rope mesh that follows a polyline path.
type Rope struct {
	node   *Node
	config RopeConfig
	cumLen []float64 // preallocated cumulative length buffer
}

// NewRope creates a rope mesh node that renders a textured ribbon along the given points.
// The image is tiled along the path (SrcX) and spans the full image height (SrcY).
func NewRope(name string, img *ebiten.Image, points []Vec2, cfg RopeConfig) (*Rope, *Node) {
	n := NewMesh(name, img, nil, nil)
	r := &Rope{node: n, config: cfg}
	r.SetPoints(points)
	return r, n
}

// Node returns the underlying mesh node.
func (r *Rope) Node() *Node {
	return r.node
}

// SetPoints updates the rope's path. For N points: 2N vertices, 6(N-1) indices.
func (r *Rope) SetPoints(points []Vec2) {
	if len(points) < 2 {
		r.node.Vertices = r.node.Vertices[:0]
		r.node.Indices = r.node.Indices[:0]
		r.node.InvalidateMeshAABB()
		return
	}

	n := len(points)
	numVerts := n * 2
	numInds := (n - 1) * 6

	// Grow vertex/index slices to high-water mark.
	if cap(r.node.Vertices) < numVerts {
		r.node.Vertices = make([]ebiten.Vertex, numVerts)
	}
	r.node.Vertices = r.node.Vertices[:numVerts]

	if cap(r.node.Indices) < numInds {
		r.node.Indices = make([]uint16, numInds)
	}
	r.node.Indices = r.node.Indices[:numInds]

	imgH := float64(0)
	if r.node.MeshImage != nil {
		imgH = float64(r.node.MeshImage.Bounds().Dy())
	}

	halfW := r.config.Width / 2

	// Compute cumulative path length for UV tiling.
	if cap(r.cumLen) < n {
		r.cumLen = make([]float64, n)
	}
	r.cumLen = r.cumLen[:n]
	r.cumLen[0] = 0
	for i := 1; i < n; i++ {
		dx := points[i].X - points[i-1].X
		dy := points[i].Y - points[i-1].Y
		r.cumLen[i] = r.cumLen[i-1] + math.Sqrt(dx*dx+dy*dy)
	}

	for i := 0; i < n; i++ {
		// Compute perpendicular normal.
		var nx, ny float64
		if i == 0 {
			nx, ny = perpendicular(points[0], points[1])
		} else if i == n-1 {
			nx, ny = perpendicular(points[n-2], points[n-1])
		} else {
			// Average of adjacent segment normals (miter).
			nx0, ny0 := perpendicular(points[i-1], points[i])
			nx1, ny1 := perpendicular(points[i], points[i+1])
			nx, ny = nx0+nx1, ny0+ny1
			ln := math.Sqrt(nx*nx + ny*ny)
			if ln > 1e-10 {
				nx /= ln
				ny /= ln
			}
			if r.config.JoinMode == RopeJoinMiter {
				// Scale to maintain width at the miter, clamped to avoid
				// exaggerated spikes at sharp corners (max 2x extension).
				dot := nx0*nx + ny0*ny
				if dot > 0.1 {
					scale := 1.0 / dot
					if scale > 2.0 {
						scale = 2.0
					}
					nx *= scale
					ny *= scale
				}
			}
		}

		srcX := float32(r.cumLen[i])
		vi := i * 2
		r.node.Vertices[vi] = ebiten.Vertex{
			DstX:   float32(points[i].X + nx*halfW),
			DstY:   float32(points[i].Y + ny*halfW),
			SrcX:   srcX,
			SrcY:   0,
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}
		r.node.Vertices[vi+1] = ebiten.Vertex{
			DstX:   float32(points[i].X - nx*halfW),
			DstY:   float32(points[i].Y - ny*halfW),
			SrcX:   srcX,
			SrcY:   float32(imgH),
			ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
		}
	}

	// Build indices: two triangles per segment.
	for i := 0; i < n-1; i++ {
		ii := i * 6
		v := uint16(i * 2)
		r.node.Indices[ii+0] = v
		r.node.Indices[ii+1] = v + 1
		r.node.Indices[ii+2] = v + 2
		r.node.Indices[ii+3] = v + 1
		r.node.Indices[ii+4] = v + 3
		r.node.Indices[ii+5] = v + 2
	}

	r.node.InvalidateMeshAABB()
}

// perpendicular returns the unit left-perpendicular of the segment from a to b.
func perpendicular(a, b Vec2) (float64, float64) {
	dx := b.X - a.X
	dy := b.Y - a.Y
	ln := math.Sqrt(dx*dx + dy*dy)
	if ln < 1e-10 {
		return 0, -1
	}
	return -dy / ln, dx / ln
}

// --- DistortionGrid ---

// DistortionGrid provides a grid mesh that can be deformed per-vertex.
type DistortionGrid struct {
	node    *Node
	cols    int
	rows    int
	restPos []Vec2 // original vertex positions for Reset
}

// NewDistortionGrid creates a grid mesh over the given image. cols and rows
// define the number of cells (vertices = (cols+1) * (rows+1)).
func NewDistortionGrid(name string, img *ebiten.Image, cols, rows int) (*DistortionGrid, *Node) {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}

	var imgW, imgH float64
	if img != nil {
		b := img.Bounds()
		imgW = float64(b.Dx())
		imgH = float64(b.Dy())
	}

	vcols := cols + 1
	vrows := rows + 1
	numVerts := vcols * vrows
	numInds := cols * rows * 6

	verts := make([]ebiten.Vertex, numVerts)
	inds := make([]uint16, numInds)
	restPos := make([]Vec2, numVerts)

	cellW := imgW / float64(cols)
	cellH := imgH / float64(rows)

	for r := 0; r < vrows; r++ {
		for c := 0; c < vcols; c++ {
			idx := r*vcols + c
			x := float64(c) * cellW
			y := float64(r) * cellH
			u := float32(x)
			v := float32(y)
			verts[idx] = ebiten.Vertex{
				DstX: float32(x), DstY: float32(y),
				SrcX: u, SrcY: v,
				ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1,
			}
			restPos[idx] = Vec2{X: x, Y: y}
		}
	}

	// Build indices.
	ii := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			tl := uint16(r*vcols + c)
			tr := tl + 1
			bl := uint16((r+1)*vcols + c)
			br := bl + 1
			inds[ii+0] = tl
			inds[ii+1] = bl
			inds[ii+2] = tr
			inds[ii+3] = tr
			inds[ii+4] = bl
			inds[ii+5] = br
			ii += 6
		}
	}

	n := NewMesh(name, img, verts, inds)
	g := &DistortionGrid{node: n, cols: cols, rows: rows, restPos: restPos}
	return g, n
}

// Node returns the underlying mesh node.
func (g *DistortionGrid) Node() *Node {
	return g.node
}

// Cols returns the number of grid columns.
func (g *DistortionGrid) Cols() int { return g.cols }

// Rows returns the number of grid rows.
func (g *DistortionGrid) Rows() int { return g.rows }

// SetVertex offsets a single grid vertex by (dx, dy) from its rest position.
func (g *DistortionGrid) SetVertex(col, row int, dx, dy float64) {
	vcols := g.cols + 1
	idx := row*vcols + col
	rest := g.restPos[idx]
	g.node.Vertices[idx].DstX = float32(rest.X + dx)
	g.node.Vertices[idx].DstY = float32(rest.Y + dy)
	g.node.InvalidateMeshAABB()
}

// SetAllVertices calls fn for each vertex, passing (col, row, restX, restY).
// fn returns the (dx, dy) displacement from the rest position.
func (g *DistortionGrid) SetAllVertices(fn func(col, row int, restX, restY float64) (dx, dy float64)) {
	vcols := g.cols + 1
	vrows := g.rows + 1
	for r := 0; r < vrows; r++ {
		for c := 0; c < vcols; c++ {
			idx := r*vcols + c
			rest := g.restPos[idx]
			dx, dy := fn(c, r, rest.X, rest.Y)
			g.node.Vertices[idx].DstX = float32(rest.X + dx)
			g.node.Vertices[idx].DstY = float32(rest.Y + dy)
		}
	}
	g.node.InvalidateMeshAABB()
}

// Reset returns all vertices to their original positions.
func (g *DistortionGrid) Reset() {
	vcols := g.cols + 1
	vrows := g.rows + 1
	for r := 0; r < vrows; r++ {
		for c := 0; c < vcols; c++ {
			idx := r*vcols + c
			rest := g.restPos[idx]
			g.node.Vertices[idx].DstX = float32(rest.X)
			g.node.Vertices[idx].DstY = float32(rest.Y)
		}
	}
	g.node.InvalidateMeshAABB()
}

// --- Polygon ---

// NewPolygon creates an untextured polygon mesh from the given vertices.
// Uses fan triangulation (convex polygons). The polygon is drawn with a shared
// 1x1 white pixel image; color comes from the node's Color field.
func NewPolygon(name string, points []Vec2) *Node {
	white := ensureWhitePixel()
	verts, inds := buildPolygonFan(points, false, nil)
	return NewMesh(name, white, verts, inds)
}

// NewPolygonTextured creates a textured polygon mesh. UVs are mapped to the
// bounding box of the points, so (0,0)→top-left and (imgW,imgH)→bottom-right.
func NewPolygonTextured(name string, img *ebiten.Image, points []Vec2) *Node {
	verts, inds := buildPolygonFan(points, true, img)
	return NewMesh(name, img, verts, inds)
}

// SetPolygonPoints updates the polygon's vertices. Maintains fan triangulation.
// If textured is true and img is non-nil, UVs are mapped to the bounding box.
func SetPolygonPoints(n *Node, points []Vec2) {
	textured := false
	var img *ebiten.Image
	if n.MeshImage != nil && n.MeshImage != ensureWhitePixel() {
		textured = true
		img = n.MeshImage
	}
	verts, inds := buildPolygonFan(points, textured, img)

	// Reuse backing arrays when possible.
	if cap(n.Vertices) >= len(verts) {
		n.Vertices = n.Vertices[:len(verts)]
		copy(n.Vertices, verts)
	} else {
		n.Vertices = verts
	}
	if cap(n.Indices) >= len(inds) {
		n.Indices = n.Indices[:len(inds)]
		copy(n.Indices, inds)
	} else {
		n.Indices = inds
	}

	n.InvalidateMeshAABB()
}

// buildPolygonFan generates vertices and indices for a fan-triangulated polygon.
// N vertices, 3*(N-2) indices.
func buildPolygonFan(points []Vec2, textured bool, img *ebiten.Image) ([]ebiten.Vertex, []uint16) {
	n := len(points)
	if n < 3 {
		return nil, nil
	}

	verts := make([]ebiten.Vertex, n)
	inds := make([]uint16, (n-2)*3)

	// Compute bounding box for UV mapping (textured mode).
	var minX, minY, maxX, maxY float64
	var imgW, imgH float64
	if textured && img != nil {
		minX, minY = points[0].X, points[0].Y
		maxX, maxY = minX, minY
		for i := 1; i < n; i++ {
			if points[i].X < minX {
				minX = points[i].X
			}
			if points[i].X > maxX {
				maxX = points[i].X
			}
			if points[i].Y < minY {
				minY = points[i].Y
			}
			if points[i].Y > maxY {
				maxY = points[i].Y
			}
		}
		b := img.Bounds()
		imgW = float64(b.Dx())
		imgH = float64(b.Dy())
	}

	for i, p := range points {
		v := &verts[i]
		v.DstX = float32(p.X)
		v.DstY = float32(p.Y)
		v.ColorR = 1
		v.ColorG = 1
		v.ColorB = 1
		v.ColorA = 1

		if textured && img != nil {
			bbW := maxX - minX
			bbH := maxY - minY
			var u, vv float64
			if bbW > 0 {
				u = (p.X - minX) / bbW * imgW
			}
			if bbH > 0 {
				vv = (p.Y - minY) / bbH * imgH
			}
			v.SrcX = float32(u)
			v.SrcY = float32(vv)
		} else {
			// Untextured: map to center of white pixel (0.5, 0.5)
			v.SrcX = 0.5
			v.SrcY = 0.5
		}
	}

	// Fan triangulation: vertex 0 is the hub.
	for i := 0; i < n-2; i++ {
		inds[i*3+0] = 0
		inds[i*3+1] = uint16(i + 1)
		inds[i*3+2] = uint16(i + 2)
	}

	return verts, inds
}
