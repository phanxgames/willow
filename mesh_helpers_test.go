package willow

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Rope ---

func TestRopeBevelJoinMode(t *testing.T) {
	// L-shaped path: bevel mode should not scale normals at the join.
	points := []Vec2{{0, 0}, {10, 0}, {10, 10}}
	_, n := NewRope("rope", nil, points, RopeConfig{Width: 4, JoinMode: RopeJoinBevel})
	if len(n.Vertices) != 6 {
		t.Errorf("vertices = %d, want 6", len(n.Vertices))
	}
	if len(n.Indices) != 12 {
		t.Errorf("indices = %d, want 12", len(n.Indices))
	}
}

func TestRopeVertexAndIndexCounts(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}, {20, 0}, {30, 0}}
	_, n := NewRope("rope", nil, points, RopeConfig{Width: 4})
	// 4 points → 8 vertices, 3 segments → 18 indices
	if len(n.Vertices) != 8 {
		t.Errorf("vertices = %d, want 8", len(n.Vertices))
	}
	if len(n.Indices) != 18 {
		t.Errorf("indices = %d, want 18", len(n.Indices))
	}
}

func TestRopeTwoPoints(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}}
	_, n := NewRope("rope", nil, points, RopeConfig{Width: 4})
	// 2 points → 4 vertices, 1 segment → 6 indices
	if len(n.Vertices) != 4 {
		t.Errorf("vertices = %d, want 4", len(n.Vertices))
	}
	if len(n.Indices) != 6 {
		t.Errorf("indices = %d, want 6", len(n.Indices))
	}

	// For horizontal segment left→right (dx=10,dy=0), left-perpendicular is (0, 1).
	// Top vertex (+ perpendicular * halfW) should be at Y = +2
	// Bottom vertex (- perpendicular * halfW) should be at Y = -2
	if !approxEqual(float64(n.Vertices[0].DstY), 2, 0.01) {
		t.Errorf("top vertex Y = %f, want 2", n.Vertices[0].DstY)
	}
	if !approxEqual(float64(n.Vertices[1].DstY), -2, 0.01) {
		t.Errorf("bottom vertex Y = %f, want -2", n.Vertices[1].DstY)
	}
}

func TestRopeUVTiling(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}, {20, 0}}
	_, n := NewRope("rope", nil, points, RopeConfig{Width: 4})
	// Cumulative length: 0, 10, 20
	if !approxEqual(float64(n.Vertices[0].SrcX), 0, 0.01) {
		t.Errorf("SrcX[0] = %f, want 0", n.Vertices[0].SrcX)
	}
	if !approxEqual(float64(n.Vertices[2].SrcX), 10, 0.01) {
		t.Errorf("SrcX[2] = %f, want 10", n.Vertices[2].SrcX)
	}
	if !approxEqual(float64(n.Vertices[4].SrcX), 20, 0.01) {
		t.Errorf("SrcX[4] = %f, want 20", n.Vertices[4].SrcX)
	}
}

func TestRopeSetPointsReusesBuffer(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}, {20, 0}}
	r, n := NewRope("rope", nil, points, RopeConfig{Width: 4})
	vertCap := cap(n.Vertices)
	indCap := cap(n.Indices)

	// Set fewer points — should not reallocate.
	r.SetPoints([]Vec2{{0, 0}, {5, 0}})
	if cap(n.Vertices) != vertCap {
		t.Errorf("vertex cap changed from %d to %d", vertCap, cap(n.Vertices))
	}
	if cap(n.Indices) != indCap {
		t.Errorf("index cap changed from %d to %d", indCap, cap(n.Indices))
	}
	if len(n.Vertices) != 4 {
		t.Errorf("vertices = %d, want 4", len(n.Vertices))
	}
	if len(n.Indices) != 6 {
		t.Errorf("indices = %d, want 6", len(n.Indices))
	}
}

func TestRopeSinglePointNoMesh(t *testing.T) {
	r, n := NewRope("rope", nil, []Vec2{{0, 0}}, RopeConfig{Width: 4})
	if len(n.Vertices) != 0 || len(n.Indices) != 0 {
		t.Error("single point rope should have no vertices/indices")
	}
	_ = r
}

func TestRopeInvalidatesMeshAABB(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}}
	r, n := NewRope("rope", nil, points, RopeConfig{Width: 4})
	n.recomputeMeshAABB()
	if n.meshAABBDirty {
		t.Error("AABB should not be dirty after recompute")
	}

	r.SetPoints([]Vec2{{0, 0}, {20, 0}})
	if !n.meshAABBDirty {
		t.Error("AABB should be dirty after SetPoints")
	}
}

// --- DistortionGrid ---

func TestDistortionGridDimensions(t *testing.T) {
	_, n := NewDistortionGrid("grid", nil, 4, 3)
	// 4 cols, 3 rows → (5*4)=20 vertices, (4*3*6)=72 indices
	if len(n.Vertices) != 20 {
		t.Errorf("vertices = %d, want 20", len(n.Vertices))
	}
	if len(n.Indices) != 72 {
		t.Errorf("indices = %d, want 72", len(n.Indices))
	}
}

func TestDistortionGridSetVertex(t *testing.T) {
	g, n := NewDistortionGrid("grid", nil, 2, 2)
	// Grid over a nil image → all at (0,0). But cols=2,rows=2, imgW=0,imgH=0
	// So all rest positions are at (0,0) since cellW=0, cellH=0.
	// Instead, test with known dimensions by checking displacement.
	// With nil image, all positions should be 0,0.
	g.SetVertex(1, 1, 5.0, 10.0)
	idx := 1*3 + 1 // row=1, vcols=3, col=1
	v := n.Vertices[idx]
	if !approxEqual(float64(v.DstX), 5, 0.01) || !approxEqual(float64(v.DstY), 10, 0.01) {
		t.Errorf("SetVertex(1,1, 5,10): got (%f,%f), want (5,10)", v.DstX, v.DstY)
	}
}

func TestDistortionGridSetAllVertices(t *testing.T) {
	g, n := NewDistortionGrid("grid", nil, 2, 2)

	g.SetAllVertices(func(col, row int, restX, restY float64) (dx, dy float64) {
		return float64(col) * 3, float64(row) * 7
	})

	// Check vertex at (2, 1): rest=(0,0) since nil image, offset=(2*3, 1*7)=(6,7)
	vcols := 3
	idx := 1*vcols + 2
	v := n.Vertices[idx]
	if !approxEqual(float64(v.DstX), 6, 0.01) || !approxEqual(float64(v.DstY), 7, 0.01) {
		t.Errorf("SetAllVertices: vertex(%d) = (%f,%f), want (6,7)", idx, v.DstX, v.DstY)
	}
}

func TestDistortionGridReset(t *testing.T) {
	g, n := NewDistortionGrid("grid", nil, 2, 2)
	g.SetVertex(0, 0, 99, 99)
	g.Reset()

	// After reset, vertex (0,0) should be back at rest position (0,0).
	v := n.Vertices[0]
	if !approxEqual(float64(v.DstX), 0, 0.01) || !approxEqual(float64(v.DstY), 0, 0.01) {
		t.Errorf("Reset: vertex(0) = (%f,%f), want (0,0)", v.DstX, v.DstY)
	}
}

func TestDistortionGridUVStability(t *testing.T) {
	g, n := NewDistortionGrid("grid", nil, 2, 2)
	// Store original UVs.
	origUVs := make([][2]float32, len(n.Vertices))
	for i, v := range n.Vertices {
		origUVs[i] = [2]float32{v.SrcX, v.SrcY}
	}

	// Deform and reset.
	g.SetVertex(1, 1, 10, 10)
	g.Reset()

	// UVs should not change.
	for i, v := range n.Vertices {
		if v.SrcX != origUVs[i][0] || v.SrcY != origUVs[i][1] {
			t.Errorf("UV changed at %d: (%f,%f) vs (%f,%f)", i, v.SrcX, v.SrcY, origUVs[i][0], origUVs[i][1])
		}
	}
}

func TestDistortionGridInvalidatesAABB(t *testing.T) {
	g, n := NewDistortionGrid("grid", nil, 2, 2)
	n.recomputeMeshAABB()
	g.SetVertex(0, 0, 5, 5)
	if !n.meshAABBDirty {
		t.Error("SetVertex should invalidate AABB")
	}

	n.recomputeMeshAABB()
	g.SetAllVertices(func(c, r int, rx, ry float64) (float64, float64) { return 0, 0 })
	if !n.meshAABBDirty {
		t.Error("SetAllVertices should invalidate AABB")
	}

	n.recomputeMeshAABB()
	g.Reset()
	if !n.meshAABBDirty {
		t.Error("Reset should invalidate AABB")
	}
}

func TestDistortionGridColsRows(t *testing.T) {
	g, _ := NewDistortionGrid("grid", nil, 5, 3)
	if g.Cols() != 5 {
		t.Errorf("Cols = %d, want 5", g.Cols())
	}
	if g.Rows() != 3 {
		t.Errorf("Rows = %d, want 3", g.Rows())
	}
}

// --- Polygon ---

func TestPolygonFanTriangulation(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	n := NewPolygon("poly", points)
	// 4 vertices, 2 triangles → 6 indices
	if len(n.Vertices) != 4 {
		t.Errorf("vertices = %d, want 4", len(n.Vertices))
	}
	if len(n.Indices) != 6 {
		t.Errorf("indices = %d, want 6", len(n.Indices))
	}
	// Fan: indices should be [0,1,2, 0,2,3]
	expected := []uint16{0, 1, 2, 0, 2, 3}
	for i, idx := range expected {
		if n.Indices[i] != idx {
			t.Errorf("Indices[%d] = %d, want %d", i, n.Indices[i], idx)
		}
	}
}

func TestPolygonTriangle(t *testing.T) {
	points := []Vec2{{0, 0}, {10, 0}, {5, 10}}
	n := NewPolygon("tri", points)
	if len(n.Vertices) != 3 {
		t.Errorf("vertices = %d, want 3", len(n.Vertices))
	}
	if len(n.Indices) != 3 {
		t.Errorf("indices = %d, want 3", len(n.Indices))
	}
}

func TestPolygonPentagon(t *testing.T) {
	// 5-sided polygon → 5 verts, 3 triangles → 9 indices
	var points []Vec2
	for i := 0; i < 5; i++ {
		angle := float64(i) * 2 * math.Pi / 5
		points = append(points, Vec2{X: math.Cos(angle) * 10, Y: math.Sin(angle) * 10})
	}
	n := NewPolygon("pent", points)
	if len(n.Vertices) != 5 {
		t.Errorf("vertices = %d, want 5", len(n.Vertices))
	}
	if len(n.Indices) != 9 {
		t.Errorf("indices = %d, want 9", len(n.Indices))
	}
}

func TestPolygonWhitePixelSharing(t *testing.T) {
	a := NewPolygon("a", []Vec2{{0, 0}, {10, 0}, {5, 10}})
	b := NewPolygon("b", []Vec2{{0, 0}, {20, 0}, {10, 20}})
	if a.MeshImage != b.MeshImage {
		t.Error("untextured polygons should share the same white pixel image")
	}
}

func TestPolygonUntexturedUV(t *testing.T) {
	n := NewPolygon("poly", []Vec2{{0, 0}, {10, 0}, {5, 10}})
	// Untextured: all UVs should be at center of white pixel (0.5, 0.5).
	for i, v := range n.Vertices {
		if v.SrcX != 0.5 || v.SrcY != 0.5 {
			t.Errorf("vertex %d UV = (%f,%f), want (0.5,0.5)", i, v.SrcX, v.SrcY)
		}
	}
}

func TestPolygonTooFewPoints(t *testing.T) {
	n := NewPolygon("tiny", []Vec2{{0, 0}, {10, 0}})
	if len(n.Vertices) != 0 || len(n.Indices) != 0 {
		t.Error("polygon with <3 points should have no vertices/indices")
	}
}

func TestSetPolygonPoints(t *testing.T) {
	n := NewPolygon("poly", []Vec2{{0, 0}, {10, 0}, {5, 10}})
	vertCap := cap(n.Vertices)

	// Update with fewer points — should reuse backing array.
	SetPolygonPoints(n, []Vec2{{0, 0}, {20, 0}, {10, 20}})
	if len(n.Vertices) != 3 {
		t.Errorf("vertices = %d, want 3", len(n.Vertices))
	}
	if cap(n.Vertices) != vertCap {
		t.Errorf("vertex backing array reallocated: was %d, now %d", vertCap, cap(n.Vertices))
	}
	if !n.meshAABBDirty {
		t.Error("SetPolygonPoints should invalidate AABB")
	}
}

func TestNewPolygonTextured(t *testing.T) {
	img := ebiten.NewImage(32, 32)
	points := []Vec2{{0, 0}, {32, 0}, {32, 32}, {0, 32}}
	n := NewPolygonTextured("texPoly", img, points)

	if len(n.Vertices) != 4 {
		t.Errorf("vertices = %d, want 4", len(n.Vertices))
	}
	if len(n.Indices) != 6 {
		t.Errorf("indices = %d, want 6", len(n.Indices))
	}
	if n.MeshImage != img {
		t.Error("MeshImage should be the provided image")
	}

	// Bottom-right vertex should have UVs mapped to image dimensions.
	v := n.Vertices[2] // point (32,32)
	if v.SrcX != 32 || v.SrcY != 32 {
		t.Errorf("textured UV at (32,32) = (%f,%f), want (32,32)", v.SrcX, v.SrcY)
	}
}

func TestSetPolygonPointsGrows(t *testing.T) {
	n := NewPolygon("poly", []Vec2{{0, 0}, {10, 0}, {5, 10}})
	// Grow to 5 points — may need new backing array.
	points := []Vec2{{0, 0}, {10, 0}, {20, 5}, {15, 15}, {0, 10}}
	SetPolygonPoints(n, points)
	if len(n.Vertices) != 5 {
		t.Errorf("vertices = %d, want 5", len(n.Vertices))
	}
	if len(n.Indices) != 9 {
		t.Errorf("indices = %d, want 9", len(n.Indices))
	}
}
