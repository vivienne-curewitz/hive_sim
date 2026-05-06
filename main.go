package main

import (
	"image/color"
	"log"

	"hive_sim/src/sim"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	HiveSim *sim.Simulation
}

func (g *Game) Update() error {
	g.HiveSim.SingleStep()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// log.Printf("Worker ants: %d\n", len(g.HiveSim.WorkerAnts))
	// the silly way with vectors for now
	cameraX := screen.Bounds().Dx() / 2
	cameraY := screen.Bounds().Dy() / 2
	for _, ant := range g.HiveSim.WorkerAnts {
		vector.FillRect(screen, float32(cameraX)-float32(ant.Pos.X()), float32(cameraY)-float32(ant.Pos.Y()), 0.1, 0.1, color.White, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1920, 1080
}

func main() {
	hive_sim := sim.NewSimulation(0.1, 100)
	hive_sim.Init()

	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{HiveSim: hive_sim}); err != nil {
		log.Fatal(err)
	}
}
