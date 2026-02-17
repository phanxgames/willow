package willow

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Screenshot queues a labeled screenshot to be captured at the end of the
// current frame's Draw call. The resulting PNG is written to ScreenshotDir
// with a timestamped filename. Safe to call from Update or Draw.
func (s *Scene) Screenshot(label string) {
	s.screenshotQueue = append(s.screenshotQueue, label)
}

// flushScreenshots captures the rendered frame for every queued label and
// writes each as a PNG file. Called at the end of Scene.Draw.
func (s *Scene) flushScreenshots(screen *ebiten.Image) {
	if len(s.screenshotQueue) == 0 {
		return
	}

	if err := os.MkdirAll(s.ScreenshotDir, 0o755); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[willow] screenshot: mkdir %s: %v\n", s.ScreenshotDir, err)
		s.screenshotQueue = s.screenshotQueue[:0]
		return
	}

	bounds := screen.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	pixels := make([]byte, 4*w*h)
	screen.ReadPixels(pixels)

	// Convert premultiplied RGBA to straight-alpha NRGBA.
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(pixels); i += 4 {
		r, g, b, a := pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]
		if a > 0 && a < 255 {
			r = uint8(min(int(r)*255/int(a), 255))
			g = uint8(min(int(g)*255/int(a), 255))
			b = uint8(min(int(b)*255/int(a), 255))
		}
		img.Pix[i] = r
		img.Pix[i+1] = g
		img.Pix[i+2] = b
		img.Pix[i+3] = a
	}

	stamp := time.Now().Format("20060102_150405")

	for _, label := range s.screenshotQueue {
		safe := sanitizeLabel(label)
		path := fmt.Sprintf("%s/%s_%s.png", s.ScreenshotDir, stamp, safe)
		if err := writePNG(path, img); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[willow] screenshot: %v\n", err)
		}
	}

	s.screenshotQueue = s.screenshotQueue[:0]
}

// writePNG encodes an image to a PNG file at the given path.
func writePNG(path string, img *image.NRGBA) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return f.Close()
}

// sanitizeLabel replaces characters that are unsafe in file names with
// underscores and falls back to "unlabeled" for empty strings.
func sanitizeLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return "unlabeled"
	}
	var b strings.Builder
	b.Grow(len(label))
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9', r == '-', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
