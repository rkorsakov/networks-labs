package logic

import proto "snake-game/internal/proto/gen"

func (gl *GameLogic) GetField() *Field {
	return gl.field
}

func (gl *GameLogic) GetFoods() []*proto.GameState_Coord { return gl.foods }

func (gl *GameLogic) GetSnakes() map[int32]*proto.GameState_Snake { return gl.snakes }

func (gl *GameLogic) GetPlayer(playerID int32) *proto.GamePlayer {
	return gl.players[playerID]
}

func (gl *GameLogic) GetPlayers() map[int32]*proto.GamePlayer {
	return gl.players
}
