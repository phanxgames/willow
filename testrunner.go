package willow

import (
	"encoding/json"
	"fmt"
)

// testStep represents a single action in a test script.
type testStep struct {
	Action string  `json:"action"`
	Label  string  `json:"label,omitempty"`
	X      float64 `json:"x,omitempty"`
	Y      float64 `json:"y,omitempty"`
	FromX  float64 `json:"fromX,omitempty"`
	FromY  float64 `json:"fromY,omitempty"`
	ToX    float64 `json:"toX,omitempty"`
	ToY    float64 `json:"toY,omitempty"`
	Frames int     `json:"frames,omitempty"`
}

// testScript is the top-level JSON structure for a test script.
type testScript struct {
	Steps []testStep `json:"steps"`
}

// TestRunner sequences injected input events and screenshots across frames
// for automated visual testing. Attach to a Scene via SetTestRunner.
type TestRunner struct {
	steps     []testStep
	cursor    int
	waitCount int
	done      bool
}

// LoadTestScript parses a JSON test script and returns a TestRunner ready
// to be attached to a Scene via SetTestRunner.
func LoadTestScript(jsonData []byte) (*TestRunner, error) {
	var script testScript
	if err := json.Unmarshal(jsonData, &script); err != nil {
		return nil, fmt.Errorf("parse test script: %w", err)
	}
	if len(script.Steps) == 0 {
		return nil, fmt.Errorf("parse test script: no steps")
	}
	return &TestRunner{steps: script.Steps}, nil
}

// SetTestRunner attaches a TestRunner to the scene. The runner's step method
// is called from Scene.Update before processInput each frame.
func (s *Scene) SetTestRunner(runner *TestRunner) {
	s.testRunner = runner
}

// Done reports whether all steps in the test script have been executed.
func (r *TestRunner) Done() bool {
	return r.done
}

// step advances the test runner by one frame. Called from Scene.Update.
func (r *TestRunner) step(s *Scene) {
	if r.done {
		return
	}
	// Wait for pending injections to drain before advancing.
	if len(s.injectQueue) > 0 {
		return
	}
	// Count down wait frames.
	if r.waitCount > 0 {
		r.waitCount--
		return
	}
	if r.cursor >= len(r.steps) {
		r.done = true
		return
	}

	st := r.steps[r.cursor]
	r.cursor++

	switch st.Action {
	case "screenshot":
		s.Screenshot(st.Label)
	case "click":
		s.InjectClick(st.X, st.Y)
	case "drag":
		frames := st.Frames
		if frames < 2 {
			frames = 2
		}
		s.InjectDrag(st.FromX, st.FromY, st.ToX, st.ToY, frames)
	case "wait":
		if st.Frames > 0 {
			r.waitCount = st.Frames - 1 // this frame counts as one
		}
	}

	// Check if we've reached the end after executing.
	if r.cursor >= len(r.steps) && r.waitCount == 0 && len(s.injectQueue) == 0 {
		r.done = true
	}
}
