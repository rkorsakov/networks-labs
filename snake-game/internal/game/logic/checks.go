package logic

import proto "snake-game/internal/proto/gen"

func (gl *GameLogic) isFoodOnSnake(coord *proto.GameState_Coord) bool {
	for _, snake := range gl.snakes {
		for _, point := range snake.Points {
			if point.X == coord.X && point.Y == coord.Y {
				return true
			}
		}
	}
	return false
}

func (gl *GameLogic) isFoodAtPosition(coord *proto.GameState_Coord) bool {
	for _, food := range gl.foods {
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
