package ant

import (
	"math"
	"math/rand/v2"

	"hive_sim/src/utils"

	"github.com/google/uuid"
)

const (
	ExhaustionRate float64 = 0.1
	MaxHunger      float64 = 100.0
	MaxThirst      float64 = 100.0
)

type Landmark struct {
	Position utils.Coordinate
	Type     Pheremone
}

const (
	PheremoneFrequency float64 = 0.05      // chance to drop per second
	PheremoneLifetime  float64 = 300_000.0 // 5 minutes in ms
)

type Ant interface {
	Move(timeStepMs float64)
	Eat()
	Drink()
	Rest()
	ChooseAction()
}

type WorkerAnt struct {
	ID                uuid.UUID
	Pos               utils.Coordinate
	hunger            float64
	thirst            float64
	tiredness         float64
	hitpoints         float64
	Direction         float64
	Speed             float64
	LastKnownLandmark Landmark
	Exhausted         bool
}

func NewWorkerAnt(pos utils.Coordinate) WorkerAnt {
	return WorkerAnt{
		ID:                uuid.New(),
		Pos:               pos,
		hunger:            0.0,
		tiredness:         0.0,
		hitpoints:         100.0,
		Speed:             1.0,
		Direction:         rand.Float64() * 2 * math.Pi,
		LastKnownLandmark: getHomeLandmark(),
		Exhausted:         false,
	}
}

func getHomeLandmark() Landmark {
	return Landmark{
		Position: utils.NewCoordinate(0, 0),
		Type:     PheremoneHome,
	}
}

func (wa *WorkerAnt) Step(timeStepMs float64) {
	wa.hunger += timeStepMs * ExhaustionRate
	wa.thirst += timeStepMs * ExhaustionRate
	wa.tiredness += timeStepMs * ExhaustionRate
	if wa.hunger > MaxHunger || wa.thirst > MaxThirst || wa.tiredness > MaxHunger {
		wa.Exhausted = true
		wa.FindBearings()
	}
	wa.Move(timeStepMs)
}

func (wa *WorkerAnt) FindBearings() {
	// look for best pheremone based on hunger or thirst, otherwise move towards home
	// currently only pheremone is home
	homeDir := wa.Pos.AngleTo(utils.NewCoordinate(0, 0))
	wa.Direction = homeDir
}

func (wa *WorkerAnt) SprayPheremone(currentTime float64) PheremoneMark {
	newLm := Landmark{
		Position: wa.Pos,
		Type:     wa.LastKnownLandmark.Type,
	}
	ph := PheremoneMark{
		Type:       newLm.Type,
		Pos:        wa.Pos,
		Direction:  wa.Pos.AngleTo(wa.LastKnownLandmark.Position),
		Expiration: PheremoneLifetime + currentTime,
	}
	wa.LastKnownLandmark = newLm
	return ph
}

func (wa *WorkerAnt) Move(timeStepMs float64) {
	dx := math.Cos(wa.Direction) * wa.Speed * timeStepMs
	dy := math.Sin(wa.Direction) * wa.Speed * timeStepMs
	wa.Pos.Add(dx, dy)
	deltaTheta := rand.Float64()*0.8 - 0.4 // small random change in direction
	wa.Direction += deltaTheta
	wa.Direction = math.Mod(wa.Direction, 2*math.Pi)
	if wa.Direction < 0 {
		wa.Direction += 2 * math.Pi
	}
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
	wa.Move(16.67) // to do make this a real function
}
