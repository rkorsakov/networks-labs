package logic

import (
	"fmt"
	"math/rand/v2"
	"time"

	proto "snake-game/internal/proto/gen"
)

type GameLogic struct {
	Config        *proto.GameConfig
	field         *Field
	state         *proto.GameState
	rnd           *rand.Rand
	pendingSteers map[int32]proto.Direction
}

func NewGameLogic(config *proto.GameConfig) *GameLogic {
	if config == nil {
		config = &proto.GameConfig{
			Width:      40,
			Height:     30,
			FoodStatic: 1,
		}
	}

	gl := &GameLogic{
		Config: config,
		field:  NewField(config.Width, config.Height),
		state: &proto.GameState{
			StateOrder: 0,
			Snakes:     make([]*proto.GameState_Snake, 0),
			Foods:      make([]*proto.GameState_Coord, 0),
			Players:    &proto.GamePlayers{Players: make([]*proto.GamePlayer, 0)},
		},
		rnd:           rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)),
		pendingSteers: make(map[int32]proto.Direction),
	}
	return gl
}

func (gl *GameLogic) Init() {
	gl.generateInitialFood()
}

func (gl *GameLogic) Update() error {
	for playerID, newDirection := range gl.pendingSteers {
		if snake := gl.GetSnakeByPlayerID(playerID); snake != nil && snake.State == proto.GameState_Snake_ALIVE {
			currentDir := snake.HeadDirection
			if !gl.isReverseDirection(currentDir, newDirection) {
				snake.HeadDirection = newDirection
			}
		}
	}
	gl.pendingSteers = make(map[int32]proto.Direction)
	gl.moveSnakes()
	gl.checkCollisions()
	gl.updateFood()
	gl.state.StateOrder++
	return nil
}

func (gl *GameLogic) moveSnakes() {
	for _, snake := range gl.state.Snakes {
		gl.moveSnake(snake)
	}
}

func (gl *GameLogic) moveSnake(snake *proto.GameState_Snake) {

	newHead := &proto.GameState_Coord{
		X: snake.Points[0].X,
		Y: snake.Points[0].Y,
	}

	switch snake.HeadDirection {
	case proto.Direction_UP:
		newHead.Y = (newHead.Y - 1 + gl.field.Height) % gl.field.Height
	case proto.Direction_DOWN:
		newHead.Y = (newHead.Y + 1) % gl.field.Height
	case proto.Direction_LEFT:
		newHead.X = (newHead.X - 1 + gl.field.Width) % gl.field.Width
	case proto.Direction_RIGHT:
		newHead.X = (newHead.X + 1) % gl.field.Width
	}

	newPoints := make([]*proto.GameState_Coord, 0, len(snake.Points)+1)
	newPoints = append(newPoints, newHead)
	newPoints = append(newPoints, snake.Points...)

	ateFood := false
	for i, food := range gl.state.Foods {
		if food.X == newHead.X && food.Y == newHead.Y {
			ateFood = true
			gl.state.Foods = append(gl.state.Foods[:i], gl.state.Foods[i+1:]...)
			if player, err := gl.GetPlayer(snake.PlayerId); err == nil {
				player.Score++
			}
			break
		}
	}

	if !ateFood && len(newPoints) > 2 {
		newPoints = newPoints[:len(newPoints)-1]
	}

	snake.Points = newPoints
}

func (gl *GameLogic) checkCollisions() {
	type coordKey struct{ X, Y int32 }
	collisions := make(map[int32]bool)
	headCoords := make(map[coordKey]int32)
	bodyCoords := make(map[coordKey]struct{})

	for _, snake := range gl.state.Snakes {
		if snake.State != proto.GameState_Snake_ALIVE {
			continue
		}

		head := snake.Points[0]
		headKey := coordKey{X: head.X, Y: head.Y}
		headCoords[headKey] = snake.PlayerId

		for i := 1; i < len(snake.Points); i++ {
			point := snake.Points[i]
			bodyKey := coordKey{X: point.X, Y: point.Y}
			bodyCoords[bodyKey] = struct{}{}
		}
	}
	collidedWith := make(map[int32][]int32)
	for headKey, playerID := range headCoords {
		if _, exists := bodyCoords[headKey]; exists {
			collisions[playerID] = true
			continue
		}
		for otherHeadKey, otherPlayerID := range headCoords {
			if playerID == otherPlayerID {
				continue
			}
			if headKey.X == otherHeadKey.X && headKey.Y == otherHeadKey.Y {
				collisions[playerID] = true
				collisions[otherPlayerID] = true
				collidedWith[playerID] = append(collidedWith[playerID], otherPlayerID)
				collidedWith[otherPlayerID] = append(collidedWith[otherPlayerID], playerID)
				break
			}
		}
	}

	for killerID, victims := range collidedWith {
		if snake := gl.GetSnakeByPlayerID(killerID); snake != nil && snake.State == proto.GameState_Snake_ALIVE {
			if player, err := gl.GetPlayer(killerID); err == nil {
				player.Score += int32(len(victims))
			}
		}
	}

	for playerID := range collisions {
		if snake := gl.GetSnakeByPlayerID(playerID); snake != nil {
			snake.State = proto.GameState_Snake_ZOMBIE

			for _, point := range snake.Points {
				if gl.rnd.Float32() < 0.5 {
					gl.state.Foods = append(gl.state.Foods, &proto.GameState_Coord{
						X: point.X,
						Y: point.Y,
					})
				}
			}
		}
	}
}

func (gl *GameLogic) updateFood() {
	targetFood := int(gl.Config.FoodStatic)
	for _, snake := range gl.state.Snakes {
		if snake.State == proto.GameState_Snake_ALIVE {
			targetFood++
		}
	}

	for len(gl.state.Foods) < targetFood {
		newFood := gl.generateFoodPosition()
		if !gl.isPositionOccupied(newFood) {
			gl.state.Foods = append(gl.state.Foods, newFood)
		} else {
			break
		}
	}
}

func (gl *GameLogic) generateInitialFood() {
	for i := 0; i < int(gl.Config.FoodStatic); i++ {
		gl.state.Foods = append(gl.state.Foods, gl.generateFoodPosition())
	}
}

func (gl *GameLogic) generateFoodPosition() *proto.GameState_Coord {
	return gl.field.GetRandomPosition(gl.rnd)
}

func (gl *GameLogic) isPositionOccupied(coord *proto.GameState_Coord) bool {

	for _, food := range gl.state.Foods {
		if food.X == coord.X && food.Y == coord.Y {
			return true
		}
	}

	for _, snake := range gl.state.Snakes {
		for _, point := range snake.Points {
			if point.X == coord.X && point.Y == coord.Y {
				return true
			}
		}
	}

	return false
}

func (gl *GameLogic) AddPlayer(player *proto.GamePlayer) {
	gl.state.Players.Players = append(gl.state.Players.Players, player)
	if player.Role != proto.NodeRole_VIEWER {
		gl.placeSnake(player)
	}
}

func (gl *GameLogic) KillPlayer(playerID int32) {
	if snake := gl.GetSnakeByPlayerID(playerID); snake != nil {
		snake.State = proto.GameState_Snake_ZOMBIE
	}
}

func (gl *GameLogic) SteerSnake(playerID int32, direction proto.Direction) error {
	snake := gl.GetSnakeByPlayerID(playerID)
	if snake == nil {
		return fmt.Errorf("snake for player %d not found", playerID)
	}
	gl.pendingSteers[playerID] = direction
	return nil
}

func (gl *GameLogic) getTailPosition(head *proto.GameState_Coord, direction proto.Direction) *proto.GameState_Coord {
	switch direction {
	case proto.Direction_UP:
		return &proto.GameState_Coord{X: head.X, Y: head.Y + 1}
	case proto.Direction_DOWN:
		return &proto.GameState_Coord{X: head.X, Y: head.Y - 1}
	case proto.Direction_LEFT:
		return &proto.GameState_Coord{X: head.X + 1, Y: head.Y}
	case proto.Direction_RIGHT:
		return &proto.GameState_Coord{X: head.X - 1, Y: head.Y}
	}
	return &proto.GameState_Coord{X: head.X, Y: head.Y + 1}
}

func (gl *GameLogic) getOppositeDirection(dir proto.Direction) proto.Direction {
	switch dir {
	case proto.Direction_UP:
		return proto.Direction_DOWN
	case proto.Direction_DOWN:
		return proto.Direction_UP
	case proto.Direction_LEFT:
		return proto.Direction_RIGHT
	case proto.Direction_RIGHT:
		return proto.Direction_LEFT
	}
	return proto.Direction_UP
}

func (gl *GameLogic) SetState(state *proto.GameState) {
	gl.state = state
}
