package utils

import (
	"math"
	"math/rand/v2"
)

type Coordinate struct {
	x float64
	y float64
}

func NewCoordinate(x float64, y float64) Coordinate {
	return Coordinate{x: x, y: y}
}

func RandomCoordinate(maxX float64, maxY float64) Coordinate {
	return Coordinate{
		x: rand.Float64() * maxX,
		y: rand.Float64() * maxY,
	}
}

func (ca Coordinate) X() float64 {
	return ca.x
}

func (ca Coordinate) Y() float64 {
	return ca.y
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

func (ca Coordinate) DistanceTo(other Coordinate) float64 {
	return math.Hypot(ca.x-other.x, ca.y-other.y)
}
