package willow

import (
	"math"
	"math/rand/v2"
)

// particle holds per-particle simulation state. Unexported; managed by ParticleEmitter.
type particle struct {
	x, y       float64
	vx, vy     float64
	life       float64 // remaining lifetime in seconds
	maxLife    float64 // initial lifetime (for computing t)
	startScale float64
	endScale   float64
	scale      float64
	startAlpha float64
	endAlpha   float64
	alpha      float64
	startR     float64
	startG     float64
	startB     float64
	endR       float64
	endG       float64
	endB       float64
	colorR     float64
	colorG     float64
	colorB     float64
}

// EmitterConfig controls how particles are spawned and behave.
type EmitterConfig struct {
	MaxParticles int
	EmitRate     float64 // particles per second
	Lifetime     Range
	Speed        Range
	Angle        Range   // emission angle in radians
	StartScale   Range
	EndScale     Range
	StartAlpha   Range
	EndAlpha     Range
	Gravity      Vec2
	StartColor   Color
	EndColor     Color
	Region       TextureRegion
	BlendMode    BlendMode
	WorldSpace   bool // when true, particles keep world position once emitted
}

// ParticleEmitter manages a pool of particles with CPU-based simulation.
type ParticleEmitter struct {
	config    EmitterConfig
	particles []particle
	alive     int
	emitAccum float64
	active    bool
	// World-space tracking: the emitter's last known world position,
	// set by the update walk so particles can be spawned at world coords.
	worldX, worldY float64
}

// newParticleEmitter creates a ParticleEmitter with a preallocated pool.
func newParticleEmitter(cfg EmitterConfig) *ParticleEmitter {
	max := cfg.MaxParticles
	if max <= 0 {
		max = 128
	}
	return &ParticleEmitter{
		config:    cfg,
		particles: make([]particle, max),
	}
}

// Start begins emitting particles.
func (e *ParticleEmitter) Start() {
	e.active = true
}

// Stop stops emitting new particles. Existing particles continue to live out.
func (e *ParticleEmitter) Stop() {
	e.active = false
}

// Reset stops emitting and kills all alive particles.
func (e *ParticleEmitter) Reset() {
	e.active = false
	e.alive = 0
	e.emitAccum = 0
}

// IsActive reports whether the emitter is currently emitting new particles.
func (e *ParticleEmitter) IsActive() bool {
	return e.active
}

// AliveCount returns the number of alive particles.
func (e *ParticleEmitter) AliveCount() int {
	return e.alive
}

// Config returns a pointer to the emitter's config for live tuning.
func (e *ParticleEmitter) Config() *EmitterConfig {
	return &e.config
}

// update advances particle simulation by dt seconds.
func (e *ParticleEmitter) update(dt float64) {
	gx := e.config.Gravity.X * dt
	gy := e.config.Gravity.Y * dt

	// Update existing particles, swap-remove dead ones.
	i := 0
	for i < e.alive {
		p := &e.particles[i]
		p.life -= dt
		if p.life <= 0 {
			// Swap with last alive particle.
			e.alive--
			e.particles[i] = e.particles[e.alive]
			continue
		}

		// Apply gravity.
		p.vx += gx
		p.vy += gy

		// Move.
		p.x += p.vx * dt
		p.y += p.vy * dt

		// Interpolate properties.
		t := 1.0 - p.life/p.maxLife
		p.scale = lerp(p.startScale, p.endScale, t)
		p.alpha = lerp(p.startAlpha, p.endAlpha, t)
		p.colorR = lerp(p.startR, p.endR, t)
		p.colorG = lerp(p.startG, p.endG, t)
		p.colorB = lerp(p.startB, p.endB, t)

		i++
	}

	// Emit new particles.
	if e.active && e.config.EmitRate > 0 {
		e.emitAccum += e.config.EmitRate * dt
		for e.emitAccum >= 1.0 {
			e.emitAccum -= 1.0
			if e.alive < len(e.particles) {
				e.spawnParticle()
			}
		}
	}
}

// spawnParticle initializes the particle at slot e.alive and increments alive.
func (e *ParticleEmitter) spawnParticle() {
	p := &e.particles[e.alive]

	angle := e.config.Angle.Random()
	speed := e.config.Speed.Random()
	p.vx = math.Cos(angle) * speed
	p.vy = math.Sin(angle) * speed

	if e.config.WorldSpace {
		p.x = e.worldX
		p.y = e.worldY
	} else {
		p.x = 0
		p.y = 0
	}

	p.life = e.config.Lifetime.Random()
	if p.life <= 0 {
		p.life = 1.0
	}
	p.maxLife = p.life

	p.startScale = e.config.StartScale.Random()
	p.endScale = e.config.EndScale.Random()
	p.scale = p.startScale

	p.startAlpha = e.config.StartAlpha.Random()
	p.endAlpha = e.config.EndAlpha.Random()
	p.alpha = p.startAlpha

	p.startR = float64(e.config.StartColor.R)
	p.startG = float64(e.config.StartColor.G)
	p.startB = float64(e.config.StartColor.B)
	p.endR = float64(e.config.EndColor.R)
	p.endG = float64(e.config.EndColor.G)
	p.endB = float64(e.config.EndColor.B)
	p.colorR = p.startR
	p.colorG = p.startG
	p.colorB = p.startB

	e.alive++
}

// updateParticles walks the node tree and updates all particle emitters.
func updateParticles(n *Node, dt float64) {
	if n.Type == NodeTypeParticleEmitter && n.Emitter != nil {
		if n.Emitter.config.WorldSpace {
			// Store the emitter's world position so particles spawn in world space.
			n.Emitter.worldX = n.worldTransform[4]
			n.Emitter.worldY = n.worldTransform[5]
		}
		n.Emitter.update(dt)
	}
	for _, child := range n.children {
		updateParticles(child, dt)
	}
}

// lerp linearly interpolates between a and b by t.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Random returns a random float64 in [Min, Max].
func (r Range) Random() float64 {
	if r.Min == r.Max {
		return r.Min
	}
	return r.Min + rand.Float64()*(r.Max-r.Min)
}
