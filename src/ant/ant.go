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
	ID        uuid.UUID
	Pos       utils.Coordinate
	hunger    float64
	thirst    float64
	tiredness float64
	hitpoints float64
	Direction float64
	Speed     float64
}

func NewWorkerAnt(pos utils.Coordinate) WorkerAnt {
	return WorkerAnt{
		ID:        uuid.New(),
		Pos:       pos,
		hunger:    0.0,
		tiredness: 0.0,
		hitpoints: 100.0,
		Speed:     1.0,
		Direction: -1.0,
	}
}

func (wa *WorkerAnt) Move() {
	if wa.Direction < 0.0 {
		wa.Direction = rand.Float64() * 2 * math.Pi
	}
	dx := math.Cos(wa.Direction) * wa.Speed
	dy := math.Sin(wa.Direction) * wa.Speed
	wa.Pos.Add(dx, dy)
	deltaTheta := rand.Float64()*0.4 - 0.2 // small random change in direction
	wa.Direction += deltaTheta
	wa.Direction = math.Mod(wa.Direction, 2*math.Pi)
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
