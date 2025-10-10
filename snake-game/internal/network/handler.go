package network

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"snake-game/internal/game/logic"
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
		m.handlePing(&msg, addr)

	case msg.GetSteer() != nil:
		m.handleSteer(&msg, addr)

	case msg.GetAck() != nil:
		m.handleAck(&msg, addr)

	case msg.GetState() != nil:
		m.handleState(&msg)

	case msg.GetAnnouncement() != nil:
		m.handleAnnouncement(&msg, addr)

	case msg.GetJoin() != nil:
		m.handleJoin(&msg, addr)

	case msg.GetError() != nil:
		m.handleError(&msg)

	case msg.GetRoleChange() != nil:
		m.handleRoleChange(&msg, addr)

	case msg.GetDiscover() != nil:
		m.handleDiscovery(&msg, addr)

	default:
		log.Printf("Unknown message type from %s", addr)
	}
}

func (m *Manager) handlePing(msg *prt.GameMessage, addr *net.UDPAddr) {}

func (m *Manager) handleSteer(msg *prt.GameMessage, addr *net.UDPAddr) {
	if m.role != prt.NodeRole_MASTER {
		return
	}
	steerMsg := msg.GetSteer()
	if steerMsg == nil {
		return
	}
	senderID := msg.GetSenderId()
	if senderID == 0 {
		return
	}
	direction := steerMsg.GetDirection()
	playerExists := false
	for _, player := range m.gameAnnounce.GetPlayers().GetPlayers() {
		if player.GetId() == senderID {
			playerExists = true
			break
		}
	}
	if !playerExists {
		log.Printf("Steer message from unknown player ID: %d", senderID)
		return
	}
	if m.steerListener != nil {
		err := m.steerListener.OnSteerReceived(senderID, direction)
		if err != nil {
			log.Printf("Error steering snake for player %d: %v", senderID, err)
		}
	}
}

func (m *Manager) handleAck(msg *prt.GameMessage, addr *net.UDPAddr) {
	ackMsg := msg.GetAck()
	if ackMsg == nil {
		return
	}
	if msg.GetReceiverId() != 0 {
		m.playerID = msg.GetReceiverId()
		if m.JoinNotify != nil {
			m.JoinNotify <- m.playerID
		}
		log.Printf("Successfully joined the game! Player ID: %d", m.playerID)
	} else {
		log.Printf("Received ACK for message seq %d from %s", msg.GetMsgSeq(), addr)
	}
}

func (m *Manager) handleDiscovery(msg *prt.GameMessage, addr *net.UDPAddr) {
}

func (m *Manager) handleJoin(msg *prt.GameMessage, addr *net.UDPAddr) {
	if m.role != prt.NodeRole_MASTER {
		return
	}
	joinMsg := msg.GetJoin()
	lgc := m.joinListener.GetLogic()
	if !lgc.CanPlaceSnake() {
		m.gameAnnounce.CanJoin = false
		errorMsg := &prt.GameMessage_ErrorMsg{
			ErrorMessage: "Cannot find suitable position for new snake",
		}
		message := &prt.GameMessage{
			MsgSeq: msg.GetMsgSeq(),
			Type:   &prt.GameMessage_Error{Error: errorMsg},
		}
		data, err := proto.Marshal(message)
		if err != nil {
			log.Printf("Error marshaling error message: %v", err)
			return
		}
		_, err = m.unicastConn.WriteToUDP(data, addr)
		if err != nil {
			log.Printf("Error sending error message: %v", err)
		}
		return
	}
	ackMsg := &prt.GameMessage_AckMsg{}
	ids := make(map[int32]struct{})
	for _, val := range m.gameAnnounce.GetPlayers().Players {
		ids[val.Id] = struct{}{}
	}
	newPlayerID := logic.GeneratePlayerID()
	for {
		if _, exists := ids[newPlayerID]; !exists {
			break
		}
		newPlayerID = logic.GeneratePlayerID()
	}
	player := &prt.GamePlayer{Name: joinMsg.PlayerName, Id: newPlayerID, Type: joinMsg.PlayerType, Role: joinMsg.RequestedRole, Score: 0, IpAddress: addr.IP.String(), Port: int32(addr.Port)}
	m.joinListener.OnGameAddPlayer(player)
	message := &prt.GameMessage{MsgSeq: msg.GetMsgSeq(), Type: &prt.GameMessage_Ack{Ack: ackMsg}, ReceiverId: newPlayerID}
	data, err := proto.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling message: %v", err)
	}
	_, err = m.unicastConn.WriteToUDP(data, addr)
	if err != nil {
		fmt.Printf("Error writing message: %v", err)
	}
}

func (m *Manager) handleState(msg *prt.GameMessage) {
	gameState := msg.GetState().State
	if m.stateListener != nil {
		m.stateListener.OnGameStateReceived(gameState)
	}
}

func (m *Manager) handleAnnouncement(msg *prt.GameMessage, addr *net.UDPAddr) {
	games := msg.GetAnnouncement().GetGames()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, game := range games {
		if m.AvailableGames == nil {
			m.AvailableGames = make(map[string]*GameInfo)
		}
		m.AvailableGames[game.GameName] = &GameInfo{
			Announcement: game,
			MasterAddr:   addr,
		}
	}
	if m.gameListener != nil {
		m.gameListener.OnGameAnnouncement(games)
	}
}

func (m *Manager) handleError(msg *prt.GameMessage) {
	fmt.Println(msg.GetError())
}

func (m *Manager) handleRoleChange(msg *prt.GameMessage, addr *net.UDPAddr) {
}
