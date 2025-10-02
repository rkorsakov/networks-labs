package logic

import (
	"fmt"
	proto "snake-game/internal/proto/gen"
)

func (gl *GameLogic) GetField() *Field {
	return gl.field
}

func (gl *GameLogic) GetFoods() []*proto.GameState_Coord { return gl.state.Foods }

func (gl *GameLogic) GetSnakes() []*proto.GameState_Snake { return gl.state.Snakes }

func (gl *GameLogic) GetPlayer(playerID int32) (*proto.GamePlayer, error) {
	for _, val := range gl.state.Players.Players {
		if val.Id == playerID {
			return val, nil
		}
	}
	return nil, fmt.Errorf("player %d not found", playerID)
}

func (gl *GameLogic) GetPlayers() *proto.GamePlayers {
	return gl.state.Players
}
