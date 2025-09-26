package logic

import (
	"math/rand/v2"
	proto "snake-game/internal/proto/gen"
	"time"
)

type GameLogic struct {
	config  *proto.GameConfig
	field   *Field
	state   *proto.GameState
	players map[int32]*proto.GamePlayer
	snakes  map[int32]*proto.GameState_Snake
	Foods   []*proto.GameState_Coord
	rng     *rand.Rand
}

func NewGameLogic(config *proto.GameConfig) *GameLogic {
	gl := &GameLogic{
		config:  config,
		field:   NewField(config.Width, config.Height),
		players: make(map[int32]*proto.GamePlayer),
		snakes:  make(map[int32]*proto.GameState_Snake),
		state:   &proto.GameState{StateOrder: 0},
		Foods:   make([]*proto.GameState_Coord, 0),
		rng:     rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)),
	}
	gl.generateInitialFood()
	return gl
}

//func (gl *GameLogic) Update() error {
//}

func (gl *GameLogic) checkCollisions() {

}

func (gl *GameLogic) generateInitialFood() {
	for i := 0; i < int(gl.config.FoodStatic); i++ {
		gl.Foods = append(gl.Foods, gl.generateFoodPosition())
	}
}

func (gl *GameLogic) generateFoodPosition() *proto.GameState_Coord {
	return gl.field.GetRandomPosition(gl.rng)

}

func (gl *GameLogic) GetField() *Field {
	return gl.field
}
