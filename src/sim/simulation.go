package sim

import (
	"log"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"hive_sim/src/ant"
	"hive_sim/src/pheremone"
	"hive_sim/src/utils"
	"hive_sim/src/world"

	"github.com/google/uuid"
)

type Simulation struct {
	// Simulation parameters
	TimeStep float64
	Duration float64

	// Simulation state
	CurrentTime float64
	// fast ant lookup if necessary
	Ants map[uuid.UUID]ant.Ant
	// where the ants live in memory
	WorkerAnts []ant.WorkerAnt
	Resources  map[uuid.UUID]*world.Resource
	World      *world.World

	Pheremones []pheremone.PheremoneMark
	// bench mark stuff
	StepCount        int
	StepTimeSum      int
	ActionTimeSum    int
	PheremoneTimeSum int
	ResolveTimeSum   int
}

func NewSimulation(timeStep, duration float64) *Simulation {
	return &Simulation{
		TimeStep:    timeStep,
		Duration:    duration,
		CurrentTime: 0,
		Ants:        make(map[uuid.UUID]ant.Ant),
		Resources:   make(map[uuid.UUID]*world.Resource),
		World:       world.NewWorld(500, 500), // Example world size
	}
}

func (s *Simulation) Init() {
	s.WorkerAnts = make([]ant.WorkerAnt, 10000)
	cx := float64(s.World.Length() / 2)
	cy := float64(s.World.Height() / 2)
	home := utils.NewCoordinate(cx, cy)
	for i := range 10000 {
		cx := rand.Float64() - 0.5 + cx
		cy := rand.Float64() - 0.5 + cy
		wa := ant.NewWorkerAnt(utils.NewCoordinate(cx, cy), home)
		s.WorkerAnts[i] = wa
		s.Ants[wa.ID] = &s.WorkerAnts[i]
	}
}

func (s *Simulation) SingleStep() {
	// hack
	home := utils.NewCoordinate(float64(s.World.Length()/2), float64(s.World.Height()/2))
	// start go routines
	startTime := time.Now().UnixMicro()
	processors := runtime.NumCPU()
	wg := sync.WaitGroup{}
	if true {
		wg.Add(processors)
		for pi := range processors {
			go func() {
				startInd := len(s.WorkerAnts) / processors * pi
				end := len(s.WorkerAnts) / processors * (pi + 1)
				if end > len(s.WorkerAnts) {
					end = len(s.WorkerAnts)
				}
				for startInd < end {
					ant := &s.WorkerAnts[startInd]
					ant.Step(s.TimeStep, s.World)
					startInd++
				}
				wg.Done()
			}()
		}
		// first, phase 1 -- mostly movement
		wg.Wait()
	} else {
		// first, phase 1 -- mostly movement
		for i := range len(s.WorkerAnts) {
			ant := &s.WorkerAnts[i]
			ant.Move(s.TimeStep)
		}
	}
	s.StepTimeSum += int(time.Now().UnixMicro() - startTime)

	// then, phase 2 -- interactions with other -- attack other ant, take resource, etc
	actionStartTime := time.Now().UnixMicro()
	maxFoodForceSpray := int(float64(len(s.WorkerAnts))*ant.PheremoneFrequency*s.TimeStep) / 2
	foodAnts := make([]*ant.WorkerAnt, maxFoodForceSpray)
	for j := range processors {
		wg.Add(1)
		go func(pIndex int) {
			startInd := len(s.WorkerAnts) / processors * pIndex
			end := len(s.WorkerAnts) / processors * (pIndex + 1)
			fsa := pIndex
			if end > len(s.WorkerAnts) {
				end = len(s.WorkerAnts)
			}
			for i := startInd; i < end; i += 1 {
				cant := &s.WorkerAnts[i]
				cant.ChooseAction(s.World, home)
				if cant.CurrentAction == ant.RetrieveFood {
					if fsa < maxFoodForceSpray {
						foodAnts[fsa] = cant
						fsa += processors
					}
				}
			}
			wg.Done()
		}(j)
	}
	wg.Wait()
	s.ActionTimeSum += int(time.Now().UnixMicro() - actionStartTime)

	// pheremones first
	phStartTime := time.Now().UnixMicro()
	s.World.CullPheremones(s.CurrentTime)
	num_to_spray := int(float64(len(s.WorkerAnts)) * ant.PheremoneFrequency * s.TimeStep)
	sprayAnts := rand.Perm(len(s.WorkerAnts))[:num_to_spray]
	for _, ant := range sprayAnts {
		ph := s.WorkerAnts[ant].SprayPheremone(s.CurrentTime)
		s.World.AddPheremone(ph)
	}
	nilCount := 0
	for _, cat := range foodAnts {
		if cat != nil {
			nilCount = 0
			ph := cat.SprayPheremone(s.CurrentTime)
			s.World.AddPheremone(ph)
		} else {
			nilCount++
			if nilCount > processors {
				break
			}
		}
	}
	s.PheremoneTimeSum += int(time.Now().UnixMicro() - phStartTime)

	// phase 3 -- resolution
	resolveStartTime := time.Now().UnixMicro()
	// remove depleted food sources
	for fi := 0; fi < len(s.World.Resources); fi += 1 {
		food := &s.World.Resources[fi]
		if food.Amount.Load() <= 0 {
			log.Printf("Food source %d depleted\n", fi)
			// remove from world
			s.World.FoodSourceCells[int(food.Pos.X())][int(food.Pos.Y())] = nil
			food.Pos = utils.RandomCoordinate(float64(s.World.Length()), float64(s.World.Height()))
			s.World.FoodSourceCells[int(food.Pos.X())][int(food.Pos.Y())] = food
			food.Amount.Store(food.MaxAmount)
		}
	}
	s.ResolveTimeSum += int(time.Now().UnixMicro() - resolveStartTime)

	s.StepCount++
	if s.StepCount%60 == 0 {
		log.Printf("Averages (µs) | Ph1: %d | Ph2: %d | Pher: %d | Res: %d | Total: %d\n",
			s.StepTimeSum/60, s.ActionTimeSum/60, s.PheremoneTimeSum/60, s.ResolveTimeSum/60, (s.StepTimeSum+s.ActionTimeSum+s.PheremoneTimeSum+s.ResolveTimeSum)/60)
		s.StepTimeSum = 0
		s.ActionTimeSum = 0
		s.PheremoneTimeSum = 0
		s.ResolveTimeSum = 0
	}

	// update simulation
	s.CurrentTime += s.TimeStep * 1000
}
