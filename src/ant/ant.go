package ant

import (
	"math"
	"math/rand/v2"

	"hive_sim/src/utils"

	"github.com/google/uuid"
)

type Ant interface {
	Move()
	Eat()
	Drink()
	Rest()
	ChooseAction()
}

type WorkerAnt struct {
	ID          uuid.UUID
	Pos         utils.Coordinate
	hunger      float64
	thirst      float64
	tiredness   float64
	hitpoints   float64
	Destination *utils.Coordinate
	Speed       float64
}

func NewWorkerAnt(pos utils.Coordinate) WorkerAnt {
	return WorkerAnt{
		ID:          uuid.New(),
		Pos:         pos,
		hunger:      0.0,
		tiredness:   0.0,
		hitpoints:   100.0,
		Speed:       1.0,
		Destination: nil,
	}
}

func (wa *WorkerAnt) Move() {
	var theta float64
	if wa.Destination != nil {
		theta = wa.Pos.AngleTo(*wa.Destination)
	} else {
		theta = rand.Float64() * 2 * math.Pi
	}
	dx := math.Cos(theta) * wa.Speed
	dy := math.Sin(theta) * wa.Speed
	wa.Pos.Add(dx, dy)
}

// just a stub - Eat, Drink, Rest all need backbone later
func (wa *WorkerAnt) Eat() {
	wa.hunger = 0
}

func (wa *WorkerAnt) Drink() {
	wa.thirst = 0
}

func (wa *WorkerAnt) Rest() {
	wa.tiredness = 0
}

// always move for now
func (wa *WorkerAnt) ChooseAction() {
	wa.Move()
}
