package network

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"snake-game/internal/game/interfaces"
	"snake-game/internal/game/ui"
	prt "snake-game/internal/proto/gen"
	"time"
)

const (
	multicastAddr = "239.255.255.250:9999"
)

type Manager struct {
	unicastConn    *net.UDPConn
	multicastConn  *net.UDPConn
	role           prt.NodeRole
	msgSeq         int64
	gameAnnounce   *prt.GameAnnouncement
	ui             *ui.ConsoleUI
	announceTicker *time.Ticker
	gameListener   interfaces.GameAnnouncementListener
	localPort      int
	availableGames map[string]*GameInfo
	playerID       int32
}

type GameInfo struct {
	Announcement *prt.GameAnnouncement
	MasterAddr   *net.UDPAddr
}

func NewNetworkManager(role prt.NodeRole, gameAnnounce *prt.GameAnnouncement) *Manager {
	return &Manager{
		role:         role,
		msgSeq:       1,
		gameAnnounce: gameAnnounce,
		ui:           ui.NewConsoleUI(),
	}
}

func (m *Manager) SetGameAnnouncementListener(listener interfaces.GameAnnouncementListener) {
	m.gameListener = listener
}

func (m *Manager) GetRole() prt.NodeRole {
	return m.role
}

func (m *Manager) Start() error {
	if err := m.setupUnicastSocket(); err != nil {
		return err
	}
	if err := m.setupMulticastSocket(); err != nil {
		return err
	}
	go m.listenForMessages()
	go m.listenForMulticast()
	if m.role == prt.NodeRole_MASTER {
		m.startAnnouncementBroadcast()
	}
	return nil
}

func (m *Manager) setupUnicastSocket() error {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	m.unicastConn = conn
	log.Printf("Unicast socket bound to %s", conn.LocalAddr())
	return nil
}

func (m *Manager) startAnnouncementBroadcast() {
	m.announceTicker = time.NewTicker(10 * time.Second)
	go func() {
		for range m.announceTicker.C {
			m.sendAnnouncement()
			for _, player := range m.gameAnnounce.GetPlayers().GetPlayers() {
				log.Println(player.Name, player.Role, player.Id)
			}
		}
	}()
}

func (m *Manager) setupMulticastSocket() error {
	groupAddr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return err
	}
	conn, err := net.ListenMulticastUDP("udp", nil, groupAddr)
	if err != nil {
		return err
	}
	m.multicastConn = conn
	log.Printf("Multicast socket joined group %s", groupAddr)
	return nil
}

func (m *Manager) listenForMessages() {
	buf := make([]byte, 4096)
	for {
		n, addr, err := m.unicastConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}
		go m.handleMessage(buf[:n], addr)
	}
}

func (m *Manager) listenForMulticast() {
	buf := make([]byte, 4096)
	for {
		n, addr, err := m.multicastConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}
		go m.handleMessage(buf[:n], addr)
	}
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

func (m *Manager) SendJoinRequest(playerType prt.PlayerType, playerName string, gameName string, role prt.NodeRole) error {
	gameInfo, _ := m.availableGames[gameName]
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

func (m *Manager) ChangeRole(role prt.NodeRole, announcement *prt.GameAnnouncement) {
	m.role = role
	m.gameAnnounce = announcement
	gp := announcement.GetPlayers().GetPlayers()[0]
	m.playerID = gp.Id
	if role == prt.NodeRole_MASTER {
		m.startAnnouncementBroadcast()
	} else if m.announceTicker != nil {
		m.announceTicker.Stop()
		m.announceTicker = nil
	}
}

func (m *Manager) GetID() int32 {
	return m.playerID
}
