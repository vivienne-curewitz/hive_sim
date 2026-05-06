package main

import (
	"image/color"
	"log"

	"hive_sim/src/sim"
	"hive_sim/src/world"

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

func DrawWorld(screen *ebiten.Image, w *world.World) {
	const zoom float32 = 10 // for now, each cell is 10 x 10 pixels
	var x, y float32 = 0, 0
	// first rectangle at 0,0 (top left, top right??)
	for i := 0; i < w.Length(); i += 1 {
		y = 0
		for j := 0; j < w.Height(); j += 1 {
			vector.FillRect(screen, x, y, zoom, zoom, w.GetColor(i, j), false)
			y += zoom
		}
		x += zoom
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// log.Printf("Worker ants: %d\n", len(g.HiveSim.WorkerAnts))
	// the silly way with vectors for now
	DrawWorld(screen, g.HiveSim.World)
	cameraX := screen.Bounds().Dx() / 2
	cameraY := screen.Bounds().Dy() / 2
	for _, ant := range g.HiveSim.WorkerAnts {
		vector.FillRect(screen, float32(cameraX)-float32(ant.Pos.X()), float32(cameraY)-float32(ant.Pos.Y()), 0.4, 0.4, color.White, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1000, 1000
}

func main() {
	hive_sim := sim.NewSimulation(0.1, 100)
	hive_sim.Init()

	ebiten.SetWindowSize(1000, 1000)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{HiveSim: hive_sim}); err != nil {
		log.Fatal(err)
	}
}
