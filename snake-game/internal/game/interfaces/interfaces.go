package interfaces

import (
	"snake-game/internal/game/logic"
	prt "snake-game/internal/proto/gen"
)

type GameAnnouncementListener interface {
	OnGameAnnouncement(games []*prt.GameAnnouncement)
}

type GameStateListener interface {
	OnGameStateReceived(state *prt.GameState)
}

type GameJoinListener interface {
	OnGameAddPlayer(player *prt.GamePlayer)
	GetLogic() *logic.GameLogic
}

type SteerListener interface {
	OnSteerReceived(playerID int32, direction prt.Direction) error
}
