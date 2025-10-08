package network

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	prt "snake-game/internal/proto/gen"
	"strconv"
)

func (m *Manager) SendJoinRequest(playerType prt.PlayerType, playerName string, gameName string, role prt.NodeRole) error {
	gameInfo, _ := m.AvailableGames[gameName]
	joinMsg := &prt.GameMessage_JoinMsg{PlayerType: playerType, PlayerName: playerName, GameName: gameName, RequestedRole: role}
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
	_, err = m.unicastConn.WriteToUDP(data, gameInfo.MasterAddr)
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
		log.Printf("Error marshaling message: %v", err)
	}
	var playerAddr *net.UDPAddr
	for _, player := range m.gameAnnounce.GetPlayers().GetPlayers() {
		if player.GetId() == m.playerID {
			continue
		}
		port := strconv.Itoa(int(player.GetPort()))
		playerAddr, err = net.ResolveUDPAddr("udp", net.JoinHostPort(player.IpAddress, port))
		if err != nil {
			log.Printf("Error resolving player address: %v", err)
		}
		_, err := m.unicastConn.WriteToUDP(data, playerAddr)
		if err != nil {
			log.Printf("Error sending state: %v", err)
		}
		m.msgSeq++
	}
	return nil
}

func (m *Manager) SendSteer(dir prt.Direction) {
	steerMsg := &prt.GameMessage_SteerMsg{Direction: dir}
	msg := &prt.GameMessage{
		MsgSeq:   m.msgSeq,
		Type:     &prt.GameMessage_Steer{Steer: steerMsg},
		SenderId: m.playerID,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
	}
	gameInfo, _ := m.AvailableGames[m.currentGameName]
	_, err = m.unicastConn.WriteToUDP(data, gameInfo.MasterAddr)
	if err != nil {
		log.Printf("Error sending steer: %v", err)
	}
	m.msgSeq++
}
