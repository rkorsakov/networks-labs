package network

import (
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	prt "snake-game/internal/proto/gen"
)

type GameMessage struct {
}

func (m *Manager) handleMessage(data []byte, addr *net.UDPAddr) {
	var msg prt.GameMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}
	log.Printf("Received message seq=%d from %s", msg.MsgSeq, addr)
	switch {
	case msg.GetPing() != nil:
		log.Println("Got PING message")
		m.handlePing(&msg, addr)

	case msg.GetSteer() != nil:
		log.Println("Got STEER message")
		m.handleSteer(&msg, addr)

	case msg.GetAck() != nil:
		log.Println("Got ACK message")
		m.handleAck(&msg, addr)

	case msg.GetState() != nil:
		log.Println("Got STATE message")
		m.handleState(&msg, addr)

	case msg.GetAnnouncement() != nil:
		log.Println("Got ANNOUNCEMENT message")
		m.handleAnnouncement(&msg)

	case msg.GetJoin() != nil:
		log.Println("Got JOIN message")
		m.handleJoin(&msg, addr)

	case msg.GetError() != nil:
		log.Println("Got ERROR message")
		m.handleError(&msg, addr)

	case msg.GetRoleChange() != nil:
		log.Println("Got ROLE_CHANGE message")
		m.handleRoleChange(&msg, addr)

	case msg.GetDiscover() != nil:
		log.Println("Got DISCOVER message")
		m.handleDiscovery(&msg, addr)

	default:
		log.Printf("Unknown message type from %s", addr)
	}
}

func (m *Manager) handlePing(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleSteer(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleAck(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleDiscovery(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleJoin(msg *prt.GameMessage, addr *net.UDPAddr) {
}

func (m *Manager) handleState(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleAnnouncement(msg *prt.GameMessage) {
	games := msg.GetAnnouncement().GetGames()
	if m.gameListener != nil {
		m.gameListener.OnGameAnnouncement(games)
	}
}

func (m *Manager) handleError(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleRoleChange(msg *prt.GameMessage, addr *net.UDPAddr) {}
