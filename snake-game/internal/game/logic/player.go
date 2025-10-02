package logic

import (
	"math/rand/v2"
	proto "snake-game/internal/proto/gen"
)

func GeneratePlayerID() int32 {
	return rand.Int32()
}

func (gl *GameLogic) NewPlayer(name string, playerType proto.PlayerType, role proto.NodeRole, id int32) *proto.GamePlayer {
	return &proto.GamePlayer{
		Name:  name,
		Id:    id,
		Role:  role,
		Type:  playerType,
		Score: 0,
	}
}
