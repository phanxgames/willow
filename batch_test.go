package willow

import "testing"

func TestBatchKeySameAtlasSameBlend(t *testing.T) {
	a := RenderCommand{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}}
	b := RenderCommand{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}}
	if commandBatchKey(&a) != commandBatchKey(&b) {
		t.Error("same atlas + same blend should produce same batch key")
	}
}

func TestBatchKeyDifferentBlend(t *testing.T) {
	a := RenderCommand{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}}
	b := RenderCommand{BlendMode: BlendAdd, TextureRegion: TextureRegion{Page: 0}}
	if commandBatchKey(&a) == commandBatchKey(&b) {
		t.Error("different blend modes should produce different batch keys")
	}
}

func TestBatchKeyDifferentPage(t *testing.T) {
	a := RenderCommand{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}}
	b := RenderCommand{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 1}}
	if commandBatchKey(&a) == commandBatchKey(&b) {
		t.Error("different pages should produce different batch keys")
	}
}

func TestBatchKeyDifferentShader(t *testing.T) {
	a := RenderCommand{ShaderID: 0}
	b := RenderCommand{ShaderID: 1}
	if commandBatchKey(&a) == commandBatchKey(&b) {
		t.Error("different shaders should produce different batch keys")
	}
}

func TestBatchKeyDifferentTarget(t *testing.T) {
	a := RenderCommand{TargetID: 0}
	b := RenderCommand{TargetID: 1}
	if commandBatchKey(&a) == commandBatchKey(&b) {
		t.Error("different targets should produce different batch keys")
	}
}

func TestBatchCountSameAtlas(t *testing.T) {
	cmds := []RenderCommand{
		{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}},
		{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}},
		{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}},
	}
	if got := countBatches(cmds); got != 1 {
		t.Errorf("batches = %d, want 1", got)
	}
}

func TestBatchCountDifferentBlends(t *testing.T) {
	cmds := []RenderCommand{
		{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}},
		{BlendMode: BlendAdd, TextureRegion: TextureRegion{Page: 0}},
	}
	if got := countBatches(cmds); got != 2 {
		t.Errorf("batches = %d, want 2", got)
	}
}

func TestBatchCountDifferentPages(t *testing.T) {
	cmds := []RenderCommand{
		{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 0}},
		{BlendMode: BlendNormal, TextureRegion: TextureRegion{Page: 1}},
	}
	if got := countBatches(cmds); got != 2 {
		t.Errorf("batches = %d, want 2", got)
	}
}

func TestBatchCountEmpty(t *testing.T) {
	if got := countBatches(nil); got != 0 {
		t.Errorf("batches = %d, want 0", got)
	}
}
