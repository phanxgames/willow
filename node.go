package willow

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// --- Placeholder types (replaced by later phases) ---

// Font and TextBlock are defined in text.go (Phase 07).
// ParticleEmitter and EmitterConfig are defined in particle.go (Phase 10).

// Filter is defined in filter.go (Phase 09).

// HitShape is used for custom hit testing regions (Phase 08).
type HitShape interface {
	Contains(x, y float64) bool
}

// --- Callback context placeholders (Phase 08) ---

// PointerContext carries pointer event data.
type PointerContext struct {
	Node      *Node
	EntityID  uint32
	UserData  any
	GlobalX   float64
	GlobalY   float64
	LocalX    float64
	LocalY    float64
	Button    MouseButton
	PointerID int
	Modifiers KeyModifiers
}

// ClickContext carries click event data.
type ClickContext struct {
	Node      *Node
	EntityID  uint32
	UserData  any
	GlobalX   float64
	GlobalY   float64
	LocalX    float64
	LocalY    float64
	Button    MouseButton
	PointerID int
	Modifiers KeyModifiers
}

// DragContext carries drag event data.
type DragContext struct {
	Node      *Node
	EntityID  uint32
	UserData  any
	GlobalX   float64
	GlobalY   float64
	LocalX    float64
	LocalY    float64
	StartX    float64
	StartY    float64
	DeltaX    float64
	DeltaY    float64
	Button    MouseButton
	PointerID int
	Modifiers KeyModifiers
}

// PinchContext carries pinch gesture data.
type PinchContext struct {
	CenterX, CenterY float64
	Scale, ScaleDelta float64
	Rotation, RotDelta float64
}

// --- ID counter ---

// nodeIDCounter is a plain counter (no atomic — willow is single-threaded).
var nodeIDCounter uint32

func nextNodeID() uint32 {
	nodeIDCounter++
	return nodeIDCounter
}

// --- Node ---

// Node is the fundamental scene graph element. A single flat struct is used for
// all node types to avoid interface dispatch on the hot path.
type Node struct {
	// Identity
	ID   uint32
	Name string
	Type NodeType

	// Hierarchy
	Parent   *Node
	children []*Node

	// Transform (local)
	X, Y         float64
	ScaleX       float64
	ScaleY       float64
	Rotation     float64
	SkewX, SkewY float64
	PivotX       float64
	PivotY       float64

	// Computed (unexported, updated during traversal in Phase 03)
	worldTransform [6]float64
	worldAlpha     float64
	transformDirty bool

	// Visibility & interaction
	Alpha        float64
	Visible      bool
	Renderable   bool
	Interactable bool

	// Ordering
	ZIndex      int
	RenderLayer uint8
	GlobalOrder int

	// Metadata
	UserData any
	EntityID uint32

	// Sprite fields (NodeTypeSprite)
	TextureRegion TextureRegion
	BlendMode     BlendMode
	Color         Color
	customImage   *ebiten.Image // user-provided offscreen canvas (RenderTexture)

	// Mesh fields (NodeTypeMesh)
	Vertices        []ebiten.Vertex
	Indices         []uint16
	MeshImage       *ebiten.Image
	transformedVerts []ebiten.Vertex // preallocated transform buffer
	meshAABB         Rect            // cached local-space AABB
	meshAABBDirty    bool            // recompute AABB when true

	// Particle fields (NodeTypeParticleEmitter)
	Emitter *ParticleEmitter

	// Text fields (NodeTypeText)
	TextBlock *TextBlock

	// Hit testing
	HitShape HitShape

	// Filters
	Filters []Filter

	// Cache fields (Phase 09)
	cacheEnabled bool
	cacheTexture *ebiten.Image
	cacheDirty   bool

	// Mask field (Phase 09)
	mask *Node

	// Per-node callbacks (nil by default; zero cost when unused)
	OnPointerDown func(PointerContext)
	OnPointerUp   func(PointerContext)
	OnPointerMove func(PointerContext)
	OnClick       func(ClickContext)
	OnDragStart   func(DragContext)
	OnDrag        func(DragContext)
	OnDragEnd      func(DragContext)
	OnPinch        func(PinchContext)
	OnPointerEnter func(PointerContext)
	OnPointerLeave func(PointerContext)

	// Internal
	disposed       bool
	childrenSorted bool
	sortedChildren []*Node // reused buffer for ZIndex-sorted traversal order
}

// nodeDefaults sets the common default field values shared by all constructors.
func nodeDefaults(n *Node) {
	n.ID = nextNodeID()
	n.ScaleX = 1
	n.ScaleY = 1
	n.Alpha = 1
	n.Color = Color{1, 1, 1, 1}
	n.Visible = true
	n.Renderable = true
	n.transformDirty = true
	n.childrenSorted = true
}

// NewContainer creates a container node with no visual representation.
func NewContainer(name string) *Node {
	n := &Node{Name: name, Type: NodeTypeContainer}
	nodeDefaults(n)
	return n
}

// NewSprite creates a sprite node that renders a texture region.
func NewSprite(name string, region TextureRegion) *Node {
	n := &Node{Name: name, Type: NodeTypeSprite, TextureRegion: region}
	nodeDefaults(n)
	return n
}

// NewMesh creates a mesh node that uses DrawTriangles for rendering.
func NewMesh(name string, img *ebiten.Image, vertices []ebiten.Vertex, indices []uint16) *Node {
	n := &Node{
		Name:          name,
		Type:          NodeTypeMesh,
		MeshImage:     img,
		Vertices:      vertices,
		Indices:       indices,
		meshAABBDirty: true,
	}
	nodeDefaults(n)
	return n
}

// NewParticleEmitter creates a particle emitter node with a preallocated pool.
func NewParticleEmitter(name string, cfg EmitterConfig) *Node {
	emitter := newParticleEmitter(cfg)
	n := &Node{
		Name:          name,
		Type:          NodeTypeParticleEmitter,
		TextureRegion: cfg.Region,
		BlendMode:     cfg.BlendMode,
		Emitter:       emitter,
	}
	nodeDefaults(n)
	return n
}

// NewText creates a text node with the given content and font.
func NewText(name string, content string, font Font) *Node {
	n := &Node{
		Name: name,
		Type: NodeTypeText,
		TextBlock: &TextBlock{
			Content:     content,
			Font:        font,
			Color:       Color{1, 1, 1, 1},
			layoutDirty: true,
			ttfPage:     -1,
		},
	}
	nodeDefaults(n)
	return n
}

// SetCustomImage sets a user-provided *ebiten.Image to display instead of TextureRegion.
// Used by RenderTexture to attach a persistent offscreen canvas to a sprite node.
func (n *Node) SetCustomImage(img *ebiten.Image) {
	n.customImage = img
}

// CustomImage returns the user-provided image, or nil if not set.
func (n *Node) CustomImage() *ebiten.Image {
	return n.customImage
}

// --- Tree manipulation ---

// AddChild appends child to this node's children.
// If child already has a parent, it is removed from that parent first.
// Panics if child is nil or child is an ancestor of this node (cycle).
func (n *Node) AddChild(child *Node) {
	if child == nil {
		panic("willow: cannot add nil child")
	}
	if globalDebug {
		debugCheckDisposed(n, "AddChild (parent)")
		debugCheckDisposed(child, "AddChild (child)")
	}
	if isAncestor(child, n) {
		panic("willow: adding child would create a cycle")
	}
	if child.Parent != nil {
		child.Parent.removeChildByPtr(child)
	}
	child.Parent = n
	n.children = append(n.children, child)
	n.childrenSorted = false
	markSubtreeDirty(child)
	if globalDebug {
		debugCheckTreeDepth(child)
		debugCheckChildCount(n)
	}
}

// AddChildAt inserts child at the given index.
// Same reparenting and cycle-check behavior as AddChild.
func (n *Node) AddChildAt(child *Node, index int) {
	if child == nil {
		panic("willow: cannot add nil child")
	}
	if globalDebug {
		debugCheckDisposed(n, "AddChildAt (parent)")
		debugCheckDisposed(child, "AddChildAt (child)")
	}
	if isAncestor(child, n) {
		panic("willow: adding child would create a cycle")
	}
	if index < 0 || index > len(n.children) {
		panic("willow: child index out of range")
	}
	if child.Parent != nil {
		child.Parent.removeChildByPtr(child)
	}
	child.Parent = n
	n.children = append(n.children, nil)
	copy(n.children[index+1:], n.children[index:])
	n.children[index] = child
	n.childrenSorted = false
	markSubtreeDirty(child)
	if globalDebug {
		debugCheckTreeDepth(child)
		debugCheckChildCount(n)
	}
}

// RemoveChild detaches child from this node.
// Panics if child.Parent != n.
func (n *Node) RemoveChild(child *Node) {
	if globalDebug {
		debugCheckDisposed(n, "RemoveChild (parent)")
		debugCheckDisposed(child, "RemoveChild (child)")
	}
	if child.Parent != n {
		panic("willow: child's parent is not this node")
	}
	n.removeChildByPtr(child)
	child.Parent = nil
	n.childrenSorted = false
	markSubtreeDirty(child)
}

// RemoveChildAt removes and returns the child at the given index.
func (n *Node) RemoveChildAt(index int) *Node {
	if globalDebug {
		debugCheckDisposed(n, "RemoveChildAt")
	}
	if index < 0 || index >= len(n.children) {
		panic("willow: child index out of range")
	}
	child := n.children[index]
	copy(n.children[index:], n.children[index+1:])
	n.children[len(n.children)-1] = nil
	n.children = n.children[:len(n.children)-1]
	child.Parent = nil
	n.childrenSorted = false
	markSubtreeDirty(child)
	return child
}

// RemoveFromParent detaches this node from its parent.
// No-op if this node has no parent.
func (n *Node) RemoveFromParent() {
	if n.Parent == nil {
		return
	}
	n.Parent.RemoveChild(n)
}

// RemoveChildren detaches all children from this node.
// Children are NOT disposed.
func (n *Node) RemoveChildren() {
	for _, child := range n.children {
		child.Parent = nil
		markSubtreeDirty(child)
	}
	n.children = n.children[:0]
	n.childrenSorted = true
}

// Children returns the child list. The returned slice MUST NOT be mutated by the caller.
func (n *Node) Children() []*Node {
	return n.children
}

// NumChildren returns the number of children.
func (n *Node) NumChildren() int {
	return len(n.children)
}

// ChildAt returns the child at the given index.
func (n *Node) ChildAt(index int) *Node {
	return n.children[index]
}

// SetChildIndex moves child to a new index among its siblings.
func (n *Node) SetChildIndex(child *Node, index int) {
	if child.Parent != n {
		panic("willow: child's parent is not this node")
	}
	nc := len(n.children)
	if index < 0 || index >= nc {
		panic("willow: child index out of range")
	}
	oldIndex := -1
	for i, c := range n.children {
		if c == child {
			oldIndex = i
			break
		}
	}
	if oldIndex == index {
		return
	}
	// Shift elements to fill the gap and open the target slot.
	if oldIndex < index {
		copy(n.children[oldIndex:], n.children[oldIndex+1:index+1])
	} else {
		copy(n.children[index+1:], n.children[index:oldIndex])
	}
	n.children[index] = child
	n.childrenSorted = false
}

// SetZIndex sets the node's ZIndex and marks the parent's children as unsorted.
func (n *Node) SetZIndex(z int) {
	if n.ZIndex == z {
		return
	}
	n.ZIndex = z
	if n.Parent != nil {
		n.Parent.childrenSorted = false
	}
}

// --- Disposal ---

// Dispose removes this node from its parent, marks it as disposed,
// and recursively disposes all descendants.
func (n *Node) Dispose() {
	if n.disposed {
		return
	}
	n.RemoveFromParent()
	n.dispose()
}

func (n *Node) dispose() {
	n.disposed = true
	n.ID = 0
	for _, child := range n.children {
		child.Parent = nil
		child.dispose()
	}
	n.children = nil
	n.sortedChildren = nil
	n.Parent = nil
	n.HitShape = nil
	n.Filters = nil
	n.cacheEnabled = false
	if n.cacheTexture != nil {
		n.cacheTexture.Deallocate()
		n.cacheTexture = nil
	}
	n.cacheDirty = false
	n.mask = nil
	n.customImage = nil
	n.MeshImage = nil
	n.transformedVerts = nil
	n.Emitter = nil
	n.TextBlock = nil
	n.UserData = nil
	n.OnPointerDown = nil
	n.OnPointerUp = nil
	n.OnPointerMove = nil
	n.OnClick = nil
	n.OnDragStart = nil
	n.OnDrag = nil
	n.OnDragEnd = nil
	n.OnPinch = nil
	n.OnPointerEnter = nil
	n.OnPointerLeave = nil
}

// IsDisposed returns true if this node has been disposed.
func (n *Node) IsDisposed() bool {
	return n.disposed
}

// --- Helpers ---

// isAncestor reports whether candidate is an ancestor of node.
func isAncestor(candidate, node *Node) bool {
	for p := node; p != nil; p = p.Parent {
		if p == candidate {
			return true
		}
	}
	return false
}

// removeChildByPtr removes child from n.children without clearing child.Parent.
// Uses copy+nil to avoid retaining a dangling pointer in the backing array.
func (n *Node) removeChildByPtr(child *Node) {
	for i, c := range n.children {
		if c == child {
			copy(n.children[i:], n.children[i+1:])
			n.children[len(n.children)-1] = nil
			n.children = n.children[:len(n.children)-1]
			return
		}
	}
}

// markSubtreeDirty sets transformDirty on node and all its descendants.
func markSubtreeDirty(node *Node) {
	node.transformDirty = true
	for _, child := range node.children {
		markSubtreeDirty(child)
	}
}
