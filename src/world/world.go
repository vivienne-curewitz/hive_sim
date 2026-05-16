package world

import (
	"image/color"
	"log"
	"math"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"hive_sim/src/pheremone"
	"hive_sim/src/utils"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Cell int

const PheremoneIndexPerCell = 1.0

const (
	Dirt Cell = iota
	Grass
	Flower
)

// ALL WORLD COORDINATES ARE POSITIVE
type World struct {
	height               int
	length               int
	time                 int32
	Cells                [][]Cell
	Pheremones           []pheremone.PheremoneMark
	AveragePheremoneCell [][]map[pheremone.Pheremone]pheremone.AveragePheremone
	LastPheremoneIndex   int
	FirstValidPheremone  int
	Resources            []FoodSource
	FoodSourceCells      [][]*FoodSource
	Images               []*ebiten.Image
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

func (w *World) initImages() {
	var err error
	beetle_img, _, err := ebitenutil.NewImageFromFile("images/beetle.png")
	if err != nil {
		log.Fatal(err)
	}
	w.Images = make([]*ebiten.Image, 2)
	w.Images[int(beetle)] = beetle_img
	flower_img, _, err := ebitenutil.NewImageFromFile("images/YellowFlower.png")
	if err != nil {
		log.Fatal(err)
	}
	w.Images[int(flower)] = flower_img
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
	home := utils.NewCoordinate(float64(w.length), float64(w.height))
	pheremoneCircleDistance := 10.0
	w.Pheremones = make([]pheremone.PheremoneMark, 20_000_000)
	w.AveragePheremoneCell = make([][]map[pheremone.Pheremone]pheremone.AveragePheremone, PheremoneIndexPerCell*w.length)
	for i := 0; i < w.length*PheremoneIndexPerCell; i += 1 {
		w.AveragePheremoneCell[i] = make([]map[pheremone.Pheremone]pheremone.AveragePheremone, PheremoneIndexPerCell*w.height)
		for j := 0; j < w.height*PheremoneIndexPerCell; j += 1 {
			w.AveragePheremoneCell[i][j] = make(map[pheremone.Pheremone]pheremone.AveragePheremone)
			dist := utils.NewCoordinate(float64(i)/PheremoneIndexPerCell, float64(j)/PheremoneIndexPerCell).DistanceTo(home)
			if dist < pheremoneCircleDistance {
				mark := pheremone.PheremoneMark{
					Type:     pheremone.PheremoneHome,
					Pos:      utils.NewCoordinate(float64(i)/PheremoneIndexPerCell, float64(j)/PheremoneIndexPerCell),
					Strength: 1.0 / dist,
				}
				w.AveragePheremoneCell[i][j][pheremone.PheremoneHome] = pheremone.NewAveragePheremone(mark)
			}
		}
	}
	// init Food Sources
	w.Resources = make([]FoodSource, 100)
	w.FoodSourceCells = make([][]*FoodSource, w.length)
	for i := 0; i < w.length; i += 1 {
		w.FoodSourceCells[i] = make([]*FoodSource, w.height)
	}
	var x, y float64
	for i := range w.Resources {
		if i < 4 {
			x = rs.Float64()*20 + float64(w.length)/2 - 10
			y = rs.Float64()*20 + float64(w.height)/2 - 10
		} else if i <= 10 {
			x = rs.Float64()*100 + float64(w.length)/2 - 50
			y = rs.Float64()*100 + float64(w.height)/2 - 50
		} else {
			x = rs.Float64() * float64(w.length)
			y = rs.Float64() * float64(w.height)
		}
		w.Resources[i] = FoodSource{
			Pos:       utils.NewCoordinate(x, y),
			Amount:    atomic.Int32{},
			MaxAmount: 100_000,
			Type:      FoodTypes[rand.IntN(len(FoodTypes))],
			Radius:    0.2,
		}
		cx := int(w.Resources[i].Pos.X())
		cy := int(w.Resources[i].Pos.Y())
		if w.FoodSourceCells[cx][cy] != nil {
			i--
			continue
		}
		w.Resources[i].Amount.Store(w.Resources[i].MaxAmount)
		w.FoodSourceCells[cx][cy] = &w.Resources[i]
	}
	// init images
	w.initImages()
}

func (w *World) findAvgPheremone(pos utils.Coordinate) map[pheremone.Pheremone]pheremone.AveragePheremone {
	cx := int(pos.X() * PheremoneIndexPerCell)
	cy := int(pos.Y() * PheremoneIndexPerCell)
	if cx < 0 || cx >= len(w.AveragePheremoneCell) || cy < 0 || cy >= len(w.AveragePheremoneCell[0]) {
		return nil
	}
	return w.AveragePheremoneCell[cx][cy]
}

func (w *World) CullPheremones(currentTime float64) {
	// binary search for most recently expired pheremone
	for i := w.FirstValidPheremone; i != w.LastPheremoneIndex; i = (i + 1) % len(w.Pheremones) {
		if w.Pheremones[i].Expiration < currentTime {
			ph := w.Pheremones[i]
			mp := w.findAvgPheremone(ph.Pos)
			avgPh, exists := mp[ph.Type]
			if !exists {
				// shouldn't happen
				continue
			}
			avgPh.RemovePheremoneMark(ph)
			if avgPh.Count > 0 {
				mp[ph.Type] = avgPh
			} else {
				delete(mp, ph.Type)
			}
		} else {
			w.FirstValidPheremone = i
			break
		}
	}
}

func (w *World) AddPheremone(ph pheremone.PheremoneMark) {
	w.LastPheremoneIndex = (w.LastPheremoneIndex + 1) % len(w.Pheremones)
	if ph.Type != pheremone.PheremoneHome {
		w.Pheremones[w.LastPheremoneIndex] = ph
	}
	// add to Pheremone cell
	mp := w.findAvgPheremone(ph.Pos)
	avgPh, exists := mp[ph.Type]
	if !exists {
		avgPh = pheremone.NewAveragePheremone(ph)
	} else {
		avgPh.AddPheremoneMark(ph)
	}
	mp[ph.Type] = avgPh
}

func (w *World) GetPheremones() []pheremone.PheremoneMark {
	if w.LastPheremoneIndex >= w.FirstValidPheremone {
		return w.Pheremones[w.FirstValidPheremone : w.LastPheremoneIndex+1]
	} else {
		return append(w.Pheremones[w.FirstValidPheremone:], w.Pheremones[:w.LastPheremoneIndex+1]...)
	}
}

func (w *World) GetAveragePheremones(pos utils.Coordinate) map[pheremone.Pheremone]pheremone.AveragePheremone {
	return w.findAvgPheremone(pos)
}

// large search
func (w *World) GetNearbyResource(pos utils.Coordinate) *FoodSource {
	x := int(pos.X())
	y := int(pos.Y())
	for i := x - 1; i <= x+1; i += 1 {
		for j := y - 1; j <= y+1; j += 1 {
			if w.FoodSourceCells[i][j] != nil {
				return w.FoodSourceCells[i][j]
			}
		}
	}
	return nil
}

// returns the direction of the average gradient of the given pheremone type in the area around the ant
func (w *World) GetPheremoneDirection(pos utils.Coordinate, phType pheremone.Pheremone) (float64, bool) {
	var sumX float64 = 0
	var sumY float64 = 0
	var real bool = false
	viewDistance := 5.0
	const step float64 = 1.0 / PheremoneIndexPerCell
	for i := pos.X() - viewDistance*step; i < pos.X()+viewDistance*step; i += step {
		for j := pos.Y() - viewDistance*step; j < pos.Y()+viewDistance*step; j += step {
			avgPhs := w.findAvgPheremone(utils.NewCoordinate(i, j))
			if avgPhs == nil {
				continue
			}
			avgPh, exists := avgPhs[phType]
			if exists {
				real = true
				loc := utils.NewCoordinate(i, j)
				dir := pos.AngleTo(loc)
				sumX += math.Cos(dir) * math.Pow(avgPh.Strength(), 2)
				sumY += math.Sin(dir) * math.Pow(avgPh.Strength(), 2)
			}
		}
	}
	return math.Atan2(sumY, sumX), real
}
