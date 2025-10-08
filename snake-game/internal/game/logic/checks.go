package logic

import proto "snake-game/internal/proto/gen"

func (gl *GameLogic) isFoodAtPosition(coord *proto.GameState_Coord) bool {
	for _, food := range gl.state.Foods {
		if food.X == coord.X && food.Y == coord.Y {
			return true
		}
	}
	return false
}

func (gl *GameLogic) isReverseDirection(current, new proto.Direction) bool {
	return (current == proto.Direction_UP && new == proto.Direction_DOWN) ||
		(current == proto.Direction_DOWN && new == proto.Direction_UP) ||
		(current == proto.Direction_LEFT && new == proto.Direction_RIGHT) ||
		(current == proto.Direction_RIGHT && new == proto.Direction_LEFT)
}

func (gl *GameLogic) CanPlaceSnake() bool {
	field := gl.GetField()
	snakes := gl.GetSnakes()
	foods := gl.GetFoods()
	for x := int32(0); x < field.Width-4; x++ {
		for y := int32(0); y < field.Height-4; y++ {
			squareEmpty := true
			for i := x; i < x+5; i++ {
				for j := y; j < y+5; j++ {
					coord := &proto.GameState_Coord{
						X: i % field.Width,
						Y: j % field.Height,
					}
					if !field.IsPositionEmpty(coord, snakes, foods) {
						squareEmpty = false
						break
					}
				}
				if !squareEmpty {
					break
				}
			}
			if squareEmpty {
				centerCoord := &proto.GameState_Coord{
					X: (x + 2) % field.Width,
					Y: (y + 2) % field.Height,
				}
				if !gl.isFoodAtPosition(centerCoord) {
					return true
				}
			}
		}
	}
	return false
}
