package logic

import proto "snake-game/internal/proto/gen"

func (gl *GameLogic) IsFoodAtPosition(coord *proto.GameState_Coord) bool {
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
