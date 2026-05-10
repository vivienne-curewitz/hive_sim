package main

import (
	"image/color"
	"log"
	"math"
	"runtime"
	"sync"

	"hive_sim/src/ant"
	"hive_sim/src/camera"
	ph "hive_sim/src/pheremone"
	"hive_sim/src/sim"
	"hive_sim/src/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	HiveSim        *sim.Simulation
	frametime      float64
	Camera         *camera.Camera
	CursorPosition CursorPos
}

type CursorPos struct {
	X int
	Y int
}

func CameraControl(cam *camera.Camera, prevCursorPos *CursorPos) {
	// case zoom in/out with mouse wheel
	_, dy := ebiten.Wheel()
	// log.Printf("Wheel: dx=%f, dy=%f\n", dx, dy)
	if dy != 0 {
		cam.Zoom(10.0 * dy) // Pan vertically
		// time.Sleep(500 * time.Microsecond)
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		dx := float64(prevCursorPos.X - cx)
		dy = float64(prevCursorPos.Y - cy)
		xs, ys := cam.GetScale()
		cam.Move(dx/float64(xs), dy/float64(ys)) // dx is in pixels, xs is pixels/world_unit, dx/xs is pixels/pixels/world_unit = pixels*(world_unit/pixels) = world_unit
	}
	prevCursorPos.X, prevCursorPos.Y = ebiten.CursorPosition()
	// case click and drag
}

func (g *Game) Update() error {
	go CameraControl(g.Camera, &g.CursorPosition)
	g.HiveSim.SingleStep()
	return nil
}

type PhDrawInfo struct {
	PColor color.RGBA
	X      float32
	Y      float32
}

func DrawPheremones(screen *ebiten.Image, w *world.World, cam *camera.Camera) {
	x_scale, y_scale := cam.GetScale()
	bounds := cam.GetBounds()

	const phIdx = 5 // should match pheremoneIndexPerCell in world.go
	step := 1.0 / float64(phIdx)

	// Calculate visible range in the pheremone grid
	minI := int(math.Max(0, bounds.Min.X()*phIdx))
	maxI := int(math.Min(float64(len(w.AveragePheremoneCell)), bounds.Max.X()*phIdx))
	minJ := int(math.Max(0, bounds.Min.Y()*phIdx))
	maxJ := int(math.Min(float64(len(w.AveragePheremoneCell[0])), bounds.Max.Y()*phIdx))

	for i := minI; i < maxI; i++ {
		for j := minJ; j < maxJ; j++ {
			mp := w.AveragePheremoneCell[i][j]
			if len(mp) == 0 {
				continue
			}

			// Implicit coordinate of this pheremone cell
			worldX := float64(i) * step
			worldY := float64(j) * step

			px := (float32(worldX) - float32(bounds.Min.X())) * x_scale
			py := (float32(worldY) - float32(bounds.Min.Y())) * y_scale

			for phType, avgPh := range mp {
				strength := avgPh.Strength()
				if strength <= 0 {
					continue
				}

				var pcolor color.RGBA
				switch phType {
				case ph.PheremoneHome:
					pcolor = color.RGBA{199, 25, 224, 255}
				case ph.PheremoneFood:
					pcolor = color.RGBA{255, 255, 0, 128}
				case ph.PheremonePath:
					pcolor = color.RGBA{0, 255, 255, 128}
				case ph.PheremoneDeath:
					pcolor = color.RGBA{255, 0, 255, 128}
				}
				// Scale radius by strength (clamped for visibility)
				radius := float32(math.Max(1.0, math.Min(4.0, 3*strength*2.0)))
				vector.FillCircle(screen, px, py, radius, pcolor, false)
			}
		}
	}
}

func DrawWorld(screen *ebiten.Image, w *world.World, cam *camera.Camera) {
	x_scale, y_scale := cam.GetScale()
	bounds := cam.GetBounds()
	var x, y float32 = 0, 0
	// first rectangle at 0,0 (top left, top right??)
	for i := int(math.Max(float64(bounds.Min.X()), 0)); i < int(math.Min(float64(bounds.Max.X()), float64(w.Length()))); i += 1 {
		for j := int(math.Max(float64(bounds.Min.Y()), 0)); j < int(math.Min(float64(bounds.Max.Y()), float64(w.Height()))); j += 1 {
			x = (float32(i) - float32(bounds.Min.X())) * x_scale
			y = (float32(j) - float32(bounds.Min.Y())) * y_scale
			vector.FillRect(screen, x, y, x_scale, y_scale, w.GetColor(i, j), false)
		}
	}

	for _, res := range w.Resources {
		if !camera.InBounds(res.Pos, bounds) {
			continue
		}
		img := w.Images[int(res.Type)]
		ibounds := img.Bounds()
		dx := ibounds.Dx()
		ixscale := float64(x_scale / float32(dx))
		dy := ibounds.Dy()
		iyscale := float64(y_scale / float32(dy))
		px := (res.Pos.X() - bounds.Min.X() - 0.5) * float64(x_scale)
		py := (res.Pos.Y() - bounds.Min.Y() - 0.5) * float64(y_scale)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(ixscale, iyscale)
		op.GeoM.Translate(px, py)
		screen.DrawImage(img, op)

		// Draw health bar
		barWidth := float32(x_scale)
		barHeight := float32(y_scale * 0.1)
		barX := float32(px)
		barY := float32(py) + float32(y_scale)

		// Background
		vector.FillRect(screen, barX, barY, barWidth, barHeight, color.RGBA{50, 50, 50, 255}, false)

		// Foreground
		percent := float32(res.Amount.Load()) / float32(res.MaxAmount)
		if percent > 0 {
			vector.FillRect(screen, barX, barY, barWidth*percent, barHeight, color.RGBA{0, 255, 0, 255}, false)
		}

	}
}

func DrawAnts(screen *ebiten.Image, ants []ant.WorkerAnt, w *world.World, cam *camera.Camera) {
	cpus := runtime.NumCPU() / 2
	x_scale, y_scale := cam.GetScale()
	bounds := cam.GetBounds()
	antDrawInfos := make([]PhDrawInfo, len(ants))
	wg := sync.WaitGroup{}
	wg.Add(cpus)
	for cpuI := range cpus {
		go func() {
			start := cpuI * (len(ants)/cpus + 1)
			end := (cpuI + 1) * (len(ants)/cpus + 1)
			if end > len(ants) {
				end = len(ants)
			}
			for ai, ant := range ants[start:end] {
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
				px := (float32(ant.Pos.X()) - float32(bounds.Min.X())) * x_scale
				py := (float32(ant.Pos.Y()) - float32(bounds.Min.Y())) * y_scale
				antDrawInfos[ai+start] = PhDrawInfo{PColor: acolor, X: px, Y: py}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for _, adi := range antDrawInfos {
		vector.FillRect(screen, adi.X, adi.Y, 1, 1, adi.PColor, false)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Camera.SetScale(screen)
	// log.Printf("Worker ants: %d\n", len(g.HiveSim.WorkerAnts))
	// the silly way with vectors for now
	DrawWorld(screen, g.HiveSim.World, g.Camera)
	// draw pheremones
	DrawPheremones(screen, g.HiveSim.World, g.Camera)
	// draw ants
	DrawAnts(screen, g.HiveSim.WorkerAnts, g.HiveSim.World, g.Camera)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func main() {
	tps := 60
	hive_sim := sim.NewSimulation(1.0/float64(tps), 100)
	hive_sim.Init()
	cam := camera.NewCamera(hive_sim.World, nil, 2.0)
	game := Game{HiveSim: hive_sim, frametime: 1.0 / float64(tps), Camera: &cam}

	ebiten.SetWindowSize(2000, int(2000.0*9.0/16.0))
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetTPS(tps) // 60 FPS
	// go CameraControl(&cam)
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
