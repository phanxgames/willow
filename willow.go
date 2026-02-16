package willow

import "github.com/hajimehoshi/ebiten/v2"

// Color represents an RGBA color with components in [0, 1]. Not premultiplied.
// Premultiplication occurs at render submission time.
type Color struct {
	R, G, B, A float64
}

// ColorWhite is the default tint (no color modification).
var ColorWhite = Color{1, 1, 1, 1}

// Vec2 is a 2D vector.
type Vec2 struct {
	X, Y float64
}

// Rect is an axis-aligned rectangle.
type Rect struct {
	X, Y, Width, Height float64
}

// Contains reports whether the point (x, y) lies inside the rectangle.
// Points on the edge are considered inside.
func (r Rect) Contains(x, y float64) bool {
	return x >= r.X && x <= r.X+r.Width &&
		y >= r.Y && y <= r.Y+r.Height
}

// Intersects reports whether r and other overlap.
// Adjacent rectangles (sharing only an edge) are considered intersecting.
func (r Rect) Intersects(other Rect) bool {
	return r.X <= other.X+other.Width &&
		r.X+r.Width >= other.X &&
		r.Y <= other.Y+other.Height &&
		r.Y+r.Height >= other.Y
}

// Range is a general-purpose min/max range.
// Used by the particle system (EmitterConfig) and potentially other systems.
type Range struct {
	Min, Max float64
}

// BlendMode selects a compositing operation. Each maps to a specific ebiten.Blend value.
type BlendMode uint8

const (
	BlendNormal   BlendMode = iota // source-over (standard alpha blending)
	BlendAdd                       // additive / lighter
	BlendMultiply                  // multiply (source * destination; only darkens)
	BlendScreen                    // screen (1 - (1-src)*(1-dst); only brightens)
	BlendErase                     // destination-out (punch transparent holes)
	BlendMask                      // clip destination to source alpha
	BlendBelow                     // destination-over (draw behind existing content)
	BlendNone                      // opaque copy (skip blending)
)

// EbitenBlend returns the ebiten.Blend value corresponding to this BlendMode.
func (b BlendMode) EbitenBlend() ebiten.Blend {
	switch b {
	case BlendNormal:
		return ebiten.BlendSourceOver
	case BlendAdd:
		return ebiten.BlendLighter
	case BlendMultiply:
		return ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorDestinationColor,
			BlendFactorSourceAlpha:      ebiten.BlendFactorDestinationAlpha,
			BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceAlpha,
			BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha,
			BlendOperationRGB:           ebiten.BlendOperationAdd,
			BlendOperationAlpha:         ebiten.BlendOperationAdd,
		}
	case BlendScreen:
		return ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorOne,
			BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
			BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceColor,
			BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha,
			BlendOperationRGB:           ebiten.BlendOperationAdd,
			BlendOperationAlpha:         ebiten.BlendOperationAdd,
		}
	case BlendErase:
		return ebiten.BlendDestinationOut
	case BlendMask:
		return ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorZero,
			BlendFactorSourceAlpha:      ebiten.BlendFactorZero,
			BlendFactorDestinationRGB:   ebiten.BlendFactorSourceAlpha,
			BlendFactorDestinationAlpha: ebiten.BlendFactorSourceAlpha,
			BlendOperationRGB:           ebiten.BlendOperationAdd,
			BlendOperationAlpha:         ebiten.BlendOperationAdd,
		}
	case BlendBelow:
		return ebiten.BlendDestinationOver
	case BlendNone:
		return ebiten.BlendCopy
	default:
		return ebiten.BlendSourceOver
	}
}

// NodeType distinguishes rendering behavior for a Node.
type NodeType uint8

const (
	NodeTypeContainer      NodeType = iota
	NodeTypeSprite
	NodeTypeMesh
	NodeTypeParticleEmitter
	NodeTypeText
)

// EventType identifies a kind of interaction event.
type EventType uint8

const (
	EventPointerDown EventType = iota
	EventPointerUp
	EventPointerMove
	EventClick
	EventDragStart
	EventDrag
	EventDragEnd
	EventPinch        // not in original spec enum; needed for ECS bridge
	EventPointerEnter // pointer entered a node's bounds
	EventPointerLeave // pointer left a node's bounds
)

// MouseButton identifies a mouse button.
type MouseButton uint8

const (
	MouseButtonLeft   MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

// KeyModifiers is a bitmask of keyboard modifier keys.
type KeyModifiers uint8

const (
	ModShift KeyModifiers = 1 << iota
	ModCtrl
	ModAlt
	ModMeta
)

// TextAlign controls horizontal text alignment within a TextBlock.
type TextAlign uint8

const (
	TextAlignLeft   TextAlign = iota
	TextAlignCenter
	TextAlignRight
)
