package willow

import "testing"

func TestInjectClick(t *testing.T) {
	s := NewScene()
	sprite := NewSprite("s", TextureRegion{OriginalW: 100, OriginalH: 100})
	sprite.Interactable = true
	s.Root().AddChild(sprite)
	updateWorldTransform(s.root, identityTransform, 1.0, false)

	var clicked bool
	s.OnClick(func(ctx ClickContext) {
		clicked = true
		if ctx.Node != sprite {
			t.Error("expected sprite node")
		}
	})

	s.InjectClick(50, 50)
	if len(s.injectQueue) != 2 {
		t.Fatalf("expected 2 queued events, got %d", len(s.injectQueue))
	}

	// Frame 1: press
	s.processInput()
	if len(s.injectQueue) != 1 {
		t.Fatalf("expected 1 remaining event after frame 1, got %d", len(s.injectQueue))
	}
	if clicked {
		t.Error("click should not fire on press frame")
	}

	// Frame 2: release → click fires
	s.processInput()
	if len(s.injectQueue) != 0 {
		t.Fatalf("expected 0 remaining events after frame 2, got %d", len(s.injectQueue))
	}
	if !clicked {
		t.Error("click should fire on release frame")
	}
}

func TestInjectDrag(t *testing.T) {
	s := NewScene()
	sprite := NewSprite("s", TextureRegion{OriginalW: 400, OriginalH: 400})
	sprite.Interactable = true
	s.Root().AddChild(sprite)
	updateWorldTransform(s.root, identityTransform, 1.0, false)

	var events []string
	s.OnDragStart(func(ctx DragContext) { events = append(events, "dragstart") })
	s.OnDrag(func(ctx DragContext) { events = append(events, "drag") })
	s.OnDragEnd(func(ctx DragContext) { events = append(events, "dragend") })

	// Drag from (10,10) to (200,200) over 5 frames:
	// frame 0: press at (10,10)
	// frame 1: move to ~(57.5, 57.5)
	// frame 2: move to ~(105, 105)
	// frame 3: move to ~(152.5, 152.5)
	// frame 4: release at (200, 200)
	s.InjectDrag(10, 10, 200, 200, 5)
	if len(s.injectQueue) != 5 {
		t.Fatalf("expected 5 queued events, got %d", len(s.injectQueue))
	}

	// Drain all frames.
	for i := 0; i < 5; i++ {
		s.processInput()
	}

	// Should see dragstart, at least one drag, and dragend.
	if len(events) < 3 {
		t.Fatalf("expected at least 3 events, got %v", events)
	}
	if events[0] != "dragstart" {
		t.Errorf("first event should be dragstart, got %s", events[0])
	}
	if events[len(events)-1] != "dragend" {
		t.Errorf("last event should be dragend, got %s", events[len(events)-1])
	}
}

func TestInjectDrag_MinFrames(t *testing.T) {
	s := NewScene()
	s.InjectDrag(0, 0, 100, 100, 1) // should clamp to 2
	if len(s.injectQueue) != 2 {
		t.Fatalf("expected 2 queued events (clamped), got %d", len(s.injectQueue))
	}
}

func TestInjectQueueOrder(t *testing.T) {
	s := NewScene()

	s.InjectPress(10, 20)
	s.InjectMove(30, 40)
	s.InjectRelease(50, 60)

	if len(s.injectQueue) != 3 {
		t.Fatalf("expected 3 events, got %d", len(s.injectQueue))
	}

	// Verify order: press, move, release.
	if !s.injectQueue[0].pressed || s.injectQueue[0].screenX != 10 {
		t.Error("first event should be press at (10,20)")
	}
	if !s.injectQueue[1].pressed || s.injectQueue[1].screenX != 30 {
		t.Error("second event should be move at (30,40)")
	}
	if s.injectQueue[2].pressed || s.injectQueue[2].screenX != 50 {
		t.Error("third event should be release at (50,60)")
	}
}

func TestProcessInjectedInput(t *testing.T) {
	s := NewScene()
	sprite := NewSprite("s", TextureRegion{OriginalW: 100, OriginalH: 100})
	sprite.Interactable = true
	s.Root().AddChild(sprite)
	updateWorldTransform(s.root, identityTransform, 1.0, false)

	var downFired bool
	s.OnPointerDown(func(ctx PointerContext) {
		downFired = true
		if ctx.GlobalX != 50 || ctx.GlobalY != 50 {
			t.Errorf("expected global (50,50), got (%v,%v)", ctx.GlobalX, ctx.GlobalY)
		}
	})

	// No camera → screen coords = world coords.
	s.InjectPress(50, 50)
	consumed := s.processInjectedInput(nil, 0)
	if !consumed {
		t.Error("expected processInjectedInput to consume an event")
	}
	if !downFired {
		t.Error("pointer down should have fired")
	}
	if len(s.injectQueue) != 0 {
		t.Errorf("queue should be empty, got %d", len(s.injectQueue))
	}
}

func TestProcessInjectedInput_EmptyQueue(t *testing.T) {
	s := NewScene()
	consumed := s.processInjectedInput(nil, 0)
	if consumed {
		t.Error("should not consume when queue is empty")
	}
}

func TestInjectWithCamera(t *testing.T) {
	s := NewScene()
	cam := s.NewCamera(Rect{X: 0, Y: 0, Width: 640, Height: 480})
	cam.X = 320
	cam.Y = 240
	cam.Zoom = 2.0
	cam.computeViewMatrix()

	sprite := NewSprite("s", TextureRegion{OriginalW: 50, OriginalH: 50})
	sprite.Interactable = true
	sprite.X = 295
	sprite.Y = 215
	s.Root().AddChild(sprite)
	updateWorldTransform(s.root, identityTransform, 1.0, false)

	var hitNode *Node
	s.OnPointerDown(func(ctx PointerContext) {
		hitNode = ctx.Node
	})

	// Screen center (320, 240) maps to world (320, 240) with camera centered there.
	s.InjectPress(320, 240)
	s.processInjectedInput(cam, 0)

	if hitNode != sprite {
		t.Errorf("expected sprite hit via camera transform, got %v", hitNode)
	}
}
