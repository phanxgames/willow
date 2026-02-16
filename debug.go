package willow

import (
	"fmt"
	"os"
	"time"
)

// debugStats holds per-frame timing and draw-call metrics.
// Only populated when Scene.debug is true.
type debugStats struct {
	traverseTime  time.Duration
	sortTime      time.Duration
	batchTime     time.Duration
	submitTime    time.Duration
	commandCount  int
	batchCount    int
	drawCallCount int
}

// debugLog prints timing and draw-call stats to stderr.
func (s *Scene) debugLog(stats debugStats) {
	if !s.debug {
		return
	}
	total := stats.traverseTime + stats.sortTime + stats.batchTime + stats.submitTime
	_, _ = fmt.Fprintf(os.Stderr,
		"[willow] traverse: %v | sort: %v | batch: %v | submit: %v | total: %v\n",
		stats.traverseTime, stats.sortTime, stats.batchTime, stats.submitTime, total)
	_, _ = fmt.Fprintf(os.Stderr,
		"[willow] commands: %d | batches: %d | draw calls: %d\n",
		stats.commandCount, stats.batchCount, stats.drawCallCount)
}

// debugCheckDisposed panics with a descriptive message when a disposed node is
// used in a tree operation. Only called when Scene.debug or the node's scene is
// in debug mode. In release mode callers skip this entirely.
func debugCheckDisposed(n *Node, op string) {
	if n.disposed {
		panic(fmt.Sprintf("willow debug: %s on disposed node %q (ID was %d)", op, n.Name, n.ID))
	}
}

// debugCheckTreeDepth warns on stderr if tree depth exceeds the threshold.
const debugMaxTreeDepth = 32

func debugCheckTreeDepth(n *Node) {
	depth := 0
	for p := n; p != nil; p = p.Parent {
		depth++
	}
	if depth > debugMaxTreeDepth {
		_, _ = fmt.Fprintf(os.Stderr, "[willow] warning: tree depth %d exceeds %d (node %q)\n",
			depth, debugMaxTreeDepth, n.Name)
	}
}

// debugCheckChildCount warns on stderr if a node has more than 1000 children.
const debugMaxChildCount = 1000

func debugCheckChildCount(n *Node) {
	if len(n.children) > debugMaxChildCount {
		_, _ = fmt.Fprintf(os.Stderr, "[willow] warning: node %q has %d children (threshold %d)\n",
			n.Name, len(n.children), debugMaxChildCount)
	}
}

// countBatches counts contiguous groups of commands sharing the same batchKey.
// This reports how many draw calls a true batching implementation would produce.
func countBatches(commands []RenderCommand) int {
	if len(commands) == 0 {
		return 0
	}
	count := 1
	prev := commandBatchKey(&commands[0])
	for i := 1; i < len(commands); i++ {
		cur := commandBatchKey(&commands[i])
		if cur != prev {
			count++
			prev = cur
		}
	}
	return count
}

// countDrawCalls counts individual draw calls from the command list.
// Meshes and direct-image sprites each count as 1. Particle commands count
// as the number of alive particles.
func countDrawCalls(commands []RenderCommand) int {
	count := 0
	for i := range commands {
		cmd := &commands[i]
		switch cmd.Type {
		case CommandParticle:
			if cmd.emitter != nil {
				count += cmd.emitter.alive
			}
		default:
			count++
		}
	}
	return count
}
