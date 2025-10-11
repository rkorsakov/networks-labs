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
		m.handleMasterTimeout(timedOutPlayer)
	case prt.NodeRole_DEPUTY:
		m.handleDeputyTimeout(timedOutPlayer)
	case prt.NodeRole_NORMAL:
		m.handleNormalTimeout(timedOutPlayer)
	}
}

func (m *Manager) handleMasterTimeout(player *prt.GamePlayer) {
	if player.Role == prt.NodeRole_DEPUTY {
		var newDeputy *prt.GamePlayer
		for _, p := range m.gameAnnounce.GetPlayers().GetPlayers() {
			if p.Id != player.Id && p.Role == prt.NodeRole_NORMAL {
				newDeputy = p
				break
			}
		}
		if newDeputy != nil {
			m.sendRoleChangeMessage(newDeputy, prt.NodeRole_DEPUTY)
		}
	}
	m.sendRoleChangeMessage(player, prt.NodeRole_VIEWER)
	m.Kill(player)
}

func (m *Manager) handleDeputyTimeout(player *prt.GamePlayer) {
	if player.Role == prt.NodeRole_MASTER {
		m.ChangeRole(player, prt.NodeRole_MASTER)
		var newDeputy *prt.GamePlayer
		for _, p := range m.gameAnnounce.GetPlayers().GetPlayers() {
			if p.Role == prt.NodeRole_NORMAL {
				newDeputy = p
				break
			}
		}
		if newDeputy != nil {
			m.sendRoleChangeMessage(newDeputy, prt.NodeRole_DEPUTY)
		}
		m.broadcastNewMaster()
	}
}

func (m *Manager) handleNormalTimeout(player *prt.GamePlayer) {
	if player.Role == prt.NodeRole_MASTER {
		var deputy *prt.GamePlayer
		for _, p := range m.gameAnnounce.GetPlayers().GetPlayers() {
			if p.Role == prt.NodeRole_DEPUTY {
				deputy = p
				break
			}
		}
		if deputy != nil {
			for _, gameInfo := range m.AvailableGames {
				if gameInfo.Announcement.GameName == m.gameAnnounce.GameName {
					deputyAddr, err := net.ResolveUDPAddr("udp",
						net.JoinHostPort(deputy.IpAddress, strconv.Itoa(int(deputy.Port))))
					if err == nil {
						gameInfo.MasterAddr = deputyAddr
					}
					break
				}
			}
		}
	}
}
