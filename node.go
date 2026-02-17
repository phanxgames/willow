package willow

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// --- Placeholder types (replaced by later phases) ---

// Font and TextBlock are defined in text.go (Phase 07).
// ParticleEmitter and EmitterConfig are defined in particle.go (Phase 10).

// Filter is defined in filter.go (Phase 09).

// HitShape defines a custom hit testing region in local coordinates.
// Implement this interface and assign it to Node.HitShape to override
// the default axis-aligned bounding box test.
type HitShape interface {
	// Contains reports whether the local-space point (x, y) is inside the shape.
	Contains(x, y float64) bool
}

// --- Callback context placeholders (Phase 08) ---

// PointerContext carries pointer event data passed to pointer callbacks.
type PointerContext struct {
	Node      *Node        // the node under the pointer, or nil if none
	EntityID  uint32       // the hit node's EntityID (for ECS bridging)
	UserData  any          // the hit node's UserData
	GlobalX   float64      // pointer X in world coordinates
	GlobalY   float64      // pointer Y in world coordinates
	LocalX    float64      // pointer X in the hit node's local coordinates
	LocalY    float64      // pointer Y in the hit node's local coordinates
	Button    MouseButton  // which mouse button is involved
	PointerID int          // 0 = mouse, 1-9 = touch contacts
	Modifiers KeyModifiers // keyboard modifier keys held during the event
}

// ClickContext carries click event data passed to click callbacks.
type ClickContext struct {
	Node      *Node        // the clicked node
	EntityID  uint32       // the clicked node's EntityID (for ECS bridging)
	UserData  any          // the clicked node's UserData
	GlobalX   float64      // click X in world coordinates
	GlobalY   float64      // click Y in world coordinates
	LocalX    float64      // click X in the node's local coordinates
	LocalY    float64      // click Y in the node's local coordinates
	Button    MouseButton  // which mouse button was clicked
	PointerID int          // 0 = mouse, 1-9 = touch contacts
	Modifiers KeyModifiers // keyboard modifier keys held during the click
}

// DragContext carries drag event data passed to drag callbacks.
type DragContext struct {
	Node         *Node        // the node being dragged
	EntityID     uint32       // the dragged node's EntityID (for ECS bridging)
	UserData     any          // the dragged node's UserData
	GlobalX      float64      // current pointer X in world coordinates
	GlobalY      float64      // current pointer Y in world coordinates
	LocalX       float64      // current pointer X in the node's local coordinates
	LocalY       float64      // current pointer Y in the node's local coordinates
	StartX       float64      // world X where the drag began
	StartY       float64      // world Y where the drag began
	DeltaX       float64      // X movement since the previous drag event
	DeltaY       float64      // Y movement since the previous drag event
	ScreenDeltaX float64      // X movement in screen pixels since the previous drag event
	ScreenDeltaY float64      // Y movement in screen pixels since the previous drag event
	Button       MouseButton  // which mouse button initiated the drag
	PointerID    int          // 0 = mouse, 1-9 = touch contacts
	Modifiers    KeyModifiers // keyboard modifier keys held during the drag
}

// PinchContext carries two-finger pinch/rotate gesture data.
type PinchContext struct {
	CenterX, CenterY   float64 // midpoint between the two touch points in world coordinates
	Scale, ScaleDelta  float64 // cumulative scale factor and frame-to-frame change
	Rotation, RotDelta float64 // cumulative rotation (radians) and frame-to-frame change
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

	// ID is a unique auto-assigned identifier (never zero for live nodes).
	ID uint32
	// Name is a human-readable label for debugging; not used for lookups.
	Name string
	// Type determines how this node is rendered (container, sprite, mesh, etc.).
	Type NodeType

	// Hierarchy

	// Parent points to this node's parent, or nil for the root.
	Parent   *Node
	children []*Node

	// Transform (local, relative to Parent)

	// X and Y are the local-space position in pixels (origin at top-left, Y down).
	X, Y float64
	// ScaleX and ScaleY are the local scale factors (1.0 = no scaling).
	ScaleX float64
	ScaleY float64
	// Rotation is the local rotation in radians (clockwise).
	Rotation float64
	// SkewX and SkewY are shear angles in radians.
	SkewX, SkewY float64
	// PivotX and PivotY are the transform origin in local pixels. Scale,
	// rotation, and skew are applied around this point.
	PivotX float64
	PivotY float64

	// Computed (unexported, updated during traversal)
	worldTransform [6]float64
	worldAlpha     float64
	transformDirty bool

	// Visibility & interaction

	// Alpha is the node's opacity in [0, 1]. Multiplied with the parent's
	// computed alpha, so children inherit transparency.
	Alpha float64
	// Visible controls whether this node and its subtree are drawn.
	// An invisible node is also excluded from hit testing.
	Visible bool
	// Renderable controls whether this node emits render commands. When false
	// the node is skipped during drawing but its children are still traversed.
	Renderable bool
	// Interactable controls whether this node responds to pointer events.
	// When false the entire subtree is excluded from hit testing.
	Interactable bool

	// Ordering

	// ZIndex controls draw order among siblings. Higher values draw on top.
	// Use SetZIndex to change this so the parent is notified to re-sort.
	ZIndex int
	// RenderLayer is the primary sort key for render commands.
	// All commands in a lower layer draw before any command in a higher layer.
	RenderLayer uint8
	// GlobalOrder is a secondary sort key within the same RenderLayer.
	// Set it to override the default tree-order sorting.
	GlobalOrder int

	// Metadata

	// UserData is an arbitrary value the application can attach to a node.
	UserData any
	// EntityID links this node to an ECS entity. When non-zero, interaction
	// events on this node are forwarded to the Scene's EntityStore.
	EntityID uint32

	// Sprite fields (NodeTypeSprite)

	// TextureRegion identifies the sub-image within an atlas page to draw.
	TextureRegion TextureRegion
	// BlendMode selects the compositing operation used when drawing this node.
	BlendMode BlendMode
	// Color is a multiplicative tint applied to the sprite. The default
	// {1,1,1,1} means no tint.
	Color       Color
	customImage *ebiten.Image // user-provided offscreen canvas (RenderTexture)

	// Mesh fields (NodeTypeMesh)

	// Vertices holds the mesh vertex data for DrawTriangles.
	Vertices []ebiten.Vertex
	// Indices holds the triangle index list for DrawTriangles.
	Indices []uint16
	// MeshImage is the texture sampled by DrawTriangles.
	MeshImage        *ebiten.Image
	transformedVerts []ebiten.Vertex // preallocated transform buffer
	meshAABB         Rect            // cached local-space AABB
	meshAABBDirty    bool            // recompute AABB when true

	// Particle fields (NodeTypeParticleEmitter)

	// Emitter manages the particle pool and simulation for this node.
	Emitter *ParticleEmitter

	// Text fields (NodeTypeText)

	// TextBlock holds the text content, font, and cached layout state.
	TextBlock *TextBlock

	// Update field (optional callback)

	// OnUpdate is called once per tick during Scene.Update if set.
	OnUpdate func(dt float64)

	// Hit testing

	// HitShape overrides the default AABB hit test with a custom shape.
	// Nil means use the node's bounding box.
	HitShape HitShape

	// Filters

	// Filters is the chain of visual effects applied to this node's rendered
	// output. Filters are applied in order; each reads from the previous
	// result and writes to a new buffer.
	Filters []Filter

	// Cache fields
	cacheEnabled bool
	cacheTexture *ebiten.Image
	cacheDirty   bool

	// Mask field
	mask *Node

	// Per-node callbacks (nil by default; zero cost when unused).
	// Scene-level handlers fire before per-node callbacks.

	// OnPointerDown fires when a pointer button is pressed over this node.
	OnPointerDown func(PointerContext)
	// OnPointerUp fires when a pointer button is released over this node.
	OnPointerUp func(PointerContext)
	// OnPointerMove fires when the pointer moves over this node (hover).
	OnPointerMove func(PointerContext)
	// OnClick fires on press then release over this node.
	OnClick func(ClickContext)
	// OnDragStart fires when a drag gesture begins on this node.
	OnDragStart func(DragContext)
	// OnDrag fires each frame while this node is being dragged.
	OnDrag func(DragContext)
	// OnDragEnd fires when a drag gesture ends on this node.
	OnDragEnd func(DragContext)
	// OnPinch fires during a two-finger pinch gesture over this node.
	OnPinch func(PinchContext)
	// OnPointerEnter fires when the pointer enters this node's bounds.
	OnPointerEnter func(PointerContext)
	// OnPointerLeave fires when the pointer leaves this node's bounds.
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
	// If no region is specified (zero value), default to WhitePixel
	if region == (TextureRegion{}) {
		n.customImage = WhitePixel
	}
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

// NewText creates a text node that renders the given string using font.
// The node's TextBlock is initialized with white color and dirty layout.
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
// Panics if the index is out of range.
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
// Panics if the index is out of range.
func (n *Node) ChildAt(index int) *Node {
	return n.children[index]
}

// SetChildIndex moves child to a new index among its siblings.
// Panics if child is not a child of n or if index is out of range.
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

// SetZIndex sets the node's ZIndex and marks the parent's children as unsorted,
// so the next traversal will re-sort siblings by ZIndex.
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
