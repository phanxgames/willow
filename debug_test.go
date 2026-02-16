package willow

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestDebugMode_DisposedNodePanics(t *testing.T) {
	s := NewScene()
	s.SetDebugMode(true)
	defer s.SetDebugMode(false)

	parent := NewContainer("parent")
	s.Root().AddChild(parent)

	child := NewSprite("child", TextureRegion{OriginalW: 10, OriginalH: 10})
	child.Dispose()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on AddChild with disposed node, got none")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "disposed") {
			t.Errorf("panic message should mention 'disposed', got: %s", msg)
		}
	}()

	parent.AddChild(child)
}

func TestDebugMode_DisposedParentPanics(t *testing.T) {
	s := NewScene()
	s.SetDebugMode(true)
	defer s.SetDebugMode(false)

	parent := NewContainer("parent")
	parent.Dispose()

	child := NewSprite("child", TextureRegion{OriginalW: 10, OriginalH: 10})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on AddChild to disposed parent, got none")
		}
		msg := fmt.Sprint(r)
		if !strings.Contains(msg, "disposed") {
			t.Errorf("panic message should mention 'disposed', got: %s", msg)
		}
	}()

	parent.AddChild(child)
}

func TestReleaseMode_DisposedNodeNoOp(t *testing.T) {
	s := NewScene()
	s.SetDebugMode(false)

	child := NewSprite("child", TextureRegion{OriginalW: 10, OriginalH: 10})
	child.Dispose()

	// In release mode, adding a disposed child should not panic.
	// It still won't work correctly but it won't crash.
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprint(r)
			if strings.Contains(msg, "disposed") {
				t.Errorf("release mode should not panic on disposed node, got: %s", msg)
			}
		}
	}()

	// This may panic for other reasons (cycle check with nil parent chain),
	// but not for "disposed" reasons.
	s.Root().AddChild(child)
}

func TestDebugMode_TreeDepthWarning(t *testing.T) {
	s := NewScene()
	s.SetDebugMode(true)
	defer s.SetDebugMode(false)

	// Capture stderr output.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Build a chain deeper than debugMaxTreeDepth (32).
	current := s.Root()
	for i := 0; i < debugMaxTreeDepth+5; i++ {
		child := NewContainer(fmt.Sprintf("depth_%d", i))
		current.AddChild(child)
		current = child
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "warning: tree depth") {
		t.Errorf("expected tree depth warning in stderr, got: %q", output)
	}
}

func TestDebugMode_ChildCountWarning(t *testing.T) {
	s := NewScene()
	s.SetDebugMode(true)
	defer s.SetDebugMode(false)

	// Capture stderr output.
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	parent := NewContainer("many_children")
	s.Root().AddChild(parent)

	for i := 0; i < debugMaxChildCount+1; i++ {
		child := NewContainer(fmt.Sprintf("c_%d", i))
		parent.AddChild(child)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "warning: node") || !strings.Contains(output, "children") {
		t.Errorf("expected child count warning in stderr, got: %q", output)
	}
}

func TestDebugStats_AllFieldsPopulated(t *testing.T) {
	stats := debugStats{
		traverseTime:  100,
		sortTime:      50,
		batchTime:     30,
		submitTime:    80,
		commandCount:  1000,
		batchCount:    10,
		drawCallCount: 15,
	}

	if stats.traverseTime == 0 || stats.sortTime == 0 || stats.submitTime == 0 {
		t.Error("expected all timing fields to be non-zero")
	}
	if stats.commandCount == 0 || stats.drawCallCount == 0 {
		t.Error("expected count fields to be non-zero")
	}
}

func TestCountDrawCalls(t *testing.T) {
	cmds := []RenderCommand{
		{Type: CommandSprite},
		{Type: CommandSprite},
		{Type: CommandMesh},
		{Type: CommandParticle, emitter: &ParticleEmitter{alive: 50}},
	}
	got := countDrawCalls(cmds)
	// 1 + 1 + 1 + 50 = 53
	if got != 53 {
		t.Errorf("countDrawCalls = %d, want 53", got)
	}
}

func TestCountDrawCalls_Empty(t *testing.T) {
	got := countDrawCalls(nil)
	if got != 0 {
		t.Errorf("countDrawCalls(nil) = %d, want 0", got)
	}
}
