package willow

import "testing"

func TestLoadTestScript(t *testing.T) {
	data := []byte(`{
		"steps": [
			{"action": "screenshot", "label": "initial"},
			{"action": "click", "x": 100, "y": 200},
			{"action": "wait", "frames": 3},
			{"action": "screenshot", "label": "after-click"}
		]
	}`)

	runner, err := LoadTestScript(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runner.steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(runner.steps))
	}
	if runner.steps[0].Action != "screenshot" || runner.steps[0].Label != "initial" {
		t.Error("step 0 mismatch")
	}
	if runner.steps[1].Action != "click" || runner.steps[1].X != 100 || runner.steps[1].Y != 200 {
		t.Error("step 1 mismatch")
	}
	if runner.steps[2].Action != "wait" || runner.steps[2].Frames != 3 {
		t.Error("step 2 mismatch")
	}
}

func TestLoadTestScript_Invalid(t *testing.T) {
	_, err := LoadTestScript([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadTestScript_Empty(t *testing.T) {
	_, err := LoadTestScript([]byte(`{"steps": []}`))
	if err == nil {
		t.Error("expected error for empty steps")
	}
}

func TestRunnerStep_Click(t *testing.T) {
	s := NewScene()
	sprite := NewSprite("s", TextureRegion{OriginalW: 200, OriginalH: 200})
	sprite.Interactable = true
	s.Root().AddChild(sprite)
	updateWorldTransform(s.root, identityTransform, 1.0, false)

	data := []byte(`{"steps": [{"action": "click", "x": 50, "y": 50}]}`)
	runner, err := LoadTestScript(data)
	if err != nil {
		t.Fatal(err)
	}
	s.SetTestRunner(runner)

	// First step call: click queues press+release (2 events).
	runner.step(s)
	if len(s.injectQueue) != 2 {
		t.Fatalf("expected 2 queued events, got %d", len(s.injectQueue))
	}
	// Runner should not be done yet — injections still pending.
	if runner.Done() {
		t.Error("runner should not be done while inject queue has events")
	}

	// Drain injections.
	s.processInput()
	s.processInput()

	// Now step again — should finalize.
	runner.step(s)
	if !runner.Done() {
		t.Error("runner should be done after all steps executed and queue drained")
	}
}

func TestRunnerStep_Wait(t *testing.T) {
	s := NewScene()

	data := []byte(`{"steps": [
		{"action": "wait", "frames": 3},
		{"action": "screenshot", "label": "done"}
	]}`)
	runner, err := LoadTestScript(data)
	if err != nil {
		t.Fatal(err)
	}

	// Frame 1: execute wait (waitCount becomes 2).
	runner.step(s)
	if runner.Done() {
		t.Error("should not be done during wait")
	}

	// Frame 2: waitCount 2→1.
	runner.step(s)
	if runner.Done() {
		t.Error("should not be done during wait countdown")
	}

	// Frame 3: waitCount 1→0.
	runner.step(s)
	if runner.Done() {
		t.Error("should not be done — screenshot step not yet executed")
	}

	// Frame 4: execute screenshot step, runner finishes.
	runner.step(s)
	if !runner.Done() {
		t.Error("runner should be done after screenshot step")
	}

	// Verify screenshot was queued.
	if len(s.screenshotQueue) != 1 || s.screenshotQueue[0] != "done" {
		t.Errorf("expected screenshot 'done', got %v", s.screenshotQueue)
	}
}

func TestRunnerStep_Drag(t *testing.T) {
	s := NewScene()
	sprite := NewSprite("s", TextureRegion{OriginalW: 400, OriginalH: 400})
	sprite.Interactable = true
	s.Root().AddChild(sprite)
	updateWorldTransform(s.root, identityTransform, 1.0, false)

	data := []byte(`{"steps": [{"action": "drag", "fromX": 10, "fromY": 10, "toX": 200, "toY": 200, "frames": 4}]}`)
	runner, err := LoadTestScript(data)
	if err != nil {
		t.Fatal(err)
	}

	runner.step(s)
	if len(s.injectQueue) != 4 {
		t.Fatalf("expected 4 queued events for drag, got %d", len(s.injectQueue))
	}
}

func TestRunnerDone(t *testing.T) {
	s := NewScene()

	data := []byte(`{"steps": [{"action": "screenshot", "label": "only"}]}`)
	runner, err := LoadTestScript(data)
	if err != nil {
		t.Fatal(err)
	}

	if runner.Done() {
		t.Error("runner should not be done before any steps")
	}

	runner.step(s)
	if !runner.Done() {
		t.Error("runner should be done after single screenshot step")
	}
}

func TestRunnerWaitsForInjectQueue(t *testing.T) {
	s := NewScene()

	data := []byte(`{"steps": [
		{"action": "click", "x": 50, "y": 50},
		{"action": "screenshot", "label": "after"}
	]}`)
	runner, err := LoadTestScript(data)
	if err != nil {
		t.Fatal(err)
	}

	// Step 1: click queues 2 events.
	runner.step(s)
	if len(s.injectQueue) != 2 {
		t.Fatalf("expected 2 events, got %d", len(s.injectQueue))
	}

	// Step again — should NOT advance because inject queue is not drained.
	runner.step(s)
	if runner.cursor != 1 {
		t.Errorf("cursor should still be 1, got %d", runner.cursor)
	}

	// Drain inject queue manually.
	s.injectQueue = s.injectQueue[:0]

	// Now step — should execute screenshot.
	runner.step(s)
	if len(s.screenshotQueue) != 1 || s.screenshotQueue[0] != "after" {
		t.Errorf("expected screenshot 'after', got %v", s.screenshotQueue)
	}
	if !runner.Done() {
		t.Error("runner should be done")
	}
}
