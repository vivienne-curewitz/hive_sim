package main

import (
	"image/color"
	"log"

	"hive_sim/src/ant"
	"hive_sim/src/camera"
	"hive_sim/src/sim"
	"hive_sim/src/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	HiveSim   *sim.Simulation
	frametime float64
	Camera    *camera.Camera
}

func (g *Game) Update() error {
	g.HiveSim.SingleStep()
	return nil
}

func DrawPheremones(screen *ebiten.Image, pheremones []ant.PheremoneMark, w *world.World, cam *camera.Camera) {
	x_scale, y_scale := cam.GetScale(screen)
	bounds := cam.GetBounds()
	for _, ph := range pheremones {
		if !camera.InBounds(ph.Pos, bounds) {
			continue
		}
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
		px := (float32(ph.Pos.X()) - float32(bounds.Min.X)) * x_scale
		py := (float32(ph.Pos.Y()) - float32(bounds.Min.Y)) * y_scale
		vector.FillCircle(screen, px, py, 1, pcolor, false)
	}
}

func DrawWorld(screen *ebiten.Image, w *world.World, cam *camera.Camera) {
	x_scale, y_scale := cam.GetScale(screen)
	bounds := cam.GetBounds()
	var x, y float32 = 0, 0
	// first rectangle at 0,0 (top left, top right??)
	for i := bounds.Min.X; i < bounds.Max.X; i += 1 {
		for j := bounds.Min.X; j < bounds.Max.X; j += 1 {
			x = (float32(i) - float32(bounds.Min.X)) * x_scale
			y = (float32(j) - float32(bounds.Min.Y)) * y_scale
			vector.FillRect(screen, x, y, x_scale, y_scale, w.GetColor(i, j), false)
		}
	}
}

func DrawAnts(screen *ebiten.Image, ants []ant.WorkerAnt, w *world.World, cam *camera.Camera) {
	x_scale, y_scale := cam.GetScale(screen)
	bounds := cam.GetBounds()
	for _, ant := range ants {
		if !camera.InBounds(ant.Pos, bounds) {
			continue
		}
		var acolor color.RGBA
		switch {
		case ant.Exhausted:
			acolor = color.RGBA{255, 0, 0, 255}
		default:
			acolor = color.RGBA{255, 255, 255, 255}
		}
		px := (float32(ant.Pos.X()) - float32(bounds.Min.X)) * x_scale
		py := (float32(ant.Pos.Y()) - float32(bounds.Min.Y)) * y_scale
		vector.FillRect(screen, px, py, 1, 1, acolor, false)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// log.Printf("Worker ants: %d\n", len(g.HiveSim.WorkerAnts))
	// the silly way with vectors for now
	DrawWorld(screen, g.HiveSim.World, g.Camera)
	// draw pheremones
	DrawPheremones(screen, g.HiveSim.World.GetPheremones(), g.HiveSim.World, g.Camera)
	// draw ants
	DrawAnts(screen, g.HiveSim.WorkerAnts, g.HiveSim.World, g.Camera)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1000, 1000
}

func main() {
	tps := 60
	hive_sim := sim.NewSimulation(1.0/float64(tps), 100)
	hive_sim.Init()
	cam := camera.NewCamera(hive_sim.World, nil, 1.0)
	game := Game{HiveSim: hive_sim, frametime: 1.0 / float64(tps), Camera: &cam}

	ebiten.SetWindowSize(1000, 1000)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetTPS(tps) // 60 FPS
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
