// Package ecs provides ECS adapters for willow.
package ecs

import (
	"github.com/phanxgames/willow"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"
)

// InteractionEventType is the Donburi event type for willow interaction events.
// Subscribe to this in your ECS systems to receive pointer, drag, and pinch events.
var InteractionEventType = events.NewEventType[willow.InteractionEvent]()

type donburiStore struct {
	world donburi.World
}

// NewDonburiStore creates an EntityStore backed by a Donburi world.
// Interaction events are published to InteractionEventType and can be
// consumed with events.Subscribe and ProcessEvents.
func NewDonburiStore(world donburi.World) willow.EntityStore {
	return &donburiStore{world: world}
}

func (s *donburiStore) EmitEvent(event willow.InteractionEvent) {
	InteractionEventType.Publish(s.world, event)
}
