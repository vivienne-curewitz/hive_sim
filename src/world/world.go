package world

import (
	"image/color"
	"math/rand/v2"
	"time"

	"hive_sim/src/ant"
)

type Cell int

const (
	Dirt Cell = iota
	Grass
	Flower
)

type AveragePheremone struct {
	Mark  ant.PheremoneMark
	Count int
}

type World struct {
	height               int
	length               int
	time                 int32
	Cells                [][]Cell
	Pheremones           []ant.PheremoneMark
	AveragePheremoneCell [][]map[ant.Pheremone]AveragePheremone
	LastPheremoneIndex   int
	FirstValidPheremone  int
}

func NewWorld(height, length int) *World {
	w := &World{
		height:              height,
		length:              length,
		time:                0,
		LastPheremoneIndex:  -1,
		FirstValidPheremone: 0,
	}
	w.Init()
	return w
}

func (w *World) Length() int {
	return w.length
}

func (w *World) Height() int {
	return w.height
}

func (w *World) GetColor(i, j int) color.Color {
	if w.Cells[i][j] == Dirt {
		return color.RGBA{
			153, 115, 0, 255,
		}
	} else {
		return color.RGBA{
			0, 153, 51, 255,
		}
	}
}

func (w *World) getSurroundingCount(i int, j int) (float64, float64) {
	var dirt float64 = 0
	var grass float64 = 0
	if i < 1 || j < 1 || i > w.length-2 || j > w.length-2 {
		return 0, 0
	}
	for x := -1; x < 2; x += 1 {
		for y := -1; y < 2; y += 1 {
			if x == 1 && y == 1 {
				continue
			}
			switch w.Cells[i+x][i+y] {
			case Dirt:
				dirt += 1
			case Grass:
				grass += 1
			}
		}
	}
	return dirt, grass
}

func (w *World) Init() {
	// initialize world state here
	var seed uint64 = uint64(time.Now().UnixMilli())
	rs := rand.New(rand.NewPCG(seed, seed))
	const genRange float64 = 12
	w.Cells = make([][]Cell, w.length)
	for i := 0; i < w.length; i += 1 {
		w.Cells[i] = make([]Cell, w.height)
	}
	// assume height and length in meters
	for i := 0; i < w.length; i += 1 {
		for j := 0; j < w.height; j += 1 {
			r1 := rs.Float64() * 100
			if r1 > 95 {
				w.Cells[i][j] = Flower
			} else {
				d, g := w.getSurroundingCount(i, j)
				rd := rs.Float64()*genRange + d
				rg := rs.Float64()*genRange + g
				if rd > rg {
					w.Cells[i][j] = Dirt
				} else {
					w.Cells[i][j] = Grass
				}
			}

		}
	}
	// init Pheremones
	w.Pheremones = make([]ant.PheremoneMark, 1_000_000)
	w.AveragePheremoneCell = make([][]map[ant.Pheremone]AveragePheremone, w.length)
	for i := 0; i < w.length; i++ {
		w.AveragePheremoneCell[i] = make([]map[ant.Pheremone]AveragePheremone, w.height)
	}
}

func (w *World) CullPheremones(currentTime float64) {
	// binary search for most recently expired pheremone
	for i := w.FirstValidPheremone; i != w.LastPheremoneIndex; i = (i + 1) % len(w.Pheremones) {
		if w.Pheremones[i].Expiration < currentTime {
			w.FirstValidPheremone = (i + 1) % len(w.Pheremones)
		} else {
			break
		}
	}
}

func (w *World) AddPheremone(ph ant.PheremoneMark) {
	w.LastPheremoneIndex = (w.LastPheremoneIndex + 1) % len(w.Pheremones)
	w.Pheremones[w.LastPheremoneIndex] = ph
	// add to Pheremone cell
	avgPh, exists := w.AveragePheremoneCell[int(ph.Pos.X())][int(ph.Pos.Y())][ph.Type]
	if !exists {
		avgPh = AveragePheremone{Mark: ph, Count: 1}
	}
	avgPh.Count++
	w.AveragePheremoneCell[int(ph.Pos.X())][int(ph.Pos.Y())][ph.Type] = avgPh
}

func (w *World) GetPheremones() []ant.PheremoneMark {
	if w.LastPheremoneIndex >= w.FirstValidPheremone {
		return w.Pheremones[w.FirstValidPheremone : w.LastPheremoneIndex+1]
	} else {
		return append(w.Pheremones[w.FirstValidPheremone:], w.Pheremones[:w.LastPheremoneIndex+1]...)
	}
}
