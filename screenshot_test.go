package willow

import "testing"

func TestSanitizeLabel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"after-spawn", "after-spawn"},
		{"frame.01", "frame.01"},
		{"has spaces", "has_spaces"},
		{"path/to/thing", "path_to_thing"},
		{"back\\slash", "back_slash"},
		{"special!@#$%", "special_____"},
		{"", "unlabeled"},
		{"   ", "unlabeled"},
		{"MixedCase123", "MixedCase123"},
	}
	for _, tt := range tests {
		got := sanitizeLabel(tt.in)
		if got != tt.want {
			t.Errorf("sanitizeLabel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestScreenshotQueueAppend(t *testing.T) {
	s := NewScene()
	s.Screenshot("a")
	s.Screenshot("b")
	s.Screenshot("c")
	if len(s.screenshotQueue) != 3 {
		t.Fatalf("queue len = %d, want 3", len(s.screenshotQueue))
	}
	if s.screenshotQueue[0] != "a" || s.screenshotQueue[1] != "b" || s.screenshotQueue[2] != "c" {
		t.Errorf("queue = %v, want [a b c]", s.screenshotQueue)
	}
}

func TestScreenshotDirDefault(t *testing.T) {
	s := NewScene()
	if s.ScreenshotDir != "screenshots" {
		t.Errorf("ScreenshotDir = %q, want %q", s.ScreenshotDir, "screenshots")
	}
}
