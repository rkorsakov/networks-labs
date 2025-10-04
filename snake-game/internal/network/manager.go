package network

import (
	"log"
	"net"
	"snake-game/internal/game/interfaces"
	"snake-game/internal/game/ui"
	prt "snake-game/internal/proto/gen"
	"sync"
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
	stateListener  interfaces.GameStateListener
	joinListener   interfaces.GameJoinListener
	AvailableGames map[string]*GameInfo
	playerID       int32
	mu             sync.Mutex
	closeMutex     sync.Mutex
	closeChan      chan struct{}
	wg             sync.WaitGroup
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
		closeChan:    make(chan struct{}),
	}
}

func (m *Manager) SetGameAnnouncementListener(listener interfaces.GameAnnouncementListener) {
	m.gameListener = listener
}

func (m *Manager) SetGameStateListener(listener interfaces.GameStateListener) {
	m.stateListener = listener
}

func (m *Manager) SetGameJoinListener(listener interfaces.GameJoinListener) {
	m.joinListener = listener
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
	m.wg.Add(2)
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
	return nil
}

func (m *Manager) startAnnouncementBroadcast() {
	m.announceTicker = time.NewTicker(1 * time.Second)
	go func() {
		for range m.announceTicker.C {
			m.sendAnnouncement()
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
	return nil
}

func (m *Manager) listenForMessages() {
	defer m.wg.Done()
	buf := make([]byte, 4096)
	for {
		select {
		case <-m.closeChan:
			return
		default:
		}
		n, addr, err := m.unicastConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}
		go m.handleMessage(buf[:n], addr)
	}
}

func (m *Manager) listenForMulticast() {
	defer m.wg.Done()
	buf := make([]byte, 4096)
	for {
		select {
		case <-m.closeChan:
			return
		default:
		}
		n, addr, err := m.multicastConn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from multicast UDP: %v", err)
			continue
		}
		go m.handleMessage(buf[:n], addr)
	}
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

func (m *Manager) Close() {
	m.closeMutex.Lock()
	defer m.closeMutex.Unlock()
	close(m.closeChan)
	if m.announceTicker != nil {
		m.announceTicker.Stop()
	}
	if m.unicastConn != nil {
		m.unicastConn.Close()
	}
	if m.multicastConn != nil {
		m.multicastConn.Close()
	}
	m.wg.Wait()
}

func (m *Manager) GetID() int32 {
	return m.playerID
}
