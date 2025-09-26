package logic

import proto "snake-game/internal/proto/gen"

func (gl *GameLogic) placeSnakes() {
	for _, player := range gl.players {
		gl.placeSnake(player)
	}
}

func (gl *GameLogic) placeSnake(player *proto.GamePlayer) {
	for attempt := 0; attempt < 100; attempt++ {
		head := gl.field.GetRandomPosition(gl.rnd)
		if gl.isFoodAtPosition(head) {
			continue
		}
		directions := []proto.Direction{
			proto.Direction_UP,
			proto.Direction_DOWN,
			proto.Direction_LEFT,
			proto.Direction_RIGHT,
		}
		gl.rnd.Shuffle(len(directions), func(i, j int) {
			directions[i], directions[j] = directions[j], directions[i]
		})
		var tail *proto.GameState_Coord
		var selectedDirection proto.Direction
		found := false
		for _, dir := range directions {
			tail = gl.getTailPosition(head, dir)
			tail = gl.field.WrapPosition(tail)
			if !gl.isFoodAtPosition(tail) {
				selectedDirection = dir
				found = true
				break
			}
		}
		if !found {
			continue
		}
		head = gl.field.WrapPosition(head)
		headDirection := gl.getOppositeDirection(selectedDirection)
		snake := &proto.GameState_Snake{
			PlayerId:      player.Id,
			Points:        []*proto.GameState_Coord{head, tail},
			State:         proto.GameState_Snake_ALIVE,
			HeadDirection: headDirection,
		}
		gl.snakes[player.Id] = snake
		return
	}
	head := gl.field.GetRandomPosition(gl.rnd)
	head = gl.field.WrapPosition(head)
	tail := gl.getTailPosition(head, proto.Direction_UP)
	tail = gl.field.WrapPosition(tail)
	snake := &proto.GameState_Snake{
		PlayerId:      player.Id,
		Points:        []*proto.GameState_Coord{head, tail},
		State:         proto.GameState_Snake_ALIVE,
		HeadDirection: proto.Direction_DOWN,
	}
	gl.snakes[player.Id] = snake
}
