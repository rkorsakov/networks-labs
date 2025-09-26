package logic

import (
	"fmt"
	"math/rand/v2"
	proto "snake-game/internal/proto/gen"
	"time"
)

type GameLogic struct {
	Config        *proto.GameConfig
	field         *Field
	state         *proto.GameState
	players       map[int32]*proto.GamePlayer
	snakes        map[int32]*proto.GameState_Snake
	foods         []*proto.GameState_Coord
	rnd           *rand.Rand
	pendingSteers map[int32]proto.Direction
	stateOrder    int64
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
		stateOrder:    0,
	}
	testPlayer := NewPlayer("TestPlayer", proto.PlayerType_HUMAN, proto.NodeRole_MASTER)
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

	gl.stateOrder++

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
		return fmt.Errorf("snake is not alive")
	}
	gl.pendingSteers[playerID] = direction
	return nil
}

func (gl *GameLogic) moveSnake(snake *proto.GameState_Snake) {
	if snake.State != proto.GameState_Snake_ALIVE {
		return
	}
	oldHead := snake.Points[0]
	newHead := &proto.GameState_Coord{
		X: oldHead.X,
		Y: oldHead.Y,
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

func (gl *GameLogic) isReverseDirection(current, new proto.Direction) bool {
	return (current == proto.Direction_UP && new == proto.Direction_DOWN) ||
		(current == proto.Direction_DOWN && new == proto.Direction_UP) ||
		(current == proto.Direction_LEFT && new == proto.Direction_RIGHT) ||
		(current == proto.Direction_RIGHT && new == proto.Direction_LEFT)
}

func (gl *GameLogic) checkCollisions() {
	occupiedCoords := make(map[*proto.GameState_Coord]int32)
	for playerID, snake := range gl.snakes {
		for _, point := range snake.Points {
			key := proto.GameState_Coord{X: point.X, Y: point.Y}
			if _, exists := occupiedCoords[&key]; !exists {
				occupiedCoords[&key] = playerID
			}
		}
	}
	collisions := make(map[int32]bool)
	for playerID, snake := range gl.snakes {
		if snake.State != proto.GameState_Snake_ALIVE {
			continue
		}
		head := snake.Points[0]
		key := proto.GameState_Coord{X: head.X, Y: head.Y}
		if occupyingPlayerID, occupied := occupiedCoords[&key]; occupied {
			isTailOfSameSnake := false
			if len(snake.Points) > 1 {
				tail := snake.Points[len(snake.Points)-1]
				if tail.X == head.X && tail.Y == head.Y {
					isTailOfSameSnake = true
				}
			}
			if !isTailOfSameSnake || occupyingPlayerID != playerID {
				collisions[playerID] = true
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

func (gl *GameLogic) generateInitialFood() {
	for i := 0; i < int(gl.Config.FoodStatic); i++ {
		gl.foods = append(gl.foods, gl.generateFoodPosition())
	}
}

func (gl *GameLogic) placeSnakes() {
	for _, player := range gl.players {
		gl.placeSnakeForPlayer(player)
	}
}

func (gl *GameLogic) placeSnakeForPlayer(player *proto.GamePlayer) {
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

func (gl *GameLogic) isFoodAtPosition(coord *proto.GameState_Coord) bool {
	for _, food := range gl.foods {
		if food.X == coord.X && food.Y == coord.Y {
			return true
		}
	}
	return false
}

func (gl *GameLogic) AddPlayer(player *proto.GamePlayer) {
	gl.players[player.Id] = player
}

func (gl *GameLogic) GetPlayer(playerID int32) *proto.GamePlayer {
	return gl.players[playerID]
}

func (gl *GameLogic) GetPlayers() map[int32]*proto.GamePlayer {
	return gl.players
}

func (gl *GameLogic) generateFoodPosition() *proto.GameState_Coord {
	return gl.field.GetRandomPosition(gl.rnd)

}

func (gl *GameLogic) GetField() *Field {
	return gl.field
}

func (gl *GameLogic) GetFoods() []*proto.GameState_Coord { return gl.foods }

func (gl *GameLogic) GetSnakes() map[int32]*proto.GameState_Snake { return gl.snakes }
