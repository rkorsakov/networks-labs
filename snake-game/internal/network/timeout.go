package network

import (
	"log"
	"net"
	prt "snake-game/internal/proto/gen"
	"strconv"
)

func (m *Manager) handleNodeTimeout(addr *net.UDPAddr) {
	log.Printf("Node %s timed out", addr)
	var timedOutPlayer *prt.GamePlayer
	for _, player := range m.gameAnnounce.GetPlayers().GetPlayers() {
		playerAddr := net.JoinHostPort(player.GetIpAddress(), strconv.Itoa(int(player.GetPort())))
		expectedAddr := net.JoinHostPort(addr.IP.String(), strconv.Itoa(addr.Port))
		if playerAddr == expectedAddr {
			timedOutPlayer = player
			break
		}
	}
	if timedOutPlayer == nil {
		return
	}
	switch m.role {
	case prt.NodeRole_MASTER:
		m.Kill(timedOutPlayer)
	case prt.NodeRole_DEPUTY:
	case prt.NodeRole_NORMAL:

	}
}
