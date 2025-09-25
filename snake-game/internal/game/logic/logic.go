package logic

import (
	"snake-game/internal/game/config"
)

type GameLogic struct {
	config *config.Config
	field  *Field
}

func NewGameLogic(config *config.Config) *GameLogic {
	return &GameLogic{
		config: config,
		field:  &Field{Width: config.Field.Width, Height: config.Field.Height},
	}
}

func (gl *GameLogic) GetField() *Field {
	return gl.field
}
