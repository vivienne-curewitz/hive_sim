package ant

import (
	"math"
	"math/rand/v2"

	"hive_sim/src/pheremone"
	"hive_sim/src/utils"
	"hive_sim/src/world"

	"github.com/google/uuid"
)

const (
	ExhaustionRate float64 = 0.1
	MaxHunger      float64 = 100.0
	MaxThirst      float64 = 100.0
)

type Landmark struct {
	Position utils.Coordinate
	Type     pheremone.Pheremone
}

const (
	PheremoneFrequency float64 = 0.05      // chance to drop per second
	PheremoneLifetime  float64 = 300_000.0 // 5 minutes in ms
)

type Action int

const (
	RetrieveFood Action = iota
	DeliverFood
	ActionRest
	Wander
)

type Ant interface {
	Move(timeStepMs float64)
	Eat()
	Drink()
	Rest()
	ChooseAction(w *world.World, home utils.Coordinate)
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
	CurrentAction     Action
}

func NewWorkerAnt(pos utils.Coordinate, home utils.Coordinate) WorkerAnt {
	return WorkerAnt{
		ID:                uuid.New(),
		Pos:               pos,
		hunger:            0.0,
		tiredness:         0.0,
		hitpoints:         100.0,
		Speed:             1.0,
		Direction:         rand.Float64() * 2 * math.Pi,
		LastKnownLandmark: getHomeLandmark(home),
		Exhausted:         false,
		CurrentAction:     Wander,
	}
}

func getHomeLandmark(home utils.Coordinate) Landmark {
	return Landmark{
		Position: home,
		Type:     pheremone.PheremoneHome,
	}
}

func (wa *WorkerAnt) Step(timeStepMs float64, w *world.World) {
	wa.hunger += timeStepMs * ExhaustionRate
	wa.thirst += timeStepMs * ExhaustionRate
	wa.tiredness += timeStepMs * ExhaustionRate
	if wa.hunger > MaxHunger || wa.thirst > MaxThirst || wa.tiredness > MaxHunger {
		wa.Exhausted = true
		wa.FindBearings(w)
	}
	wa.Move(timeStepMs)
}

func (wa *WorkerAnt) FindBearings(w *world.World) {
	// look for best pheremone based on hunger or thirst, otherwise move towards home
	// currently only pheremone is home
	w.GetAveragePheremones(wa.Pos)
}

func (wa *WorkerAnt) SprayPheremone(currentTime float64) pheremone.PheremoneMark {
	newLm := Landmark{
		Position: wa.Pos,
		Type:     wa.LastKnownLandmark.Type,
	}
	ph := pheremone.PheremoneMark{
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
	deltaTheta := rand.Float64()*0.4 - 0.2 // small random change in direction
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

// set action and direction
func (wa *WorkerAnt) ChooseAction(w *world.World, home utils.Coordinate) {
	if wa.CurrentAction == Wander {
		// check for resource nearby
		res := w.GetNearbyResource(wa.Pos)
		fph, exists := w.GetAveragePheremones(wa.Pos)[pheremone.PheremoneFood]
		hph, hexists := w.GetAveragePheremones(wa.Pos)[pheremone.PheremoneHome]
		if res != nil {
			wa.CurrentAction = DeliverFood
			wa.Direction = math.Mod(wa.Direction+math.Pi, 2*math.Pi) // turn around to go back to home
			wa.LastKnownLandmark = Landmark{
				Position: res.Pos,
				Type:     pheremone.PheremoneFood,
			}
		} else if exists {
			wa.CurrentAction = RetrieveFood
			wa.Direction = fph.AverageDirection()
		} else if wa.Exhausted {
			wa.CurrentAction = ActionRest
			if hexists {
				wa.Direction = hph.AverageDirection()
			} else {
				wa.Direction = wa.Pos.AngleTo(wa.LastKnownLandmark.Position)
			}
		} else {
			wa.CurrentAction = Wander
		}
	} else if wa.CurrentAction == DeliverFood {
		delta := math.Sqrt(math.Pow(wa.Pos.X()-home.X(), 2.0) + math.Pow(wa.Pos.Y()-home.Y(), 2))
		if delta < 1.0 {
			// food is delivered
			wa.CurrentAction = Wander
			wa.LastKnownLandmark = Landmark{
				Position: home,
				Type:     pheremone.PheremoneHome,
			}
			wa.Direction = rand.Float64() * 2 * math.Pi
		} else {
			// orient again
			hph, hexists := w.GetAveragePheremones(wa.Pos)[pheremone.PheremoneHome]
			if hexists {
				wa.Direction = hph.AverageDirection()
				//wa.Direction += rand.Float64() * 2 * math.Pi
				//if wa.Direction < 0.0 {
				//	wa.Direction += 2 * math.Pi
				//}
			}

		}
	} else if wa.CurrentAction == RetrieveFood {
		res := w.GetNearbyResource(wa.Pos)
		if res != nil {
			// found food!!
			wa.CurrentAction = DeliverFood
			wa.Direction = math.Mod(wa.Direction+math.Pi, 2*math.Pi) // turn around to go back to home
			wa.LastKnownLandmark = Landmark{
				Position: res.Pos,
				Type:     pheremone.PheremoneFood,
			}
		} else {
			// reorient
			fph, exists := w.GetAveragePheremones(wa.Pos)[pheremone.PheremoneFood]
			if exists {
				wa.Direction = fph.AverageDirection()
				//wa.Direction += rand.Float64() * 2 * math.Pi
				//if wa.Direction < 0.0 {
				//	wa.Direction += 2 * math.Pi
				//}
			}
		}
	}

	// otherwise continue the current action
}
