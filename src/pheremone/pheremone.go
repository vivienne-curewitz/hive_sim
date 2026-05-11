package pheremone

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
	Strength   float64
	Expiration float64
}

type AveragePheremone struct {
	Type     Pheremone
	strength float64
	Count    int
}

func NewAveragePheremone(ph PheremoneMark) AveragePheremone {
	return AveragePheremone{
		Type:     ph.Type,
		strength: ph.Strength,
		Count:    1,
	}
}

func (ap *AveragePheremone) AddPheremoneMark(pm PheremoneMark) {
	ap.strength += pm.Strength
	ap.Count++
}

func (ap *AveragePheremone) RemovePheremoneMark(pm PheremoneMark) {
	ap.strength -= pm.Strength
	ap.Count--
}

func (ap *AveragePheremone) Strength() float64 {
	if ap.Count == 0 {
		return 0.0
	}
	return ap.strength / float64(ap.Count)
}
