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
	for x := int32(0); x < field.Width-4; x++ {
		for y := int32(0); y < field.Height-4; y++ {
			squareEmpty := true
			for i := x; i < x+5; i++ {
				for j := y; j < y+5; j++ {
					coord := &proto.GameState_Coord{
						X: i % field.Width,
						Y: j % field.Height,
					}
					for _, snake := range snakes {
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
				headCoord := &proto.GameState_Coord{X: x + 2, Y: y + 2}
				tailPositions := []proto.GameState_Coord{
					{X: headCoord.X + 1, Y: headCoord.Y},
					{X: headCoord.X - 1, Y: headCoord.Y},
					{X: headCoord.X, Y: headCoord.Y + 1},
					{X: headCoord.X, Y: headCoord.Y - 1},
				}
				for _, tailCoord := range tailPositions {
					tailCoord.X = (tailCoord.X + field.Width) % field.Width
					tailCoord.Y = (tailCoord.Y + field.Height) % field.Height
					headHasFood := gl.isFoodAtPosition(headCoord)
					tailHasFood := gl.isFoodAtPosition(&tailCoord)
					if !headHasFood && !tailHasFood {
						return true
					}
				}
			}
		}
	}
	return false
}
