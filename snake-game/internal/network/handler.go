package network

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	prt "snake-game/internal/proto/gen"
)

func (m *Manager) handleMessage(data []byte, addr *net.UDPAddr) {
	var msg prt.GameMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}
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
		m.handleAnnouncement(&msg, addr)

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

func (m *Manager) handleAck(msg *prt.GameMessage, addr *net.UDPAddr) {
	ackMsg := msg.GetAck()
	if ackMsg == nil {
		return
	}
	if msg.GetReceiverId() != 0 {
		m.playerID = msg.GetReceiverId()
		log.Printf("Successfully joined the game! Player ID: %d", m.playerID)
	} else {
		log.Printf("Received ACK for message seq %d from %s", msg.GetMsgSeq(), addr)
	}
}

func (m *Manager) handleDiscovery(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleJoin(msg *prt.GameMessage, addr *net.UDPAddr) {
	if m.role != prt.NodeRole_MASTER {
		return
	}
	joinMsg := msg.GetJoin()
	ackMsg := &prt.GameMessage_AckMsg{}
	if joinMsg.RequestedRole == prt.NodeRole_VIEWER {
		message := &prt.GameMessage{MsgSeq: msg.GetMsgSeq(), Type: &prt.GameMessage_Ack{Ack: ackMsg}, ReceiverId: 3}
		data, err := proto.Marshal(message)
		if err != nil {
			fmt.Printf("Error marshaling message: %v", err)
		}
		_, err = m.unicastConn.WriteToUDP(data, addr)
		if err != nil {
			fmt.Printf("Error writing message: %v", err)
		}

	}
}

func (m *Manager) handleState(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleAnnouncement(msg *prt.GameMessage, addr *net.UDPAddr) {
	games := msg.GetAnnouncement().GetGames()
	for _, game := range games {
		if m.availableGames == nil {
			m.availableGames = make(map[string]*GameInfo)
		}
		m.availableGames[game.GameName] = &GameInfo{
			Announcement: game,
			MasterAddr:   addr,
		}
	}
	if m.gameListener != nil {
		m.gameListener.OnGameAnnouncement(games)
	}
}

func (m *Manager) handleError(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleRoleChange(msg *prt.GameMessage, addr *net.UDPAddr) {}
