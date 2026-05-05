package world

type World struct {
	width  float32
	length float32
	time   int32
}

func NewWorld(width, length float32) *World {
	return &World{
		width:  width,
		length: length,
		time:   0,
	}
}
