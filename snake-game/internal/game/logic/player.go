package logic

import (
	proto "snake-game/internal/proto/gen"
	"sync/atomic"
)

var playerIDCounter int32 = 0

func GeneratePlayerID() int32 {
	return atomic.AddInt32(&playerIDCounter, 1)
}

func NewPlayer(name string, playerType proto.PlayerType, role proto.NodeRole) *proto.GamePlayer {
	return &proto.GamePlayer{
		Name:  name,
		Id:    GeneratePlayerID(),
		Role:  role,
		Type:  playerType,
		Score: 0,
	}
}
