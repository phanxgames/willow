// Package ecs provides ECS adapters for willow's interaction event system.
//
// The primary adapter is [NewDonburiStore], which bridges willow interaction
// events (pointer, click, drag, pinch) into a [Donburi] world as typed events.
// Subscribe to [InteractionEventType] in your ECS systems to receive them.
//
// Usage:
//
//	store := ecs.NewDonburiStore(world)
//	scene.SetEntityStore(store)
//
// [Donburi]: https://github.com/yohamta/donburi
package ecs
