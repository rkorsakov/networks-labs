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
	for x := int32(0); x < gl.GetField().Width-4; x++ {
		for y := int32(0); y < gl.GetField().Height-4; y++ {
			squareEmpty := true
			for i := x; i < x+5; i++ {
				for j := y; j < y+5; j++ {
					coord := &proto.GameState_Coord{
						X: i % gl.GetField().Width,
						Y: j % gl.GetField().Height,
					}
					for _, snake := range gl.GetSnakes() {
						for _, point := range snake.Points {
							if point.X == coord.X && point.Y == coord.Y {
								squareEmpty = false
								break
							}
						}
						if !squareEmpty {
							break
						}
					}
					if !squareEmpty {
						break
					}
				}
				if !squareEmpty {
					break
				}
			}
			if squareEmpty {
				centerX := x + 2
				centerY := y + 2
				centerCoord := &proto.GameState_Coord{X: centerX, Y: centerY}
				if !gl.isFoodAtPosition(centerCoord) {
					return true
				}
			}
		}
	}
	return false
}
