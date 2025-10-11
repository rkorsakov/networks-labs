package network

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	prt "snake-game/internal/proto/gen"
	"strconv"
)

func (m *Manager) SendUnicastMessage(data []byte, addr *net.UDPAddr) error {
	if m.activityManager != nil {
		m.activityManager.RecordMessageSent(addr)
	}
	_, err := m.unicastConn.WriteToUDP(data, addr)
	return err
}

func (m *Manager) SendJoinRequest(playerType prt.PlayerType, playerName string, gameName string, role prt.NodeRole) error {
	gameInfo, exists := m.AvailableGames[gameName]
	if !exists {
		return fmt.Errorf("game %s not found", gameName)
	}

	joinMsg := &prt.GameMessage_JoinMsg{
		PlayerType:    playerType,
		PlayerName:    playerName,
		GameName:      gameName,
		RequestedRole: role,
	}
	msg := &prt.GameMessage{
		MsgSeq: m.msgSeq,
		Type: &prt.GameMessage_Join{
			Join: joinMsg,
		},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling join message: %v", err)
	}

	err = m.SendUnicastMessage(data, gameInfo.MasterAddr)
	if err != nil {
		return fmt.Errorf("sending join request: %v", err)
	}

	m.msgSeq++
	log.Printf("Join request sent to %s for game: %s, role: %v",
		gameInfo.MasterAddr, gameName, role)
	return nil
}

func (m *Manager) sendAnnouncement() {
	if m.role != prt.NodeRole_MASTER {
		return
	}
	announcementMsg := &prt.GameMessage_AnnouncementMsg{
		Games: []*prt.GameAnnouncement{m.gameAnnounce},
	}
	msg := &prt.GameMessage{
		MsgSeq: m.msgSeq,
		Type: &prt.GameMessage_Announcement{
			Announcement: announcementMsg,
		},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	groupAddr, _ := net.ResolveUDPAddr("udp", multicastAddr)
	_, err = m.multicastConn.WriteToUDP(data, groupAddr)
	if err != nil {
		log.Printf("Error sending announcement: %v", err)
		return
	}
	m.msgSeq++
}

func (m *Manager) SendState(gameState *prt.GameState) error {
	stateMsg := &prt.GameMessage_StateMsg{State: gameState}
	msg := &prt.GameMessage{
		MsgSeq: m.msgSeq,
		Type:   &prt.GameMessage_State{State: stateMsg},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling state message: %v", err)
	}

	for _, player := range m.gameAnnounce.GetPlayers().GetPlayers() {
		if player.GetId() == m.playerID {
			continue
		}
		port := strconv.Itoa(int(player.GetPort()))
		playerAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(player.IpAddress, port))
		if err != nil {
			log.Printf("Error resolving player address: %v", err)
			continue
		}

		err = m.SendUnicastMessage(data, playerAddr)
		if err != nil {
			log.Printf("Error sending state to player %d: %v", player.GetId(), err)
		}
	}

	m.msgSeq++
	return nil
}

func (m *Manager) SendSteer(dir prt.Direction) error {
	steerMsg := &prt.GameMessage_SteerMsg{Direction: dir}
	msg := &prt.GameMessage{
		MsgSeq:   m.msgSeq,
		Type:     &prt.GameMessage_Steer{Steer: steerMsg},
		SenderId: m.playerID,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling steer message: %v", err)
	}

	gameInfo, exists := m.AvailableGames[m.gameAnnounce.GameName]
	if !exists {
		return fmt.Errorf("game info not found")
	}

	err = m.SendUnicastMessage(data, gameInfo.MasterAddr)
	if err != nil {
		return fmt.Errorf("error sending steer: %v", err)
	}

	m.msgSeq++
	return nil
}

func (m *Manager) sendPing(addr *net.UDPAddr) error {
	pingMsg := &prt.GameMessage_PingMsg{}
	msg := &prt.GameMessage{
		MsgSeq: m.msgSeq,
		Type:   &prt.GameMessage_Ping{Ping: pingMsg},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling ping message: %v", err)
	}

	err = m.SendUnicastMessage(data, addr)
	if err != nil {
		return fmt.Errorf("sending ping: %v", err)
	}

	m.msgSeq++
	//log.Printf("Sent PING to %s", addr)
	return nil
}

func (m *Manager) SendAck(msgSeq int64, receiverId int32, addr *net.UDPAddr) error {
	ackMsg := &prt.GameMessage_AckMsg{}
	msg := &prt.GameMessage{
		MsgSeq:     msgSeq,
		ReceiverId: receiverId,
		Type:       &prt.GameMessage_Ack{Ack: ackMsg},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling ack message: %v", err)
	}

	err = m.SendUnicastMessage(data, addr)
	if err != nil {
		return fmt.Errorf("sending ack: %v", err)
	}

	log.Printf("Sent ACK for message seq %d to %s", msgSeq, addr)
	return nil
}

func (m *Manager) sendRoleChangeMessage(player *prt.GamePlayer, newRole prt.NodeRole) {
	roleChangeMsg := &prt.GameMessage_RoleChangeMsg{
		SenderRole:   m.role,
		ReceiverRole: newRole,
	}

	msg := &prt.GameMessage{
		MsgSeq:     m.msgSeq,
		SenderId:   m.playerID,
		ReceiverId: player.GetId(),
		Type:       &prt.GameMessage_RoleChange{RoleChange: roleChangeMsg},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling role change message: %v", err)
		return
	}

	playerAddr, err := net.ResolveUDPAddr("udp",
		net.JoinHostPort(player.GetIpAddress(), strconv.Itoa(int(player.GetPort()))))
	if err != nil {
		log.Printf("Error resolving player address: %v", err)
		return
	}

	err = m.SendUnicastMessage(data, playerAddr)
	if err != nil {
		log.Printf("Error sending role change message: %v", err)
		return
	}

	log.Printf("Sent role change message to player %s, new role: %v", player.GetName(), newRole)
	m.msgSeq++
}

func (m *Manager) broadcastNewMaster() {
	for _, player := range m.gameAnnounce.GetPlayers().GetPlayers() {
		if player.Role == prt.NodeRole_VIEWER || player.Id == m.playerID {
			continue
		}
		roleChangeMsg := &prt.GameMessage_RoleChangeMsg{
			SenderRole:   prt.NodeRole_MASTER,
			ReceiverRole: player.Role,
		}
		msg := &prt.GameMessage{
			MsgSeq:     m.msgSeq,
			SenderId:   m.playerID,
			ReceiverId: player.Id,
			Type:       &prt.GameMessage_RoleChange{RoleChange: roleChangeMsg},
		}
		data, err := proto.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling new master announcement: %v", err)
			continue
		}
		playerAddr, err := net.ResolveUDPAddr("udp",
			net.JoinHostPort(player.IpAddress, strconv.Itoa(int(player.Port))))
		if err != nil {
			log.Printf("Error resolving player address: %v", err)
			continue
		}
		err = m.SendUnicastMessage(data, playerAddr)
		if err != nil {
			log.Printf("Error sending new master announcement: %v", err)
			continue
		}
		m.msgSeq++
	}
	log.Printf("Broadcasted new master announcement to all players")
}
