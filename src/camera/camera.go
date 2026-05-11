package camera

import (
	"hive_sim/src/utils"
	"hive_sim/src/world"

	"github.com/hajimehoshi/ebiten/v2"
)

type FBounds struct {
	Min utils.Coordinate
	Max utils.Coordinate
}

func NewFBounds(xmin, ymin, xmax, ymax float64) FBounds {
	return FBounds{
		Min: utils.NewCoordinate(xmin, ymin),
		Max: utils.NewCoordinate(xmax, ymax),
	}
}

type Camera struct {
	Position     utils.Coordinate
	x_width      float64
	y_width      float64
	max_width    int
	min_width    int
	aspect_ratio float64
	xs           float32
	ys           float32
}

func NewCamera(w *world.World, position *utils.Coordinate, aspect_ratio float64) Camera {
	var pos utils.Coordinate
	if position == nil {
		pos = utils.NewCoordinate(float64(w.Length())/2, float64(w.Height())/2) // center of the screen if not given
	} else {
		pos = *position
	}
	// camera thinks in terms of world coordinates, so we need to convert the screen coordinates to world coordinates
	return Camera{
		Position:     pos,
		x_width:      float64(w.Length()) / 10.0 * aspect_ratio,
		y_width:      float64(w.Height()) / 10.0,
		max_width:    w.Length(),
		min_width:    2,
		aspect_ratio: aspect_ratio,
	}
}

func (c *Camera) Move(dx float64, dy float64) {
	c.Position.Add(dx, dy)
}

func (c *Camera) Zoom(step float64) {
	c.x_width = float64(c.x_width) - step
	if c.x_width < float64(c.min_width) {
		c.x_width = float64(c.min_width)
	} else if c.x_width > float64(c.max_width) {
		c.x_width = float64(c.max_width)
	}
	c.y_width = c.x_width / c.aspect_ratio
}

func (c *Camera) SetScale(screen *ebiten.Image) {
	c.xs = float32(screen.Bounds().Dx()) / float32(c.x_width)
	c.ys = float32(screen.Bounds().Dy()) / float32(c.y_width)
}

func (c *Camera) GetScale() (float32, float32) {
	return c.xs, c.ys
}

func (c *Camera) GetBounds() FBounds {
	return NewFBounds(
		c.Position.X()-c.x_width/2,
		c.Position.Y()-c.y_width/2,
		c.Position.X()+c.x_width/2+1.0,
		c.Position.Y()+c.y_width/2+1.0,
	)
}

func (c *Camera) ScreenToWorld(sx, sy float64) (float64, float64) {
	bounds := c.GetBounds()
	worldX := sx/float64(c.xs) + bounds.Min.X()
	worldY := sy/float64(c.ys) + bounds.Min.Y()
	return worldX, worldY
}

func InBounds(pos utils.Coordinate, bounds FBounds) bool {
	return pos.X() > float64(bounds.Min.X()) && pos.X() < float64(bounds.Max.X()) && pos.Y() > float64(bounds.Min.Y()) && pos.Y() < float64(bounds.Max.Y())
}
