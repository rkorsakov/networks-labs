package logic

import (
	proto "snake-game/internal/proto/gen"
	"sync/atomic"
)

func (gl *GameLogic) GeneratePlayerID() int32 {
	return atomic.AddInt32(&gl.playerIDCounter, 1)
}

func (gl *GameLogic) NewPlayer(name string, playerType proto.PlayerType, role proto.NodeRole) *proto.GamePlayer {
	return &proto.GamePlayer{
		Name:  name,
		Id:    gl.GeneratePlayerID(),
		Role:  role,
		Type:  playerType,
		Score: 0,
	}
}
