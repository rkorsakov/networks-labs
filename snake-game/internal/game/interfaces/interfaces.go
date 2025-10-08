package interfaces

import prt "snake-game/internal/proto/gen"

type GameAnnouncementListener interface {
	OnGameAnnouncement(games []*prt.GameAnnouncement)
}

type GameStateListener interface {
	OnGameStateReceived(state *prt.GameState)
}

type GameJoinListener interface {
	OnGameAddPlayer(player *prt.GamePlayer)
}

type SteerListener interface {
	OnSteerReceived(playerID int32, direction prt.Direction) error
}
