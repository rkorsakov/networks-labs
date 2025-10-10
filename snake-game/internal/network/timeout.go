package network

import (
	"log"
	"net"
	prt "snake-game/internal/proto/gen"
)

func (m *Manager) handleNodeTimeout(addr *net.UDPAddr) {
	log.Printf("Node %s timed out", addr)
	switch m.role {
	case prt.NodeRole_MASTER:
	case prt.NodeRole_DEPUTY:
	case prt.NodeRole_NORMAL:
	}
}
