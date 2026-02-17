package willow

// syntheticPointerEvent represents a single injected pointer event.
// Screen coordinates are used (matching what an AI sees in screenshots)
// and converted to world coordinates via the primary camera, identical
// to real mouse input.
type syntheticPointerEvent struct {
	screenX, screenY float64
	pressed          bool
	button           MouseButton
}

// InjectPress queues a pointer press event at the given screen coordinates
// (left button). The event is consumed on the next frame's processInput call.
func (s *Scene) InjectPress(x, y float64) {
	s.injectQueue = append(s.injectQueue, syntheticPointerEvent{
		screenX: x, screenY: y,
		pressed: true,
		button:  MouseButtonLeft,
	})
}

// InjectMove queues a pointer move event at the given screen coordinates
// with the button held down. Use this between InjectPress and InjectRelease
// to simulate a drag.
func (s *Scene) InjectMove(x, y float64) {
	s.injectQueue = append(s.injectQueue, syntheticPointerEvent{
		screenX: x, screenY: y,
		pressed: true,
		button:  MouseButtonLeft,
	})
}

// InjectRelease queues a pointer release event at the given screen coordinates.
func (s *Scene) InjectRelease(x, y float64) {
	s.injectQueue = append(s.injectQueue, syntheticPointerEvent{
		screenX: x, screenY: y,
		pressed: false,
		button:  MouseButtonLeft,
	})
}

// InjectClick is a convenience that queues a press followed by a release
// at the same screen coordinates. Consumes two frames.
func (s *Scene) InjectClick(x, y float64) {
	s.InjectPress(x, y)
	s.InjectRelease(x, y)
}

// InjectDrag queues a full drag sequence: press at (fromX, fromY),
// linearly interpolated moves over frames-2 intermediate frames, and
// release at (toX, toY). The total sequence consumes `frames` frames.
// Minimum frames is 2 (press + release).
func (s *Scene) InjectDrag(fromX, fromY, toX, toY float64, frames int) {
	if frames < 2 {
		frames = 2
	}
	s.InjectPress(fromX, fromY)
	steps := frames - 2
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps+1)
		x := fromX + (toX-fromX)*t
		y := fromY + (toY-fromY)*t
		s.InjectMove(x, y)
	}
	s.InjectRelease(toX, toY)
}

// processInjectedInput pops one event from the inject queue, converts
// screen→world via the primary camera, and feeds it through processPointer.
// Returns true if an event was consumed (real mouse input should be skipped).
func (s *Scene) processInjectedInput(cam *Camera, mods KeyModifiers) bool {
	if len(s.injectQueue) == 0 {
		return false
	}
	evt := s.injectQueue[0]
	copy(s.injectQueue, s.injectQueue[1:])
	s.injectQueue = s.injectQueue[:len(s.injectQueue)-1]

	wx, wy := screenToWorld(cam, evt.screenX, evt.screenY)
	s.processPointer(0, wx, wy, evt.screenX, evt.screenY, evt.pressed, evt.button, mods)
	return true
}
