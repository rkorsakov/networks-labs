package interfaces

import prt "snake-game/internal/proto/gen"

type GameAnnouncementListener interface {
	OnGameAnnouncement(games []*prt.GameAnnouncement)
}

type GameStateListener interface {
	OnGameStateReceived(state *prt.GameState)
}
