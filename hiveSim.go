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

func DrawPheremones(screen *ebiten.Image, pheremones []ph.PheremoneMark, w *world.World, cam *camera.Camera) {
	cpus := runtime.NumCPU() / 2
	x_scale, y_scale := cam.GetScale()
	bounds := cam.GetBounds()
	drawInfos := make([]PhDrawInfo, len(pheremones))
	wg := sync.WaitGroup{}
	wg.Add(cpus)
	for cpuI := range cpus {
		// multithreading here causes the bad bad
		go func() {
			start := cpuI * (len(pheremones)/cpus + 1)
			end := (cpuI + 1) * (len(pheremones)/cpus + 1)
			if end > len(pheremones) {
				end = len(pheremones)
			}
			for phi := start; phi < end; phi += 1 {
				phm := pheremones[phi]
				if !camera.InBounds(phm.Pos, bounds) {
					continue
				}
				var pcolor color.RGBA
				switch phm.Type {
				case ph.PheremoneHome:
					pcolor = color.RGBA{199, 25, 224, 255}
				case ph.PheremoneFood:
					pcolor = color.RGBA{255, 255, 0, 128}
				case ph.PheremonePath:
					pcolor = color.RGBA{0, 255, 255, 128}
				case ph.PheremoneDeath:
					pcolor = color.RGBA{255, 0, 255, 128}
				}
				px := (float32(phm.Pos.X()) - float32(bounds.Min.X())) * x_scale
				py := (float32(phm.Pos.Y()) - float32(bounds.Min.Y())) * y_scale
				drawInfos[phi] = PhDrawInfo{PColor: pcolor, X: px, Y: py}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	for _, di := range drawInfos {
		if di.X == 0 && di.Y == 0 {
			continue
		}
		vector.FillCircle(screen, di.X, di.Y, 1, di.PColor, false)
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
	DrawPheremones(screen, g.HiveSim.World.GetPheremones(), g.HiveSim.World, g.Camera)
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
