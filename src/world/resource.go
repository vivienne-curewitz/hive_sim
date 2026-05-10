package world

import (
	"sync/atomic"

	"hive_sim/src/utils"

	"github.com/google/uuid"
)

type ResourceType int

const (
	water ResourceType = iota
	food
)

type FoodType int

const (
	flower FoodType = iota
	beetle
)

var FoodTypes = []FoodType{flower, beetle}

type FoodSource struct {
	Type      FoodType
	Amount    atomic.Int32
	MaxAmount int32
	Pos       utils.Coordinate
	Radius    float64
}

func (fs *FoodSource) TakeX(x int32) {
	remove := -1 * x
	fs.Amount.Add(remove)
}

type Resource struct {
	location utils.Coordinate
	id       uuid.UUID
	amount   float32
	radius   float32
}
