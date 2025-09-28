package logic

import (
	"fmt"
	"math/rand/v2"
	proto "snake-game/internal/proto/gen"
	"time"
)

//еще тоже надо с rnd решить нужен или нет

//еще нужно в graphics сделать чтобы Layout или че там крч норм окно было чтобы

//потом надо сделать отдельное в graphics типо gui и нужен еще guicontroller наверное или чет такое

type GameLogic struct {
	Config          *proto.GameConfig
	field           *Field
	state           *proto.GameState
	players         map[int32]*proto.GamePlayer
	snakes          map[int32]*proto.GameState_Snake
	foods           []*proto.GameState_Coord
	rnd             *rand.Rand
	pendingSteers   map[int32]proto.Direction
	playerIDCounter int32
}

func NewGameLogic(Config *proto.GameConfig) *GameLogic {
	gl := &GameLogic{
		Config:        Config,
		field:         NewField(Config.Width, Config.Height),
		players:       make(map[int32]*proto.GamePlayer),
		snakes:        make(map[int32]*proto.GameState_Snake),
		state:         &proto.GameState{StateOrder: 0},
		foods:         make([]*proto.GameState_Coord, 0),
		rnd:           rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)),
		pendingSteers: make(map[int32]proto.Direction),
	}
	testPlayer := gl.NewPlayer("TestPlayer", proto.PlayerType_HUMAN, proto.NodeRole_MASTER)
	gl.AddPlayer(testPlayer)
	gl.generateInitialFood()
	gl.placeSnakes()
	return gl
}

func (gl *GameLogic) Update() error {
	for playerID, newDirection := range gl.pendingSteers {
		if snake, exists := gl.snakes[playerID]; exists && snake.State == proto.GameState_Snake_ALIVE {
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
	for _, snake := range gl.snakes {
		gl.moveSnake(snake)
	}
}

func (gl *GameLogic) SteerSnake(playerID int32, direction proto.Direction) error {
	snake, exists := gl.snakes[playerID]
	if !exists {
		return fmt.Errorf("snake for player %d not found", playerID)
	}
	if snake.State != proto.GameState_Snake_ALIVE {
		return fmt.Errorf("snake is ZOMBIE")
	}
	gl.pendingSteers[playerID] = direction
	return nil
}

func (gl *GameLogic) moveSnake(snake *proto.GameState_Snake) {
	if snake.State != proto.GameState_Snake_ALIVE {
		return
	}
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
	for i, food := range gl.foods {
		if food.X == newHead.X && food.Y == newHead.Y {
			ateFood = true
			gl.foods = append(gl.foods[:i], gl.foods[i+1:]...)
			break
		}
	}

	if !ateFood && len(newPoints) > 2 {
		newPoints = newPoints[:len(newPoints)-1]
	}

	snake.Points = newPoints
}

func (gl *GameLogic) checkCollisions() {
	collisions := make(map[int32]bool)
	occupiedCoords := make([]*proto.GameState_Coord, 0)
	for _, snake := range gl.snakes {
		if len(snake.Points) > 1 {
			occupiedCoords = append(occupiedCoords, snake.Points[1:]...)
		}
	}
	for playerID, snake := range gl.snakes {
		if snake.State != proto.GameState_Snake_ALIVE {
			continue
		}
		head := snake.Points[0]
		for _, bodyCoord := range occupiedCoords {
			if head.X == bodyCoord.X && head.Y == bodyCoord.Y {
				collisions[playerID] = true
				break
			}
		}
		if !collisions[playerID] {
			for otherPlayerID, otherSnake := range gl.snakes {
				if otherPlayerID == playerID || otherSnake.State != proto.GameState_Snake_ALIVE {
					continue
				}
				otherHead := otherSnake.Points[0]
				if head.X == otherHead.X && head.Y == otherHead.Y {
					collisions[playerID] = true
					collisions[otherPlayerID] = true
					break
				}
			}
		}
	}
	for playerID := range collisions {
		if snake, exists := gl.snakes[playerID]; exists {
			for _, point := range snake.Points {
				if gl.rnd.Float32() < 0.5 {
					gl.foods = append(gl.foods, &proto.GameState_Coord{
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
	for _, snake := range gl.snakes {
		if snake.State == proto.GameState_Snake_ALIVE {
			targetFood++
		}
	}
	for len(gl.foods) < targetFood {
		newFood := gl.generateFoodPosition()
		if !gl.isFoodOnSnake(newFood) {
			gl.foods = append(gl.foods, newFood)
		} else {
			break
		}
	}
}

func (gl *GameLogic) generateInitialFood() {
	for i := 0; i < int(gl.Config.FoodStatic); i++ {
		gl.foods = append(gl.foods, gl.generateFoodPosition())
	}
}

func (gl *GameLogic) getTailPosition(head *proto.GameState_Coord, direction proto.Direction) *proto.GameState_Coord {
	switch direction {
	case proto.Direction_UP:
		return &proto.GameState_Coord{X: head.X, Y: head.Y - 1}
	case proto.Direction_DOWN:
		return &proto.GameState_Coord{X: head.X, Y: head.Y + 1}
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

func (gl *GameLogic) generateFoodPosition() *proto.GameState_Coord {
	return gl.field.GetRandomPosition(gl.rnd)

}

func (gl *GameLogic) AddPlayer(player *proto.GamePlayer) {
	gl.players[player.Id] = player
}
