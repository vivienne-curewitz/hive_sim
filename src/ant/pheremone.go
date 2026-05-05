package ant

import "hive_sim/src/utils"

type Pheremone int

const (
	PheremoneFood Pheremone = iota
	PheremoneHome
	PheremoneDeath
	PheremonePath
)

// PheremoneMark - struct that describes a Pheremone in the world
/*
pheremones are pointers in the world left by ants
they signal all sorts of things about the world around them
after some time, they expire.
this is the primary way ants navigate and think
*/
type PheremoneMark struct {
	Type Pheremone
	Pos  utils.Coordinate
	// points to the last dropped pheremonemark, forming a chain
	Direction         float64
	LifeTimeRemaining int64
}

func (ph *PheremoneMark) Step(step int64) {
	ph.LifeTimeRemaining -= step
}
