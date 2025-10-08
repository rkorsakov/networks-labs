package logic

import (
	"math/rand/v2"
	proto "snake-game/internal/proto/gen"
)

type Field struct {
	Width  int32
	Height int32
}

func NewField(width, height int32) *Field {
	return &Field{
		Width:  width,
		Height: height,
	}
}

func (f *Field) IsValidPosition(coord *proto.GameState_Coord) bool {
	return coord.X >= 0 && coord.X < f.Width &&
		coord.Y >= 0 && coord.Y < f.Height
}

func (f *Field) WrapPosition(coord *proto.GameState_Coord) *proto.GameState_Coord {
	x := coord.X
	y := coord.Y

	if x < 0 {
		x = f.Width - 1
	} else if x >= f.Width {
		x = 0
	}

	if y < 0 {
		y = f.Height - 1
	} else if y >= f.Height {
		y = 0
	}

	return &proto.GameState_Coord{X: x, Y: y}
}

func (f *Field) GetRandomPosition(rnd *rand.Rand) *proto.GameState_Coord {
	return &proto.GameState_Coord{
		X: int32(rnd.IntN(int(f.Width))),
		Y: int32(rnd.IntN(int(f.Height))),
	}
}
