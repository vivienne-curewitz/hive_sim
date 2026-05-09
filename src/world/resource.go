package world

import (
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
	Type   FoodType
	Amount float32
	Pos    utils.Coordinate
	Radius float64
}

type Resource struct {
	location utils.Coordinate
	id       uuid.UUID
	amount   float32
	radius   float32
}
