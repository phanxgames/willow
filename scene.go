package willow

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// EntityStore is the interface for optional ECS integration.
// When set on a Scene, interaction events are forwarded to the ECS.
type EntityStore interface {
	EmitEvent(event InteractionEvent)
}

// InteractionEvent carries interaction data for the ECS bridge.
type InteractionEvent struct {
	Type      EventType
	EntityID  uint32
	GlobalX   float64
	GlobalY   float64
	LocalX    float64
	LocalY    float64
	Button    MouseButton
	Modifiers KeyModifiers
	// Drag fields (valid for EventDragStart, EventDrag, EventDragEnd)
	StartX float64
	StartY float64
	DeltaX float64
	DeltaY float64
	// Pinch fields (valid for EventPinch)
	Scale      float64
	ScaleDelta float64
	Rotation   float64
	RotDelta   float64
}

const defaultCommandCap = 1024

// Scene is the top-level object that owns the node tree, cameras, input state,
// and render buffers.
type Scene struct {
	root  *Node
	store EntityStore
	debug bool

	// Cameras
	cameras []*Camera

	// Render state
	commands  []RenderCommand
	sortBuf   []RenderCommand
	pages     []*ebiten.Image
	nextPage  int // next available page index for LoadAtlas
	cullBounds    Rect       // current camera cull bounds (set per-camera during Draw)
	cullActive    bool       // whether culling is active for the current camera
	viewTransform [6]float64 // current camera view matrix for world-space particles

	// Render target pool and offscreen buffers (Phase 09)
	rtPool        renderTexturePool
	rtDeferred    []*ebiten.Image
	offscreenCmds []RenderCommand

	// Input state (Phase 08)
	handlers     handlerRegistry
	captured     [maxPointers]*Node
	pointers     [maxPointers]pointerState
	hitBuf       []*Node
	dragDeadZone float64
	touchMap     [maxPointers]ebiten.TouchID
	touchUsed    [maxPointers]bool
	prevTouchIDs []ebiten.TouchID
	pinch        pinchState
}

// NewScene creates a new scene with a pre-created root container.
func NewScene() *Scene {
	root := NewContainer("root")
	root.Interactable = true
	return &Scene{
		root:         root,
		commands:     make([]RenderCommand, 0, defaultCommandCap),
		sortBuf:      make([]RenderCommand, 0, defaultCommandCap),
		dragDeadZone: defaultDragDeadZone,
	}
}

// Root returns the scene's root container node.
func (s *Scene) Root() *Node {
	return s.root
}

// Update processes input, advances animations, and simulates particles.
func (s *Scene) Update() {
	dt := float32(1.0 / float64(ebiten.TPS()))

	// Refresh world transforms first so camera follow targets and hit testing
	// have accurate positions this frame.
	updateWorldTransform(s.root, identityTransform, 1.0, false)

	for _, cam := range s.cameras {
		cam.update(dt)
	}
	updateParticles(s.root, float64(dt))
	s.processInput()
}

// Draw traverses the scene tree, emits render commands, sorts them, and submits
// batches to the given screen image.
func (s *Scene) Draw(screen *ebiten.Image) {
	if len(s.cameras) == 0 {
		// No explicit cameras: use implicit identity camera, full screen.
		s.drawWithCamera(screen, nil)
		return
	}

	for _, cam := range s.cameras {
		cam.computeViewMatrix()
		vp := cam.Viewport
		viewportImg := screen.SubImage(image.Rect(
			int(vp.X), int(vp.Y),
			int(vp.X+vp.Width), int(vp.Y+vp.Height),
		)).(*ebiten.Image)
		s.drawWithCamera(viewportImg, cam)
	}
}

// drawWithCamera renders the scene from a camera's perspective.
// If cam is nil, uses identity view (no camera).
func (s *Scene) drawWithCamera(target *ebiten.Image, cam *Camera) {
	s.commands = s.commands[:0]

	var viewTransform [6]float64
	viewAlpha := 1.0

	if cam != nil {
		viewTransform = cam.computeViewMatrix()
		s.cullActive = cam.CullEnabled
		if cam.CullEnabled {
			s.cullBounds = cam.VisibleBounds()
		}
	} else {
		viewTransform = identityTransform
		s.cullActive = false
	}
	s.viewTransform = viewTransform

	var stats debugStats
	var t0 time.Time

	if s.debug {
		t0 = time.Now()
	}

	treeOrder := 0
	s.traverse(s.root, viewTransform, viewAlpha, false, &treeOrder)

	if s.debug {
		stats.traverseTime = time.Since(t0)
		t0 = time.Now()
	}

	s.mergeSort()

	if s.debug {
		stats.sortTime = time.Since(t0)
		stats.commandCount = len(s.commands)
		t0 = time.Now()
	}

	s.submitBatches(target)

	if s.debug {
		stats.submitTime = time.Since(t0)
		stats.batchCount = countBatches(s.commands)
		stats.drawCallCount = countDrawCalls(s.commands)
		s.debugLog(stats)
	}

	// Release deferred pooled textures used as directImage during this frame.
	for _, img := range s.rtDeferred {
		s.rtPool.Release(img)
	}
	s.rtDeferred = s.rtDeferred[:0]
}

// NewCamera creates a camera with the given viewport and adds it to the scene.
func (s *Scene) NewCamera(viewport Rect) *Camera {
	cam := newCamera(viewport)
	s.cameras = append(s.cameras, cam)
	return cam
}

// RemoveCamera removes a camera from the scene.
func (s *Scene) RemoveCamera(cam *Camera) {
	for i, c := range s.cameras {
		if c == cam {
			s.cameras = append(s.cameras[:i], s.cameras[i+1:]...)
			return
		}
	}
}

// Cameras returns the scene's camera list. The returned slice MUST NOT be mutated.
func (s *Scene) Cameras() []*Camera {
	return s.cameras
}

// SetEntityStore sets the optional ECS bridge.
func (s *Scene) SetEntityStore(store EntityStore) {
	s.store = store
}

// SetDebugMode enables or disables debug mode. When enabled, disposed-node
// access panics, tree depth and child count warnings are printed, and per-frame
// timing stats are logged to stderr.
func (s *Scene) SetDebugMode(enabled bool) {
	s.debug = enabled
	globalDebug = enabled
}

// globalDebug mirrors the most recently set Scene debug flag so that node
// operations (which lack a Scene pointer) can check it cheaply. Only valid
// with a single Scene; multiple Scenes with differing debug modes will
// reflect whichever called SetDebugMode last.
var globalDebug bool

// RegisterPage stores an atlas page image at the given index.
// The render compiler uses these to SubImage sprite regions.
func (s *Scene) RegisterPage(index int, img *ebiten.Image) {
	for len(s.pages) <= index {
		s.pages = append(s.pages, nil)
	}
	s.pages[index] = img
}

// LoadAtlas parses TexturePacker JSON, registers atlas pages with the scene,
// and returns the Atlas for region lookups. Pages are registered starting at
// the next available page index.
func (s *Scene) LoadAtlas(jsonData []byte, pages []*ebiten.Image) (*Atlas, error) {
	atlas, err := LoadAtlas(jsonData, pages)
	if err != nil {
		return nil, err
	}
	startIndex := s.nextPage
	for i, page := range pages {
		s.RegisterPage(startIndex+i, page)
	}
	s.nextPage = startIndex + len(pages)
	// Remap region page indices to account for startIndex offset.
	if startIndex > 0 {
		for name, r := range atlas.regions {
			r.Page += uint16(startIndex)
			atlas.regions[name] = r
		}
	}
	return atlas, nil
}
