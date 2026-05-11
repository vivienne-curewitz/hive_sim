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
	PheremoneFrequency float64 = 1.0      // chance to drop per second
	PheremoneLifetime  float64 = 10_000.0 // 5 minutes in ms
)

type Action int

const (
	FindFood Action = iota
	RetrieveFood
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

type phDirCache struct {
	Direction float64
	Type      pheremone.Pheremone
	Location  utils.Coordinate
	hits      int
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
	DirCache          phDirCache
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
	//mult := 0.0
	//switch wa.LastKnownLandmark.Type {
	//case pheremone.PheremoneFood:
	//	mult = 1
	//case pheremone.PheremoneHome:
	//	mult = 10
	//}
	ph := pheremone.PheremoneMark{
		Type:       wa.LastKnownLandmark.Type,
		Pos:        wa.Pos,
		Strength:   1.0 / wa.Pos.DistanceTo(wa.LastKnownLandmark.Position),
		Expiration: PheremoneLifetime + currentTime,
	}
	return ph
}

// add +- ent radians to the direction
func (wa *WorkerAnt) DirectionEntropy(ent float64) {
	deltaTheta := rand.Float64()*2*ent - ent // small random change in direction
	wa.Direction += deltaTheta
	wa.Direction = math.Mod(wa.Direction, 2*math.Pi)
}

func (wa *WorkerAnt) Move(timeStepMs float64) {
	dx := math.Cos(wa.Direction) * wa.Speed * timeStepMs
	dy := math.Sin(wa.Direction) * wa.Speed * timeStepMs
	wa.Pos.Add(dx, dy)
	wa.DirectionEntropy(0.02)
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

func (wa *WorkerAnt) foundFood(res *world.FoodSource, home utils.Coordinate) {
	wa.CurrentAction = RetrieveFood
	wa.Direction = wa.Pos.AngleTo(home) // kind of know where to go
	res.TakeX(5)
	wa.LastKnownLandmark = Landmark{
		Position: res.Pos,
		Type:     pheremone.PheremoneFood,
	}
}

func (wa *WorkerAnt) setPhDirection(w *world.World, ptype pheremone.Pheremone) bool {
	if wa.DirCache.Type == ptype && wa.DirCache.hits < 10 {
		if wa.Pos.EqualsModStep(wa.DirCache.Location, world.PheremoneIndexPerCell) {
			// cache hit
			// wa.Direction = wa.Direction*0.9 + wa.DirCache.Direction*0.1
			wa.Direction = wa.DirCache.Direction
			wa.DirCache.hits += 1
			return true
		}
	}
	newDir, exists := w.GetPheremoneDirection(wa.Pos, ptype)
	if exists {
		// update wa.DirCache and wa.Direction
		wa.DirCache.Type = ptype
		wa.DirCache.Direction = newDir
		wa.DirCache.Location = wa.Pos
		wa.DirCache.hits = 0
		wa.Direction = newDir
		// wa.Direction = wa.Direction*0.9 + newDir*0.1
		return true
	}
	return false
}

// set action and direction
func (wa *WorkerAnt) ChooseAction(w *world.World, home utils.Coordinate) {
	if wa.CurrentAction == Wander {
		// check for resource nearby
		res := w.GetNearbyResource(wa.Pos)
		if res != nil && wa.Pos.DistanceTo(res.Pos) < res.Radius {
			wa.foundFood(res, home)
			return
		} else if res != nil {
			wa.CurrentAction = FindFood
			return
		} else {
			realDirection := wa.setPhDirection(w, pheremone.PheremoneFood)
			if realDirection {
				wa.CurrentAction = FindFood
				return
			}
		}
		wa.DirectionEntropy(0.2)
		return
	} else if wa.CurrentAction == RetrieveFood {
		delta := wa.Pos.DistanceTo(home)
		if delta < 1.0 {
			// food is delivered
			wa.CurrentAction = Wander
			wa.LastKnownLandmark = Landmark{
				Position: home,
				Type:     pheremone.PheremoneHome,
			}
			wa.Direction = rand.Float64() * 2 * math.Pi
			return
		} else {
			// orient again
			if wa.Pos.DistanceTo(home) < 10.0 { // the ant can just see home
				wa.Direction = wa.Pos.AngleTo(home)
				wa.DirectionEntropy(0.002)
				return
			} else {
				hexists := wa.setPhDirection(w, pheremone.PheremoneHome)
				if hexists {
					wa.DirectionEntropy(0.002)
				}
				return
			}
		}
	} else if wa.CurrentAction == FindFood {
		res := w.GetNearbyResource(wa.Pos)
		if res != nil && wa.Pos.DistanceTo(res.Pos) < res.Radius {
			// found food!!
			wa.foundFood(res, home)
			return
		} else if res != nil {
			wa.Direction = wa.Pos.AngleTo(res.Pos)
			return
		} else {
			// reorient
			exists := wa.setPhDirection(w, pheremone.PheremoneFood)
			if exists {
				wa.DirectionEntropy(0.002)
			} else {
				wa.CurrentAction = Wander
			}
			return
		}
	}
}
