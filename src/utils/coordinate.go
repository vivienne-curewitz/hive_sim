package utils

import "math"

type Coordinate struct {
	x float64
	y float64
}

func (ca *Coordinate) Add(dx float64, dy float64) {
	ca.x += dx
	ca.y += dy
}

func (ca Coordinate) AngleTo(other Coordinate) float64 {
	rise := other.y - ca.y
	over := other.x - ca.x
	return math.Atan2(rise, over)
}
