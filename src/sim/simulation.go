package sim

import (
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"hive_sim/src/ant"
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
	WorkerAnts  []ant.WorkerAnt
	Resources   map[uuid.UUID]*world.Resource
	World       *world.World
	StepCount   int
	StepTimeSum int
	Pheremones  []ant.PheremoneMark
}

func NewSimulation(timeStep, duration float64) *Simulation {
	return &Simulation{
		TimeStep:    timeStep,
		Duration:    duration,
		CurrentTime: 0,
		Ants:        make(map[uuid.UUID]ant.Ant),
		Resources:   make(map[uuid.UUID]*world.Resource),
		World:       world.NewWorld(100, 100), // Example world size
	}
}

func (s *Simulation) Init() {
	s.WorkerAnts = make([]ant.WorkerAnt, 10000)
	for i := range 10000 {
		cx := rand.Float64() * 10
		cy := rand.Float64() * 10
		wa := ant.NewWorkerAnt(utils.NewCoordinate(cx, cy))
		s.WorkerAnts[i] = wa
		s.Ants[wa.ID] = &s.WorkerAnts[i]
	}
}

func (s *Simulation) SingleStep(timeStepMs float64) {
	// start go routines
	startTime := time.Now().UnixMicro()
	if true {
		processors := 4
		wg := sync.WaitGroup{}
		wg.Add(processors)
		for pi := range processors {
			go func() {
				startInd := len(s.WorkerAnts) / processors * pi
				for startInd < len(s.WorkerAnts)/processors*(pi+1) {
					ant := &s.WorkerAnts[startInd]
					ant.Step(timeStepMs)
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
			ant.Move(timeStepMs)
		}
	}
	s.StepCount++
	s.StepTimeSum += int(time.Now().UnixMicro() - startTime)
	if s.StepCount%60 == 0 {
		log.Printf("Phase 1 average %d microseconds\n", s.StepTimeSum/60)
		s.StepTimeSum = 0
	}
	// then, phase 2 -- interactions with other -- attack other ant, take resource, etc
	// pheremones first
	sprayAnts := rand.Perm(len(s.WorkerAnts))[:len(s.WorkerAnts)*int(ant.PheremoneFrequency)]
	for _, ant := range sprayAnts {
		ph := s.WorkerAnts[ant].SprayPheremone(timeStepMs)
		s.World.AddPheremone(ph)
	}
	// phase 3 -- resolution

	// update simulation
	s.CurrentTime += s.TimeStep
}
