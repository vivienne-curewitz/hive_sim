package simulation

import (
	"hive_sim/src/ant"
	"hive_sim/src/world"

	"github.com/google/uuid"
)

type Simulation struct {
	// Simulation parameters
	TimeStep float64
	Duration float64

	// Simulation state
	CurrentTime float64
	Ants        map[uuid.UUID]*ant.Ant
	Resources   map[uuid.UUID]*world.Resource
	World       *world.World
}

func NewSimulation(timeStep, duration float64) *Simulation {
	return &Simulation{
		TimeStep:    timeStep,
		Duration:    duration,
		CurrentTime: 0,
		Ants:        make(map[uuid.UUID]*ant.Ant),
		Resources:   make(map[uuid.UUID]*world.Resource),
		World:       world.NewWorld(100, 100), // Example world size
	}
}

func (s *Simulation) SingleStep() {
	// first, phase 1 -- mostly movement

	// then, phase 2 -- interactions with other -- attack other ant, take resource, etc

	// phase 3 -- resolution
}
