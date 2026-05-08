package main

import (
	"image/color"
	"log"

	"hive_sim/src/ant"
	"hive_sim/src/sim"
	"hive_sim/src/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	HiveSim   *sim.Simulation
	frametime float64
}

func (g *Game) Update() error {
	g.HiveSim.SingleStep()
	return nil
}

func DrawPheremones(screen *ebiten.Image, pheremones []ant.PheremoneMark) {
	cameraX := screen.Bounds().Dx() / 2
	cameraY := screen.Bounds().Dy() / 2
	for _, ph := range pheremones {
		var pcolor color.RGBA
		switch ph.Type {
		case ant.PheremoneHome:
			pcolor = color.RGBA{199, 25, 224, 255}
		case ant.PheremoneFood:
			pcolor = color.RGBA{255, 255, 0, 128}
		case ant.PheremonePath:
			pcolor = color.RGBA{0, 255, 255, 128}
		case ant.PheremoneDeath:
			pcolor = color.RGBA{255, 0, 255, 128}
		}
		vector.FillCircle(screen, float32(cameraX)-float32(ph.Pos.X()), float32(cameraY)-float32(ph.Pos.Y()), 1, pcolor, false)
	}
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

func DrawAnts(screen *ebiten.Image, ants []ant.WorkerAnt) {
	cameraX := screen.Bounds().Dx() / 2
	cameraY := screen.Bounds().Dy() / 2
	for _, ant := range ants {
		var acolor color.RGBA
		switch {
		case ant.Exhausted:
			acolor = color.RGBA{255, 0, 0, 255}
		default:
			acolor = color.RGBA{255, 255, 255, 255}
		}
		vector.FillRect(screen, float32(cameraX)-float32(ant.Pos.X()), float32(cameraY)-float32(ant.Pos.Y()), 1, 1, acolor, false)
	}
}
func (g *Game) Draw(screen *ebiten.Image) {
	// log.Printf("Worker ants: %d\n", len(g.HiveSim.WorkerAnts))
	// the silly way with vectors for now
	DrawWorld(screen, g.HiveSim.World)
	// draw pheremones
	DrawPheremones(screen, g.HiveSim.World.GetPheremones())
	// draw ants
	DrawAnts(screen, g.HiveSim.WorkerAnts)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1000, 1000
}

func main() {
	tps := 60
	hive_sim := sim.NewSimulation(1.0, 100)
	hive_sim.Init()

	ebiten.SetWindowSize(1000, 1000)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetTPS(tps) // 60 FPS
	if err := ebiten.RunGame(&Game{HiveSim: hive_sim, frametime: 1.0 / float64(tps)}); err != nil {
		log.Fatal(err)
	}
}
