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

type Resource struct {
	location utils.Coordinate
	id       uuid.UUID
	amount   float32
	radius   float32
}
